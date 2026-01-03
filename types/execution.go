package types

import (
	"fmt"
	"time"
)

// ==================== EXECUTION REPORT STRUCTURE ====================

// ExecutionReport represents the result of executing an order
// This is the output from ExecuteOrder()
type ExecutionReport struct {
	// OrderID is the unique identifier for this order
	OrderID string

	// Timestamp is when the order was executed
	Timestamp time.Time

	// Action is what was executed: BUY or SELL
	Action string

	// RequestedSize is the size the agent asked for
	RequestedSize float64

	// FilledSize is the actual size that was filled
	// May be less than RequestedSize if partial fill
	FilledSize float64

	// FillPrice is the average price this was filled at
	// Includes slippage but not commission
	FillPrice float64

	// SlippageUnits is the slippage in decimal units (pips for forex, cents for stocks, etc)
	SlippageUnits float64

	// Commission is the trading fee paid
	Commission float64

	// PositionAfter is the position size after this execution
	// Positive = LONG, negative = SHORT, 0 = FLAT
	PositionAfter float64

	// EntryPrice is the entry price for current position (if open)
	EntryPrice float64

	// UnrealizedPnL is the mark-to-market profit/loss on open position
	UnrealizedPnL float64

	// RealizedPnL is the profit/loss from closed trades
	RealizedPnL float64

	// TotalPnL is realized + unrealized - cumulative commissions
	TotalPnL float64

	// Status is the execution status: FILLED, PARTIAL, REJECTED
	Status string

	// ErrorCode is populated if Status is REJECTED
	ErrorCode string

	// ErrorMessage is the error description if rejected
	ErrorMessage string

	// Latency is the delay in milliseconds before execution
	Latency int64

	// AvailableDepth is the available volume at execution time
	AvailableDepth int64

	// AverageFillPrice is the price including slippage and commission impact
	AverageFillPrice float64
}

// ==================== EXECUTION REPORT CONSTRUCTORS ====================

// NewExecutionReport creates a new ExecutionReport for a successful fill
func NewExecutionReport(
	orderID string,
	timestamp time.Time,
	action string,
	requestedSize, filledSize, fillPrice, slippage, commission float64,
	positionAfter, entryPrice, unrealizedPnL, realizedPnL, totalPnL float64,
) *ExecutionReport {
	return &ExecutionReport{
		OrderID:          orderID,
		Timestamp:        timestamp,
		Action:           action,
		RequestedSize:    requestedSize,
		FilledSize:       filledSize,
		FillPrice:        fillPrice,
		SlippageUnits:    slippage,
		Commission:       commission,
		PositionAfter:    positionAfter,
		EntryPrice:       entryPrice,
		UnrealizedPnL:    unrealizedPnL,
		RealizedPnL:      realizedPnL,
		TotalPnL:         totalPnL,
		Status:           OrderStatusFilled,
		AverageFillPrice: fillPrice, // Without commission for now
	}
}

// NewRejectedExecution creates an ExecutionReport for a rejected order
func NewRejectedExecution(
	orderID string,
	timestamp time.Time,
	action string,
	requestedSize float64,
	errorCode, errorMessage string,
) *ExecutionReport {
	return &ExecutionReport{
		OrderID:       orderID,
		Timestamp:     timestamp,
		Action:        action,
		RequestedSize: requestedSize,
		FilledSize:    0,
		Status:        OrderStatusRejected,
		ErrorCode:     errorCode,
		ErrorMessage:  errorMessage,
	}
}

// NewPartialExecution creates an ExecutionReport for a partial fill
func NewPartialExecution(
	orderID string,
	timestamp time.Time,
	action string,
	requestedSize, filledSize, fillPrice, slippage, commission float64,
	positionAfter, entryPrice, unrealizedPnL, realizedPnL, totalPnL float64,
) *ExecutionReport {
	report := NewExecutionReport(
		orderID, timestamp, action,
		requestedSize, filledSize, fillPrice, slippage, commission,
		positionAfter, entryPrice, unrealizedPnL, realizedPnL, totalPnL,
	)
	report.Status = OrderStatusPartial
	return report
}

// ==================== EXECUTION REPORT METHODS ====================

// IsFilled returns true if the order was fully filled
func (er *ExecutionReport) IsFilled() bool {
	return er.Status == OrderStatusFilled
}

// IsPartial returns true if the order was partially filled
func (er *ExecutionReport) IsPartial() bool {
	return er.Status == OrderStatusPartial
}

