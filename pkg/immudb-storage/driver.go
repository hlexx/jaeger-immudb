package immudb_storage

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/codenotary/immudb/pkg/api/schema"
	immudb "github.com/codenotary/immudb/pkg/client"
	badgerV3 "github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/badger/v3/pb"
	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/go-hclog"
	"github.com/hlexx/jaeger-immudb/pkg/cached"
	"github.com/hlexx/jaeger-immudb/pkg/utils"
	"github.com/jaegertracing/jaeger/model"
	"github.com/jaegertracing/jaeger/storage/spanstore"
	"io/ioutil"
	"strings"
	"sync"
	"time"
)

const (
	spansTable      = "spans"
	operationsTable = "operations"
	tracesTable     = "traces"
	indexStartTime  = "startTime"
	servicesTable   = "services"
	tagsTable       = "tags"
	ttl             = time.Hour * 24
	MaxKeyLen       = 1024
)

var (
	mtx    sync.Mutex
	tables = []string{
		spansTable,
		operationsTable,
		tracesTable,
		servicesTable}
)

type Config struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	User     string `yaml:"user"`
	Pwd      string `yaml:"pwd"`
}

type ImmuDbDriver struct {
	cfg    *Config
	logger hclog.Logger
	Client immudb.ImmuClient
}
type BWriter struct {
	client      *ImmuDbDriver
	cacheBackup *cached.Data
}

func New(cfgPath string) (*ImmuDbDriver, error) {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:       "jaeger",
		Level:      hclog.Warn,
		JSONFormat: true,
	})
	yamlFile, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}
	var c Config
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		return nil, err
	}
	return &ImmuDbDriver{
		cfg:    &c,
		logger: logger,
	}, err
}

func (driver *ImmuDbDriver) CreateDataBase() error {
	client, err := driver.OpenSession()
	defer client.CloseSession(context.Background())
	if err != nil {
		return err
	}
	for _, table := range tables {
		_, err = client.CreateDatabaseV2(context.Background(), table, nil)
		if err != nil && err.Error() != "database already exists" {
			return err
		}
	}
	return nil
}

