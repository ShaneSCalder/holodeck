package types

import (
	"fmt"
	"time"
)

// ==================== BALANCE STRUCTURE ====================

// Balance represents the account's financial state
// Tracks equity, margin, P&L, and account status
type Balance struct {
	// InitialBalance is the starting account balance
	InitialBalance float64

	// CurrentBalance is the current account equity
	// Calculated as: InitialBalance + TotalPnL - CommissionPaid
	CurrentBalance float64

	// Currency is the account currency (USD, EUR, etc)
	Currency string

	// TotalRealizedPnL is profit/loss from closed trades
	TotalRealizedPnL float64

	// TotalUnrealizedPnL is profit/loss from open positions (mark-to-market)
	TotalUnrealizedPnL float64

	// CommissionPaid is the total fees/commissions paid
	CommissionPaid float64

	// Leverage is the account leverage multiplier (1.0 = no leverage)
	Leverage float64

	// UsedMargin is the margin in use by open positions
	UsedMargin float64

	// AvailableMargin is the margin available for new trades
	AvailableMargin float64

	// BuyingPower is the total amount that can be traded (balance * leverage)
	BuyingPower float64

	// MaxDrawdownPercent is the maximum allowed drawdown before account blown
	MaxDrawdownPercent float64

	// MaxPositionSize is the maximum size allowed per position
	MaxPositionSize float64

	// TradeCount is the total number of trades executed
	TradeCount int

	// WinningTrades is the count of profitable trades
	WinningTrades int

	// LosingTrades is the count of losing trades
	LosingTrades int

	// BreakevenTrades is the count of trades with 0 P&L
	BreakevenTrades int

	// AccountStatus is ACTIVE, BLOWN, or AT_LIMIT
	AccountStatus string

	// LastUpdateTime is when balance was last updated
	LastUpdateTime time.Time

	// HighWaterMark is the highest balance reached
	HighWaterMark float64

	// LowWaterMark is the lowest balance reached
	LowWaterMark float64

	// MaxDrawdown is the largest peak-to-trough drawdown experienced
	MaxDrawdownExperienced float64

	// StartTime is when the account was opened
	StartTime time.Time

	// UpdateHistory tracks balance changes over time
	UpdateHistory []*BalanceUpdate
}

// ==================== BALANCE UPDATE RECORD ====================

// BalanceUpdate records a balance change event
type BalanceUpdate struct {
	// Timestamp of the update
	Timestamp time.Time

	// BalanceBefore is the balance before this update
	BalanceBefore float64

	// BalanceAfter is the balance after this update
	BalanceAfter float64

	// Change is the net change
	Change float64

	// Reason describes why balance changed (trade, commission, etc)
	Reason string

	// OrderID is the order that caused this update (if applicable)
	OrderID string

	// ReferencePnL is the P&L that caused the change
	ReferencePnL float64
}

// ==================== BALANCE CONSTRUCTORS ====================

// NewBalance creates a new balance account
func NewBalance(initialBalance float64, currency string, leverage, maxDrawdown, maxPositionSize float64) *Balance {
	now := time.Now()
	return &Balance{
		InitialBalance:     initialBalance,
		CurrentBalance:     initialBalance,
		Currency:           currency,
		Leverage:           leverage,
		MaxDrawdownPercent: maxDrawdown,
		MaxPositionSize:    maxPositionSize,
		AccountStatus:      AccountStatusActive,
		LastUpdateTime:     now,
		StartTime:          now,
		HighWaterMark:      initialBalance,
		LowWaterMark:       initialBalance,
		BuyingPower:        initialBalance * leverage,
		UpdateHistory:      make([]*BalanceUpdate, 0),
	}
}

// ==================== BALANCE QUERIES ====================

// GetTotalPnL returns realized + unrealized P&L
func (b *Balance) GetTotalPnL() float64 {
	return b.TotalRealizedPnL + b.TotalUnrealizedPnL
}

// GetNetPnL returns total P&L minus commissions
func (b *Balance) GetNetPnL() float64 {
	return b.GetTotalPnL() - b.CommissionPaid
}

// IsAccountActive returns true if account status is ACTIVE
func (b *Balance) IsAccountActive() bool {
	return b.AccountStatus == AccountStatusActive
}

// IsAccountBlown returns true if account is blown
func (b *Balance) IsAccountBlown() bool {
	return b.AccountStatus == AccountStatusBlown
}

// IsAccountAtLimit returns true if at drawdown limit
func (b *Balance) IsAccountAtLimit() bool {
	return b.AccountStatus == AccountStatusAtLimit
}

