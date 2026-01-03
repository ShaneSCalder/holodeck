package logger

import (
	"fmt"
	"math"
	"time"
)

// ==================== METRICS CALCULATOR ====================

// MetricsCalculator calculates performance metrics
type MetricsCalculator struct {
	initialBalance float64
	tradeLogger    *TradeLogger
	startTime      time.Time
}

// ==================== CREATION ====================

// NewMetricsCalculator creates a new metrics calculator
func NewMetricsCalculator(initialBalance float64, tradeLogger *TradeLogger) *MetricsCalculator {
	return &MetricsCalculator{
		initialBalance: initialBalance,
		tradeLogger:    tradeLogger,
		startTime:      time.Now(),
	}
}

// ==================== CALCULATION METHODS ====================

// CalculateMetrics calculates all metrics and returns MetricsLog
func (mc *MetricsCalculator) CalculateMetrics(
	sessionID string,
	currentBalance float64,
	ticksProcessed int64,
	errorCount int64,
	rejectedOrders int64,
) *MetricsLog {

	trades := mc.tradeLogger.GetTrades()
	totalTrades := int64(len(trades))

	totalPnL := currentBalance - mc.initialBalance
	totalPnLPercent := 0.0
	if mc.initialBalance > 0 {
		totalPnLPercent = (totalPnL / mc.initialBalance) * 100
	}

	winningTrades := mc.tradeLogger.GetWinningTrades()
	losingTrades := mc.tradeLogger.GetLosingTrades()

	winRate := 0.0
	if totalTrades > 0 {
		winRate = (float64(winningTrades) / float64(totalTrades)) * 100
	}

	maxDrawdown, maxDrawdownPercent := mc.CalculateMaxDrawdown()
	avgTradePnL := mc.CalculateAverageTradePnL()
	largestWin := mc.tradeLogger.GetLargestWin()
	largestLoss := mc.tradeLogger.GetLargestLoss()

	ratios := mc.tradeLogger.AnalyzeWinLossRatio()
	meanWin := ratios["average_win"]
	meanLoss := ratios["average_loss"]
	profitFactor := mc.tradeLogger.GetProfitFactor()

	sharpeRatio := mc.CalculateSharpeRatio()
	avgHoldTime := mc.CalculateAverageHoldTime()

	commissionTotal := mc.CalculateTotalCommission()
	slippageTotal := mc.CalculateTotalSlippage()

	return &MetricsLog{
		Timestamp:          time.Now(),
		SessionID:          sessionID,
		SessionDuration:    time.Since(mc.startTime),
		InitialBalance:     mc.initialBalance,
		CurrentBalance:     currentBalance,
		TotalPnL:           totalPnL,
		TotalPnLPercent:    totalPnLPercent,
		TradeCount:         totalTrades,
		WinningTrades:      winningTrades,
		LosingTrades:       losingTrades,
		WinRate:            winRate,
		MaxDrawdown:        maxDrawdown,
		MaxDrawdownPercent: maxDrawdownPercent,
		CommissionTotal:    commissionTotal,
		SlippageTotal:      slippageTotal,
		AverageTradePnL:    avgTradePnL,
		LargestWin:         largestWin,
		LargestLoss:        largestLoss,
		MeanWin:            meanWin,
		MeanLoss:           meanLoss,
		ProfitFactor:       profitFactor,
		SharpeRatio:        sharpeRatio,
		MDD:                maxDrawdown,
		MWL:                mc.tradeLogger.GetMaxWinStreak(),
		MLS:                mc.tradeLogger.GetMaxLoseStreak(),
		AvgHoldTime:        avgHoldTime,
		TicksProcessed:     ticksProcessed,
		ErrorCount:         errorCount,
		RejectedOrders:     rejectedOrders,
	}
}

// ==================== INDIVIDUAL METRIC CALCULATIONS ====================

