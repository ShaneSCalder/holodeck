# Instrument Package Documentation

## Overview

The `instrument` package provides comprehensive management of tradeable financial instruments across multiple asset classes (FOREX, STOCKS, COMMODITIES, CRYPTO). The package is organized into six focused files:

- **base.go** - Core Instrument type and utilities
- **forex.go** - FOREX instrument definitions
- **stocks.go** - STOCKS instrument definitions
- **commodities.go** - COMMODITIES instrument definitions
- **crypto.go** - CRYPTO instrument definitions
- **instrument.go** - Main package helpers and validation

## Package Structure

```
instrument/
├── base.go          (255 lines) - Core types & utilities
├── forex.go         (39 lines)  - FOREX asset class
├── stocks.go        (39 lines)  - STOCKS asset class
├── commodities.go   (39 lines)  - COMMODITIES asset class
├── crypto.go        (39 lines)  - CRYPTO asset class
└── instrument.go    (66 lines)  - Helpers & validation
```

**Total: 477 lines | 6 files | 4 asset classes**

---

## Base.go - Core Infrastructure

### Types

#### Instrument
The main instrument structure representing a tradeable financial product.

```go
type Instrument struct {
    // Identity
    Symbol      string  // e.g., "EUR/USD", "AAPL", "GOLD", "BTC/USD"
    Type        string  // FOREX, STOCKS, COMMODITIES, CRYPTO
    Description string  // Human-readable description
    Exchange    string  // Exchange name (FOREX, NYSE/NASDAQ, etc)

    // Price Configuration
    DecimalPlaces  int      // Number of decimal places (2-8)
    PipValue       float64  // Value of one pip (0.0001 for FOREX, 0.01 for stocks)
    TickSize       float64  // Minimum price movement
    ContractSize   int64    // Units per contract (100,000 for FOREX, 1 for stocks)
    MinimumLotSize float64  // Minimum tradeable lot

    // Trading Parameters
    Commission        float64  // Commission amount
    CommissionType    string   // per_million, per_share, per_lot, percentage
    Spread            float64  // Typical bid-ask spread
    MaxSpread         float64  // Maximum spread in adverse conditions
    MinSpread         float64  // Minimum spread in favorable conditions
    TradingDays       int      // Trading days per year
    AverageVolume     int64    // Average daily volume
    TypicalVolatility float64  // Annual volatility (0.10 = 10%)

    // Restrictions
    MinVolume float64  // Minimum trade size
    MaxVolume float64  // Maximum trade size
    MinPrice  float64  // Minimum price limit
    MaxPrice  float64  // Maximum price limit

    // Session Info
    OpenHour  int   // Market open hour (UTC)
    CloseHour int   // Market close hour (UTC)
    IsOpen    bool  // Is market currently open
}
```

#### InstrumentList
Container for managing multiple instruments.

```go
type InstrumentList struct {
    instruments map[string]*Instrument
}
```

### Type Constants

```go
const (
    TypeForex       = "FOREX"
    TypeStocks      = "STOCKS"
    TypeCommodities = "COMMODITIES"
    TypeCrypto      = "CRYPTO"
)
```

### Price Operations

#### RoundPrice
Rounds a price to the instrument's pip value.

```go
func (i *Instrument) RoundPrice(price float64) float64
```

**Parameters:**
- `price` - Price to round

**Returns:** Rounded price

**Example:**
```go
eurusd := NewForex("EUR/USD")
rounded := eurusd.RoundPrice(1.20456)  // Returns: 1.2046
```

#### FormatPrice
Formats price to the instrument's decimal places.

```go
func (i *Instrument) FormatPrice(price float64) string
```

**Returns:** Formatted price string

**Example:**
```go
aapl := NewStock("AAPL")
formatted := aapl.FormatPrice(150.50123)  // Returns: "150.50"
```

#### NormalizeLot
Normalizes a lot size to the minimum lot size.

```go
func (i *Instrument) NormalizeLot(lot float64) float64
```

