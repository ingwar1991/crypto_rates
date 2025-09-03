package redis_helper

import (
	"context"
	"fmt"
	"strings"
	"time"

	redis "github.com/redis/go-redis/v9"

	"crypto_rates_collector/internal/config"
)

type Writer struct {
	Client *redis.Client
	ctx    context.Context

	tickTTL   time.Duration
	candleTTL time.Duration

	availableSymbols []string
}

func NewWriter(cfg *config.Config) *Writer {
	cli := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	return &Writer{
		Client: cli,
		ctx:    context.Background(),

		tickTTL:   cfg.TickTTL,
		candleTTL: cfg.CandleTTL,

		availableSymbols: cfg.Symbols,
	}
}

func (w *Writer) GetContext() context.Context {
	return w.ctx
}

func (w *Writer) GetTicksStreamKey(symbol string) string {
	return fmt.Sprintf("ticks:%s", strings.ToUpper(symbol))
}

func (w *Writer) GetCandlesStreamKey(symbol string) string {
	return fmt.Sprintf("candles:%v:%s", w.tickTTL, strings.ToUpper(symbol))
}

func (w *Writer) GetAvailableTicksStreams() []string {
	streams := make([]string, len(w.availableSymbols))
	for i, symbol := range w.availableSymbols {
		streams[i] = w.GetTicksStreamKey(symbol)
	}

	return streams
}

func (w *Writer) GetAvailableCandlesStreams() []string {
	streams := make([]string, len(w.availableSymbols))
	for i, symbol := range w.availableSymbols {
		streams[i] = w.GetCandlesStreamKey(symbol)
	}

	return streams
}

func (w *Writer) GetTickTTL() time.Duration {
	return w.tickTTL
}

func (w *Writer) GetCandleTTL() time.Duration {
	return w.candleTTL
}

func (w *Writer) WriteTick(t *TradeEvent) (err error) {
	tm := time.Unix(t.TradeTime/1000, (t.TradeTime%1000)*int64(time.Millisecond))
	_, err = w.Client.XAdd(w.ctx, &redis.XAddArgs{
		Stream: w.GetTicksStreamKey(t.Symbol),
		Values: map[string]any{
			"ts":    t.TradeTime, // ms
			"tm":    tm.Format("2006-01-02 15:04:05.000"),
			"price": t.Price,
			"qty":   t.Quantity,
		},
	}).Result()

	return err
}

func (w *Writer) WriteCandle(c *Candle) (err error) {
	// tm := time.Unix(c.Start/1000, (c.Start%1000)*int64(time.Millisecond))
	_, err = w.Client.XAdd(w.ctx, &redis.XAddArgs{
		Stream: w.GetCandlesStreamKey(c.Symbol),
		Values: map[string]any{
			"ts": c.Start, // ms
			// used for testing
			// "tm": tm.Format("2006-01-02 15:04:05.000"),
			"o": c.Open,
			"h": c.High,
			"l": c.Low,
			"c": c.Close,
			"v": c.Volume,
		},
	}).Result()

	return err
}
