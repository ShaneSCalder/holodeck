package slippage

import (
	"fmt"
	"math"

	"holodeck/types"
)

// ==================== MOMENTUM MODEL ====================

// MomentumModel adjusts slippage based on price momentum
// Higher momentum (fast moving market) increases slippage
// Lower momentum (stable market) decreases slippage
type MomentumModel struct {
	// Constants
	BaseMultiplier float64 // Default 1.0
	MaxMultiplier  float64 // Maximum adjustment (default 2.0)

	// Statistics
	totalAdjustment float64
	adjustmentCount int64
	avgMomentum     float64
	maxMomentum     float64
	minMomentum     float64
}

// ==================== MODEL CREATION ====================

// NewMomentumModel creates a new momentum model
func NewMomentumModel() *MomentumModel {
	return &MomentumModel{
		BaseMultiplier: 1.0,
		MaxMultiplier:  2.0,
		minMomentum:    1e9,
	}
}

// ==================== CORE CALCULATION ====================

// AdjustSlippage adjusts base slippage using momentum multiplier
// Parameters:
//   - baseSlippage: Base slippage from depth model
//   - momentum: Momentum multiplier (1.0 = neutral, >1.0 = uptrend, <1.0 = downtrend)
//   - tick: Market tick for volatility analysis
//
// Returns: Adjusted slippage
func (mm *MomentumModel) AdjustSlippage(
	baseSlippage float64,
	momentum float64,
	tick *types.Tick,
) (float64, error) {

	if tick == nil {
		// Use base slippage with momentum adjustment if no tick
		adjustedSlippage := baseSlippage * momentum
		mm.recordAdjustment(adjustedSlippage, momentum)
		return adjustedSlippage, nil
	}

	// Calculate volatility from tick
	spread := tick.GetSpread()
	midPrice := tick.GetMidPrice()
	volatilityPercent := 0.0
	if midPrice > 0 {
		volatilityPercent = (spread / midPrice) * 100
	}

	// Calculate momentum adjustment factor
	adjustmentFactor := mm.calculateAdjustmentFactor(momentum, volatilityPercent)

	// Apply adjustment to base slippage
	adjustedSlippage := baseSlippage * adjustmentFactor

	// Track adjustment
	mm.recordAdjustment(adjustedSlippage, momentum)

	return adjustedSlippage, nil
}

// ==================== ADJUSTMENT FACTOR CALCULATION ====================

// calculateAdjustmentFactor calculates the momentum adjustment factor
func (mm *MomentumModel) calculateAdjustmentFactor(
	momentum float64,
	volatilityPercent float64,
) float64 {

	// Start with base multiplier
	factor := mm.BaseMultiplier

	// Apply momentum adjustment
	// momentum > 1.0 = uptrend/strong price movement = increase slippage
	// momentum < 1.0 = downtrend/weak price movement = decrease slippage
	// momentum = 1.0 = neutral = no adjustment
	if momentum > 1.0 {
		// Increase slippage in strong momentum
		momentumAdjustment := (momentum - 1.0) * 0.5 // Scale the adjustment
		factor += momentumAdjustment
	} else if momentum < 1.0 && momentum > 0 {
		// Decrease slippage in weak momentum
		momentumAdjustment := (1.0 - momentum) * 0.3 // Scale the reduction
		factor -= momentumAdjustment
	}

	// Apply volatility adjustment (0-10% spread = 0-1x multiplier)
	volAdjustment := volatilityPercent / 10.0
	if volAdjustment > 1.0 {
		volAdjustment = 1.0 // Cap at 1.0
	}
	factor += volAdjustment * 0.3 // Apply 30% of volatility adjustment

	// Ensure factor doesn't exceed maximum multiplier
	if factor > mm.MaxMultiplier {
		factor = mm.MaxMultiplier
	}

	// Ensure factor doesn't go below minimum (0.2x)
	if factor < 0.2 {
		factor = 0.2
	}

	return factor
}

// ==================== RECORDING ====================

