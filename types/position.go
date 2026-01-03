package types

import (
	"fmt"
	"time"
)

// ==================== POSITION STRUCTURE ====================

// Position represents the current open trading position
// Tracks a single position (LONG, SHORT, or FLAT)
type Position struct {
	// Size is the current position size
	// Positive = LONG (own the asset)
	// Negative = SHORT (owe the asset)
	// 0 = FLAT (no position)
	Size float64

	// EntryPrice is the average entry price for the current position
	EntryPrice float64

	// EntryTime is when the position was opened
	EntryTime time.Time

	// EntryCommission is the commission paid when opening the position
	EntryCommission float64

	// CurrentPrice is the latest market price (updated each tick)
	CurrentPrice float64

	// RealizedPnL is profit/loss from closed trades
	RealizedPnL float64

	// UnrealizedPnL is mark-to-market P&L on open position
	UnrealizedPnL float64

	// CommissionPaid is total commission on this position
	CommissionPaid float64

	// TradeCount is the number of trades that make up this position
	TradeCount int

	// TradeHistory tracks all trades that affect this position
	TradeHistory []*Trade

	// PeakProfit is the highest unrealized P&L reached
	PeakProfit float64

	// PeakLoss is the lowest unrealized P&L reached
	PeakLoss float64

	// MaxAdverseExcursion is the worst mark-to-market during position
	MaxAdverseExcursion float64

	// MaxFavorableExcursion is the best mark-to-market during position
	MaxFavorableExcursion float64
}

// ==================== TRADE RECORD ====================

// Trade represents a single trade (entry or partial exit)
type Trade struct {
	// TradeID is a unique identifier for this trade
	TradeID string

	// Timestamp is when the trade occurred
	Timestamp time.Time

	// Action is BUY or SELL
	Action string

	// Size is the quantity traded
	Size float64

	// Price is the execution price
	Price float64

	// Commission paid for this trade
	Commission float64

	// Slippage on this trade
	Slippage float64

	// IsEntry indicates if this opened the position (true) or modified it (false)
	IsEntry bool

	// IsExit indicates if this closed or reduced the position
	IsExit bool

	// PnLAtClose is the P&L if this was a close
	PnLAtClose float64
}

// ==================== POSITION CONSTRUCTORS ====================

// NewPosition creates an empty flat position
func NewPosition() *Position {
	return &Position{
		Size:           0,
		RealizedPnL:    0,
		UnrealizedPnL:  0,
		CommissionPaid: 0,
		TradeCount:     0,
		TradeHistory:   make([]*Trade, 0),
		PeakProfit:     0,
		PeakLoss:       0,
	}
}

// NewLongPosition creates a position that is long
func NewLongPosition(size, entryPrice float64, entryTime time.Time, commission float64) *Position {
	pos := NewPosition()
	pos.Size = size
	pos.EntryPrice = entryPrice
	pos.EntryTime = entryTime
	pos.EntryCommission = commission
	pos.CommissionPaid = commission
	pos.TradeCount = 1
	return pos
}

// NewShortPosition creates a position that is short
func NewShortPosition(size, entryPrice float64, entryTime time.Time, commission float64) *Position {
	pos := NewPosition()
	pos.Size = -size // Negative for short
	pos.EntryPrice = entryPrice
	pos.EntryTime = entryTime
	pos.EntryCommission = commission
	pos.CommissionPaid = commission
	pos.TradeCount = 1
	return pos
}

// ==================== POSITION QUERIES ====================

// GetStatus returns the position status: LONG, SHORT, or FLAT
func (p *Position) GetStatus() string {
	return GetPositionStatusFromSize(p.Size)
}

// IsFlat returns true if position size is 0
func (p *Position) IsFlat() bool {
	return p.Size == 0
}

// IsLong returns true if position size > 0
func (p *Position) IsLong() bool {
	return p.Size > 0
}

// IsShort returns true if position size < 0
func (p *Position) IsShort() bool {
	return p.Size < 0
}

// GetAbsoluteSize returns the absolute value of position size
func (p *Position) GetAbsoluteSize() float64 {
	if p.Size < 0 {
		return -p.Size
	}
	return p.Size
}

// GetDirection returns 1 for LONG, -1 for SHORT, 0 for FLAT
func (p *Position) GetDirection() int {
	if p.IsLong() {
		return 1
	} else if p.IsShort() {
		return -1
	}
	return 0
}

// GetOpenDuration returns how long the position has been open
func (p *Position) GetOpenDuration(now time.Time) time.Duration {
	if p.IsFlat() {
		return 0
	}
	return now.Sub(p.EntryTime)
}

// GetOpenHours returns how many hours the position has been open
func (p *Position) GetOpenHours(now time.Time) float64 {
	return p.GetOpenDuration(now).Hours()
}

// ==================== POSITION UPDATE METHODS ====================

