.PHONY: run build test clean help

BINARY_API_NAME=dns-api
BINARY_DNS_NAME=dns-server

# Default target
all: build

# Run the API and DNS servers
run:
	@echo "Starting API server..."
	go run ./cmd/api/main.go & \
	@echo "Starting DNS server..."
	go run ./cmd/dns/main.go

# Build the binaries
build:
	@echo "Building API server..."
	go build -o bin/$(BINARY_API_NAME) ./cmd/api
	@echo "Building DNS server..."
	go build -o bin/$(BINARY_DNS_NAME) ./cmd/dns

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. ./...

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	go clean
	rm -f bin/$(BINARY_API_NAME)
	rm -f bin/$(BINARY_DNS_NAME)

# Help message
help:
	@echo "Available commands:"
	@echo "  run    - Run the API and DNS servers"
	@echo "  build  - Build the binaries"
	@echo "  test   - Run all tests"
	@echo "  bench  - Run all benchmarks"
	@echo "  clean  - Clean up build artifacts"
```
```go
