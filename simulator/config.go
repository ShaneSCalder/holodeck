package simulator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"holodeck/types"
)

// ==================== CONFIGURATION STRUCTURES ====================

// Config is the root configuration structure loaded from JSON
type Config struct {
	CSV        CSVConfig        `json:"csv"`
	Instrument InstrumentConfig `json:"instrument"`
	Account    AccountConfig    `json:"account"`
	Execution  ExecutionConfig  `json:"execution"`
	OrderTypes OrderTypesConfig `json:"order_types"`
	Speed      SpeedConfig      `json:"speed"`
	Session    SessionConfig    `json:"session"`
	Logging    LoggingConfig    `json:"logging"`
}

// CSVConfig defines the CSV data source
type CSVConfig struct {
	FilePath string `json:"filepath"`
}

// InstrumentConfig defines instrument-specific parameters
type InstrumentConfig struct {
	Type           string  `json:"type"`
	Symbol         string  `json:"symbol"`
	Description    string  `json:"description"`
	DecimalPlaces  int     `json:"decimal_places"`
	PipValue       float64 `json:"pip_value"`
	ContractSize   int64   `json:"contract_size"`
	MinimumLotSize float64 `json:"minimum_lot_size"`
	TickSize       float64 `json:"tick_size"`
}

// AccountConfig defines account parameters
type AccountConfig struct {
	InitialBalance     float64 `json:"initial_balance"`
	Currency           string  `json:"currency"`
	Leverage           float64 `json:"leverage"`
	MaxPositionSize    float64 `json:"max_position_size"`
	MaxDrawdownPercent float64 `json:"max_drawdown_percent"`
}

// ExecutionConfig defines execution parameters
type ExecutionConfig struct {
	Slippage           bool    `json:"slippage"`
	SlippageModel      string  `json:"slippage_model"`
	Latency            bool    `json:"latency"`
	LatencyMs          int64   `json:"latency_ms"`
	Commission         bool    `json:"commission"`
	CommissionType     string  `json:"commission_type"`
	CommissionValue    float64 `json:"commission_value"`
	PartialFills       bool    `json:"partial_fills"`
	PartialFillBasedOn string  `json:"partial_fill_based_on"`
}

// OrderTypesConfig defines supported order types
type OrderTypesConfig struct {
	Supported []string `json:"supported"`
	Default   string   `json:"default"`
}

// SpeedConfig defines simulation speed
type SpeedConfig struct {
	Multiplier float64 `json:"multiplier"`
}

// SessionConfig defines session parameters
type SessionConfig struct {
	ClosePositionsAtEnd bool `json:"close_positions_at_end"`
}

// LoggingConfig defines logging parameters
type LoggingConfig struct {
	Verbose       bool   `json:"verbose"`
	LogFile       string `json:"log_file"`
	LogEveryTick  bool   `json:"log_every_tick"`
	LogEveryTrade bool   `json:"log_every_trade"`
	LogMetrics    bool   `json:"log_metrics"`
}

// ==================== CONFIGURATION LOADER ====================

// ConfigLoader handles loading and validating configurations
type ConfigLoader struct {
	ConfigPath string
	Config     *Config
	Errors     []*types.HolodeckError
}

// NewConfigLoader creates a new configuration loader
func NewConfigLoader(configPath string) *ConfigLoader {
	return &ConfigLoader{
		ConfigPath: configPath,
		Errors:     make([]*types.HolodeckError, 0),
	}
}

// Load loads the configuration from JSON file
func (cl *ConfigLoader) Load() error {
	// Check if file exists
	if _, err := os.Stat(cl.ConfigPath); os.IsNotExist(err) {
		return types.NewConfigError("filepath", fmt.Sprintf("config file not found: %s", cl.ConfigPath))
	}

	// Read file
	data, err := ioutil.ReadFile(cl.ConfigPath)
	if err != nil {
		return types.NewConfigError("filepath", fmt.Sprintf("failed to read config file: %v", err))
	}

	// Parse JSON
	config := &Config{}
	if err := json.Unmarshal(data, config); err != nil {
		return types.NewConfigError("json", fmt.Sprintf("failed to parse JSON: %v", err))
	}

	cl.Config = config

	// Validate
	if err := cl.Validate(); err != nil {
		return err
	}

	return nil
}

