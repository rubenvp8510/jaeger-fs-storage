package filesystem

import (
	"context"
	"github.com/gogo/protobuf/proto"
	"github.com/jaegertracing/jaeger/model"
	"github.com/jaegertracing/jaeger/storage/spanstore"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"sync"
	"time"
)

var directories = []string{"operations", "traces", "services"}

type Store struct {
	config        Config
	readSemaphore chan struct{}
	writeChannel  chan *model.Span
}

func NewStorage(options Options) *Store {
	cfg := options.Configuration
	if cfg.Ephemeral {
		cfg.DataDir, _ = ioutil.TempDir("", "filesystem")
	}
	return &Store{
		config: cfg,
	}
}

func (fs *Store) Init() {
	for _, dir := range directories {
		subPath := path.Join(fs.config.DataDir, dir)
		if _, err := os.Stat(subPath); os.IsNotExist(err) {
			_ = os.Mkdir(subPath, 0755)
		}
	}
	fs.readSemaphore = make(chan struct{}, fs.config.NumberReadWorkers)
	fs.writeChannel = make(chan *model.Span, fs.config.WriteBufferSize)
	fs.startWriteWorkers()

}

func (fs *Store) GetTrace(ctx context.Context, traceID model.TraceID) (*model.Trace, error) {
	tracePath := path.Join(fs.config.DataDir, "traces", traceID.String())
	files, err := ioutil.ReadDir(tracePath)
	if err != nil {
		return nil, err
	}
	nSpans := len(files)
	span := make([]*model.Span, nSpans)

	var wg sync.WaitGroup
	wg.Add(nSpans)

	for i, spanFile := range files {
		fs.getTraceConcurrent(&wg, span, i, path.Join(tracePath, spanFile.Name()))
	}
	wg.Wait()
	return &model.Trace{
		Spans: span,
	}, nil
}

func (fs *Store) GetServices(ctx context.Context) ([]string, error) {
	files, err := ioutil.ReadDir(path.Join(fs.config.DataDir, "services"))
	if err != nil {
		return nil, err
	}
	services := make([]string, len(files))
	for i, srv := range files {
		srvName, _ := url.QueryUnescape(srv.Name())
		services[i] = srvName
	}
	return services, nil
}

func (fs *Store) GetOperations(ctx context.Context, query spanstore.OperationQueryParameters) ([]spanstore.Operation, error) {
	files, err := ioutil.ReadDir(path.Join(fs.config.DataDir, "operations"))
	if err != nil {
		return nil, err
	}
	operations := make([]spanstore.Operation, len(files))
	for i, op := range files {
		opName, _ := url.QueryUnescape(op.Name())
		operations[i].Name = opName
	}
	return operations, nil
}

func (fs *Store) FindTraces(ctx context.Context, query *spanstore.TraceQueryParameters) ([]*model.Trace, error) {

	// Ignore query because we cannot do complex queries just with filesystem
	// This will return all available traces and is just for demo purposes.
	traceIds, err := ioutil.ReadDir(path.Join(fs.config.DataDir, "traces"))
	if err != nil {
		return nil, err
	}
	traces := make([]*model.Trace, len(traceIds))

	for i, traceId := range traceIds {
		tid, err := model.TraceIDFromString(traceId.Name())
		if err != nil {
			return nil, err
		}

		trace, err := fs.GetTrace(ctx, tid)
		if err != nil {
			return nil, err
		}
		traces[i] = trace
	}
	return traces, nil
}

func (fs *Store) FindTraceIDs(ctx context.Context, query *spanstore.TraceQueryParameters) ([]model.TraceID, error) {
	return nil, nil
}

func (fs *Store) GetDependencies(ctx context.Context, endTs time.Time, lookback time.Duration) ([]model.DependencyLink, error) {
	return nil, nil
}

func (fs *Store) WriteSpan(context context.Context, span *model.Span) error {
	_ = os.MkdirAll(path.Join(fs.config.DataDir, "traces", span.TraceID.String()), 0755)
	// Send span to write pipeline so one of the workers can write it to the disk
	fs.writeChannel <- span
	return nil
}

func (fs *Store) getTraceConcurrent(wg *sync.WaitGroup, spans []*model.Span, index int, spanPath string) {
	fs.readSemaphore <- struct{}{}
	go func(index int, path string) {
		msg, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}
		s := &model.Span{}
		_ = proto.Unmarshal(msg, s)
		spans[index] = s
		wg.Done()
		<-fs.readSemaphore
	}(index, spanPath)
}

func (fs *Store) startWriteWorkers() {
	for i := 0; i < fs.config.NumberWriteWorkers; i++ {
		go func() {
			for span := range fs.writeChannel {
				_ = fs.writeSingleSpan(span)
			}
		}()
	}
}

func (fs *Store) writeSingleSpan(span *model.Span) error {
	_ = fs.writeMetaData(span)
	writePath := path.Join(fs.config.DataDir, "traces", span.TraceID.String(), span.SpanID.String())
	bytes, _ := proto.Marshal(span)
	err := ioutil.WriteFile(writePath, bytes, 0644)
	if err != nil {
		panic(err)
	}
	return nil
}

func (fs *Store) writeMetaData(span *model.Span) error {
	err := fs.writeOperation(span.OperationName)
	if err != nil {
		return err
	}
	err = fs.writeService(span.Process.ServiceName)
	if err != nil {
		return err
	}
	return nil
}

func (fs *Store) writeOperation(operation string) error {
	f, err := os.OpenFile(path.Join(fs.config.DataDir, "operations", url.QueryEscape(operation)), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	return f.Close()
}

func (fs *Store) writeService(serviceName string) error {
	f, err := os.OpenFile(path.Join(fs.config.DataDir, "services", url.QueryEscape(serviceName)), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	return f.Close()
}
