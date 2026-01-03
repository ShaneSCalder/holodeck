#!/bin/bash

# Holodeck - Mock Broker API Module Setup
# Standalone Go Module for multi-instrument backtesting
# Based on: Holodeck Canonical Specification (Multi-Instrument)

set -e

MODULE_NAME="holodeck"
MODULE_PATH="github.com/yourusername/holodeck"

echo "=============================================="
echo "Holodeck Module Setup"
echo "=============================================="
echo ""
echo "Creating standalone Holodeck module..."
echo ""

# Create module directory
mkdir -p "$MODULE_NAME"
cd "$MODULE_NAME"

# ==================== CORE FILES ====================

echo "Setting up core files..."

touch holodeck.go
touch types.go
touch config.go

# ==================== INSTRUMENT PACKAGE ====================

echo "Setting up instrument package..."

mkdir -p instrument
touch instrument/instrument.go
touch instrument/forex.go
touch instrument/stocks.go
touch instrument/commodities.go
touch instrument/crypto.go
touch instrument/base.go

# ==================== READER PACKAGE ====================

echo "Setting up CSV reader..."

mkdir -p reader
touch reader/csv.go
touch reader/parser.go

# ==================== EXECUTOR PACKAGE ====================

echo "Setting up order executor..."

mkdir -p executor
touch executor/executor.go
touch executor/market_order.go
touch executor/limit_order.go
touch executor/partial_fill.go
touch executor/errors.go
touch executor/validation.go

# ==================== POSITION PACKAGE ====================

echo "Setting up position tracking..."

mkdir -p position
touch position/tracker.go
touch position/pnl.go
touch position/state.go

# ==================== ACCOUNT PACKAGE ====================

echo "Setting up account management..."

mkdir -p account
touch account/balance.go
touch account/leverage.go
touch account/drawdown.go
touch account/manager.go

# ==================== LOGGER PACKAGE ====================

echo "Setting up logging..."

mkdir -p logger
touch logger/logger.go
touch logger/file_logger.go
touch logger/metrics.go
touch logger/trade_logger.go

# ==================== SPEED PACKAGE ====================

echo "Setting up speed controller..."

mkdir -p speed
touch speed/controller.go
touch speed/timer.go

# ==================== TYPES PACKAGE ====================

echo "Setting up data types..."

mkdir -p types
touch types/tick.go
touch types/order.go
touch types/execution.go
touch types/position.go
touch types/balance.go
touch types/errors.go
touch types/instrument.go
touch types/constants.go

# ==================== COMMISSION PACKAGE ====================

echo "Setting up commission calculator..."

mkdir -p commission
touch commission/calculator.go
touch commission/forex.go
touch commission/stocks.go
touch commission/commodities.go
touch commission/crypto.go

# ==================== SLIPPAGE PACKAGE ====================

echo "Setting up slippage calculator..."

mkdir -p slippage
touch slippage/calculator.go
touch slippage/depth_model.go
touch slippage/momentum_model.go

# ==================== TESTS ====================

echo "Setting up test suite..."

mkdir -p tests/unit
mkdir -p tests/integration
mkdir -p tests/fixtures

touch tests/unit/instrument_test.go
touch tests/unit/executor_test.go
touch tests/unit/pnl_test.go
touch tests/unit/commission_test.go
touch tests/unit/slippage_test.go
touch tests/unit/leverage_test.go
touch tests/unit/drawdown_test.go
touch tests/unit/csv_reader_test.go

touch tests/integration/forex_test.go
touch tests/integration/stocks_test.go
touch tests/integration/commodities_test.go
touch tests/integration/crypto_test.go
touch tests/integration/full_session_test.go
touch tests/integration/order_flow_test.go

touch tests/fixtures/sample_forex_ticks.csv
touch tests/fixtures/sample_stocks_ticks.csv
touch tests/fixtures/sample_commodities_ticks.csv

# ==================== CMD PACKAGE ====================

echo "Setting up command-line application..."

mkdir -p cmd/holodeck
touch cmd/holodeck/main.go
touch cmd/holodeck/runner.go

mkdir -p cmd/backtest
touch cmd/backtest/main.go
touch cmd/backtest/processor.go

# ==================== CONFIG DIRECTORY ====================

echo "Setting up configuration directory..."

mkdir -p config

