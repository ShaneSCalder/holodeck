package simulator

import (
	"fmt"
	"sync"
	"time"

	"holodeck/types"
)

// ==================== HOLODECK MAIN API ====================

// Holodeck is the main orchestrator for the mock broker system
// It manages state, executes orders, and coordinates all subsystems
type Holodeck struct {
	// Configuration
	config *HolodeckConfig
	state  *HolodeckState

	// Subsystems
	executor OrderExecutor
	reader   TickReader
	logger   Logger

	// Synchronization
	mu       sync.RWMutex
	running  bool
	stopped  bool
	stopChan chan bool

	// Callbacks (for integration with agents)
	callbacks HolodeckCallbacks

	// Performance tracking
	startTime      time.Time
	lastTickTime   time.Time
	ticksPerSecond float64
}

// ==================== SUBSYSTEM INTERFACES ====================

// OrderExecutor defines the order execution interface
type OrderExecutor interface {
	// Execute executes an order and returns an execution report
	Execute(order *types.Order, tick *types.Tick, instrument types.Instrument) (*types.ExecutionReport, error)

	// Validate validates an order before execution
	Validate(order *types.Order, instrument types.Instrument, availableBalance float64) error

	// CalculateCommission calculates commission for an order
	CalculateCommission(price, size float64, instrument types.Instrument, side string) float64

	// CalculateSlippage calculates slippage for an order
	CalculateSlippage(size float64, availableDepth int64, momentum int, instrument types.Instrument) float64
}

// TickReader defines the tick data source interface
type TickReader interface {
	// HasNext checks if there are more ticks to read
	HasNext() bool

	// Next returns the next tick
	Next() (*types.Tick, error)

	// Close closes the reader
	Close() error

	// GetTickCount returns the number of ticks read
	GetTickCount() int64

	// Reset resets the reader to the beginning
	Reset() error
}

// Logger defines the logging interface
type Logger interface {
	// LogTick logs a tick
	LogTick(tick *types.Tick)

	// LogOrder logs an order
	LogOrder(order *types.Order)

	// LogExecution logs an execution report
	LogExecution(exec *types.ExecutionReport)

	// LogError logs an error
	LogError(err error)

	// LogMetrics logs performance metrics
	LogMetrics(metrics map[string]interface{})

	// Close closes the logger
	Close() error
}

// HolodeckCallbacks are optional callbacks for integration
type HolodeckCallbacks struct {
	// OnTick is called when a new tick is received
	OnTick func(tick *types.Tick) error

	// OnExecution is called after an order is executed
	OnExecution func(exec *types.ExecutionReport) error

	// OnError is called when an error occurs
	OnError func(err error)

	// OnStatusChange is called when account status changes
	OnStatusChange func(oldStatus, newStatus string)

	// OnSessionEnd is called when the session ends
	OnSessionEnd func(status *SessionStatus)
}

// ==================== HOLODECK CREATION ====================

// NewHolodeck creates a new Holodeck instance
func NewHolodeck(config *HolodeckConfig) (*Holodeck, error) {
	// Validate configuration
	if err := ValidateHolodeckConfig(config); err != nil {
		return nil, err
	}

	// Create state
	state, err := NewHolodeckState(config)
	if err != nil {
		return nil, err
	}

	// Create Holodeck instance
	h := &Holodeck{
		config:    config,
		state:     state,
		running:   false,
		stopped:   false,
		stopChan:  make(chan bool, 1),
		startTime: time.Now(),
	}

	return h, nil
}

// WithExecutor sets the order executor
func (h *Holodeck) WithExecutor(executor OrderExecutor) *Holodeck {
	h.executor = executor
	return h
}

// WithReader sets the tick reader
func (h *Holodeck) WithReader(reader TickReader) *Holodeck {
	h.reader = reader
	return h
}

// WithLogger sets the logger
func (h *Holodeck) WithLogger(logger Logger) *Holodeck {
	h.logger = logger
	return h
}