**Example:**
```go
eurusd := NewForex("EUR/USD")
normalized := eurusd.NormalizeLot(0.05)  // Returns: 0.05 (min is 0.01)
```

### Validation Methods

#### IsValidVolume
Checks if a volume is within the instrument's limits.

```go
func (i *Instrument) IsValidVolume(volume float64) bool
```

**Returns:** `true` if volume is between MinVolume and MaxVolume

#### IsValidPrice
Checks if a price is within the instrument's limits.

```go
func (i *Instrument) IsValidPrice(price float64) bool
```

**Returns:** `true` if price is within MinPrice and MaxPrice limits

**Example:**
```go
stock := NewStock("AAPL")
if stock.IsValidPrice(150.00) {
    // Price is valid
}
```

### Type Check Methods

#### IsForex
```go
func (i *Instrument) IsForex() bool
```

#### IsStock
```go
func (i *Instrument) IsStock() bool
```

#### IsCommodity
```go
func (i *Instrument) IsCommodity() bool
```

#### IsCrypto
```go
func (i *Instrument) IsCrypto() bool
```

### Statistics Methods

#### GetVolatilityCategory
Classifies volatility into categories.

```go
func (i *Instrument) GetVolatilityCategory() string
```

**Returns:** "LOW", "MEDIUM", "HIGH", or "VERY_HIGH"

**Thresholds:**
- LOW: < 10% (0.10)
- MEDIUM: 10-20% (0.10-0.20)
- HIGH: 20-40% (0.20-0.40)
- VERY_HIGH: > 40% (0.40+)

**Example:**
```go
btc := NewCrypto("BTC/USD")
fmt.Println(btc.GetVolatilityCategory())  // Returns: "VERY_HIGH"
```

#### GetLiquidityCategory
Classifies liquidity based on average volume.

```go
func (i *Instrument) GetLiquidityCategory() string
```

**Returns:** "VERY_HIGH", "HIGH", "MEDIUM", or "LOW"

**Thresholds:**
- VERY_HIGH: > 5,000,000 volume
- HIGH: > 1,000,000 volume
- MEDIUM: > 100,000 volume
- LOW: <= 100,000 volume

### Position Sizing Methods

#### GetRiskAmount
Calculates the risk amount per pip movement.

```go
func (i *Instrument) GetRiskAmount(lotSize float64) float64
```

**Formula:** `lotSize × ContractSize × PipValue`

**Example:**
```go
eurusd := NewForex("EUR/USD")
risk := eurusd.GetRiskAmount(1.0)
// 1.0 lot × 100,000 contract × 0.0001 pip = 10.00
```

#### GetRequiredMargin
Calculates the required margin for a trade.

```go
func (i *Instrument) GetRequiredMargin(lotSize float64, leverage float64, price float64) float64
```

**Formula:** `(lotSize × ContractSize × price) / leverage`

**Example:**
```go
stock := NewStock("AAPL")
margin := stock.GetRequiredMargin(100, 2.0, 150.50)
// (100 × 1 × 150.50) / 2.0 = 7,525.00
```

### Display Methods

#### String
Returns a formatted one-line summary.

```go
func (i *Instrument) String() string
```

**Output Example:**
```
EUR/USD (FOREX) | Pip: 0.000100 | Spread: 0.000200 | Vol: 1000000 | Volatility: MEDIUM
```

#### Details
Returns detailed formatted information.

```go
func (i *Instrument) Details() string
```

**Output Example:**
```
Symbol:         EUR/USD
Type:           FOREX
Description:    Foreign Exchange: EUR/USD
Exchange:       FOREX
Decimals:       5
Pip Value:      0.000100
Tick Size:      0.000010
Contract Size:  100000
Min Lot:        0.010
Commission:     25.000000 (per_million)
Spread:         0.000200 (min: 0.000050, max: 0.000500)
Volume:         1000000
Volatility:     10.00% (MEDIUM)
Hours:          00:00 - 24:00 UTC
```

### InstrumentList Methods

#### NewInstrumentList
Creates a new empty instrument list.