// GetDrawdownPercent returns current drawdown as percentage
func (b *Balance) GetDrawdownPercent() float64 {
	if b.InitialBalance == 0 {
		return 0
	}
	return ((b.InitialBalance - b.CurrentBalance) / b.InitialBalance) * 100.0
}

// GetReturnPercent returns total return as percentage
func (b *Balance) GetReturnPercent() float64 {
	if b.InitialBalance == 0 {
		return 0
	}
	return ((b.CurrentBalance - b.InitialBalance) / b.InitialBalance) * 100.0
}

// GetWinRate returns winning trades as percentage
func (b *Balance) GetWinRate() float64 {
	totalTrades := b.WinningTrades + b.LosingTrades
	if totalTrades == 0 {
		return 0
	}
	return (float64(b.WinningTrades) / float64(totalTrades)) * 100.0
}

// GetAverageTradePnL returns average P&L per trade
func (b *Balance) GetAverageTradePnL() float64 {
	if b.TradeCount == 0 {
		return 0
	}
	return b.TotalRealizedPnL / float64(b.TradeCount)
}

// GetProfitFactor returns profit factor (gross profits / gross losses)
func (b *Balance) GetProfitFactor() float64 {
	if b.WinningTrades == 0 || b.LosingTrades == 0 {
		return 0
	}
	grossWins := (b.GetAverageTradePnL() * float64(b.WinningTrades))
	grossLosses := (b.GetAverageTradePnL() * float64(b.LosingTrades))

	if grossLosses == 0 {
		return 0
	}
	return grossWins / -grossLosses
}

// GetSharpeRatio is a simplified sharpe ratio approximation
// (Real Sharpe needs daily returns, this is simplified)
func (b *Balance) GetSharpeRatio() float64 {
	if b.TradeCount == 0 {
		return 0
	}

	avgReturn := b.GetAverageTradePnL()
	// Simplified: assume standard deviation is 20% of avg trade
	stdDev := (avgReturn * 0.2)
	if stdDev == 0 {
		return 0
	}

	// Annualize (assuming ~250 trades per year)
	return (avgReturn / stdDev) * (250.0 / float64(b.TradeCount))
}

// IsMarginCall returns true if margin is violated
func (b *Balance) IsMarginCall() bool {
	return b.AvailableMargin < 0
}

// CanTrade returns true if account can trade (active and has margin)
func (b *Balance) CanTrade() bool {
	return b.IsAccountActive() && b.AvailableMargin > 0
}

// ==================== BALANCE UPDATE METHODS ====================

// UpdateFromExecution updates balance from an execution report
func (b *Balance) UpdateFromExecution(report *ExecutionReport) error {
	if report == nil {
		return fmt.Errorf("execution report cannot be nil")
	}

	if report.IsRejected() {
		return nil // No balance change on rejection
	}

	// Calculate P&L change
	pnlChange := 0.0
	if report.IsSell() && report.RealizedPnL > 0 {
		// Position closing, add realized P&L
		pnlChange = report.RealizedPnL
		b.TotalRealizedPnL += report.RealizedPnL
	}

	// Add unrealized P&L from open position
	if report.IsPartial() || (report.IsFilled() && report.PositionAfter != 0) {
		b.TotalUnrealizedPnL = report.UnrealizedPnL
	}

	// Add commission
	b.CommissionPaid += report.Commission

	// Update trade counts
	if report.IsFilled() || report.IsPartial() {
		b.TradeCount++
		if report.RealizedPnL > 0 {
			b.WinningTrades++
		} else if report.RealizedPnL < 0 {
			b.LosingTrades++
		} else if report.RealizedPnL == 0 && report.IsSell() {
			b.BreakevenTrades++
		}
	}

	// Recalculate balance
	b.RecalculateBalance()

	// Record update
	b.recordUpdate(
		fmt.Sprintf("Execution %s", report.OrderID),
		report.OrderID,
		pnlChange,
	)

	return nil
}

// UpdateMargin updates the margin calculations
func (b *Balance) UpdateMargin(usedMargin float64) {
	b.UsedMargin = usedMargin
	b.AvailableMargin = (b.CurrentBalance * b.Leverage) - b.UsedMargin
	b.BuyingPower = b.AvailableMargin
}

