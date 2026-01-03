package position

import (
	"time"
)

// ==================== P&L UPDATES ====================

// UpdatePrice updates the current price and recalculates P&L
func (p *Position) UpdatePrice(newPrice float64) {
	p.CurrentPrice = newPrice
	p.LastUpdateTime = time.Now()
	p.UpdatePnL()
}

// UpdatePnL updates all P&L calculations
func (p *Position) UpdatePnL() {
	if p.Size == 0 {
		p.UnrealizedPnL = 0
		return
	}

	// Calculate unrealized P&L
	if p.Type == "LONG" {
		p.UnrealizedPnL = (p.CurrentPrice - p.AveragePrice) * p.Size
	} else if p.Type == "SHORT" {
		p.UnrealizedPnL = (p.AveragePrice - p.CurrentPrice) * p.Size
	} else {
		p.UnrealizedPnL = 0
	}

	// Subtract commissions
	p.UnrealizedPnL -= p.CommissionPaid

	// Update peak profit/loss
	if p.UnrealizedPnL > p.PeakProfit {
		p.PeakProfit = p.UnrealizedPnL
		p.MaxFavorableExcursion = p.PeakProfit
	}
	if p.UnrealizedPnL < p.PeakLoss {
		p.PeakLoss = p.UnrealizedPnL
		if p.PeakLoss < p.MaxAdverseExcursion {
			p.MaxAdverseExcursion = p.PeakLoss
		}
	}

	// Update run-up and draw-down
	if p.PeakProfit > 0 {
		p.RunUp = p.PeakProfit
	}
	if p.PeakLoss < 0 {
		p.DrawDown = -p.PeakLoss
	}
}

// ==================== P&L QUERIES ====================

// IsProfitable checks if position is currently profitable
func (p *Position) IsProfitable() bool {
	return p.UnrealizedPnL > 0
}

// IsNegative checks if position is currently losing money
func (p *Position) IsNegative() bool {
	return p.UnrealizedPnL < 0
}

// GetProfit returns current profit/loss (realized or unrealized)
func (p *Position) GetProfit() float64 {
	if p.IsActive {
		return p.UnrealizedPnL
	}
	return p.RealizedPnL
}

// GetTotalPnL returns total P&L (realized + unrealized)
func (p *Position) GetTotalPnL() float64 {
	return p.RealizedPnL + p.UnrealizedPnL
}

// GetNetPnL returns net P&L after commissions
func (p *Position) GetNetPnL() float64 {
	return p.GetTotalPnL() - p.CommissionPaid
}

// ==================== RISK METRICS ====================

// GetRatio returns profit/loss ratio (peak profit / peak loss)
func (p *Position) GetRatio() float64 {
	if p.PeakLoss >= 0 {
		return 0
	}
	return p.PeakProfit / (-p.PeakLoss)
}

// GetRiskReward returns risk/reward ratio (MFE / MAE)
func (p *Position) GetRiskReward() float64 {
	if p.MaxAdverseExcursion >= 0 {
		return 0
	}
	if p.MaxFavorableExcursion <= 0 {
		return 0
	}
	return p.MaxFavorableExcursion / (-p.MaxAdverseExcursion)
}

// GetMaxFavorableExcursion returns maximum favorable excursion (MFE)
func (p *Position) GetMaxFavorableExcursion() float64 {
	return p.MaxFavorableExcursion
}

// GetMaxAdverseExcursion returns maximum adverse excursion (MAE)
func (p *Position) GetMaxAdverseExcursion() float64 {
	return p.MaxAdverseExcursion
}

// GetRunUp returns maximum run-up from entry
func (p *Position) GetRunUp() float64 {
	return p.RunUp
}

// GetDrawDown returns maximum draw-down from entry
func (p *Position) GetDrawDown() float64 {
	return p.DrawDown
}
