package aggregator

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"crypto_rates_collector/internal/config"
	redis_helper "crypto_rates_collector/internal/redis"
)

func Run(ctx context.Context, cfg *config.Config, rw *redis_helper.Writer) {
	errCh := make(chan error, len(cfg.Symbols))

	go func() {
		<-ctx.Done()
		log.Printf("[ohlcv aggregator] received cancel signal, closing err chan\n")
		close(errCh)
	}()

	for _, symbol := range cfg.Symbols {
		go aggregateCandle(symbol, ctx, rw, errCh)
	}

	for err := range errCh {
		if err != nil {
			// placeholder for some error handlers
			log.Printf("[ohlcv aggregator] error: %v\n", err)
		}
	}
}

func aggregateCandle(symbol string, ctx context.Context, rw *redis_helper.Writer, errCh chan error) {
	var candle *redis_helper.Candle
	var periodCur time.Time

	lastID := "0"
	for {
		select {
		case <-ctx.Done():
			return
		default:
			res, err := rw.Client.XRead(ctx, &redis.XReadArgs{
				Streams: []string{
					rw.GetTicksStreamKey(symbol),
					lastID,
				},
				Count: 100,
				Block: 5 * time.Second,
			}).Result()

			if err != nil && err != redis.Nil {
				errCh <- fmt.Errorf("Failed to read ticks stream [%s]: %v", rw.GetCandlesStreamKey(symbol), err)

				return
			}

			errCh <- nil

			for _, str := range res {
				for _, msg := range str.Messages {
					lastID = msg.ID

					tsStr := fmt.Sprintf("%v", msg.Values["ts"])
					priceStr := fmt.Sprintf("%v", msg.Values["price"])
					qtyStr := fmt.Sprintf("%v", msg.Values["qty"])

					tsInt, _ := strconv.ParseInt(tsStr, 10, 64)
					price, _ := strconv.ParseFloat(priceStr, 64)
					qty, _ := strconv.ParseFloat(qtyStr, 64)

					ts := time.UnixMilli(tsInt)
					period := ts.Truncate(rw.GetTickTTL())

					if candle == nil {
						candle = redis_helper.NewCandle(symbol, tsInt, price, qty)
						periodCur = period
					} else if period.Equal(periodCur) {
						candle.Add(price, qty)
					} else {
						if err := rw.WriteCandle(candle); err != nil {
							errCh <- err
						}

						candle = redis_helper.NewCandle(symbol, tsInt, price, qty)
						periodCur = period
					}
				}
			}
		}
	}
}