// RecalculateBalance recalculates the current balance
func (b *Balance) RecalculateBalance() {
	newBalance := b.InitialBalance + b.GetNetPnL()
	b.CurrentBalance = newBalance

	// Update high/low water marks
	if b.CurrentBalance > b.HighWaterMark {
		b.HighWaterMark = b.CurrentBalance
	}
	if b.CurrentBalance < b.LowWaterMark {
		b.LowWaterMark = b.CurrentBalance
	}

	// Update max drawdown experienced
	currentDrawdown := b.GetDrawdownPercent()
	if currentDrawdown > b.MaxDrawdownExperienced {
		b.MaxDrawdownExperienced = currentDrawdown
	}

	// Update account status based on drawdown
	b.updateAccountStatus()

	// Update buying power
	b.BuyingPower = b.CurrentBalance * b.Leverage

	// Update last update time
	b.LastUpdateTime = time.Now()
}

// updateAccountStatus updates the account status based on drawdown
func (b *Balance) updateAccountStatus() {
	currentDrawdown := b.GetDrawdownPercent()

	if currentDrawdown > b.MaxDrawdownPercent {
		b.AccountStatus = AccountStatusBlown
	} else if currentDrawdown >= (b.MaxDrawdownPercent * 0.95) {
		// Within 5% of limit
		b.AccountStatus = AccountStatusAtLimit
	} else {
		b.AccountStatus = AccountStatusActive
	}
}

// recordUpdate records a balance update event
func (b *Balance) recordUpdate(reason, orderID string, pnlChange float64) {
	balanceBefore := b.CurrentBalance - pnlChange
	update := &BalanceUpdate{
		Timestamp:     b.LastUpdateTime,
		BalanceBefore: balanceBefore,
		BalanceAfter:  b.CurrentBalance,
		Change:        pnlChange,
		Reason:        reason,
		OrderID:       orderID,
		ReferencePnL:  pnlChange,
	}
	b.UpdateHistory = append(b.UpdateHistory, update)
}

// ==================== BALANCE METRICS ====================

// GetMetrics returns comprehensive balance metrics
func (b *Balance) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"initial_balance":          b.InitialBalance,
		"current_balance":          b.CurrentBalance,
		"currency":                 b.Currency,
		"total_pnl":                b.GetTotalPnL(),
		"realized_pnl":             b.TotalRealizedPnL,
		"unrealized_pnl":           b.TotalUnrealizedPnL,
		"net_pnl":                  b.GetNetPnL(),
		"commission_paid":          b.CommissionPaid,
		"return_percent":           b.GetReturnPercent(),
		"drawdown_percent":         b.GetDrawdownPercent(),
		"max_drawdown_percent":     b.MaxDrawdownPercent,
		"max_drawdown_experienced": b.MaxDrawdownExperienced,
		"leverage":                 b.Leverage,
		"used_margin":              b.UsedMargin,
		"available_margin":         b.AvailableMargin,
		"buying_power":             b.BuyingPower,
		"trade_count":              b.TradeCount,
		"winning_trades":           b.WinningTrades,
		"losing_trades":            b.LosingTrades,
		"breakeven_trades":         b.BreakevenTrades,
		"win_rate":                 b.GetWinRate(),
		"avg_trade_pnl":            b.GetAverageTradePnL(),
		"profit_factor":            b.GetProfitFactor(),
		"sharpe_ratio":             b.GetSharpeRatio(),
		"account_status":           b.AccountStatus,
		"high_water_mark":          b.HighWaterMark,
		"low_water_mark":           b.LowWaterMark,
		"last_update_time":         b.LastUpdateTime,
		"session_duration":         time.Since(b.StartTime),
	}
}

// ==================== BALANCE DISPLAY ====================

// String returns a human-readable representation
func (b *Balance) String() string {
	return fmt.Sprintf(
		"Balance[%s %.2f | Return: %.2f%% | Drawdown: %.2f%% | Status: %s]",
		b.Currency,
		b.CurrentBalance,
		b.GetReturnPercent(),
		b.GetDrawdownPercent(),
		b.AccountStatus,
	)
}

