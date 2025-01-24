package config

import (
    "os"
    "strconv"
    "time"
)

type Config struct {
    App struct {
        Environment  string
        LogLevel    string
        NumWorkers  int
        BufferSize  int
        BatchSize   int
        TimeoutSecs int
    }

    ClickHouse struct {
        Host            string
        Port            int
        User            string
        Password        string
        Database        string
        MaxOpenConns    int
        MaxIdleConns    int
        ConnMaxLifetime time.Duration
        QueryTimeout    time.Duration
        Debug          bool
    }

    Security struct {
        TLSEnabled     bool
        CertFile       string
        KeyFile        string
        RequestTimeout time.Duration
    }

    Metrics struct {
        Prefix        string
        EnableDebug   bool
        Labels        map[string]string
    }
}

func Load() (*Config, error) {
    cfg := &Config{}
    
    // App settings
    cfg.App.Environment = getEnvOrDefault("APP_ENV", "production")
    cfg.App.LogLevel = getEnvOrDefault("LOG_LEVEL", "info")
    cfg.App.NumWorkers = getEnvAsIntOrDefault("NUM_WORKERS", 5)
    cfg.App.BufferSize = getEnvAsIntOrDefault("BUFFER_SIZE", 1000)
    cfg.App.BatchSize = getEnvAsIntOrDefault("BATCH_SIZE", 1000)
    cfg.App.TimeoutSecs = getEnvAsIntOrDefault("TIMEOUT_SECS", 30)

    // ClickHouse settings
    cfg.ClickHouse.Host = getEnvOrDefault("CLICKHOUSE_HOST", "localhost")
    cfg.ClickHouse.Port = getEnvAsIntOrDefault("CLICKHOUSE_PORT", 9000)
    cfg.ClickHouse.User = getEnvOrDefault("CLICKHOUSE_USER", "default")
    cfg.ClickHouse.Password = os.Getenv("CLICKHOUSE_PASSWORD")
    cfg.ClickHouse.Database = getEnvOrDefault("CLICKHOUSE_DB", "default")
    cfg.ClickHouse.MaxOpenConns = getEnvAsIntOrDefault("CLICKHOUSE_MAX_OPEN_CONNS", 10)
    cfg.ClickHouse.MaxIdleConns = getEnvAsIntOrDefault("CLICKHOUSE_MAX_IDLE_CONNS", 5)
    cfg.ClickHouse.ConnMaxLifetime = time.Duration(getEnvAsIntOrDefault("CLICKHOUSE_CONN_MAX_LIFETIME_MINS", 60)) * time.Minute
    cfg.ClickHouse.QueryTimeout = time.Duration(getEnvAsIntOrDefault("CLICKHOUSE_QUERY_TIMEOUT_SECS", 30)) * time.Second
    cfg.ClickHouse.Debug = getEnvOrDefault("APP_ENV", "production") != "production"

    return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intVal, err := strconv.Atoi(value); err == nil {
            return intVal
        }
    }
    return defaultValue
}
