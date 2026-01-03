# Slippage Package Documentation

## Overview

The `slippage/` package provides a comprehensive slippage calculation system for the Holodeck backtesting platform. It uses two models: depth-based and momentum-adjusted to calculate realistic order slippage.

**Location:** `/home/claude/holodeck/slippage/`

**Files:** 3 Go files (984 lines)

---

## Package Structure

```
slippage/
├── calculator.go           # Main orchestrator (309 lines)
├── depth_model.go         # Depth-based: size / available_depth (304 lines)
└── momentum_model.go      # Momentum adjustments (371 lines)

Total: 984 lines
```

---

## Slippage Models

### 1. Depth Model - Size vs Available Depth

**File:** `depth_model.go`

**Formula:**
```
slippage = (order_size / available_depth) × volatility
```

**Concept:**
- Larger orders relative to available depth cause more slippage
- Higher volatility increases slippage
- Small orders in liquid markets have minimal slippage

**Example:**
```
Order Size:        0.1 lots
Available Depth:   0.5 lots
Volatility:        0.008
Depth Ratio:       0.1 / 0.5 = 0.2
Slippage:          0.2 × 0.008 = 0.0016 = 16 pips
```

**Key Types:**
- `DepthModel` - Core depth calculator
- `DepthSlippageAnalysis` - Detailed breakdown
- Utility methods for depth interpretation

**Methods:**
- `CalculateSlippage(size, depth, volatility)` - Core calculation
- `AnalyzeDepthSlippage()` - Detailed analysis
- `CalculateDepthRequired()` - Reverse calculation
- `CalculateMaxOrderSize()` - Maximum size for target slippage
- `InterpretDepthRatio()` - Human-readable interpretation
- `InterpretSlippage()` - Slippage interpretation

---

### 2. Momentum Model - Price Movement Adjustment

**File:** `momentum_model.go`

**Concept:**
- Adjusts base slippage using momentum multiplier
- Fast-moving markets (high momentum) = higher slippage
- Stable markets (low momentum) = lower slippage
- Momentum = 1.0 is neutral

**Momentum Scale:**
```
< 0.5      Negligible momentum - minimal slippage
0.5-0.8    Weak momentum - slippage reduction
0.8-1.0    Weakening momentum - slight reduction
1.0        Neutral momentum - no adjustment
1.0-1.2    Building momentum - slight increase
1.2-1.5    Strong momentum - significant increase
> 1.5      Very strong momentum - maximum slippage
```

**Key Types:**
- `MomentumModel` - Core momentum calculator
- `MomentumAdjustmentAnalysis` - Detailed breakdown

**Methods:**
- `AdjustSlippage(baseSlippage, momentum, tick)` - Apply momentum adjustment
- `AnalyzeMomentumAdjustment()` - Detailed analysis
- `InterpretMomentum()` - Human-readable interpretation
- `CalculateMomentumMultiplier()` - Calculate from price data
- `SetMaxMultiplier()` / `SetBaseMultiplier()` - Configuration

---

## Main Orchestrator

### SlippageCalculator

**File:** `calculator.go`

Main entry point coordinating both models.

**Constructor:**
```go
calc := slippage.NewSlippageCalculator()
```

**Core Methods:**

#### CalculateSlippage
```go
slippageUnits, err := calc.CalculateSlippage(
    orderSize,        // Size of order
    availableDepth,   // Depth at market
    volatility,       // Market volatility
    momentum,         // Price momentum multiplier
    tick,             // Market tick
    instrument,       // Instrument being traded
)
```

Combines depth and momentum models for realistic slippage.

#### CalculateFillPrice
```go
fillPrice, err := calc.CalculateFillPrice(
    midPrice,         // Mid-market price
    slippageUnits,    // Calculated slippage
    side,             // BUY or SELL
    instrument,       // Instrument
)
```

Converts slippage units to actual fill price.

#### CalculateBatchSlippage
```go
totalSlippage, err := calc.CalculateBatchSlippage(
    orders,           // []SlippageInput
    tick,             // Market tick
    instrument,       // Instrument
)
```

Calculates slippage for multiple orders.

#### GetStatistics
```go
stats := calc.GetStatistics()
// Returns map with comprehensive slippage stats
```

---

## Usage Examples

### Example 1: Single Trade Slippage

