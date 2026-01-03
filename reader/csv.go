package reader

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"holodeck/types"
)

// ==================== CSV READER ====================

// CSVTickReader reads tick data from a CSV file
type CSVTickReader struct {
	filePath    string
	file        *os.File
	reader      *csv.Reader
	tickCount   int64
	lineNumber  int64
	currentLine []string
	closed      bool
	hasNext     bool

	// Parser configuration
	config *ParserConfig

	// Statistics
	validTicks   int64
	invalidTicks int64
	parseErrors  int64
}

// ParserConfig holds configuration for CSV parsing
type ParserConfig struct {
	// Column indices (0-based)
	TimestampCol int
	BidCol       int
	AskCol       int
	BidQtyCol    int
	AskQtyCol    int
	LastPriceCol int
	VolumeCol    int

	// Timestamp format
	TimestampFormat string

	// Skip first line (header)
	SkipHeader bool

	// Validation
	ValidateData bool
}

// DefaultParserConfig returns a default parser configuration
// Expects CSV format: timestamp,bid,ask,bid_qty,ask_qty,last_price,volume
func DefaultParserConfig() *ParserConfig {
	return &ParserConfig{
		TimestampCol:    0,
		BidCol:          1,
		AskCol:          2,
		BidQtyCol:       3,
		AskQtyCol:       4,
		LastPriceCol:    5,
		VolumeCol:       6,
		TimestampFormat: time.RFC3339Nano,
		SkipHeader:      true,
		ValidateData:    true,
	}
}

// ==================== CONSTRUCTOR ====================

// NewCSVTickReader creates a new CSV tick reader
func NewCSVTickReader(filePath string) (*CSVTickReader, error) {
	return NewCSVTickReaderWithConfig(filePath, DefaultParserConfig())
}

// NewCSVTickReaderWithConfig creates a CSV reader with custom configuration
func NewCSVTickReaderWithConfig(filePath string, config *ParserConfig) (*CSVTickReader, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, types.NewConfigError("filePath", fmt.Sprintf("CSV file not found: %s", filePath))
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, types.NewConfigError("filePath", fmt.Sprintf("failed to open CSV file: %v", err))
	}

	// Create reader
	csvReader := csv.NewReader(file)

	reader := &CSVTickReader{
		filePath:     filePath,
		file:         file,
		reader:       csvReader,
		config:       config,
		tickCount:    0,
		lineNumber:   0,
		closed:       false,
		hasNext:      true,
		validTicks:   0,
		invalidTicks: 0,
		parseErrors:  0,
	}

	// Skip header if configured
	if config.SkipHeader {
		if _, err := csvReader.Read(); err != nil && err != io.EOF {
			file.Close()
			return nil, types.NewConfigError("csv", fmt.Sprintf("failed to read header: %v", err))
		}
		reader.lineNumber++
	}

	return reader, nil
}

// ==================== READING OPERATIONS ====================

// HasNext checks if there are more ticks to read
func (ctr *CSVTickReader) HasNext() bool {
	if ctr.closed {
		return false
	}
	return ctr.hasNext
}

// Next returns the next tick from the CSV file
func (ctr *CSVTickReader) Next() (*types.Tick, error) {
	if ctr.closed {
		return nil, types.NewConfigError("reader", "reader is closed")
	}

	// Read next line
	line, err := ctr.reader.Read()
	if err != nil {
		if err == io.EOF {
			ctr.hasNext = false
			return nil, fmt.Errorf("EOF")
		}
		ctr.lineNumber++
		ctr.parseErrors++
		return nil, types.NewCSVReadError(ctr.filePath, int(ctr.lineNumber), fmt.Sprintf("read error: %v", err))
	}

	ctr.lineNumber++

	// Parse line
	tick, err := ctr.parseLine(line)
	if err != nil {
		ctr.invalidTicks++
		return nil, err
	}

	ctr.tickCount++
	ctr.validTicks++

	return tick, nil
}

// ==================== PARSING ====================

