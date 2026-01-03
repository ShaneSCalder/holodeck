package instrument

import "fmt"

// ==================== CRYPTO INSTRUMENT ====================

// NewCrypto creates a CRYPTO instrument with standard market parameters
func NewCrypto(symbol string) *Instrument {
	return &Instrument{
		Symbol:            symbol,
		Type:              TypeCrypto,
		Description:       fmt.Sprintf("Cryptocurrency: %s", symbol),
		Exchange:          "CRYPTO",
		DecimalPlaces:     8,
		PipValue:          0.00000001,
		TickSize:          0.00000001,
		ContractSize:      1,
		MinimumLotSize:    0.001,
		Commission:        0.001,
		CommissionType:    "percentage",
		Spread:            0.0001,
		MaxSpread:         0.001,
		MinSpread:         0.00001,
		TradingDays:       365,
		AverageVolume:     1000000,
		TypicalVolatility: 0.50,
		MinVolume:         0.001,
		MaxVolume:         1000.0,
		OpenHour:          0,
		CloseHour:         24,
		IsOpen:            true,
	}
}

// CryptoDefaults returns default CRYPTO market parameters
// Useful for creating multiple CRYPTO instruments with consistent settings
func CryptoDefaults() *Instrument {
	return NewCrypto("BTC/USD")
}
