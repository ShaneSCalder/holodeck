package commission

import (
	"fmt"
)

// ==================== FOREX COMMISSION CALCULATOR ====================

// ForexCommissionCalculator calculates FOREX commissions
// Commission: $25 per $1,000,000 notional
// Formula: (price × size × contractSize / 1,000,000) × 25
type ForexCommissionCalculator struct {
	// Constants
	CommissionPerMillion float64 // $25
	ContractSize         int64   // 100,000 units per lot

	// Statistics
	totalCommission float64
	commissionCount int64
	totalNotional   float64
}

// ==================== CALCULATOR CREATION ====================

// NewForexCommissionCalculator creates a new FOREX commission calculator
func NewForexCommissionCalculator() *ForexCommissionCalculator {
	return &ForexCommissionCalculator{
		CommissionPerMillion: 25.0,
		ContractSize:         100000,
	}
}

// ==================== CORE CALCULATION ====================

// CalculateCommission calculates FOREX commission
// Parameters:
//   - price: Price per unit (e.g., 1.08505 for EUR/USD)
//   - size: Size in lots (e.g., 0.01 = 1,000 units)
//
// Returns: Commission in USD
func (fcc *ForexCommissionCalculator) CalculateCommission(
	price float64,
	sizeInLots float64,
) (float64, error) {

	// Convert lots to units
	sizeInUnits := sizeInLots * float64(fcc.ContractSize)

	// Calculate notional value in USD
	notional := price * sizeInUnits

	// Calculate commission: (notional / 1,000,000) × $25
	commission := (notional / 1000000.0) * fcc.CommissionPerMillion

	// Track statistics
	fcc.totalCommission += commission
	fcc.commissionCount++
	fcc.totalNotional += notional

	return commission, nil
}

// CalculateBatchCommission calculates commission for multiple FOREX trades
func (fcc *ForexCommissionCalculator) CalculateBatchCommission(
	trades []ForexCommissionInput,
) (float64, error) {

	totalCommission := 0.0

	for _, trade := range trades {
		commission, err := fcc.CalculateCommission(trade.Price, trade.SizeInLots)
		if err != nil {
			return 0, err
		}
		totalCommission += commission
	}

	return totalCommission, nil
}

// ==================== STATISTICS ====================

// GetTotalCommission returns total commission collected
func (fcc *ForexCommissionCalculator) GetTotalCommission() float64 {
	return fcc.totalCommission
}

// GetCommissionCount returns number of commissions calculated
func (fcc *ForexCommissionCalculator) GetCommissionCount() int64 {
	return fcc.commissionCount
}

// GetAverageCommission returns average commission per trade
func (fcc *ForexCommissionCalculator) GetAverageCommission() float64 {
	if fcc.commissionCount == 0 {
		return 0
	}
	return fcc.totalCommission / float64(fcc.commissionCount)
}

// GetTotalNotional returns total notional value traded
func (fcc *ForexCommissionCalculator) GetTotalNotional() float64 {
	return fcc.totalNotional
}

// GetAverageNotional returns average notional value per trade
func (fcc *ForexCommissionCalculator) GetAverageNotional() float64 {
	if fcc.commissionCount == 0 {
		return 0
	}
	return fcc.totalNotional / float64(fcc.commissionCount)
}

// GetCommissionRate returns the effective commission rate as percentage
func (fcc *ForexCommissionCalculator) GetCommissionRate() float64 {
	if fcc.totalNotional == 0 {
		return 0
	}
	return (fcc.totalCommission / fcc.totalNotional) * 100
}

// GetStatistics returns comprehensive FOREX commission statistics
func (fcc *ForexCommissionCalculator) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"total_commission":    fcc.totalCommission,
		"commission_count":    fcc.commissionCount,
		"average_commission":  fcc.GetAverageCommission(),
		"total_notional":      fcc.totalNotional,
		"average_notional":    fcc.GetAverageNotional(),
		"commission_rate_pct": fcc.GetCommissionRate(),
		"contract_size":       fcc.ContractSize,
		"commission_per_mm":   fcc.CommissionPerMillion,
	}
}

