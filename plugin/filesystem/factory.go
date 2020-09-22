package filesystem

import (
	"flag"
	"github.com/jaegertracing/jaeger/storage/dependencystore"
	"github.com/jaegertracing/jaeger/storage/spanstore"
	"github.com/spf13/viper"
	"github.com/uber/jaeger-lib/metrics"
	"go.uber.org/zap"
)

const (
	pathFlag = "fsstorage.path"
	ephemeralFlag = "fsstorage.Ephemeral"

)

// Config holds the configuration for redbull.
type Config struct {
	DataDir   string
	Ephemeral bool
}

// Factory is the redbull factory that implements storage.Factory.
type Factory struct {
	cfg Config
	*FileSystem
}

// NewFactory returns a new factory.
func NewFactory() *Factory {
	return &Factory{}
}

// AddFlags implements plugin.Configurable.
func (f *Factory) AddFlags(flagset *flag.FlagSet) {
	flagset.String(pathFlag, ".", "Path to store traces information.")
	flagset.Bool(ephemeralFlag, true, "Ephemeral store data on a temp file")
}

// InitFromViper implements plugin.Configurable.
func (f *Factory) InitFromViper(v *viper.Viper) {
	f.cfg.DataDir = v.GetString(pathFlag)
	f.cfg.Ephemeral = v.GetBool(ephemeralFlag)
}

func (f *Factory) InitFromConfig(c *Config) {
	f.cfg.DataDir = c.DataDir
	f.cfg.Ephemeral = c.Ephemeral
}


// Initialize implements storage.Factory.
func (f *Factory) Initialize(metricsFactory metrics.Factory, zapLogger *zap.Logger) error {
	f.FileSystem = NewFileSystem(f.cfg)
	f.FileSystem.Init()
	return nil
}

// CreateSpanReader implements storage.Factory.
func (f *Factory) CreateSpanReader() (spanstore.Reader, error) {
	return f.FileSystem.SpanReader(), nil
}

// CreateSpanWriter implements storage.Factory.
func (f *Factory) CreateSpanWriter() (spanstore.Writer, error) {
	return f.FileSystem.SpanWriter(), nil
}

// CreateDependencyReader implements storage.Factory.
func (f *Factory) CreateDependencyReader() (dependencystore.Reader, error) {
	return f.FileSystem.DependencyReader(), nil
}