// WithCallbacks sets the callbacks
func (h *Holodeck) WithCallbacks(callbacks HolodeckCallbacks) *Holodeck {
	h.callbacks = callbacks
	return h
}

// ==================== HOLODECK VALIDATION ====================

// Validate validates that Holodeck is properly configured
func (h *Holodeck) Validate() error {
	if h.executor == nil {
		return types.NewInvalidOperationError("Validate", "executor not set")
	}

	if h.reader == nil {
		return types.NewInvalidOperationError("Validate", "reader not set")
	}

	if h.state == nil {
		return types.NewInvalidOperationError("Validate", "state not initialized")
	}

	return nil
}

// ==================== SESSION LIFECYCLE ====================

// Start starts the Holodeck session
func (h *Holodeck) Start() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.running {
		return types.NewInvalidOperationError("Start", "already running")
	}

	// Validate setup
	if err := h.Validate(); err != nil {
		return err
	}

	h.running = true
	h.stopped = false
	h.config.IsRunning = true
	h.startTime = time.Now()
	h.state.SessionStart = h.startTime

	if h.logger != nil {
		metrics := map[string]interface{}{
			"event":      "session_start",
			"session_id": h.config.SessionID,
			"instrument": h.config.Instrument.GetSymbol(),
			"timestamp":  h.startTime,
		}
		h.logger.LogMetrics(metrics)
	}

	return nil
}

// Stop stops the Holodeck session
func (h *Holodeck) Stop() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.running {
		return types.NewInvalidOperationError("Stop", "not running")
	}

	h.running = false
	h.stopped = true
	h.config.IsRunning = false
	h.state.SessionEnd = time.Now()
	h.stopChan <- true

	if h.logger != nil {
		metrics := map[string]interface{}{
			"event":      "session_stop",
			"session_id": h.config.SessionID,
			"timestamp":  h.state.SessionEnd,
		}
		h.logger.LogMetrics(metrics)
	}

	// Call session end callback
	if h.callbacks.OnSessionEnd != nil {
		status := h.state.GetStatus()
		h.callbacks.OnSessionEnd(status)
	}

	return nil
}

// IsRunning checks if Holodeck is running
func (h *Holodeck) IsRunning() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.running
}

// ==================== TICK PROCESSING ====================

// ProcessTick processes a single tick
func (h *Holodeck) ProcessTick(tick *types.Tick) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.running {
		return types.NewInvalidOperationError("ProcessTick", "not running")
	}

	// Validate tick
	if !tick.IsValid() {
		err := types.NewConfigError("tick", "invalid tick data")
		h.state.AddError(err)
		if h.callbacks.OnError != nil {
			h.callbacks.OnError(err)
		}
		return err
	}

	// Update state with tick
	if err := h.state.UpdateTick(tick); err != nil {
		// Type assert the error to HolodeckError if possible
		if hErr, ok := err.(*types.HolodeckError); ok {
			h.state.AddError(hErr)
		}
		if h.callbacks.OnError != nil {
			h.callbacks.OnError(err)
		}
		return err
	}

	// Update position with new price
	if !h.state.Position.IsFlat() {
		h.state.Position.UpdatePrice(tick.GetMidPrice(), h.config.Instrument.GetPipValue())
		h.state.UpdateBalance(h.state.Balance)
	}

	// Call tick callback
	if h.callbacks.OnTick != nil {
		if err := h.callbacks.OnTick(tick); err != nil {
			h.state.AddError(types.NewOrderRejectedError(fmt.Sprintf("tick callback error: %v", err)))
			if h.callbacks.OnError != nil {
				h.callbacks.OnError(err)
			}
		}
	}

	// Log tick
	if h.logger != nil {
		h.logger.LogTick(tick)
	}

	h.lastTickTime = time.Now()

	return nil
}

