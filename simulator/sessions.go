package simulator

import (
	"fmt"
	"sync"
	"time"

	"holodeck/types"
)

// ==================== HOLODECK CONFIGURATION ====================

// HolodeckConfig is the main configuration for the Holodeck system
type HolodeckConfig struct {
	// Config is the loaded JSON configuration
	Config *Config

	// Instrument is the active instrument
	Instrument types.Instrument

	// Session information
	SessionID string
	StartTime time.Time
	EndTime   time.Time
	IsRunning bool

	// Execution parameters
	ExecutionConfig ExecutionParameters

	// Data source
	DataSource DataSourceConfig

	// State tracking
	StateConfig StateConfiguration
}

// ExecutionParameters holds all execution-related parameters
type ExecutionParameters struct {
	// Commission settings
	CommissionEnabled bool
	CommissionType    string
	CommissionValue   float64

	// Slippage settings
	SlippageEnabled bool
	SlippageModel   string

	// Latency settings
	LatencyEnabled bool
	LatencyMs      int64

	// Partial fill settings
	PartialFillsEnabled bool
	PartialFillLogic    string

	// Speed control
	SpeedMultiplier float64

	// Order type defaults
	DefaultOrderType string
	SupportedTypes   []string
}

// DataSourceConfig holds data source configuration
type DataSourceConfig struct {
	FilePath string
	Format   string // CSV, JSON, etc
}

// StateConfiguration holds state tracking configuration
type StateConfiguration struct {
	MaxTicksToKeep          int
	MaxPositionHistorySize  int
	MaxBalanceHistorySize   int
	MaxExecutionHistorySize int
}

// ==================== HOLODECK STATE ====================

// HolodeckState represents the current state of Holodeck
type HolodeckState struct {
	// Configuration
	Config *HolodeckConfig

	// Mutable state (protected by mutex)
	mu sync.RWMutex

	// Current tick
	CurrentTick *types.Tick
	TickCount   int64

	// Position tracking
	Position *types.Position

	// Account tracking
	Balance *types.Balance

	// Execution history
	ExecutionHistory []*types.ExecutionReport
	ExecutionCount   int

	// Error tracking
	ErrorLog *types.ErrorLog

	// Performance metrics
	StartBalance   float64
	CurrentBalance float64
	PeakBalance    float64
	TroughBalance  float64
	TotalPnL       float64

	// Timing
	LastUpdateTime time.Time
	SessionStart   time.Time
	SessionEnd     time.Time
}

// ==================== HOLODECK INITIALIZATION ====================

// NewHolodeckConfig creates a new Holodeck configuration from Config
func NewHolodeckConfig(config *Config) (*HolodeckConfig, error) {
	// Validate config
	if config == nil {
		return nil, types.NewConfigError("config", "configuration cannot be nil")
	}

	// Create instrument
	instrument, err := types.NewInstrument(
		config.Instrument.Type,
		config.Instrument.Symbol,
		config.Instrument.Description,
	)
	if err != nil {
		return nil, err
	}

	// Create execution parameters
	execParams := ExecutionParameters{
		CommissionEnabled:   config.Execution.Commission,
		CommissionType:      config.Execution.CommissionType,
		CommissionValue:     config.Execution.CommissionValue,
		SlippageEnabled:     config.Execution.Slippage,
		SlippageModel:       config.Execution.SlippageModel,
		LatencyEnabled:      config.Execution.Latency,
		LatencyMs:           config.Execution.LatencyMs,
		PartialFillsEnabled: config.Execution.PartialFills,
		PartialFillLogic:    config.Execution.PartialFillBasedOn,
		SpeedMultiplier:     config.Speed.Multiplier,
		DefaultOrderType:    config.OrderTypes.Default,
		SupportedTypes:      config.OrderTypes.Supported,
	}

	// Create data source config
	dataSource := DataSourceConfig{
		FilePath: config.CSV.FilePath,
		Format:   "CSV",
	}

	// Create state config with reasonable defaults
	stateConfig := StateConfiguration{
		MaxTicksToKeep:          10000,
		MaxPositionHistorySize:  1000,
		MaxBalanceHistorySize:   1000,
		MaxExecutionHistorySize: 10000,
	}

	// Create Holodeck config
	hConfig := &HolodeckConfig{
		Config:          config,
		Instrument:      instrument,
		SessionID:       generateSessionID(),
		StartTime:       time.Now(),
		ExecutionConfig: execParams,
		DataSource:      dataSource,
		StateConfig:     stateConfig,
	}

	return hConfig, nil
}

