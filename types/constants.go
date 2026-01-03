package types

// ==================== INSTRUMENT TYPES ====================

const (
	InstrumentTypeForex       = "FOREX"
	InstrumentTypeStocks      = "STOCKS"
	InstrumentTypeCommodities = "COMMODITIES"
	InstrumentTypeCrypto      = "CRYPTO"
)

// ==================== ORDER ACTIONS ====================

const (
	OrderActionBuy  = "BUY"
	OrderActionSell = "SELL"
	OrderActionHold = "HOLD"
)

// ==================== ORDER TYPES ====================

const (
	OrderTypeMarket = "MARKET"
	OrderTypeLimit  = "LIMIT"
)

// ==================== ORDER STATUS ====================

const (
	OrderStatusFilled    = "FILLED"
	OrderStatusPartial   = "PARTIAL"
	OrderStatusRejected  = "REJECTED"
	OrderStatusPending   = "PENDING"
	OrderStatusCancelled = "CANCELLED"
)

// ==================== ACCOUNT STATUS ====================

const (
	AccountStatusActive  = "ACTIVE"
	AccountStatusBlown   = "BLOWN"
	AccountStatusAtLimit = "AT_LIMIT"
)

// ==================== POSITION STATUS ====================

const (
	PositionStatusFlat  = "FLAT"
	PositionStatusLong  = "LONG"
	PositionStatusShort = "SHORT"
)

// ==================== ERROR CODES ====================

const (
	ErrorCodeInsufficientBalance   = "INSUFFICIENT_BALANCE"
	ErrorCodePositionLimitExceeded = "POSITION_LIMIT_EXCEEDED"
	ErrorCodeInvalidOrderType      = "INVALID_ORDER_TYPE"
	ErrorCodeInvalidLimitPrice     = "INVALID_LIMIT_PRICE"
	ErrorCodeInvalidOrderSize      = "INVALID_ORDER_SIZE"
	ErrorCodeOrderRejected         = "ORDER_REJECTED"
	ErrorCodeAccountBlown          = "ACCOUNT_BLOWN"
	ErrorCodeInvalidOperation      = "INVALID_OPERATION"
	ErrorCodeInvalidLotSize        = "INVALID_LOT_SIZE"
	ErrorCodeCSVReadError          = "CSV_READ_ERROR"
	ErrorCodeConfigError           = "CONFIG_ERROR"
	ErrorCodeInstrumentNotFound    = "INSTRUMENT_NOT_FOUND"
	ErrorCodeInvalidInstrumentType = "INVALID_INSTRUMENT_TYPE"
)

// ==================== COMMISSION TYPES ====================

const (
	CommissionTypePerMillion = "per_million"
	CommissionTypePerShare   = "per_share"
	CommissionTypePerLot     = "per_lot"
	CommissionTypePercentage = "percentage"
)

// ==================== SLIPPAGE MODELS ====================

const (
	SlippageModelDepth    = "depth"
	SlippageModelMomentum = "momentum"
	SlippageModelFixed    = "fixed"
	SlippageModelNone     = "none"
)

// ==================== MOMENTUM LEVELS ====================

const (
	MomentumStrong = "STRONG"
	MomentumNormal = "NORMAL"
	MomentumWeak   = "WEAK"
)

// ==================== PARTIAL FILL LOGIC ====================

const (
	PartialFillByVolumeMomentum = "volume_momentum"
	PartialFillByDepth          = "depth"
	PartialFillNone             = "none"
)

// ==================== LOGGING LEVELS ====================

const (
	LogLevelDebug   = "DEBUG"
	LogLevelInfo    = "INFO"
	LogLevelWarning = "WARNING"
	LogLevelError   = "ERROR"
	LogLevelFatal   = "FATAL"
)

// ==================== DEFAULT VALUES ====================

const (
	DefaultInitialBalance     = 100000.00
	DefaultLeverage           = 1.0
	DefaultMaxDrawdownPercent = 20.0
	DefaultSpeedMultiplier    = 100.0
	DefaultLatencyMs          = 5
)

// ==================== FOREX CONSTANTS ====================

const (
	ForexContractSize   = 100000
	ForexMinimumLotSize = 0.01
	ForexDecimalPlaces  = 4
	ForexPipValue       = 0.0001
	ForexTickSize       = 0.00001
)

// ==================== STOCKS CONSTANTS ====================

const (
	StocksContractSize   = 1
	StocksMinimumLotSize = 1.0
	StocksDecimalPlaces  = 2
	StocksPipValue       = 0.01
	StocksTickSize       = 0.01
)

// ==================== COMMODITIES CONSTANTS ====================

const (
	CommoditiesContractSize   = 1
	CommoditiesMinimumLotSize = 0.1
	CommoditiesDecimalPlaces  = 2
	CommoditiesPipValue       = 0.01
	CommoditiesTickSize       = 0.01
)

// ==================== CRYPTO CONSTANTS ====================

const (
	CryptoContractSize   = 1
	CryptoMinimumLotSize = 0.001
	CryptoDecimalPlaces  = 2
	CryptoPipValue       = 0.01
	CryptoTickSize       = 1.00
)

// ==================== COMMISSION DEFAULTS ====================

const (
	ForexCommissionType  = CommissionTypePerMillion
	ForexCommissionValue = 25.0

	StocksCommissionType  = CommissionTypePerShare
	StocksCommissionValue = 0.01

	CommoditiesCommissionType  = CommissionTypePerLot
	CommoditiesCommissionValue = 5.0

	CryptoCommissionType  = CommissionTypePercentage
	CryptoCommissionValue = 0.002 // 0.2%
)

