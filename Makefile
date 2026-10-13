TESTS ?= ./...

.EXPORT_ALL_VARIABLES:
GO111MODULE = on

.PHONY: all
all: clean fmt test

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: test
test: test-unit

.PHONY: test-unit
test-unit:
	go test $(TESTS) -run 'Unit' -coverprofile=coverage.out

.PHONY: clean
clean:
	go mod tidy