// ProcessTickStream processes ticks from the reader in a loop
func (h *Holodeck) ProcessTickStream() error {
	// Start session
	if err := h.Start(); err != nil {
		return err
	}

	defer func() {
		if h.IsRunning() {
			h.Stop()
		}
		if h.reader != nil {
			h.reader.Close()
		}
		if h.logger != nil {
			h.logger.Close()
		}
	}()

	// Process ticks
	for h.IsRunning() && h.reader.HasNext() {
		tick, err := h.reader.Next()
		if err != nil {
			herr := types.NewConfigError("tick_reader", fmt.Sprintf("error reading tick: %v", err))
			h.state.AddError(herr)
			if h.callbacks.OnError != nil {
				h.callbacks.OnError(herr)
			}
			continue
		}

		if err := h.ProcessTick(tick); err != nil {
			// Log but continue
			continue
		}

		// Check for account blown
		if h.state.Balance.IsAccountBlown() {
			err := types.NewAccountBlownError(
				h.state.GetDrawdownPercent(),
				h.config.Config.Account.MaxDrawdownPercent,
			)
			h.state.AddError(err)

			if h.callbacks.OnStatusChange != nil {
				h.callbacks.OnStatusChange(types.AccountStatusActive, types.AccountStatusBlown)
			}
			if h.callbacks.OnError != nil {
				h.callbacks.OnError(err)
			}

			h.Stop()
			break
		}
	}

	return nil
}

// ==================== ORDER EXECUTION ====================

// ExecuteOrder executes an order
func (h *Holodeck) ExecuteOrder(order *types.Order) (*types.ExecutionReport, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.running {
		return nil, types.NewInvalidOperationError("ExecuteOrder", "not running")
	}

	if order == nil {
		return nil, types.NewOrderRejectedError("order cannot be nil")
	}

	// Get current tick
	tick := h.state.GetCurrentTick()
	if tick == nil {
		return nil, types.NewOrderRejectedError("no tick data available")
	}

	// Handle HOLD orders
	if order.IsHold() {
		return &types.ExecutionReport{
			OrderID:       order.OrderID,
			Timestamp:     tick.Timestamp,
			Action:        types.OrderActionHold,
			Status:        types.OrderStatusFilled,
			PositionAfter: h.state.Position.Size,
		}, nil
	}

	// Validate order
	if err := h.executor.Validate(order, h.config.Instrument, h.state.Balance.BuyingPower); err != nil {
		rejection := types.NewOrderRejectedError(fmt.Sprintf("validation failed: %v", err))
		h.state.AddError(rejection)
		if h.callbacks.OnError != nil {
			h.callbacks.OnError(rejection)
		}

		// Return rejection
		herr, _ := err.(*types.HolodeckError)
		return types.NewRejectedExecution(
			order.OrderID,
			tick.Timestamp,
			order.Action,
			order.Size,
			herr.Code,
			herr.Message,
		), nil
	}

	// Log order
	if h.logger != nil {
		h.logger.LogOrder(order)
	}

	// Execute order
	exec, err := h.executor.Execute(order, tick, h.config.Instrument)
	if err != nil {
		rejection := types.NewOrderRejectedError(fmt.Sprintf("execution failed: %v", err))
		h.state.AddError(rejection)
		if h.callbacks.OnError != nil {
			h.callbacks.OnError(rejection)
		}
		return nil, err
	}

	// Check for rejection
	if exec.IsRejected() {
		h.state.AddError(types.NewOrderRejectedError(exec.ErrorMessage))
		if h.callbacks.OnError != nil {
			h.callbacks.OnError(types.NewOrderRejectedError(exec.ErrorMessage))
		}
		if h.logger != nil {
			h.logger.LogExecution(exec)
		}
		return exec, nil
	}

	// Update position
	if exec.WasExecuted() {
		// Create trade record
		trade := &types.Trade{
			TradeID:    fmt.Sprintf("TRADE-%d", h.state.GetExecutionCount()),
			Timestamp:  tick.Timestamp,
			Action:     order.Action,
			Size:       exec.FilledSize,
			Price:      exec.FillPrice,
			Commission: exec.Commission,
			Slippage:   exec.SlippageUnits,
			IsEntry:    h.state.Position.IsFlat() && exec.WasExecuted(),
			IsExit:     !h.state.Position.IsFlat() && order.Action == opposite(h.state.Position.GetStatus()),
			PnLAtClose: exec.RealizedPnL,
		}

		h.state.Position.AddTrade(trade)
		h.state.Position.Size = exec.PositionAfter
		h.state.UpdatePosition(h.state.Position)
	}

	// Update balance
	h.state.Balance.UpdateFromExecution(exec)
	h.state.UpdateBalance(h.state.Balance)

	// Record execution
	h.state.AddExecution(exec)

	// Log execution
	if h.logger != nil {
		h.logger.LogExecution(exec)
	}

	// Call execution callback
	if h.callbacks.OnExecution != nil {
		if err := h.callbacks.OnExecution(exec); err != nil {
			h.state.AddError(types.NewOrderRejectedError(fmt.Sprintf("execution callback error: %v", err)))
			if h.callbacks.OnError != nil {
				h.callbacks.OnError(err)
			}
		}
	}

	return exec, nil
}

