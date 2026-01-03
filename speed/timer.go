package speed

import (
	"fmt"
	"time"
)

// ==================== TICK TIMER ====================

// TickTimer manages timing for individual ticks with adaptive sleep logic
type TickTimer struct {
	controller       *SpeedController
	tickStartTime    time.Time
	lastTickDuration time.Duration
}

// ==================== CREATION ====================

// NewTickTimer creates a new tick timer
func NewTickTimer(controller *SpeedController) *TickTimer {
	return &TickTimer{
		controller:    controller,
		tickStartTime: time.Now(),
	}
}

// ==================== TICK TIMING ====================

// StartTick marks the start of a tick
func (tt *TickTimer) StartTick() {
	tt.tickStartTime = time.Now()
}

// EndTick marks the end of a tick and performs appropriate timing adjustment
func (tt *TickTimer) EndTick() error {
	processingTime := time.Since(tt.tickStartTime)
	tt.lastTickDuration = processingTime

	return tt.controller.WaitTick(processingTime)
}

// GetLastTickDuration returns the duration of the last tick
func (tt *TickTimer) GetLastTickDuration() time.Duration {
	return tt.lastTickDuration
}

// ==================== BATCH TIMER ====================

// BatchTimer manages timing for a batch of ticks
type BatchTimer struct {
	controller     *SpeedController
	batchStartTime time.Time
	batchSize      int64
	ticksProcessed int64
	totalDuration  time.Duration
	minTickTime    time.Duration
	maxTickTime    time.Duration
	avgTickTime    time.Duration
}

// ==================== CREATION ====================

// NewBatchTimer creates a new batch timer
func NewBatchTimer(controller *SpeedController, batchSize int64) *BatchTimer {
	return &BatchTimer{
		controller:     controller,
		batchStartTime: time.Now(),
		batchSize:      batchSize,
		minTickTime:    time.Duration(1<<63 - 1), // Max int64
	}
}

// ==================== BATCH TIMING ====================

// StartBatch marks the start of a batch
func (bt *BatchTimer) StartBatch() {
	bt.batchStartTime = time.Now()
	bt.ticksProcessed = 0
	bt.totalDuration = 0
	bt.minTickTime = time.Duration(1<<63 - 1)
	bt.maxTickTime = 0
}

// RecordTick records a tick duration
func (bt *BatchTimer) RecordTick(tickDuration time.Duration) {
	bt.ticksProcessed++
	bt.totalDuration += tickDuration

	if tickDuration < bt.minTickTime {
		bt.minTickTime = tickDuration
	}
	if tickDuration > bt.maxTickTime {
		bt.maxTickTime = tickDuration
	}

	if bt.ticksProcessed > 0 {
		bt.avgTickTime = bt.totalDuration / time.Duration(bt.ticksProcessed)
	}
}

// EndBatch marks the end of a batch and returns statistics
func (bt *BatchTimer) EndBatch() BatchStatistics {
	wallClockTime := time.Since(bt.batchStartTime)

	var progress float64
	if bt.batchSize > 0 {
		progress = float64(bt.ticksProcessed) / float64(bt.batchSize) * 100
	}

	return BatchStatistics{
		WallClockTime:    wallClockTime,
		TicksProcessed:   bt.ticksProcessed,
		BatchSize:        bt.batchSize,
		AverageTickTime:  bt.avgTickTime,
		MinTickTime:      bt.minTickTime,
		MaxTickTime:      bt.maxTickTime,
		TotalProcessTime: bt.totalDuration,
		Progress:         progress,
		TicksPerSecond:   calculateTicksPerSecond(bt.ticksProcessed, wallClockTime),
	}
}

// GetProgress returns batch progress percentage
func (bt *BatchTimer) GetProgress() float64 {
	if bt.batchSize <= 0 {
		return 0
	}
	return float64(bt.ticksProcessed) / float64(bt.batchSize) * 100
}

