package instrument

import "fmt"

// ==================== FOREX INSTRUMENT ====================

// NewForex creates a FOREX instrument with standard market parameters
func NewForex(symbol string) *Instrument {
	return &Instrument{
		Symbol:            symbol,
		Type:              TypeForex,
		Description:       fmt.Sprintf("Foreign Exchange: %s", symbol),
		Exchange:          "FOREX",
		DecimalPlaces:     5,
		PipValue:          0.0001,
		TickSize:          0.00001,
		ContractSize:      100000,
		MinimumLotSize:    0.01,
		Commission:        25,
		CommissionType:    "per_million",
		Spread:            0.0002,
		MaxSpread:         0.0005,
		MinSpread:         0.00005,
		TradingDays:       252,
		AverageVolume:     1000000,
		TypicalVolatility: 0.10,
		MinVolume:         0.01,
		MaxVolume:         1000.0,
		OpenHour:          0,
		CloseHour:         24,
		IsOpen:            true,
	}
}

// ForexDefaults returns default FOREX market parameters
// Useful for creating multiple FOREX instruments with consistent settings
func ForexDefaults() *Instrument {
	return NewForex("EUR/USD")
}
