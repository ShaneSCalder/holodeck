# Types Package Documentation

The `types` package contains all fundamental data structures used throughout Holodeck. These files define the core concepts: instruments, market data, orders, executions, positions, accounts, and error handling.

---

## Overview

```
types/
├── constants.go      # Enums, constants, utility functions
├── tick.go           # Market price data structure
├── order.go          # Trading order structure
├── execution.go      # Order execution results
├── position.go       # Open position tracking
├── balance.go        # Account equity tracking
├── errors.go         # Error types and handling
└── instrument.go     # Instrument definitions (FOREX, STOCKS, etc)
```

---

## File Descriptions

### 1. constants.go

**Purpose:** Central definition of all constants, enums, and utility functions.

**Key Constants:**
- **Instrument Types:** FOREX, STOCKS, COMMODITIES, CRYPTO
- **Order Actions:** BUY, SELL, HOLD
- **Order Types:** MARKET, LIMIT
- **Order Status:** FILLED, PARTIAL, REJECTED, PENDING, CANCELLED
- **Account Status:** ACTIVE, BLOWN, AT_LIMIT
- **Position Status:** FLAT, LONG, SHORT
- **Error Codes:** 13 standardized error codes
- **Commission Types:** per_million, per_share, per_lot, percentage
- **Slippage Models:** depth, momentum, fixed, none
- **Momentum Levels:** STRONG, NORMAL, WEAK

**Instrument-Specific Defaults:**
- **FOREX:** ContractSize=100k, PipValue=0.0001, MinLot=0.01, Commission=$25/M
- **STOCKS:** ContractSize=1, PipValue=0.01, MinLot=1.0, Commission=$0.01/share
- **COMMODITIES:** ContractSize=1, PipValue=0.01, MinLot=0.1, Commission=$5/lot
- **CRYPTO:** ContractSize=1, PipValue=0.01, MinLot=0.001, Commission=0.2%

**Utility Functions:**
- `IsValidInstrumentType()` - Validates instrument type
- `IsValidOrderAction()` - Validates order action
- `GetPositionStatusFromSize()` - Derives status from position size
- `GetInstrumentDefaults()` - Returns defaults for instrument type
- `GetMomentumMultiplier()` - Gets fill multiplier from momentum
- `GetVolumeMultiplier()` - Gets fill multiplier from volume

---

### 2. tick.go

**Purpose:** Represents a single market data point (price quote).

**Main Struct: Tick**
```go
type Tick struct {
  Timestamp   time.Time  // When this price occurred
  Bid         float64    // Price we can SELL at
  Ask         float64    // Price we can BUY at
  BidQty      int64      // Volume at bid
  AskQty      int64      // Volume at ask
  LastPrice   float64    // Last traded price
  Volume      int64      // Tick volume
  Sequence    int64      // Monotonic counter
  MidPrice    float64    // Calculated (Bid+Ask)/2
  SpreadPips  float64    // Calculated Ask-Bid
}
```

**Key Methods:**
- `GetMidPrice()` - Mid price
- `GetSpread()` - Bid-ask spread
- `GetBuyPrice()` - Ask price (for BUY orders)
- `GetSellPrice()` - Bid price (for SELL orders)
- `GetAvailableDepth()` - Min of bid/ask qty (for slippage calc)
- `IsValid()` - Validates tick integrity

**Bonus Structures:**
- `TickBuffer` - Circular buffer for rolling window of ticks
- `TickStats` - OHLC and statistics calculated from tick set

---

### 3. order.go

**Purpose:** Represents a trading order from Agent to Holodeck.

**Main Struct: Order**
```go
type Order struct {
  Action      string      // BUY, SELL, HOLD
  Size        float64     // Quantity (lots, shares, oz, etc)
  OrderType   string      // MARKET or LIMIT
  LimitPrice  float64     // For LIMIT orders only
  Timestamp   time.Time   // When order placed
  OrderID     string      // Unique identifier
  Description string      // Human note
}
```

**Quick Constructors:**
- `NewMarketOrder(action, size, timestamp)` - MARKET order
- `NewLimitOrder(action, size, limit, timestamp)` - LIMIT order
- `NewBuyOrder(size, timestamp)` - BUY MARKET shortcut
- `NewSellOrder(size, timestamp)` - SELL MARKET shortcut
- `NewHoldOrder(timestamp)` - Do nothing

