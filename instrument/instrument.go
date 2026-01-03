package instrument

// ==================== INSTRUMENT PACKAGE ====================

// This file serves as the main entry point for the instrument package
// and provides convenience functions for working with multiple instruments

// GetInstrumentType returns the type string for an instrument
func GetInstrumentType(instrument *Instrument) string {
	return instrument.Type
}

// IsValidInstrument checks if an instrument is properly configured
func IsValidInstrument(instrument *Instrument) bool {
	if instrument == nil {
		return false
	}
	if instrument.Symbol == "" {
		return false
	}
	if instrument.Type == "" {
		return false
	}
	if instrument.PipValue <= 0 {
		return false
	}
	if instrument.ContractSize <= 0 {
		return false
	}
	return true
}

// CompareInstruments compares two instruments for equality
func CompareInstruments(a, b *Instrument) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Symbol == b.Symbol && a.Type == b.Type && a.Exchange == b.Exchange
}

// CreateCustomInstrument creates a custom instrument with provided parameters
func CreateCustomInstrument(symbol string, instrumentType string, decimals int,
	pipValue float64, tickSize float64, contractSize int64, minLot float64) *Instrument {
	return &Instrument{
		Symbol:            symbol,
		Type:              instrumentType,
		DecimalPlaces:     decimals,
		PipValue:          pipValue,
		TickSize:          tickSize,
		ContractSize:      contractSize,
		MinimumLotSize:    minLot,
		Commission:        0.001,
		CommissionType:    "percentage",
		Spread:            0.0001,
		MaxSpread:         0.001,
		MinSpread:         0.00001,
		TradingDays:       252,
		AverageVolume:     1000000,
		TypicalVolatility: 0.20,
		MinVolume:         0.01,
		MaxVolume:         100000.0,
		OpenHour:          0,
		CloseHour:         24,
		IsOpen:            true,
	}
}
