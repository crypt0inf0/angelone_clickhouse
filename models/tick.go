package models

import "time"

type MarketTick struct {
    Timestamp   time.Time `ch:"timestamp"`
    Symbol      string    `ch:"symbol"`
    LastPrice   float64   `ch:"last_price"`
    Volume      int64     `ch:"volume"`
    BidPrice    float64   `ch:"bid_price"`
    AskPrice    float64   `ch:"ask_price"`
    OpenPrice   float64   `ch:"open_price"`
    HighPrice   float64   `ch:"high_price"`
    LowPrice    float64   `ch:"low_price"`
    ClosePrice  float64   `ch:"close_price"`
}
