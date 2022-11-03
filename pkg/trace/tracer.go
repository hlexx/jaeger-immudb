package trace

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"time"
)

// Trace add trace to jaeger
func Trace(id, component string, args map[string]string) {
	provider, err := newProvider(id, component)
	if err != nil {
		return
	}
	otel.SetTracerProvider(provider)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func(ctx context.Context) {
		ctx, cancel = context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := provider.Shutdown(ctx); err != nil {
			fmt.Printf("provider %s\n", err.Error())
		}
	}(ctx)
	tr := provider.Tracer(component)
	_, span := tr.Start(ctx, component)
	call := Call(4)
	args["call.file"] = call.FileName
	args["call.line"] = fmt.Sprintf("%d", call.Line)
	args["call.function"] = call.FunctionName
	var attKV []attribute.KeyValue
	for k, v := range args {
		attKV = append(attKV, attribute.String(k, v))
	}
	if val, exists := args[ErrorKey]; exists {
		span.SetStatus(codes.Error, val)
	}
	span.SetAttributes(attKV...)
	defer span.End()
}