```go
func NewInstrumentList() *InstrumentList
```

#### Add
Adds an instrument to the list.

```go
func (il *InstrumentList) Add(instrument *Instrument)
```

**Example:**
```go
list := NewInstrumentList()
list.Add(NewForex("EUR/USD"))
list.Add(NewStock("AAPL"))
```

#### Get
Retrieves an instrument by symbol (case-insensitive).

```go
func (il *InstrumentList) Get(symbol string) (*Instrument, bool)
```

**Returns:** Instrument and boolean indicating if found

**Example:**
```go
if eurusd, ok := list.Get("EUR/USD"); ok {
    fmt.Println(eurusd.String())
}
```

#### Remove
Removes an instrument from the list.

```go
func (il *InstrumentList) Remove(symbol string)
```

#### List
Returns all instruments in the list.

```go
func (il *InstrumentList) List() []*Instrument
```

#### Count
Returns the number of instruments.

```go
func (il *InstrumentList) Count() int
```

#### Contains
Checks if an instrument exists.

```go
func (il *InstrumentList) Contains(symbol string) bool
```

---

## Forex.go - FOREX Instruments

### Constructor

#### NewForex
Creates a FOREX instrument with standard market parameters.

```go
func NewForex(symbol string) *Instrument
```

**Parameters:**
- `symbol` - Currency pair (e.g., "EUR/USD", "GBP/USD")

**Returns:** Configured FOREX Instrument

**Standard Parameters:**
- DecimalPlaces: 5
- PipValue: 0.0001
- TickSize: 0.00001
- ContractSize: 100,000
- MinimumLotSize: 0.01
- Commission: 25 (per_million)
- Spread: 0.0002 (typical)
- MaxSpread: 0.0005 (wide)
- MinSpread: 0.00005 (tight)
- TradingDays: 252
- AverageVolume: 1,000,000
- TypicalVolatility: 0.10 (10%)
- Trading Hours: 00:00 - 24:00 UTC (24/5)

**Example:**
```go
eurusd := NewForex("EUR/USD")
gbpusd := NewForex("GBP/USD")
```

### Helper

#### ForexDefaults
Returns default FOREX parameters.

```go
func ForexDefaults() *Instrument
```

---

## Stocks.go - STOCKS Instruments

### Constructor

#### NewStock
Creates a STOCKS instrument with standard market parameters.

```go
func NewStock(symbol string) *Instrument
```

**Parameters:**
- `symbol` - Stock ticker (e.g., "AAPL", "GOOGL", "MSFT")

**Returns:** Configured STOCKS Instrument

**Standard Parameters:**
- DecimalPlaces: 2
- PipValue: 0.01
- TickSize: 0.01
- ContractSize: 1
- MinimumLotSize: 1.0
- Commission: 0.001 (0.1% percentage)
- Spread: 0.01 (typical)
- MaxSpread: 0.05 (wide)
- MinSpread: 0.001 (tight)
- TradingDays: 252
- AverageVolume: 1,000,000
- TypicalVolatility: 0.25 (25%)
- Trading Hours: 13:00 - 21:00 UTC (NYSE/NASDAQ hours)

**Example:**
```go
aapl := NewStock("AAPL")
googl := NewStock("GOOGL")
```

### Helper

#### StockDefaults
Returns default STOCKS parameters.

```go
func StockDefaults() *Instrument
```

---

## Commodities.go - COMMODITIES Instruments

### Constructor

#### NewCommodity
Creates a COMMODITIES instrument with standard market parameters.

```go
func NewCommodity(symbol string) *Instrument
```

**Parameters:**
- `symbol` - Commodity name (e.g., "GOLD", "CRUDE_OIL", "WHEAT")

**Returns:** Configured COMMODITIES Instrument