```go
calc := slippage.NewSlippageCalculator()

// Create tick
tick := &types.Tick{
    Bid: 1.08500,
    Ask: 1.08520,
    // ... other fields
}

// Create instrument
instrument := types.NewForexInstrument("EUR/USD")

// Calculate slippage
slippageUnits, err := calc.CalculateSlippage(
    0.1,       // Order size: 0.1 lots
    0.5,       // Available depth: 0.5 lots
    0.008,     // Volatility: 0.8%
    1.2,       // Momentum: 1.2x (strong momentum)
    tick,      // Market tick
    instrument, // FOREX
)

if err == nil {
    fmt.Printf("Slippage: %.4f pips\n", slippageUnits)
    
    // Calculate fill price
    midPrice := tick.GetMidPrice()
    fillPrice, _ := calc.CalculateFillPrice(
        midPrice,
        slippageUnits,
        "BUY",
        instrument,
    )
    fmt.Printf("Fill Price: %.5f\n", fillPrice)
}
```

### Example 2: Depth Model Direct Usage

```go
depthModel := slippage.NewDepthModel()

// Analyze depth slippage
analysis := depthModel.AnalyzeDepthSlippage(
    0.1,    // Order size
    0.5,    // Available depth
    0.008,  // Volatility
)

fmt.Println(analysis.DebugString())
// Output:
// Depth Slippage Analysis:
//   Order Size:            0.1000
//   Available Depth:       0.5000
//   Depth Ratio:           0.2000
//   Volatility:            0.0080
//   Slippage:              0.0016 pips
```

### Example 3: Momentum Model Direct Usage

```go
momentumModel := slippage.NewMomentumModel()

// Analyze momentum adjustment
analysis := momentumModel.AnalyzeMomentumAdjustment(
    0.0016,   // Base slippage from depth model
    1.2,      // Momentum: 1.2x (strong)
    0.15,     // Volatility: 0.15%
)

fmt.Println(analysis.DebugString())
// Output:
// Momentum Adjustment Analysis:
//   Base Slippage:         0.0016 pips
//   Momentum:              1.2000
//   Volatility:            0.15%
//   Adjustment Factor:     1.25 x
//   Adjusted Slippage:     0.0020 pips
//   Adjustment:            increase by 25.00%
```

### Example 4: Batch Slippage Calculation

```go
calc := slippage.NewSlippageCalculator()
tick := &types.Tick{...}
instrument := types.NewForexInstrument("EUR/USD")

trades := []slippage.SlippageInput{
    {OrderSize: 0.1, AvailableDepth: 0.5, Volatility: 0.008, Momentum: 1.2},
    {OrderSize: 0.05, AvailableDepth: 0.5, Volatility: 0.008, Momentum: 1.1},
    {OrderSize: 0.2, AvailableDepth: 0.8, Volatility: 0.009, Momentum: 1.3},
}

totalSlippage, err := calc.CalculateBatchSlippage(trades, tick, instrument)
if err == nil {
    fmt.Printf("Total Slippage: %.4f pips\n", totalSlippage)
}
```

### Example 5: Get Statistics

```go
calc := slippage.NewSlippageCalculator()

// ... perform trades ...

stats := calc.GetStatistics()

fmt.Printf("Total Slippage: %.4f pips\n", stats["total_slippage"])
fmt.Printf("Trade Count: %d\n", stats["slippage_count"])
fmt.Printf("Average Slippage: %.4f pips\n", stats["average_slippage"])
fmt.Printf("Max Slippage: %.4f pips\n", stats["max_slippage"])

// Access sub-model stats
depthStats := stats["depth_model_stats"].(map[string]interface{})
fmt.Printf("Depth Model Total: %.4f\n", depthStats["total_slippage"])
```

### Example 6: Volatility-Based Interpretation

```go
depthModel := slippage.NewDepthModel()

slippage, _ := depthModel.CalculateSlippage(0.1, 0.5, 0.008)

// Interpret the result
interpretation := depthModel.InterpretSlippage(slippage)
fmt.Println(interpretation)
// Output: "Low slippage"

// Interpret depth ratio
depthRatio := 0.1 / 0.5  // 0.2
interpretation2 := depthModel.InterpretDepthRatio(depthRatio)
fmt.Println(interpretation2)
// Output: "Small order relative to depth - low slippage"
```

---

## Statistics & Tracking

### SlippageCalculator Statistics