// LoadFromString loads configuration from JSON string (useful for testing)
func (cl *ConfigLoader) LoadFromString(jsonString string) error {
	config := &Config{}
	if err := json.Unmarshal([]byte(jsonString), config); err != nil {
		return types.NewConfigError("json", fmt.Sprintf("failed to parse JSON: %v", err))
	}

	cl.Config = config

	// Validate
	if err := cl.Validate(); err != nil {
		return err
	}

	return nil
}

// ==================== VALIDATION ====================

// Validate validates the entire configuration
func (cl *ConfigLoader) Validate() error {
	if cl.Config == nil {
		return types.NewConfigError("config", "configuration not loaded")
	}

	cl.Errors = make([]*types.HolodeckError, 0)

	// Validate each section
	cl.validateCSV()
	cl.validateInstrument()
	cl.validateAccount()
	cl.validateExecution()
	cl.validateOrderTypes()
	cl.validateSpeed()
	cl.validateLogging()

	// Return first error if any
	if len(cl.Errors) > 0 {
		return cl.Errors[0]
	}

	return nil
}

// validateCSV validates CSV configuration
func (cl *ConfigLoader) validateCSV() {
	if cl.Config.CSV.FilePath == "" {
		cl.Errors = append(cl.Errors,
			types.NewConfigError("csv.filepath", "CSV filepath cannot be empty"))
	}
}

// validateInstrument validates instrument configuration
func (cl *ConfigLoader) validateInstrument() {
	// Check instrument type
	if !types.IsValidInstrumentType(cl.Config.Instrument.Type) {
		cl.Errors = append(cl.Errors,
			types.NewConfigError("instrument.type", fmt.Sprintf("invalid instrument type: %s", cl.Config.Instrument.Type)))
		return
	}

	// Check symbol
	if cl.Config.Instrument.Symbol == "" {
		cl.Errors = append(cl.Errors,
			types.NewConfigError("instrument.symbol", "symbol cannot be empty"))
	}

	// Check decimal places
	if cl.Config.Instrument.DecimalPlaces < 0 {
		cl.Errors = append(cl.Errors,
			types.NewConfigError("instrument.decimal_places", "decimal places cannot be negative"))
	}

	// Check pip value
	if cl.Config.Instrument.PipValue <= 0 {
		cl.Errors = append(cl.Errors,
			types.NewConfigError("instrument.pip_value", "pip value must be positive"))
	}

	// Check contract size
	if cl.Config.Instrument.ContractSize <= 0 {
		cl.Errors = append(cl.Errors,
			types.NewConfigError("instrument.contract_size", "contract size must be positive"))
	}

	// Check minimum lot size
	if cl.Config.Instrument.MinimumLotSize <= 0 {
		cl.Errors = append(cl.Errors,
			types.NewConfigError("instrument.minimum_lot_size", "minimum lot size must be positive"))
	}

	// Check tick size
	if cl.Config.Instrument.TickSize <= 0 {
		cl.Errors = append(cl.Errors,
			types.NewConfigError("instrument.tick_size", "tick size must be positive"))
	}
}

// validateAccount validates account configuration
func (cl *ConfigLoader) validateAccount() {
	// Check initial balance
	if cl.Config.Account.InitialBalance <= 0 {
		cl.Errors = append(cl.Errors,
			types.NewConfigError("account.initial_balance", "initial balance must be positive"))
	}

	// Check currency
	if cl.Config.Account.Currency == "" {
		cl.Errors = append(cl.Errors,
			types.NewConfigError("account.currency", "currency cannot be empty"))
	}

	// Check leverage
	if cl.Config.Account.Leverage < 1.0 {
		cl.Errors = append(cl.Errors,
			types.NewConfigError("account.leverage", "leverage must be >= 1.0"))
	}

	// Check max position size
	if cl.Config.Account.MaxPositionSize <= 0 {
		cl.Errors = append(cl.Errors,
			types.NewConfigError("account.max_position_size", "max position size must be positive"))
	}

	// Check max drawdown
	if cl.Config.Account.MaxDrawdownPercent <= 0 || cl.Config.Account.MaxDrawdownPercent > 100 {
		cl.Errors = append(cl.Errors,
			types.NewConfigError("account.max_drawdown_percent", "max drawdown must be between 0 and 100"))
	}
}

