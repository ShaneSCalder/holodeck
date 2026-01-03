package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ==================== FILE LOGGER ====================

// FileLogger implements Logger interface with file-based logging
type FileLogger struct {
	// Configuration
	sessionID  string
	logDir     string
	verbosity  VerbosityLevel
	bufferSize int

	// File handles
	tradeFile   *os.File
	errorFile   *os.File
	metricsFile *os.File
	infoFile    *os.File

	// Buffering
	buffer      []string
	bufferMutex sync.Mutex

	// Statistics
	entriesLogged int64
	lastFlush     time.Time
	createdTime   time.Time
}

// ==================== CREATION ====================

// NewFileLogger creates a new file logger
func NewFileLogger(logDir string) (*FileLogger, error) {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	return &FileLogger{
		logDir:      logDir,
		verbosity:   VerbosityNormal,
		bufferSize:  100,
		lastFlush:   time.Now(),
		createdTime: time.Now(),
	}, nil
}

// ==================== SESSION MANAGEMENT ====================

// StartSession initializes logger for a new session
func (fl *FileLogger) StartSession(sessionID string) error {
	fl.sessionID = sessionID
	timestamp := time.Now().Format("2006-01-02_15-04-05")

	baseName := fmt.Sprintf("%s_%s", sessionID, timestamp)

	// Open trade log
	tradeFile, err := os.OpenFile(
		filepath.Join(fl.logDir, baseName+"_trades.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return err
	}
	fl.tradeFile = tradeFile

	// Open error log
	errorFile, err := os.OpenFile(
		filepath.Join(fl.logDir, baseName+"_errors.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return err
	}
	fl.errorFile = errorFile

	// Open metrics log
	metricsFile, err := os.OpenFile(
		filepath.Join(fl.logDir, baseName+"_metrics.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return err
	}
	fl.metricsFile = metricsFile

	// Open info log
	infoFile, err := os.OpenFile(
		filepath.Join(fl.logDir, baseName+"_info.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return err
	}
	fl.infoFile = infoFile

	// Write session header
	header := fmt.Sprintf("=== Holodeck Session %s ===\n", sessionID)
	header += fmt.Sprintf("Started: %s\n\n", time.Now().Format(time.RFC3339))

	fl.tradeFile.WriteString(header)
	fl.errorFile.WriteString(header)
	fl.metricsFile.WriteString(header)
	fl.infoFile.WriteString(header)

	return nil
}

// EndSession closes all log files for session
func (fl *FileLogger) EndSession(sessionID string) error {
	// Flush remaining buffer
	if err := fl.Flush(); err != nil {
		return err
	}

	// Write session footer
	footer := fmt.Sprintf("\n=== Session %s Ended ===\n", sessionID)
	footer += fmt.Sprintf("Ended: %s\n", time.Now().Format(time.RFC3339))
	footer += fmt.Sprintf("Total Entries Logged: %d\n", fl.entriesLogged)

	if fl.tradeFile != nil {
		fl.tradeFile.WriteString(footer)
		fl.tradeFile.Close()
	}
	if fl.errorFile != nil {
		fl.errorFile.WriteString(footer)
		fl.errorFile.Close()
	}
	if fl.metricsFile != nil {
		fl.metricsFile.WriteString(footer)
		fl.metricsFile.Close()
	}
	if fl.infoFile != nil {
		fl.infoFile.WriteString(footer)
		fl.infoFile.Close()
	}

	return nil
}

// GetSessionID returns current session ID
func (fl *FileLogger) GetSessionID() string {
	return fl.sessionID
}

// ==================== LOGGING METHODS ====================

// LogTrade logs a trade entry
func (fl *FileLogger) LogTrade(trade *TradeLog) error {
	if fl.verbosity < VerbosityMinimal {
		return nil
	}

	if fl.tradeFile == nil {
		return fmt.Errorf("trade log file not initialized")
	}

	entry := fmt.Sprintf(
		"[%s] TRADE: %s\n"+
			"  Order ID: %s\n"+
			"  Instrument: %s\n"+
			"  Action: %s | Type: %s\n"+
			"  Requested: %.4f | Filled: %.4f @ %.5f\n"+
			"  Commission: %.2f | Slippage: %.4f pips\n"+
			"  P&L: %.2f | Status: %s\n\n",
		trade.Timestamp.Format("2006-01-02 15:04:05.000"),
		trade.TradeID,
		trade.OrderID,
		trade.Instrument,
		trade.Action,
		trade.OrderType,
		trade.RequestedSize,
		trade.FilledSize,
		trade.FillPrice,
		trade.Commission,
		trade.Slippage,
		trade.RealizedPnL,
		trade.Status,
	)

	fl.bufferMutex.Lock()
	fl.buffer = append(fl.buffer, entry)
	fl.bufferMutex.Unlock()

	fl.entriesLogged++

	// Auto-flush if buffer is full
	if len(fl.buffer) >= fl.bufferSize {
		return fl.Flush()
	}

	return nil
}

// LogError logs an error entry
func (fl *FileLogger) LogError(errLog *ErrorLog) error {
	if fl.verbosity < VerbosityMinimal {
		return nil
	}

	if fl.errorFile == nil {
		return fmt.Errorf("error log file not initialized")
	}

	entry := fmt.Sprintf(
		"[%s] %s - %s\n"+
			"  Code: %s | Type: %s\n"+
			"  Message: %s\n"+
			"  Details: %s\n"+
			"  Trade ID: %s | Order ID: %s\n\n",
		errLog.Timestamp.Format("2006-01-02 15:04:05.000"),
		errLog.Severity.String(),
		errLog.ErrorCode,
		errLog.ErrorCode,
		errLog.ErrorType,
		errLog.Message,
		errLog.Details,
		errLog.TradeID,
		errLog.OrderID,
	)

	_, err := fl.errorFile.WriteString(entry)
	if err != nil {
		return err
	}

	fl.entriesLogged++
	return nil
}

// LogMetrics logs periodic metrics
func (fl *FileLogger) LogMetrics(metrics *MetricsLog) error {
	if fl.verbosity < VerbosityNormal {
		return nil
	}

	if fl.metricsFile == nil {
		return fmt.Errorf("metrics log file not initialized")
	}

	entry := fmt.Sprintf(
		"[%s] METRICS SNAPSHOT\n"+
			"  Session Duration: %v\n"+
			"  Initial Balance: $%.2f\n"+
			"  Current Balance: $%.2f\n"+
			"  Total P&L: $%.2f (%.2f%%)\n"+
			"  Trades: %d (Won: %d | Lost: %d | Win Rate: %.1f%%)\n"+
			"  Largest Win: $%.2f | Largest Loss: $%.2f\n"+
			"  Commission: $%.2f | Slippage: $%.2f\n"+
			"  Max Drawdown: %.2f%%\n"+
			"  Sharpe Ratio: %.2f\n"+
			"  Ticks Processed: %d | Errors: %d\n\n",
		metrics.Timestamp.Format("2006-01-02 15:04:05.000"),
		metrics.SessionDuration,
		metrics.InitialBalance,
		metrics.CurrentBalance,
		metrics.TotalPnL,
		metrics.TotalPnLPercent,
		metrics.TradeCount,
		metrics.WinningTrades,
		metrics.LosingTrades,
		metrics.WinRate,
		metrics.LargestWin,
		metrics.LargestLoss,
		metrics.CommissionTotal,
		metrics.SlippageTotal,
		metrics.MaxDrawdownPercent,
		metrics.SharpeRatio,
		metrics.TicksProcessed,
		metrics.ErrorCount,
	)

	_, err := fl.metricsFile.WriteString(entry)
	if err != nil {
		return err
	}

	fl.entriesLogged++
	return nil
}

// LogInfo logs informational message
func (fl *FileLogger) LogInfo(message string) error {
	if fl.verbosity < VerbosityVerbose {
		return nil
	}

	if fl.infoFile == nil {
		return fmt.Errorf("info log file not initialized")
	}

	entry := fmt.Sprintf("[%s] INFO: %s\n", time.Now().Format("2006-01-02 15:04:05.000"), message)

	_, err := fl.infoFile.WriteString(entry)
	if err != nil {
		return err
	}

	fl.entriesLogged++
	return nil
}

// LogWarning logs a warning message
func (fl *FileLogger) LogWarning(message string) error {
	if fl.verbosity < VerbosityMinimal {
		return nil
	}

	if fl.infoFile == nil {
		return fmt.Errorf("info log file not initialized")
	}

	entry := fmt.Sprintf("[%s] WARNING: %s\n", time.Now().Format("2006-01-02 15:04:05.000"), message)

	_, err := fl.infoFile.WriteString(entry)
	if err != nil {
		return err
	}

	fl.entriesLogged++
	return nil
}

// LogDebug logs a debug message
func (fl *FileLogger) LogDebug(message string) error {
	if fl.verbosity < VerbosityDebug {
		return nil
	}

	if fl.infoFile == nil {
		return fmt.Errorf("info log file not initialized")
	}

	entry := fmt.Sprintf("[%s] DEBUG: %s\n", time.Now().Format("2006-01-02 15:04:05.000"), message)

	_, err := fl.infoFile.WriteString(entry)
	if err != nil {
		return err
	}

	fl.entriesLogged++
	return nil
}

// ==================== CONTROL METHODS ====================

// SetVerbosity sets the verbosity level
func (fl *FileLogger) SetVerbosity(level VerbosityLevel) error {
	fl.verbosity = level
	return nil
}

// Flush flushes all buffered entries to disk
func (fl *FileLogger) Flush() error {
	fl.bufferMutex.Lock()
	defer fl.bufferMutex.Unlock()

	for _, entry := range fl.buffer {
		if fl.tradeFile != nil {
			fl.tradeFile.WriteString(entry)
		}
	}

	fl.buffer = make([]string, 0, fl.bufferSize)
	fl.lastFlush = time.Now()

	// Sync files to disk
	if fl.tradeFile != nil {
		fl.tradeFile.Sync()
	}
	if fl.errorFile != nil {
		fl.errorFile.Sync()
	}
	if fl.metricsFile != nil {
		fl.metricsFile.Sync()
	}
	if fl.infoFile != nil {
		fl.infoFile.Sync()
	}

	return nil
}

// Close closes all open log files
func (fl *FileLogger) Close() error {
	// Flush first
	if err := fl.Flush(); err != nil {
		return err
	}

	// Close files
	var lastErr error
	if fl.tradeFile != nil {
		if err := fl.tradeFile.Close(); err != nil {
			lastErr = err
		}
	}
	if fl.errorFile != nil {
		if err := fl.errorFile.Close(); err != nil {
			lastErr = err
		}
	}
	if fl.metricsFile != nil {
		if err := fl.metricsFile.Close(); err != nil {
			lastErr = err
		}
	}
	if fl.infoFile != nil {
		if err := fl.infoFile.Close(); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// ==================== STATISTICS ====================

// GetStatistics returns logger statistics
func (fl *FileLogger) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"entries_logged": fl.entriesLogged,
		"last_flush":     fl.lastFlush,
		"buffer_size":    len(fl.buffer),
		"verbosity":      fl.verbosity.String(),
		"session_id":     fl.sessionID,
		"uptime":         time.Since(fl.createdTime),
	}
}
