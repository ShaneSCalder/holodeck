package executor

import (
	"fmt"
	"time"

	"holodeck/types"
)

// ==================== LIMIT ORDER EXECUTOR ====================

// LimitOrderExecutor executes LIMIT orders
type LimitOrderExecutor struct {
	validator OrderValidator
}

// NewLimitOrderExecutor creates a new limit order executor
func NewLimitOrderExecutor() *LimitOrderExecutor {
	return &LimitOrderExecutor{
		validator: NewOrderValidator(),
	}
}

// Execute executes a limit order
func (loe *LimitOrderExecutor) Execute(
	order *types.Order,
	tick *types.Tick,
	instrument types.Instrument,
) (*types.ExecutionReport, error) {

	// Validate it's a LIMIT order
	if !order.IsLimit() {
		return nil, types.NewInvalidOrderTypeError("not a limit order")
	}

	// Check if limit price would be filled
	filled, price := loe.checkFillCondition(order, tick)

	if !filled {
		// Limit order not filled - return unfilled status
		return &types.ExecutionReport{
			OrderID:       order.OrderID,
			Timestamp:     tick.Timestamp,
			Action:        order.Action,
			RequestedSize: order.Size,
			FilledSize:    0, // Not filled
			FillPrice:     order.LimitPrice,
			Status:        types.OrderStatusPending,
		}, nil
	}

	// Limit order is filled - create execution report
	exec := &types.ExecutionReport{
		OrderID:       order.OrderID,
		Timestamp:     tick.Timestamp,
		Action:        order.Action,
		RequestedSize: order.Size,
		FilledSize:    order.Size, // Assume full fill when condition met
		FillPrice:     price,
		Status:        types.OrderStatusFilled,
	}

	return exec, nil
}

// checkFillCondition checks if limit order would be filled
// Returns: (filled bool, fillPrice float64)
func (loe *LimitOrderExecutor) checkFillCondition(
	order *types.Order,
	tick *types.Tick,
) (bool, float64) {

	if order.IsBuy() {
		// Buy limit: fill if ask price <= limit price
		if tick.GetBuyPrice() <= order.LimitPrice {
			return true, tick.GetBuyPrice()
		}
	} else if order.IsSell() {
		// Sell limit: fill if bid price >= limit price
		if tick.GetSellPrice() >= order.LimitPrice {
			return true, tick.GetSellPrice()
		}
	}

	return false, 0
}

// ==================== LIMIT ORDER VALIDATION ====================

// ValidateLimitOrder validates a limit order
func (loe *LimitOrderExecutor) ValidateLimitOrder(
	order *types.Order,
	instrument types.Instrument,
	availableBalance float64,
	minSize float64,
	maxSize float64,
) error {

	if !order.IsLimit() {
		return types.NewInvalidOrderTypeError("not a limit order")
	}

	// Basic order validation
	if err := loe.validator.ValidateOrder(
		order,
		instrument,
		availableBalance,
		minSize,
		maxSize,
		maxSize,
	); err != nil {
		return err
	}

	// Validate limit price
	return loe.validator.ValidateLimitPrice(
		order.LimitPrice,
		0, // We don't have current price context here
		order.Action,
		instrument,
	)
}

// ==================== LIMIT ORDER DETAILS ====================

// LimitOrderDetails contains details about a limit order
type LimitOrderDetails struct {
	OrderID       string
	Action        string
	Size          float64
	LimitPrice    float64
	CurrentPrice  float64
	Status        string // FILLED, PENDING, EXPIRED
	FillPrice     float64
	DistanceTicks int64 // How many ticks away from fill
	TimeToFill    *time.Duration
}

// CheckLimitOrderStatus checks if a limit order would fill
func CheckLimitOrderStatus(
	order *types.Order,
	tick *types.Tick,
) *LimitOrderDetails {

	var currentPrice float64
	var status string
	distanceTicks := int64(0)

	if order.IsBuy() {
		currentPrice = tick.GetBuyPrice()
		if currentPrice <= order.LimitPrice {
			status = types.OrderStatusFilled
		} else {
			status = types.OrderStatusPending
			// Distance in smallest units (pips for forex, cents for stocks)
			distanceTicks = int64((currentPrice - order.LimitPrice) * 10000)
		}
	} else {
		currentPrice = tick.GetSellPrice()
		if currentPrice >= order.LimitPrice {
			status = types.OrderStatusFilled
		} else {
			status = types.OrderStatusPending
			distanceTicks = int64((order.LimitPrice - currentPrice) * 10000)
		}
	}

	return &LimitOrderDetails{
		OrderID:       order.OrderID,
		Action:        order.Action,
		Size:          order.Size,
		LimitPrice:    order.LimitPrice,
		CurrentPrice:  currentPrice,
		Status:        status,
		FillPrice:     currentPrice,
		DistanceTicks: distanceTicks,
	}
}

// String returns string representation
func (lod *LimitOrderDetails) String() string {
	return fmt.Sprintf(
		"Limit %s: %.2f @ %.8f, current: %.8f, status: %s",
		lod.Action,
		lod.Size,
		lod.LimitPrice,
		lod.CurrentPrice,
		lod.Status,
	)
}

// DebugString returns detailed debug info
func (lod *LimitOrderDetails) DebugString() string {
	return fmt.Sprintf(
		"Limit Order Details:\n"+
			"  Order ID:          %s\n"+
			"  Action:            %s\n"+
			"  Size:              %.6f\n"+
			"  Limit Price:       %.8f\n"+
			"  Current Price:     %.8f\n"+
			"  Status:            %s\n"+
			"  Distance (ticks):  %d\n"+
			"  Would Fill:        %v",
		lod.OrderID,
		lod.Action,
		lod.Size,
		lod.LimitPrice,
		lod.CurrentPrice,
		lod.Status,
		lod.DistanceTicks,
		lod.Status == types.OrderStatusFilled,
	)
}

// ==================== LIMIT ORDER TRACKING ====================

// LimitOrderTracker tracks pending limit orders
type LimitOrderTracker struct {
	pendingOrders map[string]*types.Order
	filledOrders  map[string]*types.Order
	expiredOrders map[string]*types.Order
}

// NewLimitOrderTracker creates a new tracker
func NewLimitOrderTracker() *LimitOrderTracker {
	return &LimitOrderTracker{
		pendingOrders: make(map[string]*types.Order),
		filledOrders:  make(map[string]*types.Order),
		expiredOrders: make(map[string]*types.Order),
	}
}

// AddPending adds a pending limit order
func (lot *LimitOrderTracker) AddPending(order *types.Order) {
	lot.pendingOrders[order.OrderID] = order
}

// CheckFills checks all pending orders for fills
func (lot *LimitOrderTracker) CheckFills(tick *types.Tick) []string {
	executor := NewLimitOrderExecutor()
	filled := make([]string, 0)

	for orderID, order := range lot.pendingOrders {
		if exec, _ := executor.Execute(order, tick, nil); exec != nil {
			if exec.IsFilled() {
				filled = append(filled, orderID)
				lot.filledOrders[orderID] = order
				delete(lot.pendingOrders, orderID)
			}
		}
	}

	return filled
}

// GetPendingCount returns number of pending orders
func (lot *LimitOrderTracker) GetPendingCount() int {
	return len(lot.pendingOrders)
}

// GetFilledCount returns number of filled orders
func (lot *LimitOrderTracker) GetFilledCount() int {
	return len(lot.filledOrders)
}
