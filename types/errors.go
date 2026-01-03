package types

import (
	"fmt"
	"time"
)

// ==================== ERROR INTERFACE ====================

// HolodeckError is the standard error type for Holodeck
type HolodeckError struct {
	// Code is the error code (e.g., ErrorCodeInsufficientBalance)
	Code string

	// Message is the human-readable error message
	Message string

	// Details is additional context about the error
	Details map[string]interface{}

	// Timestamp is when the error occurred
	Timestamp time.Time

	// SourceFunc is the function that generated the error
	SourceFunc string

	// SourceFile is the file that generated the error
	SourceFile string

	// SourceLine is the line number that generated the error
	SourceLine int

	// ParentError is the underlying error if wrapping another error
	ParentError error
}

// ==================== ERROR CONSTRUCTORS ====================

// NewHolodeckError creates a new Holodeck error
func NewHolodeckError(code, message string) *HolodeckError {
	return &HolodeckError{
		Code:      code,
		Message:   message,
		Details:   make(map[string]interface{}),
		Timestamp: time.Now(),
	}
}

// NewInsufficientBalanceError creates an INSUFFICIENT_BALANCE error
func NewInsufficientBalanceError(required, available float64) *HolodeckError {
	err := NewHolodeckError(
		ErrorCodeInsufficientBalance,
		fmt.Sprintf("insufficient balance: required %.2f, available %.2f", required, available),
	)
	err.Details["required"] = required
	err.Details["available"] = available
	err.Details["shortfall"] = required - available
	return err
}

// NewPositionLimitError creates a POSITION_LIMIT_EXCEEDED error
func NewPositionLimitError(requested, maxAllowed float64) *HolodeckError {
	err := NewHolodeckError(
		ErrorCodePositionLimitExceeded,
		fmt.Sprintf("position limit exceeded: requested %.2f, max %.2f", requested, maxAllowed),
	)
	err.Details["requested"] = requested
	err.Details["max_allowed"] = maxAllowed
	err.Details["excess"] = requested - maxAllowed
	return err
}

// NewInvalidOrderTypeError creates an INVALID_ORDER_TYPE error
func NewInvalidOrderTypeError(orderType string) *HolodeckError {
	err := NewHolodeckError(
		ErrorCodeInvalidOrderType,
		fmt.Sprintf("invalid order type: %s (must be %s or %s)", orderType, OrderTypeMarket, OrderTypeLimit),
	)
	err.Details["provided_type"] = orderType
	err.Details["valid_types"] = []string{OrderTypeMarket, OrderTypeLimit}
	return err
}

// NewInvalidLimitPriceError creates an INVALID_LIMIT_PRICE error
func NewInvalidLimitPriceError(limitPrice float64, reason string) *HolodeckError {
	err := NewHolodeckError(
		ErrorCodeInvalidLimitPrice,
		fmt.Sprintf("invalid limit price: %.8f (%s)", limitPrice, reason),
	)
	err.Details["limit_price"] = limitPrice
	err.Details["reason"] = reason
	return err
}

// NewInvalidOrderSizeError creates an INVALID_ORDER_SIZE error
func NewInvalidOrderSizeError(size float64, minSize float64) *HolodeckError {
	err := NewHolodeckError(
		ErrorCodeInvalidOrderSize,
		fmt.Sprintf("invalid order size: %.2f (min: %.2f)", size, minSize),
	)
	err.Details["provided_size"] = size
	err.Details["minimum_size"] = minSize
	return err
}

// NewInvalidLotSizeError creates an INVALID_LOT_SIZE error
func NewInvalidLotSizeError(size float64, minimumLotSize float64) *HolodeckError {
	err := NewHolodeckError(
		ErrorCodeInvalidLotSize,
		fmt.Sprintf("invalid lot size: %.6f (minimum: %.6f)", size, minimumLotSize),
	)
	err.Details["provided_size"] = size
	err.Details["minimum_lot_size"] = minimumLotSize
	return err
}

// NewOrderRejectedError creates an ORDER_REJECTED error
func NewOrderRejectedError(reason string) *HolodeckError {
	err := NewHolodeckError(
		ErrorCodeOrderRejected,
		fmt.Sprintf("order rejected: %s", reason),
	)
	err.Details["reason"] = reason
	return err
}