// GetEstimatedTimeRemaining estimates time remaining for batch
func (bt *BatchTimer) GetEstimatedTimeRemaining() time.Duration {
	if bt.ticksProcessed <= 0 || bt.avgTickTime <= 0 {
		return 0
	}

	remaining := bt.batchSize - bt.ticksProcessed
	return time.Duration(remaining) * (bt.avgTickTime / time.Duration(bt.controller.GetActualMultiplier()))
}

// ==================== BATCH STATISTICS ====================

// BatchStatistics represents batch timing statistics
type BatchStatistics struct {
	WallClockTime    time.Duration
	TicksProcessed   int64
	BatchSize        int64
	AverageTickTime  time.Duration
	MinTickTime      time.Duration
	MaxTickTime      time.Duration
	TotalProcessTime time.Duration
	Progress         float64
	TicksPerSecond   float64
}

// String returns formatted statistics string
func (bs BatchStatistics) String() string {
	return fmt.Sprintf(
		"Batch Statistics:\n"+
			"  Wall Clock Time:    %s\n"+
			"  Ticks Processed:    %d / %d (%.1f%%)\n"+
			"  Average Tick Time:  %v\n"+
			"  Min Tick Time:      %v\n"+
			"  Max Tick Time:      %v\n"+
			"  Total Process Time: %s\n"+
			"  Ticks Per Second:   %.1f\n",
		bs.WallClockTime,
		bs.TicksProcessed,
		bs.BatchSize,
		bs.Progress,
		bs.AverageTickTime,
		bs.MinTickTime,
		bs.MaxTickTime,
		bs.TotalProcessTime,
		bs.TicksPerSecond,
	)
}

// ==================== SESSION TIMER ====================

// SessionTimer manages timing for an entire backtesting session
type SessionTimer struct {
	controller          *SpeedController
	sessionStartTime    time.Time
	sessionName         string
	batches             []*BatchTimer
	currentBatch        *BatchTimer
	totalTicksProcessed int64
}

// ==================== CREATION ====================

// NewSessionTimer creates a new session timer
func NewSessionTimer(controller *SpeedController, sessionName string) *SessionTimer {
	return &SessionTimer{
		controller:       controller,
		sessionStartTime: time.Now(),
		sessionName:      sessionName,
		batches:          make([]*BatchTimer, 0),
	}
}

// ==================== SESSION TIMING ====================

// StartBatch starts a new batch within the session
func (st *SessionTimer) StartBatch(batchSize int64) {
	st.currentBatch = NewBatchTimer(st.controller, batchSize)
	st.currentBatch.StartBatch()
}

// RecordTick records a tick in the current batch
func (st *SessionTimer) RecordTick(tickDuration time.Duration) error {
	if st.currentBatch == nil {
		return fmt.Errorf("no batch in progress")
	}

	st.currentBatch.RecordTick(tickDuration)
	st.totalTicksProcessed++
	return nil
}

// EndBatch ends the current batch
func (st *SessionTimer) EndBatch() BatchStatistics {
	if st.currentBatch == nil {
		return BatchStatistics{}
	}

	stats := st.currentBatch.EndBatch()
	st.batches = append(st.batches, st.currentBatch)
	st.currentBatch = nil

	return stats
}

// EndSession ends the session and returns summary statistics
func (st *SessionTimer) EndSession() SessionStatistics {
	wallClockTime := time.Since(st.sessionStartTime)

	var totalBatchTime time.Duration
	var avgBatchTime time.Duration
	var totalAvgTickTime time.Duration

	if len(st.batches) > 0 {
		for _, batch := range st.batches {
			totalBatchTime += batch.EndBatch().WallClockTime
			totalAvgTickTime += batch.avgTickTime
		}
		avgBatchTime = totalBatchTime / time.Duration(len(st.batches))
		totalAvgTickTime = totalAvgTickTime / time.Duration(len(st.batches))
	}

	return SessionStatistics{
		SessionName:         st.sessionName,
		WallClockTime:       wallClockTime,
		TotalTicksProcessed: st.totalTicksProcessed,
		BatchCount:          int64(len(st.batches)),
		AverageBatchTime:    avgBatchTime,
		AverageTickTime:     totalAvgTickTime,
		ActualSpeed:         st.controller.GetActualMultiplier(),
	}
}

