package position

import (
	"fmt"
	"time"
)

// ==================== TRADE ENTRY ====================

// AddTrade adds a trade to the position history
func (p *Position) AddTrade(trade *PositionTrade) {
	p.Trades = append(p.Trades, trade)
	p.TradeCount++
	p.LastUpdateTime = time.Now()

	if trade.IsEntry && p.EntryTradeID == "" {
		p.EntryTradeID = trade.TradeID
	}

	// Update commission
	p.CommissionPaid += trade.Commission

	// Update average price for slippage
	if trade.IsEntry {
		// Adjust entry price for slippage
		if trade.Action == "BUY" {
			p.AveragePrice = (p.EntryPrice*p.Size + trade.Price*trade.Size + trade.Slippage) / (p.Size + trade.Size)
		} else {
			p.AveragePrice = (p.EntryPrice*p.Size + trade.Price*trade.Size - trade.Slippage) / (p.Size + trade.Size)
		}
	}
}

// ==================== POSITION CLOSING ====================

// ClosePartial closes part of the position
func (p *Position) ClosePartial(closeSize float64, closePrice float64, commission float64) float64 {
	if closeSize > p.Size {
		closeSize = p.Size
	}

	// Calculate P&L for this close
	var closePnL float64
	if p.Type == "LONG" {
		closePnL = (closePrice - p.AveragePrice) * closeSize
	} else if p.Type == "SHORT" {
		closePnL = (p.AveragePrice - closePrice) * closeSize
	}
	closePnL -= commission

	p.RealizedPnL += closePnL
	p.Size -= closeSize
	p.CommissionPaid += commission

	if p.Size == 0 {
		p.Status = "CLOSED"
		p.IsActive = false
		p.CloseTime = timePtr(time.Now())
	} else {
		p.Status = "PARTIAL"
	}

	p.UpdatePnL()
	return closePnL
}

// Close closes the entire position
func (p *Position) Close(closePrice float64, commission float64) float64 {
	return p.ClosePartial(p.Size, closePrice, commission)
}

// ==================== STATISTICS ====================

// GetDuration returns position duration
func (p *Position) GetDuration() time.Duration {
	endTime := time.Now()
	if p.CloseTime != nil {
		endTime = *p.CloseTime
	}
	return endTime.Sub(p.OpenTime)
}

// GetTradeCount returns number of trades in position
func (p *Position) GetTradeCount() int {
	return p.TradeCount
}

// GetEntryTime returns the entry time
func (p *Position) GetEntryTime() time.Time {
	return p.OpenTime
}

// ==================== STRING REPRESENTATION ====================

// String returns formatted position info
func (p *Position) String() string {
	return fmt.Sprintf(
		"%s %s %.2f @ %.4f | P&L: %.2f | Status: %s",
		p.Symbol, p.GetDirection(), p.Size, p.AveragePrice,
		p.GetProfit(), p.Status,
	)
}

// Details returns detailed position information
func (p *Position) Details() string {
	duration := p.GetDuration()
	durationStr := fmt.Sprintf("%d days %d hours", duration/24, (duration%24)/time.Hour)
	if p.CloseTime != nil {
		durationStr = fmt.Sprintf("%d days %d hours", duration/24, (duration%24)/time.Hour)
	}

	return fmt.Sprintf(
		"Position ID:            %s\n"+
			"Symbol:                 %s\n"+
			"Type:                   %s\n"+
			"Direction:              %s\n"+
			"Size:                   %.4f\n"+
			"Entry Price:            %.6f\n"+
			"Average Price:          %.6f\n"+
			"Current Price:          %.6f\n"+
			"Open Time:              %s\n"+
			"Duration:               %s\n"+
			"Status:                 %s\n"+
			"Realized P&L:           %.2f\n"+
			"Unrealized P&L:         %.2f\n"+
			"Total P&L:              %.2f\n"+
			"Commission Paid:        %.2f\n"+
			"Peak Profit:            %.2f\n"+
			"Peak Loss:              %.2f\n"+
			"Max Favorable Excursion: %.2f\n"+
			"Max Adverse Excursion:   %.2f\n"+
			"Trade Count:            %d",
		p.PositionID, p.Symbol, p.Type, p.GetDirection(), p.Size,
		p.EntryPrice, p.AveragePrice, p.CurrentPrice, p.OpenTime.Format("2006-01-02 15:04:05"),
		durationStr, p.Status,
		p.RealizedPnL, p.UnrealizedPnL, p.RealizedPnL+p.UnrealizedPnL,
		p.CommissionPaid, p.PeakProfit, p.PeakLoss,
		p.MaxFavorableExcursion, p.MaxAdverseExcursion,
		p.TradeCount,
	)
}

// ==================== PORTFOLIO MANAGEMENT ====================

// Add adds a position to the portfolio
func (pf *Portfolio) Add(position *Position) {
	pf.positions[position.PositionID] = position
}

// Remove removes a position from the portfolio
func (pf *Portfolio) Remove(positionID string) {
	delete(pf.positions, positionID)
}

// Get retrieves a position by ID
func (pf *Portfolio) Get(positionID string) (*Position, bool) {
	pos, ok := pf.positions[positionID]
	return pos, ok
}

// GetBySymbol gets all positions for a symbol
func (pf *Portfolio) GetBySymbol(symbol string) []*Position {
	var positions []*Position
	for _, pos := range pf.positions {
		if pos.Symbol == symbol {
			positions = append(positions, pos)
		}
	}
	return positions
}

// List returns all positions in portfolio
func (pf *Portfolio) List() []*Position {
	positions := make([]*Position, 0, len(pf.positions))
	for _, pos := range pf.positions {
		positions = append(positions, pos)
	}
	return positions
}

// Count returns number of active (open) positions
func (pf *Portfolio) Count() int {
	count := 0
	for _, pos := range pf.positions {
		if pos.IsActive {
			count++
		}
	}
	return count
}

// TotalExposure calculates total exposure (sum of all sizes)
func (pf *Portfolio) TotalExposure() float64 {
	var total float64
	for _, pos := range pf.positions {
		if pos.IsActive {
			total += pos.Size
		}
	}
	return total
}

// UpdatePrices updates prices for all positions
func (pf *Portfolio) UpdatePrices(prices map[string]float64) {
	for _, pos := range pf.positions {
		if price, ok := prices[pos.Symbol]; ok {
			pos.UpdatePrice(price)
		}
	}
	pf.UpdateTotalPnL()
}

// UpdateTotalPnL recalculates total portfolio P&L
func (pf *Portfolio) UpdateTotalPnL() {
	pf.TotalPnL = 0
	for _, pos := range pf.positions {
		if pos.IsActive {
			pf.TotalPnL += pos.UnrealizedPnL
		} else {
			pf.TotalPnL += pos.RealizedPnL
		}
	}
}

// GetTotalPnL returns total portfolio P&L
func (pf *Portfolio) GetTotalPnL() float64 {
	return pf.TotalPnL
}
