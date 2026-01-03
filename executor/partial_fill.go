package executor

import (
	"fmt"
	"math"
)

// ==================== PARTIAL FILL CALCULATOR ====================

// PartialFillCalculator calculates partial fill sizes
type PartialFillCalculator struct{}

// NewPartialFillCalculator creates a new partial fill calculator
func NewPartialFillCalculator() PartialFillCalculator {
	return PartialFillCalculator{}
}

// CalculateFilledSize calculates the size that will actually be filled
// Returns: amount of order that gets filled (may be less than requested)
func (pfc PartialFillCalculator) CalculateFilledSize(
	requestedSize float64,
	availableDepth int64,
	volume int64,
) float64 {

	// If depth is sufficient, fill entire order
	if float64(availableDepth) >= requestedSize {
		return requestedSize
	}

	// Partial fill: limited by depth
	filledSize := math.Min(requestedSize, float64(availableDepth))

	// Further adjust by volume momentum
	if volume > 0 {
		volumeMultiplier := pfc.getVolumeMultiplier(volume)
		filledSize = filledSize * volumeMultiplier
	}

	return filledSize
}

// getVolumeMultiplier adjusts fill size based on volume level
// Higher volume = higher fill rate
func (pfc PartialFillCalculator) getVolumeMultiplier(volume int64) float64 {
	// Simple multiplier based on volume level
	// Could be enhanced with typical volume data
	switch {
	case volume > 2000000:
		return 1.0 // 100% fill
	case volume > 1000000:
		return 0.9 // 90% fill
	case volume > 500000:
		return 0.8 // 80% fill
	case volume > 250000:
		return 0.7 // 70% fill
	default:
		return 0.5 // 50% fill
	}
}

// ==================== DEPTH-BASED FILLS ====================

// CalculateDepthBasedFill calculates fill size limited by available depth
// Formula: min(requested_size, available_depth)
func (pfc PartialFillCalculator) CalculateDepthBasedFill(
	requestedSize float64,
	availableDepth int64,
) float64 {

	return math.Min(requestedSize, float64(availableDepth))
}

// ==================== MOMENTUM-BASED FILLS ====================

// CalculateMomentumBasedFill calculates fill size adjusted by market momentum
// Strong momentum = larger fills
// Weak momentum = smaller fills
func (pfc PartialFillCalculator) CalculateMomentumBasedFill(
	requestedSize float64,
	availableDepth int64,
	momentum int,
) float64 {

	// Start with depth-based fill
	depthFill := pfc.CalculateDepthBasedFill(requestedSize, availableDepth)

	// Adjust by momentum
	multiplier := pfc.getMomentumMultiplier(momentum)
	return depthFill * multiplier
}

// getMomentumMultiplier returns fill multiplier for momentum level
func (pfc PartialFillCalculator) getMomentumMultiplier(momentum int) float64 {
	switch momentum {
	case 0: // WEAK - harder to fill
		return 0.5
	case 1: // NORMAL - normal fills
		return 1.0
	case 2: // STRONG - easier to fill
		return 1.5
	default:
		return 1.0
	}
}

// ==================== VOLUME-BASED FILLS ====================

// CalculateVolumeLimitedFill calculates fill limited by available volume
func (pfc PartialFillCalculator) CalculateVolumeLimitedFill(
	requestedSize float64,
	availableVolume int64,
	typicalVolume int64,
) float64 {

	if typicalVolume == 0 {
		return requestedSize
	}

	volumeRatio := float64(availableVolume) / float64(typicalVolume)

	// Larger orders get smaller fills in low-volume situations
	maxFillPercent := math.Min(volumeRatio*100, 100) // Cap at 100%

	return requestedSize * (maxFillPercent / 100)
}

// ==================== ICEBERG-STYLE FILLS ====================

// IcebergFillCalculator handles iceberg order fills
type IcebergFillCalculator struct {
	totalSize      float64
	visibleSize    float64
	filledSoFar    float64
	remainingOrder float64
}

// NewIcebergFillCalculator creates a calculator for iceberg orders
func NewIcebergFillCalculator(totalSize float64, visibleSize float64) *IcebergFillCalculator {
	return &IcebergFillCalculator{
		totalSize:      totalSize,
		visibleSize:    visibleSize,
		filledSoFar:    0,
		remainingOrder: totalSize,
	}
}

