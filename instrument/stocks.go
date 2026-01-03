package instrument

import "fmt"

// ==================== STOCKS INSTRUMENT ====================

// NewStock creates a STOCKS instrument with standard market parameters
func NewStock(symbol string) *Instrument {
	return &Instrument{
		Symbol:            symbol,
		Type:              TypeStocks,
		Description:       fmt.Sprintf("Stock: %s", symbol),
		Exchange:          "NYSE/NASDAQ",
		DecimalPlaces:     2,
		PipValue:          0.01,
		TickSize:          0.01,
		ContractSize:      1,
		MinimumLotSize:    1.0,
		Commission:        0.001,
		CommissionType:    "percentage",
		Spread:            0.01,
		MaxSpread:         0.05,
		MinSpread:         0.001,
		TradingDays:       252,
		AverageVolume:     1000000,
		TypicalVolatility: 0.25,
		MinVolume:         1.0,
		MaxVolume:         10000.0,
		OpenHour:          13,
		CloseHour:         21,
		IsOpen:            true,
	}
}

// StockDefaults returns default STOCKS market parameters
// Useful for creating multiple STOCKS instruments with consistent settings
func StockDefaults() *Instrument {
	return NewStock("AAPL")
}
