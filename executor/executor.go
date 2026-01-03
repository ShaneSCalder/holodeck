package executor

import (
	"fmt"

	"holodeck/types"
)

// ==================== ORDER EXECUTOR ====================

// OrderExecutor orchestrates order execution with validation and partial fills
type OrderExecutor struct {
	config ExecutorConfig

	// Statistics
	ordersReceived   int64
	ordersExecuted   int64
	ordersRejected   int64
	executionHistory []*types.ExecutionReport
}

// ExecutorConfig holds executor configuration
type ExecutorConfig struct {
	// Features enabled
	CommissionEnabled   bool
	SlippageEnabled     bool
	LatencyEnabled      bool
	PartialFillsEnabled bool

	// Order limits
	MaxOrderSize     float64
	MaxPositionSize  float64
	MinimumOrderSize float64
}

// ==================== EXECUTOR CREATION ====================

// NewOrderExecutor creates a new order executor
func NewOrderExecutor(config ExecutorConfig) *OrderExecutor {
	return &OrderExecutor{
		config:           config,
		executionHistory: make([]*types.ExecutionReport, 0),
	}
}

// ==================== CORE EXECUTION ====================

// Execute orchestrates the execution of an order
func (oe *OrderExecutor) Execute(
	order *types.Order,
	tick *types.Tick,
	instrument types.Instrument,
) (*types.ExecutionReport, error) {

	oe.ordersReceived++

	// Validate inputs
	if order == nil {
		return nil, types.NewOrderRejectedError("order cannot be nil")
	}

	if tick == nil {
		return nil, types.NewOrderRejectedError("tick cannot be nil")
	}

	if instrument == nil {
		return nil, types.NewOrderRejectedError("instrument cannot be nil")
	}

	// Handle HOLD orders
	if order.IsHold() {
		oe.ordersExecuted++
		return &types.ExecutionReport{
			OrderID:       order.OrderID,
			Timestamp:     tick.Timestamp,
			Action:        types.OrderActionHold,
			Status:        types.OrderStatusFilled,
			RequestedSize: 0,
			FilledSize:    0,
		}, nil
	}

	// Validate order
	validator := NewOrderValidator()
	if err := validator.ValidateOrder(
		order,
		instrument,
		10000000, // Default available balance
		oe.config.MinimumOrderSize,
		oe.config.MaxOrderSize,
		oe.config.MaxPositionSize,
	); err != nil {
		oe.ordersRejected++
		herr := err.(*types.HolodeckError)
		return types.NewRejectedExecution(
			order.OrderID,
			tick.Timestamp,
			order.Action,
			order.Size,
			herr.Code,
			herr.Message,
		), nil
	}

	// Route to appropriate executor
	var exec *types.ExecutionReport
	var err error

	if order.IsMarket() {
		moe := NewMarketOrderExecutor()
		exec, err = moe.Execute(order, tick, instrument)
	} else if order.IsLimit() {
		loe := NewLimitOrderExecutor()
		exec, err = loe.Execute(order, tick, instrument)
	} else {
		return types.NewRejectedExecution(
			order.OrderID,
			tick.Timestamp,
			order.Action,
			order.Size,
			types.ErrorCodeOrderRejected,
			"unknown order type",
		), nil
	}

	if err != nil {
		oe.ordersRejected++
		return nil, err
	}

	// Handle partial fills if enabled
	if oe.config.PartialFillsEnabled && exec.IsFilled() {
		pfc := NewPartialFillCalculator()
		filledSize := pfc.CalculateFilledSize(
			exec.RequestedSize,
			int64(tick.GetAvailableDepth()),
			tick.Volume,
		)

		if filledSize < exec.RequestedSize {
			exec.FilledSize = filledSize
			exec.Status = types.OrderStatusPartial
		}
	}

	// Record execution
	oe.recordExecution(exec)
	if !exec.IsRejected() {
		oe.ordersExecuted++
	} else {
		oe.ordersRejected++
	}

	return exec, nil
}

// ==================== VALIDATION ====================