// ==================== MOMENTUM MULTIPLIERS ====================

const (
	MomentumStrongMultiplier = 1.5 // +50% fill
	MomentumNormalMultiplier = 1.0 // 100% of available
	MomentumWeakMultiplier   = 0.5 // 50% of available
)

// ==================== VOLUME MULTIPLIERS ====================

const (
	HighVolumeMultiplier   = 1.0 // 100% fill ratio
	NormalVolumeMultiplier = 0.8 // 80% fill ratio
	LowVolumeMultiplier    = 0.5 // 50% fill ratio
)

// ==================== P&L CALCULATION THRESHOLDS ====================

const (
	MinimumPnLTrackingThreshold = 0.01 // Track P&L changes > 0.01
)

// ==================== VALIDATION THRESHOLDS ====================

const (
	MinimumOrderSize = 0.001
	MaximumOrderSize = 1000000.0
)

// ==================== TIMING CONSTANTS ====================

const (
	TickProcessingDelay = 1000000000 // 1 second in nanoseconds
	MinimumLatency      = 0          // milliseconds
	MaximumLatency      = 5000       // milliseconds
)

// ==================== UTILITY FUNCTIONS ====================

// IsValidInstrumentType checks if the instrument type is supported
func IsValidInstrumentType(instrumentType string) bool {
	switch instrumentType {
	case InstrumentTypeForex, InstrumentTypeStocks, InstrumentTypeCommodities, InstrumentTypeCrypto:
		return true
	default:
		return false
	}
}

// IsValidOrderAction checks if the action is valid
func IsValidOrderAction(action string) bool {
	switch action {
	case OrderActionBuy, OrderActionSell, OrderActionHold:
		return true
	default:
		return false
	}
}

// IsValidOrderType checks if the order type is valid
func IsValidOrderType(orderType string) bool {
	switch orderType {
	case OrderTypeMarket, OrderTypeLimit:
		return true
	default:
		return false
	}
}

// IsValidOrderStatus checks if the order status is valid
func IsValidOrderStatus(status string) bool {
	switch status {
	case OrderStatusFilled, OrderStatusPartial, OrderStatusRejected, OrderStatusPending, OrderStatusCancelled:
		return true
	default:
		return false
	}
}

// IsValidAccountStatus checks if the account status is valid
func IsValidAccountStatus(status string) bool {
	switch status {
	case AccountStatusActive, AccountStatusBlown, AccountStatusAtLimit:
		return true
	default:
		return false
	}
}

// IsValidPositionStatus checks if the position status is valid
func IsValidPositionStatus(status string) bool {
	switch status {
	case PositionStatusFlat, PositionStatusLong, PositionStatusShort:
		return true
	default:
		return false
	}
}

// GetPositionStatusFromSize returns position status based on size
func GetPositionStatusFromSize(size float64) string {
	if size == 0 {
		return PositionStatusFlat
	} else if size > 0 {
		return PositionStatusLong
	} else {
		return PositionStatusShort
	}
}

// GetInstrumentDefaults returns default values for an instrument type
func GetInstrumentDefaults(instrumentType string) map[string]interface{} {
	switch instrumentType {
	case InstrumentTypeForex:
		return map[string]interface{}{
			"contract_size":    ForexContractSize,
			"minimum_lot_size": ForexMinimumLotSize,
			"decimal_places":   ForexDecimalPlaces,
			"pip_value":        ForexPipValue,
			"tick_size":        ForexTickSize,
			"commission_type":  ForexCommissionType,
			"commission_value": ForexCommissionValue,
		}
	case InstrumentTypeStocks:
		return map[string]interface{}{
			"contract_size":    StocksContractSize,
			"minimum_lot_size": StocksMinimumLotSize,
			"decimal_places":   StocksDecimalPlaces,
			"pip_value":        StocksPipValue,
			"tick_size":        StocksTickSize,
			"commission_type":  StocksCommissionType,
			"commission_value": StocksCommissionValue,
		}
	case InstrumentTypeCommodities:
		return map[string]interface{}{
			"contract_size":    CommoditiesContractSize,
			"minimum_lot_size": CommoditiesMinimumLotSize,
			"decimal_places":   CommoditiesDecimalPlaces,
			"pip_value":        CommoditiesPipValue,
			"tick_size":        CommoditiesTickSize,
			"commission_type":  CommoditiesCommissionType,
			"commission_value": CommoditiesCommissionValue,
		}
	case InstrumentTypeCrypto:
		return map[string]interface{}{
			"contract_size":    CryptoContractSize,
			"minimum_lot_size": CryptoMinimumLotSize,
			"decimal_places":   CryptoDecimalPlaces,
			"pip_value":        CryptoPipValue,
			"tick_size":        CryptoTickSize,
			"commission_type":  CryptoCommissionType,
			"commission_value": CryptoCommissionValue,
		}
	default:
		return nil
	}
}

// GetMomentumMultiplier returns the fill multiplier based on momentum level
func GetMomentumMultiplier(momentum string) float64 {
	switch momentum {
	case MomentumStrong:
		return MomentumStrongMultiplier
	case MomentumNormal:
		return MomentumNormalMultiplier
	case MomentumWeak:
		return MomentumWeakMultiplier
	default:
		return MomentumNormalMultiplier
	}
}

// GetVolumeMultiplier returns the fill multiplier based on volume level
func GetVolumeMultiplier(volume, avgVolume int64) float64 {
	ratio := float64(volume) / float64(avgVolume)
	if ratio > 1.0 {
		return HighVolumeMultiplier
	} else if ratio < 0.5 {
		return LowVolumeMultiplier
	}
	return NormalVolumeMultiplier
}
