package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"holodeck/simulator"
	"holodeck/types"
)

// ==================== VERSION ====================

const (
	Version = "1.0.0"
	Name    = "Holodeck"
)

// ==================== MAIN ====================

func main() {
	// Define command-line flags
	configFile := flag.String("config", "", "Path to configuration JSON file (REQUIRED)")
	speed := flag.Float64("speed", 100.0, "Simulation speed multiplier (default 100.0)")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	showHelp := flag.Bool("help", false, "Show help message")
	showVersion := flag.Bool("version", false, "Show version information")

	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("%s version %s\n", Name, Version)
		fmt.Println("Live market simulator for trading strategies")
		os.Exit(0)
	}

	// Handle help flag
	if *showHelp {
		printUsage()
		os.Exit(0)
	}

	// Validate required config flag
	if *configFile == "" {
		fmt.Println("Error: -config flag is required")
		fmt.Println("\nUsage: holodeck -config <file.json> [-speed <multiplier>] [-verbose]")
		fmt.Println("       holodeck -help")
		fmt.Println("       holodeck -version")
		os.Exit(1)
	}

	// Check if config file exists
	if _, err := os.Stat(*configFile); os.IsNotExist(err) {
		fmt.Printf("Error: Configuration file not found: %s\n", *configFile)
		os.Exit(1)
	}

	// Step 1: Load configuration
	if *verbose {
		fmt.Printf("[INFO] Loading configuration from %s\n", *configFile)
	}

	config, err := loadConfigFromFile(*configFile)
	if err != nil {
		log.Fatalf("[ERROR] Failed to load configuration: %v", err)
	}

	// Step 2: Create Holodeck from config
	if *verbose {
		fmt.Println("[INFO] Initializing Holodeck simulator...")
	}

	holodeck, err := config.NewHolodeck()
	if err != nil {
		log.Fatalf("[ERROR] Failed to initialize Holodeck: %v", err)
	}

	// Step 3: Override speed if specified
	if *speed > 0 {
		if err := holodeck.SetSpeed(*speed); err != nil {
			log.Fatalf("[ERROR] Failed to set speed: %v", err)
		}
	}

	// Step 4: Start simulation
	if *verbose {
		fmt.Printf("[INFO] Starting simulation at %.1fx speed\n", *speed)
	}

	if err := holodeck.Start(); err != nil {
		log.Fatalf("[ERROR] Failed to start simulation: %v", err)
	}

	// Step 5: Main simulation loop
	tickCount := 0
	tradeCount := 0

	for holodeck.IsRunning() && !holodeck.IsAccountBlown() {
		// Get next tick from data source
		tick, err := holodeck.GetNextTick()
		if err != nil {
			// No more ticks available
			break
		}

		tickCount++

		// Print progress every 10000 ticks
		if *verbose && tickCount%10000 == 0 {
			balance := holodeck.GetBalance()
			fmt.Printf("[PROGRESS] Processed %d ticks | Balance: $%.2f\n",
				tickCount, balance.CurrentBalance)
		}

		// TODO: Add agent decision logic here
		// Example:
		// if shouldExecuteOrder(tick) {
		//     order := createOrder(tick)
		//     exec, err := holodeck.ExecuteOrder(order)
		//     if err == nil && exec.FilledSize > 0 {
		//         tradeCount++
		//     }
		// }

		_ = tick // Placeholder to use tick variable
	}

	// Step 6: Stop simulation
	if *verbose {
		fmt.Println("[INFO] Stopping simulation...")
	}

	if err := holodeck.Stop(); err != nil {
		log.Fatalf("[ERROR] Failed to stop simulation: %v", err)
	}

	// Step 7: Retrieve final metrics
	metrics := holodeck.GetMetrics()
	balance := holodeck.GetBalance()
	position := holodeck.GetPosition()

	// Step 8: Print results
	printResults(metrics, balance, position, tickCount, tradeCount)
}

// loadConfigFromFile loads configuration from a JSON file
func loadConfigFromFile(filePath string) (*simulator.Config, error) {
	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	config := &simulator.Config{}
	if err := json.Unmarshal(content, config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON config: %w", err)
	}

	return config, nil
}

