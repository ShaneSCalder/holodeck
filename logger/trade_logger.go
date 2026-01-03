package logger

import (
	"fmt"
	"sync"
	"time"
)

// ==================== TRADE LOGGER ====================

// TradeLogger tracks and logs all trade-related events
type TradeLogger struct {
	logger Logger

	// Trade tracking
	trades      []*TradeLog
	tradesMutex sync.RWMutex

	// Statistics
	totalTrades     int64
	winningTrades   int64
	losingTrades    int64
	breakEvenTrades int64
	totalWinAmount  float64
	totalLossAmount float64
	largestWin      float64
	largestLoss     float64
	successRate     float64
	profitFactor    float64

	// Streaks
	currentWinStreak  int64
	currentLoseStreak int64
	maxWinStreak      int64
	maxLoseStreak     int64

	// Timing
	createdAt time.Time
}

// ==================== CREATION ====================

// NewTradeLogger creates a new trade logger
func NewTradeLogger(baseLogger Logger) *TradeLogger {
	return &TradeLogger{
		logger:    baseLogger,
		trades:    make([]*TradeLog, 0),
		createdAt: time.Now(),
	}
}

// ==================== LOGGING ====================

// LogTrade logs a trade and updates statistics
func (tl *TradeLogger) LogTrade(trade *TradeLog) error {
	// Log to underlying logger
	if err := tl.logger.LogTrade(trade); err != nil {
		return err
	}

	// Track trade
	tl.tradesMutex.Lock()
	tl.trades = append(tl.trades, trade)
	tl.tradesMutex.Unlock()

	// Update statistics
	tl.updateStatistics(trade)

	return nil
}

// ==================== STATISTICS UPDATES ====================

// updateStatistics updates all trade statistics
func (tl *TradeLogger) updateStatistics(trade *TradeLog) {
	tl.totalTrades++

	// P&L classification
	if trade.RealizedPnL > 0 {
		tl.winningTrades++
		tl.totalWinAmount += trade.RealizedPnL
		if trade.RealizedPnL > tl.largestWin {
			tl.largestWin = trade.RealizedPnL
		}
		// Win streak
		tl.currentWinStreak++
		tl.currentLoseStreak = 0
		if tl.currentWinStreak > tl.maxWinStreak {
			tl.maxWinStreak = tl.currentWinStreak
		}
	} else if trade.RealizedPnL < 0 {
		tl.losingTrades++
		tl.totalLossAmount += trade.RealizedPnL
		if trade.RealizedPnL < tl.largestLoss {
			tl.largestLoss = trade.RealizedPnL
		}
		// Lose streak
		tl.currentLoseStreak++
		tl.currentWinStreak = 0
		if tl.currentLoseStreak > tl.maxLoseStreak {
			tl.maxLoseStreak = tl.currentLoseStreak
		}
	} else {
		tl.breakEvenTrades++
	}

	// Calculate ratios
	if tl.totalTrades > 0 {
		tl.successRate = float64(tl.winningTrades) / float64(tl.totalTrades)
	}

	// Calculate profit factor
	if tl.totalLossAmount != 0 {
		tl.profitFactor = -tl.totalWinAmount / tl.totalLossAmount
	}
}

// ==================== QUERY METHODS ====================

// GetTotalTrades returns total number of trades
func (tl *TradeLogger) GetTotalTrades() int64 {
	return tl.totalTrades
}

// GetWinningTrades returns number of winning trades
func (tl *TradeLogger) GetWinningTrades() int64 {
	return tl.winningTrades
}

// GetLosingTrades returns number of losing trades
func (tl *TradeLogger) GetLosingTrades() int64 {
	return tl.losingTrades
}

// GetBreakEvenTrades returns number of break-even trades
func (tl *TradeLogger) GetBreakEvenTrades() int64 {
	return tl.breakEvenTrades
}

// GetWinRate returns win rate as percentage
func (tl *TradeLogger) GetWinRate() float64 {
	return tl.successRate * 100
}

// GetProfitFactor returns profit factor
func (tl *TradeLogger) GetProfitFactor() float64 {
	return tl.profitFactor
}

// GetLargestWin returns largest win amount
func (tl *TradeLogger) GetLargestWin() float64 {
	return tl.largestWin
}

// GetLargestLoss returns largest loss amount
func (tl *TradeLogger) GetLargestLoss() float64 {
	return tl.largestLoss
}

// GetTotalWins returns total winning amount
func (tl *TradeLogger) GetTotalWins() float64 {
	return tl.totalWinAmount
}

// GetTotalLosses returns total losing amount
func (tl *TradeLogger) GetTotalLosses() float64 {
	return tl.totalLossAmount
}

// GetMaxWinStreak returns maximum winning streak
func (tl *TradeLogger) GetMaxWinStreak() int64 {
	return tl.maxWinStreak
}

// GetMaxLoseStreak returns maximum losing streak
func (tl *TradeLogger) GetMaxLoseStreak() int64 {
	return tl.maxLoseStreak
}

// GetCurrentWinStreak returns current winning streak
func (tl *TradeLogger) GetCurrentWinStreak() int64 {
	return tl.currentWinStreak
}

// GetCurrentLoseStreak returns current losing streak
func (tl *TradeLogger) GetCurrentLoseStreak() int64 {
	return tl.currentLoseStreak
}

// GetTrades returns slice of all logged trades
func (tl *TradeLogger) GetTrades() []*TradeLog {
	tl.tradesMutex.RLock()
	defer tl.tradesMutex.RUnlock()

	trades := make([]*TradeLog, len(tl.trades))
	copy(trades, tl.trades)
	return trades
}

