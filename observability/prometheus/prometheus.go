package prometheus

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"unicode"

	collectionmapping "github.com/DaiYuANg/arcgo/collectionx/mapping"
	"github.com/DaiYuANg/arcgo/observability"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/samber/lo"
)

const defaultNamespace = "arcgo"

// Option configures Prometheus observability adapter.
type Option func(*config)

type config struct {
	logger    *slog.Logger
	namespace string
	register  prom.Registerer
	gatherer  prom.Gatherer
	buckets   []float64
}

// WithLogger sets logger used by this adapter.
func WithLogger(logger *slog.Logger) Option {
	return func(cfg *config) {
		cfg.logger = logger
	}
}

// WithNamespace sets namespace prefix for metric names.
func WithNamespace(namespace string) Option {
	return func(cfg *config) {
		cfg.namespace = strings.TrimSpace(namespace)
	}
}

// WithRegisterer sets custom metric registerer.
func WithRegisterer(registerer prom.Registerer) Option {
	return func(cfg *config) {
		if registerer != nil {
			cfg.register = registerer
		}
	}
}

// WithGatherer sets custom metric gatherer.
func WithGatherer(gatherer prom.Gatherer) Option {
	return func(cfg *config) {
		if gatherer != nil {
			cfg.gatherer = gatherer
		}
	}
}

// WithHistogramBuckets sets custom histogram buckets.
func WithHistogramBuckets(buckets []float64) Option {
	return func(cfg *config) {
		if len(buckets) > 0 {
			cfg.buckets = buckets
		}
	}
}

// Adapter is a Prometheus-backed observability adapter.
type Adapter struct {
	logger    *slog.Logger
	namespace string
	register  prom.Registerer
	gatherer  prom.Gatherer
	buckets   []float64

	counters   *collectionmapping.ConcurrentMap[string, *counterInstrument]
	histograms *collectionmapping.ConcurrentMap[string, *histInstrument]
}

type counterInstrument struct {
	labels []string
	vec    *prom.CounterVec
}

type histInstrument struct {
	labels []string
	vec    *prom.HistogramVec
}

// New creates a Prometheus adapter.
func New(opts ...Option) *Adapter {
	cfg := config{
		logger:    slog.Default(),
		namespace: defaultNamespace,
		register:  prom.DefaultRegisterer,
		gatherer:  prom.DefaultGatherer,
		buckets:   prom.DefBuckets,
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&cfg)
	}

	return &Adapter{
		logger:     observability.NormalizeLogger(cfg.logger),
		namespace:  normalizeMetricSegment(cfg.namespace, defaultNamespace),
		register:   cfg.register,
		gatherer:   cfg.gatherer,
		buckets:    cfg.buckets,
		counters:   collectionmapping.NewConcurrentMap[string, *counterInstrument](),
		histograms: collectionmapping.NewConcurrentMap[string, *histInstrument](),
	}
}

// Logger returns logger for this adapter.
func (a *Adapter) Logger() *slog.Logger {
	return observability.NormalizeLogger(a.logger)
}

// StartSpan is a no-op for Prometheus adapter.
func (a *Adapter) StartSpan(
	ctx context.Context,
	name string,
	attrs ...observability.Attribute,
) (context.Context, observability.Span) {
	_ = name
	_ = attrs
	if ctx == nil {
		ctx = context.Background()
	}
	return ctx, noopSpan{}
}

// AddCounter records an increment for a named counter.
func (a *Adapter) AddCounter(
	ctx context.Context,
	name string,
	value int64,
	attrs ...observability.Attribute,
) {
	_ = ctx
	if value == 0 {
		return
	}

	metricName := a.normalizeMetricName(name)
	metricLabels := attrsToLabelMap(attrs)
	instrument, err := a.counter(metricName, metricLabels)
	if err != nil {
		a.Logger().Warn("prometheus counter setup failed", "metric", metricName, "error", err.Error())
		return
	}

	labelValues := toPromLabels(instrument.labels, metricLabels)
	counter, err := instrument.vec.GetMetricWith(labelValues)
	if err != nil {
		a.Logger().Warn("prometheus counter labels mismatch", "metric", metricName, "error", err.Error())
		return
	}
	counter.Add(float64(value))
}

// RecordHistogram records a value for a named histogram.
func (a *Adapter) RecordHistogram(
	ctx context.Context,
	name string,
	value float64,
	attrs ...observability.Attribute,
) {
	_ = ctx
	metricName := a.normalizeMetricName(name)
	metricLabels := attrsToLabelMap(attrs)
	instrument, err := a.histogram(metricName, metricLabels)
	if err != nil {
		a.Logger().Warn("prometheus histogram setup failed", "metric", metricName, "error", err.Error())
		return
	}

	labelValues := toPromLabels(instrument.labels, metricLabels)
	histogram, err := instrument.vec.GetMetricWith(labelValues)
	if err != nil {
		a.Logger().Warn("prometheus histogram labels mismatch", "metric", metricName, "error", err.Error())
		return
	}
	histogram.Observe(value)
}