// printResults prints the simulation results in a formatted way
func printResults(metrics map[string]interface{}, balance *types.Balance, position *types.Position, ticks int, trades int) {
	fmt.Println("\n" + strings.Repeat("=", 63))
	fmt.Println(strings.Repeat(" ", 15) + "SIMULATION RESULTS")
	fmt.Println(strings.Repeat("=", 63) + "\n")

	// Market data
	fmt.Println("MARKET DATA:")
	if ticks > 0 {
		fmt.Printf("  Ticks Processed:           %d\n", ticks)
	}
	if v, ok := metrics["total_ticks_available"]; ok {
		fmt.Printf("  Total Available Ticks:     %v\n", v)
	}

	// Trades
	fmt.Println("\nTRADES:")
	if trades > 0 {
		fmt.Printf("  Trades Executed:           %d\n", trades)
	} else {
		fmt.Printf("  Trades Executed:           0 (Demo mode)\n")
	}
	if v, ok := metrics["trades_executed"]; ok {
		fmt.Printf("  Total Executed:            %v\n", v)
	}

	// Account information
	fmt.Println("\nACCOUNT:")
	if balance != nil && balance.InitialBalance > 0 {
		fmt.Printf("  Initial Balance:           $%.2f\n", balance.InitialBalance)
		fmt.Printf("  Final Balance:             $%.2f\n", balance.CurrentBalance)
		netPnL := balance.CurrentBalance - balance.InitialBalance
		fmt.Printf("  Net P&L:                   $%.2f\n", netPnL)
		fmt.Printf("  Commission Paid:           $%.2f\n", balance.CommissionPaid)
	}

	// Performance metrics
	fmt.Println("\nPERFORMANCE:")
	if v, ok := metrics["return_percent"]; ok {
		fmt.Printf("  Return %%:                   %.2f%%\n", v)
	}
	if v, ok := metrics["drawdown_percent"]; ok {
		fmt.Printf("  Max Drawdown %%:            %.2f%%\n", v)
	}
	if v, ok := metrics["win_rate"]; ok {
		fmt.Printf("  Win Rate:                  %.2f%%\n", v)
	}

	// Position information
	fmt.Println("\nPOSITION:")
	if position != nil {
		fmt.Printf("  Size:                      %.2f\n", position.Size)
		fmt.Printf("  Entry Price:               %.4f\n", position.EntryPrice)
		fmt.Printf("  Unrealized P&L:            $%.2f\n", position.UnrealizedPnL)
		fmt.Printf("  Realized P&L:              $%.2f\n", position.RealizedPnL)
	}

	// Session duration
	if v, ok := metrics["session_duration"]; ok {
		fmt.Printf("\nSession Duration:           %v\n", v)
	}

	fmt.Println("\n" + strings.Repeat("=", 63) + "\n")
}

// ==================== USAGE ====================

func printUsage() {
	fmt.Printf(`%s - Live Market Simulator
Version %s

USAGE:
    holodeck -config <file.json> [options]

OPTIONS:
    -config <file>      Configuration file (JSON) - REQUIRED
    -speed <multiplier> Simulation speed multiplier (default: 100.0)
    -verbose            Enable verbose output
    -help               Show this help message
    -version            Show version information

EXAMPLES:
    # Basic simulation at default 100x speed
    holodeck -config config.json

    # Simulation at 1000x speed with verbose output
    holodeck -config config.json -speed 1000.0 -verbose

    # Show version
    holodeck -version

CONFIGURATION FILE:
    The configuration file should be in JSON format. Example:

    {
      "csv": {
        "filepath": "data/eurusd_ticks.csv"
      },
      "instrument": {
        "type": "FOREX",
        "symbol": "EURUSD",
        "description": "Euro vs US Dollar"
      },
      "account": {
        "initial_balance": 10000.0,
        "currency": "USD",
        "leverage": 1.0,
        "max_drawdown_percent": 20.0
      },
      "execution": {
        "commission": true,
        "commission_value": 0.0001,
        "slippage": true,
        "slippage_model": "fixed",
        "latency": false
      },
      "speed": {
        "multiplier": 100.0
      },
      "logging": {
        "verbose": true,
        "log_file": "logs/simulation.log"
      }
    }

SPEED MULTIPLIERS:
    0.1x - 0.5x  Slower than real-time
    1.0x         Real-time
    10x          10x faster
    100x         100x faster (1 year in ~2.5 min) - Recommended
    1000x        1000x faster
    10000x+      Ultra-fast

SUPPORTED INSTRUMENTS:
    - FOREX: Foreign exchange pairs (EUR/USD, GBP/USD, etc.)
    - STOCKS: Individual stocks (AAPL, MSFT, etc.)
    - COMMODITIES: Commodities (GOLD, OIL, etc.)
    - CRYPTO: Cryptocurrencies (BTC, ETH, etc.)

DESCRIPTION:
    Holodeck is a live market simulator that processes historical tick data
    at configurable speeds. It provides a realistic simulation environment
    for testing trading strategies, agents, and algorithms.

    All executions include realistic friction:
    - Commission fees based on instrument type
    - Slippage from market depth and order size
    - Latency simulation
    - Partial fills for large orders

    The simulator is event-driven, processing one tick at a time and
    supporting flexible agent decision logic.

`, Name, Version)
}