// validateExecution validates execution configuration
func (cl *ConfigLoader) validateExecution() {
	// Check slippage model if enabled
	if cl.Config.Execution.Slippage {
		validModels := []string{types.SlippageModelDepth, types.SlippageModelMomentum, types.SlippageModelFixed, types.SlippageModelNone}
		found := false
		for _, m := range validModels {
			if cl.Config.Execution.SlippageModel == m {
				found = true
				break
			}
		}
		if !found {
			cl.Errors = append(cl.Errors,
				types.NewConfigError("execution.slippage_model", fmt.Sprintf("invalid slippage model: %s", cl.Config.Execution.SlippageModel)))
		}
	}

	// Check latency
	if cl.Config.Execution.Latency && cl.Config.Execution.LatencyMs < 0 {
		cl.Errors = append(cl.Errors,
			types.NewConfigError("execution.latency_ms", "latency cannot be negative"))
	}

	// Check commission
	if cl.Config.Execution.Commission && cl.Config.Execution.CommissionValue < 0 {
		cl.Errors = append(cl.Errors,
			types.NewConfigError("execution.commission_value", "commission value cannot be negative"))
	}

	// Check partial fills
	if cl.Config.Execution.PartialFills {
		validLogic := []string{types.PartialFillByVolumeMomentum, types.PartialFillByDepth, types.PartialFillNone}
		found := false
		for _, l := range validLogic {
			if cl.Config.Execution.PartialFillBasedOn == l {
				found = true
				break
			}
		}
		if !found {
			cl.Errors = append(cl.Errors,
				types.NewConfigError("execution.partial_fill_based_on", fmt.Sprintf("invalid partial fill logic: %s", cl.Config.Execution.PartialFillBasedOn)))
		}
	}
}

// validateOrderTypes validates order types configuration
func (cl *ConfigLoader) validateOrderTypes() {
	// Check that at least one order type is supported
	if len(cl.Config.OrderTypes.Supported) == 0 {
		cl.Errors = append(cl.Errors,
			types.NewConfigError("order_types.supported", "must support at least one order type"))
	}

	// Check that all supported types are valid
	for _, ot := range cl.Config.OrderTypes.Supported {
		if !types.IsValidOrderType(ot) {
			cl.Errors = append(cl.Errors,
				types.NewConfigError("order_types.supported", fmt.Sprintf("invalid order type: %s", ot)))
		}
	}

	// Check default is valid
	if cl.Config.OrderTypes.Default == "" {
		cl.Errors = append(cl.Errors,
			types.NewConfigError("order_types.default", "default order type cannot be empty"))
	}

	if !types.IsValidOrderType(cl.Config.OrderTypes.Default) {
		cl.Errors = append(cl.Errors,
			types.NewConfigError("order_types.default", fmt.Sprintf("invalid default order type: %s", cl.Config.OrderTypes.Default)))
	}

	// Check default is in supported list
	found := false
	for _, ot := range cl.Config.OrderTypes.Supported {
		if ot == cl.Config.OrderTypes.Default {
			found = true
			break
		}
	}
	if !found {
		cl.Errors = append(cl.Errors,
			types.NewConfigError("order_types.default", "default order type must be in supported list"))
	}
}

// validateSpeed validates speed configuration
func (cl *ConfigLoader) validateSpeed() {
	// Check speed multiplier
	if cl.Config.Speed.Multiplier < 0.1 || cl.Config.Speed.Multiplier > 10000.0 {
		cl.Errors = append(cl.Errors,
			types.NewConfigError("speed.multiplier", "speed multiplier must be between 0.1 and 10000"))
	}
}

// validateLogging validates logging configuration
func (cl *ConfigLoader) validateLogging() {
	// Check log file path if logging is enabled
	if cl.Config.Logging.LogFile == "" && (cl.Config.Logging.LogEveryTick || cl.Config.Logging.LogEveryTrade || cl.Config.Logging.LogMetrics) {
		cl.Errors = append(cl.Errors,
			types.NewConfigError("logging.log_file", "log file path required if logging is enabled"))
	}
}

// ==================== GETTERS WITH DEFAULTS ====================

// GetCSVFilePath returns the CSV file path
func (c *Config) GetCSVFilePath() string {
	return c.CSV.FilePath
}

// GetInstrumentType returns the instrument type
func (c *Config) GetInstrumentType() string {
	return c.Instrument.Type
}

// GetInitialBalance returns the initial account balance
func (c *Config) GetInitialBalance() float64 {
	return c.Account.InitialBalance
}

