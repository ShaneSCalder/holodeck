package executor

import (
	"fmt"

	"holodeck/types"
)

// ==================== MARKET ORDER EXECUTOR ====================

// MarketOrderExecutor executes MARKET orders
type MarketOrderExecutor struct {
	validator OrderValidator
}

// NewMarketOrderExecutor creates a new market order executor
func NewMarketOrderExecutor() *MarketOrderExecutor {
	return &MarketOrderExecutor{
		validator: NewOrderValidator(),
	}
}

// Execute executes a market order
func (moe *MarketOrderExecutor) Execute(
	order *types.Order,
	tick *types.Tick,
	instrument types.Instrument,
) (*types.ExecutionReport, error) {

	// Validate it's a MARKET order
	if !order.IsMarket() {
		return nil, types.NewInvalidOrderTypeError("not a market order")
	}

	// Get fill price based on order side
	var fillPrice float64
	if order.IsBuy() {
		fillPrice = tick.GetBuyPrice() // Use ask for buy
	} else {
		fillPrice = tick.GetSellPrice() // Use bid for sell
	}

	// Validate fill price is reasonable
	if err := ValidateFillPrice(fillPrice, tick.Bid, tick.Ask, instrument.GetPipValue()); err != nil {
		return types.NewRejectedExecution(
			order.OrderID,
			tick.Timestamp,
			order.Action,
			order.Size,
			types.ErrorCodeOrderRejected,
			fmt.Sprintf("invalid fill price: %v", err),
		), nil
	}

	// Create execution report
	exec := &types.ExecutionReport{
		OrderID:       order.OrderID,
		Timestamp:     tick.Timestamp,
		Action:        order.Action,
		OrderType:     types.OrderTypeMarket,
		RequestedSize: order.Size,
		FilledSize:    order.Size, // Market orders fill immediately
		FillPrice:     fillPrice,
		Status:        types.OrderStatusFilled,
	}

	return exec, nil
}

// ==================== MARKET ORDER VALIDATION ====================

// ValidateMarketOrder validates a market order
func (moe *MarketOrderExecutor) ValidateMarketOrder(
	order *types.Order,
	instrument types.Instrument,
	availableBalance float64,
	minSize float64,
	maxSize float64,
) error {

	if !order.IsMarket() {
		return types.NewInvalidOrderTypeError("not a market order")
	}

	return moe.validator.ValidateOrder(
		order,
		instrument,
		availableBalance,
		minSize,
		maxSize,
		maxSize, // Use maxSize as max position
	)
}

// ==================== MARKET ORDER DETAILS ====================

// MarketOrderDetails contains details about a market order execution
type MarketOrderDetails struct {
	OrderID   string
	Action    string
	Size      float64
	FillPrice float64
	AskPrice  float64
	BidPrice  float64
	Slippage  float64
	IsAdverse bool // true if slippage is worse than expected
}

// AnalyzeMarketFill analyzes fill quality for a market order
func AnalyzeMarketFill(
	order *types.Order,
	fillPrice float64,
	bidPrice float64,
	askPrice float64,
) *MarketOrderDetails {

	midPrice := (bidPrice + askPrice) / 2

	// Calculate slippage from mid price
	var slippage float64
	isAdverse := false

	if order.IsBuy() {
		slippage = fillPrice - midPrice
		isAdverse = slippage > 0 // Bought worse than mid
	} else {
		slippage = midPrice - fillPrice
		isAdverse = slippage > 0 // Sold worse than mid
	}

	return &MarketOrderDetails{
		OrderID:   order.OrderID,
		Action:    order.Action,
		Size:      order.Size,
		FillPrice: fillPrice,
		AskPrice:  askPrice,
		BidPrice:  bidPrice,
		Slippage:  slippage,
		IsAdverse: isAdverse,
	}
}

// String returns string representation
func (mod *MarketOrderDetails) String() string {
	adverse := ""
	if mod.IsAdverse {
		adverse = " (ADVERSE)"
	}
	return fmt.Sprintf(
		"Market %s: %.2f @ %.8f, slippage: %.8f%s",
		mod.Action,
		mod.Size,
		mod.FillPrice,
		mod.Slippage,
		adverse,
	)
}

// DebugString returns detailed debug info
func (mod *MarketOrderDetails) DebugString() string {
	return fmt.Sprintf(
		"Market Order Analysis:\n"+
			"  Order ID:     %s\n"+
			"  Action:       %s\n"+
			"  Size:         %.6f\n"+
			"  Fill Price:   %.8f\n"+
			"  Mid Price:    %.8f\n"+
			"  Bid Price:    %.8f\n"+
			"  Ask Price:    %.8f\n"+
			"  Slippage:     %.8f\n"+
			"  Is Adverse:   %v",
		mod.OrderID,
		mod.Action,
		mod.Size,
		mod.FillPrice,
		(mod.BidPrice+mod.AskPrice)/2,
		mod.BidPrice,
		mod.AskPrice,
		mod.Slippage,
		mod.IsAdverse,
	)
}
