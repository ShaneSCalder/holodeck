package simulator

import (
	"encoding/json"
)

// ==================== CONFIGURATION STRUCTURES ====================

// Config represents the simulation configuration (JSON)
type Config struct {
	Instrument InstrumentConfig `json:"instrument"`
	Account    AccountConfig    `json:"account"`
	Execution  ExecutionConfig  `json:"execution"`
	Speed      SpeedConfig      `json:"speed"`
	CSV        CSVConfig        `json:"csv"`
}

// InstrumentConfig specifies the instrument to trade
type InstrumentConfig struct {
	Type        string `json:"type"`   // FOREX, STOCKS, COMMODITIES, CRYPTO
	Symbol      string `json:"symbol"` // e.g., EURUSD, AAPL, GOLD, BTC
	Description string `json:"description"`
}

// AccountConfig specifies account parameters
type AccountConfig struct {
	InitialBalance     float64 `json:"initialBalance"`
	Currency           string  `json:"currency"`
	Leverage           float64 `json:"leverage"`
	MaxDrawdownPercent float64 `json:"maxDrawdownPercent"`
	MaxPositionSize    float64 `json:"maxPositionSize"`
}

// ExecutionConfig specifies execution parameters
type ExecutionConfig struct {
	Commission       bool    `json:"commission"`
	CommissionType   string  `json:"commissionType"`
	CommissionValue  float64 `json:"commissionValue"`
	Slippage         bool    `json:"slippage"`
	SlippageModel    string  `json:"slippageModel"`
	Latency          bool    `json:"latency"`
	LatencyMs        int     `json:"latencyMs"`
	PartialFills     bool    `json:"partialFills"`
	PartialFillBased string  `json:"partialFillBasedOn"`
}

// SpeedConfig specifies speed control
type SpeedConfig struct {
	Multiplier float64 `json:"multiplier"`
}

// CSVConfig specifies CSV data source
type CSVConfig struct {
	FilePath string `json:"filePath"`
}

// ==================== JSON MARSHALING ====================

// UnmarshalJSON handles JSON unmarshaling with validation
func (c *Config) UnmarshalJSON(data []byte) error {
	type Alias Config
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	return c.Validate()
}

// ==================== VALIDATION ====================

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Instrument.Type == "" {
		return NewConfigError("instrument.type", "is required")
	}

	if c.Instrument.Symbol == "" {
		return NewConfigError("instrument.symbol", "is required")
	}

	if c.Account.InitialBalance <= 0 {
		return NewConfigError("account.initialBalance", "must be positive")
	}

	if c.Account.Leverage <= 0 {
		c.Account.Leverage = 1.0 // Default
	}

	if c.CSV.FilePath == "" {
		return NewConfigError("csv.filePath", "is required")
	}

	return nil
}

// ==================== CONFIG ERROR ====================

// ConfigError represents a configuration error
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return "config error: " + e.Field + " " + e.Message
}

// NewConfigError creates a new configuration error
func NewConfigError(field, message string) *ConfigError {
	return &ConfigError{
		Field:   field,
		Message: message,
	}
}