// NewAccountBlownError creates an ACCOUNT_BLOWN error
func NewAccountBlownError(currentDrawdown, maxDrawdown float64) *HolodeckError {
	err := NewHolodeckError(
		ErrorCodeAccountBlown,
		fmt.Sprintf("account blown: drawdown %.2f%% exceeds limit %.2f%%", currentDrawdown, maxDrawdown),
	)
	err.Details["current_drawdown"] = currentDrawdown
	err.Details["max_drawdown"] = maxDrawdown
	return err
}

// NewInvalidOperationError creates an INVALID_OPERATION error
func NewInvalidOperationError(operation, reason string) *HolodeckError {
	err := NewHolodeckError(
		ErrorCodeInvalidOperation,
		fmt.Sprintf("invalid operation %s: %s", operation, reason),
	)
	err.Details["operation"] = operation
	err.Details["reason"] = reason
	return err
}

// NewCSVReadError creates a CSV_READ_ERROR error
func NewCSVReadError(filename string, lineNumber int, reason string) *HolodeckError {
	err := NewHolodeckError(
		ErrorCodeCSVReadError,
		fmt.Sprintf("CSV read error in %s at line %d: %s", filename, lineNumber, reason),
	)
	err.Details["filename"] = filename
	err.Details["line_number"] = lineNumber
	err.Details["reason"] = reason
	return err
}

// NewConfigError creates a CONFIG_ERROR error
func NewConfigError(field, reason string) *HolodeckError {
	err := NewHolodeckError(
		ErrorCodeConfigError,
		fmt.Sprintf("configuration error in %s: %s", field, reason),
	)
	err.Details["field"] = field
	err.Details["reason"] = reason
	return err
}

// NewInstrumentNotFoundError creates an INSTRUMENT_NOT_FOUND error
func NewInstrumentNotFoundError(instrumentType string) *HolodeckError {
	err := NewHolodeckError(
		ErrorCodeInstrumentNotFound,
		fmt.Sprintf("instrument not found: %s", instrumentType),
	)
	err.Details["instrument_type"] = instrumentType
	return err
}

// NewInvalidInstrumentTypeError creates an INVALID_INSTRUMENT_TYPE error
func NewInvalidInstrumentTypeError(instrumentType string) *HolodeckError {
	err := NewHolodeckError(
		ErrorCodeInvalidInstrumentType,
		fmt.Sprintf("invalid instrument type: %s", instrumentType),
	)
	err.Details["provided_type"] = instrumentType
	err.Details["valid_types"] = []string{
		InstrumentTypeForex,
		InstrumentTypeStocks,
		InstrumentTypeCommodities,
		InstrumentTypeCrypto,
	}
	return err
}

// ==================== ERROR METHODS ====================

// WithDetail adds a detail to the error
func (e *HolodeckError) WithDetail(key string, value interface{}) *HolodeckError {
	e.Details[key] = value
	return e
}