// NewHolodeckState creates a new Holodeck state
func NewHolodeckState(hConfig *HolodeckConfig) (*HolodeckState, error) {
	if hConfig == nil {
		return nil, types.NewConfigError("hConfig", "Holodeck configuration cannot be nil")
	}

	// Create balance
	balance := types.NewBalance(
		hConfig.Config.Account.InitialBalance,
		hConfig.Config.Account.Currency,
		hConfig.Config.Account.Leverage,
		hConfig.Config.Account.MaxDrawdownPercent,
		hConfig.Config.Account.MaxPositionSize,
	)

	// Create position
	position := types.NewPosition()

	// Create error log
	errorLog := types.NewErrorLog()

	// Create state
	now := time.Now()
	state := &HolodeckState{
		Config:           hConfig,
		CurrentTick:      nil,
		TickCount:        0,
		Position:         position,
		Balance:          balance,
		ExecutionHistory: make([]*types.ExecutionReport, 0, hConfig.StateConfig.MaxExecutionHistorySize),
		ExecutionCount:   0,
		ErrorLog:         errorLog,
		StartBalance:     hConfig.Config.Account.InitialBalance,
		CurrentBalance:   hConfig.Config.Account.InitialBalance,
		PeakBalance:      hConfig.Config.Account.InitialBalance,
		TroughBalance:    hConfig.Config.Account.InitialBalance,
		TotalPnL:         0,
		LastUpdateTime:   now,
		SessionStart:     now,
	}

	return state, nil
}

// ==================== STATE QUERIES ====================

// GetCurrentTick returns the current tick (thread-safe)
func (hs *HolodeckState) GetCurrentTick() *types.Tick {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	return hs.CurrentTick
}

// GetPosition returns the current position (thread-safe)
func (hs *HolodeckState) GetPosition() *types.Position {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	return hs.Position
}

// GetBalance returns the current balance (thread-safe)
func (hs *HolodeckState) GetBalance() *types.Balance {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	return hs.Balance
}

// GetExecutionCount returns the total number of executions (thread-safe)
func (hs *HolodeckState) GetExecutionCount() int {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	return hs.ExecutionCount
}

// GetTickCount returns the total number of ticks processed (thread-safe)
func (hs *HolodeckState) GetTickCount() int64 {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	return hs.TickCount
}

// GetTotalPnL returns total P&L (thread-safe)
func (hs *HolodeckState) GetTotalPnL() float64 {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	return hs.TotalPnL
}

// GetErrorCount returns the total number of errors (thread-safe)
func (hs *HolodeckState) GetErrorCount() int {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	return hs.ErrorLog.Size()
}

// ==================== STATE UPDATES ====================

// UpdateTick updates the current tick and increments counter (thread-safe)
func (hs *HolodeckState) UpdateTick(tick *types.Tick) error {
	if tick == nil {
		return types.NewInvalidOperationError("UpdateTick", "tick cannot be nil")
	}

	hs.mu.Lock()
	defer hs.mu.Unlock()

	hs.CurrentTick = tick
	hs.TickCount++
	hs.LastUpdateTime = time.Now()

	return nil
}

// UpdatePosition updates the position (thread-safe)
func (hs *HolodeckState) UpdatePosition(position *types.Position) error {
	if position == nil {
		return types.NewInvalidOperationError("UpdatePosition", "position cannot be nil")
	}

	hs.mu.Lock()
	defer hs.mu.Unlock()

	hs.Position = position
	hs.LastUpdateTime = time.Now()

	return nil
}

// UpdateBalance updates the balance (thread-safe)
func (hs *HolodeckState) UpdateBalance(balance *types.Balance) error {
	if balance == nil {
		return types.NewInvalidOperationError("UpdateBalance", "balance cannot be nil")
	}

	hs.mu.Lock()
	defer hs.mu.Unlock()

	hs.Balance = balance
	hs.CurrentBalance = balance.CurrentBalance

	// Update peak and trough
	if hs.CurrentBalance > hs.PeakBalance {
		hs.PeakBalance = hs.CurrentBalance
	}
	if hs.CurrentBalance < hs.TroughBalance {
		hs.TroughBalance = hs.CurrentBalance
	}

	hs.LastUpdateTime = time.Now()

	return nil
}

// AddExecution adds an execution to the history (thread-safe)
func (hs *HolodeckState) AddExecution(execution *types.ExecutionReport) error {
	if execution == nil {
		return types.NewInvalidOperationError("AddExecution", "execution cannot be nil")
	}

	hs.mu.Lock()
	defer hs.mu.Unlock()

	hs.ExecutionHistory = append(hs.ExecutionHistory, execution)
	hs.ExecutionCount++

	// Update total P&L
	hs.TotalPnL = execution.TotalPnL

	hs.LastUpdateTime = time.Now()

	return nil
}

// AddError adds an error to the log (thread-safe)
func (hs *HolodeckState) AddError(err *types.HolodeckError) {
	if err == nil {
		return
	}

	hs.mu.Lock()
	defer hs.mu.Unlock()

	hs.ErrorLog.Add(err)
	hs.LastUpdateTime = time.Now()
}