// IsRejected returns true if the order was rejected
func (er *ExecutionReport) IsRejected() bool {
	return er.Status == OrderStatusRejected
}

// WasExecuted returns true if order was filled (fully or partially)
func (er *ExecutionReport) WasExecuted() bool {
	return er.IsPartial() || er.IsFilled()
}

// GetFillPercentage returns the percentage of order that was filled (0-100)
func (er *ExecutionReport) GetFillPercentage() float64 {
	if er.RequestedSize == 0 {
		return 0
	}
	return (er.FilledSize / er.RequestedSize) * 100.0
}

// GetUnfilledSize returns the remaining unfilled quantity
func (er *ExecutionReport) GetUnfilledSize() float64 {
	return er.RequestedSize - er.FilledSize
}

// GetAverageSlippage returns average slippage per unit
func (er *ExecutionReport) GetAverageSlippage() float64 {
	if er.FilledSize == 0 {
		return 0
	}
	return er.SlippageUnits / er.FilledSize
}

// GetAverageCommission returns average commission per unit
func (er *ExecutionReport) GetAverageCommission() float64 {
	if er.FilledSize == 0 {
		return 0
	}
	return er.Commission / er.FilledSize
}

// GetNotional returns the notional value of the fill
func (er *ExecutionReport) GetNotional() float64 {
	return er.FilledSize * er.FillPrice
}

// IsBuy returns true if this was a BUY execution
func (er *ExecutionReport) IsBuy() bool {
	return er.Action == OrderActionBuy
}

// IsSell returns true if this was a SELL execution
func (er *ExecutionReport) IsSell() bool {
	return er.Action == OrderActionSell
}

// GetPositionStatus returns the position status after execution
func (er *ExecutionReport) GetPositionStatus() string {
	return GetPositionStatusFromSize(er.PositionAfter)
}

// IsLongPosition returns true if position is long after execution
func (er *ExecutionReport) IsLongPosition() bool {
	return er.PositionAfter > 0
}

// IsShortPosition returns true if position is short after execution
func (er *ExecutionReport) IsShortPosition() bool {
	return er.PositionAfter < 0
}

// IsFlatPosition returns true if position is flat after execution
func (er *ExecutionReport) IsFlatPosition() bool {
	return er.PositionAfter == 0
}

// String returns a human-readable representation
func (er *ExecutionReport) String() string {
	if er.IsRejected() {
		return fmt.Sprintf(
			"Execution[REJECTED %s %f | Error: %s]",
			er.Action, er.RequestedSize, er.ErrorMessage,
		)
	}

	status := "FILLED"
	if er.IsPartial() {
		status = "PARTIAL"
	}

	return fmt.Sprintf(
		"Execution[%s %s %f @ %.5f | Filled: %f (%.1f%%) | P&L: %.2f | Status: %s]",
		status, er.Action, er.RequestedSize, er.FillPrice, er.FilledSize,
		er.GetFillPercentage(), er.TotalPnL, er.Status,
	)
}

// DebugString returns detailed execution information
func (er *ExecutionReport) DebugString() string {
	if er.IsRejected() {
		return fmt.Sprintf(
			"ExecutionReport (REJECTED):\n"+
				"  OrderID:       %s\n"+
				"  Timestamp:     %s\n"+
				"  Action:        %s\n"+
				"  Requested:     %f\n"+
				"  Error Code:    %s\n"+
				"  Error Message: %s",
			er.OrderID,
			er.Timestamp.Format("2006-01-02T15:04:05.000000"),
			er.Action,
			er.RequestedSize,
			er.ErrorCode,
			er.ErrorMessage,
		)
	}

	return fmt.Sprintf(
		"ExecutionReport:\n"+
			"  OrderID:        %s\n"+
			"  Timestamp:      %s\n"+
			"  Action:         %s\n"+
			"  Requested:      %f\n"+
			"  Filled:         %f (%.1f%%)\n"+
			"  Fill Price:     %.8f\n"+
			"  Slippage:       %.8f\n"+
			"  Commission:     %.2f\n"+
			"  Notional:       %.2f\n"+
			"  Avg Comm/Unit:  %.6f\n"+
			"  Position After: %f (%s)\n"+
			"  Entry Price:    %.8f\n"+
			"  Unrealized P&L: %.2f\n"+
			"  Realized P&L:   %.2f\n"+
			"  Total P&L:      %.2f\n"+
			"  Status:         %s\n"+
			"  Latency:        %d ms",
		er.OrderID,
		er.Timestamp.Format("2006-01-02T15:04:05.000000"),
		er.Action,
		er.RequestedSize,
		er.FilledSize,
		er.GetFillPercentage(),
		er.FillPrice,
		er.SlippageUnits,
		er.Commission,
		er.GetNotional(),
		er.GetAverageCommission(),
		er.PositionAfter,
		er.GetPositionStatus(),
		er.EntryPrice,
		er.UnrealizedPnL,
		er.RealizedPnL,
		er.TotalPnL,
		er.Status,
		er.Latency,
	)
}

