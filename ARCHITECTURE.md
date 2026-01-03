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
