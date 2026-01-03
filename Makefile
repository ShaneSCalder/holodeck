.PHONY: help build test clean run-forex run-stocks run-commodities run-crypto fmt lint

help:
	@echo "Holodeck Module - Available Commands"
	@echo "====================================="
	@echo "make build                    - Build all binaries"
	@echo "make test                     - Run all tests"
	@echo "make test-unit                - Run unit tests only"
	@echo "make test-integration         - Run integration tests only"
	@echo "make clean                    - Clean build artifacts"
	@echo "make run-forex                - Run Forex backtest"
	@echo "make run-stocks               - Run Stocks backtest"
	@echo "make run-commodities          - Run Commodities backtest"
	@echo "make run-crypto               - Run Crypto backtest"
	@echo "make benchmark                - Run performance benchmarks"
	@echo "make fmt                      - Format code"
	@echo "make lint                     - Lint code"
	@echo "make help                     - Show this help message"

build:
	@echo "Building Holodeck..."
	go build -o bin/holodeck ./cmd/holodeck
	go build -o bin/backtest ./cmd/backtest
	@echo "Build complete!"

test:
	@echo "Running all tests..."
	go test -v ./...

test-unit:
	@echo "Running unit tests..."
	go test -v ./tests/unit/...

test-integration:
	@echo "Running integration tests..."
	go test -v ./tests/integration/...

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

run-forex:
	@echo "Running Forex backtest..."
	go run ./cmd/backtest/main.go -config config/forex_eurusd.json

run-stocks:
	@echo "Running Stocks backtest..."
	go run ./cmd/backtest/main.go -config config/stocks_aapl.json

run-commodities:
	@echo "Running Commodities backtest..."
	go run ./cmd/backtest/main.go -config config/commodities_gold.json

run-crypto:
	@echo "Running Crypto backtest..."
	go run ./cmd/backtest/main.go -config config/crypto_btc.json

benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

fmt:
	@echo "Formatting code..."
	go fmt ./...

lint:
	@echo "Linting code..."
	golangci-lint run ./...

.PHONY: all
all: clean build test