// ==================== EXECUTION STATISTICS ====================

// ExecutionStats holds statistics about a series of executions
type ExecutionStats struct {
	// Total number of executions
	TotalExecutions int

	// Number of filled executions
	FilledExecutions int

	// Number of partial fills
	PartialFills int

	// Number of rejected orders
	RejectedOrders int

	// Total filled volume
	TotalFilledVolume float64

	// Total requested volume
	TotalRequestedVolume float64

	// Fill rate (filled / requested)
	FillRate float64

	// Total commission paid
	TotalCommission float64

	// Total slippage cost
	TotalSlippage float64

	// Best fill price
	BestFillPrice float64

	// Worst fill price
	WorstFillPrice float64

	// Average fill price
	AverageFillPrice float64

	// Total realized P&L
	RealizedPnL float64

	// Total unrealized P&L
	UnrealizedPnL float64

	// Total P&L
	TotalPnL float64

	// Best trade P&L
	BestTradeP_L float64

	// Worst trade P&L
	WorstTradeP_L float64

	// Winning trades count
	WinningTrades int

	// Losing trades count
	LosingTrades int

	// Win rate percentage
	WinRate float64

	// Average trade P&L
	AverageTradeP_L float64
}

// CalculateExecutionStats calculates statistics from a set of execution reports
func CalculateExecutionStats(reports []*ExecutionReport) *ExecutionStats {
	if len(reports) == 0 {
		return &ExecutionStats{}
	}

	stats := &ExecutionStats{
		TotalExecutions: len(reports),
		BestFillPrice:   reports[0].FillPrice,
		WorstFillPrice:  reports[0].FillPrice,
		BestTradeP_L:    reports[0].TotalPnL,
		WorstTradeP_L:   reports[0].TotalPnL,
	}

	var sumFillPrice float64

	for _, report := range reports {
		if report.IsRejected() {
			stats.RejectedOrders++
			stats.TotalRequestedVolume += report.RequestedSize
			continue
		}

		stats.FilledExecutions++
		if report.IsPartial() {
			stats.PartialFills++
		}

		stats.TotalFilledVolume += report.FilledSize
		stats.TotalRequestedVolume += report.RequestedSize
		stats.TotalCommission += report.Commission
		stats.TotalSlippage += report.SlippageUnits
		stats.RealizedPnL += report.RealizedPnL
		stats.UnrealizedPnL += report.UnrealizedPnL
		stats.TotalPnL += report.TotalPnL

		// Track best/worst prices and P&L
		if report.FillPrice < stats.BestFillPrice {
			stats.BestFillPrice = report.FillPrice
		}
		if report.FillPrice > stats.WorstFillPrice {
			stats.WorstFillPrice = report.FillPrice
		}
		if report.RealizedPnL > stats.BestTradeP_L {
			stats.BestTradeP_L = report.RealizedPnL
		}
		if report.RealizedPnL < stats.WorstTradeP_L {
			stats.WorstTradeP_L = report.RealizedPnL
		}

		// Count winning/losing trades (only closed trades)
		if report.RealizedPnL > 0 {
			stats.WinningTrades++
		} else if report.RealizedPnL < 0 {
			stats.LosingTrades++
		}

		sumFillPrice += report.FillPrice
	}

	// Calculate derived stats
	if stats.FilledExecutions > 0 {
		stats.AverageFillPrice = sumFillPrice / float64(stats.FilledExecutions)
		stats.AverageTradeP_L = stats.RealizedPnL / float64(stats.FilledExecutions)

		totalClosedTrades := stats.WinningTrades + stats.LosingTrades
		if totalClosedTrades > 0 {
			stats.WinRate = (float64(stats.WinningTrades) / float64(totalClosedTrades)) * 100.0
		}
	}

	if stats.TotalRequestedVolume > 0 {
		stats.FillRate = (stats.TotalFilledVolume / stats.TotalRequestedVolume) * 100.0
	}

	return stats
}

