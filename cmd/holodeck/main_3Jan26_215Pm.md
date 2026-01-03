package main

import (
	"fmt"
	"os"
)

// ==================== VERSION ====================

const (
	Version = "1.0.0"
	Name    = "Holodeck"
)

// ==================== MAIN ====================

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	command := os.Args[1]

	// Parse common flags
	configFile := ""
	speed := 100.0
	logLevel := "NORMAL"
	outputDir := "./data_out"

	// Simple argument parsing
	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "-config":
			if i+1 < len(os.Args) {
				configFile = os.Args[i+1]
				i++
			}
		case "-speed":
			if i+1 < len(os.Args) {
				fmt.Sscanf(os.Args[i+1], "%f", &speed)
				i++
			}
		case "-log-level":
			if i+1 < len(os.Args) {
				logLevel = os.Args[i+1]
				i++
			}
		case "-output":
			if i+1 < len(os.Args) {
				outputDir = os.Args[i+1]
				i++
			}
		}
	}

	// Route to command
	switch command {
	case "backtest", "simulate", "run":
		// All map to the same simulation engine
		if configFile == "" {
			fmt.Println("Error: -config flag is required")
			printSimulationUsage()
			os.Exit(1)
		}

		processor := NewBacktestProcessor(configFile, speed, logLevel, outputDir)
		if err := processor.Process(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	case "version", "-v", "--version":
		fmt.Printf("%s version %s\n", Name, Version)
		fmt.Println("Live market simulator for trading strategies")

	case "help", "-h", "--help":
		printUsage()

	default:
		fmt.Printf("Error: unknown command '%s'\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

// ==================== USAGE ====================

func printUsage() {
	fmt.Printf(`%s - Live Market Simulator
Version %s

USAGE:
    holodeck <command> [options]

COMMANDS:
    backtest    Run a simulation (accepts simulated or real tick data)
    simulate    Alias for backtest
    run         Alias for backtest
    version     Show version information
    help        Show this help message

EXAMPLES:
    # Run simulation with simulated tick data
    holodeck backtest -config data/forex_eurusd.json

    # Run at 1000x speed
    holodeck backtest -config data/forex_eurusd.json -speed 1000

    # Show version
    holodeck version

RUN SIMULATION:
    holodeck backtest [options]
    
    Options:
      -config <file>      Configuration file (JSON) - REQUIRED
      -speed <multiplier> Simulation speed (default: 100)
      -log-level <level>  Logging level (default: NORMAL)
      -output <dir>       Output directory (default: ./data_out)

CONFIGURATION FILE:
    The configuration file should be in JSON format with the following structure:

    {
      "instrument": {
        "type": "FOREX",
        "symbol": "EURUSD",
        "description": "Euro vs US Dollar"
      },
      "account": {
        "initialBalance": 10000,
        "currency": "USD",
        "leverage": 100,
        "maxDrawdownPercent": 20,
        "maxPositionSize": 100000
      },
      "execution": {
        "commission": true,
        "commissionType": "per_million",
        "commissionValue": 25,
        "slippage": true,
        "slippageModel": "depth_based",
        "latency": true,
        "latencyMs": 50,
        "partialFills": true
      },
      "speed": {
        "multiplier": 1.0
      },
      "csv": {
        "filePath": "data/ticks_eurusd.csv"
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

See: holodeck backtest -h for simulation-specific options

`, Name, Version)
}

func printSimulationUsage() {
	fmt.Printf(`%s - Run Simulation
Version %s

USAGE:
    holodeck backtest [options]

OPTIONS:
    -config <file>      Configuration file (JSON) - REQUIRED
    -speed <multiplier> Simulation speed multiplier (default: 100)
    -log-level <level>  Logging level: QUIET, MINIMAL, NORMAL, VERBOSE, DEBUG (default: NORMAL)
    -output <dir>       Output directory for results (default: ./data_out)
    -h, --help          Show this help message

EXAMPLES:
    # Basic simulation
    holodeck backtest -config data/forex_eurusd.json

    # Fast 1000x speed simulation
    holodeck backtest -config data/forex_eurusd.json -speed 1000

    # Verbose with debug logging
    holodeck backtest -config data/forex_eurusd.json -log-level DEBUG

    # Custom output directory
    holodeck backtest -config data/forex_eurusd.json -output ./results

SPEED MULTIPLIERS:
    0.1x - 0.5x  Slower than real-time
    1.0x         Real-time
    10x          10x faster
    100x         100x faster (1 year in ~2.5 min) - Recommended
    1000x        1000x faster
    10000x+      Ultra-fast

`, Name, Version)
}