// ==================== STATE QUERIES ====================

// GetState returns the current Holodeck state
func (h *Holodeck) GetState() *HolodeckState {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.state
}

// GetStatus returns the current session status
func (h *Holodeck) GetStatus() *SessionStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.state.GetStatus()
}

// GetMetrics returns comprehensive metrics
func (h *Holodeck) GetMetrics() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	metrics := h.state.GetMetrics()

	// Add Holodeck-specific metrics
	metrics["executor"] = "active"
	metrics["reader"] = h.reader
	metrics["logger"] = h.logger
	// Check if any callbacks are set (since we can't compare function fields directly)
	hasCallbacks := h.callbacks.OnTick != nil || h.callbacks.OnError != nil ||
		h.callbacks.OnStatusChange != nil || h.callbacks.OnSessionEnd != nil
	metrics["callbacks_set"] = hasCallbacks

	return metrics
}

// GetConfig returns the Holodeck configuration
func (h *Holodeck) GetConfig() *HolodeckConfig {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.config
}

// GetPosition returns the current position
func (h *Holodeck) GetPosition() *types.Position {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.state.Position
}

// GetBalance returns the current balance
func (h *Holodeck) GetBalance() *types.Balance {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.state.Balance
}

// GetExecutionHistory returns the execution history
func (h *Holodeck) GetExecutionHistory() []*types.ExecutionReport {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.state.ExecutionHistory
}

// GetErrors returns all recorded errors
func (h *Holodeck) GetErrors() []*types.HolodeckError {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.state.ErrorLog.Errors
}

// ==================== CONTROL & DIAGNOSTICS ====================

// Reset resets the Holodeck to initial state
func (h *Holodeck) Reset() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.running {
		return types.NewInvalidOperationError("Reset", "cannot reset while running")
	}

	return h.state.Reset()
}

// PrintStatus prints the current status to console
func (h *Holodeck) PrintStatus() {
	status := h.GetStatus()
	fmt.Println(status.DebugString())
}

// PrintMetrics prints all metrics to console
func (h *Holodeck) PrintMetrics() {
	metrics := h.GetMetrics()
	fmt.Println("=== Holodeck Metrics ===")
	for key, value := range metrics {
		fmt.Printf("  %s: %v\n", key, value)
	}
}

// PrintErrors prints all errors to console
func (h *Holodeck) PrintErrors() {
	errors := h.GetErrors()
	if len(errors) == 0 {
		fmt.Println("No errors recorded")
		return
	}

	fmt.Printf("=== Errors (%d total) ===\n", len(errors))
	for i, err := range errors {
		fmt.Printf("%d. %s\n", i+1, err.Error())
	}
}

