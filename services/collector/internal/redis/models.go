package redis_helper

type TradeEvent struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
	TradeTime int64  `json:"T"`
	Symbol    string `json:"s"`
	Price     string `json:"p"`
	Quantity  string `json:"q"`
}

type Candle struct {
    Symbol string
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
	Start  int64 
}

func NewCandle(symbol string, start int64, price float64, qty float64) *Candle{
    return &Candle{
        Symbol: symbol,
        Open: price,
        High: price,
        Low: price,
        Close: price,
        Volume: qty,
        Start: start,
    }
}

func (c *Candle) Add(price float64, qty float64) {
    c.Close = price
    c.Volume += qty

    if price < c.Low {
        c.Low = price
    } else if price > c.High {
        c.High = price
    }
}
