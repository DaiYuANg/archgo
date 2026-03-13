package prometheus

import (
	"errors"
	"fmt"

	prom "github.com/prometheus/client_golang/prometheus"
)

func (a *Adapter) counter(metricName string, labels map[string]string) (*counterInstrument, error) {
	if existing, ok := a.counters.Get(metricName); ok {
		return existing, nil
	}

	labelNames := sortedLabelKeys(labels)
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

	labelNames := sortedLabelKeys(labels)
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
