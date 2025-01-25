package models

type TokenConfig struct {
    Symbol   string `json:"symbol"`
    Token    string `json:"token"`
    Exchange string `json:"exchange"`
}

const (
    // Actions
    SubscribeAction   = 1
    UnsubscribeAction = 0

    // Subscription Modes
    LtpMode    = 1
    QuoteMode  = 2
    SnapQuote  = 3
    DepthMode  = 4

    // Exchange Types
    NSE_CM  = 1
    NSE_FO  = 2
    BSE_CM  = 3
    BSE_FO  = 4
    MCX_FO  = 5
    NCX_FO  = 7
    CDE_FO  = 13
)

var ExchangeMap = map[string]int{
    "NSE_CM":  NSE_CM,
    "NSE_FO":  NSE_FO,
    "BSE_CM":  BSE_CM,
    "BSE_FO":  BSE_FO,
    "MCX_FO":  MCX_FO,
    "NCX_FO":  NCX_FO,
    "CDE_FO":  CDE_FO,
}
