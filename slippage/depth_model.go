package slippage

import (
	"fmt"
)

// ==================== DEPTH MODEL ====================

// DepthModel calculates slippage based on order size relative to available depth
// Formula: slippage = (order_size / available_depth) × volatility
type DepthModel struct {
	// Statistics
	totalSlippage float64
	slippageCount int64
	avgDepthRatio float64
	maxDepthRatio float64
	minDepthRatio float64
}

// ==================== MODEL CREATION ====================

// NewDepthModel creates a new depth model
func NewDepthModel() *DepthModel {
	return &DepthModel{
		minDepthRatio: 1e9,
	}
}

// ==================== CORE CALCULATION ====================

// CalculateSlippage calculates slippage based on depth
// Formula: slippage = (order_size / available_depth) × volatility
// Parameters:
//   - orderSize: Size of the order
//   - availableDepth: Available depth at bid/ask
//   - volatility: Market volatility (0.0 to 1.0+)
//
// Returns: Slippage in pips/units
func (dm *DepthModel) CalculateSlippage(
	orderSize float64,
	availableDepth float64,
	volatility float64,
) (float64, error) {

	// Prevent division by zero
	if availableDepth <= 0 {
		availableDepth = 0.001 // Minimum depth
	}

	// Calculate depth ratio
	depthRatio := orderSize / availableDepth

	// Calculate slippage: depth_ratio × volatility
	slippage := depthRatio * volatility

	// Track statistics
	dm.totalSlippage += slippage
	dm.slippageCount++
	dm.avgDepthRatio = (dm.avgDepthRatio*(float64(dm.slippageCount-1)) + depthRatio) / float64(dm.slippageCount)
	if depthRatio > dm.maxDepthRatio {
		dm.maxDepthRatio = depthRatio
	}
	if depthRatio < dm.minDepthRatio {
		dm.minDepthRatio = depthRatio
	}

	return slippage, nil
}

// ==================== STATISTICS ====================

// GetTotalSlippage returns total slippage from depth model
func (dm *DepthModel) GetTotalSlippage() float64 {
	return dm.totalSlippage
}

// GetSlippageCount returns number of slippage calculations
func (dm *DepthModel) GetSlippageCount() int64 {
	return dm.slippageCount
}

// GetAverageSlippage returns average slippage
func (dm *DepthModel) GetAverageSlippage() float64 {
	if dm.slippageCount == 0 {
		return 0
	}
	return dm.totalSlippage / float64(dm.slippageCount)
}

// GetAverageDepthRatio returns average depth ratio
func (dm *DepthModel) GetAverageDepthRatio() float64 {
	return dm.avgDepthRatio
}

// GetMaxDepthRatio returns maximum depth ratio
func (dm *DepthModel) GetMaxDepthRatio() float64 {
	if dm.slippageCount == 0 {
		return 0
	}
	return dm.maxDepthRatio
}

// GetMinDepthRatio returns minimum depth ratio
func (dm *DepthModel) GetMinDepthRatio() float64 {
	if dm.slippageCount == 0 {
		return 0
	}
	return dm.minDepthRatio
}

// GetStatistics returns comprehensive depth model statistics
func (dm *DepthModel) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"total_slippage":   dm.totalSlippage,
		"slippage_count":   dm.slippageCount,
		"average_slippage": dm.GetAverageSlippage(),
		"avg_depth_ratio":  dm.avgDepthRatio,
		"max_depth_ratio":  dm.GetMaxDepthRatio(),
		"min_depth_ratio":  dm.GetMinDepthRatio(),
	}
}

// ==================== DEBUG ====================

// String returns a human-readable representation
func (dm *DepthModel) String() string {
	return fmt.Sprintf(
		"DepthModel[Total:%.4f, Avg:%.4f, DepthRatio:%.4f]",
		dm.totalSlippage,
		dm.GetAverageSlippage(),
		dm.avgDepthRatio,
	)
}

// DebugString returns detailed debug information
func (dm *DepthModel) DebugString() string {
	return fmt.Sprintf(
		"Depth Model:\n"+
			"  Total Slippage:        %.4f pips\n"+
			"  Slippage Count:        %d\n"+
			"  Average Slippage:      %.4f pips\n"+
			"  Avg Depth Ratio:       %.4f\n"+
			"  Max Depth Ratio:       %.4f\n"+
			"  Min Depth Ratio:       %.4f",
		dm.totalSlippage,
		dm.slippageCount,
		dm.GetAverageSlippage(),
		dm.avgDepthRatio,
		dm.GetMaxDepthRatio(),
		dm.GetMinDepthRatio(),
	)
}

