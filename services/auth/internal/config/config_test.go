package config_test

import (
    "os"
    "testing"

	"crypto_rates_auth/internal/config"
)

const sampleYAML = `
MONGO:
  ADDR: "localhost:27017"
  USER: "mongo_user"
  PASS: "mongo_pass"
SMTP:
  host: "smtp.example.com"
  port: 587
  user: "smtp_user"
  pass: "smtp_pass"
  from_email: "noreply@example.com"
JWT_SECRET: "supersecretjwtkey"
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
    if cfg.Mongo.Addr != "localhost:27017" {
        t.Errorf("expected Mongo.Addr to be 'localhost:27017', got '%s'", cfg.Mongo.Addr)
    }
    if cfg.SMTP.Port != 587 {
        t.Errorf("expected SMTP.Port to be 587, got %d", cfg.SMTP.Port)
    }
    if cfg.JWTSecret != "supersecretjwtkey" {
        t.Errorf("expected JWTSecret to be 'supersecretjwtkey', got '%s'", cfg.JWTSecret)
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