// UpdatePrice updates the current market price (for unrealized P&L calculation)
func (p *Position) UpdatePrice(newPrice float64, pipValue float64) {
	p.CurrentPrice = newPrice

	if p.IsFlat() {
		p.UnrealizedPnL = 0
		return
	}

	// Calculate unrealized P&L based on position direction
	if p.IsLong() {
		// For long: profit when price goes up
		priceDiff := newPrice - p.EntryPrice
		p.UnrealizedPnL = priceDiff * p.Size / pipValue
	} else {
		// For short: profit when price goes down
		priceDiff := p.EntryPrice - newPrice
		p.UnrealizedPnL = priceDiff * p.GetAbsoluteSize() / pipValue
	}

	// Track peak/trough
	if p.UnrealizedPnL > p.MaxFavorableExcursion {
		p.MaxFavorableExcursion = p.UnrealizedPnL
	}
	if p.UnrealizedPnL < p.MaxAdverseExcursion {
		p.MaxAdverseExcursion = p.UnrealizedPnL
	}
}

// AddTrade adds a trade to the position history and updates position
func (p *Position) AddTrade(trade *Trade) {
	p.TradeHistory = append(p.TradeHistory, trade)
	p.TradeCount++

	if trade.IsEntry {
		// This is opening a new position
		p.EntryPrice = trade.Price
		p.EntryTime = trade.Timestamp
		p.EntryCommission = trade.Commission
	}

	if trade.IsExit && !p.IsFlat() {
		// This is closing or reducing the position
		p.RealizedPnL += trade.PnLAtClose
	}

	p.CommissionPaid += trade.Commission
}

// ==================== POSITION CALCULATIONS ====================

// CalculateUnrealizedPnL calculates unrealized P&L based on current price
// pipValue is the smallest price unit (0.0001 for Forex, 0.01 for stocks, etc)
func (p *Position) CalculateUnrealizedPnL(currentPrice, pipValue float64) float64 {
	if p.IsFlat() {
		return 0
	}

	if p.IsLong() {
		// For long positions: profit = (currentPrice - entryPrice) * size
		priceDiff := currentPrice - p.EntryPrice
		return priceDiff * p.Size / pipValue
	}

	// For short positions: profit = (entryPrice - currentPrice) * size
	priceDiff := p.EntryPrice - currentPrice
	return priceDiff * p.GetAbsoluteSize() / pipValue
}

// CalculateTotalPnL returns realized + unrealized P&L
func (p *Position) CalculateTotalPnL() float64 {
	return p.RealizedPnL + p.UnrealizedPnL - p.CommissionPaid
}

// CalculateROE returns Return On Equity (P&L / Entry Notional)
func (p *Position) CalculateROE() float64 {
	if p.IsFlat() {
		return 0
	}

	entryNotional := p.GetAbsoluteSize() * p.EntryPrice
	if entryNotional == 0 {
		return 0
	}

	return (p.CalculateTotalPnL() / entryNotional) * 100
}

// CalculateDrawdown returns the peak-to-trough drawdown
func (p *Position) CalculateDrawdown() float64 {
	if p.PeakProfit == 0 {
		return 0
	}
	return ((p.PeakProfit - p.MaxAdverseExcursion) / p.PeakProfit) * 100
}

// GetAverageEntryPrice returns the average entry price (same as EntryPrice in single position)
func (p *Position) GetAverageEntryPrice() float64 {
	return p.EntryPrice
}

// GetNotional returns the notional value of the position
func (p *Position) GetNotional() float64 {
	if p.IsFlat() {
		return 0
	}
	return p.GetAbsoluteSize() * p.CurrentPrice
}

// GetBreakevenPrice returns the breakeven price accounting for commission
func (p *Position) GetBreakevenPrice() float64 {
	if p.IsFlat() || p.EntryPrice == 0 {
		return 0
	}

	// Adjust entry price by commission impact
	if p.IsLong() {
		return p.EntryPrice + (p.CommissionPaid / p.Size)
	} else {
		return p.EntryPrice - (p.CommissionPaid / p.GetAbsoluteSize())
	}
}

// ==================== POSITION METRICS ====================

// GetMetrics returns a summary of position metrics
func (p *Position) GetMetrics(currentPrice, pipValue float64) map[string]interface{} {
	p.UpdatePrice(currentPrice, pipValue)

	return map[string]interface{}{
		"status":                  p.GetStatus(),
		"size":                    p.Size,
		"absolute_size":           p.GetAbsoluteSize(),
		"entry_price":             p.EntryPrice,
		"entry_time":              p.EntryTime,
		"current_price":           p.CurrentPrice,
		"breakeven_price":         p.GetBreakevenPrice(),
		"notional":                p.GetNotional(),
		"unrealized_pnl":          p.UnrealizedPnL,
		"realized_pnl":            p.RealizedPnL,
		"total_pnl":               p.CalculateTotalPnL(),
		"commission_paid":         p.CommissionPaid,
		"roe":                     p.CalculateROE(),
		"max_favorable_excursion": p.MaxFavorableExcursion,
		"max_adverse_excursion":   p.MaxAdverseExcursion,
		"drawdown":                p.CalculateDrawdown(),
		"trade_count":             p.TradeCount,
	}
}

// ==================== POSITION DISPLAY ====================

