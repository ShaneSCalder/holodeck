# Position Package Documentation

## Overview

The `position` package provides comprehensive position lifecycle management for trading systems, including P&L tracking, risk metrics, trade history, and portfolio management. The package is organized into three focused files:

- **state.go** - Position types and state management
- **pnl.go** - P&L tracking and calculations
- **tracker.go** - Position lifecycle, closing, and portfolio management

## Package Structure

```
position/
├── state.go      (136 lines) - Types & state management
├── pnl.go        (125 lines) - P&L tracking & metrics
└── tracker.go    (228 lines) - Lifecycle & portfolio
```

**Total: 489 lines | 3 files**

---

## State.go - Position Types & State Management

### Types

#### Position
The main position structure representing a single open or closed trade.

```go
type Position struct {
    // Identification
    PositionID    string     // Unique position ID
    Symbol        string     // Instrument symbol (e.g., "EUR/USD")
    OpenTime      time.Time  // When position was opened
    CloseTime     *time.Time // When position was closed

    // Position Details
    Type           string   // LONG, SHORT, FLAT
    Size           float64  // Positive for LONG, Negative for SHORT
    EntryPrice     float64  // Entry price
    AveragePrice   float64  // Average entry price (with slippage)
    CurrentPrice   float64  // Current market price
    LastUpdateTime time.Time // Last update time

    // P&L Tracking
    RealizedPnL    float64  // Realized P&L from closed portions
    UnrealizedPnL  float64  // Mark-to-market P&L
    CommissionPaid float64  // Total commissions paid
    TotalCost      float64  // Total cost of position

    // Risk Metrics
    PeakProfit              float64 // Maximum unrealized profit
    PeakLoss                float64 // Maximum unrealized loss
    MaxAdverseExcursion     float64 // Worst unrealized P&L (MAE)
    MaxFavorableExcursion   float64 // Best unrealized P&L (MFE)
    RunUp                   float64 // Maximum profit from entry
    DrawDown                float64 // Maximum loss from entry

    // Trade History
    Trades       []*PositionTrade // All trades in position
    TradeCount   int              // Number of trades
    EntryTradeID string           // ID of entry trade

    // Status
    Status   string // OPEN, CLOSED, PARTIAL
    IsActive bool   // Is position currently open
}
```

#### PositionTrade
Represents a single trade within a position.

```go
type PositionTrade struct {
    TradeID    string    // Unique trade ID
    Timestamp  time.Time // Trade timestamp
    Action     string    // BUY or SELL
    Size       float64   // Trade size
    Price      float64   // Trade price
    Commission float64   // Commission for this trade
    Slippage   float64   // Slippage for this trade
    IsEntry    bool      // Is this the entry trade
    IsExit     bool      // Is this an exit trade
    PnLAtClose float64   // P&L when position closed
}
```

#### Portfolio
Container for managing multiple positions.

```go
type Portfolio struct {
    positions map[string]*Position // All positions by ID
    TotalPnL  float64              // Total portfolio P&L
}
```

### Constructors

#### NewPosition
Creates a new open position.

```go
func NewPosition(id, symbol string, posType string, size, price float64) *Position
```

**Parameters:**
- `id` - Unique position ID
- `symbol` - Instrument symbol
- `posType` - Position type (LONG, SHORT, FLAT)
- `size` - Position size
- `price` - Entry price

**Returns:** New Position struct

**Example:**
```go
pos := NewPosition("POS001", "EUR/USD", "LONG", 1.0, 1.2000)
```

#### NewPortfolio
Creates a new empty portfolio.

```go
func NewPortfolio() *Portfolio
```

**Example:**
```go
portfolio := NewPortfolio()
```

### Status Check Methods

#### IsLong
Checks if position is a long position.

```go
func (p *Position) IsLong() bool
```

#### IsShort
Checks if position is a short position.

```go
func (p *Position) IsShort() bool
```

#### IsFlat
Checks if position is closed (flat).

```go
func (p *Position) IsFlat() bool
```

#### GetDirection
Returns position direction string.

```go
func (p *Position) GetDirection() string
```

**Returns:** "LONG", "SHORT", or "FLAT"

---

## PnL.go - P&L Tracking & Calculations

### Price Updates

#### UpdatePrice
Updates the current market price and recalculates P&L.

```go
func (p *Position) UpdatePrice(newPrice float64)
```

**Parameters:**
- `newPrice` - New market price

**Example:**
```go
pos.UpdatePrice(1.2050)  // Update EUR/USD to 1.2050
```

#### UpdatePnL
Recalculates all P&L metrics.

```go
func (p *Position) UpdatePnL()
```

**Calculations:**
- Unrealized P&L (mark-to-market)
- Peak profit/loss
- Maximum favorable/adverse excursion
- Run-up and draw-down

