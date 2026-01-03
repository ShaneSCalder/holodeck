package instrument

import (
	"fmt"
	"strings"
)

// ==================== INSTRUMENT TYPES ====================

const (
	TypeForex       = "FOREX"
	TypeStocks      = "STOCKS"
	TypeCommodities = "COMMODITIES"
	TypeCrypto      = "CRYPTO"
)

// ==================== INSTRUMENT ====================

// Instrument represents a tradeable financial instrument
type Instrument struct {
	// Identity
	Symbol      string
	Type        string
	Description string
	Exchange    string

	// Price Configuration
	DecimalPlaces  int
	PipValue       float64
	TickSize       float64
	ContractSize   int64
	MinimumLotSize float64

	// Trading Parameters
	Commission        float64
	CommissionType    string // per_million, per_share, per_lot, percentage
	Spread            float64
	MaxSpread         float64
	MinSpread         float64
	TradingDays       int
	AverageVolume     int64
	TypicalVolatility float64

	// Restrictions
	MinVolume float64
	MaxVolume float64
	MinPrice  float64
	MaxPrice  float64

	// Session Info
	OpenHour  int
	CloseHour int
	IsOpen    bool
}

// ==================== PRICE HELPERS ====================

// RoundPrice rounds price to pip value
func (i *Instrument) RoundPrice(price float64) float64 {
	if i.TickSize == 0 {
		return price
	}
	return round(price/i.TickSize) * i.TickSize
}

// FormatPrice formats price to decimal places
func (i *Instrument) FormatPrice(price float64) string {
	return fmt.Sprintf("%.*f", i.DecimalPlaces, price)
}

// NormalizeLot normalizes lot size to minimum lot
func (i *Instrument) NormalizeLot(lot float64) float64 {
	if lot == 0 {
		return 0
	}
	return round(lot/i.MinimumLotSize) * i.MinimumLotSize
}

// ==================== VALIDATION ====================

// IsValidVolume checks if volume is within limits
func (i *Instrument) IsValidVolume(volume float64) bool {
	return volume >= i.MinVolume && volume <= i.MaxVolume
}

// IsValidPrice checks if price is within limits
func (i *Instrument) IsValidPrice(price float64) bool {
	if i.MinPrice > 0 && price < i.MinPrice {
		return false
	}
	if i.MaxPrice > 0 && price > i.MaxPrice {
		return false
	}
	return true
}

// ==================== TYPE CHECKS ====================

// IsForex checks if instrument is FOREX
func (i *Instrument) IsForex() bool {
	return i.Type == TypeForex
}

// IsStock checks if instrument is STOCKS
func (i *Instrument) IsStock() bool {
	return i.Type == TypeStocks
}

// IsCommodity checks if instrument is COMMODITIES
func (i *Instrument) IsCommodity() bool {
	return i.Type == TypeCommodities
}

// IsCrypto checks if instrument is CRYPTO
func (i *Instrument) IsCrypto() bool {
	return i.Type == TypeCrypto
}

// ==================== STATISTICS ====================

// GetVolatilityCategory returns volatility category
func (i *Instrument) GetVolatilityCategory() string {
	switch {
	case i.TypicalVolatility < 0.10:
		return "LOW"
	case i.TypicalVolatility < 0.20:
		return "MEDIUM"
	case i.TypicalVolatility < 0.40:
		return "HIGH"
	default:
		return "VERY_HIGH"
	}
}

// GetLiquidityCategory returns liquidity category based on volume
func (i *Instrument) GetLiquidityCategory() string {
	switch {
	case i.AverageVolume > 5000000:
		return "VERY_HIGH"
	case i.AverageVolume > 1000000:
		return "HIGH"
	case i.AverageVolume > 100000:
		return "MEDIUM"
	default:
		return "LOW"
	}
}

// ==================== POSITION SIZING ====================

// GetRiskAmount calculates risk amount per pip
func (i *Instrument) GetRiskAmount(lotSize float64) float64 {
	return lotSize * float64(i.ContractSize) * i.PipValue
}

// GetRequiredMargin calculates required margin (simplified)
func (i *Instrument) GetRequiredMargin(lotSize float64, leverage float64, price float64) float64 {
	notional := lotSize * float64(i.ContractSize) * price
	return notional / leverage
}

// ==================== STRING REPRESENTATION ====================

// String returns formatted instrument info
func (i *Instrument) String() string {
	return fmt.Sprintf(
		"%s (%s) | Pip: %.6f | Spread: %.6f | Vol: %.0f | Volatility: %s",
		i.Symbol, i.Type, i.PipValue, i.Spread, float64(i.AverageVolume),
		i.GetVolatilityCategory(),
	)
}

// Details returns detailed instrument information
func (i *Instrument) Details() string {
	return fmt.Sprintf(
		"Symbol:         %s\n"+
			"Type:           %s\n"+
			"Description:    %s\n"+
			"Exchange:       %s\n"+
			"Decimals:       %d\n"+
			"Pip Value:      %.6f\n"+
			"Tick Size:      %.6f\n"+
			"Contract Size:  %d\n"+
			"Min Lot:        %.3f\n"+
			"Commission:     %.6f (%s)\n"+
			"Spread:         %.6f (min: %.6f, max: %.6f)\n"+
			"Volume:         %.0f\n"+
			"Volatility:     %.2f%% (%s)\n"+
			"Hours:          %02d:00 - %02d:00 UTC",
		i.Symbol, i.Type, i.Description, i.Exchange, i.DecimalPlaces,
		i.PipValue, i.TickSize, i.ContractSize, i.MinimumLotSize,
		i.Commission, i.CommissionType,
		i.Spread, i.MinSpread, i.MaxSpread,
		float64(i.AverageVolume),
		i.TypicalVolatility*100, i.GetVolatilityCategory(),
		i.OpenHour, i.CloseHour,
	)
}

// ==================== UTILITY ====================

// round rounds to nearest integer
func round(x float64) float64 {
	return float64(int64(x + 0.5))
}

// ==================== INSTRUMENT LIST ====================

// InstrumentList manages multiple instruments
type InstrumentList struct {
	instruments map[string]*Instrument
}

// NewInstrumentList creates a new instrument list
func NewInstrumentList() *InstrumentList {
	return &InstrumentList{
		instruments: make(map[string]*Instrument),
	}
}

// Add adds an instrument to the list
func (il *InstrumentList) Add(instrument *Instrument) {
	il.instruments[strings.ToUpper(instrument.Symbol)] = instrument
}

// Get retrieves an instrument by symbol
func (il *InstrumentList) Get(symbol string) (*Instrument, bool) {
	inst, ok := il.instruments[strings.ToUpper(symbol)]
	return inst, ok
}

// Remove removes an instrument from the list
func (il *InstrumentList) Remove(symbol string) {
	delete(il.instruments, strings.ToUpper(symbol))
}

// List returns all instruments
func (il *InstrumentList) List() []*Instrument {
	instruments := make([]*Instrument, 0, len(il.instruments))
	for _, inst := range il.instruments {
		instruments = append(instruments, inst)
	}
	return instruments
}

// Count returns number of instruments
func (il *InstrumentList) Count() int {
	return len(il.instruments)
}

// Contains checks if instrument exists
func (il *InstrumentList) Contains(symbol string) bool {
	_, ok := il.instruments[strings.ToUpper(symbol)]
	return ok
}