**Key Methods:**
- `IsBuy() / IsSell() / IsHold()` - Check action
- `IsMarket() / IsLimit()` - Check type
- `IsTradeOrder()` - Is BUY or SELL?
- `GetDirection()` - Returns 1/-1/0
- `Validate(minLotSize, maxPositionSize)` - Validate order

**Bonus Features:**
- `OrderBuilder` - Fluent builder pattern
- `OrderBatch` - Group multiple orders
- Full validation with structured error reporting

---

### 4. execution.go

**Purpose:** Represents the result of executing an order.

**Main Struct: ExecutionReport**
```go
type ExecutionReport struct {
  OrderID         string      // Order identifier
  Timestamp       time.Time   // Execution time
  Action          string      // BUY or SELL
  RequestedSize   float64     // What was asked for
  FilledSize      float64     // What was actually filled
  FillPrice       float64     // Price per unit (with slippage)
  SlippageUnits   float64     // Slippage in pips/cents/dollars
  Commission      float64     // Trading fee
  PositionAfter   float64     // Position size after (+ = long, - = short)
  EntryPrice      float64     // Entry price for open position
  UnrealizedPnL   float64     // Mark-to-market P&L
  RealizedPnL     float64     // Closed trade P&L
  TotalPnL        float64     // Realized + Unrealized - Commission
  Status          string      // FILLED, PARTIAL, REJECTED
  ErrorCode       string      // If rejected
  ErrorMessage    string      // Error details
  Latency         int64       // Execution delay (ms)
  AvailableDepth  int64       // Volume available at exec
}
```

**Constructors:**
- `NewExecutionReport()` - Successful fill
- `NewPartialExecution()` - Partial fill
- `NewRejectedExecution()` - Rejected order

**Query Methods:**
- `IsFilled() / IsPartial() / IsRejected()` - Check status
- `WasExecuted()` - Filled or partial?
- `GetFillPercentage()` - 0-100%
- `GetUnfilledSize()` - Remaining quantity
- `GetNotional()` - FilledSize × FillPrice
- `GetPositionStatus()` - LONG/SHORT/FLAT after

**Bonus Structures:**
- `ExecutionStats` - Statistics from execution set
- `ExecutionBatch` - Group multiple executions

---

### 5. position.go

**Purpose:** Tracks the current open trading position.

**Main Struct: Position**
```go
type Position struct {
  Size                    float64       // Current size (+ long, - short, 0 flat)
  EntryPrice              float64       // Average entry price
  EntryTime               time.Time     // When opened
  EntryCommission         float64       // Commission on entry
  CurrentPrice            float64       // Latest market price
  RealizedPnL             float64       // From closed trades
  UnrealizedPnL           float64       // Mark-to-market
  CommissionPaid          float64       // Total commissions
  TradeHistory            []*Trade      // All trades
  MaxFavorableExcursion   float64       // Best P&L reached
  MaxAdverseExcursion     float64       // Worst P&L reached
}
```

**Status Queries:**
- `GetStatus()` - LONG/SHORT/FLAT
- `IsFlat() / IsLong() / IsShort()`
- `GetAbsoluteSize()` - |Size|
- `GetDirection()` - 1/-1/0

**P&L Calculations:**
- `CalculateUnrealizedPnL(price, pipValue)` - Mark-to-market
- `CalculateTotalPnL()` - Realized + Unrealized - Commission
- `CalculateROE()` - Return on Equity %
- `GetBreakevenPrice()` - Price to break even
- `GetNotional()` - Size × CurrentPrice

**Bonus Structures:**
- `Trade` - Individual trade record
- `PositionHistory` - Snapshot tracking over time
- `PositionSnapshot` - Point-in-time state

---

### 6. balance.go

**Purpose:** Tracks account equity and performance metrics.

**Main Struct: Balance**
```go
type Balance struct {
  InitialBalance          float64       // Starting money
  CurrentBalance          float64       // Current equity
  Currency                string        // USD, EUR, etc
  TotalRealizedPnL        float64       // From closed trades
  TotalUnrealizedPnL      float64       // From open positions
  CommissionPaid          float64       // Total fees
  Leverage                float64       // 1.0 = no leverage
  UsedMargin              float64       // Margin in use
  AvailableMargin         float64       // Margin available
  BuyingPower             float64       // Balance × Leverage
  MaxDrawdownPercent      float64       // Limit before blown
  AccountStatus           string        // ACTIVE/BLOWN/AT_LIMIT
  TradeCount              int           // Total trades
  WinningTrades           int
  LosingTrades            int
  HighWaterMark           float64       // Peak balance
  LowWaterMark            float64       // Lowest balance
}
```

