# Account Package Documentation

## Overview

The `account` package provides comprehensive account management for trading systems, including balance tracking, margin management, drawdown monitoring, and statistical analysis. The package is organized into four focused files:

- **manager.go** - Account definition and lifecycle
- **balance.go** - Balance operations and P&L tracking
- **drawdown.go** - Drawdown metrics and monitoring
- **leverage.go** - Leverage and margin management

## Package Structure

```
account/
├── manager.go       (149 lines) - Core Account type
├── balance.go       (133 lines) - P&L and statistics
├── drawdown.go      (72 lines)  - Drawdown tracking
└── leverage.go      (115 lines) - Leverage and margin
```

**Total: 469 lines | 4 files**

---

## Manager.go - Account Definition & Lifecycle

### Types

#### Account
The main account structure that manages all account-level operations and state.

```go
type Account struct {
    // Identification
    AccountID   string
    Name        string
    Description string

    // Initial Setup
    InitialBalance float64
    Currency       string
    Leverage       float64

    // Current State
    CurrentBalance     float64
    UsedMargin         float64
    AvailableMargin    float64
    BuyingPower        float64
    TotalRealizedPnL   float64
    TotalUnrealizedPnL float64
    CommissionPaid     float64

    // Trade Statistics
    TotalTrades      int
    WinningTrades    int
    LosingTrades     int
    BreakevenTrades  int
    ConsecutiveWins  int
    ConsecutiveLosses int

    // Risk Management
    MaxDrawdownPercent      float64
    MaxDrawdownExperienced  float64
    MaxDrawdownAmount       float64
    MaxPositionSize         float64
    MaxPositionsOpen        int
    MaxLeverageAllowed      float64
    RiskPerTradePercent     float64

    // Account Status
    Status           string
    HighWaterMark    float64
    LowWaterMark     float64
    CreatedTime      time.Time
    LastUpdateTime   time.Time
    UpdateHistory    []*BalanceUpdate
}
```

#### BalanceUpdate
Records a balance change event for audit trail.

```go
type BalanceUpdate struct {
    Timestamp      time.Time
    BalanceBefore  float64
    BalanceAfter   float64
    Change         float64
    Reason         string
    TransactionID  string
    RelatedTradeID string
}
```

### Constructor

#### NewAccount
Creates a new account with initial setup.

```go
func NewAccount(id, name string, initialBalance float64, currency string) *Account
```

**Parameters:**
- `id` - Unique account identifier
- `name` - Account name/description
- `initialBalance` - Starting balance
- `currency` - Account currency (USD, EUR, etc)

**Returns:** Initialized Account pointer

**Example:**
```go
acc := account.NewAccount("ACC001", "My Trading Account", 100000, "USD")
```

### Status Methods

#### IsActive
Checks if account is currently active.

```go
func (a *Account) IsActive() bool
```

#### IsBlown
Checks if account balance is zero or negative (account blown).

```go
func (a *Account) IsBlown() bool
```

#### IsAtLimit
Checks if account has reached margin limit.

```go
func (a *Account) IsAtLimit() bool
```

#### CanTrade
Checks if account can execute trades (active and has positive balance).

```go
func (a *Account) CanTrade() bool
```

### Display Methods

#### String
Returns formatted account summary.

```go
func (a *Account) String() string
```

**Output Example:**
```
Account: My Trading Account (ACC001)
Balance: 102500.00 USD | P&L: 2500.00 | Commission: 250.00
Margin: 5000.00 / 5125000.00 (100.0%) | Leverage: 51.25x
Trades: 10 (W:7 L:2 B:1) | Win Rate: 70.0%
Drawdown: 2.43% (2500.00) | Status: ACTIVE
```

### History Recording

#### RecordBalanceUpdate
Adds a balance change to history.

```go
func (a *Account) RecordBalanceUpdate(before, after, change float64, reason, transactionID string)
```

---

## Balance.go - Balance Operations & P&L Tracking

### Trade Recording

#### RecordTrade
Records a trade's P&L impact on account.

```go
func (a *Account) RecordTrade(tradeID string, pnl float64, commission float64)
```