// WithDetails adds multiple details to the error
func (e *HolodeckError) WithDetails(details map[string]interface{}) *HolodeckError {
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// WithParent wraps another error
func (e *HolodeckError) WithParent(parent error) *HolodeckError {
	e.ParentError = parent
	return e
}

// WithSource sets the source location of the error
func (e *HolodeckError) WithSource(funcName, fileName string, lineNumber int) *HolodeckError {
	e.SourceFunc = funcName
	e.SourceFile = fileName
	e.SourceLine = lineNumber
	return e
}

// Error implements the error interface
func (e *HolodeckError) Error() string {
	if e.ParentError != nil {
		return fmt.Sprintf("[%s] %s (caused by: %v)", e.Code, e.Message, e.ParentError)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// String returns a detailed string representation
func (e *HolodeckError) String() string {
	return e.DebugString()
}

// DebugString returns a detailed debug representation
func (e *HolodeckError) DebugString() string {
	source := ""
	if e.SourceFunc != "" {
		source = fmt.Sprintf("\n  Source: %s in %s (line %d)", e.SourceFunc, e.SourceFile, e.SourceLine)
	}

	parent := ""
	if e.ParentError != nil {
		parent = fmt.Sprintf("\n  Caused by: %v", e.ParentError)
	}

	details := ""
	if len(e.Details) > 0 {
		details = "\n  Details:"
		for k, v := range e.Details {
			details += fmt.Sprintf("\n    %s: %v", k, v)
		}
	}

	return fmt.Sprintf(
		"HolodeckError:\n"+
			"  Code: %s\n"+
			"  Message: %s\n"+
			"  Timestamp: %s%s%s%s",
		e.Code,
		e.Message,
		e.Timestamp.Format("2006-01-02T15:04:05.000"),
		source,
		parent,
		details,
	)
}

// ==================== ERROR TYPE CHECKS ====================

// IsInsufficientBalance checks if error is insufficient balance
func (e *HolodeckError) IsInsufficientBalance() bool {
	return e.Code == ErrorCodeInsufficientBalance
}

// IsPositionLimitExceeded checks if error is position limit exceeded
func (e *HolodeckError) IsPositionLimitExceeded() bool {
	return e.Code == ErrorCodePositionLimitExceeded
}

// IsInvalidOrderType checks if error is invalid order type
func (e *HolodeckError) IsInvalidOrderType() bool {
	return e.Code == ErrorCodeInvalidOrderType
}

// IsOrderRejected checks if error is order rejected
func (e *HolodeckError) IsOrderRejected() bool {
	return e.Code == ErrorCodeOrderRejected
}

// IsAccountBlown checks if error is account blown
func (e *HolodeckError) IsAccountBlown() bool {
	return e.Code == ErrorCodeAccountBlown
}

// IsCritical checks if error is critical (account blown)
func (e *HolodeckError) IsCritical() bool {
	return e.IsAccountBlown()
}

// IsRetryable checks if error is retryable
func (e *HolodeckError) IsRetryable() bool {
	// Most errors are not retryable
	switch e.Code {
	case ErrorCodeInsufficientBalance, ErrorCodePositionLimitExceeded,
		ErrorCodeAccountBlown, ErrorCodeInvalidOperation:
		return false
	default:
		return false
	}
}

// ==================== ERROR COLLECTION ====================

// ErrorLog collects multiple errors
type ErrorLog struct {
	Errors []*HolodeckError
}

// NewErrorLog creates a new error log
func NewErrorLog() *ErrorLog {
	return &ErrorLog{
		Errors: make([]*HolodeckError, 0),
	}
}

// Add adds an error to the log
func (el *ErrorLog) Add(err *HolodeckError) {
	if err != nil {
		el.Errors = append(el.Errors, err)
	}
}

// Size returns the number of errors in the log
func (el *ErrorLog) Size() int {
	return len(el.Errors)
}

// IsEmpty returns true if no errors
func (el *ErrorLog) IsEmpty() bool {
	return len(el.Errors) == 0
}

// HasErrors returns true if there are errors
func (el *ErrorLog) HasErrors() bool {
	return len(el.Errors) > 0
}

// GetLatest returns the most recent error
func (el *ErrorLog) GetLatest() *HolodeckError {
	if len(el.Errors) == 0 {
		return nil
	}
	return el.Errors[len(el.Errors)-1]
}

// GetOldest returns the oldest error
func (el *ErrorLog) GetOldest() *HolodeckError {
	if len(el.Errors) == 0 {
		return nil
	}
	return el.Errors[0]
}

// GetByCritical returns errors filtered by critical status
func (el *ErrorLog) GetByCritical() []*HolodeckError {
	critical := make([]*HolodeckError, 0)
	for _, err := range el.Errors {
		if err.IsCritical() {
			critical = append(critical, err)
		}
	}
	return critical
}

// GetByCode returns errors with a specific code
func (el *ErrorLog) GetByCode(code string) []*HolodeckError {
	filtered := make([]*HolodeckError, 0)
	for _, err := range el.Errors {
		if err.Code == code {
			filtered = append(filtered, err)
		}
	}
	return filtered
}

// Clear empties the error log
func (el *ErrorLog) Clear() {
	el.Errors = make([]*HolodeckError, 0)
}

// String returns a summary of all errors
func (el *ErrorLog) String() string {
	if len(el.Errors) == 0 {
		return "ErrorLog: no errors"
	}

	summary := fmt.Sprintf("ErrorLog: %d errors\n", len(el.Errors))
	for i, err := range el.Errors {
		summary += fmt.Sprintf("  [%d] %s\n", i+1, err.Error())
	}
	return summary
}

// DebugString returns detailed error log information
func (el *ErrorLog) DebugString() string {
	if len(el.Errors) == 0 {
		return "ErrorLog: no errors"
	}

	summary := fmt.Sprintf("ErrorLog: %d errors\n\n", len(el.Errors))
	for i, err := range el.Errors {
		summary += fmt.Sprintf("Error %d:\n%s\n\n", i+1, err.DebugString())
	}
	return summary
}

// ==================== ERROR HELPERS ====================

// IsHolodeckError checks if an error is a HolodeckError
func IsHolodeckError(err error) bool {
	_, ok := err.(*HolodeckError)
	return ok
}

// AsHolodeckError converts an error to HolodeckError if possible
func AsHolodeckError(err error) (*HolodeckError, bool) {
	he, ok := err.(*HolodeckError)
	return he, ok
}

// ==================== ERROR BUILDER ====================

// ErrorBuilder allows building errors fluently
type ErrorBuilder struct {
	code    string
	message string
	details map[string]interface{}
	parent  error
	source  struct {
		funcName string
		fileName string
		lineNum  int
	}
}

// NewErrorBuilder creates a new error builder
func NewErrorBuilder(code, message string) *ErrorBuilder {
	return &ErrorBuilder{
		code:    code,
		message: message,
		details: make(map[string]interface{}),
	}
}

// WithDetail adds a detail
func (eb *ErrorBuilder) WithDetail(key string, value interface{}) *ErrorBuilder {
	eb.details[key] = value
	return eb
}

// WithParent adds a parent error
func (eb *ErrorBuilder) WithParent(err error) *ErrorBuilder {
	eb.parent = err
	return eb
}

// WithSource adds source location information
func (eb *ErrorBuilder) WithSource(funcName, fileName string, lineNum int) *ErrorBuilder {
	eb.source.funcName = funcName
	eb.source.fileName = fileName
	eb.source.lineNum = lineNum
	return eb
}

// Build builds the error
func (eb *ErrorBuilder) Build() *HolodeckError {
	err := NewHolodeckError(eb.code, eb.message)
	err.Details = eb.details

	if eb.parent != nil {
		err.ParentError = eb.parent
	}

	if eb.source.funcName != "" {
		err.WithSource(eb.source.funcName, eb.source.fileName, eb.source.lineNum)
	}

	return err
}

// ==================== ERROR SUMMARY ====================

// ErrorSummary provides a summary of errors
type ErrorSummary struct {
	TotalErrors    int
	CriticalErrors int
	ErrorCounts    map[string]int // Count by error code
	FirstError     *HolodeckError
	LastError      *HolodeckError
}

// SummarizeErrors creates a summary from an error log
func SummarizeErrors(el *ErrorLog) *ErrorSummary {
	summary := &ErrorSummary{
		TotalErrors:    el.Size(),
		ErrorCounts:    make(map[string]int),
		FirstError:     el.GetOldest(),
		LastError:      el.GetLatest(),
		CriticalErrors: len(el.GetByCritical()),
	}

	for _, err := range el.Errors {
		summary.ErrorCounts[err.Code]++
	}

	return summary
}

// String returns a string representation of the error summary
func (es *ErrorSummary) String() string {
	return fmt.Sprintf(
		"ErrorSummary: %d total errors (%d critical)",
		es.TotalErrors,
		es.CriticalErrors,
	)
}

// DebugString returns a detailed error summary
func (es *ErrorSummary) DebugString() string {
	summary := fmt.Sprintf(
		"Error Summary:\n"+
			"  Total Errors:       %d\n"+
			"  Critical Errors:    %d\n",
		es.TotalErrors,
		es.CriticalErrors,
	)

	if es.FirstError != nil {
		summary += fmt.Sprintf("  First Error:        [%s] %s\n", es.FirstError.Code, es.FirstError.Message)
	}

	if es.LastError != nil {
		summary += fmt.Sprintf("  Last Error:         [%s] %s\n", es.LastError.Code, es.LastError.Message)
	}

	if len(es.ErrorCounts) > 0 {
		summary += "  Error Counts:\n"
		for code, count := range es.ErrorCounts {
			summary += fmt.Sprintf("    %s: %d\n", code, count)
		}
	}

	return summary
}
