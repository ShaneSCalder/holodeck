package types

import (
	"fmt"
	"time"
)

// ==================== ORDER STRUCTURE ====================

// Order represents a trading order placed by an agent/trader
// This is the input to ExecuteOrder()
type Order struct {
	// Action is what to do: BUY, SELL, or HOLD
	Action string

	// Size is the quantity to trade (in lots, shares, oz, etc depending on instrument)
	Size float64

	// OrderType is how to execute: MARKET or LIMIT
	OrderType string

	// LimitPrice is the price threshold for LIMIT orders (optional)
	// For BUY LIMIT: will only buy if ask <= LimitPrice
	// For SELL LIMIT: will only sell if bid >= LimitPrice
	LimitPrice float64

	// Timestamp is when the order was created
	Timestamp time.Time

	// OrderID is a unique identifier (optional, can be set by executor)
	OrderID string

	// Description is a human-readable note about the order
	Description string
}

// ==================== ORDER CONSTRUCTORS ====================

// NewMarketOrder creates a MARKET order
func NewMarketOrder(action string, size float64, timestamp time.Time) *Order {
	return &Order{
		Action:    action,
		Size:      size,
		OrderType: OrderTypeMarket,
		Timestamp: timestamp,
	}
}

// NewLimitOrder creates a LIMIT order
func NewLimitOrder(action string, size, limitPrice float64, timestamp time.Time) *Order {
	return &Order{
		Action:     action,
		Size:       size,
		OrderType:  OrderTypeLimit,
		LimitPrice: limitPrice,
		Timestamp:  timestamp,
	}
}

// NewBuyOrder creates a BUY MARKET order
func NewBuyOrder(size float64, timestamp time.Time) *Order {
	return NewMarketOrder(OrderActionBuy, size, timestamp)
}

// NewSellOrder creates a SELL MARKET order
func NewSellOrder(size float64, timestamp time.Time) *Order {
	return NewMarketOrder(OrderActionSell, size, timestamp)
}

// NewHoldOrder creates a HOLD order (do nothing)
func NewHoldOrder(timestamp time.Time) *Order {
	return &Order{
		Action:    OrderActionHold,
		Size:      0,
		OrderType: OrderTypeMarket,
		Timestamp: timestamp,
	}
}

// NewBuyLimitOrder creates a BUY LIMIT order
func NewBuyLimitOrder(size, limitPrice float64, timestamp time.Time) *Order {
	return NewLimitOrder(OrderActionBuy, size, limitPrice, timestamp)
}

// NewSellLimitOrder creates a SELL LIMIT order
func NewSellLimitOrder(size, limitPrice float64, timestamp time.Time) *Order {
	return NewLimitOrder(OrderActionSell, size, limitPrice, timestamp)
}

// ==================== ORDER METHODS ====================

// IsBuy returns true if this is a BUY action
func (o *Order) IsBuy() bool {
	return o.Action == OrderActionBuy
}

// IsSell returns true if this is a SELL action
func (o *Order) IsSell() bool {
	return o.Action == OrderActionSell
}

// IsHold returns true if this is a HOLD action (no trade)
func (o *Order) IsHold() bool {
	return o.Action == OrderActionHold
}

// IsMarket returns true if this is a MARKET order
func (o *Order) IsMarket() bool {
	return o.OrderType == OrderTypeMarket
}

// IsLimit returns true if this is a LIMIT order
func (o *Order) IsLimit() bool {
	return o.OrderType == OrderTypeLimit
}

// IsTradeOrder returns true if this is a BUY or SELL (not HOLD)
func (o *Order) IsTradeOrder() bool {
	return o.Action == OrderActionBuy || o.Action == OrderActionSell
}

// GetDirection returns 1 for BUY, -1 for SELL, 0 for HOLD
func (o *Order) GetDirection() int {
	switch o.Action {
	case OrderActionBuy:
		return 1
	case OrderActionSell:
		return -1
	case OrderActionHold:
		return 0
	default:
		return 0
	}
}