// parseLine parses a CSV line into a Tick
func (ctr *CSVTickReader) parseLine(line []string) (*types.Tick, error) {
	// Check minimum columns
	minCols := ctr.config.VolumeCol + 1
	if len(line) < minCols {
		ctr.invalidTicks++
		return nil, types.NewCSVReadError(
			ctr.filePath,
			int(ctr.lineNumber),
			fmt.Sprintf("insufficient columns: expected at least %d, got %d", minCols, len(line)),
		)
	}

	// Parse timestamp
	timestamp, err := time.Parse(ctr.config.TimestampFormat, line[ctr.config.TimestampCol])
	if err != nil {
		return nil, types.NewCSVReadError(
			ctr.filePath,
			int(ctr.lineNumber),
			fmt.Sprintf("invalid timestamp format: %s (expected %s)", line[ctr.config.TimestampCol], ctr.config.TimestampFormat),
		)
	}

	// Parse bid
	bid, err := strconv.ParseFloat(line[ctr.config.BidCol], 64)
	if err != nil {
		return nil, types.NewCSVReadError(
			ctr.filePath,
			int(ctr.lineNumber),
			fmt.Sprintf("invalid bid price: %s", line[ctr.config.BidCol]),
		)
	}

	// Parse ask
	ask, err := strconv.ParseFloat(line[ctr.config.AskCol], 64)
	if err != nil {
		return nil, types.NewCSVReadError(
			ctr.filePath,
			int(ctr.lineNumber),
			fmt.Sprintf("invalid ask price: %s", line[ctr.config.AskCol]),
		)
	}

	// Parse bid quantity
	bidQty, err := strconv.ParseInt(line[ctr.config.BidQtyCol], 10, 64)
	if err != nil {
		return nil, types.NewCSVReadError(
			ctr.filePath,
			int(ctr.lineNumber),
			fmt.Sprintf("invalid bid quantity: %s", line[ctr.config.BidQtyCol]),
		)
	}

	// Parse ask quantity
	askQty, err := strconv.ParseInt(line[ctr.config.AskQtyCol], 10, 64)
	if err != nil {
		return nil, types.NewCSVReadError(
			ctr.filePath,
			int(ctr.lineNumber),
			fmt.Sprintf("invalid ask quantity: %s", line[ctr.config.AskQtyCol]),
		)
	}

	// Parse last price
	lastPrice, err := strconv.ParseFloat(line[ctr.config.LastPriceCol], 64)
	if err != nil {
		return nil, types.NewCSVReadError(
			ctr.filePath,
			int(ctr.lineNumber),
			fmt.Sprintf("invalid last price: %s", line[ctr.config.LastPriceCol]),
		)
	}

	// Parse volume
	volume, err := strconv.ParseInt(line[ctr.config.VolumeCol], 10, 64)
	if err != nil {
		return nil, types.NewCSVReadError(
			ctr.filePath,
			int(ctr.lineNumber),
			fmt.Sprintf("invalid volume: %s", line[ctr.config.VolumeCol]),
		)
	}

	// Create tick
	tick := types.NewTick(timestamp, bid, ask, lastPrice, bidQty, askQty, volume, ctr.tickCount)

	// Validate if configured
	if ctr.config.ValidateData {
		if !tick.IsValid() {
			return nil, types.NewCSVReadError(
				ctr.filePath,
				int(ctr.lineNumber),
				fmt.Sprintf("invalid tick data: bid=%.8f ask=%.8f last=%.8f", bid, ask, lastPrice),
			)
		}
	}

	return tick, nil
}

// ==================== STATE QUERIES ====================

// GetTickCount returns the number of ticks read
func (ctr *CSVTickReader) GetTickCount() int64 {
	return ctr.tickCount
}

// GetLineNumber returns the current line number
func (ctr *CSVTickReader) GetLineNumber() int64 {
	return ctr.lineNumber
}

// GetValidTickCount returns the number of valid ticks
func (ctr *CSVTickReader) GetValidTickCount() int64 {
	return ctr.validTicks
}

// GetInvalidTickCount returns the number of invalid ticks
func (ctr *CSVTickReader) GetInvalidTickCount() int64 {
	return ctr.invalidTicks
}

// GetParseErrorCount returns the number of parse errors
func (ctr *CSVTickReader) GetParseErrorCount() int64 {
	return ctr.parseErrors
}

// IsClosed checks if the reader is closed
func (ctr *CSVTickReader) IsClosed() bool {
	return ctr.closed
}

// ==================== CONTROL OPERATIONS ====================

// Reset resets the reader to the beginning
func (ctr *CSVTickReader) Reset() error {
	if ctr.closed {
		return types.NewInvalidOperationError("Reset", "reader is closed")
	}

	// Close and reopen file
	if err := ctr.file.Close(); err != nil {
		return types.NewConfigError("reader", fmt.Sprintf("failed to close file: %v", err))
	}

	// Reopen file
	file, err := os.Open(ctr.filePath)
	if err != nil {
		return types.NewConfigError("filePath", fmt.Sprintf("failed to reopen file: %v", err))
	}

	// Create new CSV reader
	csvReader := csv.NewReader(file)

	// Update reader state
	ctr.file = file
	ctr.reader = csvReader
	ctr.tickCount = 0
	ctr.lineNumber = 0
	ctr.hasNext = true
	ctr.validTicks = 0
	ctr.invalidTicks = 0
	ctr.parseErrors = 0

	// Skip header if configured
	if ctr.config.SkipHeader {
		if _, err := csvReader.Read(); err != nil && err != io.EOF {
			file.Close()
			return types.NewConfigError("csv", fmt.Sprintf("failed to read header: %v", err))
		}
		ctr.lineNumber++
	}

	return nil
}