**Key Query Methods:**
- `GetTotalPnL()` - Realized + Unrealized
- `GetNetPnL()` - Total P&L minus commissions
- `GetDrawdownPercent()` - Current drawdown %
- `GetReturnPercent()` - Return %
- `GetWinRate()` - Win % of trades
- `GetAverageTradePnL()` - Avg P&L per trade
- `GetProfitFactor()` - Gross profits / losses
- `GetSharpeRatio()` - Risk-adjusted return

**Status Checks:**
- `IsAccountActive()` - Can trade?
- `IsAccountBlown()` - Account blown?
- `IsAccountAtLimit()` - Near limit?
- `IsMarginCall()` - Margin violated?
- `CanTrade()` - Active and has margin?

**Update Methods:**
- `UpdateFromExecution()` - Auto-update from execution
- `UpdateMargin()` - Recalculate margin
- `RecalculateBalance()` - Recalc balance and status

**Bonus Features:**
- `BalanceUpdate` - Records individual updates
- Comprehensive metrics via `GetMetrics()`

---

### 7. errors.go

**Purpose:** Structured error handling with detailed context.

**Main Struct: HolodeckError**
```go
type HolodeckError struct {
  Code        string                 // Error code
  Message     string                 // Human-readable message
  Details     map[string]interface{} // Additional context
  Timestamp   time.Time              // When error occurred
  SourceFunc  string                 // Function that errored
  SourceFile  string                 // File name
  SourceLine  int                    // Line number
  ParentError error                  // Wrapped error
}
```

**Error Constructors (13 types):**
- `NewInsufficientBalanceError(required, available)`
- `NewPositionLimitError(requested, maxAllowed)`
- `NewInvalidOrderTypeError(orderType)`
- `NewInvalidLimitPriceError(price, reason)`
- `NewInvalidOrderSizeError(size, minSize)`
- `NewOrderRejectedError(reason)`
- `NewAccountBlownError(drawdown, maxDrawdown)`
- `NewInvalidOperationError(operation, reason)`
- `NewCSVReadError(filename, lineNumber, reason)`
- `NewConfigError(field, reason)`
- `NewInstrumentNotFoundError(instrumentType)`
- `NewInvalidInstrumentTypeError(instrumentType)`
- Plus generic `NewHolodeckError(code, message)`

**Error Methods:**
- `WithDetail(key, value)` - Add detail
- `WithDetails(map)` - Add multiple details
- `WithParent(error)` - Wrap another error
- `WithSource(func, file, line)` - Set source location

**Error Type Checks:**
- `IsInsufficientBalance()`
- `IsPositionLimitExceeded()`
- `IsAccountBlown()`
- `IsCritical()`
- `IsRetryable()`

**Bonus Features:**
- `ErrorLog` - Collect multiple errors
- `ErrorBuilder` - Fluent error building
- `ErrorSummary` - Statistics on error collection

---

### 8. instrument.go

**Purpose:** Defines how different instruments work (FOREX, STOCKS, COMMODITIES, CRYPTO).

**Main Interface: Instrument**
```go
type Instrument interface {
  // Basic info
  GetType() string
  GetSymbol() string
  GetDescription() string
  GetDecimalPlaces() int
  GetPipValue() float64
  GetContractSize() int64
  GetMinimumLotSize() float64
  GetTickSize() float64
  
  // Core calculations (differ per instrument)
  CalculatePnL(entryPrice, exitPrice, size, direction) float64
  CalculateCommission(price, size, side) float64
  CalculateSlippage(size, availableDepth, momentum) float64
  
  // Validation
  ValidateOrderSize(size) error
  ValidateLimitPrice(limitPrice, currentPrice, action) error
  
  // Formatting
  FormatPrice(price) string
  GetConfig() *InstrumentConfig
}
```

**Four Implementations:**

**1. ForexInstrument (EURUSD, GBPUSD, etc)**
- Pips: 0.0001
- Contract size: 100,000
- Commission: $25 per $1M notional
- P&L: pips × size × contract_size × pip_value