# Forex config
cat > config/forex_eurusd.json << 'EOF'
{
  "csv": {
    "filepath": "data/forex_eurusd_ticks.csv"
  },
  "instrument": {
    "type": "FOREX",
    "symbol": "EURUSD",
    "description": "Euro vs US Dollar",
    "decimal_places": 4,
    "pip_value": 0.0001,
    "contract_size": 100000,
    "minimum_lot_size": 0.01,
    "tick_size": 0.00001
  },
  "account": {
    "initial_balance": 100000.00,
    "currency": "USD",
    "leverage": 1.0,
    "max_position_size": 10.0,
    "max_drawdown_percent": 20.0
  },
  "execution": {
    "slippage": true,
    "slippage_model": "depth",
    "latency": true,
    "latency_ms": 5,
    "commission": true,
    "commission_type": "per_million",
    "commission_value": 25,
    "partial_fills": true,
    "partial_fill_based_on": "volume_momentum"
  },
  "order_types": {
    "supported": ["MARKET", "LIMIT"],
    "default": "MARKET"
  },
  "speed": {
    "multiplier": 100
  },
  "session": {
    "close_positions_at_end": false
  },
  "logging": {
    "verbose": true,
    "log_file": "logs/holodeck.log",
    "log_every_tick": false,
    "log_every_trade": true,
    "log_metrics": true
  }
}
EOF

# Stocks config
cat > config/stocks_aapl.json << 'EOF'
{
  "csv": {
    "filepath": "data/stocks_aapl_ticks.csv"
  },
  "instrument": {
    "type": "STOCKS",
    "symbol": "AAPL",
    "description": "Apple Inc.",
    "decimal_places": 2,
    "pip_value": 0.01,
    "contract_size": 1,
    "minimum_lot_size": 1.0,
    "tick_size": 0.01
  },
  "account": {
    "initial_balance": 100000.00,
    "currency": "USD",
    "leverage": 1.0,
    "max_position_size": 1000.0,
    "max_drawdown_percent": 20.0
  },
  "execution": {
    "slippage": true,
    "slippage_model": "depth",
    "latency": true,
    "latency_ms": 10,
    "commission": true,
    "commission_type": "per_share",
    "commission_value": 0.01,
    "partial_fills": true,
    "partial_fill_based_on": "volume_momentum"
  },
  "order_types": {
    "supported": ["MARKET", "LIMIT"],
    "default": "MARKET"
  },
  "speed": {
    "multiplier": 100
  },
  "session": {
    "close_positions_at_end": false
  },
  "logging": {
    "verbose": true,
    "log_file": "logs/holodeck.log",
    "log_every_tick": false,
    "log_every_trade": true,
    "log_metrics": true
  }
}
EOF

# Commodities config
cat > config/commodities_gold.json << 'EOF'
{
  "csv": {
    "filepath": "data/commodities_gold_ticks.csv"
  },
  "instrument": {
    "type": "COMMODITIES",
    "symbol": "GOLD",
    "description": "Gold per troy ounce",
    "decimal_places": 2,
    "pip_value": 0.01,
    "contract_size": 1,
    "minimum_lot_size": 0.1,
    "tick_size": 0.01
  },
  "account": {
    "initial_balance": 100000.00,
    "currency": "USD",
    "leverage": 1.0,
    "max_position_size": 100.0,
    "max_drawdown_percent": 20.0
  },
  "execution": {
    "slippage": true,
    "slippage_model": "depth",
    "latency": true,
    "latency_ms": 5,
    "commission": true,
    "commission_type": "per_lot",
    "commission_value": 5.00,
    "partial_fills": true,
    "partial_fill_based_on": "volume_momentum"
  },
  "order_types": {
    "supported": ["MARKET", "LIMIT"],
    "default": "MARKET"
  },
  "speed": {
    "multiplier": 100
  },
  "session": {
    "close_positions_at_end": false
  },
  "logging": {
    "verbose": true,
    "log_file": "logs/holodeck.log",
    "log_every_tick": false,
    "log_every_trade": true,
    "log_metrics": true
  }
}
EOF