// ==================== DEBUG ====================

// String returns a human-readable representation
func (fcc *ForexCommissionCalculator) String() string {
	return fmt.Sprintf(
		"ForexCommission[Total:$%.2f, Count:%d, Rate:%.4f%%]",
		fcc.totalCommission,
		fcc.commissionCount,
		fcc.GetCommissionRate(),
	)
}

// DebugString returns detailed debug information
func (fcc *ForexCommissionCalculator) DebugString() string {
	return fmt.Sprintf(
		"FOREX Commission Calculator:\n"+
			"  Total Commission:      $%.2f\n"+
			"  Commission Count:      %d\n"+
			"  Average Commission:    $%.2f\n"+
			"  Total Notional:        $%.2f\n"+
			"  Average Notional:      $%.2f\n"+
			"  Commission Rate:       %.4f%%\n"+
			"  Contract Size:         %d units\n"+
			"  Rate:                  $%.2f per $1M",
		fcc.totalCommission,
		fcc.commissionCount,
		fcc.GetAverageCommission(),
		fcc.totalNotional,
		fcc.GetAverageNotional(),
		fcc.GetCommissionRate(),
		fcc.ContractSize,
		fcc.CommissionPerMillion,
	)
}

// Reset resets calculator statistics
func (fcc *ForexCommissionCalculator) Reset() {
	fcc.totalCommission = 0
	fcc.commissionCount = 0
	fcc.totalNotional = 0
}

// ==================== ANALYSIS ====================

// AnalyzeCommission provides detailed analysis of a single commission calculation
func (fcc *ForexCommissionCalculator) AnalyzeCommission(
	price float64,
	sizeInLots float64,
) *ForexCommissionAnalysis {

	sizeInUnits := sizeInLots * float64(fcc.ContractSize)
	notional := price * sizeInUnits
	commission := (notional / 1000000.0) * fcc.CommissionPerMillion
	commissionPct := (commission / notional) * 100

	return &ForexCommissionAnalysis{
		Price:           price,
		SizeInLots:      sizeInLots,
		SizeInUnits:     sizeInUnits,
		Notional:        notional,
		Commission:      commission,
		CommissionPct:   commissionPct,
		ContractSize:    fcc.ContractSize,
		CommissionPerMM: fcc.CommissionPerMillion,
	}
}

// ==================== ANALYSIS TYPES ====================

// ForexCommissionAnalysis provides detailed breakdown of a commission calculation
type ForexCommissionAnalysis struct {
	Price           float64
	SizeInLots      float64
	SizeInUnits     float64
	Notional        float64
	Commission      float64
	CommissionPct   float64
	ContractSize    int64
	CommissionPerMM float64
}

// String returns string representation
func (fca *ForexCommissionAnalysis) String() string {
	return fmt.Sprintf(
		"FOREX: %.2f lots @ %.5f = $%.2f notional = $%.2f commission (%.4f%%)",
		fca.SizeInLots,
		fca.Price,
		fca.Notional,
		fca.Commission,
		fca.CommissionPct,
	)
}

// DebugString returns detailed debug information
func (fca *ForexCommissionAnalysis) DebugString() string {
	return fmt.Sprintf(
		"FOREX Commission Analysis:\n"+
			"  Price:                 %.8f\n"+
			"  Size (lots):           %.6f\n"+
			"  Size (units):          %.0f\n"+
			"  Contract Size:         %d\n"+
			"  Notional Value:        $%.2f\n"+
			"  Commission Rate:       $%.2f per $1M\n"+
			"  Commission:            $%.2f\n"+
			"  Commission Pct:        %.6f%%",
		fca.Price,
		fca.SizeInLots,
		fca.SizeInUnits,
		fca.ContractSize,
		fca.Notional,
		fca.CommissionPerMM,
		fca.Commission,
		fca.CommissionPct,
	)
}

// ==================== INPUT TYPES ====================

// ForexCommissionInput represents input for FOREX commission calculation
type ForexCommissionInput struct {
	Price      float64 // Price per unit
	SizeInLots float64 // Size in lots (0.01 = 1,000 units)
}