// GetLeverage returns the account leverage
func (c *Config) GetLeverage() float64 {
	return c.Account.Leverage
}

// GetMaxDrawdownPercent returns the max drawdown limit
func (c *Config) GetMaxDrawdownPercent() float64 {
	return c.Account.MaxDrawdownPercent
}

// GetSpeedMultiplier returns the speed multiplier
func (c *Config) GetSpeedMultiplier() float64 {
	return c.Speed.Multiplier
}

// GetCommissionValue returns the commission value
func (c *Config) GetCommissionValue() float64 {
	return c.Execution.CommissionValue
}

// GetCommissionType returns the commission type
func (c *Config) GetCommissionType() string {
	return c.Execution.CommissionType
}

// GetLatencyMs returns the latency in milliseconds
func (c *Config) GetLatencyMs() int64 {
	return c.Execution.LatencyMs
}

// IsSlippageEnabled checks if slippage is enabled
func (c *Config) IsSlippageEnabled() bool {
	return c.Execution.Slippage
}

// IsCommissionEnabled checks if commission is enabled
func (c *Config) IsCommissionEnabled() bool {
	return c.Execution.Commission
}

// IsLatencyEnabled checks if latency is enabled
func (c *Config) IsLatencyEnabled() bool {
	return c.Execution.Latency
}

// ArePartialFillsEnabled checks if partial fills are enabled
func (c *Config) ArePartialFillsEnabled() bool {
	return c.Execution.PartialFills
}

// IsVerboseLogging checks if verbose logging is enabled
func (c *Config) IsVerboseLogging() bool {
	return c.Logging.Verbose
}

// ShouldLogEveryTick checks if every tick should be logged
func (c *Config) ShouldLogEveryTick() bool {
	return c.Logging.LogEveryTick
}

// ShouldLogEveryTrade checks if every trade should be logged
func (c *Config) ShouldLogEveryTrade() bool {
	return c.Logging.LogEveryTrade
}

// ShouldLogMetrics checks if metrics should be logged
func (c *Config) ShouldLogMetrics() bool {
	return c.Logging.LogMetrics
}

// ==================== EXPORT TO TYPES ====================

// ToInstrumentConfig converts Config to types.InstrumentConfig
func (c *Config) ToInstrumentConfig() *types.InstrumentConfig {
	return &types.InstrumentConfig{
		Type:            c.Instrument.Type,
		Symbol:          c.Instrument.Symbol,
		Description:     c.Instrument.Description,
		DecimalPlaces:   c.Instrument.DecimalPlaces,
		PipValue:        c.Instrument.PipValue,
		ContractSize:    c.Instrument.ContractSize,
		MinimumLotSize:  c.Instrument.MinimumLotSize,
		TickSize:        c.Instrument.TickSize,
		CommissionType:  c.Execution.CommissionType,
		CommissionValue: c.Execution.CommissionValue,
	}
}

// ==================== CONFIGURATION MANAGER ====================

// ConfigManager manages a set of configurations
type ConfigManager struct {
	Configs map[string]*Config
	Default string
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		Configs: make(map[string]*Config),
	}
}

// LoadConfig loads a configuration and registers it
func (cm *ConfigManager) LoadConfig(name, filePath string) error {
	loader := NewConfigLoader(filePath)
	if err := loader.Load(); err != nil {
		return err
	}

	cm.Configs[name] = loader.Config

	// Set as default if first one
	if cm.Default == "" {
		cm.Default = name
	}

	return nil
}

// GetConfig retrieves a configuration by name
func (cm *ConfigManager) GetConfig(name string) (*Config, error) {
	config, ok := cm.Configs[name]
	if !ok {
		return nil, types.NewConfigError("name", fmt.Sprintf("configuration not found: %s", name))
	}
	return config, nil
}

// GetDefault retrieves the default configuration
func (cm *ConfigManager) GetDefault() (*Config, error) {
	if cm.Default == "" {
		return nil, types.NewConfigError("default", "no default configuration set")
	}
	return cm.GetConfig(cm.Default)
}

// SetDefault sets the default configuration
func (cm *ConfigManager) SetDefault(name string) error {
	if _, ok := cm.Configs[name]; !ok {
		return types.NewConfigError("name", fmt.Sprintf("configuration not found: %s", name))
	}
	cm.Default = name
	return nil
}

