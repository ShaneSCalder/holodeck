package types

import (
	"fmt"
)

// ==================== INSTRUMENT CONFIGURATION ====================

// InstrumentConfig defines all parameters for a specific instrument
type InstrumentConfig struct {
	// Type is the instrument type: FOREX, STOCKS, COMMODITIES, CRYPTO
	Type string

	// Symbol is the unique identifier (EURUSD, AAPL, GOLD, BTC/USD)
	Symbol string

	// Description is a human-readable description
	Description string

	// DecimalPlaces is the number of decimal places for display
	// FOREX: 4-5, STOCKS: 2, COMMODITIES: 2-3, CRYPTO: 2-8
	DecimalPlaces int

	// PipValue is the smallest price unit
	// FOREX: 0.0001, STOCKS: 0.01, COMMODITIES: 0.01, CRYPTO: 0.01
	PipValue float64

	// ContractSize is units per lot
	// FOREX: 100000, STOCKS: 1, COMMODITIES: 1, CRYPTO: 1
	ContractSize int64

	// MinimumLotSize is the smallest tradeable amount
	// FOREX: 0.01, STOCKS: 1.0, COMMODITIES: 0.1, CRYPTO: 0.001
	MinimumLotSize float64

	// TickSize is the price increment
	// FOREX: 0.00001, STOCKS: 0.01, COMMODITIES: 0.01, CRYPTO: 1.00
	TickSize float64

	// CommissionType is how commissions are calculated
	// per_million, per_share, per_lot, percentage
	CommissionType string

	// CommissionValue is the commission rate
	// FOREX: 25 (per $1M), STOCKS: 0.01 (per share), etc
	CommissionValue float64

	// SessionHours defines trading hours per session (if applicable)
	SessionHours []SessionHour

	// TradingDaysPerYear is used for return calculations
	TradingDaysPerYear int

	// AverageVolume is typical daily/hourly volume (for slippage calc)
	AverageVolume int64

	// MaxSpread is the maximum typical spread
	MaxSpread float64

	// MinSpread is the minimum typical spread
	MinSpread float64

	// TypicalVolatility is the typical volatility (for slippage)
	TypicalVolatility float64
}

// ==================== SESSION HOUR ====================

// SessionHour defines trading hours for a session
type SessionHour struct {
	// Name of the session (Asian, London, NY, etc)
	Name string

	// OpenHour in UTC (0-23)
	OpenHour int

	// CloseHour in UTC (0-23)
	CloseHour int

	// IsActive indicates if this session is currently open
	IsActive bool
}

// ==================== INSTRUMENT INTERFACE ====================

// Instrument defines the interface all instruments must implement
type Instrument interface {
	// GetType returns the instrument type
	GetType() string

	// GetSymbol returns the instrument symbol
	GetSymbol() string

	// GetDescription returns human-readable description
	GetDescription() string

	// GetDecimalPlaces returns number of decimal places
	GetDecimalPlaces() int

	// GetPipValue returns the pip value
	GetPipValue() float64

	// GetContractSize returns the contract size
	GetContractSize() int64

	// GetMinimumLotSize returns minimum tradeable size
	GetMinimumLotSize() float64

	// GetTickSize returns the price increment
	GetTickSize() float64

	// CalculatePnL calculates profit/loss for a trade
	// Params: entryPrice, exitPrice, size (in lots), direction (1=long, -1=short)
	// Returns: P&L in account currency
	CalculatePnL(entryPrice, exitPrice, size float64, direction int) float64

	// CalculateCommission calculates trading commission
	// Params: price, size (in lots), side (BUY or SELL)
	// Returns: commission in account currency
	CalculateCommission(price, size float64, side string) float64

	// CalculateSlippage calculates expected slippage
	// Params: size (in lots), availableDepth (in units), momentum (0=weak, 1=normal, 2=strong)
	// Returns: slippage in decimal units (pips for forex, cents for stocks, etc)
	CalculateSlippage(size float64, availableDepth int64, momentum int) float64

	// ValidateOrderSize checks if order size is valid
	// Returns error if invalid, nil if valid
	ValidateOrderSize(size float64) error

	// ValidateLimitPrice checks if limit price is valid
	ValidateLimitPrice(limitPrice, currentPrice float64, action string) error

	// FormatPrice formats a price with correct decimals
	FormatPrice(price float64) string

	// GetConfig returns the underlying configuration
	GetConfig() *InstrumentConfig
}