// String returns a human-readable representation of execution stats
func (es *ExecutionStats) String() string {
	return fmt.Sprintf(
		"ExecutionStats[%d exec, %.1f%% fill rate, %.2f total P&L, %.2f win rate]",
		es.TotalExecutions,
		es.FillRate,
		es.TotalPnL,
		es.WinRate,
	)
}

// DebugString returns detailed statistics
func (es *ExecutionStats) DebugString() string {
	return fmt.Sprintf(
		"Execution Statistics:\n"+
			"  Total Executions:      %d\n"+
			"  Filled:                %d\n"+
			"  Partial Fills:         %d\n"+
			"  Rejected:              %d\n"+
			"\n"+
			"  Volume:\n"+
			"    Requested:           %f\n"+
			"    Filled:              %f\n"+
			"    Fill Rate:           %.2f%%\n"+
			"\n"+
			"  Pricing:\n"+
			"    Best Fill:           %.8f\n"+
			"    Worst Fill:          %.8f\n"+
			"    Average Fill:        %.8f\n"+
			"    Total Slippage:      %.8f\n"+
			"\n"+
			"  Costs:\n"+
			"    Total Commission:    %.2f\n"+
			"\n"+
			"  P&L:\n"+
			"    Realized P&L:        %.2f\n"+
			"    Unrealized P&L:      %.2f\n"+
			"    Total P&L:           %.2f\n"+
			"\n"+
			"  Trade Performance:\n"+
			"    Winning Trades:      %d\n"+
			"    Losing Trades:       %d\n"+
			"    Win Rate:            %.2f%%\n"+
			"    Best Trade:          %.2f\n"+
			"    Worst Trade:         %.2f\n"+
			"    Average Trade:       %.2f",
		es.TotalExecutions,
		es.FilledExecutions,
		es.PartialFills,
		es.RejectedOrders,
		es.TotalRequestedVolume,
		es.TotalFilledVolume,
		es.FillRate,
		es.BestFillPrice,
		es.WorstFillPrice,
		es.AverageFillPrice,
		es.TotalSlippage,
		es.TotalCommission,
		es.RealizedPnL,
		es.UnrealizedPnL,
		es.TotalPnL,
		es.WinningTrades,
		es.LosingTrades,
		es.WinRate,
		es.BestTradeP_L,
		es.WorstTradeP_L,
		es.AverageTradeP_L,
	)
}

// ==================== EXECUTION BATCH ====================

// ExecutionBatch groups multiple execution reports together
type ExecutionBatch struct {
	Reports   []*ExecutionReport
	Timestamp time.Time
}

// NewExecutionBatch creates a new batch
func NewExecutionBatch(timestamp time.Time) *ExecutionBatch {
	return &ExecutionBatch{
		Reports:   make([]*ExecutionReport, 0),
		Timestamp: timestamp,
	}
}

// Add adds an execution report to the batch
func (eb *ExecutionBatch) Add(report *ExecutionReport) {
	eb.Reports = append(eb.Reports, report)
}

// Size returns the number of reports in batch
func (eb *ExecutionBatch) Size() int {
	return len(eb.Reports)
}

// GetSuccessfulExecutions returns only filled or partial executions
func (eb *ExecutionBatch) GetSuccessfulExecutions() []*ExecutionReport {
	successful := make([]*ExecutionReport, 0)
	for _, report := range eb.Reports {
		if report.WasExecuted() {
			successful = append(successful, report)
		}
	}
	return successful
}

// GetFailedExecutions returns only rejected executions
func (eb *ExecutionBatch) GetFailedExecutions() []*ExecutionReport {
	failed := make([]*ExecutionReport, 0)
	for _, report := range eb.Reports {
		if report.IsRejected() {
			failed = append(failed, report)
		}
	}
	return failed
}

// GetTotalPnL returns the sum of all P&L from all executions
func (eb *ExecutionBatch) GetTotalPnL() float64 {
	total := 0.0
	for _, report := range eb.Reports {
		total += report.TotalPnL
	}
	return total
}

// GetStats calculates statistics for the batch
func (eb *ExecutionBatch) GetStats() *ExecutionStats {
	return CalculateExecutionStats(eb.Reports)
}

// String returns a human-readable representation
func (eb *ExecutionBatch) String() string {
	return fmt.Sprintf("ExecutionBatch[%d executions at %s]", eb.Size(), eb.Timestamp.Format("2006-01-02T15:04:05.000"))
}
