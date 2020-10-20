FIXTURES_PATH ?= "../data/traces"

.PHONY: build
build:
	go build main.go

.PHONY: perftest
perftest:
	@FIXTURES_PATH=$(FIXTURES_PATH) go test -bench=. ./perf/... -count=20 -timeout=0