### P&L Query Methods

#### IsProfitable
Checks if position is currently profitable.

```go
func (p *Position) IsProfitable() bool
```

**Returns:** `true` if unrealized P&L > 0

#### IsNegative
Checks if position is currently losing money.

```go
func (p *Position) IsNegative() bool
```

**Returns:** `true` if unrealized P&L < 0

#### GetProfit
Returns current profit/loss.

```go
func (p *Position) GetProfit() float64
```

**Returns:**
- Unrealized P&L if position is open
- Realized P&L if position is closed

**Example:**
```go
pnl := pos.GetProfit()
fmt.Printf("Current P&L: %.2f\n", pnl)
```

#### GetTotalPnL
Returns total P&L (realized + unrealized).

```go
func (p *Position) GetTotalPnL() float64
```

#### GetNetPnL
Returns net P&L after commissions.

```go
func (p *Position) GetNetPnL() float64
```

### Risk Metrics

#### GetRatio
Returns profit/loss ratio.

```go
func (p *Position) GetRatio() float64
```

**Formula:** `PeakProfit / |PeakLoss|`

**Returns:** Ratio (higher is better)

**Example:**
```go
ratio := pos.GetRatio()  // Returns: 2.5 (win 2.5x loss)
```

#### GetRiskReward
Returns risk/reward ratio (MFE / MAE).

```go
func (p *Position) GetRiskReward() float64
```

**Formula:** `MaxFavorableExcursion / |MaxAdverseExcursion|`

**Example:**
```go
rr := pos.GetRiskReward()  // Returns: 3.0
```

#### GetMaxFavorableExcursion
Returns maximum favorable excursion (MFE).

```go
func (p *Position) GetMaxFavorableExcursion() float64
```

**Returns:** Best unrealized P&L reached

#### GetMaxAdverseExcursion
Returns maximum adverse excursion (MAE).

```go
func (p *Position) GetMaxAdverseExcursion() float64
```

**Returns:** Worst unrealized P&L reached

#### GetRunUp
Returns maximum run-up from entry.

```go
func (p *Position) GetRunUp() float64
```

#### GetDrawDown
Returns maximum draw-down from entry.

```go
func (p *Position) GetDrawDown() float64
```

---

## Tracker.go - Position Lifecycle & Portfolio Management

### Trade Entry

#### AddTrade
Adds a trade to position history.

```go
func (p *Position) AddTrade(trade *PositionTrade)
```

**Parameters:**
- `trade` - PositionTrade to add

**Example:**
```go
trade := &PositionTrade{
    TradeID:    "TRADE001",
    Timestamp:  time.Now(),
    Action:     "BUY",
    Size:       1.0,
    Price:      1.2000,
    Commission: 10.50,
    IsEntry:    true,
}
pos.AddTrade(trade)
```

### Position Closing

#### Close
Closes the entire position.

```go
func (p *Position) Close(closePrice float64, commission float64) float64
```

**Parameters:**
- `closePrice` - Price at which to close
- `commission` - Commission for closing

**Returns:** P&L from closing

**Example:**
```go
pnl := pos.Close(1.2150, 10.50)  // Close entire position
```

#### ClosePartial
Closes part of the position.

```go
func (p *Position) ClosePartial(closeSize float64, closePrice float64, commission float64) float64
```

**Parameters:**
- `closeSize` - Size to close
- `closePrice` - Close price
- `commission` - Commission

**Returns:** P&L from partial close

**Example:**
```go
pnl := pos.ClosePartial(0.5, 1.2150, 5.25)  // Close half position
```

### Statistics

#### GetDuration
Returns position duration.

```go
func (p *Position) GetDuration() time.Duration
```

**Returns:** Time from open to close (or now)

**Example:**
```go
duration := pos.GetDuration()
fmt.Printf("Held for: %v\n", duration)
```

#### GetTradeCount
Returns number of trades in position.

```go
func (p *Position) GetTradeCount() int
```

#### GetEntryTime
Returns when position was opened.

```go
func (p *Position) GetEntryTime() time.Time
```

### Display Methods

#### String
Returns formatted one-line summary.

```go
func (p *Position) String() string
```

**Output Example:**
```
EUR/USD LONG 1.00 @ 1.2000 | P&L: 150.00 | Status: OPEN
```

#### Details
Returns detailed position information.

```go
func (p *Position) Details() string
```

**Output Example:**
```
Position ID:             POS001
Symbol:                  EUR/USD
Type:                    LONG
Direction:               LONG
Size:                    1.0000
Entry Price:             1.200000
Average Price:           1.200000
Current Price:           1.215000
Open Time:               2024-12-26 10:30:45
Duration:                0 days 2 hours
Status:                  OPEN
Realized P&L:            0.00
Unrealized P&L:          1500.00
Total P&L:               1500.00
Commission Paid:         21.00
Peak Profit:             1500.00
Peak Loss:               0.00
Max Favorable Excursion: 1500.00
Max Adverse Excursion:   0.00
Trade Count:             1
```

