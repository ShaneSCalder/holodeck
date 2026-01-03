package reader

import (
	"fmt"
	"time"

	"holodeck/types"
)

// ==================== CSV FORMAT DETECTION ====================

// DetectTimestampFormat attempts to detect the timestamp format from a sample line
func DetectTimestampFormat(sample string) string {
	// Try common formats
	formats := []string{
		time.RFC3339Nano,          // 2024-01-15T07:00:00.000000000Z
		time.RFC3339,              // 2024-01-15T07:00:00Z
		"2006-01-02T15:04:05.000", // 2024-01-15T07:00:00.000
		"2006-01-02 15:04:05.000", // 2024-01-15 07:00:00.000
		"2006-01-02 15:04:05",     // 2024-01-15 07:00:00
		"01/02/2006 15:04:05",     // 01/15/2024 07:00:00
		"2006-01-02",              // 2024-01-15
	}

	for _, format := range formats {
		if _, err := time.Parse(format, sample); err == nil {
			return format
		}
	}

	// Default to RFC3339Nano
	return time.RFC3339Nano
}

// ==================== CSV VALIDATION ====================

// ValidateCSVHeader checks if a CSV header matches expected columns
func ValidateCSVHeader(header []string, expectedColumns []string) error {
	if len(header) < len(expectedColumns) {
		return types.NewCSVReadError(
			"unknown",
			1,
			fmt.Sprintf("header has %d columns, expected at least %d", len(header), len(expectedColumns)),
		)
	}

	// Check if expected columns are present (in any order)
	headerMap := make(map[string]bool)
	for _, col := range header {
		headerMap[col] = true
	}

	for _, expected := range expectedColumns {
		if !headerMap[expected] {
			return types.NewCSVReadError(
				"unknown",
				1,
				fmt.Sprintf("missing expected column: %s", expected),
			)
		}
	}

	return nil
}

// ==================== TICK VALIDATION ====================

// TickValidator validates ticks for data quality
type TickValidator struct {
	minBid       float64
	maxBid       float64
	minAsk       float64
	maxAsk       float64
	maxSpread    float64
	requireDepth bool
}

// NewTickValidator creates a new tick validator
func NewTickValidator() *TickValidator {
	return &TickValidator{
		minBid:       0.00001,
		maxBid:       100000.0,
		minAsk:       0.00001,
		maxAsk:       100000.0,
		maxSpread:    1.0,
		requireDepth: false,
	}
}

// WithBidRange sets the acceptable bid price range
func (tv *TickValidator) WithBidRange(min, max float64) *TickValidator {
	tv.minBid = min
	tv.maxBid = max
	return tv
}

// WithAskRange sets the acceptable ask price range
func (tv *TickValidator) WithAskRange(min, max float64) *TickValidator {
	tv.minAsk = min
	tv.maxAsk = max
	return tv
}

// WithMaxSpread sets the maximum acceptable spread
func (tv *TickValidator) WithMaxSpread(maxSpread float64) *TickValidator {
	tv.maxSpread = maxSpread
	return tv
}

// WithDepthRequirement sets whether depth is required
func (tv *TickValidator) WithDepthRequirement(required bool) *TickValidator {
	tv.requireDepth = required
	return tv
}

// ValidateTick validates a tick against rules
func (tv *TickValidator) ValidateTick(tick *types.Tick) error {
	// Check bid range
	if tick.Bid < tv.minBid || tick.Bid > tv.maxBid {
		return types.NewConfigError(
			"tick.bid",
			fmt.Sprintf("bid %.8f outside range [%.8f, %.8f]", tick.Bid, tv.minBid, tv.maxBid),
		)
	}

	// Check ask range
	if tick.Ask < tv.minAsk || tick.Ask > tv.maxAsk {
		return types.NewConfigError(
			"tick.ask",
			fmt.Sprintf("ask %.8f outside range [%.8f, %.8f]", tick.Ask, tv.minAsk, tv.maxAsk),
		)
	}

	// Check spread
	spread := tick.Ask - tick.Bid
	if spread > tv.maxSpread {
		return types.NewConfigError(
			"tick.spread",
			fmt.Sprintf("spread %.8f exceeds max %.8f", spread, tv.maxSpread),
		)
	}

	// Check depth if required
	if tv.requireDepth {
		if tick.BidQty == 0 || tick.AskQty == 0 {
			return types.NewConfigError(
				"tick.depth",
				fmt.Sprintf("insufficient depth: bid_qty=%d, ask_qty=%d", tick.BidQty, tick.AskQty),
			)
		}
	}

	return nil
}

// ==================== CSV COLUMN DETECTION ====================

// AutodetectColumns attempts to autodetect CSV column positions
// Assumes header row with standard column names
func AutodetectColumns(header []string) (*ParserConfig, error) {
	config := DefaultParserConfig()

	// Map to find columns
	columnMap := make(map[string]int)
	for i, col := range header {
		columnMap[col] = i
	}

	// Standard column names (case-sensitive for now)
	stdColumns := map[string]*int{
		"timestamp": &config.TimestampCol,
		"time":      &config.TimestampCol,
		"date":      &config.TimestampCol,
		"datetime":  &config.TimestampCol,

		"bid":       &config.BidCol,
		"bid_price": &config.BidCol,

		"ask":       &config.AskCol,
		"ask_price": &config.AskCol,

		"bid_qty":      &config.BidQtyCol,
		"bid_quantity": &config.BidQtyCol,
		"bid_size":     &config.BidQtyCol,

		"ask_qty":      &config.AskQtyCol,
		"ask_quantity": &config.AskQtyCol,
		"ask_size":     &config.AskQtyCol,

		"last":       &config.LastPriceCol,
		"last_price": &config.LastPriceCol,
		"price":      &config.LastPriceCol,

		"volume": &config.VolumeCol,
		"vol":    &config.VolumeCol,
		"qty":    &config.VolumeCol,
		"size":   &config.VolumeCol,
	}

	// Try to find columns
	for name, configPtr := range stdColumns {
		if col, ok := columnMap[name]; ok {
			*configPtr = col
		}
	}

	return config, nil
}