// List returns a list of all configuration names
func (cm *ConfigManager) List() []string {
	names := make([]string, 0, len(cm.Configs))
	for name := range cm.Configs {
		names = append(names, name)
	}
	return names
}

// Size returns the number of configurations
func (cm *ConfigManager) Size() int {
	return len(cm.Configs)
}

// ==================== ENVIRONMENT-BASED LOADING ====================

// LoadFromDirectory loads all JSON config files from a directory
func (cm *ConfigManager) LoadFromDirectory(dirPath string) error {
	// Check if directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return types.NewConfigError("dirpath", fmt.Sprintf("directory not found: %s", dirPath))
	}

	// Read directory
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return types.NewConfigError("dirpath", fmt.Sprintf("failed to read directory: %v", err))
	}

	// Load all JSON files
	loadedCount := 0
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			// Extract name from filename (without extension)
			name := file.Name()[:len(file.Name())-5]

			// Load config
			filePath := filepath.Join(dirPath, file.Name())
			if err := cm.LoadConfig(name, filePath); err != nil {
				// Log error but continue
				fmt.Printf("Warning: Failed to load config %s: %v\n", name, err)
				continue
			}

			loadedCount++
		}
	}

	if loadedCount == 0 {
		return types.NewConfigError("dirpath", fmt.Sprintf("no JSON config files found in %s", dirPath))
	}

	return nil
}

// ==================== CONFIGURATION EXPORT ====================

// ToJSON exports the configuration as JSON string
func (c *Config) ToJSON() (string, error) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SaveToFile saves the configuration to a JSON file
func (c *Config) SaveToFile(filePath string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return err
	}

	return nil
}

// ==================== CONFIGURATION SUMMARY ====================

// Summary returns a string summary of the configuration
func (c *Config) Summary() string {
	return fmt.Sprintf(
		"Configuration Summary:\n"+
			"  Instrument:  %s (%s)\n"+
			"  CSV File:    %s\n"+
			"  Initial:     %s %.2f\n"+
			"  Leverage:    %.1fx\n"+
			"  Max PosSize: %.2f\n"+
			"  Max DD:      %.1f%%\n"+
			"  Speed:       %.1fx\n"+
			"  Commission:  %s (%.4f)\n"+
			"  Slippage:    %v (%s)\n"+
			"  Latency:     %v (%dms)\n"+
			"  Partial:     %v\n"+
			"  Logging:     %v (Tick:%v Trade:%v Metrics:%v)",
		c.Instrument.Type, c.Instrument.Symbol,
		c.CSV.FilePath,
		c.Account.Currency, c.Account.InitialBalance,
		c.Account.Leverage,
		c.Account.MaxPositionSize,
		c.Account.MaxDrawdownPercent,
		c.Speed.Multiplier,
		c.Execution.CommissionType, c.Execution.CommissionValue,
		c.Execution.Slippage, c.Execution.SlippageModel,
		c.Execution.Latency, c.Execution.LatencyMs,
		c.Execution.PartialFills,
		c.Logging.Verbose, c.Logging.LogEveryTick, c.Logging.LogEveryTrade, c.Logging.LogMetrics,
	)
}

// DebugString returns a detailed debug representation
func (c *Config) DebugString() string {
	json, _ := c.ToJSON()
	return fmt.Sprintf("Config (JSON):\n%s", json)
}

// ==================== HOLODECK INITIALIZATION METHODS ====================

// NewCSVReader creates a CSV tick reader from config
func (c *Config) NewCSVReader() (TickReader, error) {
	if c.CSV.FilePath == "" {
		return nil, fmt.Errorf("CSV file path not configured")
	}

	// Check if file exists
	if _, err := os.Stat(c.CSV.FilePath); err != nil {
		return nil, fmt.Errorf("CSV file not found: %s (%w)", c.CSV.FilePath, err)
	}

	// TODO: Uncomment when reader package is available
	// return reader.NewCSVTickReader(c.CSV.FilePath)
	return nil, fmt.Errorf("reader.NewCSVTickReader not yet available")
}

// NewExecutor creates an order executor from config
func (c *Config) NewExecutor() (OrderExecutor, error) {
	// TODO: Uncomment when executor package is available
	// return executor.NewOrderExecutor(c.Execution)
	return nil, fmt.Errorf("executor.NewOrderExecutor not yet available")
}