// Handler returns HTTP metrics handler for the configured gatherer.
func (a *Adapter) Handler() http.Handler {
	return promhttp.HandlerFor(a.gatherer, promhttp.HandlerOpts{})
}

func (a *Adapter) counter(metricName string, labels map[string]string) (*counterInstrument, error) {
	if existing, ok := a.counters.Get(metricName); ok {
		return existing, nil
	}

	labelNames := loKeysSorted(labels)
	vec := prom.NewCounterVec(prom.CounterOpts{
		Name: metricName,
		Help: fmt.Sprintf("Counter metric for %s", metricName),
	}, labelNames)

	if err := a.register.Register(vec); err != nil {
		alreadyRegisteredError := &prom.AlreadyRegisteredError{}
		if !errors.As(err, alreadyRegisteredError) {
			return nil, err
		}
		existingVec, ok := alreadyRegisteredError.ExistingCollector.(*prom.CounterVec)
		if !ok {
			return nil, err
		}
		vec = existingVec
	}

	inst := &counterInstrument{
		labels: labelNames,
		vec:    vec,
	}

	actual, _ := a.counters.GetOrStore(metricName, inst)
	return actual, nil
}

func (a *Adapter) histogram(metricName string, labels map[string]string) (*histInstrument, error) {
	if existing, ok := a.histograms.Get(metricName); ok {
		return existing, nil
	}

	labelNames := loKeysSorted(labels)
	vec := prom.NewHistogramVec(prom.HistogramOpts{
		Name:    metricName,
		Help:    fmt.Sprintf("Histogram metric for %s", metricName),
		Buckets: a.buckets,
	}, labelNames)

	if err := a.register.Register(vec); err != nil {
		alreadyRegisteredError := &prom.AlreadyRegisteredError{}
		if !errors.As(err, alreadyRegisteredError) {
			return nil, err
		}
		existingVec, ok := alreadyRegisteredError.ExistingCollector.(*prom.HistogramVec)
		if !ok {
			return nil, err
		}
		vec = existingVec
	}

	inst := &histInstrument{
		labels: labelNames,
		vec:    vec,
	}

	actual, _ := a.histograms.GetOrStore(metricName, inst)
	return actual, nil
}

func (a *Adapter) normalizeMetricName(name string) string {
	metricSegment := normalizeMetricSegment(name, "metric")
	return normalizeMetricSegment(a.namespace+"_"+metricSegment, "arcgo_metric")
}

func normalizeMetricSegment(raw, fallback string) string {
	clean := strings.TrimSpace(raw)
	if clean == "" {
		clean = fallback
	}
	replaced := strings.Map(func(r rune) rune {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r), r == '_', r == ':':
			return unicode.ToLower(r)
		default:
			return '_'
		}
	}, clean)
	replaced = strings.Trim(replaced, "_")
	if replaced == "" {
		replaced = fallback
	}
	firstRune := rune(replaced[0])
	if !unicode.IsLetter(firstRune) && firstRune != '_' && firstRune != ':' {
		replaced = "_" + replaced
	}
	return replaced
}

func attrsToLabelMap(attrs []observability.Attribute) map[string]string {
	if len(attrs) == 0 {
		return nil
	}

	labels := make(map[string]string, len(attrs))
	for _, attr := range attrs {
		labelKey := normalizeLabelKey(attr.Key)
		if labelKey == "" {
			continue
		}
		labels[labelKey] = fmt.Sprint(attr.Value)
	}
	return labels
}

func loKeysSorted(values map[string]string) []string {
	if len(values) == 0 {
		return nil
	}
	keys := lo.Keys(values)
	slices.Sort(keys)
	return keys
}

func toPromLabels(labelNames []string, values map[string]string) prom.Labels {
	if len(labelNames) == 0 {
		return prom.Labels{}
	}
	labels := make(prom.Labels, len(labelNames))
	for _, labelName := range labelNames {
		labels[labelName] = values[labelName]
	}
	return labels
}

func normalizeLabelKey(raw string) string {
	clean := strings.TrimSpace(raw)
	if clean == "" {
		return ""
	}

	replaced := strings.Map(func(r rune) rune {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r), r == '_':
			return unicode.ToLower(r)
		default:
			return '_'
		}
	}, clean)
	replaced = strings.Trim(replaced, "_")
	if replaced == "" {
		return ""
	}

	firstRune := rune(replaced[0])
	if !unicode.IsLetter(firstRune) && firstRune != '_' {
		replaced = "_" + replaced
	}
	return replaced
}

type noopSpan struct{}

func (noopSpan) End() {}

func (noopSpan) RecordError(err error) {
	_ = err
}

func (noopSpan) SetAttributes(attrs ...observability.Attribute) {
	_ = attrs
}