**Standard Parameters:**
- DecimalPlaces: 3
- PipValue: 0.01
- TickSize: 0.01
- ContractSize: 100
- MinimumLotSize: 0.1
- Commission: 50 (per_lot)
- Spread: 0.02 (typical)
- MaxSpread: 0.10 (wide)
- MinSpread: 0.01 (tight)
- TradingDays: 252
- AverageVolume: 500,000
- TypicalVolatility: 0.18 (18%)
- Trading Hours: 00:00 - 24:00 UTC

**Example:**
```go
gold := NewCommodity("GOLD")
oil := NewCommodity("CRUDE_OIL")
```

### Helper

#### CommodityDefaults
Returns default COMMODITIES parameters.

```go
func CommodityDefaults() *Instrument
```

---

## Crypto.go - CRYPTO Instruments

### Constructor

#### NewCrypto
Creates a CRYPTO instrument with standard market parameters.

```go
func NewCrypto(symbol string) *Instrument
```

**Parameters:**
- `symbol` - Cryptocurrency pair (e.g., "BTC/USD", "ETH/USD")

**Returns:** Configured CRYPTO Instrument

**Standard Parameters:**
- DecimalPlaces: 8
- PipValue: 0.00000001
- TickSize: 0.00000001
- ContractSize: 1
- MinimumLotSize: 0.001
- Commission: 0.001 (0.1% percentage)
- Spread: 0.0001 (typical)
- MaxSpread: 0.001 (wide)
- MinSpread: 0.00001 (tight)
- TradingDays: 365 (24/7 trading)
- AverageVolume: 1,000,000
- TypicalVolatility: 0.50 (50%)
- Trading Hours: 00:00 - 24:00 UTC (24/7)

**Example:**
```go
btc := NewCrypto("BTC/USD")
eth := NewCrypto("ETH/USD")
```

### Helper

#### CryptoDefaults
Returns default CRYPTO parameters.

```go
func CryptoDefaults() *Instrument
```

---

## Instrument.go - Main Helpers & Validation

### Helper Functions

#### GetInstrumentType
Returns the type of an instrument.

```go
func GetInstrumentType(instrument *Instrument) string
```

**Example:**
```go
eurusd := NewForex("EUR/USD")
fmt.Println(GetInstrumentType(eurusd))  // Returns: "FOREX"
```

#### IsValidInstrument
Checks if an instrument is properly configured.

```go
func IsValidInstrument(instrument *Instrument) bool
```

**Checks:**
- Instrument is not nil
- Symbol is not empty
- Type is not empty
- PipValue > 0
- ContractSize > 0

**Example:**
```go
inst := NewForex("EUR/USD")
if IsValidInstrument(inst) {
    fmt.Println("Valid instrument")
}
```

#### CompareInstruments
Compares two instruments for equality.

```go
func CompareInstruments(a, b *Instrument) bool
```

**Returns:** `true` if symbols, types, and exchanges match

**Example:**
```go
eurusd1 := NewForex("EUR/USD")
eurusd2 := NewForex("EUR/USD")
if CompareInstruments(eurusd1, eurusd2) {
    fmt.Println("Same instrument")
}
```

#### CreateCustomInstrument
Creates a custom instrument with provided parameters.

```go
func CreateCustomInstrument(symbol string, instrumentType string, decimals int,
    pipValue float64, tickSize float64, contractSize int64, minLot float64) *Instrument
```

**Example:**
```go
custom := CreateCustomInstrument(
    "INDEX",
    "CUSTOM",
    2,
    0.01,
    0.01,
    100,
    0.1,
)
```

---

## Complete Usage Examples

### Basic Asset Class Creation

```go
package main

import (
    "fmt"
    "holodeck/instrument"
)

func main() {
    // Create different asset classes
    eurusd := instrument.NewForex("EUR/USD")
    aapl := instrument.NewStock("AAPL")
    gold := instrument.NewCommodity("GOLD")
    btc := instrument.NewCrypto("BTC/USD")

    // Display information
    fmt.Println(eurusd)
    fmt.Println(aapl)
    fmt.Println(gold)
    fmt.Println(btc)
}
```

### Instrument Validation & Operations