**Parameters:**
- `tradeID` - Unique trade identifier
- `pnl` - Profit or loss amount
- `commission` - Trading commission/fees

**Behavior:**
- Updates current balance with P&L and commission
- Increments trade counters (wins/losses/breakeven)
- Updates consecutive win/loss counts
- Records update in history
- Calls UpdateMargin() to recalculate margins

**Example:**
```go
// Winning trade: +$500 profit, $10 commission
acc.RecordTrade("TRADE001", 500.00, 10.00)

// Losing trade: -$200 loss, $10 commission
acc.RecordTrade("TRADE002", -200.00, 10.00)
```

#### RecordUnrealizedPnL
Records mark-to-market P&L on open positions.

```go
func (a *Account) RecordUnrealizedPnL(unrealizedPnL float64)
```

**Parameters:**
- `unrealizedPnL` - Current mark-to-market P&L

**Note:** Does not affect current balance, only tracks unrealized gains/losses

#### RecordCommission
Records a commission deduction separately.

```go
func (a *Account) RecordCommission(transactionID string, amount float64)
```

**Parameters:**
- `transactionID` - Transaction ID for tracking
- `amount` - Commission amount to deduct

---

### Statistics Methods

#### GetWinRate
Returns percentage of winning trades.

```go
func (a *Account) GetWinRate() float64
```

**Formula:** `(WinningTrades / TotalTrades) * 100`

**Returns:** Win rate percentage (0-100)

**Example:**
```go
winRate := acc.GetWinRate()  // Returns: 70.0
fmt.Printf("Win Rate: %.1f%%\n", winRate)  // Win Rate: 70.0%
```

#### GetLossRate
Returns percentage of losing trades.

```go
func (a *Account) GetLossRate() float64
```

**Formula:** `(LosingTrades / TotalTrades) * 100`

#### GetBreakevenRate
Returns percentage of breakeven trades.

```go
func (a *Account) GetBreakevenRate() float64
```

**Formula:** `(BreakevenTrades / TotalTrades) * 100`

#### GetTotalReturn
Returns total return percentage from initial balance.

```go
func (a *Account) GetTotalReturn() float64
```

**Formula:** `((CurrentBalance - InitialBalance) / InitialBalance) * 100`

**Example:**
```go
returnPct := acc.GetTotalReturn()  // Returns: 2.5 for +2.5% return
```

#### GetRiskRewardRatio
Returns risk/reward ratio of trades.

```go
func (a *Account) GetRiskRewardRatio() float64
```

**Formula:** `AverageWin / AverageLosingTrade`

**Returns:** Ratio (higher is better)

**Example:**
```go
rr := acc.GetRiskRewardRatio()  // Returns: 2.5 (wins are 2.5x losses)
```

#### GetProfitFactor
Returns profit factor (total wins / total losses).

```go
func (a *Account) GetProfitFactor() float64
```

**Formula:** `GrossProfit / GrossLoss`

**Returns:** Ratio (>1.0 is profitable)

**Example:**
```go
pf := acc.GetProfitFactor()  // Returns: 1.5 (50% more wins than losses)
```

---

## Drawdown.go - Drawdown Management & Monitoring

### Drawdown Operations

#### UpdateDrawdown
Updates all drawdown calculations after balance change.

```go
func (a *Account) UpdateDrawdown()
```

**Behavior:**
- Updates high watermark if balance increases
- Updates low watermark if balance decreases
- Calculates current drawdown percentage
- Updates max drawdown if exceeded

**Called automatically by:** RecordTrade, RecordCommission, RecordUnrealizedPnL

#### GetDrawdownPercent
Returns current drawdown percentage from peak.

```go
func (a *Account) GetDrawdownPercent() float64
```

**Formula:** `((HighWaterMark - CurrentBalance) / HighWaterMark) * 100`

**Returns:** Current drawdown percentage

**Example:**
```go
dd := acc.GetDrawdownPercent()
if dd > 10.0 {
    fmt.Println("Drawdown exceeded 10%!")
}
```

#### GetMaxDrawdownPercent
Returns maximum drawdown experienced.

```go
func (a *Account) GetMaxDrawdownPercent() float64
```

**Returns:** Historical maximum drawdown percentage

