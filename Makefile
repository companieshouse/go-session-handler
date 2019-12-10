TESTS ?= ./...
lint_output  := lint.txt

.EXPORT_ALL_VARIABLES:
GO111MODULE = on

.PHONY: all
all: fmt test lint

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: test
test: test-unit

.PHONY: test-unit
test-unit:
	go test $(TESTS) -run 'Unit' -coverprofile=coverage.out

.PHONY: lint
lint: export GO111MODULE=off
lint:
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install
	gometalinter ./... > $(lint_output); true
