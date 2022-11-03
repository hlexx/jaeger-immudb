package trace

import (
	"github.com/hlexx/jaeger-immudb/pkg/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

const (
	ServiceKey = "service.name"
	ErrorKey   = "err"
)

func newProvider(id string, service string) (*sdk.TracerProvider, error) {
	url := utils.GetEnv("JAEGER_URL", "http://localhost:14268/api/traces")
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}
	attKV := []attribute.KeyValue{
		semconv.ServiceNameKey.String(service),
		semconv.ServiceInstanceIDKey.String(id),
	}
	semAtt := resource.NewWithAttributes(semconv.SchemaURL, attKV...)
	provider := sdk.NewTracerProvider(
		sdk.WithIDGenerator(IdGenerator{RequestId: id}),
		sdk.WithBatcher(exp),
		sdk.WithResource(semAtt),
	)
	return provider, nil
}
