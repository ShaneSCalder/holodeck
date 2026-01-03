package types

import (
	"fmt"
	"time"
)

// ==================== TICK STRUCTURE ====================

// Tick represents a single market data point (price quote)
// This is the most granular data unit - one tick per timestamp
type Tick struct {
	// Timestamp of the tick (when this price occurred)
	Timestamp time.Time

	// Bid price (price we can SELL at)
	Bid float64

	// Ask price (price we can BUY at)
	Ask float64

	// Bid quantity (volume available at bid price)
	BidQty int64

	// Ask quantity (volume available at ask price)
	AskQty int64

	// Last executed price (actual last traded price)
	LastPrice float64

	// Tick volume (number of shares/contracts traded in this tick)
	Volume int64

	// Sequence number (monotonic counter for ordering)
	Sequence int64

	// Spread in pips (calculated, not from CSV)
	SpreadPips float64

	// Mid price (calculated as (Bid + Ask) / 2)
	MidPrice float64
}

// ==================== TICK METHODS ====================

// NewTick creates a new Tick with calculated fields
// bid, ask, and lastPrice should be in instrument units (e.g., 1.08505 for EURUSD)
// bidQty, askQty should be in contract units (e.g., 500000 for EURUSD lot)
func NewTick(timestamp time.Time, bid, ask, lastPrice float64, bidQty, askQty, volume int64, sequence int64) *Tick {
	t := &Tick{
		Timestamp: timestamp,
		Bid:       bid,
		Ask:       ask,
		LastPrice: lastPrice,
		BidQty:    bidQty,
		AskQty:    askQty,
		Volume:    volume,
		Sequence:  sequence,
	}

	// Calculate derived fields
	t.MidPrice = (bid + ask) / 2.0
	t.SpreadPips = ask - bid // Spread in decimal units (will be converted to pips by instrument)

	return t
}

// GetMidPrice returns the mid price (average of bid and ask)
func (t *Tick) GetMidPrice() float64 {
	return t.MidPrice
}

// GetSpread returns the spread (difference between ask and bid)
func (t *Tick) GetSpread() float64 {
	return t.Ask - t.Bid
}

// GetSpreadPips returns the spread in the calculation format
// (same as decimal units at this level, will be converted by instrument)
func (t *Tick) GetSpreadPips() float64 {
	return t.SpreadPips
}

// GetBidAskCenter returns the mid price (same as MidPrice)
func (t *Tick) GetBidAskCenter() float64 {
	return (t.Bid + t.Ask) / 2.0
}

// GetAvailableDepth returns the available depth (minimum of bid and ask qty)
// Used for partial fill and slippage calculations
func (t *Tick) GetAvailableDepth() int64 {
	if t.BidQty < t.AskQty {
		return t.BidQty
	}
	return t.AskQty
}

// GetBuyPrice returns the price we would pay to BUY (the Ask price)
// This is the price for market BUY orders
func (t *Tick) GetBuyPrice() float64 {
	return t.Ask
}

// GetSellPrice returns the price we would receive to SELL (the Bid price)
// This is the price for market SELL orders
func (t *Tick) GetSellPrice() float64 {
	return t.Bid
}

// GetAskQtyAvailable returns the quantity available at ask price (for buying)
func (t *Tick) GetAskQtyAvailable() int64 {
	return t.AskQty
}

// GetBidQtyAvailable returns the quantity available at bid price (for selling)
func (t *Tick) GetBidQtyAvailable() int64 {
	return t.BidQty
}

// IsValid checks if the tick has valid price data
func (t *Tick) IsValid() bool {
	// Check that prices are positive
	if t.Bid <= 0 || t.Ask <= 0 || t.LastPrice <= 0 {
		return false
	}

	// Check that bid < ask (normal market structure)
	if t.Bid > t.Ask {
		return false
	}

	// Check that quantities are non-negative
	if t.BidQty < 0 || t.AskQty < 0 {
		return false
	}

	// Check that volume is non-negative
	if t.Volume < 0 {
		return false
	}

	// Check that last price is between bid and ask (usually)
	// Allow small deviation for edge cases
	if t.LastPrice < (t.Bid-0.001) || t.LastPrice > (t.Ask+0.001) {
		// This is a warning but not necessarily invalid
	}

	return true
}

// String returns a human-readable representation of the tick
func (t *Tick) String() string {
	return fmt.Sprintf(
		"Tick[Time=%s, Bid=%.5f, Ask=%.5f, Mid=%.5f, BidQty=%d, AskQty=%d, Volume=%d, Seq=%d]",
		t.Timestamp.Format("2006-01-02T15:04:05.000"),
		t.Bid,
		t.Ask,
		t.MidPrice,
		t.BidQty,
		t.AskQty,
		t.Volume,
		t.Sequence,
	)
}

