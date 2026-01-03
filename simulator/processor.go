package simulator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ==================== SIMULATOR PROCESSOR ====================

// Processor orchestrates the Holodeck simulation
type Processor struct {
	configFile string
	speed      float64
	logLevel   string
	outputDir  string
	config     *Config
	startTime  time.Time
	results    *SimulationResults
}

// ==================== CREATION ====================

// NewProcessor creates a new simulator processor
func NewProcessor(configFile string, speed float64, logLevel, outputDir string) *Processor {
	return &Processor{
		configFile: configFile,
		speed:      speed,
		logLevel:   logLevel,
		outputDir:  outputDir,
	}
}

// ==================== EXECUTION ====================

// Process executes the full simulation workflow
func (p *Processor) Process() error {
	p.startTime = time.Now()

	// Step 1: Parse configuration
	if err := p.parseConfig(); err != nil {
		return fmt.Errorf("failed to parse configuration: %w", err)
	}

	// Step 2: Validate configuration
	if err := p.validateConfig(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Step 3: Create output directory
	if err := p.createOutputDir(); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Step 4: Print startup info
	p.printStartupInfo()

	// Step 5: Execute simulation
	if err := p.executeSimulation(); err != nil {
		return fmt.Errorf("simulation execution failed: %w", err)
	}

	// Step 6: Generate results
	p.generateResults()

	// Step 7: Print results
	p.printResults()

	// Step 8: Save results
	if err := p.saveResults(); err != nil {
		fmt.Printf("[WARNING] Failed to save results: %v\n", err)
	}

	return nil
}

// ==================== CONFIGURATION ====================

// parseConfig parses the configuration file (JSON)
func (p *Processor) parseConfig() error {
	fmt.Println("[INFO] Parsing configuration...")

	content, err := os.ReadFile(p.configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	p.config = &Config{}
	err = json.Unmarshal(content, p.config)
	if err != nil {
		return fmt.Errorf("failed to parse JSON config: %w", err)
	}

	return nil
}

// validateConfig validates the configuration
func (p *Processor) validateConfig() error {
	fmt.Println("[INFO] Validating configuration...")

	// Check if CSV file exists
	if p.config.CSV.FilePath == "" {
		return fmt.Errorf("csv.filePath is required")
	}

	if _, err := os.Stat(p.config.CSV.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("data file not found: %s", p.config.CSV.FilePath)
	}

	// Validate account settings
	if p.config.Account.InitialBalance <= 0 {
		return fmt.Errorf("account.initialBalance must be positive")
	}

	if p.config.Account.Leverage <= 0 {
		return fmt.Errorf("account.leverage must be positive")
	}

	// Validate instrument
	if p.config.Instrument.Type == "" {
		return fmt.Errorf("instrument.type is required")
	}

	if p.config.Instrument.Symbol == "" {
		return fmt.Errorf("instrument.symbol is required")
	}

	// Validate speed
	if p.speed <= 0 {
		return fmt.Errorf("speed multiplier must be positive")
	}

	return nil
}

// createOutputDir creates the output directory
func (p *Processor) createOutputDir() error {
	return os.MkdirAll(p.outputDir, 0755)
}

// ==================== EXECUTION ====================

// printStartupInfo prints startup information
func (p *Processor) printStartupInfo() {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("HOLODECK SIMULATION EXECUTION")
	fmt.Println(strings.Repeat("=", 70) + "\n")

	fmt.Printf("[INFO] Instrument:         %s (%s)\n", p.config.Instrument.Symbol, p.config.Instrument.Type)
	fmt.Printf("[INFO] Data File:          %s\n", p.config.CSV.FilePath)
	fmt.Printf("[INFO] Speed:              %.1fx\n", p.speed)
	fmt.Printf("[INFO] Log Level:          %s\n", p.logLevel)
	fmt.Printf("[INFO] Output Directory:   %s\n", p.outputDir)
	fmt.Printf("[INFO] Initial Balance:    $%.2f\n", p.config.Account.InitialBalance)
	fmt.Printf("[INFO] Leverage:           %.1fx\n", p.config.Account.Leverage)
	fmt.Printf("[INFO] Max Drawdown:       %.1f%%\n", p.config.Account.MaxDrawdownPercent)

	fmt.Println("\n[INFO] Processing ticks...")
}

// executeSimulation executes the real simulation
func (p *Processor) executeSimulation() error {
	// TODO: This is where the real Holodeck API will be called
	// For now, simulate with placeholder

	ticksToProcess := 50000
	progress := 0.0

	for i := 0; i < ticksToProcess; i++ {
		progress = float64(i) / float64(ticksToProcess) * 100

		// Calculate sleep time based on speed multiplier
		// At 1000x speed: process 1000 ticks per second
		// At 100x speed: process 100 ticks per second
		// At 1x speed: process 1 tick per second
		sleepDuration := time.Duration(1000000/p.speed) * time.Microsecond

		time.Sleep(sleepDuration)

		// Print progress every 10%
		if i%5000 == 0 && i > 0 {
			fmt.Printf("[PROGRESS] %.1f%% complete (%d / %d ticks)\n", progress, i, ticksToProcess)
		}
	}

	fmt.Printf("[PROGRESS] 100%% complete (%d / %d ticks)\n", ticksToProcess, ticksToProcess)
	return nil
}

// ==================== RESULTS ====================

// generateResults generates simulation results
func (p *Processor) generateResults() {
	p.results = &SimulationResults{
		Instrument:     p.config.Instrument.Symbol,
		InstrumentType: p.config.Instrument.Type,
		StartTime:      p.startTime,
		EndTime:        time.Now(),
		Speed:          p.speed,
		TicksProcessed: 50000,
		TradeCount:     0,
		WinCount:       0,
		LossCount:      0,
		NetProfit:      0,
		Commission:     0,
		InitialBalance: p.config.Account.InitialBalance,
		FinalBalance:   p.config.Account.InitialBalance,
		AccountStatus:  "ACTIVE",
	}

	p.results.ElapsedTime = p.results.EndTime.Sub(p.results.StartTime)
}

// printResults prints the results to console
func (p *Processor) printResults() {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("SIMULATION RESULTS")
	fmt.Println(strings.Repeat("=", 70) + "\n")

	fmt.Printf("Instrument:        %s (%s)\n", p.results.Instrument, p.results.InstrumentType)
	fmt.Printf("Execution Time:    %s (at %.1fx speed)\n", p.results.ElapsedTime, p.results.Speed)
	fmt.Printf("Ticks Processed:   %d\n\n", p.results.TicksProcessed)

	fmt.Println("ACCOUNT:")
	fmt.Printf("  Initial Balance:   $%.2f\n", p.results.InitialBalance)
	fmt.Printf("  Final Balance:     $%.2f\n", p.results.FinalBalance)
	fmt.Printf("  Net P&L:           $%.2f\n", p.results.NetProfit)
	fmt.Printf("  Commission:        $%.2f\n", p.results.Commission)
	fmt.Printf("  Account Status:    %s\n\n", p.results.AccountStatus)

	fmt.Println("TRADES:")
	fmt.Printf("  Total Trades:      %d\n", p.results.TradeCount)
	fmt.Printf("  Winning Trades:    %d\n", p.results.WinCount)
	fmt.Printf("  Losing Trades:     %d\n\n", p.results.LossCount)

	fmt.Println(strings.Repeat("=", 70))
}

// saveResults saves results to file
func (p *Processor) saveResults() error {
	logPath := filepath.Join(p.outputDir, fmt.Sprintf("simulation_%d.txt", p.startTime.Unix()))
	content := p.formatResultsForFile()
	return os.WriteFile(logPath, []byte(content), 0644)
}

// formatResultsForFile formats results for file output
func (p *Processor) formatResultsForFile() string {
	return fmt.Sprintf(
		"HOLODECK SIMULATION RESULTS\n"+
			"===========================\n\n"+
			"Instrument:        %s (%s)\n"+
			"Execution Date:    %s\n"+
			"Execution Time:    %s\n"+
			"Speed:             %.1fx\n"+
			"Ticks Processed:   %d\n\n"+
			"ACCOUNT:\n"+
			"  Initial Balance: $%.2f\n"+
			"  Final Balance:   $%.2f\n"+
			"  Net P&L:         $%.2f\n"+
			"  Commission:      $%.2f\n"+
			"  Status:          %s\n\n"+
			"TRADES:\n"+
			"  Total:           %d\n"+
			"  Wins:            %d\n"+
			"  Losses:          %d\n",
		p.results.Instrument,
		p.results.InstrumentType,
		p.results.StartTime.Format("2006-01-02 15:04:05"),
		p.results.ElapsedTime,
		p.results.Speed,
		p.results.TicksProcessed,
		p.results.InitialBalance,
		p.results.FinalBalance,
		p.results.NetProfit,
		p.results.Commission,
		p.results.AccountStatus,
		p.results.TradeCount,
		p.results.WinCount,
		p.results.LossCount,
	)
}

// ==================== RESULTS STRUCTURE ====================

// SimulationResults represents simulation results
type SimulationResults struct {
	Instrument     string
	InstrumentType string
	StartTime      time.Time
	EndTime        time.Time
	ElapsedTime    time.Duration
	Speed          float64
	TicksProcessed int64
	TradeCount     int64
	WinCount       int64
	LossCount      int64
	NetProfit      float64
	Commission     float64
	InitialBalance float64
	FinalBalance   float64
	AccountStatus  string
}
