package eventx

import (
	"reflect"
	"strings"

	"github.com/DaiYuANg/arcgo/observability"
)

func (b *Bus) observabilitySafe() observability.Observability {
	if b == nil {
		return observability.Nop()
	}
	return observability.Normalize(b.observability, b.logger)
}

func eventName(event Event) string {
	if event == nil {
		return ""
	}

	name := strings.TrimSpace(event.Name())
	if name != "" {
		return name
	}
	return reflect.TypeOf(event).String()
}
