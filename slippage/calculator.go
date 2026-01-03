package slippage

import (
	"fmt"

	"holodeck/types"
)

// ==================== SLIPPAGE CALCULATOR ====================

// SlippageCalculator orchestrates slippage calculation using depth and momentum models
type SlippageCalculator struct {
	depthModel    *DepthModel
	momentumModel *MomentumModel

	// Statistics
	totalSlippage      float64
	slippageCount      int64
	totalSlippageUnits float64
	maxSlippage        float64
	minSlippage        float64
}

// ==================== CALCULATOR CREATION ====================

// NewSlippageCalculator creates a new slippage calculator
func NewSlippageCalculator() *SlippageCalculator {
	return &SlippageCalculator{
		depthModel:    NewDepthModel(),
		momentumModel: NewMomentumModel(),
		minSlippage:   1e9, // Initialize to large value
	}
}

// ==================== CORE CALCULATION ====================

// CalculateSlippage calculates slippage based on order size and available depth
// Parameters:
//   - orderSize: Size of the order
//   - availableDepth: Available depth at bid/ask
//   - volatility: Market volatility (0.0 to 1.0+)
//   - momentum: Price momentum multiplier (default 1.0)
//   - tick: Market tick for context
//   - instrument: Instrument being traded
//
// Returns: Slippage in pips/units
func (sc *SlippageCalculator) CalculateSlippage(
	orderSize float64,
	availableDepth float64,
	volatility float64,
	momentum float64,
	tick *types.Tick,
	instrument types.Instrument,
) (float64, error) {

	if tick == nil {
		return 0, types.NewOrderRejectedError("tick cannot be nil")
	}

	if instrument == nil {
		return 0, types.NewOrderRejectedError("instrument cannot be nil")
	}

	// Calculate depth-based slippage
	depthSlippage, err := sc.depthModel.CalculateSlippage(orderSize, availableDepth, volatility)
	if err != nil {
		return 0, err
	}

	// Apply momentum adjustment
	adjustedSlippage, err := sc.momentumModel.AdjustSlippage(depthSlippage, momentum, tick)
	if err != nil {
		return 0, err
	}

	// Track statistics
	sc.totalSlippage += adjustedSlippage
	sc.slippageCount++
	sc.totalSlippageUnits += adjustedSlippage
	if adjustedSlippage > sc.maxSlippage {
		sc.maxSlippage = adjustedSlippage
	}
	if adjustedSlippage < sc.minSlippage {
		sc.minSlippage = adjustedSlippage
	}

	return adjustedSlippage, nil
}

// CalculateFillPrice calculates the fill price accounting for slippage
// Parameters:
//   - midPrice: Mid-market price (bid + ask) / 2
//   - slippageUnits: Amount of slippage in pips/units
//   - side: BUY or SELL
//   - instrument: Instrument being traded
//
// Returns: Fill price
func (sc *SlippageCalculator) CalculateFillPrice(
	midPrice float64,
	slippageUnits float64,
	side string,
	instrument types.Instrument,
) (float64, error) {

	if instrument == nil {
		return 0, types.NewOrderRejectedError("instrument cannot be nil")
	}

	// Get pip value
	pipValue := instrument.GetPipValue()

	// Calculate slippage in price units
	slippagePrice := slippageUnits * pipValue

	// Adjust price based on side
	fillPrice := midPrice
	if side == "BUY" {
		// Slippage increases the price we pay
		fillPrice += slippagePrice
	} else if side == "SELL" {
		// Slippage decreases the price we receive
		fillPrice -= slippagePrice
	}

	return fillPrice, nil
}

// CalculateBatchSlippage calculates slippage for multiple orders
func (sc *SlippageCalculator) CalculateBatchSlippage(
	orders []SlippageInput,
	tick *types.Tick,
	instrument types.Instrument,
) (float64, error) {

	totalSlippage := 0.0

	for _, order := range orders {
		slippage, err := sc.CalculateSlippage(
			order.OrderSize,
			order.AvailableDepth,
			order.Volatility,
			order.Momentum,
			tick,
			instrument,
		)
		if err != nil {
			return 0, err
		}
		totalSlippage += slippage
	}

	return totalSlippage, nil
}

// ==================== STATISTICS ====================

// GetTotalSlippage returns total slippage accumulated
func (sc *SlippageCalculator) GetTotalSlippage() float64 {
	return sc.totalSlippage
}

// GetSlippageCount returns number of slippage calculations
func (sc *SlippageCalculator) GetSlippageCount() int64 {
	return sc.slippageCount
}

// GetAverageSlippage returns average slippage per trade
func (sc *SlippageCalculator) GetAverageSlippage() float64 {
	if sc.slippageCount == 0 {
		return 0
	}
	return sc.totalSlippage / float64(sc.slippageCount)
}

