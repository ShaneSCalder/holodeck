package instrument

import "fmt"

// ==================== COMMODITIES INSTRUMENT ====================

// NewCommodity creates a COMMODITIES instrument with standard market parameters
func NewCommodity(symbol string) *Instrument {
	return &Instrument{
		Symbol:            symbol,
		Type:              TypeCommodities,
		Description:       fmt.Sprintf("Commodity: %s", symbol),
		Exchange:          "COMEX/NYMEX",
		DecimalPlaces:     3,
		PipValue:          0.01,
		TickSize:          0.01,
		ContractSize:      100,
		MinimumLotSize:    0.1,
		Commission:        50,
		CommissionType:    "per_lot",
		Spread:            0.02,
		MaxSpread:         0.10,
		MinSpread:         0.01,
		TradingDays:       252,
		AverageVolume:     500000,
		TypicalVolatility: 0.18,
		MinVolume:         0.1,
		MaxVolume:         100.0,
		OpenHour:          0,
		CloseHour:         24,
		IsOpen:            true,
	}
}

// CommodityDefaults returns default COMMODITIES market parameters
// Useful for creating multiple COMMODITIES instruments with consistent settings
func CommodityDefaults() *Instrument {
	return NewCommodity("GOLD")
}
