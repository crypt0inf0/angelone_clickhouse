package angel

type MarketData struct {
    Token                  string  `json:"token"`
    LastTradedPrice       float64 `json:"last_traded_price"`
    OpenPriceOfTheDay    float64 `json:"open_price_of_the_day"`
    HighPriceOfTheDay    float64 `json:"high_price_of_the_day"`
    LowPriceOfTheDay     float64 `json:"low_price_of_the_day"`
    ClosedPrice          float64 `json:"closed_price"`
    VolumeTrade          float64 `json:"volume_trade_for_the_day"`
}

type TokenSubscription struct {
    ExchangeType int      `json:"exchangeType"`
    Tokens       []string `json:"tokens"`
}

type SubscribeRequest struct {
    CorrelationID string             `json:"correlationID"`
    Action        int                `json:"action"`
    Params        SubscriptionParams `json:"params"`
}

type SubscriptionParams struct {
    Mode      int                `json:"mode"`
    TokenList []TokenSubscription `json:"tokenList"`
}