// ValidateOrder validates an order before execution
func (oe *OrderExecutor) ValidateOrder(
	order *types.Order,
	instrument types.Instrument,
	availableBalance float64,
) error {

	validator := NewOrderValidator()
	return validator.ValidateOrder(
		order,
		instrument,
		availableBalance,
		oe.config.MinimumOrderSize,
		oe.config.MaxOrderSize,
		oe.config.MaxPositionSize,
	)
}

// ==================== STATISTICS ====================

// GetOrdersReceived returns total orders received
func (oe *OrderExecutor) GetOrdersReceived() int64 {
	return oe.ordersReceived
}

// GetOrdersExecuted returns total orders executed
func (oe *OrderExecutor) GetOrdersExecuted() int64 {
	return oe.ordersExecuted
}

// GetOrdersRejected returns total orders rejected
func (oe *OrderExecutor) GetOrdersRejected() int64 {
	return oe.ordersRejected
}

// GetExecutionRate returns percentage of orders executed
func (oe *OrderExecutor) GetExecutionRate() float64 {
	if oe.ordersReceived == 0 {
		return 0
	}
	return (float64(oe.ordersExecuted) / float64(oe.ordersReceived)) * 100
}

// GetExecutionHistory returns execution history
func (oe *OrderExecutor) GetExecutionHistory() []*types.ExecutionReport {
	return oe.executionHistory
}

// GetStatistics returns comprehensive executor statistics
func (oe *OrderExecutor) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"orders_received":        oe.ordersReceived,
		"orders_executed":        oe.ordersExecuted,
		"orders_rejected":        oe.ordersRejected,
		"execution_rate":         oe.GetExecutionRate(),
		"execution_history_size": int64(len(oe.executionHistory)),
	}
}

// recordExecution records execution details
func (oe *OrderExecutor) recordExecution(exec *types.ExecutionReport) {
	oe.executionHistory = append(oe.executionHistory, exec)

	// Trim history if too large
	maxHistory := 10000
	if len(oe.executionHistory) > maxHistory {
		oe.executionHistory = oe.executionHistory[len(oe.executionHistory)-maxHistory:]
	}
}

// ==================== DEBUG ====================

// String returns a human-readable representation
func (oe *OrderExecutor) String() string {
	return fmt.Sprintf(
		"OrderExecutor[Received:%d, Executed:%d, Rejected:%d, Rate:%.1f%%]",
		oe.ordersReceived,
		oe.ordersExecuted,
		oe.ordersRejected,
		oe.GetExecutionRate(),
	)
}

// DebugString returns detailed debug information
func (oe *OrderExecutor) DebugString() string {
	return fmt.Sprintf(
		"Order Executor:\n"+
			"  Orders Received:      %d\n"+
			"  Orders Executed:      %d\n"+
			"  Orders Rejected:      %d\n"+
			"  Execution Rate:       %.2f%%\n"+
			"  Execution History:    %d\n"+
			"\n"+
			"  Configuration:\n"+
			"    Commission:         %v\n"+
			"    Slippage:           %v\n"+
			"    Latency:            %v\n"+
			"    Partial Fills:      %v\n"+
			"    Min Order Size:     %.6f\n"+
			"    Max Order Size:     %.6f\n"+
			"    Max Position Size:  %.6f",
		oe.ordersReceived,
		oe.ordersExecuted,
		oe.ordersRejected,
		oe.GetExecutionRate(),
		len(oe.executionHistory),
		oe.config.CommissionEnabled,
		oe.config.SlippageEnabled,
		oe.config.LatencyEnabled,
		oe.config.PartialFillsEnabled,
		oe.config.MinimumOrderSize,
		oe.config.MaxOrderSize,
		oe.config.MaxPositionSize,
	)
}

// Reset resets executor statistics
func (oe *OrderExecutor) Reset() {
	oe.ordersReceived = 0
	oe.ordersExecuted = 0
	oe.ordersRejected = 0
	oe.executionHistory = make([]*types.ExecutionReport, 0)
}