### Portfolio Management

#### Add
Adds a position to portfolio.

```go
func (pf *Portfolio) Add(position *Position)
```

#### Remove
Removes a position from portfolio.

```go
func (pf *Portfolio) Remove(positionID string)
```

#### Get
Retrieves a position by ID.

```go
func (pf *Portfolio) Get(positionID string) (*Position, bool)
```

**Returns:** Position and boolean indicating if found

#### GetBySymbol
Gets all positions for a specific symbol.

```go
func (pf *Portfolio) GetBySymbol(symbol string) []*Position
```

**Returns:** Slice of positions for that symbol

#### List
Returns all positions in portfolio.

```go
func (pf *Portfolio) List() []*Position
```

#### Count
Returns number of active (open) positions.

```go
func (pf *Portfolio) Count() int
```

#### TotalExposure
Calculates total exposure (sum of all sizes).

```go
func (pf *Portfolio) TotalExposure() float64
```

**Example:**
```go
exposure := portfolio.TotalExposure()  // Sum of all position sizes
```

#### UpdatePrices
Updates market prices for all positions.

```go
func (pf *Portfolio) UpdatePrices(prices map[string]float64)
```

**Example:**
```go
priceUpdate := map[string]float64{
    "EUR/USD": 1.2050,
    "AAPL":    150.50,
}
portfolio.UpdatePrices(priceUpdate)
```

#### UpdateTotalPnL
Recalculates total portfolio P&L.

```go
func (pf *Portfolio) UpdateTotalPnL()
```

#### GetTotalPnL
Returns total portfolio P&L.

```go
func (pf *Portfolio) GetTotalPnL() float64
```

---

## Complete Usage Examples

### Basic Position Management

```go
package main

import (
    "fmt"
    "holodeck/position"
)

func main() {
    // Create a new position
    pos := position.NewPosition("POS001", "EUR/USD", "LONG", 1.0, 1.2000)

    // Add entry trade
    trade := &position.PositionTrade{
        TradeID:    "TRADE001",
        Action:     "BUY",
        Size:       1.0,
        Price:      1.2000,
        Commission: 10.50,
        IsEntry:    true,
    }
    pos.AddTrade(trade)

    // Update price and check P&L
    pos.UpdatePrice(1.2150)
    fmt.Printf("Current P&L: %.2f\n", pos.GetProfit())

    // Close position
    pnl := pos.Close(1.2150, 10.50)
    fmt.Printf("Realized P&L: %.2f\n", pnl)
}
```

### Portfolio Management

```go
// Create portfolio
portfolio := position.NewPortfolio()

// Add multiple positions
pos1 := position.NewPosition("POS001", "EUR/USD", "LONG", 1.0, 1.2000)
pos2 := position.NewPosition("POS002", "AAPL", "LONG", 100, 150.00)
pos3 := position.NewPosition("POS003", "GOLD", "SHORT", 10, 2050.00)

portfolio.Add(pos1)
portfolio.Add(pos2)
portfolio.Add(pos3)

// Update all prices
prices := map[string]float64{
    "EUR/USD": 1.2050,
    "AAPL":    151.00,
    "GOLD":    2048.00,
}
portfolio.UpdatePrices(prices)

// Check portfolio stats
fmt.Printf("Total P&L: %.2f\n", portfolio.GetTotalPnL())
fmt.Printf("Open Positions: %d\n", portfolio.Count())
fmt.Printf("Total Exposure: %.2f\n", portfolio.TotalExposure())

// List all positions
for _, pos := range portfolio.List() {
    fmt.Println(pos.String())
}
```

### Risk Analysis

```go
pos := position.NewPosition("POS001", "EUR/USD", "LONG", 1.0, 1.2000)

// Simulate price movement
prices := []float64{1.2050, 1.2100, 1.2050, 1.2200, 1.1950, 1.2300}

for _, price := range prices {
    pos.UpdatePrice(price)
}

// Analyze risk metrics
fmt.Printf("Max Favorable Excursion (MFE): %.2f\n", pos.GetMaxFavorableExcursion())
fmt.Printf("Max Adverse Excursion (MAE): %.2f\n", pos.GetMaxAdverseExcursion())
fmt.Printf("Risk/Reward Ratio: %.2f\n", pos.GetRiskReward())
fmt.Printf("Profit/Loss Ratio: %.2f\n", pos.GetRatio())
fmt.Printf("Run-up: %.2f\n", pos.GetRunUp())
fmt.Printf("Draw-down: %.2f\n", pos.GetDrawDown())
```