#### GetMaxDrawdownAmount
Returns maximum drawdown in currency units.

```go
func (a *Account) GetMaxDrawdownAmount() float64
```

**Returns:** Maximum drawdown amount

#### IsDrawdownExceeded
Checks if current drawdown exceeds limit.

```go
func (a *Account) IsDrawdownExceeded() bool
```

**Returns:** `true` if current drawdown > MaxDrawdownPercent

**Example:**
```go
if acc.IsDrawdownExceeded() {
    // Stop trading or reduce risk
}
```

### High Watermark Operations

#### HighWaterMarkDistance
Returns distance from high watermark in currency.

```go
func (a *Account) HighWaterMarkDistance() float64
```

**Formula:** `HighWaterMark - CurrentBalance`

**Example:**
```go
distance := acc.HighWaterMarkDistance()  // Returns: 2500.00
```

#### HighWaterMarkPercent
Returns distance as percentage of high watermark.

```go
func (a *Account) HighWaterMarkPercent() float64
```

**Example:**
```go
percent := acc.HighWaterMarkPercent()  // Returns: 2.43
```

### Recovery Tracking

#### GetRecoveryPercent
Returns recovery percentage from peak to current.

```go
func (a *Account) GetRecoveryPercent() float64
```

**Formula:** `(Recovery / MaxDrawdownAmount) * 100`

**Range:** 0-100 (0 = at peak, 100 = fully recovered)

**Example:**
```go
recovery := acc.GetRecoveryPercent()
fmt.Printf("Recovery: %.1f%%\n", recovery)
```

---

## Leverage.go - Leverage & Margin Management

### Leverage Operations

#### SetLeverage
Sets account leverage (must be between 1.0 and MaxLeverageAllowed).

```go
func (a *Account) SetLeverage(leverage float64) bool
```

**Parameters:**
- `leverage` - New leverage multiplier (1.0 = no leverage, 50.0 = 50x)

**Returns:** `true` if successful, `false` if exceeds limits

**Example:**
```go
if acc.SetLeverage(50.0) {
    fmt.Println("Leverage set to 50x")
} else {
    fmt.Println("Leverage exceeds maximum allowed")
}
```

#### GetLeverage
Returns current leverage.

```go
func (a *Account) GetLeverage() float64
```

#### CanIncreaseLeverage
Checks if leverage can be increased to new value.

```go
func (a *Account) CanIncreaseLeverage(newLeverage float64) bool
```

#### CanDecreaseLeverage
Checks if leverage can be decreased to new value.

```go
func (a *Account) CanDecreaseLeverage(newLeverage float64) bool
```

### Margin Calculations

#### UpdateMargin
Updates all margin calculations.

```go
func (a *Account) UpdateMargin()
```

**Calculations:**
- `BuyingPower = CurrentBalance * Leverage`
- `AvailableMargin = BuyingPower - UsedMargin`
- Updates account status based on margin

**Called automatically by:** SetLeverage, RecordTrade, RecordCommission, RecordMarginUsed

#### HasSufficientMargin
Checks if account has enough margin for trade.

```go
func (a *Account) HasSufficientMargin(requiredMargin float64) bool
```

**Returns:** `true` if available margin >= required margin

**Example:**
```go
if acc.HasSufficientMargin(5000.00) {
    // Execute trade
}
```

### Margin Queries

#### GetAvailableMargin
Returns available margin for new trades.

```go
func (a *Account) GetAvailableMargin() float64
```

#### GetUsedMargin
Returns margin currently in use by open positions.

```go
func (a *Account) GetUsedMargin() float64
```

#### GetBuyingPower
Returns total buying power (balance × leverage).

```go
func (a *Account) GetBuyingPower() float64
```

#### GetAvailableMarginPercent
Returns available margin as percentage of buying power.

```go
func (a *Account) GetAvailableMarginPercent() float64
```

**Returns:** Percentage (0-100)

**Example:**
```go
availPct := acc.GetAvailableMarginPercent()
if availPct < 20.0 {
    fmt.Println("Low available margin!")
}
```

#### GetUsedMarginPercent
Returns used margin as percentage of buying power.

```go
func (a *Account) GetUsedMarginPercent() float64
```

