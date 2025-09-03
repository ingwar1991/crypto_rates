package config_test

import (
	"os"
	"reflect"
	"testing"
	"time"

	"crypto_rates_collector/internal/config"
)

const sampleYAML = `
BINANCE_WS: wss://stream.binance.com:9443
SYMBOLS: [BTCEUR,ETHEUR,LTCEUR]
REDIS:
  ADDR: crypto_rates_redis:6379
  PASSWORD: ""
  DB: 0
TICK_TTL_SECONDS: 10 
CANDLE_TTL_SECONDS: 300 # 5m 
`

func TestLoadConfig(t *testing.T) {
    // Create a temporary config.yaml file
    tmpFile := "config.yaml"
    err := os.WriteFile(tmpFile, []byte(sampleYAML), 0644)
    if err != nil {
        t.Fatalf("failed to write temp config file: %v", err)
    }
    defer os.Remove(tmpFile)

    cfg, err := config.Load()
    if err != nil {
        t.Fatalf("Load() returned error: %v", err)
    }

    // Validate parsed values
    if cfg.Redis.Addr != "crypto_rates_redis:6379" {
        t.Errorf("expected Redis.Addr to be 'crypto_rates_redis:6379', got '%s'", cfg.Redis.Addr)
    }
    if !reflect.DeepEqual(cfg.Symbols, []string{"BTCEUR", "ETHEUR", "LTCEUR"}) {
        t.Errorf("expected Symbols to be [BTCEUR, ETHEUR, LTCEUR], got '%s'", cfg.Symbols)
    }

    if cfg.TickTTLInt != 10 {
        t.Errorf("expected TickTTlInt to be 10, got %d", cfg.TickTTLInt)
    }
	cfg.ReadTTL()
    if cfg.TickTTL != time.Duration(10 * time.Second) {
        t.Errorf("expected TickTTl to be 10s, got %v", cfg.TickTTL)
    }
}

func TestLoadMissingFile(t *testing.T) {
    // Ensure config.yaml doesn't exist
    _ = os.Remove("config.yaml")

    _, err := config.Load()
    if err == nil {
        t.Error("expected error when config.yaml is missing, got nil")
    }
}