// DebugString returns detailed balance information
func (b *Balance) DebugString() string {
	sessionDuration := time.Since(b.StartTime)

	return fmt.Sprintf(
		"Balance Details:\n"+
			"  Currency:              %s\n"+
			"  Initial Balance:       %.2f\n"+
			"  Current Balance:       %.2f\n"+
			"  Account Status:        %s\n"+
			"\n"+
			"  P&L:\n"+
			"    Realized:            %.2f\n"+
			"    Unrealized:          %.2f\n"+
			"    Total:               %.2f\n"+
			"    Net (After Comm):    %.2f\n"+
			"    Commission Paid:     %.2f\n"+
			"\n"+
			"  Returns:\n"+
			"    Return %%:            %.2f%%\n"+
			"    Drawdown %%:          %.2f%%\n"+
			"    Max Drawdown %%:      %.2f%%\n"+
			"    Max Drawdown Exp:    %.2f%%\n"+
			"\n"+
			"  Margin & Leverage:\n"+
			"    Leverage:            %.2fx\n"+
			"    Buying Power:        %.2f\n"+
			"    Used Margin:         %.2f\n"+
			"    Available Margin:    %.2f\n"+
			"    Margin Call:         %v\n"+
			"\n"+
			"  Trading Statistics:\n"+
			"    Total Trades:        %d\n"+
			"    Winning:             %d\n"+
			"    Losing:              %d\n"+
			"    Breakeven:           %d\n"+
			"    Win Rate:            %.2f%%\n"+
			"    Avg Trade P&L:       %.2f\n"+
			"    Profit Factor:       %.2f\n"+
			"    Sharpe Ratio:        %.2f\n"+
			"\n"+
			"  Water Marks:\n"+
			"    High Water Mark:     %.2f\n"+
			"    Low Water Mark:      %.2f\n"+
			"\n"+
			"  Timing:\n"+
			"    Session Duration:    %v\n"+
			"    Last Updated:        %s",
		b.Currency,
		b.InitialBalance,
		b.CurrentBalance,
		b.AccountStatus,
		b.TotalRealizedPnL,
		b.TotalUnrealizedPnL,
		b.GetTotalPnL(),
		b.GetNetPnL(),
		b.CommissionPaid,
		b.GetReturnPercent(),
		b.GetDrawdownPercent(),
		b.MaxDrawdownPercent,
		b.MaxDrawdownExperienced,
		b.Leverage,
		b.BuyingPower,
		b.UsedMargin,
		b.AvailableMargin,
		b.IsMarginCall(),
		b.TradeCount,
		b.WinningTrades,
		b.LosingTrades,
		b.BreakevenTrades,
		b.GetWinRate(),
		b.GetAverageTradePnL(),
		b.GetProfitFactor(),
		b.GetSharpeRatio(),
		b.HighWaterMark,
		b.LowWaterMark,
		sessionDuration,
		b.LastUpdateTime.Format("2006-01-02T15:04:05.000"),
	)
}

// ==================== BALANCE HISTORY ====================

// GetUpdateHistory returns all balance updates within a timeframe
func (b *Balance) GetUpdateHistory(limit int) []*BalanceUpdate {
	if limit <= 0 || limit > len(b.UpdateHistory) {
		return b.UpdateHistory
	}
	return b.UpdateHistory[len(b.UpdateHistory)-limit:]
}

// GetLastUpdate returns the most recent balance update
func (b *Balance) GetLastUpdate() *BalanceUpdate {
	if len(b.UpdateHistory) == 0 {
		return nil
	}
	return b.UpdateHistory[len(b.UpdateHistory)-1]
}

// ==================== BALANCE COMPARISON ====================

// CompareToBaseline compares current balance to baseline
type BalanceComparison struct {
	InitialDifference  float64
	CurrentDifference  float64
	ReturnDifference   float64
	DrawdownDifference float64
}

// CompareTo compares this balance to another
func (b *Balance) CompareTo(other *Balance) *BalanceComparison {
	return &BalanceComparison{
		InitialDifference:  b.InitialBalance - other.InitialBalance,
		CurrentDifference:  b.CurrentBalance - other.CurrentBalance,
		ReturnDifference:   b.GetReturnPercent() - other.GetReturnPercent(),
		DrawdownDifference: b.GetDrawdownPercent() - other.GetDrawdownPercent(),
	}
}

// ==================== BALANCE RESET ====================

// Reset resets the balance to initial state
func (b *Balance) Reset() {
	b.CurrentBalance = b.InitialBalance
	b.TotalRealizedPnL = 0
	b.TotalUnrealizedPnL = 0
	b.CommissionPaid = 0
	b.TradeCount = 0
	b.WinningTrades = 0
	b.LosingTrades = 0
	b.BreakevenTrades = 0
	b.AccountStatus = AccountStatusActive
	b.HighWaterMark = b.InitialBalance
	b.LowWaterMark = b.InitialBalance
	b.MaxDrawdownExperienced = 0
	b.UpdateHistory = make([]*BalanceUpdate, 0)
	b.StartTime = time.Now()
	b.LastUpdateTime = time.Now()
	b.BuyingPower = b.InitialBalance * b.Leverage
}