// GetNextTranche returns the next visible tranche of an iceberg order
func (ifc *IcebergFillCalculator) GetNextTranche() float64 {
	if ifc.remainingOrder <= 0 {
		return 0
	}
	return math.Min(ifc.visibleSize, ifc.remainingOrder)
}

// RecordFill records that a tranche was filled
func (ifc *IcebergFillCalculator) RecordFill(filledSize float64) {
	ifc.filledSoFar += filledSize
	ifc.remainingOrder = ifc.totalSize - ifc.filledSoFar
}

// GetFillProgress returns the percentage of order filled (0-100)
func (ifc *IcebergFillCalculator) GetFillProgress() float64 {
	if ifc.totalSize == 0 {
		return 0
	}
	return (ifc.filledSoFar / ifc.totalSize) * 100
}

// IsComplete checks if entire order is filled
func (ifc *IcebergFillCalculator) IsComplete() bool {
	return ifc.remainingOrder <= 0
}

// ==================== ANALYSIS ====================

// FillAnalysis provides analysis of a fill scenario
type FillAnalysis struct {
	RequestedSize  float64
	AvailableDepth float64
	FilledSize     float64
	UnfilledSize   float64
	FillPercentage float64
	Reason         string
}

// AnalyzeFill analyzes a fill scenario
func (pfc PartialFillCalculator) AnalyzeFill(
	requestedSize float64,
	availableDepth int64,
	volume int64,
) *FillAnalysis {

	filledSize := pfc.CalculateFilledSize(requestedSize, availableDepth, volume)
	unfilled := requestedSize - filledSize
	fillPercent := (filledSize / requestedSize) * 100

	reason := "FULL_FILL"
	if filledSize < requestedSize {
		if float64(availableDepth) < requestedSize {
			reason = "DEPTH_LIMITED"
		} else {
			reason = "VOLUME_LIMITED"
		}
	}

	return &FillAnalysis{
		RequestedSize:  requestedSize,
		AvailableDepth: float64(availableDepth),
		FilledSize:     filledSize,
		UnfilledSize:   unfilled,
		FillPercentage: fillPercent,
		Reason:         reason,
	}
}

// ==================== DEBUG ====================

// String returns a string representation
func (fa *FillAnalysis) String() string {
	return fmt.Sprintf(
		"Fill: %.2f of %.2f (%.1f%%) - %s",
		fa.FilledSize,
		fa.RequestedSize,
		fa.FillPercentage,
		fa.Reason,
	)
}

// DebugString returns detailed fill analysis
func (fa *FillAnalysis) DebugString() string {
	return fmt.Sprintf(
		"Fill Analysis:\n"+
			"  Requested Size:    %.6f\n"+
			"  Available Depth:   %.6f\n"+
			"  Filled Size:       %.6f\n"+
			"  Unfilled Size:     %.6f\n"+
			"  Fill Percentage:   %.2f%%\n"+
			"  Reason:            %s",
		fa.RequestedSize,
		fa.AvailableDepth,
		fa.FilledSize,
		fa.UnfilledSize,
		fa.FillPercentage,
		fa.Reason,
	)
}

// ==================== FILL REJECTION RULES ====================

// ShouldRejectFill checks if a fill should be rejected
func ShouldRejectFill(
	filledSize float64,
	requestedSize float64,
	minFillPercentage float64,
) bool {

	if requestedSize == 0 {
		return false
	}

	fillPercent := (filledSize / requestedSize) * 100
	return fillPercent < minFillPercentage
}

// ==================== BATCH OPERATIONS ====================

// CalculateBatchFills calculates fills for multiple orders
func (pfc PartialFillCalculator) CalculateBatchFills(
	orders []OrderForFill,
) []float64 {

	fills := make([]float64, len(orders))

	for i, order := range orders {
		fills[i] = pfc.CalculateFilledSize(
			order.Size,
			order.AvailableDepth,
			order.Volume,
		)
	}

	return fills
}

// OrderForFill holds data for fill calculation
type OrderForFill struct {
	Size           float64
	AvailableDepth int64
	Volume         int64
}