// ==================== STATE METRICS ====================

// GetSessionDuration returns how long the session has been running
func (hs *HolodeckState) GetSessionDuration() time.Duration {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	if hs.SessionEnd.IsZero() {
		return time.Since(hs.SessionStart)
	}
	return hs.SessionEnd.Sub(hs.SessionStart)
}

// GetDrawdownPercent returns current drawdown percentage
func (hs *HolodeckState) GetDrawdownPercent() float64 {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	if hs.StartBalance == 0 {
		return 0
	}
	return ((hs.StartBalance - hs.CurrentBalance) / hs.StartBalance) * 100
}

// GetReturnPercent returns total return percentage
func (hs *HolodeckState) GetReturnPercent() float64 {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	if hs.StartBalance == 0 {
		return 0
	}
	return ((hs.CurrentBalance - hs.StartBalance) / hs.StartBalance) * 100
}

// GetMaxDrawdown returns the maximum drawdown experienced
func (hs *HolodeckState) GetMaxDrawdown() float64 {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	if hs.PeakBalance == 0 {
		return 0
	}
	return ((hs.PeakBalance - hs.TroughBalance) / hs.PeakBalance) * 100
}

// GetMetrics returns a comprehensive metrics map
func (hs *HolodeckState) GetMetrics() map[string]interface{} {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	sessionDuration := hs.GetSessionDuration()
	drawdown := hs.GetDrawdownPercent()
	returnPct := hs.GetReturnPercent()
	maxDD := hs.GetMaxDrawdown()

	return map[string]interface{}{
		"session_id":           hs.Config.SessionID,
		"instrument":           hs.Config.Instrument.GetSymbol(),
		"tick_count":           hs.TickCount,
		"execution_count":      hs.ExecutionCount,
		"error_count":          hs.ErrorLog.Size(),
		"session_duration":     sessionDuration,
		"start_balance":        hs.StartBalance,
		"current_balance":      hs.CurrentBalance,
		"peak_balance":         hs.PeakBalance,
		"trough_balance":       hs.TroughBalance,
		"total_pnl":            hs.TotalPnL,
		"return_percent":       returnPct,
		"drawdown_percent":     drawdown,
		"max_drawdown_percent": maxDD,
		"last_update_time":     hs.LastUpdateTime,
		"is_running":           hs.Config.IsRunning,
	}
}

// ==================== HOLODECK STATUS ====================

// SessionStatus represents the current status of a Holodeck session
type SessionStatus struct {
	SessionID        string
	InstrumentType   string
	InstrumentSymbol string
	StartTime        time.Time
	CurrentTime      time.Time
	IsRunning        bool
	TicksProcessed   int64
	ExecutionsCount  int
	ErrorsCount      int
	CurrentBalance   float64
	StartBalance     float64
	TotalPnL         float64
	DrawdownPercent  float64
	ReturnPercent    float64
	AccountStatus    string
}

// GetStatus returns the current session status
func (hs *HolodeckState) GetStatus() *SessionStatus {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	return &SessionStatus{
		SessionID:        hs.Config.SessionID,
		InstrumentType:   hs.Config.Instrument.GetType(),
		InstrumentSymbol: hs.Config.Instrument.GetSymbol(),
		StartTime:        hs.SessionStart,
		CurrentTime:      hs.LastUpdateTime,
		IsRunning:        hs.Config.IsRunning,
		TicksProcessed:   hs.TickCount,
		ExecutionsCount:  hs.ExecutionCount,
		ErrorsCount:      hs.ErrorLog.Size(),
		CurrentBalance:   hs.CurrentBalance,
		StartBalance:     hs.StartBalance,
		TotalPnL:         hs.TotalPnL,
		DrawdownPercent:  hs.GetDrawdownPercent(),
		ReturnPercent:    hs.GetReturnPercent(),
		AccountStatus:    hs.Balance.AccountStatus,
	}
}

// String returns a human-readable status string
func (ss *SessionStatus) String() string {
	return fmt.Sprintf(
		"SessionStatus[%s | %s/%s | Ticks:%d Execs:%d Errors:%d | PnL:%.2f (%.2f%%) | DD:%.2f%%]",
		ss.SessionID,
		ss.InstrumentType,
		ss.InstrumentSymbol,
		ss.TicksProcessed,
		ss.ExecutionsCount,
		ss.ErrorsCount,
		ss.TotalPnL,
		ss.ReturnPercent,
		ss.DrawdownPercent,
	)
}

