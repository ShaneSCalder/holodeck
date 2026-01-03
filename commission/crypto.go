package commission

import (
	"fmt"
)

// ==================== CRYPTO COMMISSION CALCULATOR ====================

// CryptoCommissionCalculator calculates CRYPTO commissions
// Commission: 0.2% of notional value
// Formula: (price × amount) × 0.002
type CryptoCommissionCalculator struct {
	// Constants
	CommissionRate float64 // 0.002 (0.2%)

	// Statistics
	totalCommission float64
	commissionCount int64
	totalNotional   float64
}

// ==================== CALCULATOR CREATION ====================

// NewCryptoCommissionCalculator creates a new CRYPTO commission calculator
func NewCryptoCommissionCalculator() *CryptoCommissionCalculator {
	return &CryptoCommissionCalculator{
		CommissionRate: 0.002,
	}
}

// ==================== CORE CALCULATION ====================

// CalculateCommission calculates CRYPTO commission
// Parameters:
//   - price: Price per unit (e.g., $45,250.50 per BTC)
//   - amount: Amount to trade (e.g., 0.5 BTC)
//
// Returns: Commission in USD
func (crc *CryptoCommissionCalculator) CalculateCommission(
	price float64,
	amount float64,
) (float64, error) {

	// Calculate notional value
	notional := price * amount

	// Calculate commission: notional × 0.2%
	commission := notional * crc.CommissionRate

	// Track statistics
	crc.totalCommission += commission
	crc.commissionCount++
	crc.totalNotional += notional

	return commission, nil
}

// CalculateBatchCommission calculates commission for multiple CRYPTO trades
func (crc *CryptoCommissionCalculator) CalculateBatchCommission(
	trades []CryptoCommissionInput,
) (float64, error) {

	totalCommission := 0.0

	for _, trade := range trades {
		commission, err := crc.CalculateCommission(trade.Price, trade.Amount)
		if err != nil {
			return 0, err
		}
		totalCommission += commission
	}

	return totalCommission, nil
}

// ==================== STATISTICS ====================

// GetTotalCommission returns total commission collected
func (crc *CryptoCommissionCalculator) GetTotalCommission() float64 {
	return crc.totalCommission
}

// GetCommissionCount returns number of commissions calculated
func (crc *CryptoCommissionCalculator) GetCommissionCount() int64 {
	return crc.commissionCount
}

// GetAverageCommission returns average commission per trade
func (crc *CryptoCommissionCalculator) GetAverageCommission() float64 {
	if crc.commissionCount == 0 {
		return 0
	}
	return crc.totalCommission / float64(crc.commissionCount)
}

// GetTotalNotional returns total notional value traded
func (crc *CryptoCommissionCalculator) GetTotalNotional() float64 {
	return crc.totalNotional
}

// GetAverageNotional returns average notional value per trade
func (crc *CryptoCommissionCalculator) GetAverageNotional() float64 {
	if crc.commissionCount == 0 {
		return 0
	}
	return crc.totalNotional / float64(crc.commissionCount)
}

// GetCommissionRatePercent returns the commission rate as percentage
func (crc *CryptoCommissionCalculator) GetCommissionRatePercent() float64 {
	return crc.CommissionRate * 100
}

// GetStatistics returns comprehensive CRYPTO commission statistics
func (crc *CryptoCommissionCalculator) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"total_commission":    crc.totalCommission,
		"commission_count":    crc.commissionCount,
		"average_commission":  crc.GetAverageCommission(),
		"total_notional":      crc.totalNotional,
		"average_notional":    crc.GetAverageNotional(),
		"commission_rate":     crc.CommissionRate,
		"commission_rate_pct": crc.GetCommissionRatePercent(),
	}
}

// ==================== DEBUG ====================

// String returns a human-readable representation
func (crc *CryptoCommissionCalculator) String() string {
	return fmt.Sprintf(
		"CryptoCommission[Total:$%.2f, Count:%d, Rate:%.2f%%]",
		crc.totalCommission,
		crc.commissionCount,
		crc.GetCommissionRatePercent(),
	)
}

// DebugString returns detailed debug information
func (crc *CryptoCommissionCalculator) DebugString() string {
	return fmt.Sprintf(
		"CRYPTO Commission Calculator:\n"+
			"  Total Commission:      $%.2f\n"+
			"  Commission Count:      %d\n"+
			"  Average Commission:    $%.2f\n"+
			"  Total Notional:        $%.2f\n"+
			"  Average Notional:      $%.2f\n"+
			"  Commission Rate:       %.4f (%.2f%%)",
		crc.totalCommission,
		crc.commissionCount,
		crc.GetAverageCommission(),
		crc.totalNotional,
		crc.GetAverageNotional(),
		crc.CommissionRate,
		crc.GetCommissionRatePercent(),
	)
}

// Reset resets calculator statistics
func (crc *CryptoCommissionCalculator) Reset() {
	crc.totalCommission = 0
	crc.commissionCount = 0
	crc.totalNotional = 0
}

// ==================== ANALYSIS ====================

// AnalyzeCommission provides detailed analysis of a single commission calculation
func (crc *CryptoCommissionCalculator) AnalyzeCommission(
	price float64,
	amount float64,
) *CryptoCommissionAnalysis {

	notional := price * amount
	commission := notional * crc.CommissionRate
	commissionPct := crc.CommissionRate * 100

	return &CryptoCommissionAnalysis{
		Price:          price,
		Amount:         amount,
		Notional:       notional,
		Commission:     commission,
		CommissionRate: crc.CommissionRate,
		CommissionPct:  commissionPct,
	}
}

// ==================== ANALYSIS TYPES ====================

// CryptoCommissionAnalysis provides detailed breakdown of a commission calculation
type CryptoCommissionAnalysis struct {
	Price          float64
	Amount         float64
	Notional       float64
	Commission     float64
	CommissionRate float64
	CommissionPct  float64
}

// String returns string representation
func (cca *CryptoCommissionAnalysis) String() string {
	return fmt.Sprintf(
		"CRYPTO: %.8f @ $%.2f = $%.2f notional = $%.2f commission (%.2f%%)",
		cca.Amount,
		cca.Price,
		cca.Notional,
		cca.Commission,
		cca.CommissionPct,
	)
}

// DebugString returns detailed debug information
func (cca *CryptoCommissionAnalysis) DebugString() string {
	return fmt.Sprintf(
		"CRYPTO Commission Analysis:\n"+
			"  Price Per Unit:        $%.2f\n"+
			"  Amount:                %.8f\n"+
			"  Notional Value:        $%.2f\n"+
			"  Commission Rate:       %.4f (%.2f%%)\n"+
			"  Commission:            $%.2f",
		cca.Price,
		cca.Amount,
		cca.Notional,
		cca.CommissionRate,
		cca.CommissionPct,
		cca.Commission,
	)
}

// ==================== INPUT TYPES ====================

// CryptoCommissionInput represents input for CRYPTO commission calculation
type CryptoCommissionInput struct {
	Price  float64 // Price per unit
	Amount float64 // Amount to trade
}
