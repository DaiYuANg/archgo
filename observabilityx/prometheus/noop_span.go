package prometheus

import "github.com/DaiYuANg/arcgo/observabilityx"

type noopSpan struct{}

func (noopSpan) End() {}

func (noopSpan) RecordError(err error) {
	_ = err
}

func (noopSpan) SetAttributes(attrs ...observabilityx.Attribute) {
	_ = attrs
}