// ==================== INSTRUMENT IMPLEMENTATIONS ====================

// ForexInstrument implements Instrument for FOREX
type ForexInstrument struct {
	config *InstrumentConfig
}

// StocksInstrument implements Instrument for STOCKS
type StocksInstrument struct {
	config *InstrumentConfig
}

// CommoditiesInstrument implements Instrument for COMMODITIES
type CommoditiesInstrument struct {
	config *InstrumentConfig
}

// CryptoInstrument implements Instrument for CRYPTO
type CryptoInstrument struct {
	config *InstrumentConfig
}

// ==================== FACTORY FUNCTION ====================

// NewInstrument creates an appropriate instrument based on type
func NewInstrument(instrumentType, symbol, description string) (Instrument, error) {
	if !IsValidInstrumentType(instrumentType) {
		return nil, NewInvalidInstrumentTypeError(instrumentType)
	}

	switch instrumentType {
	case InstrumentTypeForex:
		return NewForexInstrument(symbol, description), nil

	case InstrumentTypeStocks:
		return NewStocksInstrument(symbol, description), nil

	case InstrumentTypeCommodities:
		return NewCommoditiesInstrument(symbol, description), nil

	case InstrumentTypeCrypto:
		return NewCryptoInstrument(symbol, description), nil

	default:
		return nil, NewInstrumentNotFoundError(instrumentType)
	}
}

// ==================== FOREX IMPLEMENTATION ====================

// NewForexInstrument creates a new Forex instrument
func NewForexInstrument(symbol, description string) *ForexInstrument {
	return &ForexInstrument{
		config: &InstrumentConfig{
			Type:               InstrumentTypeForex,
			Symbol:             symbol,
			Description:        description,
			DecimalPlaces:      ForexDecimalPlaces,
			PipValue:           ForexPipValue,
			ContractSize:       int64(ForexContractSize),
			MinimumLotSize:     ForexMinimumLotSize,
			TickSize:           ForexTickSize,
			CommissionType:     ForexCommissionType,
			CommissionValue:    ForexCommissionValue,
			TradingDaysPerYear: 250,
			TypicalVolatility:  0.01, // 1%
		},
	}
}

func (f *ForexInstrument) GetType() string              { return f.config.Type }
func (f *ForexInstrument) GetSymbol() string            { return f.config.Symbol }
func (f *ForexInstrument) GetDescription() string       { return f.config.Description }
func (f *ForexInstrument) GetDecimalPlaces() int        { return f.config.DecimalPlaces }
func (f *ForexInstrument) GetPipValue() float64         { return f.config.PipValue }
func (f *ForexInstrument) GetContractSize() int64       { return f.config.ContractSize }
func (f *ForexInstrument) GetMinimumLotSize() float64   { return f.config.MinimumLotSize }
func (f *ForexInstrument) GetTickSize() float64         { return f.config.TickSize }
func (f *ForexInstrument) GetConfig() *InstrumentConfig { return f.config }

func (f *ForexInstrument) CalculatePnL(entryPrice, exitPrice, size float64, direction int) float64 {
	priceDiff := (exitPrice - entryPrice) * float64(direction)
	pips := priceDiff / f.config.PipValue
	return pips * size * float64(f.config.ContractSize) * f.config.PipValue
}

func (f *ForexInstrument) CalculateCommission(price, size float64, side string) float64 {
	notional := price * size * float64(f.config.ContractSize)
	if f.config.CommissionType == CommissionTypePerMillion {
		return (notional / 1000000.0) * f.config.CommissionValue
	}
	return 0
}

func (f *ForexInstrument) CalculateSlippage(size float64, availableDepth int64, momentum int) float64 {
	if availableDepth == 0 {
		return 0
	}

	baseSlippage := (size / float64(availableDepth)) * f.config.TypicalVolatility
	multiplier := GetMomentumMultiplier([]string{MomentumWeak, MomentumNormal, MomentumStrong}[momentum])
	return baseSlippage * float64(f.config.ContractSize) * multiplier
}