// String returns a human-readable representation of the order
func (o *Order) String() string {
	if o.IsHold() {
		return fmt.Sprintf("Order[HOLD at %s]", o.Timestamp.Format("2006-01-02T15:04:05.000"))
	}

	if o.IsMarket() {
		return fmt.Sprintf("Order[%s MARKET %f at %s]",
			o.Action, o.Size, o.Timestamp.Format("2006-01-02T15:04:05.000"))
	}

	return fmt.Sprintf("Order[%s LIMIT %f @ %.5f at %s]",
		o.Action, o.Size, o.LimitPrice, o.Timestamp.Format("2006-01-02T15:04:05.000"))
}

// DebugString returns detailed order information
func (o *Order) DebugString() string {
	limitInfo := ""
	if o.IsLimit() {
		limitInfo = fmt.Sprintf("\n  Limit Price: %.8f", o.LimitPrice)
	}

	description := ""
	if o.Description != "" {
		description = fmt.Sprintf("\n  Description: %s", o.Description)
	}

	return fmt.Sprintf(
		"Order Details:\n"+
			"  OrderID:     %s\n"+
			"  Action:      %s\n"+
			"  Size:        %f\n"+
			"  OrderType:   %s%s\n"+
			"  Timestamp:   %s%s",
		o.OrderID,
		o.Action,
		o.Size,
		o.OrderType,
		limitInfo,
		o.Timestamp.Format("2006-01-02T15:04:05.000000"),
		description,
	)
}

// ==================== ORDER VALIDATION ====================

// ValidationError holds validation error details
type OrderValidationError struct {
	Code    string
	Message string
}

// Validate checks if the order is valid
// Returns nil if valid, or ValidationError if invalid
func (o *Order) Validate(minLotSize, maxPositionSize float64) *OrderValidationError {
	// Check action
	if !IsValidOrderAction(o.Action) {
		return &OrderValidationError{
			Code:    ErrorCodeInvalidOrderType,
			Message: fmt.Sprintf("invalid action: %s (must be %s, %s, or %s)", o.Action, OrderActionBuy, OrderActionSell, OrderActionHold),
		}
	}

	// HOLD orders don't need further validation
	if o.IsHold() {
		return nil
	}

	// Check order type
	if !IsValidOrderType(o.OrderType) {
		return &OrderValidationError{
			Code:    ErrorCodeInvalidOrderType,
			Message: fmt.Sprintf("invalid order type: %s (must be %s or %s)", o.OrderType, OrderTypeMarket, OrderTypeLimit),
		}
	}

	// Check size is positive
	if o.Size <= 0 {
		return &OrderValidationError{
			Code:    ErrorCodeInvalidOrderSize,
			Message: fmt.Sprintf("order size must be positive, got: %f", o.Size),
		}
	}

	// Check size meets minimum
	if o.Size < minLotSize {
		return &OrderValidationError{
			Code:    ErrorCodeInvalidLotSize,
			Message: fmt.Sprintf("order size %f is less than minimum lot size %f", o.Size, minLotSize),
		}
	}

	// Check size doesn't exceed maximum
	if o.Size > maxPositionSize {
		return &OrderValidationError{
			Code:    ErrorCodePositionLimitExceeded,
			Message: fmt.Sprintf("order size %f exceeds maximum position size %f", o.Size, maxPositionSize),
		}
	}

	// For LIMIT orders, check limit price is positive
	if o.IsLimit() {
		if o.LimitPrice <= 0 {
			return &OrderValidationError{
				Code:    ErrorCodeInvalidLimitPrice,
				Message: fmt.Sprintf("limit price must be positive, got: %f", o.LimitPrice),
			}
		}
	}

	// All checks passed
	return nil
}

// ==================== ORDER COMPARISON ====================

// IsSameAs checks if two orders are functionally the same
func (o *Order) IsSameAs(other *Order) bool {
	if other == nil {
		return false
	}

	if o.Action != other.Action {
		return false
	}

	if o.Size != other.Size {
		return false
	}

	if o.OrderType != other.OrderType {
		return false
	}

	if o.IsLimit() {
		if o.LimitPrice != other.LimitPrice {
			return false
		}
	}

	return true
}

// ==================== ORDER BUILDER PATTERN ====================

// OrderBuilder allows fluent building of orders
type OrderBuilder struct {
	order *Order
	err   error
}

// NewOrderBuilder creates a new order builder
func NewOrderBuilder() *OrderBuilder {
	return &OrderBuilder{
		order: &Order{
			Timestamp: time.Now(),
		},
	}
}