// GetMaxSlippage returns maximum slippage observed
func (sc *SlippageCalculator) GetMaxSlippage() float64 {
	if sc.slippageCount == 0 {
		return 0
	}
	return sc.maxSlippage
}

// GetMinSlippage returns minimum slippage observed
func (sc *SlippageCalculator) GetMinSlippage() float64 {
	if sc.slippageCount == 0 {
		return 0
	}
	return sc.minSlippage
}

// GetStatistics returns comprehensive slippage statistics
func (sc *SlippageCalculator) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"total_slippage":       sc.totalSlippage,
		"slippage_count":       sc.slippageCount,
		"average_slippage":     sc.GetAverageSlippage(),
		"max_slippage":         sc.GetMaxSlippage(),
		"min_slippage":         sc.GetMinSlippage(),
		"depth_model_stats":    sc.depthModel.GetStatistics(),
		"momentum_model_stats": sc.momentumModel.GetStatistics(),
	}
}

// ==================== DEBUG ====================

// String returns a human-readable representation
func (sc *SlippageCalculator) String() string {
	return fmt.Sprintf(
		"SlippageCalculator[Total:%.4f, Count:%d, Avg:%.4f, Max:%.4f]",
		sc.totalSlippage,
		sc.slippageCount,
		sc.GetAverageSlippage(),
		sc.GetMaxSlippage(),
	)
}

// DebugString returns detailed debug information
func (sc *SlippageCalculator) DebugString() string {
	return fmt.Sprintf(
		"Slippage Calculator:\n"+
			"  Total Slippage:        %.4f pips\n"+
			"  Slippage Count:        %d\n"+
			"  Average Slippage:      %.4f pips\n"+
			"  Max Slippage:          %.4f pips\n"+
			"  Min Slippage:          %.4f pips\n"+
			"\n"+
			"  Sub-models:\n"+
			"    Depth Model:         %s\n"+
			"    Momentum Model:      %s",
		sc.totalSlippage,
		sc.slippageCount,
		sc.GetAverageSlippage(),
		sc.GetMaxSlippage(),
		sc.GetMinSlippage(),
		sc.depthModel.String(),
		sc.momentumModel.String(),
	)
}

// Reset resets calculator statistics
func (sc *SlippageCalculator) Reset() {
	sc.totalSlippage = 0
	sc.slippageCount = 0
	sc.totalSlippageUnits = 0
	sc.maxSlippage = 0
	sc.minSlippage = 1e9
	sc.depthModel.Reset()
	sc.momentumModel.Reset()
}

// ==================== SLIPPAGE INPUT ====================

// SlippageInput represents input for slippage calculation
type SlippageInput struct {
	OrderSize      float64 // Size of order
	AvailableDepth float64 // Available depth at market
	Volatility     float64 // Market volatility
	Momentum       float64 // Price momentum multiplier
}

// ==================== SLIPPAGE ANALYSIS ====================

// SlippageAnalysis provides detailed breakdown of a slippage calculation
type SlippageAnalysis struct {
	OrderSize        float64
	AvailableDepth   float64
	Volatility       float64
	Momentum         float64
	DepthSlippage    float64
	AdjustedSlippage float64
	MidPrice         float64
	BuyFillPrice     float64
	SellFillPrice    float64
}

// String returns string representation
func (sa *SlippageAnalysis) String() string {
	return fmt.Sprintf(
		"Size:%.4f Depth:%.4f Vol:%.4f Mom:%.4f => Slippage:%.4f pips",
		sa.OrderSize,
		sa.AvailableDepth,
		sa.Volatility,
		sa.Momentum,
		sa.AdjustedSlippage,
	)
}

// DebugString returns detailed debug information
func (sa *SlippageAnalysis) DebugString() string {
	return fmt.Sprintf(
		"Slippage Analysis:\n"+
			"  Order Size:            %.4f\n"+
			"  Available Depth:       %.4f\n"+
			"  Volatility:            %.4f\n"+
			"  Momentum:              %.4f\n"+
			"  Depth Slippage:        %.4f pips\n"+
			"  Adjusted Slippage:     %.4f pips\n"+
			"  Mid Price:             %.5f\n"+
			"  Buy Fill Price:        %.5f\n"+
			"  Sell Fill Price:       %.5f",
		sa.OrderSize,
		sa.AvailableDepth,
		sa.Volatility,
		sa.Momentum,
		sa.DepthSlippage,
		sa.AdjustedSlippage,
		sa.MidPrice,
		sa.BuyFillPrice,
		sa.SellFillPrice,
	)
}
