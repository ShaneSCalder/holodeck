package executor

import (
	"fmt"
	"time"

	"holodeck/types"
)

// ==================== EXECUTOR ERROR CODES ====================

const (
	// Validation errors
	ErrorCodeInvalidOrderSize      = "INVALID_ORDER_SIZE"
	ErrorCodeInvalidLimitPrice     = "INVALID_LIMIT_PRICE"
	ErrorCodeInsufficientBalance   = "INSUFFICIENT_BALANCE"
	ErrorCodePositionLimitExceeded = "POSITION_LIMIT_EXCEEDED"
	ErrorCodeMarketClosed          = "MARKET_CLOSED"

	// Execution errors
	ErrorCodePartialFill      = "PARTIAL_FILL"
	ErrorCodeNoLiquidity      = "NO_LIQUIDITY"
	ErrorCodeSlippageExceeded = "SLIPPAGE_EXCEEDED"
	ErrorCodeLimitNotHit      = "LIMIT_NOT_HIT"

	// System errors
	ErrorCodeInvalidInstrument = "INVALID_INSTRUMENT"
	ErrorCodeExecutorError     = "EXECUTOR_ERROR"
)

// ==================== EXECUTOR ERRORS ====================

// OrderValidationError is returned when order validation fails
type OrderValidationError struct {
	Code        string
	Message     string
	OrderID     string
	Field       string
	Value       interface{}
	Expected    interface{}
	Timestamp   time.Time
	ParentError error
}

// NewOrderValidationError creates a new order validation error
func NewOrderValidationError(
	code string,
	field string,
	message string,
	value interface{},
) *OrderValidationError {
	return &OrderValidationError{
		Code:      code,
		Message:   message,
		Field:     field,
		Value:     value,
		Timestamp: time.Now(),
	}
}

// Error implements error interface
func (ove *OrderValidationError) Error() string {
	return fmt.Sprintf("[%s] %s: %s (got: %v)", ove.Code, ove.Field, ove.Message, ove.Value)
}

// WithOrderID adds order ID to error
func (ove *OrderValidationError) WithOrderID(orderID string) *OrderValidationError {
	ove.OrderID = orderID
	return ove
}

// WithExpected adds expected value to error
func (ove *OrderValidationError) WithExpected(expected interface{}) *OrderValidationError {
	ove.Expected = expected
	return ove
}

// WithParent adds parent error
func (ove *OrderValidationError) WithParent(err error) *OrderValidationError {
	ove.ParentError = err
	return ove
}

// DebugString returns detailed error information
func (ove *OrderValidationError) DebugString() string {
	return fmt.Sprintf(
		"Order Validation Error:\n"+
			"  Code:         %s\n"+
			"  Message:      %s\n"+
			"  Order ID:     %s\n"+
			"  Field:        %s\n"+
			"  Got Value:    %v\n"+
			"  Expected:     %v\n"+
			"  Timestamp:    %s",
		ove.Code,
		ove.Message,
		ove.OrderID,
		ove.Field,
		ove.Value,
		ove.Expected,
		ove.Timestamp.Format(time.RFC3339),
	)
}

// ==================== EXECUTION ERROR ====================

// ExecutionError is returned when order execution fails
type ExecutionError struct {
	Code          string
	Message       string
	OrderID       string
	Reason        string
	Timestamp     time.Time
	Tick          *types.Tick
	FillPrice     float64
	RequestedSize float64
	FilledSize    float64
	ParentError   error
}

// NewExecutionError creates a new execution error
func NewExecutionError(
	code string,
	message string,
	reason string,
) *ExecutionError {
	return &ExecutionError{
		Code:      code,
		Message:   message,
		Reason:    reason,
		Timestamp: time.Now(),
	}
}

// Error implements error interface
func (ee *ExecutionError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", ee.Code, ee.Message, ee.Reason)
}

// WithOrderID adds order ID
func (ee *ExecutionError) WithOrderID(orderID string) *ExecutionError {
	ee.OrderID = orderID
	return ee
}

// WithTick adds tick context
func (ee *ExecutionError) WithTick(tick *types.Tick) *ExecutionError {
	ee.Tick = tick
	return ee
}

// WithFillInfo adds fill information
func (ee *ExecutionError) WithFillInfo(
	fillPrice float64,
	requested float64,
	filled float64,
) *ExecutionError {
	ee.FillPrice = fillPrice
	ee.RequestedSize = requested
	ee.FilledSize = filled
	return ee
}

// DebugString returns detailed error information
func (ee *ExecutionError) DebugString() string {
	tickInfo := "none"
	if ee.Tick != nil {
		tickInfo = fmt.Sprintf("%.8f/%.8f", ee.Tick.Bid, ee.Tick.Ask)
	}

	return fmt.Sprintf(
		"Execution Error:\n"+
			"  Code:            %s\n"+
			"  Message:         %s\n"+
			"  Reason:          %s\n"+
			"  Order ID:        %s\n"+
			"  Timestamp:       %s\n"+
			"  Tick (bid/ask):  %s\n"+
			"  Fill Price:      %.8f\n"+
			"  Requested Size:  %.6f\n"+
			"  Filled Size:     %.6f",
		ee.Code,
		ee.Message,
		ee.Reason,
		ee.OrderID,
		ee.Timestamp.Format(time.RFC3339),
		tickInfo,
		ee.FillPrice,
		ee.RequestedSize,
		ee.FilledSize,
	)
}