// WithAction sets the order action
func (ob *OrderBuilder) WithAction(action string) *OrderBuilder {
	if ob.err != nil {
		return ob
	}
	if !IsValidOrderAction(action) {
		ob.err = fmt.Errorf("invalid action: %s", action)
		return ob
	}
	ob.order.Action = action
	return ob
}

// WithSize sets the order size
func (ob *OrderBuilder) WithSize(size float64) *OrderBuilder {
	if ob.err != nil {
		return ob
	}
	if size <= 0 {
		ob.err = fmt.Errorf("size must be positive, got %f", size)
		return ob
	}
	ob.order.Size = size
	return ob
}

// WithMarketOrder sets order type to MARKET
func (ob *OrderBuilder) WithMarketOrder() *OrderBuilder {
	if ob.err != nil {
		return ob
	}
	ob.order.OrderType = OrderTypeMarket
	ob.order.LimitPrice = 0
	return ob
}

// WithLimitOrder sets order type to LIMIT with a price
func (ob *OrderBuilder) WithLimitOrder(price float64) *OrderBuilder {
	if ob.err != nil {
		return ob
	}
	if price <= 0 {
		ob.err = fmt.Errorf("limit price must be positive, got %f", price)
		return ob
	}
	ob.order.OrderType = OrderTypeLimit
	ob.order.LimitPrice = price
	return ob
}

// WithTimestamp sets the order timestamp
func (ob *OrderBuilder) WithTimestamp(ts time.Time) *OrderBuilder {
	if ob.err != nil {
		return ob
	}
	ob.order.Timestamp = ts
	return ob
}

// WithDescription sets the order description
func (ob *OrderBuilder) WithDescription(desc string) *OrderBuilder {
	if ob.err != nil {
		return ob
	}
	ob.order.Description = desc
	return ob
}

// Buy shortcut for BUY action
func (ob *OrderBuilder) Buy() *OrderBuilder {
	return ob.WithAction(OrderActionBuy)
}

// Sell shortcut for SELL action
func (ob *OrderBuilder) Sell() *OrderBuilder {
	return ob.WithAction(OrderActionSell)
}

// Build returns the constructed order or error
func (ob *OrderBuilder) Build() (*Order, error) {
	if ob.err != nil {
		return nil, ob.err
	}

	// Validate action is set
	if ob.order.Action == "" {
		return nil, fmt.Errorf("action not set")
	}

	// HOLD orders don't need size
	if ob.order.Action == OrderActionHold {
		return ob.order, nil
	}

	// Other orders need size
	if ob.order.Size == 0 {
		return nil, fmt.Errorf("size not set")
	}

	return ob.order, nil
}

// MustBuild builds the order and panics on error
func (ob *OrderBuilder) MustBuild() *Order {
	order, err := ob.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build order: %v", err))
	}
	return order
}

// ==================== ORDER BATCH ====================

// OrderBatch represents multiple orders
type OrderBatch struct {
	Orders    []*Order
	Timestamp time.Time
}

// NewOrderBatch creates a new batch of orders
func NewOrderBatch(timestamp time.Time) *OrderBatch {
	return &OrderBatch{
		Orders:    make([]*Order, 0),
		Timestamp: timestamp,
	}
}

// Add adds an order to the batch
func (ob *OrderBatch) Add(order *Order) {
	ob.Orders = append(ob.Orders, order)
}

// Size returns the number of orders in batch
func (ob *OrderBatch) Size() int {
	return len(ob.Orders)
}

// GetTradeOrders returns only BUY/SELL orders (no HOLD)
func (ob *OrderBatch) GetTradeOrders() []*Order {
	trades := make([]*Order, 0)
	for _, order := range ob.Orders {
		if order.IsTradeOrder() {
			trades = append(trades, order)
		}
	}
	return trades
}

// HasTradeOrders returns true if batch contains any BUY/SELL orders
func (ob *OrderBatch) HasTradeOrders() bool {
	return len(ob.GetTradeOrders()) > 0
}

// String returns a human-readable representation
func (ob *OrderBatch) String() string {
	return fmt.Sprintf("OrderBatch[%d orders at %s]", ob.Size(), ob.Timestamp.Format("2006-01-02T15:04:05.000"))
}