// CalculateMaxDrawdown calculates maximum drawdown
func (mc *MetricsCalculator) CalculateMaxDrawdown() (float64, float64) {
	trades := mc.tradeLogger.GetTrades()
	if len(trades) == 0 {
		return 0, 0
	}

	runningBalance := mc.initialBalance
	peakBalance := mc.initialBalance
	maxDrawdown := 0.0
	maxDrawdownPercent := 0.0

	for _, trade := range trades {
		runningBalance += trade.RealizedPnL

		if runningBalance > peakBalance {
			peakBalance = runningBalance
		}

		drawdown := peakBalance - runningBalance
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
			drawdownPercent := (drawdown / peakBalance) * 100
			if drawdownPercent > maxDrawdownPercent {
				maxDrawdownPercent = drawdownPercent
			}
		}
	}

	return maxDrawdown, maxDrawdownPercent
}

// CalculateAverageTradePnL calculates average P&L per trade
func (mc *MetricsCalculator) CalculateAverageTradePnL() float64 {
	trades := mc.tradeLogger.GetTrades()
	if len(trades) == 0 {
		return 0
	}

	totalPnL := 0.0
	for _, trade := range trades {
		totalPnL += trade.RealizedPnL
	}

	return totalPnL / float64(len(trades))
}

// CalculateTotalCommission calculates total commission paid
func (mc *MetricsCalculator) CalculateTotalCommission() float64 {
	trades := mc.tradeLogger.GetTrades()
	totalCommission := 0.0

	for _, trade := range trades {
		totalCommission += trade.Commission
	}

	return totalCommission
}

// CalculateTotalSlippage calculates total slippage in value
func (mc *MetricsCalculator) CalculateTotalSlippage() float64 {
	trades := mc.tradeLogger.GetTrades()
	totalSlippage := 0.0

	for _, trade := range trades {
		// Convert slippage pips to value (simplified)
		totalSlippage += trade.Slippage * trade.FillPrice * trade.FilledSize * 0.0001
	}

	return totalSlippage
}

// CalculateSharpeRatio calculates Sharpe ratio
func (mc *MetricsCalculator) CalculateSharpeRatio() float64 {
	trades := mc.tradeLogger.GetTrades()
	if len(trades) < 2 {
		return 0
	}

	// Calculate returns
	returns := make([]float64, len(trades))
	for i, trade := range trades {
		if mc.initialBalance > 0 {
			returns[i] = trade.RealizedPnL / mc.initialBalance
		}
	}

	// Calculate mean return
	meanReturn := 0.0
	for _, r := range returns {
		meanReturn += r
	}
	meanReturn /= float64(len(returns))

	// Calculate standard deviation
	variance := 0.0
	for _, r := range returns {
		diff := r - meanReturn
		variance += diff * diff
	}
	variance /= float64(len(returns))
	stdDev := math.Sqrt(variance)

	// Sharpe ratio = mean return / std dev (assuming risk-free rate = 0)
	if stdDev > 0 {
		return meanReturn / stdDev
	}
	return 0
}

// CalculateAverageHoldTime calculates average holding time per trade
func (mc *MetricsCalculator) CalculateAverageHoldTime() time.Duration {
	trades := mc.tradeLogger.GetTrades()
	if len(trades) == 0 {
		return 0
	}

	// Simple approximation: assume trades are evenly spaced
	if len(trades) == 1 {
		return time.Since(trades[0].Timestamp)
	}

	firstTrade := trades[0].Timestamp
	lastTrade := trades[len(trades)-1].Timestamp
	totalTime := lastTrade.Sub(firstTrade)

	return totalTime / time.Duration(len(trades)-1)
}

// ==================== RETURN CALCULATIONS ====================

// CalculateCumulativeReturn calculates cumulative return
func (mc *MetricsCalculator) CalculateCumulativeReturn(finalBalance float64) float64 {
	if mc.initialBalance <= 0 {
		return 0
	}
	return ((finalBalance - mc.initialBalance) / mc.initialBalance) * 100
}