// ==================== PERFORMANCE MONITORING ====================

// CalculateTicksPerSecond calculates the processing speed
func (h *Holodeck) CalculateTicksPerSecond() float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.lastTickTime.IsZero() || h.startTime.IsZero() {
		return 0
	}

	elapsed := h.lastTickTime.Sub(h.startTime).Seconds()
	if elapsed == 0 {
		return 0
	}

	return float64(h.state.TickCount) / elapsed
}

// GetPerformanceSummary returns a performance summary
func (h *Holodeck) GetPerformanceSummary() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	sessionDuration := h.state.GetSessionDuration()
	ticksPerSecond := h.CalculateTicksPerSecond()

	return map[string]interface{}{
		"session_id":       h.config.SessionID,
		"session_duration": sessionDuration.String(),
		"ticks_processed":  h.state.TickCount,
		"executions":       h.state.ExecutionCount,
		"ticks_per_second": fmt.Sprintf("%.2f", ticksPerSecond),
		"errors":           h.state.ErrorLog.Size(),
		"start_balance":    h.state.StartBalance,
		"current_balance":  h.state.CurrentBalance,
		"total_pnl":        h.state.TotalPnL,
		"return_percent":   fmt.Sprintf("%.2f%%", h.state.GetReturnPercent()),
		"drawdown_percent": fmt.Sprintf("%.2f%%", h.state.GetDrawdownPercent()),
		"account_status":   h.state.Balance.AccountStatus,
		"is_running":       h.running,
	}
}

// ==================== UTILITY FUNCTIONS ====================

// opposite returns the opposite position action
func opposite(status string) string {
	switch status {
	case types.PositionStatusLong:
		return types.OrderActionSell
	case types.PositionStatusShort:
		return types.OrderActionBuy
	default:
		return ""
	}
}

// ==================== HOLODECK BUILDER PATTERN ====================

// HolodeckBuilder allows fluent construction of Holodeck
type HolodeckBuilder struct {
	holodeck *Holodeck
	err      error
}

// NewHolodeckBuilder creates a new Holodeck builder
func NewHolodeckBuilder(config *HolodeckConfig) *HolodeckBuilder {
	h, err := NewHolodeck(config)
	return &HolodeckBuilder{
		holodeck: h,
		err:      err,
	}
}

// WithExecutor adds an executor
func (hb *HolodeckBuilder) WithExecutor(executor OrderExecutor) *HolodeckBuilder {
	if hb.err != nil {
		return hb
	}
	hb.holodeck.WithExecutor(executor)
	return hb
}

// WithReader adds a reader
func (hb *HolodeckBuilder) WithReader(reader TickReader) *HolodeckBuilder {
	if hb.err != nil {
		return hb
	}
	hb.holodeck.WithReader(reader)
	return hb
}

// WithLogger adds a logger
func (hb *HolodeckBuilder) WithLogger(logger Logger) *HolodeckBuilder {
	if hb.err != nil {
		return hb
	}
	hb.holodeck.WithLogger(logger)
	return hb
}

// WithCallbacks adds callbacks
func (hb *HolodeckBuilder) WithCallbacks(callbacks HolodeckCallbacks) *HolodeckBuilder {
	if hb.err != nil {
		return hb
	}
	hb.holodeck.WithCallbacks(callbacks)
	return hb
}

// Build returns the constructed Holodeck or error
func (hb *HolodeckBuilder) Build() (*Holodeck, error) {
	if hb.err != nil {
		return nil, hb.err
	}

	if err := hb.holodeck.Validate(); err != nil {
		return nil, err
	}

	return hb.holodeck, nil
}

// MustBuild builds and panics on error
func (hb *HolodeckBuilder) MustBuild() *Holodeck {
	h, err := hb.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build Holodeck: %v", err))
	}
	return h
}
