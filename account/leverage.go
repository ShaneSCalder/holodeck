package account

// ==================== LEVERAGE MANAGEMENT ====================

// SetLeverage sets the account leverage
func (a *Account) SetLeverage(leverage float64) bool {
	if leverage > a.MaxLeverageAllowed {
		return false
	}
	if leverage < 1.0 {
		return false
	}
	a.Leverage = leverage
	a.UpdateMargin()
	return true
}

// GetLeverage returns the current leverage
func (a *Account) GetLeverage() float64 {
	return a.Leverage
}

// CanIncreaseLeverage checks if leverage can be increased
func (a *Account) CanIncreaseLeverage(newLeverage float64) bool {
	return newLeverage <= a.MaxLeverageAllowed && newLeverage > a.Leverage
}

// CanDecreaseLeverage checks if leverage can be decreased
func (a *Account) CanDecreaseLeverage(newLeverage float64) bool {
	return newLeverage >= 1.0 && newLeverage < a.Leverage
}

// ==================== MARGIN MANAGEMENT ====================

// UpdateMargin updates margin calculations
func (a *Account) UpdateMargin() {
	a.BuyingPower = a.CurrentBalance * a.Leverage
	a.AvailableMargin = a.BuyingPower - a.UsedMargin

	// Check if account should be blown
	if a.CurrentBalance <= 0 {
		a.Status = "BLOWN"
	} else if a.AvailableMargin <= 0 {
		a.Status = "AT_LIMIT"
	} else if a.Status == "BLOWN" || a.Status == "AT_LIMIT" {
		a.Status = "ACTIVE"
	}
}

// HasSufficientMargin checks if account has margin for trade
func (a *Account) HasSufficientMargin(requiredMargin float64) bool {
	return a.AvailableMargin >= requiredMargin
}

// GetAvailableMargin returns available margin
func (a *Account) GetAvailableMargin() float64 {
	return a.AvailableMargin
}

// GetUsedMargin returns used margin
func (a *Account) GetUsedMargin() float64 {
	return a.UsedMargin
}

// GetBuyingPower returns total buying power
func (a *Account) GetBuyingPower() float64 {
	return a.BuyingPower
}

// GetAvailableMarginPercent returns available margin as percentage of buying power
func (a *Account) GetAvailableMarginPercent() float64 {
	if a.BuyingPower == 0 {
		return 0
	}
	return (a.AvailableMargin / a.BuyingPower) * 100
}

// GetUsedMarginPercent returns used margin as percentage of buying power
func (a *Account) GetUsedMarginPercent() float64 {
	if a.BuyingPower == 0 {
		return 0
	}
	return (a.UsedMargin / a.BuyingPower) * 100
}

// RecordMarginUsed records margin being used by open positions
func (a *Account) RecordMarginUsed(marginAmount float64) {
	a.UsedMargin = marginAmount
	a.UpdateMargin()
}

// ReleaseMargin releases margin from closed positions
func (a *Account) ReleaseMargin(marginAmount float64) {
	if a.UsedMargin >= marginAmount {
		a.UsedMargin -= marginAmount
	} else {
		a.UsedMargin = 0
	}
	a.UpdateMargin()
}

// GetMarginLevel returns margin level percentage
// Formula: (CurrentBalance / UsedMargin) * 100
func (a *Account) GetMarginLevel() float64 {
	if a.UsedMargin == 0 {
		return 0
	}
	return (a.CurrentBalance / a.UsedMargin) * 100
}

// IsMarginCall checks if margin call condition is met
// Typically margin level drops below 100%
func (a *Account) IsMarginCall() bool {
	return a.GetMarginLevel() < 100 && a.UsedMargin > 0
}