### Partial Closing

```go
pos := position.NewPosition("POS001", "EUR/USD", "LONG", 2.0, 1.2000)
pos.AddTrade(&position.PositionTrade{
    TradeID: "TRADE001",
    Action:  "BUY",
    Size:    2.0,
    Price:   1.2000,
    IsEntry: true,
})

pos.UpdatePrice(1.2100)

// Close half position
pnl1 := pos.ClosePartial(1.0, 1.2100, 5.25)
fmt.Printf("Partial close P&L: %.2f\n", pnl1)
fmt.Printf("Status: %s\n", pos.Status)  // PARTIAL

// Close remaining
pnl2 := pos.Close(1.2150, 5.25)
fmt.Printf("Final close P&L: %.2f\n", pnl2)
fmt.Printf("Status: %s\n", pos.Status)  // CLOSED
```

---

## Position Status States

| Status | Meaning | Condition |
|--------|---------|-----------|
| **OPEN** | Position is fully open | Size > 0 |
| **PARTIAL** | Part of position closed | Size > 0 but < original |
| **CLOSED** | Position fully closed | Size = 0 |

---

## Position Direction

| Type | Meaning | Size | Profit Formula |
|------|---------|------|-----------------|
| **LONG** | Buy position | Positive | (Close - Entry) × Size |
| **SHORT** | Sell position | Negative | (Entry - Close) × Size |
| **FLAT** | No position | Zero | N/A |

---

## Key Metrics Explained

### P&L (Profit & Loss)
- **Realized P&L**: Locked-in profit/loss from closed positions
- **Unrealized P&L**: Mark-to-market profit/loss on open positions
- **Total P&L**: Realized + Unrealized

### Risk Metrics
- **Peak Profit**: Maximum unrealized profit reached
- **Peak Loss**: Maximum unrealized loss reached
- **MFE (Max Favorable Excursion)**: Best unrealized P&L
- **MAE (Max Adverse Excursion)**: Worst unrealized P&L
- **Run-up**: Maximum profit from entry
- **Draw-down**: Maximum loss from entry

### Ratios
- **Risk/Reward Ratio**: MFE / |MAE| (higher is better)
- **Profit/Loss Ratio**: PeakProfit / |PeakLoss|

---

## Best Practices

### 1. Always Create with Proper ID
```go
pos := position.NewPosition("POS_" + fmt.Sprint(time.Now().UnixNano()), symbol, posType, size, price)
```

### 2. Track All Trades
```go
// Always record entry and exit trades
pos.AddTrade(entryTrade)
pos.AddTrade(exitTrade)
```

### 3. Update Prices Regularly
```go
// Update prices as market moves
pos.UpdatePrice(newMarketPrice)
```

### 4. Check P&L Before Closing
```go
if pos.IsProfitable() {
    pnl := pos.Close(closePrice, commission)
}
```

### 5. Use Portfolio for Multiple Positions
```go
// Use Portfolio, don't manage positions individually
portfolio := position.NewPortfolio()
portfolio.Add(pos1)
portfolio.Add(pos2)
portfolio.UpdatePrices(prices)
```

### 6. Monitor Risk Metrics
```go
if pos.GetRiskReward() > 2.0 {
    // Good risk/reward
}
if pos.GetDrawDown() > maxAllowedDD {
    // Stop loss triggered
}
```

---

## Integration Points

### With Account Package
```go
// Position P&L affects account balance
acc.RecordTrade(pos.PositionID, pos.GetProfit(), pos.CommissionPaid)
```

### With Instrument Package
```go
// Use instrument specs for risk calculations
riskAmount := instrument.GetRiskAmount(pos.Size)
```

### With Executor Package
```go
// Executor creates positions
pos := position.NewPosition(orderID, symbol, "LONG", size, price)
```

---

## Performance Considerations

- **Memory:** Position struct is ~500 bytes, plus trade history
- **Speed:** All operations are O(1) except portfolio listing
- **Trade History:** Consider pruning for long-running systems

---

## Error Handling

### Invalid Position
```go
if pos.IsFlat() {
    return fmt.Errorf("cannot operate on flat position")
}
```

### Invalid Close
```go
if pos.Status == "CLOSED" {
    return fmt.Errorf("position already closed")
}
```

### Portfolio Errors
```go
if retrieved, ok := portfolio.Get(posID); !ok {
    return fmt.Errorf("position not found in portfolio")
}
```

---

## Related Documentation

- [Account Package](account.md) - Account management
- [Instrument Package](instrument.md) - Instrument definitions
- [Executor Package](executor.md) - Order execution

---

**Created:** December 26, 2024  
**Version:** 1.0  
**Status:** Production Ready  
**Lines of Code:** 489 (3 files)  
**Methods:** 30+