package immudb_storage

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/jaegertracing/jaeger/model"
	"github.com/jaegertracing/jaeger/plugin/storage/grpc/shared"
	"github.com/jaegertracing/jaeger/storage/dependencystore"
	"github.com/jaegertracing/jaeger/storage/spanstore"
	"time"
)

var (
	_ shared.StoragePlugin   = (*Store)(nil)
	_ spanstore.Reader       = (*Store)(nil)
	_ dependencystore.Reader = (*Store)(nil)
)

type Store struct {
	logger hclog.Logger
	client *ImmuDbDriver
}

func NewStoreQuery(client *ImmuDbDriver, logger hclog.Logger) (*Store, error) {
	client.logger = logger
	return &Store{
		logger: logger,
		client: client,
	}, nil
}

type SpanWriter struct {
	client *ImmuDbDriver
	logger hclog.Logger
}

func (s *Store) SpanWriter() spanstore.Writer {
	s.logger.Warn("init span")
	return &SpanWriter{
		logger: s.logger,
		client: s.client,
	}
}

func (s *SpanWriter) WriteSpan(ctx context.Context, span *model.Span) error {
	s.logger.Warn("write span: ", span)
	return nil
}

func (s *Store) GetTrace(ctx context.Context, traceID model.TraceID) (*model.Trace, error) {
	s.logger.Warn(fmt.Sprintf("Get Trace: %v", traceID))
	return s.client.GetTrace(ctx, traceID)
}

func (s *Store) GetServices(ctx context.Context) ([]string, error) {
	s.logger.Warn(fmt.Sprintf("Get Services"))
	return s.client.GetServices(ctx)
}

func (s *Store) GetOperations(ctx context.Context, query spanstore.OperationQueryParameters) ([]spanstore.Operation, error) {
	s.logger.Warn(fmt.Sprintf("Get Trace: %v", query))
	return s.client.GetOperations(ctx, query)
}

func (s *Store) FindTraces(ctx context.Context, query *spanstore.TraceQueryParameters) ([]*model.Trace, error) {
	s.logger.Warn(fmt.Sprintf("Find Traces: %v", query))
	return s.client.FindTraces(ctx, query)
}

func (s *Store) FindTraceIDs(ctx context.Context, query *spanstore.TraceQueryParameters) ([]model.TraceID, error) {
	s.logger.Warn(fmt.Sprintf("Find Trace IDs: %v", query))
	return []model.TraceID{}, nil
}

func (s *Store) GetDependencies(ctx context.Context, endTs time.Time, lookback time.Duration) ([]model.DependencyLink, error) {
	return nil, nil
}

func (s *Store) SpanReader() spanstore.Reader {
	return s
}
func (s *Store) DependencyReader() dependencystore.Reader {
	return s
}