// Reset resets model statistics
func (dm *DepthModel) Reset() {
	dm.totalSlippage = 0
	dm.slippageCount = 0
	dm.avgDepthRatio = 0
	dm.maxDepthRatio = 0
	dm.minDepthRatio = 1e9
}

// ==================== ANALYSIS ====================

// AnalyzeDepthSlippage provides detailed analysis of depth slippage
func (dm *DepthModel) AnalyzeDepthSlippage(
	orderSize float64,
	availableDepth float64,
	volatility float64,
) *DepthSlippageAnalysis {

	if availableDepth <= 0 {
		availableDepth = 0.001
	}

	depthRatio := orderSize / availableDepth
	slippage := depthRatio * volatility

	return &DepthSlippageAnalysis{
		OrderSize:      orderSize,
		AvailableDepth: availableDepth,
		DepthRatio:     depthRatio,
		Volatility:     volatility,
		Slippage:       slippage,
	}
}

// ==================== ANALYSIS TYPES ====================

// DepthSlippageAnalysis provides detailed breakdown of depth slippage
type DepthSlippageAnalysis struct {
	OrderSize      float64
	AvailableDepth float64
	DepthRatio     float64
	Volatility     float64
	Slippage       float64
}

// String returns string representation
func (dsa *DepthSlippageAnalysis) String() string {
	return fmt.Sprintf(
		"Size:%.4f Depth:%.4f Ratio:%.4f Vol:%.4f => Slippage:%.4f",
		dsa.OrderSize,
		dsa.AvailableDepth,
		dsa.DepthRatio,
		dsa.Volatility,
		dsa.Slippage,
	)
}

// DebugString returns detailed debug information
func (dsa *DepthSlippageAnalysis) DebugString() string {
	return fmt.Sprintf(
		"Depth Slippage Analysis:\n"+
			"  Order Size:            %.4f\n"+
			"  Available Depth:       %.4f\n"+
			"  Depth Ratio:           %.4f\n"+
			"  Volatility:            %.4f\n"+
			"  Slippage:              %.4f pips",
		dsa.OrderSize,
		dsa.AvailableDepth,
		dsa.DepthRatio,
		dsa.Volatility,
		dsa.Slippage,
	)
}

// ==================== UTILITY FUNCTIONS ====================

// EstimateSlippageForSize estimates slippage for a given order size
func (dm *DepthModel) EstimateSlippageForSize(
	orderSize float64,
	currentDepthRatio float64,
	currentVolatility float64,
) float64 {
	// Estimate based on current conditions
	return orderSize * currentVolatility * currentDepthRatio
}

// CalculateDepthRequired calculates depth required for target slippage
func (dm *DepthModel) CalculateDepthRequired(
	orderSize float64,
	targetSlippage float64,
	volatility float64,
) float64 {
	// Reverse formula: depth = order_size × volatility / target_slippage
	if targetSlippage <= 0 || volatility <= 0 {
		return orderSize // Return order size if invalid params
	}
	return (orderSize * volatility) / targetSlippage
}

// CalculateMaxOrderSize calculates maximum order size for target slippage
func (dm *DepthModel) CalculateMaxOrderSize(
	availableDepth float64,
	targetSlippage float64,
	volatility float64,
) float64 {
	// Reverse formula: size = (target_slippage × depth) / volatility
	if volatility <= 0 {
		return 0
	}
	return (targetSlippage * availableDepth) / volatility
}

// ==================== INTERPRETATION ====================

// InterpretDepthRatio provides interpretation of depth ratio
func (dm *DepthModel) InterpretDepthRatio(depthRatio float64) string {
	switch {
	case depthRatio < 0.1:
		return "Very small order relative to depth - minimal slippage"
	case depthRatio < 0.25:
		return "Small order relative to depth - low slippage"
	case depthRatio < 0.5:
		return "Moderate order size - moderate slippage"
	case depthRatio < 1.0:
		return "Large order - significant slippage"
	case depthRatio < 2.0:
		return "Very large order - substantial slippage"
	default:
		return "Extremely large order - severe slippage, consider splitting"
	}
}

// InterpretSlippage provides interpretation of slippage amount
func (dm *DepthModel) InterpretSlippage(slippage float64) string {
	switch {
	case slippage < 0.001:
		return "Negligible slippage"
	case slippage < 0.005:
		return "Very low slippage"
	case slippage < 0.01:
		return "Low slippage"
	case slippage < 0.05:
		return "Moderate slippage"
	case slippage < 0.1:
		return "Significant slippage"
	case slippage < 0.5:
		return "High slippage"
	default:
		return "Very high slippage - execution risk"
	}
}