func (f *ForexInstrument) ValidateOrderSize(size float64) error {
	if size < f.config.MinimumLotSize {
		return NewInvalidLotSizeError(size, f.config.MinimumLotSize)
	}
	return nil
}

func (f *ForexInstrument) ValidateLimitPrice(limitPrice, currentPrice float64, action string) error {
	if limitPrice <= 0 {
		return NewInvalidLimitPriceError(limitPrice, "price must be positive")
	}
	return nil
}

func (f *ForexInstrument) FormatPrice(price float64) string {
	format := fmt.Sprintf("%%.%df", f.config.DecimalPlaces)
	return fmt.Sprintf(format, price)
}

// ==================== STOCKS IMPLEMENTATION ====================

// NewStocksInstrument creates a new Stocks instrument
func NewStocksInstrument(symbol, description string) *StocksInstrument {
	return &StocksInstrument{
		config: &InstrumentConfig{
			Type:               InstrumentTypeStocks,
			Symbol:             symbol,
			Description:        description,
			DecimalPlaces:      StocksDecimalPlaces,
			PipValue:           StocksPipValue,
			ContractSize:       int64(StocksContractSize),
			MinimumLotSize:     StocksMinimumLotSize,
			TickSize:           StocksTickSize,
			CommissionType:     StocksCommissionType,
			CommissionValue:    StocksCommissionValue,
			TradingDaysPerYear: 252,
			TypicalVolatility:  0.02, // 2%
		},
	}
}

func (s *StocksInstrument) GetType() string              { return s.config.Type }
func (s *StocksInstrument) GetSymbol() string            { return s.config.Symbol }
func (s *StocksInstrument) GetDescription() string       { return s.config.Description }
func (s *StocksInstrument) GetDecimalPlaces() int        { return s.config.DecimalPlaces }
func (s *StocksInstrument) GetPipValue() float64         { return s.config.PipValue }
func (s *StocksInstrument) GetContractSize() int64       { return s.config.ContractSize }
func (s *StocksInstrument) GetMinimumLotSize() float64   { return s.config.MinimumLotSize }
func (s *StocksInstrument) GetTickSize() float64         { return s.config.TickSize }
func (s *StocksInstrument) GetConfig() *InstrumentConfig { return s.config }

func (s *StocksInstrument) CalculatePnL(entryPrice, exitPrice, size float64, direction int) float64 {
	priceDiff := (exitPrice - entryPrice) * float64(direction)
	return priceDiff * size
}

func (s *StocksInstrument) CalculateCommission(price, size float64, side string) float64 {
	if s.config.CommissionType == CommissionTypePerShare {
		return size * s.config.CommissionValue
	}
	return 0
}

func (s *StocksInstrument) CalculateSlippage(size float64, availableDepth int64, momentum int) float64 {
	if availableDepth == 0 {
		return 0
	}

	baseSlippage := (size / float64(availableDepth)) * s.config.TypicalVolatility
	multiplier := GetMomentumMultiplier([]string{MomentumWeak, MomentumNormal, MomentumStrong}[momentum])
	return baseSlippage * multiplier
}

func (s *StocksInstrument) ValidateOrderSize(size float64) error {
	if size < s.config.MinimumLotSize {
		return NewInvalidLotSizeError(size, s.config.MinimumLotSize)
	}
	return nil
}

func (s *StocksInstrument) ValidateLimitPrice(limitPrice, currentPrice float64, action string) error {
	if limitPrice <= 0 {
		return NewInvalidLimitPriceError(limitPrice, "price must be positive")
	}
	return nil
}

func (s *StocksInstrument) FormatPrice(price float64) string {
	format := fmt.Sprintf("%%.%df", s.config.DecimalPlaces)
	return fmt.Sprintf(format, price)
}

// ==================== COMMODITIES IMPLEMENTATION ====================

