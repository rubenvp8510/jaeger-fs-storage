package filesystem

import (
	"flag"
	"github.com/jaegertracing/jaeger/storage/dependencystore"
	"github.com/jaegertracing/jaeger/storage/spanstore"
	"github.com/spf13/viper"
	"github.com/uber/jaeger-lib/metrics"
	"go.uber.org/zap"
)

type Factory struct {
	options Options
	*Store
}

// NewFactory returns a new factory.
func NewFactory() *Factory {
	return &Factory{}
}

func (f *Factory) AddFlags(flagset *flag.FlagSet) {
	AddFlags(flagset)
}

// InitFromViper implements plugin.Configurable.
func (f *Factory) InitFromViper(v *viper.Viper) {
	f.options.InitFromViper(v)
}

func (f *Factory) InitFromConfig(c Config) {
	f.options.Configuration = c
}

// Initialize implements storage.Factory.
func (f *Factory) Initialize(metricsFactory metrics.Factory, zapLogger *zap.Logger) error {
	f.Store = NewStorage(f.options.Configuration)
	f.Store.Init()
	return nil
}

// CreateSpanReader implements storage.Factory.
func (f *Factory) CreateSpanReader() (spanstore.Reader, error) {
	return f.Store, nil
}

// CreateSpanWriter implements storage.Factory.
func (f *Factory) CreateSpanWriter() (spanstore.Writer, error) {
	return f.Store, nil
}

// CreateDependencyReader implements storage.Factory.
func (f *Factory) CreateDependencyReader() (dependencystore.Reader, error) {
	return f.Store, nil
}
