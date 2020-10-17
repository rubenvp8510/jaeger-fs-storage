package perf

import (
	"github.com/jaegertracing/jaeger/storage"
	"github.com/rubenvp8510/jaeger-fs-storage/plugin/filesystem"
	"github.com/rubenvp8510/jaeger-storage-perf/perftest"
	"os"
	"testing"
)

const fixturesPathEnvKey = "FIXTURES_PATH"

func initFactory() storage.Factory {
	factory := filesystem.NewFactory()
	factory.InitFromConfig(filesystem.Config{
		Ephemeral: true,
	})
	return factory
}

func BenchmarkRead(t *testing.B) {
	factories := map[string]storage.Factory{
		"filesystem": initFactory(),
	}
	perftest.RunRead(t,os.Getenv(fixturesPathEnvKey), factories)
}

func BenchmarkWrite(t *testing.B) {
	factories := map[string]storage.Factory{
		"filesystem": initFactory(),
	}
	perftest.RunWrite(t, os.Getenv(fixturesPathEnvKey), factories)
}
