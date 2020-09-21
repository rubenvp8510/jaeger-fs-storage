package filesystem

import (
	"flag"
	"github.com/jaegertracing/jaeger/storage/dependencystore"
	"github.com/jaegertracing/jaeger/storage/spanstore"
	"github.com/spf13/viper"
	"github.com/uber/jaeger-lib/metrics"
	"go.uber.org/zap"
)

var (
	logger *zap.SugaredLogger
)

const (
	pathFlag = "fsstorage.path"
)

// Config holds the configuration for redbull.
type Config struct {
	dataDir   string
}

// Factory is the redbull factory that implements storage.Factory.
type Factory struct {
	cfg Config
	*fileSystem
}

// NewFactory returns a new factory.
func NewFactory() *Factory {
	return &Factory{}
}

// AddFlags implements plugin.Configurable.
func (f *Factory) AddFlags(flagset *flag.FlagSet) {
	flagset.String(pathFlag, ".", "Path to store traces information.")
}


// InitFromViper implements plugin.Configurable.
func (f *Factory) InitFromViper(v *viper.Viper) {
	f.cfg.dataDir = v.GetString(pathFlag)
}

// Initialize implements storage.Factory.
func (f *Factory) Initialize(metricsFactory metrics.Factory, zapLogger *zap.Logger) error {
	logger = zapLogger.Sugar()
	f.fileSystem = NewFileSystem(f.cfg)
	f.fileSystem.Init()
	return nil
}

// CreateSpanReader implements storage.Factory.
func (f *Factory) CreateSpanReader() (spanstore.Reader, error) {
	return f.fileSystem.SpanReader(), nil
}

// CreateSpanWriter implements storage.Factory.
func (f *Factory) CreateSpanWriter() (spanstore.Writer, error) {
	return f.fileSystem.SpanWriter(), nil
}

// CreateDependencyReader implements storage.Factory.
func (f *Factory) CreateDependencyReader() (dependencystore.Reader, error) {
	return f.fileSystem.DependencyReader(), nil
}