// Close closes the CSV reader
func (ctr *CSVTickReader) Close() error {
	if ctr.closed {
		return nil
	}

	ctr.closed = true
	ctr.hasNext = false

	if ctr.file != nil {
		return ctr.file.Close()
	}

	return nil
}

// ==================== STATISTICS ====================

// GetStatistics returns reader statistics
func (ctr *CSVTickReader) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"file_path":       ctr.filePath,
		"ticks_read":      ctr.tickCount,
		"lines_processed": ctr.lineNumber,
		"valid_ticks":     ctr.validTicks,
		"invalid_ticks":   ctr.invalidTicks,
		"parse_errors":    ctr.parseErrors,
		"success_rate":    ctr.getSuccessRate(),
		"is_closed":       ctr.closed,
		"has_next":        ctr.hasNext,
	}
}

// getSuccessRate calculates the success rate
func (ctr *CSVTickReader) getSuccessRate() float64 {
	if ctr.lineNumber == 0 {
		return 0
	}
	return (float64(ctr.validTicks) / float64(ctr.lineNumber)) * 100
}

// String returns a human-readable string representation
func (ctr *CSVTickReader) String() string {
	return fmt.Sprintf(
		"CSVTickReader[File=%s, Ticks=%d, Valid=%d, Invalid=%d, Success=%.1f%%]",
		ctr.filePath,
		ctr.tickCount,
		ctr.validTicks,
		ctr.invalidTicks,
		ctr.getSuccessRate(),
	)
}

// DebugString returns detailed debug information
func (ctr *CSVTickReader) DebugString() string {
	return fmt.Sprintf(
		"CSV Tick Reader:\n"+
			"  File:              %s\n"+
			"  Status:            %s\n"+
			"  Ticks Read:        %d\n"+
			"  Lines Processed:   %d\n"+
			"  Valid Ticks:       %d\n"+
			"  Invalid Ticks:     %d\n"+
			"  Parse Errors:      %d\n"+
			"  Success Rate:      %.1f%%\n"+
			"  Has Next:          %v\n"+
			"  Parser Config:\n"+
			"    Timestamp Col:   %d\n"+
			"    Bid Col:         %d\n"+
			"    Ask Col:         %d\n"+
			"    Bid Qty Col:     %d\n"+
			"    Ask Qty Col:     %d\n"+
			"    Last Price Col:  %d\n"+
			"    Volume Col:      %d\n"+
			"    Timestamp Fmt:   %s\n"+
			"    Skip Header:     %v\n"+
			"    Validate:        %v",
		ctr.filePath,
		func() string {
			if ctr.closed {
				return "CLOSED"
			}
			return "OPEN"
		}(),
		ctr.tickCount,
		ctr.lineNumber,
		ctr.validTicks,
		ctr.invalidTicks,
		ctr.parseErrors,
		ctr.getSuccessRate(),
		ctr.hasNext,
		ctr.config.TimestampCol,
		ctr.config.BidCol,
		ctr.config.AskCol,
		ctr.config.BidQtyCol,
		ctr.config.AskQtyCol,
		ctr.config.LastPriceCol,
		ctr.config.VolumeCol,
		ctr.config.TimestampFormat,
		ctr.config.SkipHeader,
		ctr.config.ValidateData,
	)
}

// ==================== ADVANCED FEATURES ====================

// ReadN reads the next N ticks and returns them as a slice
func (ctr *CSVTickReader) ReadN(n int) ([]*types.Tick, error) {
	ticks := make([]*types.Tick, 0, n)

	for i := 0; i < n && ctr.HasNext(); i++ {
		tick, err := ctr.Next()
		if err != nil {
			// Continue on error for partial reads
			continue
		}
		ticks = append(ticks, tick)
	}

	if len(ticks) == 0 {
		return nil, types.NewConfigError("reader", "no ticks read")
	}

	return ticks, nil
}

// ReadUntil reads ticks until a time limit is reached
func (ctr *CSVTickReader) ReadUntil(maxTime time.Time) ([]*types.Tick, error) {
	ticks := make([]*types.Tick, 0)

	for ctr.HasNext() {
		tick, err := ctr.Next()
		if err != nil {
			// Continue on error
			continue
		}

		if tick.Timestamp.After(maxTime) {
			break
		}

		ticks = append(ticks, tick)
	}

	if len(ticks) == 0 {
		return nil, types.NewConfigError("reader", "no ticks read")
	}

	return ticks, nil
}
