package metrics

import (
    "sync/atomic"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // Prometheus metrics
    processedTicksMetric = promauto.NewCounter(prometheus.CounterOpts{
        Name: "processed_ticks_total",
        Help: "The total number of processed market ticks",
    })
    
    errorCountMetric = promauto.NewCounter(prometheus.CounterOpts{
        Name: "error_count_total",
        Help: "Total number of errors encountered",
    })
    
    processingDuration = promauto.NewHistogram(prometheus.HistogramOpts{
        Name:    "tick_processing_seconds",
        Help:    "Time spent processing each tick",
        Buckets: prometheus.LinearBuckets(0.001, 0.001, 10),
    })

    // Internal counters
    processedTicks uint64
    errorCount     uint64
    lastProcessed  time.Time
    startTime      = time.Now()
)

func IncrementProcessed() {
    atomic.AddUint64(&processedTicks, 1)
    processedTicksMetric.Inc()
    lastProcessed = time.Now()
}

func IncrementErrors() {
    atomic.AddUint64(&errorCount, 1)
    errorCountMetric.Inc()
}

func GetStats() (uint64, uint64, time.Time, time.Duration) {
    return atomic.LoadUint64(&processedTicks),
           atomic.LoadUint64(&errorCount),
           lastProcessed,
           time.Since(startTime)
}

func RecordProcessingDuration(duration time.Duration) {
    processingDuration.Observe(duration.Seconds())
}
