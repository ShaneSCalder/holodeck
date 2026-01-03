package logger

import (
	"time"

	"holodeck/types"
)

// ==================== LOGGER INTERFACE ====================

// Logger defines the contract for all logging implementations
type Logger interface {
	// Core logging methods
	LogTrade(trade *TradeLog) error
	LogError(errLog *ErrorLog) error
	LogMetrics(metrics *MetricsLog) error
	LogInfo(message string) error
	LogWarning(message string) error
	LogDebug(message string) error

	// Session management
	StartSession(sessionID string) error
	EndSession(sessionID string) error
	GetSessionID() string

	// Control methods
	SetVerbosity(level VerbosityLevel) error
	Flush() error
	Close() error
}

// ==================== VERBOSITY LEVELS ====================

// VerbosityLevel defines logging verbosity
type VerbosityLevel int

const (
	VerbosityQuiet   VerbosityLevel = iota // Only errors
	VerbosityMinimal                       // Trades and errors
	VerbosityNormal                        // Trades, errors, metrics
	VerbosityVerbose                       // Full details
	VerbosityDebug                         // All including debug
)

// String returns string representation of verbosity level
func (vl VerbosityLevel) String() string {
	switch vl {
	case VerbosityQuiet:
		return "QUIET"
	case VerbosityMinimal:
		return "MINIMAL"
	case VerbosityNormal:
		return "NORMAL"
	case VerbosityVerbose:
		return "VERBOSE"
	case VerbosityDebug:
		return "DEBUG"
	default:
		return "UNKNOWN"
	}
}

// ==================== TRADE LOG ====================

// TradeLog represents a single trade entry
type TradeLog struct {
	Timestamp     time.Time
	TradeID       string
	OrderID       string
	Instrument    string
	Action        string // BUY, SELL, HOLD
	OrderType     string // MARKET, LIMIT
	RequestedSize float64
	FilledSize    float64
	FillPrice     float64
	Commission    float64
	Slippage      float64
	RealizedPnL   float64
	Status        string // FILLED, PARTIAL, REJECTED
	ErrorMessage  string
	EntryPrice    float64
	CurrentPrice  float64
	PositionSize  float64
	PositionValue float64
	UnrealizedPnL float64
}

// ==================== ERROR LOG ====================

// ErrorLog represents an error entry
type ErrorLog struct {
	Timestamp  time.Time
	ErrorCode  string
	ErrorType  string
	Message    string
	Details    string
	TradeID    string
	OrderID    string
	StackTrace string
	Severity   ErrorSeverity
}

// ErrorSeverity defines error severity levels
type ErrorSeverity int

const (
	SeverityInfo ErrorSeverity = iota
	SeverityWarning
	SeverityError
	SeverityCritical
)

