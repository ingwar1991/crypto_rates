package main

import (
	"context"
	"log"

	"crypto_rates_collector/internal/config"
	redis_helper "crypto_rates_collector/internal/redis"
	"crypto_rates_collector/internal/aggregator"
	"crypto_rates_collector/internal/cleanup"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Printf("[ohlcv aggregator] Can't get the config: %v\n", err)

        return
    }

    log.Printf("[ohlcv aggregator] symbols: %v", cfg.Symbols)


    rw := redis_helper.NewWriter(cfg)
    if err := rw.Client.Ping(context.Background()).Err(); err != nil {
        log.Fatalf("[ohlcv aggregator] redis ping failed: %v", err)
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // cleanup in a separate routine
    go cleanup.CleanupCandles(ctx, rw)

    aggregator.Run(ctx, cfg, rw)
}
