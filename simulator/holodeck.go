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
	startTime    time.Time
	lastTickTime time.Time
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

// ==================== PUBLIC API METHODS ====================
// These are the 11 core methods that agents/strategies use

// GetNextTick returns the next market tick from the data source
// Returns types.Tick and error if no more ticks or read error
func (h *Holodeck) GetNextTick() (*types.Tick, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if !h.running {
		return nil, fmt.Errorf("holodeck not running")
	}

	if h.reader == nil {
		return nil, fmt.Errorf("reader not set")
	}

	// Check if there are more ticks
	if !h.reader.HasNext() {
		return nil, fmt.Errorf("no more ticks available")
	}

	// Get next tick
	tick, err := h.reader.Next()
	if err != nil {
		if h.logger != nil {
			h.logger.LogError(err)
		}
		return nil, err
	}

	// Update state - use actual field name: CurrentTick
	h.state.CurrentTick = tick
	h.state.TickCount++
	h.lastTickTime = time.Now()

	// Log tick if logger available
	if h.logger != nil {
		h.logger.LogTick(tick)
	}

	// Call callback if set
	if h.callbacks.OnTick != nil {
		if err := h.callbacks.OnTick(tick); err != nil {
			if h.logger != nil {
				h.logger.LogError(err)
			}
		}
	}

	return tick, nil
}

// ExecuteOrder executes a buy/sell order and returns execution report
// Applies realistic friction: commission, slippage, partial fills
func (h *Holodeck) ExecuteOrder(order *types.Order) (*types.ExecutionReport, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.running {
		return nil, fmt.Errorf("holodeck not running")
	}

	if h.executor == nil {
		return nil, fmt.Errorf("executor not set")
	}

	if h.state.CurrentTick == nil {
		return nil, fmt.Errorf("no tick data available")
	}

	// Execute the order
	exec, err := h.executor.Execute(order, h.state.CurrentTick, h.config.Instrument)
	if err != nil {
		// Log error
		if h.logger != nil {
			h.logger.LogError(err)
		}
		// Call error callback
		if h.callbacks.OnError != nil {
			h.callbacks.OnError(err)
		}
		return nil, err
	}

	// Update state if executed (not rejected)
	if !exec.IsRejected() && exec.FilledSize > 0 {
		// Use correct field name: Position (it's *types.Position)
		if h.state.Position != nil {
			h.state.Position.Size = exec.FilledSize
			h.state.Position.EntryPrice = exec.FillPrice
		}

		// Use correct field name: ExecutionHistory
		h.state.ExecutionHistory = append(h.state.ExecutionHistory, exec)
		h.state.ExecutionCount++

		// Update balance - use correct field name: CurrentBalance (not Current)
		if h.state.Balance != nil {
			h.state.Balance.UpdateFromExecution(exec)
		}
	}

	// Log execution
	if h.logger != nil {
		h.logger.LogExecution(exec)
	}

	// Call execution callback
	if h.callbacks.OnExecution != nil {
		if err := h.callbacks.OnExecution(exec); err != nil {
			if h.logger != nil {
				h.logger.LogError(err)
			}
		}
	}

	return exec, nil
}

// GetPosition returns the current position state
// Returns position size, entry price, unrealized P&L
func (h *Holodeck) GetPosition() *types.Position {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.state == nil || h.state.Position == nil {
		return &types.Position{}
	}

	// Return copy of position state
	position := &types.Position{
		Size:          h.state.Position.Size,
		EntryPrice:    h.state.Position.EntryPrice,
		UnrealizedPnL: h.state.Position.UnrealizedPnL,
		RealizedPnL:   h.state.Position.RealizedPnL,
	}

	return position
}

// GetBalance returns the current account balance state
// Returns balance, initial balance, drawdown info
func (h *Holodeck) GetBalance() *types.Balance {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.state == nil || h.state.Balance == nil {
		return &types.Balance{}
	}

	// Return copy of balance state using correct field names
	return &types.Balance{
		InitialBalance:     h.state.Balance.InitialBalance,
		CurrentBalance:     h.state.Balance.CurrentBalance,
		Currency:           h.state.Balance.Currency,
		TotalRealizedPnL:   h.state.Balance.TotalRealizedPnL,
		TotalUnrealizedPnL: h.state.Balance.TotalUnrealizedPnL,
		CommissionPaid:     h.state.Balance.CommissionPaid,
		Leverage:           h.state.Balance.Leverage,
		UsedMargin:         h.state.Balance.UsedMargin,
		AvailableMargin:    h.state.Balance.AvailableMargin,
		BuyingPower:        h.state.Balance.BuyingPower,
		AccountStatus:      h.state.Balance.AccountStatus,
	}
}

