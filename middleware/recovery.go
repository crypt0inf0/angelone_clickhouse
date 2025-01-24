package middleware

import (
    "context"
    "runtime/debug"
    "sync"
    "time"
    
    "broker_clickhouse/utils"
    "github.com/sony/gobreaker"
)

var (
    circuitBreaker *gobreaker.CircuitBreaker
    once sync.Once
)

func init() {
    once.Do(func() {
        circuitBreaker = gobreaker.NewCircuitBreaker(gobreaker.Settings{
            Name:        "database-breaker",
            MaxRequests: 3,
            Interval:    10 * time.Second,
            Timeout:     60 * time.Second,
            ReadyToTrip: func(counts gobreaker.Counts) bool {
                failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
                return counts.Requests >= 3 && failureRatio >= 0.6
            },
            OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
                utils.Logger.Infow("Circuit breaker state changed",
                    "from", from.String(),
                    "to", to.String())
            },
        })
    })
}

func WithCircuitBreaker(ctx context.Context, operation string, fn func() error) error {
    _, err := circuitBreaker.Execute(func() (interface{}, error) {
        return nil, fn()
    })
    return err
}

func RecoverMiddleware(next func()) {
    defer func() {
        if r := recover(); r != nil {
            stack := debug.Stack()
            utils.Logger.Errorw("Panic recovered",
                "error", r,
                "stack", string(stack))
            
            // Implement graceful shutdown
            GracefulShutdown()
        }
    }()
    next()
}
