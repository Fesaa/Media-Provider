package tracing

import "go.opentelemetry.io/otel/attribute"

func WithServiceName(name string) attribute.KeyValue {
	return attribute.String("service.name", name)
}