// DebugString returns detailed status information
func (ss *SessionStatus) DebugString() string {
	return fmt.Sprintf(
		"Session Status:\n"+
			"  Session ID:        %s\n"+
			"  Instrument:        %s (%s)\n"+
			"  Start Time:        %s\n"+
			"  Current Time:      %s\n"+
			"  Is Running:        %v\n"+
			"\n"+
			"  Processing:\n"+
			"    Ticks:           %d\n"+
			"    Executions:      %d\n"+
			"    Errors:          %d\n"+
			"\n"+
			"  Balance:\n"+
			"    Start:           %.2f\n"+
			"    Current:         %.2f\n"+
			"    Peak:            N/A (see metrics)\n"+
			"    Total P&L:       %.2f\n"+
			"\n"+
			"  Performance:\n"+
			"    Return:          %.2f%%\n"+
			"    Drawdown:        %.2f%%\n"+
			"    Account Status:  %s",
		ss.SessionID,
		ss.InstrumentType, ss.InstrumentSymbol,
		ss.StartTime.Format("2006-01-02T15:04:05.000"),
		ss.CurrentTime.Format("2006-01-02T15:04:05.000"),
		ss.IsRunning,
		ss.TicksProcessed,
		ss.ExecutionsCount,
		ss.ErrorsCount,
		ss.StartBalance,
		ss.CurrentBalance,
		ss.TotalPnL,
		ss.ReturnPercent,
		ss.DrawdownPercent,
		ss.AccountStatus,
	)
}

// ==================== SESSION ID GENERATION ====================

// generateSessionID creates a unique session ID
func generateSessionID() string {
	return fmt.Sprintf("HOLO-%d", time.Now().UnixNano())
}

// ==================== VALIDATION HELPERS ====================

// ValidateHolodeckConfig validates a Holodeck configuration
func ValidateHolodeckConfig(hConfig *HolodeckConfig) error {
	if hConfig == nil {
		return types.NewConfigError("hConfig", "Holodeck configuration cannot be nil")
	}

	if hConfig.Config == nil {
		return types.NewConfigError("hConfig.config", "underlying config cannot be nil")
	}

	if hConfig.Instrument == nil {
		return types.NewConfigError("hConfig.instrument", "instrument cannot be nil")
	}

	if hConfig.SessionID == "" {
		return types.NewConfigError("hConfig.sessionID", "session ID cannot be empty")
	}

	return nil
}

// ValidateHolodeckState validates a Holodeck state
func ValidateHolodeckState(state *HolodeckState) error {
	if state == nil {
		return types.NewConfigError("state", "Holodeck state cannot be nil")
	}

	if state.Config == nil {
		return types.NewConfigError("state.config", "configuration cannot be nil")
	}

	if state.Position == nil {
		return types.NewConfigError("state.position", "position cannot be nil")
	}

	if state.Balance == nil {
		return types.NewConfigError("state.balance", "balance cannot be nil")
	}

	if state.ErrorLog == nil {
		return types.NewConfigError("state.errorLog", "error log cannot be nil")
	}

	return nil
}

// ==================== CONFIGURATION RESET ====================

// Reset resets the state to initial conditions
func (hs *HolodeckState) Reset() error {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	// Reset position
	hs.Position = types.NewPosition()

	// Reset balance
	hs.Balance = types.NewBalance(
		hs.Config.Config.Account.InitialBalance,
		hs.Config.Config.Account.Currency,
		hs.Config.Config.Account.Leverage,
		hs.Config.Config.Account.MaxDrawdownPercent,
		hs.Config.Config.Account.MaxPositionSize,
	)

	// Reset tracking
	hs.CurrentTick = nil
	hs.TickCount = 0
	hs.ExecutionCount = 0
	hs.ExecutionHistory = make([]*types.ExecutionReport, 0, hs.Config.StateConfig.MaxExecutionHistorySize)
	hs.ErrorLog = types.NewErrorLog()

	// Reset metrics
	hs.StartBalance = hs.Config.Config.Account.InitialBalance
	hs.CurrentBalance = hs.Config.Config.Account.InitialBalance
	hs.PeakBalance = hs.Config.Config.Account.InitialBalance
	hs.TroughBalance = hs.Config.Config.Account.InitialBalance
	hs.TotalPnL = 0

	// Reset timing
	now := time.Now()
	hs.LastUpdateTime = now
	hs.SessionStart = now
	hs.SessionEnd = time.Time{}

	return nil
}

// Snapshot creates a snapshot of the current state for storage
func (hs *HolodeckState) Snapshot() map[string]interface{} {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	return map[string]interface{}{
		"session_id":      hs.Config.SessionID,
		"timestamp":       hs.LastUpdateTime,
		"tick_count":      hs.TickCount,
		"execution_count": hs.ExecutionCount,
		"position":        hs.Position,
		"balance":         hs.Balance,
		"total_pnl":       hs.TotalPnL,
		"current_balance": hs.CurrentBalance,
		"error_count":     hs.ErrorLog.Size(),
	}
}