# Crypto config
cat > config/crypto_btc.json << 'EOF'
{
  "csv": {
    "filepath": "data/crypto_btc_ticks.csv"
  },
  "instrument": {
    "type": "CRYPTO",
    "symbol": "BTC/USD",
    "description": "Bitcoin vs US Dollar",
    "decimal_places": 2,
    "pip_value": 0.01,
    "contract_size": 1,
    "minimum_lot_size": 0.001,
    "tick_size": 1.00
  },
  "account": {
    "initial_balance": 100000.00,
    "currency": "USD",
    "leverage": 1.0,
    "max_position_size": 10.0,
    "max_drawdown_percent": 20.0
  },
  "execution": {
    "slippage": true,
    "slippage_model": "depth",
    "latency": true,
    "latency_ms": 100,
    "commission": true,
    "commission_type": "percentage",
    "commission_value": 0.002,
    "partial_fills": true,
    "partial_fill_based_on": "volume_momentum"
  },
  "order_types": {
    "supported": ["MARKET", "LIMIT"],
    "default": "MARKET"
  },
  "speed": {
    "multiplier": 100
  },
  "session": {
    "close_positions_at_end": false
  },
  "logging": {
    "verbose": true,
    "log_file": "logs/holodeck.log",
    "log_every_tick": false,
    "log_every_trade": true,
    "log_metrics": true
  }
}
EOF

# ==================== DATA DIRECTORY ====================

echo "Setting up data directories..."

mkdir -p data/ticks
mkdir -p data/results
mkdir -p logs
mkdir -p reports

# Create .gitkeep files
touch data/ticks/.gitkeep
touch logs/.gitkeep
touch reports/.gitkeep

# ==================== SCRIPTS ====================

echo "Setting up utility scripts..."

mkdir -p scripts

cat > scripts/run_forex_backtest.sh << 'EOF'
#!/bin/bash
go run ./cmd/backtest/main.go -config config/forex_eurusd.json
EOF

cat > scripts/run_stocks_backtest.sh << 'EOF'
#!/bin/bash
go run ./cmd/backtest/main.go -config config/stocks_aapl.json
EOF

cat > scripts/run_commodities_backtest.sh << 'EOF'
#!/bin/bash
go run ./cmd/backtest/main.go -config config/commodities_gold.json
EOF

cat > scripts/run_crypto_backtest.sh << 'EOF'
#!/bin/bash
go run ./cmd/backtest/main.go -config config/crypto_btc.json
EOF

cat > scripts/test_all.sh << 'EOF'
#!/bin/bash
echo "Running all tests..."
go test -v ./tests/unit/...
go test -v ./tests/integration/...
EOF

cat > scripts/benchmark.sh << 'EOF'
#!/bin/bash
echo "Running benchmarks..."
go test -bench=. -benchmem ./...
EOF

