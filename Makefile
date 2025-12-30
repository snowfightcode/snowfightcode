.PHONY: build test clean install help

# Default target
.DEFAULT_GOAL := help

# Build the snowfight binary
build:
	CGO_ENABLED=1 go build -mod=mod -ldflags="-buildid=" -o snowfight ./cmd/snowfight

# Run tests
test:
	CGO_ENABLED=1 go test -mod=mod ./...

# Run tests with verbose output
test-verbose:
	CGO_ENABLED=1 go test -mod=mod -v ./...

# Run scenario tests
test-scenarios:
	CGO_ENABLED=1 go test -mod=mod -v ./scenarios_test/...

# Clean build artifacts
clean:
	rm -f snowfight
	rm -f match.jsonl

# Install to GOPATH/bin
install:
	CGO_ENABLED=1 go install -mod=mod -ldflags="-buildid=" ./cmd/snowfight

# Download dependencies
deps:
	CGO_ENABLED=1 go mod download github.com/buke/quickjs-go@v0.6.6

# Run a sample match
demo: build
	./snowfight match testdata/p1.js testdata/p2.js > match.jsonl
	./snowfight visualize match.jsonl
	@echo "Open dist/index.html in your browser to view the match"

# Display help
help:
	@echo "SnowFight: Code - Makefile targets"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build           Build the snowfight binary"
	@echo "  test            Run all tests"
	@echo "  test-verbose    Run tests with verbose output"
	@echo "  test-scenarios  Run scenario tests only"
	@echo "  clean           Remove build artifacts"
	@echo "  install         Install to GOPATH/bin"
	@echo "  deps            Download dependencies"
	@echo "  demo            Build and run a sample match"
	@echo "  help            Show this help message"