// CalculateMonthlyReturn calculates monthly return (simplified)
func (mc *MetricsCalculator) CalculateMonthlyReturn(finalBalance float64) float64 {
	dailyReturn := mc.CalculateCumulativeReturn(finalBalance) / 100
	months := time.Since(mc.startTime).Hours() / 24 / 30

	if months <= 0 {
		return 0
	}

	return dailyReturn / months * 100
}

// ==================== RISK METRICS ====================

// CalculateRiskRewardRatio calculates risk/reward ratio
func (mc *MetricsCalculator) CalculateRiskRewardRatio() float64 {
	avgWinTrades := mc.tradeLogger.GetWinningTrades()
	avgLossTrades := mc.tradeLogger.GetLosingTrades()

	if avgWinTrades == 0 || avgLossTrades == 0 {
		return 0
	}

	ratios := mc.tradeLogger.AnalyzeWinLossRatio()
	avgWin := ratios["average_win"]
	avgLoss := ratios["average_loss"]

	if avgLoss <= 0 {
		return 0
	}

	return avgWin / avgLoss
}

// CalculateRecoveryFactor calculates recovery factor
func (mc *MetricsCalculator) CalculateRecoveryFactor(finalBalance float64) float64 {
	totalPnL := finalBalance - mc.initialBalance
	maxDrawdown, _ := mc.CalculateMaxDrawdown()

	if maxDrawdown <= 0 {
		return 0
	}

	return totalPnL / maxDrawdown
}

// ==================== STRING REPRESENTATIONS ====================

// GetMetricsString returns formatted metrics string
func (mc *MetricsCalculator) GetMetricsString(finalBalance float64) string {
	maxDrawdown, maxDrawdownPct := mc.CalculateMaxDrawdown()
	sharpeRatio := mc.CalculateSharpeRatio()
	riskRewardRatio := mc.CalculateRiskRewardRatio()
	recoveryFactor := mc.CalculateRecoveryFactor(finalBalance)
	cumulativeReturn := mc.CalculateCumulativeReturn(finalBalance)

	return fmt.Sprintf(
		"=== PERFORMANCE METRICS ===\n"+
			"Cumulative Return:      %.2f%%\n"+
			"Sharpe Ratio:           %.2f\n"+
			"Max Drawdown:           $%.2f (%.2f%%)\n"+
			"Risk/Reward Ratio:      %.2f\n"+
			"Recovery Factor:        %.2f\n"+
			"Session Duration:       %v\n",
		cumulativeReturn,
		sharpeRatio,
		maxDrawdown,
		maxDrawdownPct,
		riskRewardRatio,
		recoveryFactor,
		time.Since(mc.startTime),
	)
}

// ==================== PERFORMANCE RATING ====================

// RatePerformance provides overall performance rating
func (mc *MetricsCalculator) RatePerformance(finalBalance float64) string {
	cumReturn := mc.CalculateCumulativeReturn(finalBalance)
	sharpeRatio := mc.CalculateSharpeRatio()
	winRate := mc.tradeLogger.GetWinRate()
	profitFactor := mc.tradeLogger.GetProfitFactor()

	score := 0.0

	// Score based on return
	if cumReturn > 20 {
		score += 25
	} else if cumReturn > 10 {
		score += 20
	} else if cumReturn > 0 {
		score += 15
	}

	// Score based on Sharpe ratio
	if sharpeRatio > 2 {
		score += 25
	} else if sharpeRatio > 1 {
		score += 20
	} else if sharpeRatio > 0 {
		score += 10
	}

	// Score based on win rate
	if winRate > 60 {
		score += 25
	} else if winRate > 50 {
		score += 15
	} else if winRate > 40 {
		score += 5
	}

	// Score based on profit factor
	if profitFactor > 2 {
		score += 25
	} else if profitFactor > 1.5 {
		score += 20
	} else if profitFactor > 1 {
		score += 10
	}

	// Return rating based on score
	switch {
	case score >= 90:
		return "EXCELLENT"
	case score >= 75:
		return "VERY GOOD"
	case score >= 60:
		return "GOOD"
	case score >= 45:
		return "FAIR"
	case score >= 30:
		return "POOR"
	default:
		return "VERY POOR"
	}
}
