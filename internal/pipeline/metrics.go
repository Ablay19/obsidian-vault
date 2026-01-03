package pipeline

import (
	"sync/atomic"
	"time"
)

type Metrics struct {
	SubmittedCount uint64
	ProcessedCount uint64
	DroppedCount   uint64
	TotalDuration  uint64 // Nanoseconds
}

func NewMetrics() *Metrics {
	return &Metrics{}
}

func (m *Metrics) JobSubmitted() {
	atomic.AddUint64(&m.SubmittedCount, 1)
}

func (m *Metrics) JobProcessed(duration time.Duration) {
	atomic.AddUint64(&m.ProcessedCount, 1)
	atomic.AddUint64(&m.TotalDuration, uint64(duration.Nanoseconds()))
}

func (m *Metrics) JobDropped() {
	atomic.AddUint64(&m.DroppedCount, 1)
}

func (m *Metrics) GetStats() map[string]interface{} {
	processed := atomic.LoadUint64(&m.ProcessedCount)
	totalDur := atomic.LoadUint64(&m.TotalDuration)
	avg := float64(0)
	if processed > 0 {
		avg = float64(totalDur) / float64(processed) / 1e6 // ms
	}

	return map[string]interface{}{
		"submitted": atomic.LoadUint64(&m.SubmittedCount),
		"processed": processed,
		"dropped":   atomic.LoadUint64(&m.DroppedCount),
		"avg_latency_ms": avg,
	}
}