// NewCommoditiesInstrument creates a new Commodities instrument
func NewCommoditiesInstrument(symbol, description string) *CommoditiesInstrument {
	return &CommoditiesInstrument{
		config: &InstrumentConfig{
			Type:               InstrumentTypeCommodities,
			Symbol:             symbol,
			Description:        description,
			DecimalPlaces:      CommoditiesDecimalPlaces,
			PipValue:           CommoditiesPipValue,
			ContractSize:       int64(CommoditiesContractSize),
			MinimumLotSize:     CommoditiesMinimumLotSize,
			TickSize:           CommoditiesTickSize,
			CommissionType:     CommoditiesCommissionType,
			CommissionValue:    CommoditiesCommissionValue,
			TradingDaysPerYear: 250,
			TypicalVolatility:  0.015, // 1.5%
		},
	}
}

func (c *CommoditiesInstrument) GetType() string              { return c.config.Type }
func (c *CommoditiesInstrument) GetSymbol() string            { return c.config.Symbol }
func (c *CommoditiesInstrument) GetDescription() string       { return c.config.Description }
func (c *CommoditiesInstrument) GetDecimalPlaces() int        { return c.config.DecimalPlaces }
func (c *CommoditiesInstrument) GetPipValue() float64         { return c.config.PipValue }
func (c *CommoditiesInstrument) GetContractSize() int64       { return c.config.ContractSize }
func (c *CommoditiesInstrument) GetMinimumLotSize() float64   { return c.config.MinimumLotSize }
func (c *CommoditiesInstrument) GetTickSize() float64         { return c.config.TickSize }
func (c *CommoditiesInstrument) GetConfig() *InstrumentConfig { return c.config }

func (c *CommoditiesInstrument) CalculatePnL(entryPrice, exitPrice, size float64, direction int) float64 {
	priceDiff := (exitPrice - entryPrice) * float64(direction)
	return priceDiff * size
}

func (c *CommoditiesInstrument) CalculateCommission(price, size float64, side string) float64 {
	if c.config.CommissionType == CommissionTypePerLot {
		return size * c.config.CommissionValue
	}
	return 0
}

func (c *CommoditiesInstrument) CalculateSlippage(size float64, availableDepth int64, momentum int) float64 {
	if availableDepth == 0 {
		return 0
	}

	baseSlippage := (size / float64(availableDepth)) * c.config.TypicalVolatility
	multiplier := GetMomentumMultiplier([]string{MomentumWeak, MomentumNormal, MomentumStrong}[momentum])
	return baseSlippage * multiplier
}

func (c *CommoditiesInstrument) ValidateOrderSize(size float64) error {
	if size < c.config.MinimumLotSize {
		return NewInvalidLotSizeError(size, c.config.MinimumLotSize)
	}
	return nil
}

func (c *CommoditiesInstrument) ValidateLimitPrice(limitPrice, currentPrice float64, action string) error {
	if limitPrice <= 0 {
		return NewInvalidLimitPriceError(limitPrice, "price must be positive")
	}
	return nil
}

func (c *CommoditiesInstrument) FormatPrice(price float64) string {
	format := fmt.Sprintf("%%.%df", c.config.DecimalPlaces)
	return fmt.Sprintf(format, price)
}

// ==================== CRYPTO IMPLEMENTATION ====================

// NewCryptoInstrument creates a new Crypto instrument
func NewCryptoInstrument(symbol, description string) *CryptoInstrument {
	return &CryptoInstrument{
		config: &InstrumentConfig{
			Type:               InstrumentTypeCrypto,
			Symbol:             symbol,
			Description:        description,
			DecimalPlaces:      CryptoDecimalPlaces,
			PipValue:           CryptoPipValue,
			ContractSize:       int64(CryptoContractSize),
			MinimumLotSize:     CryptoMinimumLotSize,
			TickSize:           CryptoTickSize,
			CommissionType:     CryptoCommissionType,
			CommissionValue:    CryptoCommissionValue,
			TradingDaysPerYear: 365,
			TypicalVolatility:  0.03, // 3%
		},
	}
}

func (cr *CryptoInstrument) GetType() string              { return cr.config.Type }
func (cr *CryptoInstrument) GetSymbol() string            { return cr.config.Symbol }
func (cr *CryptoInstrument) GetDescription() string       { return cr.config.Description }
func (cr *CryptoInstrument) GetDecimalPlaces() int        { return cr.config.DecimalPlaces }
func (cr *CryptoInstrument) GetPipValue() float64         { return cr.config.PipValue }
func (cr *CryptoInstrument) GetContractSize() int64       { return cr.config.ContractSize }
func (cr *CryptoInstrument) GetMinimumLotSize() float64   { return cr.config.MinimumLotSize }
func (cr *CryptoInstrument) GetTickSize() float64         { return cr.config.TickSize }
func (cr *CryptoInstrument) GetConfig() *InstrumentConfig { return cr.config }