// ==================== LIMIT ORDER ERROR ====================

// LimitOrderError is returned when limit order processing fails
type LimitOrderError struct {
	Code       string
	Message    string
	OrderID    string
	LimitPrice float64
	TickPrice  float64
	Timestamp  time.Time
}

// NewLimitOrderError creates a new limit order error
func NewLimitOrderError(
	code string,
	message string,
	limitPrice float64,
) *LimitOrderError {
	return &LimitOrderError{
		Code:       code,
		Message:    message,
		LimitPrice: limitPrice,
		Timestamp:  time.Now(),
	}
}

// Error implements error interface
func (loe *LimitOrderError) Error() string {
	return fmt.Sprintf("[%s] %s (limit: %.8f)", loe.Code, loe.Message, loe.LimitPrice)
}

// WithTickPrice adds the current tick price
func (loe *LimitOrderError) WithTickPrice(price float64) *LimitOrderError {
	loe.TickPrice = price
	return loe
}

// WithOrderID adds order ID
func (loe *LimitOrderError) WithOrderID(orderID string) *LimitOrderError {
	loe.OrderID = orderID
	return loe
}

// ==================== PARTIAL FILL ERROR ====================

// PartialFillError indicates a partial fill occurred
type PartialFillError struct {
	Code          string
	Message       string
	OrderID       string
	RequestedSize float64
	FilledSize    float64
	Reason        string
	Timestamp     time.Time
}

// NewPartialFillError creates a new partial fill error
func NewPartialFillError(
	orderID string,
	requested float64,
	filled float64,
	reason string,
) *PartialFillError {
	return &PartialFillError{
		Code:          ErrorCodePartialFill,
		Message:       "Order partially filled",
		OrderID:       orderID,
		RequestedSize: requested,
		FilledSize:    filled,
		Reason:        reason,
		Timestamp:     time.Now(),
	}
}

// Error implements error interface
func (pfe *PartialFillError) Error() string {
	fillPercent := (pfe.FilledSize / pfe.RequestedSize) * 100
	return fmt.Sprintf(
		"[%s] %s: %.2f of %.2f (%.1f%%) - %s",
		pfe.Code,
		pfe.Message,
		pfe.FilledSize,
		pfe.RequestedSize,
		fillPercent,
		pfe.Reason,
	)
}

// ==================== SLIPPAGE ERROR ====================

// SlippageError indicates slippage exceeded tolerance
type SlippageError struct {
	Code           string
	Message        string
	OrderID        string
	ActualSlippage float64
	MaxSlippage    float64
	Timestamp      time.Time
}

// NewSlippageError creates a new slippage error
func NewSlippageError(
	orderID string,
	actual float64,
	max float64,
) *SlippageError {
	return &SlippageError{
		Code:           ErrorCodeSlippageExceeded,
		Message:        "Slippage exceeded maximum tolerance",
		OrderID:        orderID,
		ActualSlippage: actual,
		MaxSlippage:    max,
		Timestamp:      time.Now(),
	}
}

// Error implements error interface
func (se *SlippageError) Error() string {
	return fmt.Sprintf(
		"[%s] %s: actual %.8f > max %.8f",
		se.Code,
		se.Message,
		se.ActualSlippage,
		se.MaxSlippage,
	)
}

// ==================== POSITION LIMIT ERROR ====================

// PositionLimitError indicates position size would exceed limits
type PositionLimitError struct {
	Code        string
	Message     string
	OrderID     string
	RequestSize float64
	CurrentSize float64
	MaxSize     float64
	Timestamp   time.Time
}

// NewPositionLimitError creates a new position limit error
func NewPositionLimitError(
	orderID string,
	requestSize float64,
	currentSize float64,
	maxSize float64,
) *PositionLimitError {
	return &PositionLimitError{
		Code:        ErrorCodePositionLimitExceeded,
		Message:     "Position would exceed maximum size",
		OrderID:     orderID,
		RequestSize: requestSize,
		CurrentSize: currentSize,
		MaxSize:     maxSize,
		Timestamp:   time.Now(),
	}
}

// Error implements error interface
func (ple *PositionLimitError) Error() string {
	resultSize := ple.CurrentSize + ple.RequestSize
	return fmt.Sprintf(
		"[%s] %s: %.2f + %.2f = %.2f > max %.2f",
		ple.Code,
		ple.Message,
		ple.CurrentSize,
		ple.RequestSize,
		resultSize,
		ple.MaxSize,
	)
}

// ==================== ERROR CONVERTERS ====================

// ConvertToHolodeckError converts an executor error to HolodeckError
func ConvertToHolodeckError(err error) *types.HolodeckError {
	if herr, ok := err.(*types.HolodeckError); ok {
		return herr
	}

	// Create a generic executor error
	return types.NewOrderRejectedError(err.Error())
}