```go
// Validate prices
stock := instrument.NewStock("AAPL")
if stock.IsValidPrice(150.00) {
    fmt.Println("Valid price")
}

// Validate volumes
eurusd := instrument.NewForex("EUR/USD")
if eurusd.IsValidVolume(10.0) {
    fmt.Println("Valid volume")
}

// Price operations
rounded := eurusd.RoundPrice(1.20456)
formatted := stock.FormatPrice(150.50123)
normalized := eurusd.NormalizeLot(0.05)
```

### Risk & Margin Calculations

```go
eurusd := instrument.NewForex("EUR/USD")

// Calculate risk per pip
riskPerPip := eurusd.GetRiskAmount(1.0)
fmt.Printf("Risk per pip: %.2f\n", riskPerPip)

// Calculate required margin (2x leverage)
stock := instrument.NewStock("AAPL")
requiredMargin := stock.GetRequiredMargin(100, 2.0, 150.50)
fmt.Printf("Required margin: %.2f\n", requiredMargin)
```

### Portfolio Management

```go
// Create instrument list
list := instrument.NewInstrumentList()

// Add instruments
list.Add(instrument.NewForex("EUR/USD"))
list.Add(instrument.NewForex("GBP/USD"))
list.Add(instrument.NewStock("AAPL"))
list.Add(instrument.NewStock("GOOGL"))
list.Add(instrument.NewCommodity("GOLD"))
list.Add(instrument.NewCrypto("BTC/USD"))

// Work with the list
fmt.Printf("Total instruments: %d\n", list.Count())

// Retrieve specific instrument
if eurusd, ok := list.Get("EUR/USD"); ok {
    fmt.Println(eurusd.Details())
}

// List all instruments
for _, inst := range list.List() {
    fmt.Printf("%s: %s\n", inst.Symbol, inst.Type)
}

// Check existence
if list.Contains("BTC/USD") {
    fmt.Println("BTC/USD is in portfolio")
}
```

### Asset Class Analysis

```go
// Analyze volatility and liquidity
btc := instrument.NewCrypto("BTC/USD")
fmt.Printf("Volatility: %s\n", btc.GetVolatilityCategory())
fmt.Printf("Liquidity: %s\n", btc.GetLiquidityCategory())

eurusd := instrument.NewForex("EUR/USD")
fmt.Printf("Volatility: %s\n", eurusd.GetVolatilityCategory())
fmt.Printf("Liquidity: %s\n", eurusd.GetLiquidityCategory())
```

### Custom Instruments

```go
// Create a custom index instrument
index := instrument.CreateCustomInstrument(
    "SPX",
    "INDEX",
    2,
    0.01,
    0.01,
    100,
    0.1,
)

if instrument.IsValidInstrument(index) {
    fmt.Println(index.Details())
}
```

---

## Asset Class Specifications

### FOREX
| Parameter | Value |
|-----------|-------|
| **Decimal Places** | 5 |
| **Pip Value** | 0.0001 |
| **Tick Size** | 0.00001 |
| **Contract Size** | 100,000 |
| **Min Lot** | 0.01 |
| **Commission** | 25 per million |
| **Typical Spread** | 0.0002 |
| **Trading Days/Year** | 252 |
| **Typical Volatility** | 10% |
| **Trading Hours** | 24/5 (Sunday-Friday) |

### STOCKS
| Parameter | Value |
|-----------|-------|
| **Decimal Places** | 2 |
| **Pip Value** | 0.01 |
| **Tick Size** | 0.01 |
| **Contract Size** | 1 |
| **Min Lot** | 1.0 |
| **Commission** | 0.1% (percentage) |
| **Typical Spread** | 0.01 |
| **Trading Days/Year** | 252 |
| **Typical Volatility** | 25% |
| **Trading Hours** | 13:00-21:00 UTC (NYSE/NASDAQ) |

