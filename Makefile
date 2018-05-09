lint_output  := lint.txt

.PHONY: all
all: test lint

.PHONY: test
test:
	go get ./...
	go test ./...

.PHONY: lint
lint:
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install
	gometalinter ./... > $(lint_output); true