// DebugString returns detailed debug information about the tick
func (t *Tick) DebugString() string {
	return fmt.Sprintf(
		"Tick Details:\n"+
			"  Timestamp:   %s\n"+
			"  Bid:         %.8f (Qty: %d)\n"+
			"  Ask:         %.8f (Qty: %d)\n"+
			"  Mid Price:   %.8f\n"+
			"  Last Price:  %.8f\n"+
			"  Spread:      %.8f\n"+
			"  Volume:      %d\n"+
			"  Sequence:    %d\n"+
			"  Valid:       %v",
		t.Timestamp.Format("2006-01-02T15:04:05.000000"),
		t.Bid, t.BidQty,
		t.Ask, t.AskQty,
		t.MidPrice,
		t.LastPrice,
		t.GetSpread(),
		t.Volume,
		t.Sequence,
		t.IsValid(),
	)
}

// ==================== TICK COLLECTION ====================

// TickBuffer holds a collection of ticks
type TickBuffer struct {
	Ticks    []*Tick
	MaxSize  int
	Sequence int64
}

// NewTickBuffer creates a new tick buffer with max size
func NewTickBuffer(maxSize int) *TickBuffer {
	return &TickBuffer{
		Ticks:   make([]*Tick, 0, maxSize),
		MaxSize: maxSize,
	}
}

// Add adds a tick to the buffer, removing oldest if at capacity
func (tb *TickBuffer) Add(tick *Tick) {
	if len(tb.Ticks) >= tb.MaxSize {
		// Remove oldest (first) tick
		tb.Ticks = tb.Ticks[1:]
	}
	tb.Ticks = append(tb.Ticks, tick)
	tb.Sequence = tick.Sequence
}

// GetLatest returns the most recent tick
func (tb *TickBuffer) GetLatest() *Tick {
	if len(tb.Ticks) == 0 {
		return nil
	}
	return tb.Ticks[len(tb.Ticks)-1]
}

// GetOldest returns the oldest tick in buffer
func (tb *TickBuffer) GetOldest() *Tick {
	if len(tb.Ticks) == 0 {
		return nil
	}
	return tb.Ticks[0]
}

// GetByIndex returns tick at specific index (0 = oldest, -1 = newest)
func (tb *TickBuffer) GetByIndex(index int) *Tick {
	if index < 0 || index >= len(tb.Ticks) {
		return nil
	}
	return tb.Ticks[index]
}

// Size returns current number of ticks in buffer
func (tb *TickBuffer) Size() int {
	return len(tb.Ticks)
}

// IsFull checks if buffer is at capacity
func (tb *TickBuffer) IsFull() bool {
	return len(tb.Ticks) >= tb.MaxSize
}

// Clear empties the buffer
func (tb *TickBuffer) Clear() {
	tb.Ticks = make([]*Tick, 0, tb.MaxSize)
	tb.Sequence = 0
}

// ==================== TICK STATISTICS ====================

// TickStats holds calculated statistics about ticks
type TickStats struct {
	// High is the highest price in period
	High float64

	// Low is the lowest price in period
	Low float64

	// Open is the first price in period
	Open float64

	// Close is the last price in period
	Close float64

	// AverageBid is the average bid price
	AverageBid float64

	// AverageAsk is the average ask price
	AverageAsk float64

	// AverageSpread is the average bid-ask spread
	AverageSpread float64

	// TotalVolume is the sum of all tick volumes
	TotalVolume int64

	// TickCount is the number of ticks
	TickCount int

	// TimeRange is the time from first to last tick
	TimeRange time.Duration
}

// CalculateStats calculates statistics from a set of ticks
func CalculateStats(ticks []*Tick) *TickStats {
	if len(ticks) == 0 {
		return &TickStats{}
	}

	stats := &TickStats{
		Open:      ticks[0].LastPrice,
		Close:     ticks[len(ticks)-1].LastPrice,
		High:      ticks[0].LastPrice,
		Low:       ticks[0].LastPrice,
		TickCount: len(ticks),
	}

	var sumBid, sumAsk, sumSpread float64

	for _, tick := range ticks {
		// Update High/Low
		if tick.LastPrice > stats.High {
			stats.High = tick.LastPrice
		}
		if tick.LastPrice < stats.Low {
			stats.Low = tick.LastPrice
		}

		// Accumulate for averages
		sumBid += tick.Bid
		sumAsk += tick.Ask
		sumSpread += tick.GetSpread()

		// Volume
		stats.TotalVolume += tick.Volume
	}

	// Calculate averages
	if len(ticks) > 0 {
		stats.AverageBid = sumBid / float64(len(ticks))
		stats.AverageAsk = sumAsk / float64(len(ticks))
		stats.AverageSpread = sumSpread / float64(len(ticks))
	}

	// Time range
	if len(ticks) > 1 {
		stats.TimeRange = ticks[len(ticks)-1].Timestamp.Sub(ticks[0].Timestamp)
	}

	return stats
}

// String returns a human-readable representation of tick stats
func (ts *TickStats) String() string {
	return fmt.Sprintf(
		"TickStats[O=%.5f, H=%.5f, L=%.5f, C=%.5f, AvgSpread=%.5f, Volume=%d, Ticks=%d]",
		ts.Open, ts.High, ts.Low, ts.Close, ts.AverageSpread, ts.TotalVolume, ts.TickCount,
	)
}