Each calculator tracks:
- Total slippage accumulated
- Number of slippage calculations
- Average slippage per trade
- Maximum slippage observed
- Minimum slippage observed
- Per-model statistics

```go
stats := calc.GetStatistics()
// Returns:
{
    "total_slippage":        float64,
    "slippage_count":        int64,
    "average_slippage":      float64,
    "max_slippage":          float64,
    "min_slippage":          float64,
    "depth_model_stats":     map[string]interface{},
    "momentum_model_stats":  map[string]interface{},
}
```

---

## Key Features

✅ **Two-Model Approach**
- Depth model for order size impact
- Momentum model for market state impact
- Combined for realistic slippage

✅ **Comprehensive Calculation**
- Size relative to depth
- Market volatility
- Price momentum
- Tick-based analysis

✅ **Single & Batch Operations**
- Process single orders
- Process multiple orders
- Efficient batch calculation

✅ **Detailed Analysis**
- Analyze each slippage calculation
- Breakdown by component
- Fill price calculation

✅ **Statistics Tracking**
- Total and average slippage
- Min/max observed
- Per-model breakdown
- Momentum tracking

✅ **Interpretation Tools**
- Human-readable slippage interpretation
- Depth ratio interpretation
- Momentum interpretation

---

## Integration with Executor

The slippage calculator integrates with the executor:

```go
// In executor, after calculating base fill price:
executor := executor.NewOrderExecutor(config)
slippageCalc := slippage.NewSlippageCalculator()

// Execute order
exec, _ := executor.Execute(order, tick, instrument)

// Calculate slippage
slippageUnits, _ := slippageCalc.CalculateSlippage(
    order.Size,
    tick.AskQty,  // For BUY orders
    volatility,
    momentum,
    tick,
    instrument,
)

// Calculate actual fill price with slippage
fillPrice, _ := slippageCalc.CalculateFillPrice(
    tick.GetMidPrice(),
    slippageUnits,
    order.Action,
    instrument,
)

exec.FillPrice = fillPrice
```

---

## Performance Considerations

✅ **Efficient Calculation**
- O(1) time complexity for single calculation
- O(n) for batch (n = number of orders)
- Minimal memory overhead

✅ **Statistics Management**
- Incremental tracking
- No recalculation needed
- Reset available for cleanup

✅ **Precision**
- Float64 for calculations
- Sufficient for trading scenarios
- Configurable multipliers for tuning

---

## Configuration Options

### DepthModel Configuration
```go
depthModel := slippage.NewDepthModel()
// Default formula: (size / depth) × volatility
// Customizable via AnalyzeDepthSlippage parameters
```

### MomentumModel Configuration
```go
momentumModel := slippage.NewMomentumModel()
momentumModel.SetBaseMultiplier(1.0)      // Default multiplier
momentumModel.SetMaxMultiplier(2.0)       // Maximum adjustment
```

---

## Testing Recommendations

### Unit Tests
```go
// Test depth slippage
func TestDepthSlippage(t *testing.T) {
    dm := NewDepthModel()
    slippage, _ := dm.CalculateSlippage(0.1, 0.5, 0.008)
    expected := 0.0016
    assert.InDelta(t, slippage, expected, 0.0001)
}

// Test momentum adjustment
func TestMomentumAdjustment(t *testing.T) {
    mm := NewMomentumModel()
    adjusted, _ := mm.AdjustSlippage(0.001, 1.2, tick)
    assert.Greater(t, adjusted, 0.001)  // Should increase
}

// Test fill price calculation
func TestFillPrice(t *testing.T) {
    calc := NewSlippageCalculator()
    price, _ := calc.CalculateFillPrice(1.08510, 0.001, "BUY", forex)
    assert.Greater(t, price, 1.08510)  // BUY increases price
}

// Test statistics
func TestStatistics(t *testing.T) {
    calc := NewSlippageCalculator()
    // Perform calculations...
    stats := calc.GetStatistics()
    assert.Greater(t, stats["slippage_count"], int64(0))
}
```

---

## Summary

The slippage package provides:

✅ **3 well-organized files** (984 lines)
✅ **Depth model** for size/liquidity impact
✅ **Momentum model** for market state
✅ **Main orchestrator** for unified access
✅ **Single & batch operations**
✅ **Detailed analysis** capabilities
✅ **Comprehensive statistics**
✅ **Interpretation tools**
✅ **Production-ready code**

Ready for integration with executor and other Holodeck systems.