// ==================== BATCH READING ====================

// BatchReader reads ticks in batches for efficiency
type BatchReader struct {
	reader    *CSVTickReader
	batchSize int
}

// NewBatchReader creates a new batch reader
func NewBatchReader(reader *CSVTickReader, batchSize int) *BatchReader {
	if batchSize < 1 {
		batchSize = 100
	}
	return &BatchReader{
		reader:    reader,
		batchSize: batchSize,
	}
}

// ReadBatch reads the next batch of ticks
func (br *BatchReader) ReadBatch() ([]*types.Tick, error) {
	ticks := make([]*types.Tick, 0, br.batchSize)

	for i := 0; i < br.batchSize && br.reader.HasNext(); i++ {
		tick, err := br.reader.Next()
		if err != nil {
			// If we have some ticks, return them
			if len(ticks) > 0 {
				return ticks, nil
			}
			// Otherwise return error
			return nil, err
		}
		ticks = append(ticks, tick)
	}

	if len(ticks) == 0 {
		return nil, types.NewConfigError("reader", "no ticks read")
	}

	return ticks, nil
}

// ReadAllBatches reads all remaining ticks in batches
func (br *BatchReader) ReadAllBatches() ([][]*types.Tick, error) {
	var allBatches [][]*types.Tick

	for br.reader.HasNext() {
		batch, err := br.ReadBatch()
		if err != nil {
			// Return what we have so far
			if len(allBatches) > 0 {
				return allBatches, nil
			}
			return nil, err
		}
		allBatches = append(allBatches, batch)
	}

	if len(allBatches) == 0 {
		return nil, types.NewConfigError("reader", "no batches read")
	}

	return allBatches, nil
}

// ==================== STREAMING READER ====================

// StreamingReader provides a channel-based interface for reading ticks
type StreamingReader struct {
	reader *CSVTickReader
	tickCh chan *types.Tick
	errCh  chan error
	done   chan bool
	stopCh chan bool
}

// NewStreamingReader creates a new streaming reader
func NewStreamingReader(reader *CSVTickReader) *StreamingReader {
	return &StreamingReader{
		reader: reader,
		tickCh: make(chan *types.Tick, 100), // Buffered channel
		errCh:  make(chan error, 10),
		done:   make(chan bool),
		stopCh: make(chan bool),
	}
}

// Start begins streaming ticks
func (sr *StreamingReader) Start() {
	go sr.stream()
}

// stream reads ticks and sends them through the channel
func (sr *StreamingReader) stream() {
	defer close(sr.tickCh)
	defer close(sr.errCh)

	for sr.reader.HasNext() {
		select {
		case <-sr.stopCh:
			sr.done <- true
			return
		default:
			tick, err := sr.reader.Next()
			if err != nil {
				sr.errCh <- err
				continue
			}
			sr.tickCh <- tick
		}
	}

	sr.done <- true
}

// GetTicks returns the tick channel
func (sr *StreamingReader) GetTicks() <-chan *types.Tick {
	return sr.tickCh
}

// GetErrors returns the error channel
func (sr *StreamingReader) GetErrors() <-chan error {
	return sr.errCh
}

// Stop stops the streaming reader
func (sr *StreamingReader) Stop() {
	sr.stopCh <- true
	<-sr.done
}

// ==================== STATISTICS ====================

// ReaderStatistics holds statistics about CSV reading
type ReaderStatistics struct {
	FilePath       string
	TicksRead      int64
	LinesProcessed int64
	ValidTicks     int64
	InvalidTicks   int64
	ParseErrors    int64
	SuccessRate    float64
}

// GetReaderStatistics extracts statistics from a reader
func GetReaderStatistics(reader *CSVTickReader) *ReaderStatistics {
	return &ReaderStatistics{
		FilePath:       reader.filePath,
		TicksRead:      reader.GetTickCount(),
		LinesProcessed: reader.GetLineNumber(),
		ValidTicks:     reader.GetValidTickCount(),
		InvalidTicks:   reader.GetInvalidTickCount(),
		ParseErrors:    reader.GetParseErrorCount(),
		SuccessRate:    reader.getSuccessRate(),
	}
}

// String returns a human-readable string representation
func (rs *ReaderStatistics) String() string {
	return fmt.Sprintf(
		"CSV Reader Stats: %d ticks read, %.1f%% success, %d errors",
		rs.TicksRead,
		rs.SuccessRate,
		rs.ParseErrors,
	)
}

// DebugString returns detailed statistics
func (rs *ReaderStatistics) DebugString() string {
	return fmt.Sprintf(
		"Reader Statistics:\n"+
			"  File:            %s\n"+
			"  Ticks Read:      %d\n"+
			"  Lines Processed: %d\n"+
			"  Valid Ticks:     %d\n"+
			"  Invalid Ticks:   %d\n"+
			"  Parse Errors:    %d\n"+
			"  Success Rate:    %.1f%%",
		rs.FilePath,
		rs.TicksRead,
		rs.LinesProcessed,
		rs.ValidTicks,
		rs.InvalidTicks,
		rs.ParseErrors,
		rs.SuccessRate,
	)
}
