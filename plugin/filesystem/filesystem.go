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

const numberReadWorkers = 100
const numberWriteWorkers = 100
const numberWriteBuffer = 1

type FileSystem struct {
	basePath      string
	readSemaphore chan struct{}
	writeChannel  chan *model.Span
}

func NewFileSystem(cfg Config) *FileSystem {
	basePath := cfg.DataDir
	if cfg.Ephemeral {
		basePath,_ = ioutil.TempDir("", "filesystem")
	}
	return &FileSystem{
		basePath: basePath,
	}
}

func (fs *FileSystem) Init() {
	for _, dir := range directories {
		subPath := path.Join(fs.basePath, dir)
		if _, err := os.Stat(subPath); os.IsNotExist(err) {
			_ = os.Mkdir(subPath, 0755)
		}
	}
	fs.readSemaphore = make(chan struct{}, numberReadWorkers)
	fs.writeChannel = make(chan *model.Span, numberWriteBuffer)
	fs.startWriteWorkers()

}

func (fs *FileSystem) SpanReader() spanstore.Reader {
	return fs
}

func (fs *FileSystem) SpanWriter() spanstore.Writer {
	return fs
}

func (fs *FileSystem) DependencyReader() dependencystore.Reader {
	return fs
}

func (fs *FileSystem) getTraceConcurrent(wg *sync.WaitGroup, spans []*model.Span, index int, spanPath string) {
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

func (fs *FileSystem) GetTrace(ctx context.Context, traceID model.TraceID) (*model.Trace, error) {
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
		fs.getTraceConcurrent(&wg, span, i, path.Join(tracePath, spanFile.Name()))
	}
	wg.Wait()
	return &model.Trace{
		Spans: span,
	}, nil
}

func (fs *FileSystem) GetServices(ctx context.Context) ([]string, error) {
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

func (fs *FileSystem) GetOperations(ctx context.Context, query spanstore.OperationQueryParameters) ([]spanstore.Operation, error) {
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

func (fs *FileSystem) FindTraces(ctx context.Context, query *spanstore.TraceQueryParameters) ([]*model.Trace, error) {

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

func (fs *FileSystem) FindTraceIDs(ctx context.Context, query *spanstore.TraceQueryParameters) ([]model.TraceID, error) {
	return nil, nil
}

func (fs *FileSystem) startWriteWorkers() {
	for i := 0; i < numberWriteWorkers; i++ {
		go func() {
			for span := range fs.writeChannel {
				_ = fs.writeSingleSpan(span)
			}
		}()
	}
}

func (fs *FileSystem) writeSingleSpan(span *model.Span) error {
	_ = fs.writeMetaData(span)
	writePath := path.Join(fs.basePath, "traces", span.TraceID.String(), span.SpanID.String())
	bytes, _ := proto.Marshal(span)
	err := ioutil.WriteFile(writePath, bytes, 0644)
	if err != nil {
		panic(err)
	}
	return nil
}

func (fs *FileSystem) WriteSpan(context context.Context, span *model.Span) error {
	_ = os.MkdirAll(path.Join(fs.basePath, "traces", span.TraceID.String()), 0755)
	// Send span to write pipeline so one of the workers can write it to the disk
	fs.writeChannel <- span
	return nil
}

func (fs *FileSystem) writeMetaData(span *model.Span) error {
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

func (fs *FileSystem) writeOperation(operation string) error {
	f, err := os.OpenFile(path.Join(fs.basePath, "operations", url.QueryEscape(operation)), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	return f.Close()
}

func (fs *FileSystem) writeService(serviceName string) error {
	f, err := os.OpenFile(path.Join(fs.basePath, "services", url.QueryEscape(serviceName)), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	return f.Close()
}

func (fs *FileSystem) GetDependencies(ctx context.Context, endTs time.Time, lookback time.Duration) ([]model.DependencyLink, error) {
	return nil, nil
}
