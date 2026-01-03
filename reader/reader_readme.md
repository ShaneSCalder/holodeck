# CSV Reader Package Documentation

The `reader` package provides a complete CSV tick data reading system for Holodeck. It handles loading market data from CSV files, parsing, validation, and provides multiple reading interfaces (sequential, batch, streaming).

---

## Table of Contents

1. [Overview](#overview)
2. [csv.go - Main Reader](#csvgo---main-reader)
3. [parser.go - Parsing Utilities](#parsergo---parsing-utilities)
4. [CSV Format](#csv-format)
5. [Usage Examples](#usage-examples)
6. [Error Handling](#error-handling)
7. [Performance Considerations](#performance-considerations)
8. [Architecture](#architecture)

---

## Overview

### Purpose

The `reader` package provides a flexible, robust system for reading market tick data from CSV files. It's designed to:

- Load data from disk efficiently
- Parse CSV rows into `types.Tick` objects
- Validate data quality
- Handle errors gracefully
- Support multiple reading patterns
- Track statistics

### Key Components

```
reader/
├── csv.go          # CSVTickReader, ParserConfig
└── parser.go       # Utilities, validators, batch/stream readers
```

### Core Interfaces

```go
// TickReader interface (implemented by CSVTickReader)
type TickReader interface {
    HasNext() bool
    Next() (*types.Tick, error)
    GetTickCount() int64
    Reset() error
    Close() error
}
```

---

## csv.go - Main Reader

### CSVTickReader Struct

```go
type CSVTickReader struct {
    filePath    string           // Path to CSV file
    file        *os.File         // Open file handle
    reader      *csv.Reader      // CSV reader
    tickCount   int64            // Total ticks read
    lineNumber  int64            // Current line number
    currentLine []string         // Current CSV row
    closed      bool             // Reader closed flag
    hasNext     bool             // More data available
    config      *ParserConfig    // Parse configuration
    validTicks   int64           // Successfully parsed ticks
    invalidTicks int64           // Rejected ticks
    parseErrors  int64           // Parse errors
}
```

### ParserConfig

Configuration for CSV parsing with sensible defaults:

```go
type ParserConfig struct {
    // Column indices (0-based)
    TimestampCol  int    // Which column has timestamp (default: 0)
    BidCol        int    // Which column has bid price (default: 1)
    AskCol        int    // Which column has ask price (default: 2)
    BidQtyCol     int    // Which column has bid quantity (default: 3)
    AskQtyCol     int    // Which column has ask quantity (default: 4)
    LastPriceCol  int    // Which column has last price (default: 5)
    VolumeCol     int    // Which column has volume (default: 6)

    // Parsing options
    TimestampFormat string // Timestamp format string (default: RFC3339Nano)
    SkipHeader      bool   // Skip first line if header (default: true)
    ValidateData    bool   // Validate each tick (default: true)
}
```

**Default Parser Config:**
```go
config := DefaultParserConfig()
// Creates config for standard format:
// timestamp,bid,ask,bid_qty,ask_qty,last_price,volume
```

### Constructor Functions

#### NewCSVTickReader

```go
func NewCSVTickReader(filePath string) (*CSVTickReader, error)
```

Creates a CSV reader with default configuration.

**Example:**
```go
reader, err := NewCSVTickReader("data/forex_eurusd.csv")
if err != nil {
    log.Fatal(err)  // File not found or I/O error
}
defer reader.Close()
```

**Errors:**
- File not found → `ConfigError("filePath", "CSV file not found")`
- File open error → `ConfigError("filePath", "failed to open CSV file")`
- Header read error → `ConfigError("csv", "failed to read header")`

#### NewCSVTickReaderWithConfig

```go
func NewCSVTickReaderWithConfig(filePath string, config *ParserConfig) (*CSVTickReader, error)
```

Creates a CSV reader with custom configuration.

**Example:**
```go
config := &ParserConfig{
    TimestampCol:    0,
    BidCol:          1,
    AskCol:          2,
    BidQtyCol:       3,
    AskQtyCol:       4,
    LastPriceCol:    5,
    VolumeCol:       6,
    TimestampFormat: "2006-01-02T15:04:05.000",
    SkipHeader:      true,
    ValidateData:    true,
}

reader, err := NewCSVTickReaderWithConfig("data.csv", config)
```

### Core Methods

#### HasNext

```go
func (ctr *CSVTickReader) HasNext() bool
```

Checks if there are more ticks to read.

**Returns:** `true` if more data available, `false` if EOF or reader closed

**Example:**
```go
for reader.HasNext() {
    tick, _ := reader.Next()
    // Process tick
}
```

#### Next

```go
func (ctr *CSVTickReader) Next() (*types.Tick, error)
```

Reads and parses the next tick from CSV file.

**Returns:**
- `*types.Tick` - Parsed tick data
- `error` - CSVReadError with line number and details

**Process:**
1. Read CSV row
2. Validate column count
3. Parse each field:
   - Timestamp with format detection
   - Prices as float64
   - Quantities as int64
4. Create `types.Tick` object
5. Optionally validate tick data
6. Return or error

**Example:**
```go
for reader.HasNext() {
    tick, err := reader.Next()
    if err != nil {
        fmt.Printf("Error on line %d: %v\n", reader.GetLineNumber(), err)
        continue
    }
    
    fmt.Printf("Tick: %s %.5f/%.5f Volume:%d\n",
               tick.Timestamp, tick.Bid, tick.Ask, tick.Volume)
}
```

**Possible Errors:**
- EOF → `io.EOF` (HasNext() becomes false)
- Invalid timestamp → `CSVReadError("invalid timestamp format")`
- Invalid price → `CSVReadError("invalid bid price")`
- Invalid quantity → `CSVReadError("invalid bid quantity")`
- Invalid tick → `CSVReadError("invalid tick data")`

#### Reset

```go
func (ctr *CSVTickReader) Reset() error
```

Resets reader to beginning of file (rewinds to start).

**Example:**
```go
// Read first 100 ticks
for i := 0; i < 100 && reader.HasNext(); i++ {
    reader.Next()
}

// Reset to beginning
if err := reader.Reset(); err != nil {
    log.Fatal(err)
}

// Read again
for reader.HasNext() {
    reader.Next()
}
```

**Use Cases:**
- Backtesting multiple strategies on same data
- Validation runs
- Data quality checks

#### Close

```go
func (ctr *CSVTickReader) Close() error
```

Closes the file and stops reading.

**Example:**
```go
reader, _ := NewCSVTickReader("data.csv")
defer reader.Close()  // Always close when done
```

### Query Methods

#### GetTickCount

```go
func (ctr *CSVTickReader) GetTickCount() int64
```

Returns total number of valid ticks read.

#### GetLineNumber

```go
func (ctr *CSVTickReader) GetLineNumber() int64
```

Returns current line number in file.

#### GetValidTickCount

```go
func (ctr *CSVTickReader) GetValidTickCount() int64
```

Returns number of successfully parsed ticks.

#### GetInvalidTickCount

```go
func (ctr *CSVTickReader) GetInvalidTickCount() int64
```

Returns number of ticks that failed validation.

#### GetParseErrorCount

```go
func (ctr *CSVTickReader) GetParseErrorCount() int64
```

Returns number of parse errors encountered.

#### IsClosed

```go
func (ctr *CSVTickReader) IsClosed() bool
```

Checks if reader is closed.

### Statistics & Diagnostics

#### GetStatistics

```go
func (ctr *CSVTickReader) GetStatistics() map[string]interface{}
```

Returns complete statistics map:

```go
stats := reader.GetStatistics()
// Returns:
// {
//   "file_path": "data/forex_eurusd.csv",
//   "ticks_read": 250000,
//   "lines_processed": 250001,
//   "valid_ticks": 250000,
//   "invalid_ticks": 1,
//   "parse_errors": 0,
//   "success_rate": 99.996,
//   "is_closed": false,
//   "has_next": false,
// }
```

#### String

```go
func (ctr *CSVTickReader) String() string
```

Returns concise summary:
```
CSVTickReader[File=data.csv, Ticks=250000, Valid=250000, Invalid=1, Success=99.9%]
```

#### DebugString

```go
func (ctr *CSVTickReader) DebugString() string
```

Returns detailed diagnostic information including configuration.

### Advanced Reading Methods

#### ReadN

```go
func (ctr *CSVTickReader) ReadN(n int) ([]*types.Tick, error)
```

Reads up to N ticks at once.

**Example:**
```go
ticks, err := reader.ReadN(1000)  // Read 1000 ticks
if err != nil {
    log.Fatal(err)
}
```

#### ReadUntil

```go
func (ctr *CSVTickReader) ReadUntil(maxTime time.Time) ([]*types.Tick, error)
```

Reads ticks until a specific time is reached.

**Example:**
```go
endTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
ticks, err := reader.ReadUntil(endTime)  // Read until noon
```

---

## parser.go - Parsing Utilities

### CSV Format Detection

#### DetectTimestampFormat

```go
func DetectTimestampFormat(sample string) string
```

Automatically detects timestamp format from a sample string.

**Supported Formats:**
- RFC3339Nano: `2024-01-15T07:00:00.000000000Z`
- RFC3339: `2024-01-15T07:00:00Z`
- ISO8601: `2024-01-15T07:00:00.000`
- Datetime: `2024-01-15 07:00:00.000`
- Date only: `2024-01-15`
- US format: `01/15/2024 07:00:00`

**Example:**
```go
format := DetectTimestampFormat("2024-01-15T07:00:00.000Z")
// Returns: "2006-01-02T15:04:05.000Z"
```

### CSV Validation

#### ValidateCSVHeader

```go
func ValidateCSVHeader(header []string, expectedColumns []string) error
```

Validates that CSV header contains all expected columns.

**Example:**
```go
header := []string{"timestamp", "bid", "ask", "bid_qty", "ask_qty", "last_price", "volume"}
expected := []string{"bid", "ask", "timestamp"}

err := ValidateCSVHeader(header, expected)
// Returns nil (all expected columns present)
```

### Tick Validation

#### TickValidator

Validates tick data quality with configurable rules:

```go
type TickValidator struct {
    minBid       float64
    maxBid       float64
    minAsk       float64
    maxAsk       float64
    maxSpread    float64
    requireDepth bool
}
```

**Creating a Validator:**

```go
validator := NewTickValidator().
    WithBidRange(0.00001, 100000).
    WithAskRange(0.00001, 100000).
    WithMaxSpread(1.0).
    WithDepthRequirement(false)

err := validator.ValidateTick(tick)
```

**Builder Methods:**

```go
validator := NewTickValidator()
validator.WithBidRange(minBid, maxBid)        // Set bid range
validator.WithAskRange(minAsk, maxAsk)        // Set ask range
validator.WithMaxSpread(maxSpread)            // Max bid-ask spread
validator.WithDepthRequirement(required)      // Require depth
```

**Validations Performed:**
- ✅ Bid in acceptable range
- ✅ Ask in acceptable range
- ✅ Spread within limit
- ✅ Sufficient depth (if required)

### Column Auto-Detection

#### AutodetectColumns

```go
func AutodetectColumns(header []string) (*ParserConfig, error)
```

Automatically detects CSV columns by name.

**Supported Column Names:**

| Data | Names |
|------|-------|
| Timestamp | timestamp, time, date, datetime |
| Bid | bid, bid_price |
| Ask | ask, ask_price |
| Bid Qty | bid_qty, bid_quantity, bid_size |
| Ask Qty | ask_qty, ask_quantity, ask_size |
| Last Price | last, last_price, price |
| Volume | volume, vol, qty, size |

**Example:**
```go
header := []string{"timestamp", "bid", "ask", "bid_qty", "ask_qty", "price", "volume"}
config, err := AutodetectColumns(header)
// Returns ParserConfig with columns auto-detected
```

### Batch Reading

#### BatchReader

Reads ticks in fixed-size batches for efficiency:

```go
type BatchReader struct {
    reader    *CSVTickReader
    batchSize int
}
```

**Creating a Batch Reader:**

```go
batchReader := NewBatchReader(reader, 100)  // 100 ticks per batch
```

**Reading Batches:**

```go
// Read one batch
batch, err := batchReader.ReadBatch()
if err != nil {
    log.Fatal(err)
}

// Read all remaining batches
allBatches, err := batchReader.ReadAllBatches()
for _, batch := range allBatches {
    for _, tick := range batch {
        // Process tick
    }
}
```

**Use Cases:**
- Processing large files without memory overload
- Batch processing with parallel workers
- Database bulk inserts

### Streaming Reader

#### StreamingReader

Provides channel-based streaming interface:

```go
type StreamingReader struct {
    reader  *CSVTickReader
    tickCh  chan *types.Tick  // Tick output channel
    errCh   chan error         // Error output channel
    done    chan bool
    stopCh  chan bool
}
```

**Creating a Streaming Reader:**

```go
streamReader := NewStreamingReader(csvReader)
streamReader.Start()  // Begin streaming
```

**Reading from Stream:**

```go
streamReader := NewStreamingReader(reader)
streamReader.Start()

go func() {
    for {
        select {
        case tick, ok := <-streamReader.GetTicks():
            if !ok {
                return  // Channel closed
            }
            fmt.Printf("Tick: %s\n", tick)
            
        case err := <-streamReader.GetErrors():
            fmt.Printf("Error: %v\n", err)
        }
    }
}()

// ...later...
streamReader.Stop()
```

**Benefits:**
- Non-blocking I/O
- Decoupled processing
- Parallel processing
- Buffered channels prevent blocking

### Statistics

#### ReaderStatistics

```go
type ReaderStatistics struct {
    FilePath       string
    TicksRead      int64
    LinesProcessed int64
    ValidTicks     int64
    InvalidTicks   int64
    ParseErrors    int64
    SuccessRate    float64
}
```

**Getting Statistics:**

```go
stats := GetReaderStatistics(reader)
fmt.Printf("Ticks: %d, Success: %.1f%%\n", 
           stats.TicksRead, stats.SuccessRate)
```

---

## CSV Format

### Standard Format

Expected default CSV format with 7 columns:

```
timestamp,bid,ask,bid_qty,ask_qty,last_price,volume
2024-01-15T07:00:00.000Z,1.08500,1.08502,500000,450000,1.08501,1000000
2024-01-15T07:00:01.000Z,1.08501,1.08503,500000,450000,1.08502,1000000
2024-01-15T07:00:02.000Z,1.08502,1.08504,500000,450000,1.08503,1000000
```

### Column Definitions

| Column | Type | Range | Example |
|--------|------|-------|---------|
| timestamp | string | Any valid format | 2024-01-15T07:00:00.000Z |
| bid | float64 | > 0 | 1.08500 |
| ask | float64 | > bid | 1.08502 |
| bid_qty | int64 | > 0 | 500000 |
| ask_qty | int64 | > 0 | 450000 |
| last_price | float64 | Bid to Ask | 1.08501 |
| volume | int64 | >= 0 | 1000000 |

### Custom Format

To use a different CSV format, create custom ParserConfig:

```go
// If columns are in different order
config := &ParserConfig{
    TimestampCol:    0,
    BidCol:          2,      // Bid in column 2
    AskCol:          3,      // Ask in column 3
    BidQtyCol:       4,
    AskQtyCol:       5,
    LastPriceCol:    1,      // Last price in column 1
    VolumeCol:       6,
    TimestampFormat: "2006-01-02 15:04:05.000",  // Custom format
    SkipHeader:      true,
    ValidateData:    true,
}

reader, _ := NewCSVTickReaderWithConfig("data.csv", config)
```

### Auto-Detect Format

If header row contains standard column names:

```go
file, _ := os.Open("data.csv")
reader := csv.NewReader(file)
header, _ := reader.Read()

config, _ := AutodetectColumns(header)
// config now has correct column positions

csvReader, _ := NewCSVTickReaderWithConfig("data.csv", config)
```

---

## Usage Examples

### Example 1: Basic Reading

```go
package main

import (
    "fmt"
    "log"
    "reader"
)

func main() {
    // Create reader
    reader, err := reader.NewCSVTickReader("data/forex_eurusd.csv")
    if err != nil {
        log.Fatal(err)
    }
    defer reader.Close()

    // Read all ticks
    count := 0
    for reader.HasNext() {
        tick, err := reader.Next()
        if err != nil {
            fmt.Printf("Error at line %d: %v\n", 
                       reader.GetLineNumber(), err)
            continue
        }
        
        fmt.Printf("Tick %d: %s %.5f/%.5f Vol:%d\n",
                   count+1,
                   tick.Timestamp.Format("15:04:05"),
                   tick.Bid, tick.Ask, tick.Volume)
        count++
    }

    // Print statistics
    stats := reader.GetStatistics()
    fmt.Printf("\nTotal Ticks: %d\n", stats["ticks_read"])
    fmt.Printf("Success Rate: %.1f%%\n", stats["success_rate"])
}
```

### Example 2: Batch Processing

```go
reader, _ := reader.NewCSVTickReader("data.csv")
defer reader.Close()

batchReader := reader.NewBatchReader(reader, 100)

for {
    batch, err := batchReader.ReadBatch()
    if err != nil {
        break
    }
    
    // Process entire batch at once
    processBatch(batch)
}
```

### Example 3: Streaming with Goroutines

```go
reader, _ := reader.NewCSVTickReader("data.csv")
defer reader.Close()

streamReader := reader.NewStreamingReader(reader)
streamReader.Start()

// Process ticks in main goroutine
tickCount := 0
for {
    select {
    case tick := <-streamReader.GetTicks():
        if tick == nil {
            return
        }
        processTick(tick)
        tickCount++
        
    case err := <-streamReader.GetErrors():
        if err != nil {
            log.Printf("Error: %v", err)
        }
    }
}
```

### Example 4: Custom Configuration

```go
config := &reader.ParserConfig{
    TimestampCol:    0,
    BidCol:          1,
    AskCol:          2,
    BidQtyCol:       3,
    AskQtyCol:       4,
    LastPriceCol:    5,
    VolumeCol:       6,
    TimestampFormat: "2006-01-02T15:04:05.000",
    SkipHeader:      true,
    ValidateData:    true,
}

reader, _ := reader.NewCSVTickReaderWithConfig("data.csv", config)

// Use reader...
```

### Example 5: Data Validation

```go
reader, _ := reader.NewCSVTickReader("data.csv")
defer reader.Close()

validator := reader.NewTickValidator().
    WithBidRange(1.0, 2.0).
    WithAskRange(1.0, 2.0).
    WithMaxSpread(0.01).
    WithDepthRequirement(true)

badTicks := 0
for reader.HasNext() {
    tick, _ := reader.Next()
    
    if err := validator.ValidateTick(tick); err != nil {
        fmt.Printf("Invalid: %v\n", err)
        badTicks++
        continue
    }
}

fmt.Printf("Bad ticks: %d\n", badTicks)
```

### Example 6: Reset and Re-read

```go
reader, _ := reader.NewCSVTickReader("data.csv")
defer reader.Close()

// First pass
count1 := 0
for reader.HasNext() {
    reader.Next()
    count1++
}

// Reset
reader.Reset()

// Second pass
count2 := 0
for reader.HasNext() {
    reader.Next()
    count2++
}

fmt.Printf("First: %d, Second: %d\n", count1, count2)
```

---

## Error Handling

### Error Types

All errors are `types.HolodeckError`:

```go
type HolodeckError struct {
    Code        string
    Message     string
    Details     map[string]interface{}
    Timestamp   time.Time
    SourceFunc  string
    SourceFile  string
    SourceLine  int
}
```

### Common Errors

#### File Not Found

```go
reader, err := NewCSVTickReader("missing.csv")
// Error: ConfigError
// Code: "filePath"
// Message: "CSV file not found: missing.csv"
```

**Handling:**
```go
if herr, ok := err.(*types.HolodeckError); ok {
    if herr.Code == "filePath" {
        fmt.Println("File missing, check path")
    }
}
```

#### Invalid Timestamp

```go
// Error on line 5:
// Code: "csv"
// Message: "invalid timestamp format"
// Details: {timestamp: "invalid", expected: "RFC3339Nano"}
```

**Handling:**
```go
for reader.HasNext() {
    tick, err := reader.Next()
    if herr, ok := err.(*types.HolodeckError); ok {
        line := reader.GetLineNumber()
        fmt.Printf("Line %d: %s\n", line, herr.Message)
        continue
    }
}
```

#### Invalid Tick Data

```go
// Code: "csv"
// Message: "invalid tick data: bid > ask"
// Details: {bid: 1.08502, ask: 1.08500}
```

**Handling:**
```go
err := validator.ValidateTick(tick)
if err != nil {
    fmt.Printf("Validation failed: %v\n", err)
    // Continue with next tick or stop
}
```

---

## Performance Considerations

### Memory Usage

**Memory-efficient approaches:**

1. **Sequential Reading** (Lowest memory)
```go
for reader.HasNext() {
    tick, _ := reader.Next()
    // Process immediately, discard
    // Only 1 tick in memory at a time
}
```

2. **Batch Reading** (Moderate memory)
```go
batchReader := NewBatchReader(reader, 1000)
batch, _ := batchReader.ReadBatch()
// 1000 ticks in memory
```

3. **Streaming** (Good for parallelization)
```go
streamReader := NewStreamingReader(reader)
// Buffered channel (default 100), background goroutine
```

4. **Load All** (Highest memory)
```go
allTicks, _ := reader.ReadUntil(time.Now())
// All ticks in memory at once
```

### Speed Benchmarks

(Approximate, depends on system)

- Sequential: ~50,000 ticks/sec
- Batch (1000): ~80,000 ticks/sec
- Streaming: ~70,000 ticks/sec

**Recommendation:** For files > 1M ticks, use batch or streaming mode.

### Optimization Tips

1. **Disable validation** if data is trusted:
   ```go
   config := DefaultParserConfig()
   config.ValidateData = false
   ```

2. **Use batch reading** for parallel processing:
   ```go
   batchReader := NewBatchReader(reader, 10000)
   ```

3. **Pre-detect timestamp format** if known:
   ```go
   config := DefaultParserConfig()
   config.TimestampFormat = "2006-01-02T15:04:05.000Z"
   ```

---

## Architecture

### Data Flow

```
CSV File
   ↓
os.Open()
   ↓
csv.Reader
   ↓
CSVTickReader.Next()
   ├─ Read line
   ├─ Parse columns
   ├─ Validate data (optional)
   └─ Create Tick
   ↓
types.Tick
```

### Error Propagation

```
CSV Format Error
   ↓
types.CSVReadError
   └─ Line number
   └─ Column info
   └─ Expected vs actual
   ↓
Application Error Handler
```

### Thread Safety

**CSVTickReader is NOT thread-safe:**
- Do not call from multiple goroutines simultaneously
- Use separate reader instances for parallel processing
- Or use StreamingReader with multiple consumers on channel

**StreamingReader is thread-safe:**
- Channels are goroutine-safe
- Can have multiple readers from same stream

---

## Summary

The `reader` package provides:

✅ **Core Functionality:**
- CSV file reading with proper file handling
- Flexible column ordering
- Multiple timestamp formats
- Data validation

✅ **Multiple Reading Patterns:**
- Sequential: Simple loop
- Batch: Process in groups
- Streaming: Channel-based async

✅ **Robustness:**
- Detailed error messages with line numbers
- Data quality validation
- Configuration flexibility
- Statistics and diagnostics

✅ **Performance:**
- Memory-efficient sequential reading
- Batch mode for parallel processing
- Streaming for async patterns

The reader is ready for integration with Holodeck's main execution system and supports all CSV-based data sources for backtesting.