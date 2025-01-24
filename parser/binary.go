package parser

import (
    "bytes"
    "encoding/binary"
)

type MarketData struct {
    SubscriptionMode       uint8   `json:"subscription_mode"`
    ExchangeType          uint8   `json:"exchange_type"`
    Token                 string  `json:"token"`
    SequenceNumber        int64   `json:"sequence_number"`
    ExchangeTimestamp     int64   `json:"exchange_timestamp"`
    LastTradedPrice       int64   `json:"last_traded_price"`
    LastTradedQuantity    int64   `json:"last_traded_quantity"`
    AverageTradedPrice    int64   `json:"average_traded_price"`
    VolumeTrade          int64   `json:"volume_trade_for_the_day"`
    TotalBuyQuantity     float64 `json:"total_buy_quantity"`
    TotalSellQuantity    float64 `json:"total_sell_quantity"`
    OpenPriceOfTheDay    int64   `json:"open_price_of_the_day"`
    HighPriceOfTheDay    int64   `json:"high_price_of_the_day"`
    LowPriceOfTheDay     int64   `json:"low_price_of_the_day"`
    ClosedPrice          int64   `json:"closed_price"`
}

// Helper methods return adjusted float64 values without modifying the original data
func (md *MarketData) GetLastTradedPrice() float64 {
    return float64(md.LastTradedPrice) / 100.0
}

func (md *MarketData) GetOpenPrice() float64 {
    return float64(md.OpenPriceOfTheDay) / 100.0
}

func (md *MarketData) GetHighPrice() float64 {
    return float64(md.HighPriceOfTheDay) / 100.0
}

func (md *MarketData) GetLowPrice() float64 {
    return float64(md.LowPriceOfTheDay) / 100.0
}

func (md *MarketData) GetClosedPrice() float64 {
    return float64(md.ClosedPrice) / 100.0
}

func ParseBinaryData(data []byte) (*MarketData, error) {
    md := &MarketData{}
    reader := bytes.NewReader(data)

    // Parse the binary data into the struct fields
    binary.Read(reader, binary.LittleEndian, &md.SubscriptionMode)
    binary.Read(reader, binary.LittleEndian, &md.ExchangeType)
    
    tokenBytes := make([]byte, 25)
    reader.Read(tokenBytes)
    md.Token = string(bytes.TrimRight(tokenBytes, "\x00"))
    
    binary.Read(reader, binary.LittleEndian, &md.SequenceNumber)
    binary.Read(reader, binary.LittleEndian, &md.ExchangeTimestamp)
    binary.Read(reader, binary.LittleEndian, &md.LastTradedPrice)
    
    if md.SubscriptionMode >= 2 {
        binary.Read(reader, binary.LittleEndian, &md.LastTradedQuantity)
        binary.Read(reader, binary.LittleEndian, &md.AverageTradedPrice)
        binary.Read(reader, binary.LittleEndian, &md.VolumeTrade)
        binary.Read(reader, binary.LittleEndian, &md.TotalBuyQuantity)
        binary.Read(reader, binary.LittleEndian, &md.TotalSellQuantity)
        binary.Read(reader, binary.LittleEndian, &md.OpenPriceOfTheDay)
        binary.Read(reader, binary.LittleEndian, &md.HighPriceOfTheDay)
        binary.Read(reader, binary.LittleEndian, &md.LowPriceOfTheDay)
        binary.Read(reader, binary.LittleEndian, &md.ClosedPrice)
    }

    return md, nil
}