// ==================== STATISTICS ====================

// GetStatistics returns comprehensive trade statistics
func (tl *TradeLogger) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"total_trades":        tl.totalTrades,
		"winning_trades":      tl.winningTrades,
		"losing_trades":       tl.losingTrades,
		"break_even_trades":   tl.breakEvenTrades,
		"win_rate":            tl.GetWinRate(),
		"profit_factor":       tl.profitFactor,
		"largest_win":         tl.largestWin,
		"largest_loss":        tl.largestLoss,
		"total_wins":          tl.totalWinAmount,
		"total_losses":        tl.totalLossAmount,
		"max_win_streak":      tl.maxWinStreak,
		"max_lose_streak":     tl.maxLoseStreak,
		"current_win_streak":  tl.currentWinStreak,
		"current_lose_streak": tl.currentLoseStreak,
	}
}

// PrintStatistics prints formatted statistics
func (tl *TradeLogger) PrintStatistics() string {
	stats := tl.GetStatistics()

	output := fmt.Sprintf(
		"=== TRADE STATISTICS ===\n"+
			"Total Trades:         %d\n"+
			"Winning:              %d (%.1f%%)\n"+
			"Losing:               %d\n"+
			"Break-Even:           %d\n"+
			"Profit Factor:        %.2f\n"+
			"Largest Win:          $%.2f\n"+
			"Largest Loss:         $%.2f\n"+
			"Total Wins:           $%.2f\n"+
			"Total Losses:         $%.2f\n"+
			"Max Win Streak:       %d\n"+
			"Max Lose Streak:      %d\n"+
			"Current Win Streak:   %d\n"+
			"Current Lose Streak:  %d\n",
		stats["total_trades"],
		stats["winning_trades"],
		stats["win_rate"],
		stats["losing_trades"],
		stats["break_even_trades"],
		stats["profit_factor"],
		stats["largest_win"],
		stats["largest_loss"],
		stats["total_wins"],
		stats["total_losses"],
		stats["max_win_streak"],
		stats["max_lose_streak"],
		stats["current_win_streak"],
		stats["current_lose_streak"],
	)

	return output
}

// ==================== FILTERING ====================

// GetTradesByInstrument returns trades for specific instrument
func (tl *TradeLogger) GetTradesByInstrument(instrument string) []*TradeLog {
	tl.tradesMutex.RLock()
	defer tl.tradesMutex.RUnlock()

	var result []*TradeLog
	for _, trade := range tl.trades {
		if trade.Instrument == instrument {
			result = append(result, trade)
		}
	}
	return result
}

// GetTradesByAction returns trades of specific action
func (tl *TradeLogger) GetTradesByAction(action string) []*TradeLog {
	tl.tradesMutex.RLock()
	defer tl.tradesMutex.RUnlock()

	var result []*TradeLog
	for _, trade := range tl.trades {
		if trade.Action == action {
			result = append(result, trade)
		}
	}
	return result
}

// GetWinningTrade returns trades with positive P&L
func (tl *TradeLogger) GetWinningTradeList() []*TradeLog {
	tl.tradesMutex.RLock()
	defer tl.tradesMutex.RUnlock()

	var result []*TradeLog
	for _, trade := range tl.trades {
		if trade.RealizedPnL > 0 {
			result = append(result, trade)
		}
	}
	return result
}

// GetLosingTradeList returns trades with negative P&L
func (tl *TradeLogger) GetLosingTradeList() []*TradeLog {
	tl.tradesMutex.RLock()
	defer tl.tradesMutex.RUnlock()

	var result []*TradeLog
	for _, trade := range tl.trades {
		if trade.RealizedPnL < 0 {
			result = append(result, trade)
		}
	}
	return result
}

// GetTradesInDateRange returns trades within date range
func (tl *TradeLogger) GetTradesInDateRange(start, end time.Time) []*TradeLog {
	tl.tradesMutex.RLock()
	defer tl.tradesMutex.RUnlock()

	var result []*TradeLog
	for _, trade := range tl.trades {
		if trade.Timestamp.After(start) && trade.Timestamp.Before(end) {
			result = append(result, trade)
		}
	}
	return result
}

// ==================== ANALYSIS ====================

// AnalyzeWinLossRatio analyzes win/loss characteristics
func (tl *TradeLogger) AnalyzeWinLossRatio() map[string]float64 {
	var avgWin, avgLoss float64

	if tl.winningTrades > 0 {
		avgWin = tl.totalWinAmount / float64(tl.winningTrades)
	}
	if tl.losingTrades > 0 {
		avgLoss = -tl.totalLossAmount / float64(tl.losingTrades)
	}

	var ratio float64
	if avgLoss > 0 {
		ratio = avgWin / avgLoss
	}

	return map[string]float64{
		"average_win":    avgWin,
		"average_loss":   avgLoss,
		"win_loss_ratio": ratio,
	}
}

// GetConsecutiveLosses returns longest consecutive loss sequence
func (tl *TradeLogger) GetConsecutiveLosses() int64 {
	tl.tradesMutex.RLock()
	defer tl.tradesMutex.RUnlock()

	maxLosses := int64(0)
	currentLosses := int64(0)

	for _, trade := range tl.trades {
		if trade.RealizedPnL < 0 {
			currentLosses++
			if currentLosses > maxLosses {
				maxLosses = currentLosses
			}
		} else {
			currentLosses = 0
		}
	}

	return maxLosses
}