### COMMODITIES
| Parameter | Value |
|-----------|-------|
| **Decimal Places** | 3 |
| **Pip Value** | 0.01 |
| **Tick Size** | 0.01 |
| **Contract Size** | 100 |
| **Min Lot** | 0.1 |
| **Commission** | 50 per lot |
| **Typical Spread** | 0.02 |
| **Trading Days/Year** | 252 |
| **Typical Volatility** | 18% |
| **Trading Hours** | 24/5 |

### CRYPTO
| Parameter | Value |
|-----------|-------|
| **Decimal Places** | 8 |
| **Pip Value** | 0.00000001 |
| **Tick Size** | 0.00000001 |
| **Contract Size** | 1 |
| **Min Lot** | 0.001 |
| **Commission** | 0.1% (percentage) |
| **Typical Spread** | 0.0001 |
| **Trading Days/Year** | 365 |
| **Typical Volatility** | 50% |
| **Trading Hours** | 24/7 |

---

## Best Practices

### 1. Always Validate Instruments
```go
if !instrument.IsValidInstrument(inst) {
    return fmt.Errorf("invalid instrument")
}
```

### 2. Use Appropriate Asset Class Constructor
```go
// Don't create generic instruments, use specific constructors
eurusd := instrument.NewForex("EUR/USD")  // ✅ Correct
stock := instrument.NewStock("AAPL")       // ✅ Correct
```

### 3. Round Prices to Pip Value
```go
rounded := inst.RoundPrice(price)  // Always round user input
```

### 4. Check Volume/Price Validity
```go
if inst.IsValidVolume(volume) && inst.IsValidPrice(price) {
    // Execute trade
}
```

### 5. Use InstrumentList for Portfolios
```go
portfolio := instrument.NewInstrumentList()
// Add instruments as needed
if inst, ok := portfolio.Get(symbol); ok {
    // Use instrument
}
```

### 6. Calculate Risk Before Trading
```go
risk := inst.GetRiskAmount(lotSize)
margin := inst.GetRequiredMargin(lotSize, leverage, price)
```

### 7. Understand Asset Class Characteristics
```go
// Check volatility and liquidity
volatility := inst.GetVolatilityCategory()
liquidity := inst.GetLiquidityCategory()

// Adjust strategy based on characteristics
if volatility == "VERY_HIGH" {
    // Use tighter stops
}
```

---

## Integration Points

### With Account Package
```go
// Account needs instrument specs for margin calculation
margin := instrument.GetRequiredMargin(lotSize, account.Leverage, price)
account.RecordMarginUsed(margin)
```

### With Position Package
```go
// Position uses instrument for P&L calculations
position := position.NewPosition("POS001", "EUR/USD", "LONG", 1.0, price)
position.Instrument = instrument

// Close position with instrument knowledge
pnl := position.Close(closePrice, commission)
```

### With Executor Package
```go
// Executor validates orders using instrument specs
if !instrument.IsValidPrice(price) {
    // Reject order
}
if !instrument.IsValidVolume(volume) {
    // Reject order
}
```

---

## Performance Considerations

- **Memory:** Instrument struct is ~400 bytes
- **Speed:** All operations are O(1)
- **InstrumentList:** O(1) for Add/Get/Remove, O(n) for List

---

## Error Handling

### Invalid Instrument
```go
if instrument == nil {
    return fmt.Errorf("instrument is nil")
}
if !IsValidInstrument(instrument) {
    return fmt.Errorf("invalid instrument configuration")
}
```

### Invalid Price/Volume
```go
if !inst.IsValidPrice(price) {
    return fmt.Errorf("price %.2f outside limits", price)
}
if !inst.IsValidVolume(volume) {
    return fmt.Errorf("volume %.2f outside limits", volume)
}
```

---

## Related Documentation

- [Account Package](account.md) - Account management
- [Position Package](position.md) - Position tracking
- [Executor Package](executor.md) - Order execution
- [Commission Package](commission.md) - Fee calculations

---

**Created:** December 26, 2024  
**Version:** 1.0  
**Status:** Production Ready  
**Lines of Code:** 477 (6 files)  
**Asset Classes:** 4 (FOREX, STOCKS, COMMODITIES, CRYPTO)