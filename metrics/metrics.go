package metrics

import (
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "broker_clickhouse/config"
    dto "github.com/prometheus/client_model/go"
    "fmt"
)

const (
    metricsNamespace = "angelone"
    metricsSubsystem = "market_data"
)

type Metrics struct {
    config          *config.Config
    processedTicks  prometheus.Counter
    errorCount      prometheus.Counter
    processingTime  prometheus.Histogram
    batchSize       prometheus.Gauge
    lastProcessed   time.Time
    startTime       time.Time
}

func NewMetrics(cfg *config.Config) *Metrics {
    m := &Metrics{
        config:    cfg,
        startTime: time.Now(),
    }
    
    m.processedTicks = promauto.NewCounter(prometheus.CounterOpts{
        Namespace: metricsNamespace,
        Subsystem: metricsSubsystem,
        Name:      "processed_ticks_total",
        Help:      "The total number of processed market ticks",
    })

    m.errorCount = promauto.NewCounter(prometheus.CounterOpts{
        Namespace: metricsNamespace,
        Subsystem: metricsSubsystem,
        Name:      "error_count_total",
        Help:      "Total number of errors encountered",
    })
    
    m.processingTime = promauto.NewHistogram(prometheus.HistogramOpts{
        Namespace: metricsNamespace,
        Subsystem: metricsSubsystem,
        Name:      "tick_processing_seconds",
        Help:      "Time spent processing each tick",
        Buckets:   prometheus.LinearBuckets(0.001, 0.001, 10),
    })

    return m
}

func (m *Metrics) IncrementProcessed() {
    m.processedTicks.Inc()
    m.lastProcessed = time.Now()
}

func (m *Metrics) IncrementErrors() {
    m.errorCount.Inc()
}

func (m *Metrics) RecordProcessingDuration(duration time.Duration) {
    m.processingTime.Observe(duration.Seconds())
}

func (m *Metrics) GetStats() (uint64, uint64, time.Time, time.Duration) {
    processed := float64(0)
    errors := float64(0)
    
    // Get metric values using dto.Metric
    if metric, err := getMetricValue(m.processedTicks); err == nil {
        processed = metric.GetCounter().GetValue()
    }
    
    if metric, err := getMetricValue(m.errorCount); err == nil {
        errors = metric.GetCounter().GetValue()
    }
    
    return uint64(processed),
           uint64(errors),
           m.lastProcessed,
           time.Since(m.startTime)
}

// Helper function to get metric value
func getMetricValue(metric prometheus.Collector) (*dto.Metric, error) {
    ch := make(chan prometheus.Metric, 1)
    metric.Collect(ch)
    if m := <-ch; m != nil {
        var metric dto.Metric
        if err := m.Write(&metric); err != nil {
            return nil, err
        }
        return &metric, nil
    }
    return nil, fmt.Errorf("no metric value available")
}