// GetMetrics returns current performance metrics as a map
// Includes: ticks processed, trades executed, balance, position info
func (h *Holodeck) GetMetrics() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	metrics := make(map[string]interface{})

	if h.state == nil {
		return metrics
	}

	// Basic metrics
	metrics["ticks_processed"] = h.state.TickCount
	metrics["trades_executed"] = h.state.ExecutionCount
	metrics["session_duration"] = time.Since(h.startTime)

	if h.state.Balance != nil {
		metrics["current_balance"] = h.state.Balance.CurrentBalance
		metrics["initial_balance"] = h.state.Balance.InitialBalance
		metrics["available_margin"] = h.state.Balance.AvailableMargin
		metrics["buying_power"] = h.state.Balance.BuyingPower
		metrics["commission_paid"] = h.state.Balance.CommissionPaid
		metrics["return_percent"] = h.state.Balance.GetReturnPercent()
		metrics["drawdown_percent"] = h.state.Balance.GetDrawdownPercent()
		metrics["win_rate"] = h.state.Balance.GetWinRate()
	}

	if h.state.Position != nil {
		metrics["position_size"] = h.state.Position.Size
		metrics["entry_price"] = h.state.Position.EntryPrice
		metrics["unrealized_pnl"] = h.state.Position.UnrealizedPnL
	}

	if h.reader != nil {
		metrics["total_ticks_available"] = h.reader.GetTickCount()
	}

	return metrics
}

// SetSpeed sets the simulation speed multiplier
// Speed 1.0 = real-time, 100.0 = 100x faster, etc.
func (h *Holodeck) SetSpeed(multiplier float64) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if multiplier <= 0 {
		return fmt.Errorf("speed multiplier must be positive")
	}

	// Store in ExecutionConfig
	h.config.ExecutionConfig.SpeedMultiplier = multiplier
	return nil
}

// Reset resets the Holodeck to initial state
// Clears trades, resets balance, closes position
func (h *Holodeck) Reset() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.running {
		return fmt.Errorf("cannot reset while running")
	}

	// Reset state by recreating it
	state, err := NewHolodeckState(h.config)
	if err != nil {
		return err
	}
	h.state = state

	// Reset reader if possible
	if h.reader != nil {
		if err := h.reader.Reset(); err != nil {
			return err
		}
	}

	return nil
}

// Start starts the Holodeck session
// Must be called before GetNextTick or ExecuteOrder
func (h *Holodeck) Start() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.running {
		return fmt.Errorf("already running")
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
		return fmt.Errorf("not running")
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

// IsRunning returns whether the Holodeck session is currently running
func (h *Holodeck) IsRunning() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.running
}

// IsAccountBlown returns whether the account has been blown
// Returns true if balance is nil or blown
func (h *Holodeck) IsAccountBlown() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.state == nil || h.state.Balance == nil {
		return false
	}

	return h.state.Balance.IsAccountBlown()
}

// ==================== HOLODECK VALIDATION ====================

// Validate validates that Holodeck is properly configured
func (h *Holodeck) Validate() error {
	if h.executor == nil {
		return fmt.Errorf("executor not set")
	}

	if h.reader == nil {
		return fmt.Errorf("reader not set")
	}

	if h.state == nil {
		return fmt.Errorf("state not initialized")
	}

	return nil
}

// ==================== BUILDER PATTERN ====================

// HolodeckBuilder provides a fluent interface for building Holodeck instances
type HolodeckBuilder struct {
	holodeck *Holodeck
	err      error
}

// NewBuilder creates a new HolodeckBuilder
func NewBuilder(config *HolodeckConfig) *HolodeckBuilder {
	holodeck, err := NewHolodeck(config)
	return &HolodeckBuilder{
		holodeck: holodeck,
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
