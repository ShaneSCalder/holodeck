package speed

import (
	"fmt"
	"sync"
	"time"
)

// ==================== SPEED CONTROLLER ====================

// SpeedController manages simulation speed and timing
type SpeedController struct {
	// Configuration
	multiplier    float64
	minMultiplier float64
	maxMultiplier float64

	// Timing
	startTime         time.Time
	baseTickDuration  time.Duration
	targetTimePerTick time.Duration
	lastTickTime      time.Time

	// Statistics
	ticksProcessed   int64
	totalProcessTime time.Duration
	totalWaitTime    time.Duration
	skippedSleeps    int64
	actualMultiplier float64

	// State
	mu         sync.RWMutex
	paused     bool
	pausedTime time.Time
}

// ==================== CREATION ====================

// NewSpeedController creates a new speed controller
func NewSpeedController() *SpeedController {
	return &SpeedController{
		multiplier:       1.0,
		minMultiplier:    0.1,
		maxMultiplier:    10000.0,
		baseTickDuration: time.Second,
		startTime:        time.Now(),
		lastTickTime:     time.Now(),
	}
}

// ==================== SPEED CONTROL ====================

// SetSpeed sets the simulation speed multiplier
// speed 1.0 = real-time (1 second per tick)
// speed 100 = 100x faster (10ms per tick)
// speed 0.1 = 10x slower (10 seconds per tick)
func (sc *SpeedController) SetSpeed(multiplier float64) error {
	if multiplier < sc.minMultiplier {
		return fmt.Errorf("speed multiplier %.1f is below minimum %.1f", multiplier, sc.minMultiplier)
	}
	if multiplier > sc.maxMultiplier {
		return fmt.Errorf("speed multiplier %.1f exceeds maximum %.1f", multiplier, sc.maxMultiplier)
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.multiplier = multiplier
	sc.calculateTargetTime()

	return nil
}

// GetSpeed returns the current speed multiplier
func (sc *SpeedController) GetSpeed() float64 {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.multiplier
}

// calculateTargetTime calculates target time per tick based on multiplier
func (sc *SpeedController) calculateTargetTime() {
	// targetTime = baseTime / multiplier
	// If multiplier = 100, target = 1s / 100 = 10ms
	sc.targetTimePerTick = time.Duration(float64(sc.baseTickDuration) / sc.multiplier)
}

// ==================== TICK TIMING ====================

// WaitTick waits the appropriate amount of time before the next tick
// Pass the actual processing time for this tick for accurate timing
func (sc *SpeedController) WaitTick(processingTime time.Duration) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	// Check if paused
	if sc.paused {
		return nil
	}

	// Calculate required sleep duration
	requiredSleep := sc.targetTimePerTick - processingTime

	// Track statistics
	sc.ticksProcessed++
	sc.totalProcessTime += processingTime

	// If processing took longer than target, no sleep needed
	if requiredSleep <= 0 {
		sc.skippedSleeps++
		return nil
	}

	// Sleep for the required time
	time.Sleep(requiredSleep)
	sc.totalWaitTime += requiredSleep

	return nil
}

// ==================== PAUSE/RESUME ====================

// Pause pauses the simulation
func (sc *SpeedController) Pause() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.paused {
		return fmt.Errorf("simulation already paused")
	}

	sc.paused = true
	sc.pausedTime = time.Now()
	return nil
}

// Resume resumes the simulation
func (sc *SpeedController) Resume() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if !sc.paused {
		return fmt.Errorf("simulation not paused")
	}

	// Adjust start time for pause duration
	pauseDuration := time.Since(sc.pausedTime)
	sc.startTime = sc.startTime.Add(pauseDuration)

	sc.paused = false
	return nil
}

// IsPaused returns whether simulation is paused
func (sc *SpeedController) IsPaused() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.paused
}

// ==================== STATISTICS ====================

// GetStatistics returns speed controller statistics
func (sc *SpeedController) GetStatistics() map[string]interface{} {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	elapsed := time.Since(sc.startTime)

	// Calculate actual multiplier (simulated time / real time)
	var actualMultiplier float64
	if elapsed > 0 {
		simulatedTime := time.Duration(float64(sc.ticksProcessed) * float64(sc.baseTickDuration))
		actualMultiplier = float64(simulatedTime) / float64(elapsed)
	}

	return map[string]interface{}{
		"configured_speed":     sc.multiplier,
		"actual_speed":         actualMultiplier,
		"target_time_per_tick": sc.targetTimePerTick.String(),
		"ticks_processed":      sc.ticksProcessed,
		"total_process_time":   sc.totalProcessTime.String(),
		"total_wait_time":      sc.totalWaitTime.String(),
		"skipped_sleeps":       sc.skippedSleeps,
		"elapsed_time":         elapsed.String(),
		"is_paused":            sc.paused,
	}
}

