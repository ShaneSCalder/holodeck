// This file should be renamed to validation.go
// The content is correct, just the filename needs to match spec

package executor

import (
	"fmt"

	"holodeck/types"
)

// ==================== ORDER VALIDATOR ====================

// OrderValidator validates orders for execution
type OrderValidator struct{}

// NewOrderValidator creates a new order validator
func NewOrderValidator() OrderValidator {
	return OrderValidator{}
}

// ValidateOrder validates an order against all rules
func (ov OrderValidator) ValidateOrder(
	order *types.Order,
	instrument types.Instrument,
	availableBalance float64,
	minOrderSize float64,
	maxOrderSize float64,
	maxPositionSize float64,
) error {

	// Check order is not nil
	if order == nil {
		return types.NewOrderRejectedError("order cannot be nil")
	}

	// Check action is valid
	if !types.IsValidOrderAction(order.Action) {
		return types.NewInvalidOrderTypeError(order.Action)
	}

	// HOLD orders don't need further validation
	if order.IsHold() {
		return nil
	}

	// Check order type is valid
	if !types.IsValidOrderType(order.OrderType) {
		return types.NewInvalidOrderTypeError(order.OrderType)
	}

	// Check size is positive
	if order.Size <= 0 {
		return types.NewInvalidOrderSizeError(order.Size, minOrderSize)
	}

	// Check size meets minimum
	if order.Size < minOrderSize {
		return types.NewInvalidLotSizeError(order.Size, minOrderSize)
	}

	// Check size doesn't exceed maximum
	if order.Size > maxOrderSize {
		return types.NewPositionLimitError(order.Size, maxOrderSize)
	}

	// Check position size limit
	if order.Size > maxPositionSize {
		return types.NewPositionLimitError(order.Size, maxPositionSize)
	}

	// Validate instrument
	if instrument == nil {
		return types.NewInstrumentNotFoundError("unknown")
	}

	// Validate order size against instrument
	if err := instrument.ValidateOrderSize(order.Size); err != nil {
		return err
	}

	// Check limit price for LIMIT orders
	if order.IsLimit() {
		if order.LimitPrice <= 0 {
			return types.NewInvalidLimitPriceError(
				order.LimitPrice,
				"limit price must be positive",
			)
		}
	}

	// Check available balance (simple check, doesn't account for leverage yet)
	notionalCost := order.Size * 100 // Approximate cost
	if notionalCost > availableBalance {
		return types.NewInsufficientBalanceError(notionalCost, availableBalance)
	}

	return nil
}

// ValidateOrderSize validates just the order size
func (ov OrderValidator) ValidateOrderSize(
	size float64,
	minSize float64,
	maxSize float64,
	instrument types.Instrument,
) error {

	if size <= 0 {
		return types.NewInvalidOrderSizeError(size, minSize)
	}

	if size < minSize {
		return types.NewInvalidLotSizeError(size, minSize)
	}

	if size > maxSize {
		return types.NewPositionLimitError(size, maxSize)
	}

	if instrument != nil {
		return instrument.ValidateOrderSize(size)
	}

	return nil
}

// ValidateLimitPrice validates a limit price
func (ov OrderValidator) ValidateLimitPrice(
	limitPrice float64,
	currentPrice float64,
	action string,
	instrument types.Instrument,
) error {

	if limitPrice <= 0 {
		return types.NewInvalidLimitPriceError(
			limitPrice,
			"limit price must be positive",
		)
	}

	if instrument != nil {
		return instrument.ValidateLimitPrice(limitPrice, currentPrice, action)
	}

	return nil
}

// ==================== PRICE VALIDATION ====================

// ValidateFillPrice validates a fill price is reasonable
func ValidateFillPrice(
	fillPrice float64,
	bid float64,
	ask float64,
	pipValue float64,
) error {

	maxSpread := ask - bid + (ask * 0.01) // Allow 1% wider spread

	if fillPrice < bid-maxSpread || fillPrice > ask+maxSpread {
		return types.NewConfigError(
			"fillPrice",
			fmt.Sprintf("fill price %.8f outside reasonable range [%.8f, %.8f]",
				fillPrice, bid-maxSpread, ask+maxSpread),
		)
	}

	return nil
}

// ==================== BALANCE VALIDATION ====================

// ValidateBalance checks if balance can support a trade
func ValidateBalance(
	availableBalance float64,
	requiredBalance float64,
	leverage float64,
) error {

	buyingPower := availableBalance * leverage

	if requiredBalance > buyingPower {
		return types.NewInsufficientBalanceError(requiredBalance, buyingPower)
	}

	return nil
}

// ==================== POSITION VALIDATION ====================

// ValidatePosition checks if position is valid after execution
func ValidatePosition(
	newPositionSize float64,
	maxPositionSize float64,
) error {

	absSize := newPositionSize
	if absSize < 0 {
		absSize = -absSize
	}

	if absSize > maxPositionSize {
		return types.NewPositionLimitError(absSize, maxPositionSize)
	}

	return nil
}