// NewLogger creates a logger from config
func (c *Config) NewLogger() (Logger, error) {
	if !c.Logging.Verbose {
		return nil, nil
	}

	if c.Logging.LogFile != "" {
		// Create directory for log file
		dir := filepath.Dir(c.Logging.LogFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		// TODO: Uncomment when logger package is available
		// return logger.NewFileLogger(c.Logging.LogFile, c.Logging)
		return nil, fmt.Errorf("logger.NewFileLogger not yet available")
	}

	// TODO: Uncomment when logger package is available
	// return logger.NewConsoleLogger(c.Logging)
	return nil, fmt.Errorf("logger.NewConsoleLogger not yet available")
}

// NewInstrument creates an instrument from config
func (c *Config) NewInstrument() (types.Instrument, error) {
	if c.Instrument.Symbol == "" {
		return nil, fmt.Errorf("instrument symbol not configured")
	}

	instrumentType := c.Instrument.Type

	switch instrumentType {
	case "FOREX":
		// TODO: Uncomment when instrument package is available
		// return instrument.NewForex(c.Instrument.Symbol)
		return nil, fmt.Errorf("instrument.NewForex not yet available")

	case "STOCKS":
		// TODO: Uncomment when instrument package is available
		// return instrument.NewStocks(c.Instrument.Symbol)
		return nil, fmt.Errorf("instrument.NewStocks not yet available")

	case "COMMODITIES":
		// TODO: Uncomment when instrument package is available
		// return instrument.NewCommodities(c.Instrument.Symbol)
		return nil, fmt.Errorf("instrument.NewCommodities not yet available")

	case "CRYPTO":
		// TODO: Uncomment when instrument package is available
		// return instrument.NewCrypto(c.Instrument.Symbol)
		return nil, fmt.Errorf("instrument.NewCrypto not yet available")

	default:
		return nil, fmt.Errorf("unknown instrument type: %s", instrumentType)
	}
}

// NewHolodeck creates and configures a complete Holodeck simulator from config
// This is the main factory method that initializes all subsystems
func (c *Config) NewHolodeck() (*Holodeck, error) {
	// Step 1: Create CSV reader
	reader, err := c.NewCSVReader()
	if err != nil {
		return nil, fmt.Errorf("failed to create CSV reader: %w", err)
	}

	// Step 2: Create executor
	executor, err := c.NewExecutor()
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	// Step 3: Create logger
	logger, err := c.NewLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// Step 4: Create instrument
	instrument, err := c.NewInstrument()
	if err != nil {
		return nil, fmt.Errorf("failed to create instrument: %w", err)
	}

	// Step 5: Create HolodeckConfig
	hConfig := &HolodeckConfig{
		SessionID:  fmt.Sprintf("session_%d", time.Now().Unix()),
		StartTime:  time.Now(),
		IsRunning:  false,
		Instrument: instrument,
		ExecutionConfig: ExecutionParameters{
			CommissionEnabled:   c.Execution.Commission,
			CommissionType:      c.Execution.CommissionType,
			CommissionValue:     c.Execution.CommissionValue,
			SlippageEnabled:     c.Execution.Slippage,
			SlippageModel:       c.Execution.SlippageModel,
			LatencyEnabled:      c.Execution.Latency,
			LatencyMs:           c.Execution.LatencyMs,
			PartialFillsEnabled: c.Execution.PartialFills,
			PartialFillLogic:    c.Execution.PartialFillBasedOn,
			SpeedMultiplier:     0,
		},
		DataSource: DataSourceConfig{
			FilePath: c.CSV.FilePath,
			Format:   "CSV",
		},
		StateConfig: StateConfiguration{
			MaxTicksToKeep:          100000,
			MaxPositionHistorySize:  10000,
			MaxBalanceHistorySize:   10000,
			MaxExecutionHistorySize: 10000,
		},
	}

	// Step 6: Create Holodeck
	holodeck, err := NewHolodeck(hConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Holodeck: %w", err)
	}

	// Step 7: Wire subsystems
	holodeck = holodeck.
		WithReader(reader).
		WithExecutor(executor).
		WithLogger(logger)

	// Step 8: Set speed
	if c.Speed.Multiplier > 0 {
		if err := holodeck.SetSpeed(c.Speed.Multiplier); err != nil {
			return nil, fmt.Errorf("failed to set speed: %w", err)
		}
	}

	return holodeck, nil
}