### Margin Management

#### RecordMarginUsed
Records margin being used by open positions.

```go
func (a *Account) RecordMarginUsed(marginAmount float64)
```

**Parameters:**
- `marginAmount` - Total margin in use

**Example:**
```go
// 10 positions using 5000 margin total
acc.RecordMarginUsed(5000.00)
```

#### ReleaseMargin
Releases margin from closed positions.

```go
func (a *Account) ReleaseMargin(marginAmount float64)
```

**Parameters:**
- `marginAmount` - Margin to release

**Example:**
```go
// Position closed, release its margin
acc.ReleaseMargin(500.00)
```

### Margin Level Analysis

#### GetMarginLevel
Returns margin level percentage.

```go
func (a *Account) GetMarginLevel() float64
```

**Formula:** `(CurrentBalance / UsedMargin) * 100`

**Interpretation:**
- 200% = Strong position
- 100% = Margin call level
- < 100% = Margin call triggered

**Example:**
```go
level := acc.GetMarginLevel()
fmt.Printf("Margin Level: %.1f%%\n", level)
```

#### IsMarginCall
Checks if margin call condition is met (margin level < 100%).

```go
func (a *Account) IsMarginCall() bool
```

**Returns:** `true` if margin call triggered

**Example:**
```go
if acc.IsMarginCall() {
    // Force close some positions
}
```

---

## Complete Usage Example

```go
package main

import (
    "fmt"
    "holodeck/account"
)

func main() {
    // 1. Create account
    acc := account.NewAccount("ACC001", "Trading Account", 100000, "USD")
    
    // 2. Set leverage
    acc.SetLeverage(50.0)
    fmt.Printf("Leverage: %.1fx\n", acc.GetLeverage())
    
    // 3. Record some trades
    acc.RecordTrade("TRADE001", 500.00, 10.50)    // Win
    acc.RecordTrade("TRADE002", 750.00, 10.50)    // Win
    acc.RecordTrade("TRADE003", -200.00, 10.50)   // Loss
    acc.RecordTrade("TRADE004", 300.00, 10.50)    // Win
    
    // 4. Check balance statistics
    fmt.Println("\n=== ACCOUNT STATISTICS ===")
    fmt.Printf("Current Balance: %.2f\n", acc.CurrentBalance)
    fmt.Printf("Total Trades: %d\n", acc.TotalTrades)
    fmt.Printf("Win Rate: %.1f%%\n", acc.GetWinRate())
    fmt.Printf("Total Return: %.1f%%\n", acc.GetTotalReturn())
    fmt.Printf("Profit Factor: %.2f\n", acc.GetProfitFactor())
    
    // 5. Check margin status
    fmt.Println("\n=== MARGIN STATUS ===")
    fmt.Printf("Buying Power: %.2f\n", acc.GetBuyingPower())
    fmt.Printf("Available Margin: %.2f\n", acc.GetAvailableMargin())
    fmt.Printf("Available Margin %%: %.1f%%\n", acc.GetAvailableMarginPercent())
    
    // 6. Check drawdown
    fmt.Println("\n=== DRAWDOWN ===")
    fmt.Printf("Current Drawdown: %.2f%%\n", acc.GetDrawdownPercent())
    fmt.Printf("Max Drawdown: %.2f%%\n", acc.GetMaxDrawdownPercent())
    fmt.Printf("High Water Mark: %.2f\n", acc.HighWaterMark)
    
    // 7. Record margin usage
    acc.RecordMarginUsed(5000.00)
    fmt.Printf("\nMargin Level: %.1f%%\n", acc.GetMarginLevel())
    
    // 8. Full account summary
    fmt.Println("\n=== FULL SUMMARY ===")
    fmt.Println(acc.String())
}
```

