package account

// ==================== DRAWDOWN MANAGEMENT ====================

// UpdateDrawdown updates drawdown calculations after a balance change
func (a *Account) UpdateDrawdown() {
	// Update high watermark
	if a.CurrentBalance > a.HighWaterMark {
		a.HighWaterMark = a.CurrentBalance
	}

	// Update low watermark
	if a.CurrentBalance < a.LowWaterMark {
		a.LowWaterMark = a.CurrentBalance
	}

	// Calculate current drawdown
	drawdown := a.HighWaterMark - a.CurrentBalance
	drawdownPercent := (drawdown / a.HighWaterMark) * 100

	// Update max drawdown experienced
	if drawdownPercent > a.MaxDrawdownExperienced {
		a.MaxDrawdownExperienced = drawdownPercent
		a.MaxDrawdownAmount = drawdown
	}
}

// GetDrawdownPercent returns current drawdown percentage
func (a *Account) GetDrawdownPercent() float64 {
	if a.HighWaterMark == 0 {
		return 0
	}
	drawdown := a.HighWaterMark - a.CurrentBalance
	return (drawdown / a.HighWaterMark) * 100
}

// GetMaxDrawdownPercent returns maximum drawdown experienced
func (a *Account) GetMaxDrawdownPercent() float64 {
	return a.MaxDrawdownExperienced
}

// GetMaxDrawdownAmount returns maximum drawdown amount in currency
func (a *Account) GetMaxDrawdownAmount() float64 {
	return a.MaxDrawdownAmount
}

// IsDrawdownExceeded checks if current drawdown exceeds limit
func (a *Account) IsDrawdownExceeded() bool {
	return a.GetDrawdownPercent() > a.MaxDrawdownPercent
}

// GetRecoveryPercent returns recovery percentage from peak to current
func (a *Account) GetRecoveryPercent() float64 {
	if a.HighWaterMark == 0 {
		return 0
	}
	recovery := a.CurrentBalance - (a.HighWaterMark - a.MaxDrawdownAmount)
	return (recovery / a.MaxDrawdownAmount) * 100
}

// HighWaterMarkDistance returns distance from high watermark
func (a *Account) HighWaterMarkDistance() float64 {
	return a.HighWaterMark - a.CurrentBalance
}

// HighWaterMarkPercent returns distance as percentage
func (a *Account) HighWaterMarkPercent() float64 {
	if a.HighWaterMark == 0 {
		return 0
	}
	return (a.HighWaterMarkDistance() / a.HighWaterMark) * 100
}
