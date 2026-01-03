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
