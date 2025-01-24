package models

import "time"

type TokenStats struct {
    Token       string
    LastUpdate  time.Time
    TickCount   int64
    MinPrice    float64
    MaxPrice    float64
    AvgPrice    float64
    TotalVolume int64
}

type WorkerStats struct {
    WorkerID    int
    ProcessedCount int64
    ErrorCount    int64
    LastProcessed time.Time
}
