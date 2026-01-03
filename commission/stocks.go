package commission

import (
	"fmt"
)

// ==================== STOCKS COMMISSION CALCULATOR ====================

// StocksCommissionCalculator calculates STOCKS commissions
// Commission: $0.01 per share
// Formula: shares × $0.01
type StocksCommissionCalculator struct {
	// Constants
	CommissionPerShare float64 // $0.01

	// Statistics
	totalCommission float64
	commissionCount int64
	totalShares     float64
}

// ==================== CALCULATOR CREATION ====================

// NewStocksCommissionCalculator creates a new STOCKS commission calculator
func NewStocksCommissionCalculator() *StocksCommissionCalculator {
	return &StocksCommissionCalculator{
		CommissionPerShare: 0.01,
	}
}

// ==================== CORE CALCULATION ====================

// CalculateCommission calculates STOCKS commission
// Parameters:
//   - shares: Number of shares
//
// Returns: Commission in USD
func (scc *StocksCommissionCalculator) CalculateCommission(
	shares float64,
) (float64, error) {

	// Calculate commission: shares × $0.01
	commission := shares * scc.CommissionPerShare

	// Track statistics
	scc.totalCommission += commission
	scc.commissionCount++
	scc.totalShares += shares

	return commission, nil
}

// CalculateBatchCommission calculates commission for multiple STOCKS trades
func (scc *StocksCommissionCalculator) CalculateBatchCommission(
	trades []StocksCommissionInput,
) (float64, error) {

	totalCommission := 0.0

	for _, trade := range trades {
		commission, err := scc.CalculateCommission(trade.Shares)
		if err != nil {
			return 0, err
		}
		totalCommission += commission
	}

	return totalCommission, nil
}

// ==================== STATISTICS ====================

// GetTotalCommission returns total commission collected
func (scc *StocksCommissionCalculator) GetTotalCommission() float64 {
	return scc.totalCommission
}

// GetCommissionCount returns number of commissions calculated
func (scc *StocksCommissionCalculator) GetCommissionCount() int64 {
	return scc.commissionCount
}

// GetAverageCommission returns average commission per trade
func (scc *StocksCommissionCalculator) GetAverageCommission() float64 {
	if scc.commissionCount == 0 {
		return 0
	}
	return scc.totalCommission / float64(scc.commissionCount)
}

// GetTotalShares returns total shares traded
func (scc *StocksCommissionCalculator) GetTotalShares() float64 {
	return scc.totalShares
}

// GetAverageShares returns average shares per trade
func (scc *StocksCommissionCalculator) GetAverageShares() float64 {
	if scc.commissionCount == 0 {
		return 0
	}
	return scc.totalShares / float64(scc.commissionCount)
}

// GetStatistics returns comprehensive STOCKS commission statistics
func (scc *StocksCommissionCalculator) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"total_commission":     scc.totalCommission,
		"commission_count":     scc.commissionCount,
		"average_commission":   scc.GetAverageCommission(),
		"total_shares":         scc.totalShares,
		"average_shares":       scc.GetAverageShares(),
		"commission_per_share": scc.CommissionPerShare,
	}
}

// ==================== DEBUG ====================

// String returns a human-readable representation
func (scc *StocksCommissionCalculator) String() string {
	return fmt.Sprintf(
		"StocksCommission[Total:$%.2f, Count:%d, Shares:%.0f]",
		scc.totalCommission,
		scc.commissionCount,
		scc.totalShares,
	)
}

// DebugString returns detailed debug information
func (scc *StocksCommissionCalculator) DebugString() string {
	return fmt.Sprintf(
		"STOCKS Commission Calculator:\n"+
			"  Total Commission:      $%.2f\n"+
			"  Commission Count:      %d\n"+
			"  Average Commission:    $%.2f\n"+
			"  Total Shares:          %.0f\n"+
			"  Average Shares:        %.2f\n"+
			"  Commission Per Share:  $%.4f",
		scc.totalCommission,
		scc.commissionCount,
		scc.GetAverageCommission(),
		scc.totalShares,
		scc.GetAverageShares(),
		scc.CommissionPerShare,
	)
}

// Reset resets calculator statistics
func (scc *StocksCommissionCalculator) Reset() {
	scc.totalCommission = 0
	scc.commissionCount = 0
	scc.totalShares = 0
}

// ==================== ANALYSIS ====================

// AnalyzeCommission provides detailed analysis of a single commission calculation
func (scc *StocksCommissionCalculator) AnalyzeCommission(
	shares float64,
) *StocksCommissionAnalysis {

	commission := shares * scc.CommissionPerShare

	return &StocksCommissionAnalysis{
		Shares:     shares,
		Commission: commission,
		Rate:       scc.CommissionPerShare,
	}
}

// ==================== ANALYSIS TYPES ====================

// StocksCommissionAnalysis provides detailed breakdown of a commission calculation
type StocksCommissionAnalysis struct {
	Shares     float64
	Commission float64
	Rate       float64
}

// String returns string representation
func (sca *StocksCommissionAnalysis) String() string {
	return fmt.Sprintf(
		"STOCKS: %.0f shares = $%.2f commission @ $%.4f/share",
		sca.Shares,
		sca.Commission,
		sca.Rate,
	)
}

// DebugString returns detailed debug information
func (sca *StocksCommissionAnalysis) DebugString() string {
	return fmt.Sprintf(
		"STOCKS Commission Analysis:\n"+
			"  Shares:                %.0f\n"+
			"  Commission Per Share:  $%.4f\n"+
			"  Total Commission:      $%.2f",
		sca.Shares,
		sca.Rate,
		sca.Commission,
	)
}

// ==================== INPUT TYPES ====================

// StocksCommissionInput represents input for STOCKS commission calculation
type StocksCommissionInput struct {
	Shares float64 // Number of shares
}
