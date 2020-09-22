package main

import (
	"flag"
	"github.com/jaegertracing/jaeger/plugin/storage/grpc"
	"github.com/jaegertracing/jaeger/storage/dependencystore"
	"github.com/jaegertracing/jaeger/storage/spanstore"
	"github.com/rubenvp8510/jaeger-fs-storage/plugin/filesystem"
	"github.com/spf13/viper"
	"path"
	"strings"
)

var configPath string

func main() {
	flag.StringVar(&configPath, "config", "", "A path to the plugin's configuration file")
	flag.Parse()

	if configPath != "" {
		viper.SetConfigFile(path.Base(configPath))
		viper.AddConfigPath(path.Dir(configPath))
	}

	v := viper.New()
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))

	store := filesystem.NewFileSystem(filesystem.Config{})

	store.Init()

	grpc.Serve(&filesystemStorage{
		store: store,
	})
}

type filesystemStorage struct {
	store *filesystem.FileSystem
}

func (ns *filesystemStorage) DependencyReader() dependencystore.Reader {
	return ns.store
}

func (ns *filesystemStorage) SpanReader() spanstore.Reader {
	return ns.store
}

func (ns *filesystemStorage) SpanWriter() spanstore.Writer {
	return ns.store
}