// String returns string representation
func (es ErrorSeverity) String() string {
	switch es {
	case SeverityInfo:
		return "INFO"
	case SeverityWarning:
		return "WARNING"
	case SeverityError:
		return "ERROR"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// ==================== METRICS LOG ====================

// MetricsLog represents periodic metrics entry
type MetricsLog struct {
	Timestamp          time.Time
	SessionID          string
	SessionDuration    time.Duration
	InitialBalance     float64
	CurrentBalance     float64
	TotalPnL           float64
	TotalPnLPercent    float64
	TradeCount         int64
	WinningTrades      int64
	LosingTrades       int64
	WinRate            float64
	MaxDrawdown        float64
	MaxDrawdownPercent float64
	CommissionTotal    float64
	SlippageTotal      float64
	AverageTradePnL    float64
	LargestWin         float64
	LargestLoss        float64
	MeanWin            float64
	MeanLoss           float64
	ProfitFactor       float64
	SharpeRatio        float64
	MDD                float64 // Maximum Drawdown
	MWL                int64   // Maximum Winning Streak Length
	MLS                int64   // Maximum Losing Streak Length
	AvgHoldTime        time.Duration
	TicksProcessed     int64
	ErrorCount         int64
	RejectedOrders     int64
}

// ==================== INFO LOG ====================

// InfoLog represents an informational entry
type InfoLog struct {
	Timestamp time.Time
	Level     LogLevel
	Message   string
	Details   map[string]interface{}
}

// LogLevel defines log levels
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarning
	LevelError
)

// String returns string representation
func (ll LogLevel) String() string {
	switch ll {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarning:
		return "WARNING"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ==================== SESSION INFO ====================

// SessionInfo represents session information
type SessionInfo struct {
	SessionID       string
	StartTime       time.Time
	EndTime         time.Time
	InitialBalance  float64
	FinalBalance    float64
	TotalTrades     int64
	TotalP_L        float64
	TotalCommission float64
	TotalSlippage   float64
}

// ==================== LOG ENTRY ====================

// LogEntry represents a generic log entry
type LogEntry struct {
	Timestamp time.Time
	Level     LogLevel
	Category  string
	Message   string
	TradeID   string
	OrderID   string
	ErrorCode string
	Details   string
	UserData  map[string]interface{}
}

// ==================== NO-OP LOGGER (for testing) ====================

// NoOpLogger is a logger that does nothing (for testing)
type NoOpLogger struct {
	sessionID string
}

// NewNoOpLogger creates a new no-op logger
func NewNoOpLogger() Logger {
	return &NoOpLogger{}
}

// LogTrade logs a trade (no-op)
func (nol *NoOpLogger) LogTrade(trade *TradeLog) error {
	return nil
}

// LogError logs an error (no-op)
func (nol *NoOpLogger) LogError(errLog *ErrorLog) error {
	return nil
}

// LogMetrics logs metrics (no-op)
func (nol *NoOpLogger) LogMetrics(metrics *MetricsLog) error {
	return nil
}

// LogInfo logs info (no-op)
func (nol *NoOpLogger) LogInfo(message string) error {
	return nil
}

// LogWarning logs warning (no-op)
func (nol *NoOpLogger) LogWarning(message string) error {
	return nil
}

// LogDebug logs debug (no-op)
func (nol *NoOpLogger) LogDebug(message string) error {
	return nil
}

// StartSession starts a session (no-op)
func (nol *NoOpLogger) StartSession(sessionID string) error {
	nol.sessionID = sessionID
	return nil
}

// EndSession ends a session (no-op)
func (nol *NoOpLogger) EndSession(sessionID string) error {
	return nil
}

// GetSessionID returns session ID
func (nol *NoOpLogger) GetSessionID() string {
	return nol.sessionID
}

// SetVerbosity sets verbosity (no-op)
func (nol *NoOpLogger) SetVerbosity(level VerbosityLevel) error {
	return nil
}

// Flush flushes logs (no-op)
func (nol *NoOpLogger) Flush() error {
	return nil
}

// Close closes logger (no-op)
func (nol *NoOpLogger) Close() error {
	return nil
}

// ==================== HELPER FUNCTIONS ====================

// NewTradeLog creates a new trade log from execution report
func NewTradeLog(
	tradeID string,
	report *types.ExecutionReport,
	instrument types.Instrument,
) *TradeLog {

	return &TradeLog{
		Timestamp:     time.Now(),
		TradeID:       tradeID,
		OrderID:       report.OrderID,
		Instrument:    instrument.GetSymbol(),
		Action:        report.Action,
		RequestedSize: report.RequestedSize,
		FilledSize:    report.FilledSize,
		FillPrice:     report.FillPrice,
		Commission:    report.Commission,
		Slippage:      report.SlippageUnits,
		RealizedPnL:   report.RealizedPnL,
		Status:        report.Status,
		ErrorMessage:  report.ErrorMessage,
	}
}

// NewErrorLog creates a new error log from error
func NewErrorLog(
	err error,
	severity ErrorSeverity,
) *ErrorLog {

	return &ErrorLog{
		Timestamp: time.Now(),
		Message:   err.Error(),
		Severity:  severity,
	}
}

// NewMetricsLog creates a new metrics log
func NewMetricsLog(sessionID string) *MetricsLog {
	return &MetricsLog{
		Timestamp: time.Now(),
		SessionID: sessionID,
	}
}
