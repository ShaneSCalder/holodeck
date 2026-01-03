# Holodeck Core Architecture: Config, Types & Holodeck

This document provides comprehensive documentation for the three core files that form the foundation of Holodeck's architecture:
- **config.go** - Configuration loading and management
- **types.go** - State management and data structures
- **holodeck.go** - Main API orchestrator

---

## Table of Contents

1. [Overview](#overview)
2. [config.go - Configuration System](#configgo---configuration-system)
3. [types.go - State Management](#typesgo---state-management)
4. [holodeck.go - Main API](#holodeckgo---main-api)
5. [Integration Flow](#integration-flow)
6. [Usage Examples](#usage-examples)
7. [Best Practices](#best-practices)

---

## Overview

### Architecture Layers

```
┌─────────────────────────────────────────────────┐
│            Holodeck (Main API)                  │
│  Orchestrates execution, state, and subsystems  │
├─────────────────────────────────────────────────┤
│  HolodeckState (Mutable State)                  │
│  - Thread-safe state management                 │
│  - Position, Balance, Execution tracking        │
├─────────────────────────────────────────────────┤
│  HolodeckConfig (Configuration)                 │
│  - Instrument setup, Account parameters         │
│  - Execution settings, Logging config           │
├─────────────────────────────────────────────────┤
│  Config (JSON Configuration)                    │
│  - Loaded from JSON files                       │
│  - Fully validated                              │
└─────────────────────────────────────────────────┘
```

### Initialization Chain

```
JSON File
    ↓
ConfigLoader.Load()
    ↓
Config (validated)
    ↓
NewHolodeckConfig(config)
    ↓
HolodeckConfig
    ↓
NewHolodeckState(hConfig)
    ↓
HolodeckState
    ↓
NewHolodeck(hConfig)
    ↓
Holodeck (with subsystems)
```

---

## config.go - Configuration System

### Purpose

Loads, validates, and manages JSON configuration files. Provides flexible configuration management with full error reporting.

### Configuration Structure

#### Root Config

```go
type Config struct {
    CSV        CSVConfig        // Data source
    Instrument InstrumentConfig // Instrument type & params
    Account    AccountConfig    // Account settings
    Execution  ExecutionConfig  // Execution behavior
    OrderTypes OrderTypesConfig // Order type settings
    Speed      SpeedConfig      // Simulation speed
    Session    SessionConfig    // Session behavior
    Logging    LoggingConfig    // Logging setup
}
```

#### Section 1: CSV Configuration

```go
type CSVConfig struct {
    FilePath string  // Path to tick data file
}
```

**Example:**
```json
{
  "csv": {
    "filepath": "data/forex_eurusd_ticks.csv"
  }
}
```

#### Section 2: Instrument Configuration

```go
type InstrumentConfig struct {
    Type            string  // FOREX, STOCKS, COMMODITIES, CRYPTO
    Symbol          string  // EURUSD, AAPL, GOLD, BTC/USD
    Description     string  // Human-readable description
    DecimalPlaces   int     // Decimal precision (2-8)
    PipValue        float64 // Smallest unit (0.0001-1.0)
    ContractSize    int64   // Units per lot
    MinimumLotSize  float64 // Minimum tradeable size
    TickSize        float64 // Price increment
}
```

**Examples:**

| Field | FOREX | STOCKS | COMMODITIES | CRYPTO |
|-------|-------|--------|-------------|--------|
| Type | FOREX | STOCKS | COMMODITIES | CRYPTO |
| Symbol | EURUSD | AAPL | GOLD | BTC/USD |
| DecimalPlaces | 4 | 2 | 2 | 2 |
| PipValue | 0.0001 | 0.01 | 0.01 | 0.01 |
| ContractSize | 100000 | 1 | 1 | 1 |
| MinimumLotSize | 0.01 | 1.0 | 0.1 | 0.001 |
| TickSize | 0.00001 | 0.01 | 0.01 | 1.00 |

#### Section 3: Account Configuration

```go
type AccountConfig struct {
    InitialBalance     float64 // Starting account balance
    Currency           string  // Account currency (USD, EUR, etc)
    Leverage           float64 // Leverage multiplier (>=1.0)
    MaxPositionSize    float64 // Max size per position
    MaxDrawdownPercent float64 // Max allowed drawdown (0-100%)
}
```

**Validation Rules:**
- `InitialBalance` > 0
- `Leverage` >= 1.0
- `MaxPositionSize` > 0
- `MaxDrawdownPercent` > 0 and <= 100

#### Section 4: Execution Configuration

```go
type ExecutionConfig struct {
    // Slippage
    Slippage          bool    // Enable slippage simulation
    SlippageModel     string  // depth, momentum, fixed, none
    
    // Latency
    Latency           bool    // Enable latency simulation
    LatencyMs         int64   // Latency in milliseconds
    
    // Commission
    Commission        bool    // Enable commission
    CommissionType    string  // per_million, per_share, per_lot, percentage
    CommissionValue   float64 // Commission rate
    
    // Partial Fills
    PartialFills      bool    // Enable partial fills
    PartialFillBasedOn string // volume_momentum, depth, none
}
```

**Slippage Models:**
- `depth` - Based on order size vs available depth
- `momentum` - Based on price movement
- `fixed` - Fixed slippage amount
- `none` - No slippage

**Commission Types:**
- `per_million` - Price per $1M notional (Forex: $25)
- `per_share` - Price per share (Stocks: $0.01)
- `per_lot` - Price per lot (Commodities: $5.00)
- `percentage` - Percentage of notional (Crypto: 0.2%)

#### Section 5: Order Types Configuration

```go
type OrderTypesConfig struct {
    Supported []string // [MARKET, LIMIT] or [MARKET] etc
    Default   string   // Default order type (MARKET or LIMIT)
}
```

#### Section 6: Speed Configuration

```go
type SpeedConfig struct {
    Multiplier float64 // Speed multiplier (0.1 to 10000)
}
```

**Examples:**
- `1.0` = Real-time
- `100.0` = 100x speed (1 year of data in ~2.5 minutes)
- `1000.0` = 1000x speed (1 year of data in ~15 seconds)

#### Section 7: Session Configuration

```go
type SessionConfig struct {
    ClosePositionsAtEnd bool // Auto-close positions at end
}
```

#### Section 8: Logging Configuration

```go
type LoggingConfig struct {
    Verbose        bool   // Verbose output
    LogFile        string // Path to log file
    LogEveryTick   bool   // Log every tick (verbose)
    LogEveryTrade  bool   // Log every trade
    LogMetrics     bool   // Log metrics periodically
}
```

### ConfigLoader

#### Loading from File

```go
loader := NewConfigLoader("config/forex_eurusd.json")
if err := loader.Load(); err != nil {
    log.Fatal(err)
}
config := loader.Config
```

#### Loading from String (Testing)

```go
jsonString := `{
    "csv": {"filepath": "data.csv"},
    "instrument": {"type": "FOREX", "symbol": "EURUSD", ...},
    ...
}`

loader := NewConfigLoader("")
if err := loader.LoadFromString(jsonString); err != nil {
    log.Fatal(err)
}
```

### Validation

```go
loader := NewConfigLoader("config/forex_eurusd.json")
if err := loader.Load(); err != nil {  // Load() includes validation
    // Error contains details about what failed
    herr, _ := err.(*types.HolodeckError)
    fmt.Printf("Field: %s\n", herr.Details["field"])
    fmt.Printf("Reason: %s\n", herr.Details["reason"])
}
```

**Validated:**
✓ All required fields present
✓ All values in valid ranges
✓ Enum values valid (FOREX, STOCKS, etc)
✓ CSV file exists
✓ Order types valid and default in list

### ConfigManager (Multiple Configs)

```go
manager := NewConfigManager()

// Load individual configs
manager.LoadConfig("forex", "config/forex_eurusd.json")
manager.LoadConfig("stocks", "config/stocks_aapl.json")

// Or load all from directory
manager.LoadFromDirectory("config/")

// Access configs
config, _ := manager.GetConfig("forex")
default, _ := manager.GetDefault()
manager.SetDefault("stocks")

// List all
names := manager.List()  // ["forex", "stocks"]
```

### Configuration Utilities

```go
config := loader.Config

// Safe getters with defaults
balance := config.GetInitialBalance()          // float64
leverage := config.GetLeverage()               // float64
speed := config.GetSpeedMultiplier()           // float64

// Boolean checks
if config.IsSlippageEnabled() { }
if config.IsCommissionEnabled() { }
if config.ShouldLogEveryTrade() { }

// Export/Save
jsonStr, _ := config.ToJSON()
config.SaveToFile("config/new_config.json")

// Summary
fmt.Println(config.Summary())
fmt.Println(config.DebugString())  // Full JSON
```

---

## types.go - State Management

### Purpose

Manages the mutable state of Holodeck with thread-safe operations. Tracks position, balance, and execution history.

### HolodeckConfig

Bridges configuration and runtime setup.

```go
type HolodeckConfig struct {
    Config          *Config              // Loaded JSON config
    Instrument      types.Instrument     // Active instrument
    SessionID       string               // Unique session ID
    StartTime       time.Time            // Session start
    ExecutionConfig ExecutionParameters  // Parsed execution params
    DataSource      DataSourceConfig     // Data source info
    StateConfig     StateConfiguration   // State tracking limits
    IsRunning       bool                 // Running flag
}
```

#### Creating HolodeckConfig

```go
// From loaded Config
hConfig, err := NewHolodeckConfig(config)
if err != nil {
    log.Fatal(err)
}
```

**What it does:**
1. Creates appropriate Instrument (FOREX, STOCKS, etc)
2. Sets up ExecutionParameters from Config
3. Generates unique SessionID
4. Validates everything

#### ExecutionParameters

```go
type ExecutionParameters struct {
    CommissionEnabled    bool
    CommissionType       string
    CommissionValue      float64
    SlippageEnabled      bool
    SlippageModel        string
    LatencyEnabled       bool
    LatencyMs            int64
    PartialFillsEnabled  bool
    PartialFillLogic     string
    SpeedMultiplier      float64
    DefaultOrderType     string
    SupportedTypes       []string
}
```

### HolodeckState

Core mutable state with thread-safe access via `sync.RWMutex`.

```go
type HolodeckState struct {
    Config *HolodeckConfig
    
    mu sync.RWMutex  // Thread safety
    
    // Tick processing
    CurrentTick *types.Tick
    TickCount   int64
    
    // Position tracking
    Position *types.Position
    
    // Account tracking
    Balance *types.Balance
    
    // Execution history
    ExecutionHistory []*types.ExecutionReport
    ExecutionCount   int
    
    // Error tracking
    ErrorLog *types.ErrorLog
    
    // Metrics
    StartBalance   float64
    CurrentBalance float64
    PeakBalance    float64
    TroughBalance  float64
    TotalPnL       float64
    
    // Timing
    LastUpdateTime time.Time
    SessionStart   time.Time
    SessionEnd     time.Time
}
```

#### Creating HolodeckState

```go
state, err := NewHolodeckState(hConfig)
if err != nil {
    log.Fatal(err)
}
```

### Thread-Safe Read Operations

All use `RLock()` for safe concurrent reading:

```go
tick := state.GetCurrentTick()          // *Tick
position := state.GetPosition()         // *Position
balance := state.GetBalance()           // *Balance
count := state.GetExecutionCount()      // int
ticks := state.GetTickCount()           // int64
pnl := state.GetTotalPnL()              // float64
errors := state.GetErrorCount()         // int
```

### Thread-Safe Write Operations

All use `Lock()` for safe updates:

```go
// Update tick and increment counter
state.UpdateTick(tick)

// Update position
state.UpdatePosition(newPosition)

// Update balance
state.UpdateBalance(newBalance)

// Add execution to history
state.AddExecution(executionReport)

// Add error to log
state.AddError(holodeckError)
```

### Metrics & Status

```go
// Session duration
duration := state.GetSessionDuration()  // time.Duration

// Performance metrics
drawdown := state.GetDrawdownPercent()   // float64 (%)
returnPct := state.GetReturnPercent()    // float64 (%)
maxDD := state.GetMaxDrawdown()          // float64 (%)

// Comprehensive metrics map
metrics := state.GetMetrics()
// Contains: session_id, tick_count, execution_count, 
//           error_count, session_duration, return_percent, etc

// Session status
status := state.GetStatus()  // *SessionStatus
fmt.Println(status.String())
fmt.Println(status.DebugString())
```

### SessionStatus

Snapshot of current session state:

```go
type SessionStatus struct {
    SessionID       string      // HOLO-1735000000000
    InstrumentType  string      // FOREX
    InstrumentSymbol string     // EURUSD
    StartTime       time.Time
    CurrentTime     time.Time
    IsRunning       bool
    TicksProcessed  int64       // Total ticks
    ExecutionsCount int         // Total orders executed
    ErrorsCount     int         // Total errors
    CurrentBalance  float64
    StartBalance    float64
    TotalPnL        float64
    DrawdownPercent float64
    ReturnPercent   float64
    AccountStatus   string      // ACTIVE, BLOWN, AT_LIMIT
}
```

### Control Operations

```go
// Reset to initial state
err := state.Reset()

// Create state snapshot for storage
snapshot := state.Snapshot()  // map[string]interface{}
```

---

## holodeck.go - Main API

### Purpose

Orchestrates the entire trading system. Coordinates configuration, state, subsystems, and callbacks.

### Main Holodeck Struct

```go
type Holodeck struct {
    // Configuration & State
    config *HolodeckConfig
    state  *HolodeckState
    
    // Subsystems (injected)
    executor OrderExecutor
    reader   TickReader
    logger   Logger
    
    // Callbacks for integration
    callbacks HolodeckCallbacks
    
    // Control
    mu       sync.RWMutex
    running  bool
    stopped  bool
    stopChan chan bool
    
    // Performance
    startTime    time.Time
    lastTickTime time.Time
}
```

### Subsystem Interfaces

These are abstract interfaces to be implemented:

#### OrderExecutor

```go
type OrderExecutor interface {
    // Execute an order given current market data
    Execute(order *types.Order, tick *types.Tick, 
            instrument types.Instrument) (*types.ExecutionReport, error)
    
    // Validate order before execution
    Validate(order *types.Order, instrument types.Instrument, 
             availableBalance float64) error
    
    // Calculate commission for an order
    CalculateCommission(price, size float64, 
                       instrument types.Instrument, side string) float64
    
    // Calculate slippage for an order
    CalculateSlippage(size float64, availableDepth int64, 
                     momentum int, instrument types.Instrument) float64
}
```

#### TickReader

```go
type TickReader interface {
    // Check if more ticks available
    HasNext() bool
    
    // Get next tick
    Next() (*types.Tick, error)
    
    // Get number of ticks read
    GetTickCount() int64
    
    // Reset reader to beginning
    Reset() error
    
    // Close the reader
    Close() error
}
```

#### Logger

```go
type Logger interface {
    LogTick(tick *types.Tick)
    LogOrder(order *types.Order)
    LogExecution(exec *types.ExecutionReport)
    LogError(err error)
    LogMetrics(metrics map[string]interface{})
    Close() error
}
```

### Callbacks

Optional callbacks for agent integration:

```go
type HolodeckCallbacks struct {
    // Called when a new tick arrives
    OnTick func(tick *types.Tick) error
    
    // Called after order execution
    OnExecution func(exec *types.ExecutionReport) error
    
    // Called on any error
    OnError func(err error)
    
    // Called when account status changes
    OnStatusChange func(oldStatus, newStatus string)
    
    // Called when session ends
    OnSessionEnd func(status *SessionStatus)
}
```

### Creation & Configuration

```go
// Create from HolodeckConfig
holodeck, err := NewHolodeck(hConfig)
if err != nil {
    log.Fatal(err)
}

// Add subsystems (method chaining)
holodeck.WithExecutor(executor).
         WithReader(reader).
         WithLogger(logger).
         WithCallbacks(callbacks)

// Validate everything is set up
if err := holodeck.Validate(); err != nil {
    log.Fatal(err)
}
```

### Builder Pattern

```go
holodeck, err := NewHolodeckBuilder(hConfig).
    WithExecutor(executor).
    WithReader(reader).
    WithLogger(logger).
    WithCallbacks(callbacks).
    Build()

// Or panic on error
holodeck := NewHolodeckBuilder(hConfig).
    WithExecutor(executor).
    WithReader(reader).
    MustBuild()
```

### Session Lifecycle

#### Start Session

```go
if err := holodeck.Start(); err != nil {
    log.Fatal(err)
}
// Session started, ready to process ticks
```

#### Stop Session

```go
if err := holodeck.Stop(); err != nil {
    log.Fatal(err)
}
// Session stopped, callbacks called, resources cleaned
```

#### Process Ticks

**Single Tick:**
```go
tick := &types.Tick{...}
if err := holodeck.ProcessTick(tick); err != nil {
    // Handle error
}
```

**Tick Stream (Main Loop):**
```go
// ProcessTickStream():
// 1. Calls Start()
// 2. Reads all ticks from reader
// 3. For each tick: ProcessTick() + agent decisions
// 4. On account blown or reader exhausted: Stop()
// 5. Cleans up resources

if err := holodeck.ProcessTickStream(); err != nil {
    log.Fatal(err)
}
```

### Order Execution

```go
order := types.NewBuyOrder(0.01, time.Now())

exec, err := holodeck.ExecuteOrder(order)
if err != nil {
    log.Printf("Execution failed: %v", err)
}

// Check execution status
if exec.IsRejected() {
    fmt.Printf("Rejected: %s\n", exec.ErrorMessage)
} else if exec.IsPartial() {
    fmt.Printf("Partial fill: %f of %f\n", 
               exec.FilledSize, exec.RequestedSize)
} else {
    fmt.Printf("Fully filled at %.5f\n", exec.FillPrice)
}
```

### State & Status Queries

```go
// Get mutable state
state := holodeck.GetState()

// Get session status
status := holodeck.GetStatus()
fmt.Println(status.String())
fmt.Println(status.DebugString())

// Get configuration
config := holodeck.GetConfig()

// Get specific objects
position := holodeck.GetPosition()
balance := holodeck.GetBalance()

// Get history
execHistory := holodeck.GetExecutionHistory()
errors := holodeck.GetErrors()

// Get metrics
metrics := holodeck.GetMetrics()
```

### Diagnostics

```go
// Print status to console
holodeck.PrintStatus()

// Print all metrics
holodeck.PrintMetrics()

// Print all errors
holodeck.PrintErrors()

// Calculate processing speed
tps := holodeck.CalculateTicksPerSecond()
fmt.Printf("Processing: %.2f ticks/sec\n", tps)

// Get performance summary
summary := holodeck.GetPerformanceSummary()
// {session_id, session_duration, ticks_processed, 
//  executions, ticks_per_second, ...}
```

### Control Operations

```go
// Check if running
if holodeck.IsRunning() {
    // ...
}

// Reset to initial state
if err := holodeck.Reset(); err != nil {
    log.Fatal(err)
}
```

---

## Integration Flow

### Complete Workflow

```
1. CONFIGURATION PHASE
   ├─ ConfigLoader.Load("config/forex_eurusd.json")
   ├─ Validate all parameters
   └─ Config loaded

2. INITIALIZATION PHASE
   ├─ NewHolodeckConfig(config)
   │  ├─ Create Instrument (FOREX)
   │  ├─ Create ExecutionParameters
   │  ├─ Generate SessionID
   │  └─ HolodeckConfig ready
   │
   ├─ NewHolodeckState(hConfig)
   │  ├─ Create Position (flat)
   │  ├─ Create Balance ($100k)
   │  ├─ Create ErrorLog
   │  └─ HolodeckState ready
   │
   └─ NewHolodeck(hConfig) + subsystems
      ├─ SetExecutor
      ├─ SetReader
      ├─ SetLogger
      ├─ SetCallbacks
      └─ Holodeck ready

3. EXECUTION PHASE
   ├─ holodeck.Start()
   │  └─ Set IsRunning = true
   │
   ├─ holodeck.ProcessTickStream()
   │  ├─ For each tick:
   │  │  ├─ ProcessTick(tick)
   │  │  │  ├─ Validate tick
   │  │  │  ├─ UpdateTick()
   │  │  │  ├─ Update position price
   │  │  │  ├─ OnTick callback
   │  │  │  └─ Log tick
   │  │  │
   │  │  ├─ Agent decision (via OnTick callback)
   │  │  │  └─ ExecuteOrder(order)
   │  │  │     ├─ Validate order
   │  │  │     ├─ Execute order
   │  │  │     ├─ Update position
   │  │  │     ├─ Update balance
   │  │  │     ├─ OnExecution callback
   │  │  │     └─ Log execution
   │  │  │
   │  │  └─ Check account status
   │  │     └─ If blown, Stop()
   │  │
   │  └─ Continue until end of data
   │
   └─ holodeck.Stop()
      ├─ Set IsRunning = false
      ├─ OnSessionEnd callback
      └─ Clean up resources

4. ANALYSIS PHASE
   ├─ status := holodeck.GetStatus()
   ├─ metrics := holodeck.GetMetrics()
   ├─ executions := holodeck.GetExecutionHistory()
   └─ errors := holodeck.GetErrors()
```

---

## Usage Examples

### Example 1: Basic Setup & Run

```go
package main

import (
    "fmt"
    "log"
)

func main() {
    // 1. Load configuration
    loader := NewConfigLoader("config/forex_eurusd.json")
    if err := loader.Load(); err != nil {
        log.Fatal(err)
    }

    // 2. Create Holodeck configuration
    hConfig, err := NewHolodeckConfig(loader.Config)
    if err != nil {
        log.Fatal(err)
    }

    // 3. Create Holodeck with subsystems
    // (assuming implementations exist)
    executor := NewOrderExecutor(hConfig)
    reader := NewCSVTickReader(hConfig.DataSource.FilePath)
    logger := NewFileLogger("logs/backtest.log")

    holodeck, err := NewHolodeck(hConfig)
    if err != nil {
        log.Fatal(err)
    }

    holodeck.WithExecutor(executor).
            WithReader(reader).
            WithLogger(logger)

    // 4. Run backtest
    if err := holodeck.ProcessTickStream(); err != nil {
        log.Fatal(err)
    }

    // 5. Print results
    status := holodeck.GetStatus()
    fmt.Println(status.DebugString())
}
```

### Example 2: With Callbacks (Agent Integration)

```go
// Define callbacks
callbacks := HolodeckCallbacks{
    OnTick: func(tick *types.Tick) error {
        // Agent makes trading decision
        if shouldBuy(tick) {
            order := types.NewBuyOrder(0.01, tick.Timestamp)
            holodeck.ExecuteOrder(order)
        }
        return nil
    },

    OnExecution: func(exec *types.ExecutionReport) error {
        fmt.Printf("Executed: %s\n", exec)
        return nil
    },

    OnError: func(err error) {
        fmt.Printf("Error: %v\n", err)
    },

    OnStatusChange: func(old, new string) {
        fmt.Printf("Account status: %s -> %s\n", old, new)
    },

    OnSessionEnd: func(status *SessionStatus) {
        fmt.Println("Session ended")
        fmt.Println(status.DebugString())
    },
}

holodeck.WithCallbacks(callbacks)
```

### Example 3: Manual Tick Processing

```go
holodeck.Start()

// Process ticks manually
for tick := range tickStream {
    holodeck.ProcessTick(tick)
    
    if !holodeck.IsRunning() {
        break
    }
}

holodeck.Stop()
```

### Example 4: Multiple Backtests

```go
manager := NewConfigManager()
manager.LoadFromDirectory("config/")

for _, name := range manager.List() {
    config, _ := manager.GetConfig(name)
    hConfig, _ := NewHolodeckConfig(config)
    
    holodeck, _ := NewHolodeck(hConfig).
        WithExecutor(executor).
        WithReader(reader).
        WithLogger(logger)
    
    holodeck.ProcessTickStream()
    
    status := holodeck.GetStatus()
    fmt.Printf("%s: %.2f%% return\n", name, status.ReturnPercent)
}
```

---

## Best Practices

### 1. Configuration Management

**✓ DO:**
```go
// Load from file with validation
loader := NewConfigLoader(configPath)
if err := loader.Load(); err != nil {
    // Handle validation error
}
```

**✗ DON'T:**
```go
// Create Config struct manually without validation
config := &Config{...}  // Missing validation
```

### 2. Thread Safety

**✓ DO:**
```go
// Use provided getters (thread-safe)
tick := state.GetCurrentTick()
position := state.GetPosition()
```

**✗ DON'T:**
```go
// Direct access without mutex
tick := state.CurrentTick  // Race condition!
```

### 3. State Updates

**✓ DO:**
```go
// Use update methods (thread-safe)
state.UpdateTick(tick)
state.AddExecution(exec)
state.UpdateBalance(balance)
```

**✗ DON'T:**
```go
// Modify state directly
state.CurrentTick = tick  // Race condition!
state.ExecutionHistory = append(...)  // Race condition!
```

### 4. Error Handling

**✓ DO:**
```go
// Check errors before proceeding
if err := holodeck.Start(); err != nil {
    switch herr := err.(type) {
    case *types.HolodeckError:
        fmt.Printf("Code: %s, Message: %s\n", 
                   herr.Code, herr.Message)
    }
}
```

**✗ DON'T:**
```go
// Ignore errors
holodeck.Start()  // What if it fails?
```

### 5. Session Lifecycle

**✓ DO:**
```go
// Always clean up
holodeck.Start()
defer holodeck.Stop()
holodeck.ProcessTickStream()
```

**✗ DON'T:**
```go
// Forget to stop
holodeck.Start()
holodeck.ProcessTickStream()
// Resources never cleaned up!
```

### 6. Configuration Validation

**✓ DO:**
```go
// Validate before creating state
if err := loader.Load(); err != nil {
    log.Printf("Config error: %v", err)
    return
}
```

**✗ DON'T:**
```go
// Assume config is valid
loader.Load()  // Ignoring potential errors
hConfig, _ := NewHolodeckConfig(loader.Config)
```

---

## Summary

### Key Takeaways

1. **config.go** - Provides JSON configuration loading with full validation
2. **types.go** - Manages mutable state with thread-safe operations
3. **holodeck.go** - Orchestrates entire system with abstract subsystems

### Architecture Strengths

✓ **Separation of Concerns** - Each file has single responsibility
✓ **Thread Safety** - All state operations use mutex protection
✓ **Validation** - All configurations validated before use
✓ **Extensibility** - Subsystems are abstract interfaces
✓ **Callback Integration** - Easy integration with agents
✓ **Error Handling** - Structured error reporting

### To Use Holodeck

1. Load config from JSON file
2. Create HolodeckConfig
3. Create HolodeckState
4. Implement subsystems (Executor, Reader, Logger)
5. Create Holodeck with subsystems
6. Process ticks
7. Analyze results

### Files Needed to Implement

- **reader/** - CSV tick reader implementation
- **executor/** - Order execution logic
- **logger/** - Logging implementation
- **commission/** - Commission calculations
- **slippage/** - Slippage calculations
- **tests/** - Unit and integration tests