func (cr *CryptoInstrument) CalculatePnL(entryPrice, exitPrice, size float64, direction int) float64 {
	priceDiff := (exitPrice - entryPrice) * float64(direction)
	return priceDiff * size
}

func (cr *CryptoInstrument) CalculateCommission(price, size float64, side string) float64 {
	if cr.config.CommissionType == CommissionTypePercentage {
		notional := price * size
		return notional * cr.config.CommissionValue
	}
	return 0
}

func (cr *CryptoInstrument) CalculateSlippage(size float64, availableDepth int64, momentum int) float64 {
	if availableDepth == 0 {
		return 0
	}

	baseSlippage := (size / float64(availableDepth)) * cr.config.TypicalVolatility
	multiplier := GetMomentumMultiplier([]string{MomentumWeak, MomentumNormal, MomentumStrong}[momentum])
	return baseSlippage * multiplier
}

func (cr *CryptoInstrument) ValidateOrderSize(size float64) error {
	if size < cr.config.MinimumLotSize {
		return NewInvalidLotSizeError(size, cr.config.MinimumLotSize)
	}
	return nil
}

func (cr *CryptoInstrument) ValidateLimitPrice(limitPrice, currentPrice float64, action string) error {
	if limitPrice <= 0 {
		return NewInvalidLimitPriceError(limitPrice, "price must be positive")
	}
	return nil
}

func (cr *CryptoInstrument) FormatPrice(price float64) string {
	format := fmt.Sprintf("%%.%df", cr.config.DecimalPlaces)
	return fmt.Sprintf(format, price)
}

// ==================== REGISTRY ====================

// InstrumentRegistry manages available instruments
type InstrumentRegistry struct {
	instruments map[string]Instrument
}

// NewInstrumentRegistry creates a new registry
func NewInstrumentRegistry() *InstrumentRegistry {
	return &InstrumentRegistry{
		instruments: make(map[string]Instrument),
	}
}

// Register registers an instrument
func (ir *InstrumentRegistry) Register(symbol string, instrument Instrument) {
	ir.instruments[symbol] = instrument
}

// Get retrieves an instrument
func (ir *InstrumentRegistry) Get(symbol string) (Instrument, bool) {
	inst, ok := ir.instruments[symbol]
	return inst, ok
}

// GetAll returns all registered instruments
func (ir *InstrumentRegistry) GetAll() map[string]Instrument {
	return ir.instruments
}

// Size returns the number of registered instruments
func (ir *InstrumentRegistry) Size() int {
	return len(ir.instruments)
}

// List returns a list of all symbol names
func (ir *InstrumentRegistry) List() []string {
	symbols := make([]string, 0, len(ir.instruments))
	for symbol := range ir.instruments {
		symbols = append(symbols, symbol)
	}
	return symbols
}

// ==================== COMPARISON ====================

// CompareInstruments compares two instruments
func CompareInstruments(a, b Instrument) map[string]interface{} {
	return map[string]interface{}{
		"both_forex":       a.GetType() == InstrumentTypeForex && b.GetType() == InstrumentTypeForex,
		"both_stocks":      a.GetType() == InstrumentTypeStocks && b.GetType() == InstrumentTypeStocks,
		"both_commodities": a.GetType() == InstrumentTypeCommodities && b.GetType() == InstrumentTypeCommodities,
		"both_crypto":      a.GetType() == InstrumentTypeCrypto && b.GetType() == InstrumentTypeCrypto,
		"pip_value_a":      a.GetPipValue(),
		"pip_value_b":      b.GetPipValue(),
		"contract_size_a":  a.GetContractSize(),
		"contract_size_b":  b.GetContractSize(),
		"min_lot_size_a":   a.GetMinimumLotSize(),
		"min_lot_size_b":   b.GetMinimumLotSize(),
	}
}
