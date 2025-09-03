package main

import (
	"context"
	"log"

	"crypto_rates_collector/internal/config"
	redis_helper "crypto_rates_collector/internal/redis"
	ws_helper "crypto_rates_collector/internal/ws"
	cleanup "crypto_rates_collector/internal/cleanup"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Printf("[worker] Can't get the config: %v\n", err)

        return
    }

    log.Printf("[worker] symbols: %v", cfg.Symbols)


    rw := redis_helper.NewWriter(cfg)
    if err := rw.Client.Ping(context.Background()).Err(); err != nil {
        log.Fatalf("[worker] redis ping failed: %v", err)
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // run cleanup in a separate routine
    go cleanup.CleanupTicks(ctx, rw)

    wsClient := ws_helper.NewClient(cfg, rw)
    wsClient.Run(ctx)
}