// recordAdjustment records adjustment statistics
func (mm *MomentumModel) recordAdjustment(adjustment float64, momentum float64) {
	mm.totalAdjustment += adjustment
	mm.adjustmentCount++
	mm.avgMomentum = (mm.avgMomentum*(float64(mm.adjustmentCount-1)) + momentum) / float64(mm.adjustmentCount)
	if momentum > mm.maxMomentum {
		mm.maxMomentum = momentum
	}
	if momentum < mm.minMomentum {
		mm.minMomentum = momentum
	}
}

// ==================== STATISTICS ====================

// GetTotalAdjustment returns total adjustment applied
func (mm *MomentumModel) GetTotalAdjustment() float64 {
	return mm.totalAdjustment
}

// GetAdjustmentCount returns number of adjustments
func (mm *MomentumModel) GetAdjustmentCount() int64 {
	return mm.adjustmentCount
}

// GetAverageAdjustment returns average adjustment per trade
func (mm *MomentumModel) GetAverageAdjustment() float64 {
	if mm.adjustmentCount == 0 {
		return 0
	}
	return mm.totalAdjustment / float64(mm.adjustmentCount)
}

// GetAverageMomentum returns average momentum observed
func (mm *MomentumModel) GetAverageMomentum() float64 {
	return mm.avgMomentum
}

// GetMaxMomentum returns maximum momentum observed
func (mm *MomentumModel) GetMaxMomentum() float64 {
	if mm.adjustmentCount == 0 {
		return 0
	}
	return mm.maxMomentum
}

// GetMinMomentum returns minimum momentum observed
func (mm *MomentumModel) GetMinMomentum() float64 {
	if mm.adjustmentCount == 0 {
		return 0
	}
	return mm.minMomentum
}

// GetStatistics returns comprehensive momentum model statistics
func (mm *MomentumModel) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"total_adjustment":   mm.totalAdjustment,
		"adjustment_count":   mm.adjustmentCount,
		"average_adjustment": mm.GetAverageAdjustment(),
		"avg_momentum":       mm.avgMomentum,
		"max_momentum":       mm.GetMaxMomentum(),
		"min_momentum":       mm.GetMinMomentum(),
		"base_multiplier":    mm.BaseMultiplier,
		"max_multiplier":     mm.MaxMultiplier,
	}
}

// ==================== DEBUG ====================

// String returns a human-readable representation
func (mm *MomentumModel) String() string {
	return fmt.Sprintf(
		"MomentumModel[Total:%.4f, Avg:%.4f, Momentum:%.4f]",
		mm.totalAdjustment,
		mm.GetAverageAdjustment(),
		mm.avgMomentum,
	)
}

// DebugString returns detailed debug information
func (mm *MomentumModel) DebugString() string {
	return fmt.Sprintf(
		"Momentum Model:\n"+
			"  Total Adjustment:      %.4f pips\n"+
			"  Adjustment Count:      %d\n"+
			"  Average Adjustment:    %.4f pips\n"+
			"  Avg Momentum:          %.4f\n"+
			"  Max Momentum:          %.4f\n"+
			"  Min Momentum:          %.4f\n"+
			"  Base Multiplier:       %.2f\n"+
			"  Max Multiplier:        %.2f",
		mm.totalAdjustment,
		mm.adjustmentCount,
		mm.GetAverageAdjustment(),
		mm.avgMomentum,
		mm.GetMaxMomentum(),
		mm.GetMinMomentum(),
		mm.BaseMultiplier,
		mm.MaxMultiplier,
	)
}

// Reset resets model statistics
func (mm *MomentumModel) Reset() {
	mm.totalAdjustment = 0
	mm.adjustmentCount = 0
	mm.avgMomentum = 0
	mm.maxMomentum = 0
	mm.minMomentum = 1e9
}

// ==================== ANALYSIS ====================

