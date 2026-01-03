package commission

import (
	"fmt"
)

// ==================== COMMODITIES COMMISSION CALCULATOR ====================

// CommoditiesCommissionCalculator calculates COMMODITIES commissions
// Commission: $5.00 per lot
// Formula: lots × $5.00
type CommoditiesCommissionCalculator struct {
	// Constants
	CommissionPerLot float64 // $5.00

	// Statistics
	totalCommission float64
	commissionCount int64
	totalLots       float64
}

// ==================== CALCULATOR CREATION ====================

// NewCommoditiesCommissionCalculator creates a new COMMODITIES commission calculator
func NewCommoditiesCommissionCalculator() *CommoditiesCommissionCalculator {
	return &CommoditiesCommissionCalculator{
		CommissionPerLot: 5.00,
	}
}

// ==================== CORE CALCULATION ====================

// CalculateCommission calculates COMMODITIES commission
// Parameters:
//   - lots: Number of lots (e.g., 10 oz of gold = 10 lots)
//
// Returns: Commission in USD
func (ccc *CommoditiesCommissionCalculator) CalculateCommission(
	lots float64,
) (float64, error) {

	// Calculate commission: lots × $5.00
	commission := lots * ccc.CommissionPerLot

	// Track statistics
	ccc.totalCommission += commission
	ccc.commissionCount++
	ccc.totalLots += lots

	return commission, nil
}

// CalculateBatchCommission calculates commission for multiple COMMODITIES trades
func (ccc *CommoditiesCommissionCalculator) CalculateBatchCommission(
	trades []CommoditiesCommissionInput,
) (float64, error) {

	totalCommission := 0.0

	for _, trade := range trades {
		commission, err := ccc.CalculateCommission(trade.Lots)
		if err != nil {
			return 0, err
		}
		totalCommission += commission
	}

	return totalCommission, nil
}

// ==================== STATISTICS ====================

// GetTotalCommission returns total commission collected
func (ccc *CommoditiesCommissionCalculator) GetTotalCommission() float64 {
	return ccc.totalCommission
}

// GetCommissionCount returns number of commissions calculated
func (ccc *CommoditiesCommissionCalculator) GetCommissionCount() int64 {
	return ccc.commissionCount
}

// GetAverageCommission returns average commission per trade
func (ccc *CommoditiesCommissionCalculator) GetAverageCommission() float64 {
	if ccc.commissionCount == 0 {
		return 0
	}
	return ccc.totalCommission / float64(ccc.commissionCount)
}

// GetTotalLots returns total lots traded
func (ccc *CommoditiesCommissionCalculator) GetTotalLots() float64 {
	return ccc.totalLots
}

// GetAverageLots returns average lots per trade
func (ccc *CommoditiesCommissionCalculator) GetAverageLots() float64 {
	if ccc.commissionCount == 0 {
		return 0
	}
	return ccc.totalLots / float64(ccc.commissionCount)
}

// GetStatistics returns comprehensive COMMODITIES commission statistics
func (ccc *CommoditiesCommissionCalculator) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"total_commission":   ccc.totalCommission,
		"commission_count":   ccc.commissionCount,
		"average_commission": ccc.GetAverageCommission(),
		"total_lots":         ccc.totalLots,
		"average_lots":       ccc.GetAverageLots(),
		"commission_per_lot": ccc.CommissionPerLot,
	}
}

// ==================== DEBUG ====================

// String returns a human-readable representation
func (ccc *CommoditiesCommissionCalculator) String() string {
	return fmt.Sprintf(
		"CommoditiesCommission[Total:$%.2f, Count:%d, Lots:%.2f]",
		ccc.totalCommission,
		ccc.commissionCount,
		ccc.totalLots,
	)
}

// DebugString returns detailed debug information
func (ccc *CommoditiesCommissionCalculator) DebugString() string {
	return fmt.Sprintf(
		"COMMODITIES Commission Calculator:\n"+
			"  Total Commission:      $%.2f\n"+
			"  Commission Count:      %d\n"+
			"  Average Commission:    $%.2f\n"+
			"  Total Lots:            %.2f\n"+
			"  Average Lots:          %.2f\n"+
			"  Commission Per Lot:    $%.2f",
		ccc.totalCommission,
		ccc.commissionCount,
		ccc.GetAverageCommission(),
		ccc.totalLots,
		ccc.GetAverageLots(),
		ccc.CommissionPerLot,
	)
}

// Reset resets calculator statistics
func (ccc *CommoditiesCommissionCalculator) Reset() {
	ccc.totalCommission = 0
	ccc.commissionCount = 0
	ccc.totalLots = 0
}

// ==================== ANALYSIS ====================

// AnalyzeCommission provides detailed analysis of a single commission calculation
func (ccc *CommoditiesCommissionCalculator) AnalyzeCommission(
	lots float64,
) *CommoditiesCommissionAnalysis {

	commission := lots * ccc.CommissionPerLot

	return &CommoditiesCommissionAnalysis{
		Lots:       lots,
		Commission: commission,
		Rate:       ccc.CommissionPerLot,
	}
}

// ==================== ANALYSIS TYPES ====================

// CommoditiesCommissionAnalysis provides detailed breakdown of a commission calculation
type CommoditiesCommissionAnalysis struct {
	Lots       float64
	Commission float64
	Rate       float64
}

// String returns string representation
func (cca *CommoditiesCommissionAnalysis) String() string {
	return fmt.Sprintf(
		"COMMODITIES: %.2f lots = $%.2f commission @ $%.2f/lot",
		cca.Lots,
		cca.Commission,
		cca.Rate,
	)
}

// DebugString returns detailed debug information
func (cca *CommoditiesCommissionAnalysis) DebugString() string {
	return fmt.Sprintf(
		"COMMODITIES Commission Analysis:\n"+
			"  Lots:                  %.2f\n"+
			"  Commission Per Lot:    $%.2f\n"+
			"  Total Commission:      $%.2f",
		cca.Lots,
		cca.Rate,
		cca.Commission,
	)
}

// ==================== INPUT TYPES ====================

// CommoditiesCommissionInput represents input for COMMODITIES commission calculation
type CommoditiesCommissionInput struct {
	Lots float64 // Number of lots
}