**2. StocksInstrument (AAPL, TSLA, SPY)**
- Decimal: 0.01 (cents)
- Contract size: 1
- Commission: $0.01 per share
- P&L: (exit - entry) × shares

**3. CommoditiesInstrument (GOLD, OIL, COPPER)**
- Decimal: 0.01
- Contract size: 1
- Commission: $5.00 per lot
- P&L: (exit - entry) × quantity

**4. CryptoInstrument (BTC/USD, ETH/USD)**
- Decimal: 0.01
- Contract size: 1
- Commission: 0.2% of notional
- P&L: (exit - entry) × quantity

**Factory & Registry:**
- `NewInstrument(type, symbol, description)` - Factory function
- `InstrumentRegistry` - Manage multiple instruments
- `CompareInstruments(a, b)` - Compare two instruments

---

## Data Flow Through Types

### Order Execution Flow

```
Order (types/order.go)
  ↓
ExecuteOrder() in Executor
  ↓
Instrument (types/instrument.go)
  ├─ CalculatePnL()
  ├─ CalculateCommission()
  ├─ CalculateSlippage()
  └─ ValidateOrderSize()
  ↓
ExecutionReport (types/execution.go)
  ↓
Position (types/position.go)
  └─ UpdatePrice()
  └─ CalculateUnrealizedPnL()
  ↓
Balance (types/balance.go)
  └─ UpdateFromExecution()
  └─ RecalculateBalance()
  ↓
Error (types/errors.go) [if validation fails]
```

### Error Flow

```
Any Operation
  ↓
Validation Fails
  ↓
HolodeckError (types/errors.go)
  ├─ Code: "ERROR_CODE"
  ├─ Message: "Human readable"
  ├─ Details: {context map}
  └─ SourceFunc/File/Line: Location info
  ↓
ErrorLog (types/errors.go)
  └─ Collect multiple errors
  └─ Generate ErrorSummary
```

---

## Usage Examples

### Creating an Order
```go
// Quick way
order := NewBuyOrder(0.01, time.Now())

// With builder
order, err := NewOrderBuilder().
  Buy().
  WithSize(0.01).
  WithMarketOrder().
  WithDescription("Entry signal").
  Build()
```

### Executing an Order
```go
// In executor
exec := NewExecutionReport(
  "ORD_001",
  time.Now(),
  OrderActionBuy,
  0.01,           // requested
  0.01,           // filled
  1.08505,        // fill price
  0.0002,         // slippage
  2.71,           // commission
  0.01,           // position after
  1.08505,        // entry price
  0,              // unrealized
  0,              // realized
  -2.71,          // total (negative due to commission)
)
```

### Tracking Position
```go
pos := NewPosition()
pos.UpdatePrice(1.08550, 0.0001)  // Update to current price

unrealized := pos.CalculateUnrealizedPnL(1.08550, 0.0001)
total := pos.CalculateTotalPnL()
roe := pos.CalculateROE()
```

### Managing Account
```go
balance := NewBalance(100000, "USD", 1.0, 20.0, 10.0)

// After execution
balance.UpdateFromExecution(executionReport)
balance.UpdateMargin(usedMargin)

// Check status
if balance.IsAccountBlown() {
  // No more trading allowed
}
```

### Creating Instruments
```go
forex := NewForexInstrument("EURUSD", "Euro vs US Dollar")
pnl := forex.CalculatePnL(1.08500, 1.08600, 0.01, 1)  // BUY

stocks := NewStocksInstrument("AAPL", "Apple Inc")
comm := stocks.CalculateCommission(150.25, 100, "BUY")  // $0.01 * 100
```

### Error Handling
```go
err := NewInsufficientBalanceError(108500, 50000)
err.WithDetail("order_id", "ORD_001")
err.WithSource("ExecuteOrder", "executor.go", 45)

if herr, ok := AsHolodeckError(err); ok {
  if herr.IsInsufficientBalance() {
    // Handle insufficient balance
  }
}
```

---

## Summary

The **types package** provides:

1. **constants.go** - Centralized enums and utilities
2. **tick.go** - Market price data
3. **order.go** - Trading orders with validation
4. **execution.go** - Order execution results
5. **position.go** - Position tracking and P&L
6. **balance.go** - Account equity and performance
7. **errors.go** - Structured error handling
8. **instrument.go** - Multi-instrument support

**Together**, these files form the foundation for all Holodeck functionality. They're instrument-agnostic, fully validated, and ready for production use.