// PrintStatistics returns formatted statistics string
func (sc *SpeedController) PrintStatistics() string {
	stats := sc.GetStatistics()

	return fmt.Sprintf(
		"=== SPEED CONTROLLER STATISTICS ===\n"+
			"Configured Speed:      %.1fx\n"+
			"Actual Speed:          %.1fx\n"+
			"Target Time Per Tick:  %s\n"+
			"Ticks Processed:       %d\n"+
			"Total Process Time:    %s\n"+
			"Total Wait Time:       %s\n"+
			"Skipped Sleeps:        %d\n"+
			"Elapsed Time:          %s\n"+
			"Paused:                %v\n",
		stats["configured_speed"],
		stats["actual_speed"],
		stats["target_time_per_tick"],
		stats["ticks_processed"],
		stats["total_process_time"],
		stats["total_wait_time"],
		stats["skipped_sleeps"],
		stats["elapsed_time"],
		stats["is_paused"],
	)
}

// GetTicksProcessed returns number of ticks processed
func (sc *SpeedController) GetTicksProcessed() int64 {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.ticksProcessed
}

// GetElapsedTime returns elapsed wall-clock time
func (sc *SpeedController) GetElapsedTime() time.Duration {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return time.Since(sc.startTime)
}

// GetActualMultiplier returns the actual achieved multiplier
func (sc *SpeedController) GetActualMultiplier() float64 {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	elapsed := time.Since(sc.startTime)
	if elapsed <= 0 {
		return 0
	}

	simulatedTime := time.Duration(float64(sc.ticksProcessed) * float64(sc.baseTickDuration))
	return float64(simulatedTime) / float64(elapsed)
}

// ==================== RESET ====================

// Reset resets the speed controller
func (sc *SpeedController) Reset() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.multiplier = 1.0
	sc.startTime = time.Now()
	sc.lastTickTime = time.Now()
	sc.ticksProcessed = 0
	sc.totalProcessTime = 0
	sc.totalWaitTime = 0
	sc.skippedSleeps = 0
	sc.paused = false

	sc.calculateTargetTime()
	return nil
}

// ==================== CONFIGURATION ====================

// SetMinMultiplier sets the minimum allowed multiplier
func (sc *SpeedController) SetMinMultiplier(min float64) error {
	if min <= 0 {
		return fmt.Errorf("minimum multiplier must be positive")
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.minMultiplier = min
	return nil
}

// SetMaxMultiplier sets the maximum allowed multiplier
func (sc *SpeedController) SetMaxMultiplier(max float64) error {
	if max <= 0 {
		return fmt.Errorf("maximum multiplier must be positive")
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.maxMultiplier = max
	return nil
}

// SetBaseTickDuration sets the base duration per tick
func (sc *SpeedController) SetBaseTickDuration(duration time.Duration) error {
	if duration <= 0 {
		return fmt.Errorf("base tick duration must be positive")
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.baseTickDuration = duration
	sc.calculateTargetTime()
	return nil
}

// ==================== DESCRIPTIVE STRINGS ====================

// DescribeSpeed returns a human-readable description of the speed
func DescribeSpeed(multiplier float64) string {
	switch {
	case multiplier < 0.1:
		return "very slow (slower than real-time)"
	case multiplier < 1.0:
		return "slow (slower than real-time)"
	case multiplier == 1.0:
		return "real-time"
	case multiplier <= 10:
		return "faster than real-time"
	case multiplier <= 100:
		return "much faster than real-time"
	case multiplier <= 1000:
		return "extremely fast"
	default:
		return "ultra-fast"
	}
}

// DescribeSimulationTime returns estimated simulation time for N ticks
func DescribeSimulationTime(ticks int64, speed float64) (wallClock time.Duration, simulated time.Duration) {
	baseTick := time.Second
	simulatedTotal := time.Duration(ticks) * baseTick
	wallClockTotal := time.Duration(float64(simulatedTotal) / speed)

	return wallClockTotal, simulatedTotal
}

// Example: 50,000 ticks at 100x speed
// wallClock: ~500ms (50,000 * 1s / 100)
// simulated: ~50,000 seconds (13.9 hours)
