package commission

import (
	"fmt"

	"holodeck/types"
)

// ==================== COMMISSION CALCULATOR ====================

// CommissionCalculator orchestrates commission calculation for all instrument types
type CommissionCalculator struct {
	forexCalc       *ForexCommissionCalculator
	stocksCalc      *StocksCommissionCalculator
	commoditiesCalc *CommoditiesCommissionCalculator
	cryptoCalc      *CryptoCommissionCalculator

	// Statistics
	totalCommission float64
	commissionCount int64
}

// ==================== CALCULATOR CREATION ====================

// NewCommissionCalculator creates a new commission calculator
func NewCommissionCalculator() *CommissionCalculator {
	return &CommissionCalculator{
		forexCalc:       NewForexCommissionCalculator(),
		stocksCalc:      NewStocksCommissionCalculator(),
		commoditiesCalc: NewCommoditiesCommissionCalculator(),
		cryptoCalc:      NewCryptoCommissionCalculator(),
	}
}

// ==================== CORE CALCULATION ====================

// CalculateCommission calculates commission based on instrument type
func (cc *CommissionCalculator) CalculateCommission(
	price float64,
	size float64,
	instrument types.Instrument,
	side string, // BUY or SELL
) (float64, error) {

	if instrument == nil {
		return 0, types.NewOrderRejectedError("instrument cannot be nil")
	}

	var commission float64
	var err error

	instType := instrument.GetInstrumentType()
	switch instType {
	case types.InstrumentTypeForex:
		commission, err = cc.forexCalc.CalculateCommission(price, size)

	case types.InstrumentTypeStocks:
		commission, err = cc.stocksCalc.CalculateCommission(size)

	case types.InstrumentTypeCommodities:
		commission, err = cc.commoditiesCalc.CalculateCommission(size)

	case types.InstrumentTypeCrypto:
		commission, err = cc.cryptoCalc.CalculateCommission(price, size)

	default:
		return 0, types.NewOrderRejectedError("unsupported instrument type")
	}

	if err != nil {
		return 0, err
	}

	// Track statistics
	cc.totalCommission += commission
	cc.commissionCount++

	return commission, nil
}

// CalculateBatchCommission calculates commission for multiple trades
func (cc *CommissionCalculator) CalculateBatchCommission(
	trades []CommissionInput,
	instrument types.Instrument,
) (float64, error) {

	if instrument == nil {
		return 0, types.NewOrderRejectedError("instrument cannot be nil")
	}

	totalCommission := 0.0

	for _, trade := range trades {
		commission, err := cc.CalculateCommission(
			trade.Price,
			trade.Size,
			instrument,
			trade.Side,
		)
		if err != nil {
			return 0, err
		}
		totalCommission += commission
	}

	return totalCommission, nil
}

// ==================== STATISTICS ====================

// GetTotalCommission returns total commission collected
func (cc *CommissionCalculator) GetTotalCommission() float64 {
	return cc.totalCommission
}

// GetCommissionCount returns number of commissions calculated
func (cc *CommissionCalculator) GetCommissionCount() int64 {
	return cc.commissionCount
}

// GetAverageCommission returns average commission per trade
func (cc *CommissionCalculator) GetAverageCommission() float64 {
	if cc.commissionCount == 0 {
		return 0
	}
	return cc.totalCommission / float64(cc.commissionCount)
}

// GetStatistics returns comprehensive commission statistics
func (cc *CommissionCalculator) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"total_commission":   cc.totalCommission,
		"commission_count":   cc.commissionCount,
		"average_commission": cc.GetAverageCommission(),
		"forex_stats":        cc.forexCalc.GetStatistics(),
		"stocks_stats":       cc.stocksCalc.GetStatistics(),
		"commodities_stats":  cc.commoditiesCalc.GetStatistics(),
		"crypto_stats":       cc.cryptoCalc.GetStatistics(),
	}
}

// ==================== DEBUG ====================

// String returns a human-readable representation
func (cc *CommissionCalculator) String() string {
	return fmt.Sprintf(
		"CommissionCalculator[Total:$%.2f, Count:%d, Avg:$%.2f]",
		cc.totalCommission,
		cc.commissionCount,
		cc.GetAverageCommission(),
	)
}

// DebugString returns detailed debug information
func (cc *CommissionCalculator) DebugString() string {
	return fmt.Sprintf(
		"Commission Calculator:\n"+
			"  Total Commission:      $%.2f\n"+
			"  Commission Count:      %d\n"+
			"  Average Commission:    $%.2f\n"+
			"\n"+
			"  Sub-calculators:\n"+
			"    Forex:               %s\n"+
			"    Stocks:              %s\n"+
			"    Commodities:         %s\n"+
			"    Crypto:              %s",
		cc.totalCommission,
		cc.commissionCount,
		cc.GetAverageCommission(),
		cc.forexCalc.String(),
		cc.stocksCalc.String(),
		cc.commoditiesCalc.String(),
		cc.cryptoCalc.String(),
	)
}

// Reset resets calculator statistics
func (cc *CommissionCalculator) Reset() {
	cc.totalCommission = 0
	cc.commissionCount = 0
	cc.forexCalc.Reset()
	cc.stocksCalc.Reset()
	cc.commoditiesCalc.Reset()
	cc.cryptoCalc.Reset()
}

// ==================== COMMISSION INPUT ====================

// CommissionInput represents input for batch commission calculation
type CommissionInput struct {
	Price float64 // Price per unit
	Size  float64 // Size in units (lots, shares, contracts, etc.)
	Side  string  // BUY or SELL
}
