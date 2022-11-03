package trace

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type IdGenerator struct {
	RequestId string
}

func (gen IdGenerator) randomHex(n int) (string, error) {
	id := uuid.NewString()
	var src [8]byte
	copy(src[:], id)
	return fmt.Sprintf("%08x", src), nil
}

func (gen IdGenerator) to16Bytes(n string) string {
	var src [16]byte
	copy(src[:], n)
	return fmt.Sprintf("%016x", src)
}

func (gen IdGenerator) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
	traceIDHex := gen.to16Bytes(gen.RequestId)
	traceId, _ := trace.TraceIDFromHex(traceIDHex)
	return traceId, gen.NewSpanID(ctx, traceId)
}

func (gen IdGenerator) NewSpanID(ctx context.Context, traceID trace.TraceID) trace.SpanID {
	spanIDHex, _ := gen.randomHex(8)
	spanID, _ := trace.SpanIDFromHex(spanIDHex)
	return spanID
}
