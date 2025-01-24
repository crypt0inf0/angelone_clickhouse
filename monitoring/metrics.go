package monitoring

import (
    "runtime"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // Request latency
    RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "market_data_request_duration_seconds",
        Help:    "Time taken to process market data requests",
        Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
    }, []string{"operation"})

    // Error rates
    ErrorCounter = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "market_data_errors_total",
        Help: "Total number of errors by type",
    }, []string{"type"})

    // System resources
    MemoryUsage = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "market_data_memory_bytes",
        Help: "Current memory usage in bytes",
    })

    CPUUsage = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "market_data_cpu_usage",
        Help: "Current CPU usage percentage",
    })

    GoroutineCount = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "market_data_goroutines",
        Help: "Current number of goroutines",
    })

    // ClickHouse metrics
    QueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "clickhouse_query_duration_seconds",
        Help:    "Time taken for ClickHouse queries",
        Buckets: prometheus.LinearBuckets(0.01, 0.05, 10),
    }, []string{"query_type"})

    BatchSize = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "market_data_batch_size",
        Help: "Current size of the batch buffer",
    })
)

// Start collecting system metrics
func StartMetricsCollection() {
    go func() {
        ticker := time.NewTicker(5 * time.Second)
        defer ticker.Stop()

        for range ticker.C {
            collectSystemMetrics()
        }
    }()
}

func collectSystemMetrics() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    MemoryUsage.Set(float64(m.Alloc))
    GoroutineCount.Set(float64(runtime.NumGoroutine()))
}
