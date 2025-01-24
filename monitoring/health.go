package monitoring

import (
    "encoding/json"
    "net/http"
    "runtime"
    "time"
)

type HealthStatus struct {
    Status            string            `json:"status"`
    Uptime           string            `json:"uptime"`
    StartTime        time.Time         `json:"start_time"`
    MemoryUsage      uint64            `json:"memory_usage"`
    GoroutineCount   int               `json:"goroutine_count"`
    LastError        string            `json:"last_error,omitempty"`
    ComponentStatus   map[string]string `json:"component_status"`
    DatabaseLatency  float64           `json:"database_latency"`
}

var (
    startTime     = time.Now()
    lastError     string
    healthChecks  = make(map[string]func() bool)
)

func RegisterHealthCheck(name string, check func() bool) {
    healthChecks[name] = check
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    status := HealthStatus{
        Status:         "ok",
        Uptime:        time.Since(startTime).String(),
        StartTime:     startTime,
        MemoryUsage:   m.Alloc,
        GoroutineCount: runtime.NumGoroutine(),
        LastError:     lastError,
        ComponentStatus: make(map[string]string),
    }

    // Check all registered components
    for name, check := range healthChecks {
        if check() {
            status.ComponentStatus[name] = "healthy"
        } else {
            status.ComponentStatus[name] = "unhealthy"
            status.Status = "degraded"
        }
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(status)
}