// ==================== SESSION STATISTICS ====================

// SessionStatistics represents session timing statistics
type SessionStatistics struct {
	SessionName         string
	WallClockTime       time.Duration
	TotalTicksProcessed int64
	BatchCount          int64
	AverageBatchTime    time.Duration
	AverageTickTime     time.Duration
	ActualSpeed         float64
}

// String returns formatted statistics string
func (ss SessionStatistics) String() string {
	return fmt.Sprintf(
		"Session Statistics: %s\n"+
			"  Wall Clock Time:     %s\n"+
			"  Total Ticks:         %d\n"+
			"  Batch Count:         %d\n"+
			"  Average Batch Time:  %s\n"+
			"  Average Tick Time:   %s\n"+
			"  Actual Speed:        %.1fx\n",
		ss.SessionName,
		ss.WallClockTime,
		ss.TotalTicksProcessed,
		ss.BatchCount,
		ss.AverageBatchTime,
		ss.AverageTickTime,
		ss.ActualSpeed,
	)
}

// ==================== UTILITY FUNCTIONS ====================

// calculateTicksPerSecond calculates ticks processed per second
func calculateTicksPerSecond(ticks int64, duration time.Duration) float64 {
	if duration <= 0 {
		return 0
	}
	return float64(ticks) / duration.Seconds()
}

// CalculateSimulationTime calculates how long a simulation will take
// baseTick: base duration per tick (typically 1 second)
// tickCount: total ticks to process
// speed: simulation speed multiplier
// Returns: (wallClockTime, simulatedTime)
func CalculateSimulationTime(baseTick time.Duration, tickCount int64, speed float64) (time.Duration, time.Duration) {
	if speed <= 0 {
		speed = 1.0
	}

	simulatedTotal := time.Duration(tickCount) * baseTick
	wallClockTotal := time.Duration(float64(simulatedTotal) / speed)

	return wallClockTotal, simulatedTotal
}

// FormatDuration returns a human-readable duration string
func FormatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.0f µs", d.Seconds()*1e6)
	}
	if d < time.Second {
		return fmt.Sprintf("%.1f ms", d.Seconds()*1000)
	}
	if d < time.Minute {
		return fmt.Sprintf("%.2f s", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1f m", d.Minutes())
	}
	return fmt.Sprintf("%.1f h", d.Hours())
}

// ==================== SPEED PRESETS ====================

// SpeedPreset represents a pre-defined speed setting
type SpeedPreset struct {
	Name        string
	Multiplier  float64
	Description string
}

// GetSpeedPresets returns common speed presets
func GetSpeedPresets() []SpeedPreset {
	return []SpeedPreset{
		{
			Name:        "Slow",
			Multiplier:  0.5,
			Description: "Half speed (slower than real-time)",
		},
		{
			Name:        "RealTime",
			Multiplier:  1.0,
			Description: "Real-time simulation",
		},
		{
			Name:        "Fast",
			Multiplier:  10.0,
			Description: "10x faster than real-time",
		},
		{
			Name:        "VeryFast",
			Multiplier:  100.0,
			Description: "100x faster (1 year in ~2.5 min)",
		},
		{
			Name:        "SuperFast",
			Multiplier:  1000.0,
			Description: "1000x faster",
		},
		{
			Name:        "UltraFast",
			Multiplier:  10000.0,
			Description: "10000x faster",
		},
	}
}

// FindPreset finds a speed preset by name
func FindPreset(name string) *SpeedPreset {
	presets := GetSpeedPresets()
	for i := range presets {
		if presets[i].Name == name {
			return &presets[i]
		}
	}
	return nil
}
