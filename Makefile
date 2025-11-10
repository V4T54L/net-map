.PHONY: all run build test bench clean help swag setup

BINARY_API_NAME=dns-api
BINARY_DNS_NAME=dns-server

all: build

run:
	@echo "Starting API server..."
	go run ./cmd/api/main.go & \
	@echo "Starting DNS server..."
	go run ./cmd/dns/main.go

build:
	@echo "Building binaries..."
	@go build -o bin/$(BINARY_API_NAME) ./cmd/api
	@go build -o bin/$(BINARY_DNS_NAME) ./cmd/dns
	@echo "Build complete."

test:
	@echo "Running tests..."
	@go test -v ./...

bench:
	@echo "Running benchmarks..."
	@go test -bench=. ./...

clean:
	@echo "Cleaning up..."
	@go clean
	@rm -f bin/$(BINARY_API_NAME)
	@rm -f bin/$(BINARY_DNS_NAME)
	@rm -rf docs/

setup:
	@echo "Installing swag-cli"
	@go install github.com/swaggo/swag/cmd/swag@latest

swag:
	@echo "Generating Swagger documentation..."
	@swag init -g cmd/api/main.go

help:
	@echo "Available commands:"
	@echo "  run    - Run the API and DNS servers"
	@echo "  build  - Build the project binaries"
	@echo "  test   - Run all tests"
	@echo "  bench  - Run all benchmarks"
	@echo "  clean  - Clean up build artifacts"
	@echo "  swag   - Generate Swagger documentation"