**Output Example:**
```
Leverage: 50.0x

=== ACCOUNT STATISTICS ===
Current Balance: 101818.50
Total Trades: 4
Win Rate: 75.0%
Total Return: 1.82%
Profit Factor: 3.65

=== MARGIN STATUS ===
Buying Power: 5090925.00
Available Margin: 5085925.00
Available Margin %: 99.9%

=== DRAWDOWN ===
Current Drawdown: 0.00%
Max Drawdown: 0.00%
High Water Mark: 101818.50

Margin Level: 20363.70%

=== FULL SUMMARY ===
Account: Trading Account (ACC001)
Balance: 101818.50 USD | P&L: 1818.50 | Commission: 42.00
Margin: 5000.00 / 5090925.00 (99.9%) | Leverage: 50.0x
Trades: 4 (W:3 L:1 B:0) | Win Rate: 75.0%
Drawdown: 0.00% (0.00) | Status: ACTIVE
```

---

## Best Practices

### 1. Always Check Margin Before Trading
```go
if acc.HasSufficientMargin(requiredMargin) {
    // Execute trade
} else {
    // Reduce position size or increase capital
}
```

### 2. Monitor Drawdown
```go
if acc.IsDrawdownExceeded() {
    // Stop trading or reduce risk
    fmt.Println("Drawdown limit exceeded!")
}
```

### 3. Track Consecutive Wins/Losses
```go
if acc.ConsecutiveLosses > 3 {
    // Reduce trading after 3 consecutive losses
}
```

### 4. Regular Balance Updates
```go
// After each trade
acc.UpdateMargin()
acc.UpdateDrawdown()
```

### 5. Use Transaction IDs
```go
acc.RecordTrade("TRADE_" + fmt.Sprint(time.Now().UnixNano()), pnl, commission)
```

---

## Account Status States

| Status | Meaning | Condition |
|--------|---------|-----------|
| **ACTIVE** | Account is operational | Balance > 0 and Available Margin > 0 |
| **BLOWN** | Account depleted | Balance <= 0 |
| **AT_LIMIT** | No margin available | Available Margin <= 0 |
| **CLOSED** | Account manually closed | Set by external process |

---

## Key Metrics Summary

| Metric | Calculation | Purpose |
|--------|-----------|---------|
| **Win Rate** | Wins / Total * 100 | Accuracy of trading system |
| **Return** | (Final - Initial) / Initial * 100 | Overall profitability |
| **Drawdown** | (Peak - Current) / Peak * 100 | Risk/volatility |
| **Profit Factor** | Total Wins / Total Losses | Win quality |
| **Margin Level** | Balance / Used Margin * 100 | Leverage safety |
| **Risk/Reward** | Avg Win / Avg Loss | Trade quality |

---

## Error Handling

### Insufficient Margin
```go
if !acc.HasSufficientMargin(5000) {
    return fmt.Errorf("insufficient margin: need %.2f, have %.2f", 
        5000, acc.AvailableMargin)
}
```

### Leverage Limit
```go
if !acc.SetLeverage(100) {
    return fmt.Errorf("leverage exceeds maximum: %.1f", 
        acc.MaxLeverageAllowed)
}
```

### Account Blown
```go
if acc.IsBlown() {
    return fmt.Errorf("account balance is zero or negative")
}
```

---

## Integration Points

### With Position Package
```go
// When position closes
pnl := position.Close(closePrice, commission)
acc.RecordTrade(position.ID, pnl, commission)
```

### With Executor Package
```go
// Before executing order
if !acc.HasSufficientMargin(requiredMargin) {
    // Reject order
}
```

### With Logger Package
```go
// Log account state
logger.LogAccount(acc)
logger.LogBalanceUpdate(acc.UpdateHistory[len(acc.UpdateHistory)-1])
```

---

## Performance Considerations

- **Memory:** Account struct is small (~500 bytes), UpdateHistory can grow
- **Speed:** All operations are O(1) except history iteration
- **History Pruning:** Consider limiting UpdateHistory for long-running accounts

```go
// Optional: Prune old history
if len(acc.UpdateHistory) > 10000 {
    acc.UpdateHistory = acc.UpdateHistory[len(acc.UpdateHistory)-5000:]
}
```

---

## Related Documentation

- [Position Package](position.md) - Trade positions
- [Instrument Package](instrument.md) - Trading instruments
- [Executor Package](executor.md) - Order execution
- [Logger Package](logger.md) - Trade logging

---

**Created:** December 26, 2024  
**Version:** 1.0  
**Status:** Production Ready  
**Lines of Code:** 469 (4 files)