// AnalyzeMomentumAdjustment provides detailed analysis of momentum adjustment
func (mm *MomentumModel) AnalyzeMomentumAdjustment(
	baseSlippage float64,
	momentum float64,
	volatilityPercent float64,
) *MomentumAdjustmentAnalysis {

	adjustmentFactor := mm.calculateAdjustmentFactor(momentum, volatilityPercent)
	adjustedSlippage := baseSlippage * adjustmentFactor

	return &MomentumAdjustmentAnalysis{
		BaseSlippage:      baseSlippage,
		Momentum:          momentum,
		VolatilityPercent: volatilityPercent,
		AdjustmentFactor:  adjustmentFactor,
		AdjustedSlippage:  adjustedSlippage,
		AdjustmentPercent: ((adjustmentFactor - 1.0) / 1.0) * 100,
	}
}

// ==================== ANALYSIS TYPES ====================

// MomentumAdjustmentAnalysis provides detailed breakdown of momentum adjustment
type MomentumAdjustmentAnalysis struct {
	BaseSlippage      float64
	Momentum          float64
	VolatilityPercent float64
	AdjustmentFactor  float64
	AdjustedSlippage  float64
	AdjustmentPercent float64
}

// String returns string representation
func (maa *MomentumAdjustmentAnalysis) String() string {
	return fmt.Sprintf(
		"Base:%.4f Mom:%.2f Vol:%.2f%% => Factor:%.2f => Adjusted:%.4f",
		maa.BaseSlippage,
		maa.Momentum,
		maa.VolatilityPercent,
		maa.AdjustmentFactor,
		maa.AdjustedSlippage,
	)
}

// DebugString returns detailed debug information
func (maa *MomentumAdjustmentAnalysis) DebugString() string {
	adjustmentDir := "increase"
	if maa.AdjustmentPercent < 0 {
		adjustmentDir = "decrease"
	}
	return fmt.Sprintf(
		"Momentum Adjustment Analysis:\n"+
			"  Base Slippage:         %.4f pips\n"+
			"  Momentum:              %.4f\n"+
			"  Volatility:            %.2f%%\n"+
			"  Adjustment Factor:     %.2f x\n"+
			"  Adjusted Slippage:     %.4f pips\n"+
			"  Adjustment:            %s by %.2f%%",
		maa.BaseSlippage,
		maa.Momentum,
		maa.VolatilityPercent,
		maa.AdjustmentFactor,
		maa.AdjustedSlippage,
		adjustmentDir,
		math.Abs(maa.AdjustmentPercent),
	)
}

// ==================== MOMENTUM INTERPRETATION ====================

// InterpretMomentum provides interpretation of momentum value
func (mm *MomentumModel) InterpretMomentum(momentum float64) string {
	switch {
	case momentum > 1.5:
		return "Very strong momentum - expect maximum slippage"
	case momentum > 1.2:
		return "Strong momentum - significant slippage increase"
	case momentum > 1.0:
		return "Positive momentum - slippage increase"
	case momentum == 1.0:
		return "Neutral momentum - no adjustment"
	case momentum > 0.8:
		return "Weak momentum - slippage decrease"
	case momentum > 0.5:
		return "Very weak momentum - significant slippage reduction"
	default:
		return "Negligible momentum - minimal slippage"
	}
}

// ==================== UTILITY FUNCTIONS ====================

// SetMaxMultiplier sets the maximum multiplier
func (mm *MomentumModel) SetMaxMultiplier(max float64) {
	if max > 0 {
		mm.MaxMultiplier = max
	}
}

// SetBaseMultiplier sets the base multiplier
func (mm *MomentumModel) SetBaseMultiplier(base float64) {
	if base > 0 {
		mm.BaseMultiplier = base
	}
}

// CalculateMomentumMultiplier calculates momentum multiplier from price data
func (mm *MomentumModel) CalculateMomentumMultiplier(
	currentPrice float64,
	previousPrice float64,
	atr float64, // Average True Range for volatility
) float64 {

	if previousPrice <= 0 || atr <= 0 {
		return 1.0 // Neutral
	}

	// Calculate price change as percentage of ATR
	priceChange := currentPrice - previousPrice
	momentumFactor := 1.0 + (priceChange/atr)*0.1 // Scale to reasonable range

	// Clamp between 0.5 and 2.0
	if momentumFactor < 0.5 {
		momentumFactor = 0.5
	}
	if momentumFactor > 2.0 {
		momentumFactor = 2.0
	}

	return momentumFactor
}