chmod +x scripts/*.sh

# ==================== MAKEFILE ====================

echo "Creating Makefile..."

cat > Makefile << 'EOF'
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
EOF

# ==================== GITIGNORE ====================

cat > .gitignore << 'EOF'
# Binaries
bin/
*.exe
*.dll
*.so
*.dylib

# Build artifacts
*.o
*.a

# Test binaries
*.test

# Output
*.out
*.log

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Go
vendor/
go.sum

# Logs
logs/*
!logs/.gitkeep

# Data
data/*
!data/ticks/.gitkeep
!data/results/.gitkeep

# Reports
reports/*
!reports/.gitkeep

# Config (local overrides)
config/local*.json
EOF

# ==================== GO MOD ====================

echo "Creating go.mod placeholder..."

cat > go.mod << 'EOF'
module github.com/yourusername/holodeck

go 1.21

require (
	// Add dependencies here as needed
)
EOF

touch go.sum

# ==================== ROOT FILES ====================

echo "Creating root documentation files..."

cat > README.md << 'EOF'
# Holodeck - Mock Broker API

Holodeck is a standalone Go module that provides a mock broker API for backtesting trading strategies.

## Features

- **Multi-Instrument Support**: FOREX, STOCKS, COMMODITIES, CRYPTO
- **CSV Data Input**: Load simulated tick data from files
- **Realistic Execution**: Slippage, latency, commission
- **Position Management**: Single position tracking with leverage
- **P&L Calculation**: Accurate profit/loss calculation per instrument
- **Speed Control**: Run simulations at 1x to 1000x speed
- **Comprehensive Logging**: Trade logs, metrics, error handling
- **Error Handling**: Full validation and error reporting
- **Configurable**: JSON-based configuration for all parameters

## Quick Start

```bash
# Install dependencies
go mod download

# Build
make build

# Run Forex backtest
make run-forex

# Run all tests
make test
```

## Configuration

See `config/` directory for example configurations for each instrument type:
- `forex_eurusd.json` - Forex trading
- `stocks_aapl.json` - Stock trading
- `commodities_gold.json` - Commodity trading
- `crypto_btc.json` - Cryptocurrency trading

## Project Structure

```
holodeck/
├── instrument/         # Instrument implementations
├── executor/          # Order execution logic
├── position/          # Position tracking
├── account/           # Account management
├── logger/            # Logging
├── commission/        # Commission calculation
├── slippage/          # Slippage modeling
├── reader/            # CSV reader
├── speed/             # Speed control
├── types/             # Data structures
├── cmd/               # Command-line tools
├── tests/             # Test suite
├── config/            # Configuration files
└── scripts/           # Utility scripts
```

## Testing

```bash
# Unit tests
make test-unit

# Integration tests
make test-integration

# All tests
make test

# Benchmarks
make benchmark
```

## Documentation

See `ARCHITECTURE.md` for detailed architecture information.
EOF

cat > ARCHITECTURE.md << 'EOF'
# Holodeck Architecture

## Overview

Holodeck is a modular mock broker API designed to:
1. Read simulated tick data from CSV files
2. Execute orders with realistic friction
3. Track position and P&L
4. Support multiple instrument types
5. Provide the same interface as real brokers

## Module Organization

### Core Modules

- **instrument/**: Instrument type implementations (FOREX, STOCKS, COMMODITIES, CRYPTO)
- **executor/**: Order execution engine with validation and partial fills
- **position/**: Position state management and tracking
- **account/**: Account balance, leverage, and drawdown management
- **logger/**: Comprehensive logging and metrics

### Calculation Modules

- **commission/**: Commission calculation per instrument type
- **slippage/**: Slippage modeling (depth-based, momentum-based)

### Data Modules

- **reader/**: CSV tick data reader
- **types/**: All data structures (Order, Tick, Position, Balance, etc.)

### Utility Modules

- **speed/**: Speed control for accelerated backtesting
- **cmd/**: Command-line applications

## Data Flow

```
CSV File
  ↓
Reader (reads ticks)
  ↓
GetNextTick()
  ↓
Aggregator (builds candles)
  ↓
Agent (decides)
  ↓
ExecuteOrder()
  ↓
Executor (validates, applies friction)
  ↓
Position Tracker (updates state)
  ↓
P&L Calculator (updates P&L)
  ↓
Logger (logs trade)
  ↓
Return ExecutionReport
```

## Configuration

All behavior is controlled via JSON configuration files:
- Instrument type and parameters
- Account settings (balance, leverage, limits)
- Execution settings (slippage, latency, commission)
- Order types and defaults
- Speed multiplier
- Logging preferences

## Error Handling

Full error handling for:
- Invalid orders (size, price, type)
- Insufficient balance
- Position limits exceeded
- Account blown (drawdown exceeded)
- CSV reading errors
- Configuration errors

## Testing Strategy

- Unit tests for each module
- Integration tests for full workflows
- Fixture data for consistent testing
- Benchmarks for performance validation

See tests/ directory for details.
EOF

touch DEVELOPMENT.md

# ==================== FINAL SUMMARY ====================

echo ""
echo "=========================================="
echo "Holodeck module structure created!"
echo "=========================================="
echo ""
echo "Project structure:"
echo ""
tree -L 2 2>/dev/null || find . -type d -not -path '*/\.*' | head -30
echo ""
echo "Next steps:"
echo "1. cd $MODULE_NAME"
echo "2. go mod init $MODULE_PATH"
echo "3. go mod tidy"
echo "4. make build"
echo "5. make test"
echo "6. make run-forex  (or any other instrument)"
echo ""
echo "Configuration files available:"
echo "  - config/forex_eurusd.json"
echo "  - config/stocks_aapl.json"
echo "  - config/commodities_gold.json"
echo "  - config/crypto_btc.json"
echo ""
echo "Run tests:"
echo "  - make test-unit"
echo "  - make test-integration"
echo "  - make benchmark"
echo ""
echo "For more info:"
echo "  - make help"
echo "  - cat README.md"
echo "  - cat ARCHITECTURE.md"
echo ""