// String returns a human-readable representation
func (p *Position) String() string {
	if p.IsFlat() {
		return fmt.Sprintf("Position[FLAT]")
	}

	return fmt.Sprintf(
		"Position[%s %f @ %.5f | U/R P&L: %.2f | Total P&L: %.2f]",
		p.GetStatus(),
		p.GetAbsoluteSize(),
		p.EntryPrice,
		p.UnrealizedPnL,
		p.CalculateTotalPnL(),
	)
}

// DebugString returns detailed position information
func (p *Position) DebugString() string {
	tradeCount := len(p.TradeHistory)
	trades := "None"
	if tradeCount > 0 {
		trades = fmt.Sprintf("%d", tradeCount)
	}

	return fmt.Sprintf(
		"Position Details:\n"+
			"  Status:                 %s\n"+
			"  Size:                   %f\n"+
			"  Entry Price:            %.8f\n"+
			"  Entry Time:             %s\n"+
			"  Current Price:          %.8f\n"+
			"  Breakeven Price:        %.8f\n"+
			"  Notional Value:         %.2f\n"+
			"  Entry Commission:       %.2f\n"+
			"  Total Commission:       %.2f\n"+
			"\n"+
			"  P&L:\n"+
			"    Unrealized:          %.2f\n"+
			"    Realized:            %.2f\n"+
			"    Total:               %.2f\n"+
			"    ROE:                 %.2f%%\n"+
			"\n"+
			"  Excursions:\n"+
			"    Max Favorable:       %.2f\n"+
			"    Max Adverse:         %.2f\n"+
			"    Drawdown:            %.2f%%\n"+
			"\n"+
			"  Trades:\n"+
			"    Count:               %s\n"+
			"    First Trade:         %s",
		p.GetStatus(),
		p.Size,
		p.EntryPrice,
		p.EntryTime.Format("2006-01-02T15:04:05.000"),
		p.CurrentPrice,
		p.GetBreakevenPrice(),
		p.GetNotional(),
		p.EntryCommission,
		p.CommissionPaid,
		p.UnrealizedPnL,
		p.RealizedPnL,
		p.CalculateTotalPnL(),
		p.CalculateROE(),
		p.MaxFavorableExcursion,
		p.MaxAdverseExcursion,
		p.CalculateDrawdown(),
		trades,
		func() string {
			if len(p.TradeHistory) > 0 {
				return p.TradeHistory[0].Timestamp.Format("2006-01-02T15:04:05.000")
			}
			return "N/A"
		}(),
	)
}

// ==================== POSITION HISTORY ====================

// PositionHistory tracks position changes over time
type PositionHistory struct {
	Snapshots []*PositionSnapshot
}

// PositionSnapshot captures position state at a point in time
type PositionSnapshot struct {
	Timestamp     time.Time
	Size          float64
	EntryPrice    float64
	CurrentPrice  float64
	UnrealizedPnL float64
	RealizedPnL   float64
	TotalPnL      float64
}

// NewPositionHistory creates a new position history
func NewPositionHistory() *PositionHistory {
	return &PositionHistory{
		Snapshots: make([]*PositionSnapshot, 0),
	}
}

// AddSnapshot adds a position snapshot
func (ph *PositionHistory) AddSnapshot(snapshot *PositionSnapshot) {
	ph.Snapshots = append(ph.Snapshots, snapshot)
}

// TakeSnapshot creates a snapshot from current position
func (ph *PositionHistory) TakeSnapshot(pos *Position) {
	snapshot := &PositionSnapshot{
		Timestamp:     time.Now(),
		Size:          pos.Size,
		EntryPrice:    pos.EntryPrice,
		CurrentPrice:  pos.CurrentPrice,
		UnrealizedPnL: pos.UnrealizedPnL,
		RealizedPnL:   pos.RealizedPnL,
		TotalPnL:      pos.CalculateTotalPnL(),
	}
	ph.AddSnapshot(snapshot)
}

// Size returns number of snapshots
func (ph *PositionHistory) Size() int {
	return len(ph.Snapshots)
}

// GetLatest returns the most recent snapshot
func (ph *PositionHistory) GetLatest() *PositionSnapshot {
	if len(ph.Snapshots) == 0 {
		return nil
	}
	return ph.Snapshots[len(ph.Snapshots)-1]
}

// GetOldest returns the oldest snapshot
func (ph *PositionHistory) GetOldest() *PositionSnapshot {
	if len(ph.Snapshots) == 0 {
		return nil
	}
	return ph.Snapshots[0]
}

// ==================== POSITION COMPARISON ====================

// PositionComparison compares two positions
type PositionComparison struct {
	SizeDifference       float64
	PriceDifference      float64
	PnLDifference        float64
	CommissionDifference float64
}

// CompareTo compares this position to another
func (p *Position) CompareTo(other *Position) *PositionComparison {
	return &PositionComparison{
		SizeDifference:       p.Size - other.Size,
		PriceDifference:      p.EntryPrice - other.EntryPrice,
		PnLDifference:        p.CalculateTotalPnL() - other.CalculateTotalPnL(),
		CommissionDifference: p.CommissionPaid - other.CommissionPaid,
	}
}
