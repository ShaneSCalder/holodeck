package account

import (
	"fmt"
)

// ==================== BALANCE OPERATIONS ====================

// RecordTrade records a trade's P&L impact
func (a *Account) RecordTrade(tradeID string, pnl float64, commission float64) {
	oldBalance := a.CurrentBalance

	// Update balance
	a.CurrentBalance += pnl
	a.CurrentBalance -= commission

	// Update P&L
	if pnl > 0 {
		a.TotalRealizedPnL += pnl
		a.WinningTrades++
		a.ConsecutiveWins++
		a.ConsecutiveLosses = 0
	} else if pnl < 0 {
		a.TotalRealizedPnL += pnl
		a.LosingTrades++
		a.ConsecutiveLosses++
		a.ConsecutiveWins = 0
	} else {
		a.BreakevenTrades++
		a.ConsecutiveWins = 0
		a.ConsecutiveLosses = 0
	}

	a.CommissionPaid += commission
	a.TotalTrades++

	// Update high/low watermarks
	if a.CurrentBalance > a.HighWaterMark {
		a.HighWaterMark = a.CurrentBalance
	}
	if a.CurrentBalance < a.LowWaterMark {
		a.LowWaterMark = a.CurrentBalance
	}

	// Record update
	a.RecordBalanceUpdate(oldBalance, a.CurrentBalance, pnl-commission,
		fmt.Sprintf("Trade P&L: %+.2f", pnl), tradeID)
}

// RecordUnrealizedPnL records mark-to-market updates
func (a *Account) RecordUnrealizedPnL(unrealizedPnL float64) {
	oldBalance := a.CurrentBalance
	a.TotalUnrealizedPnL = unrealizedPnL

	// For account equity calculation
	equityWithUnrealized := a.CurrentBalance + unrealizedPnL
	a.RecordBalanceUpdate(oldBalance, equityWithUnrealized, unrealizedPnL,
		"Unrealized P&L", "")
}

// RecordCommission records a commission deduction
func (a *Account) RecordCommission(transactionID string, amount float64) {
	oldBalance := a.CurrentBalance
	a.CurrentBalance -= amount
	a.CommissionPaid += amount

	if a.CurrentBalance < a.LowWaterMark {
		a.LowWaterMark = a.CurrentBalance
	}

	a.RecordBalanceUpdate(oldBalance, a.CurrentBalance, -amount,
		"Commission", transactionID)
}

// ==================== STATISTICS ====================

// GetWinRate returns winning trades percentage
func (a *Account) GetWinRate() float64 {
	if a.TotalTrades == 0 {
		return 0
	}
	return (float64(a.WinningTrades) / float64(a.TotalTrades)) * 100
}

// GetLossRate returns losing trades percentage
func (a *Account) GetLossRate() float64 {
	if a.TotalTrades == 0 {
		return 0
	}
	return (float64(a.LosingTrades) / float64(a.TotalTrades)) * 100
}

// GetBreakevenRate returns breakeven trades percentage
func (a *Account) GetBreakevenRate() float64 {
	if a.TotalTrades == 0 {
		return 0
	}
	return (float64(a.BreakevenTrades) / float64(a.TotalTrades)) * 100
}

// GetTotalReturn returns total return percentage
func (a *Account) GetTotalReturn() float64 {
	if a.InitialBalance == 0 {
		return 0
	}
	profit := a.CurrentBalance - a.InitialBalance
	return (profit / a.InitialBalance) * 100
}

// GetRiskRewardRatio calculates risk/reward ratio
func (a *Account) GetRiskRewardRatio() float64 {
	if a.LosingTrades == 0 {
		return 0
	}
	avgWin := a.TotalRealizedPnL / float64(a.WinningTrades)
	avgLoss := a.TotalRealizedPnL / float64(a.LosingTrades)
	if avgLoss == 0 {
		return 0
	}
	return -avgWin / avgLoss
}

// GetProfitFactor returns profit factor (total wins / total losses)
func (a *Account) GetProfitFactor() float64 {
	if a.TotalRealizedPnL == 0 {
		return 0
	}
	grossProfit := a.TotalRealizedPnL
	if a.LosingTrades > 0 {
		return grossProfit / (-a.TotalRealizedPnL)
	}
	return 0
}
