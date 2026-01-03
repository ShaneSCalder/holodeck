package position

import (
	"time"
)

// ==================== POSITION TYPES ====================

// Position represents an open or closed trading position
type Position struct {
	// Identification
	PositionID string
	Symbol     string
	OpenTime   time.Time
	CloseTime  *time.Time

	// Position Details
	Type           string  // LONG, SHORT, FLAT
	Size           float64 // Positive for LONG, Negative for SHORT
	EntryPrice     float64
	AveragePrice   float64
	CurrentPrice   float64
	LastUpdateTime time.Time

	// P&L Tracking
	RealizedPnL    float64
	UnrealizedPnL  float64
	CommissionPaid float64
	TotalCost      float64

	// Risk Metrics
	PeakProfit            float64
	PeakLoss              float64
	MaxAdverseExcursion   float64
	MaxFavorableExcursion float64
	RunUp                 float64
	DrawDown              float64

	// Trade History
	Trades       []*PositionTrade
	TradeCount   int
	EntryTradeID string

	// Status
	Status   string // OPEN, CLOSED, PARTIAL
	IsActive bool
}

// ==================== POSITION TRADE ====================

// PositionTrade represents a single trade within a position
type PositionTrade struct {
	TradeID    string
	Timestamp  time.Time
	Action     string // BUY, SELL
	Size       float64
	Price      float64
	Commission float64
	Slippage   float64
	IsEntry    bool
	IsExit     bool
	PnLAtClose float64
}

// ==================== PORTFOLIO ====================

// Portfolio manages multiple positions
type Portfolio struct {
	positions map[string]*Position
	TotalPnL  float64
}

// ==================== CONSTRUCTORS ====================

// NewPosition creates a new open position
func NewPosition(id, symbol string, posType string, size, price float64) *Position {
	position := &Position{
		PositionID:            id,
		Symbol:                symbol,
		Type:                  posType,
		Size:                  size,
		EntryPrice:            price,
		AveragePrice:          price,
		CurrentPrice:          price,
		OpenTime:              time.Now(),
		Trades:                make([]*PositionTrade, 0),
		Status:                "OPEN",
		IsActive:              true,
		TotalCost:             size * price,
		MaxAdverseExcursion:   0,
		MaxFavorableExcursion: 0,
	}
	return position
}

// NewPortfolio creates a new portfolio
func NewPortfolio() *Portfolio {
	return &Portfolio{
		positions: make(map[string]*Position),
		TotalPnL:  0,
	}
}

// ==================== STATUS CHECKS ====================

// IsLong checks if position is long
func (p *Position) IsLong() bool {
	return p.Type == "LONG" && p.Size > 0
}

// IsShort checks if position is short
func (p *Position) IsShort() bool {
	return p.Type == "SHORT" && p.Size < 0
}

// IsFlat checks if position is flat (closed)
func (p *Position) IsFlat() bool {
	return p.Size == 0
}

// GetDirection returns position direction string
func (p *Position) GetDirection() string {
	if p.IsLong() {
		return "LONG"
	} else if p.IsShort() {
		return "SHORT"
	}
	return "FLAT"
}

// ==================== UTILITY ====================

// timePtr returns a pointer to a time
func timePtr(t time.Time) *time.Time {
	return &t
}
