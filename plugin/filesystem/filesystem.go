package filesystem

import (
	"context"
	"github.com/gogo/protobuf/proto"
	"github.com/jaegertracing/jaeger/model"
	"github.com/jaegertracing/jaeger/storage/dependencystore"
	"github.com/jaegertracing/jaeger/storage/spanstore"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"sync"
	"time"
)

var directories = []string{"operations", "traces", "services"}

type fileSystem struct {
	basePath string
}

func NewFileSystem(cfg Config) *fileSystem {
	return &fileSystem{
		basePath: cfg.dataDir,
	}
}

func (fs *fileSystem) Init() {
	for _, dir := range directories {
		subPath := path.Join(fs.basePath, dir)
		if _, err := os.Stat(subPath); os.IsNotExist(err) {
			os.Mkdir(subPath, 0755)
		}
	}
}

func (fs *fileSystem) SpanReader() spanstore.Reader {
	return fs
}

func (fs *fileSystem) SpanWriter() spanstore.Writer {
	return fs
}

func (fs *fileSystem) DependencyReader() dependencystore.Reader {
	return fs
}

func (fs *fileSystem) GetTrace(ctx context.Context, traceID model.TraceID) (*model.Trace, error) {
	tracePath := path.Join(fs.basePath, "traces", traceID.String())
	files, err := ioutil.ReadDir(tracePath)
	if err != nil {
		return nil, err
	}
	nSpans := len(files)
	span := make([]*model.Span, nSpans)

	var wg sync.WaitGroup
	wg.Add(nSpans)

	for i, spanFile := range files {
		go func(spanId string, index int) {
			defer wg.Done()
			msg, err := ioutil.ReadFile(path.Join(tracePath, spanId))
			if err != nil {
				return
			}
			s := &model.Span{}
			_ = proto.Unmarshal(msg, s)
			span[index] = s
		}(spanFile.Name(), i)
	}

	wg.Wait()
	return &model.Trace{
		Spans: span,
	}, nil
}

func (fs *fileSystem) GetServices(ctx context.Context) ([]string, error) {
	files, err := ioutil.ReadDir(path.Join(fs.basePath, "services"))
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

func (fs *fileSystem) GetOperations(ctx context.Context, query spanstore.OperationQueryParameters) ([]spanstore.Operation, error) {
	files, err := ioutil.ReadDir(path.Join(fs.basePath, "operations"))
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

func (fs *fileSystem) FindTraces(ctx context.Context, query *spanstore.TraceQueryParameters) ([]*model.Trace, error) {

	// Ignore query because we cannot do complex queries just with filesystem
	// This will return all available traces and is just for demo purposes.
	traceIds, err := ioutil.ReadDir(path.Join(fs.basePath, "traces"))
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

func (fs *fileSystem) FindTraceIDs(ctx context.Context, query *spanstore.TraceQueryParameters) ([]model.TraceID, error) {
	return nil, nil
}

func (fs *fileSystem) writeSingleSpan(span *model.Span) error {
	_ = os.MkdirAll(path.Join(fs.basePath, "traces", span.TraceID.String()), 0755)

	writePath := path.Join(fs.basePath, "traces", span.TraceID.String(), span.SpanID.String())

	bytes, err := proto.Marshal(span)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(writePath, bytes, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (fs *fileSystem) WriteSpan(span *model.Span) error {
	go fs.writeMetaData(span)
	go fs.writeSingleSpan(span)
	return nil
}

func (fs *fileSystem) writeMetaData(span *model.Span) {
	fs.writeOperation(span.OperationName)
	fs.writeService(span.Process.ServiceName)
}

func (fs *fileSystem) writeOperation(operation string) error {
	f, err := os.OpenFile(path.Join(fs.basePath, "operations", url.QueryEscape(operation)), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	return f.Close()
}

func (fs *fileSystem) writeService(serviceName string) error {
	f, err := os.OpenFile(path.Join(fs.basePath, "services", url.QueryEscape(serviceName)), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	return f.Close()
}

func (fs *fileSystem) GetDependencies(endTs time.Time, lookback time.Duration) ([]model.DependencyLink, error) {
	return nil, nil
}
