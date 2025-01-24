package config

type Config struct {
    App struct {
        Environment string
        LogLevel    string
        Port        int
        Workers     int
        BatchSize   int
    }
    
    AngelOne struct {
        // ... existing angel one config ...
    }
    
    ClickHouse struct {
        // ... existing clickhouse config ...
    }
    
    Monitoring struct {
        PrometheusPort int
        AlertThresholds map[string]float64
    }
}