func (driver *ImmuDbDriver) OpenSession(database ...string) (immudb.ImmuClient, error) {
	initDatabase := immudb.DefaultDB
	ctx := context.Background()
	if len(database) > 0 {
		initDatabase = database[0]
	}
	opts := immudb.DefaultOptions().
		WithAddress(driver.cfg.Host).
		WithPort(driver.cfg.Port).
		WithMetrics(false)
	client := immudb.NewClient().WithOptions(opts)
	err := client.OpenSession(ctx, []byte(driver.cfg.User), []byte(driver.cfg.Pwd), initDatabase)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (driver *ImmuDbDriver) GetOperations(ctx context.Context, query spanstore.OperationQueryParameters) ([]spanstore.Operation, error) {
	client, err := driver.OpenSession()
	defer client.CloseSession(context.Background())
	if err != nil {
		return nil, err
	}
	resp, err := client.Scan(ctx, &schema.ScanRequest{
		Prefix: []byte(fmt.Sprintf("%s", operationsTable)),
	})
	if err != nil {
		return nil, err
	}
	var operations []spanstore.Operation
	for _, entry := range resp.Entries {
		var operation spanstore.Operation
		err := json.Unmarshal(entry.Value, &operation)
		if err != nil {
			return nil, err
		}
		operations = append(operations, operation)
	}
	return operations, nil
}

func (driver *ImmuDbDriver) GetServices(ctx context.Context) ([]string, error) {
	client, err := driver.OpenSession()
	defer client.CloseSession(context.Background())
	if err != nil {
		return nil, err
	}
	resp, err := client.Scan(ctx, &schema.ScanRequest{
		Prefix: []byte(fmt.Sprintf("%s", servicesTable)),
	})
	if err != nil {
		return nil, err
	}
	var services []string
	for _, entry := range resp.Entries {
		services = append(services, string(entry.Value))
	}
	return services, nil
}

func (driver *ImmuDbDriver) Writer(ctx context.Context, key, value []byte) error {
	client, err := driver.OpenSession()
	if err != nil {
		driver.logger.Error("writer not open session err: %v", err)
		return err
	}
	defer func() {
		err := client.CloseSession(context.Background())
		if err != nil {
			driver.logger.Error("writer not close session err: %v", err)
			return
		}
	}()
	sp := model.Span{}
	err = sp.Unmarshal(value)
	if err != nil {
		driver.logger.Error("writer not unmarshal value err: %v", err)
		return err
	}
	if sp.Process == nil {
		return nil
	}
	traceId := sp.TraceID
	spanId := sp.SpanID
	spanKey := fmt.Sprintf("%s-%s-%s", spansTable, traceId, spanId)
	/*begin transaction*/
	spanIndex, err := client.VerifiedSet(ctx, []byte(spanKey), value)
	if err != nil {
		driver.logger.Error("add span err: %v", err)
		return err
	}
	var KVList []*schema.KeyValue
	for _, tag := range sp.Tags {
		indexTag := fmt.Sprintf("%s-%s-%s-%s-%s", tagsTable, tag.Key, tag.Value(), traceId, spanId)
		bytesTag := []byte(indexTag)
		if len(bytesTag) >= MaxKeyLen {
			continue
		}
		KVList = append(KVList, &schema.KeyValue{
			Key:   bytesTag,
			Value: utils.ToJsonBytes(spanIndex.Id),
		})
	}
	for _, tag := range sp.Process.Tags {
		indexTag := fmt.Sprintf("%s-%s-%s-%s-%s", tagsTable, tag.Key, tag.Value(), traceId, spanId)
		bytesTag := []byte(indexTag)
		if len(bytesTag) >= MaxKeyLen {
			continue
		}
		KVList = append(KVList, &schema.KeyValue{
			Key:   bytesTag,
			Value: utils.ToJsonBytes(spanIndex.Id),
		})
	}
	setRequest := &schema.SetRequest{KVs: KVList}
	_, err = client.SetAll(ctx, setRequest)
	if err != nil {
		return err
	}
	startTime := sp.StartTime.Unix()
	_, err = client.ZAddAt(ctx, []byte(indexStartTime), float64(startTime), []byte(spanKey), spanIndex.Id)
	if err != nil {
		driver.logger.Error("add [indexStartTime] to trace meta data err: ", err)
		return err
	}
	spanKind, _ := sp.GetSpanKind()
	operation := spanstore.Operation{
		Name:     sp.OperationName,
		SpanKind: spanKind,
	}
	operationKey := fmt.Sprintf("%s-%s", operationsTable, operation.Name)
	_, err = client.VerifiedSet(ctx, []byte(operationKey), utils.ToJsonBytes(operation))
	if err != nil {
		driver.logger.Error("add operation err: %v", err)
		return err
	}
	serviceName := sp.Process.ServiceName
	serviceKey := fmt.Sprintf("%s-%s", servicesTable, serviceName)
	_, err = client.VerifiedSet(ctx, []byte(serviceKey), []byte(serviceName))
	if err != nil {
		driver.logger.Error("add service err: %v", err)
		return err
	}
	return nil
}
func (driver *ImmuDbDriver) FindTraces(ctx context.Context, query *spanstore.TraceQueryParameters) ([]*model.Trace, error) {
	if len(query.Tags) <= 0 {
		return driver.FindTracesTime(ctx, query)
	}
	client, err := driver.OpenSession()
	defer client.CloseSession(context.Background())
	if err != nil {
		return nil, err
	}
	scanOpts := &schema.ScanRequest{}
	for k, v := range query.Tags {
		indexTag := fmt.Sprintf("%s-%s-%s", tagsTable, k, v)
		scanOpts.Prefix = []byte(indexTag)
	}
	resp, err := client.Scan(ctx, scanOpts)
	if err != nil {
		return nil, err
	}
	var traces []*model.Trace
	trace := &model.Trace{
		Spans: []*model.Span{},
	}
	for _, entry := range resp.Entries {
		var txId uint64
		err := json.Unmarshal(entry.Value, &txId)
		if err != nil {
			return nil, err
		}
		client, err := driver.OpenSession()
		defer client.CloseSession(context.Background())
		if err != nil {
			return nil, err
		}
		txByID, err := client.TxByID(context.Background(), txId)
		if err != nil {
			return nil, err
		}
		kvEntries := txByID.GetEntries()
		for _, kvEntry := range kvEntries {
			var span model.Span
			get, err := client.Get(ctx, kvEntry.Key)
			if err != nil {
				return nil, err
			}
			if err != nil {
				return nil, err
			}
			err = span.Unmarshal(get.Value)
			if err != nil {
				return nil, err
			}
			trace.Spans = append(trace.Spans, &span)
		}
	}
	traces = append(traces, trace)
	return traces, nil
}

func (driver *ImmuDbDriver) GetTrace(ctx context.Context, traceID model.TraceID) (*model.Trace, error) {
	client, err := driver.OpenSession()
	defer client.CloseSession(context.Background())
	if err != nil {
		return nil, err
	}
	traceId := traceID.String()
	resp, err := client.Scan(ctx, &schema.ScanRequest{
		Prefix: []byte(fmt.Sprintf("%s-%s", spansTable, traceId)),
	})
	if err != nil {
		return nil, err
	}
	trace := &model.Trace{
		Spans: []*model.Span{},
	}
	for _, entry := range resp.Entries {
		var span model.Span
		err := span.Unmarshal(entry.Value)
		if err != nil {
			return nil, err
		}
		trace.Spans = append(trace.Spans, &span)
	}
	return trace, nil
}

func (driver *ImmuDbDriver) GetAllSpan(ctx context.Context) ([]*model.Span, error) {
	client, err := driver.OpenSession()
	if err != nil {
		return nil, err
	}
	defer client.CloseSession(context.Background())
	resp, err := client.Scan(ctx, &schema.ScanRequest{
		Limit: 100,
	})
	if err != nil {
		return nil, err
	}
	var spans []*model.Span
	for _, entry := range resp.Entries {
		var span model.Span
		err := json.Unmarshal(entry.Value, &span)
		if err != nil {
			return nil, err
		}
		spans = append(spans, &span)
	}
	return spans, nil

}

func (driver *ImmuDbDriver) FindTracesTime(ctx context.Context, query *spanstore.TraceQueryParameters) ([]*model.Trace, error) {
	client, err := driver.OpenSession()
	if err != nil {
		return nil, err
	}
	defer client.CloseSession(context.Background())
	chunks, err := driver.scanRangeIndex(ctx, query)
	if err != nil {
		return nil, err
	}
	limit := 999
	var traces []*model.Trace
	for _, k := range chunks.Keys() {
		offset := 0
		trace := &model.Trace{
			Spans: []*model.Span{},
		}
		for {
			response, err := client.Scan(ctx, &schema.ScanRequest{
				Prefix: []byte(k),
				Limit:  uint64(limit),
				Offset: uint64(offset),
				Desc:   true,
				NoWait: true,
			})
			if err != nil {
				return nil, err
			}
			if len(response.Entries) == 0 {
				break
			}
			for _, entry := range response.Entries {
				if len(entry.Key) == 0 {
					break
				}
				var span model.Span
				err := span.Unmarshal(entry.Value)
				if err != nil {
					return nil, err
				}
				trace.Spans = append(trace.Spans, &span)
			}
			offset += limit
		}
		if len(trace.Spans) > 0 {
			if len(traces) > query.NumTraces-1 {
				break
			}
			traces = append(traces, trace)
		}
	}
	return traces, nil
}

func (driver *ImmuDbDriver) scanRangeIndex(ctx context.Context, query *spanstore.TraceQueryParameters) (*utils.SortedMap, error) {
	client, err := driver.OpenSession()
	if err != nil {
		driver.logger.Warn("scan time failed", err.Error())
		return nil, err
	}
	defer client.CloseSession(context.Background())
	var zScanOpts *schema.ZScanRequest
	offset := 0
	limit := 999
	chunks := utils.NewMap()
	for {
		zScanOpts = &schema.ZScanRequest{
			Set:      []byte(indexStartTime),
			Limit:    uint64(limit),
			Offset:   uint64(offset),
			Desc:     true,
			MinScore: &schema.Score{Score: float64(query.StartTimeMin.Unix())},
			MaxScore: &schema.Score{Score: float64(query.StartTimeMax.Unix())},
		}
		resp, err := client.ZScan(ctx, zScanOpts)
		if err != nil {
			return nil, err
		}
		if len(resp.Entries) == 0 {
			break
		}
		for _, entry := range resp.Entries {
			items := strings.Split(string(entry.Key), "-")
			tracePrefix := fmt.Sprintf("%s-%s", items[0], items[1])
			chunks.Set(tracePrefix, 1)
		}
		offset += limit
	}
	return chunks, nil
}

func (receiver *BWriter) Write(p []byte) (n int, err error) {
	var list pb.KVList
	if len(p) <= 8 {
		return
	}
	err = proto.Unmarshal(p, &list)
	if err != nil {
		return 0, err
	}
	for _, kv := range list.Kv {
		if kv.Key == nil || kv.Value == nil {
			continue
		}
		keyVersion := fmt.Sprintf("%d", kv.Version)
		get, err := receiver.cacheBackup.Exist(keyVersion)
		if err != nil {
			return 0, err
		}
		if get != nil {
			continue
		}
		err = receiver.client.Writer(context.Background(), kv.Key, kv.Value)
		if err != nil {
			return 0, err
		}
		err = receiver.cacheBackup.AddWithTTL(keyVersion, []byte(keyVersion), ttl)
		if err != nil {
			return 0, err
		}
	}
	return
}

func (driver *ImmuDbDriver) ImportFromBackup(db *badgerV3.DB) error {
	cache, err := cached.Connect("backup")
	if err != nil {
		return err
	}
	bWriter := BWriter{client: driver, cacheBackup: cache}
	_, err = db.Backup(&bWriter, 0)
	if err != nil {
		return err
	}
	return db.Close()
}
