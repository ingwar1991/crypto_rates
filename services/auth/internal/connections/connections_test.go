package auth_connections_test

import (
    "context"
    "testing"
    "time"

	mongo_helper "crypto_rates_auth/internal/mongo"
    active_conns "crypto_rates_auth/internal/connections"
)

func TestAddAndGet(t *testing.T) {
    ctx := context.Background()
    ac := active_conns.NewActiveConnections(ctx)

    secret := "test-secret"
    expiration := time.Now().Add(10 * time.Second)
    mock := &mongo_helper.ActiveSecret{Secret: secret, ExpiresAt: expiration}

    err := ac.Add((*mongo_helper.ActiveSecret)(mock))
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }

    got, ok := ac.Get(secret)
    if !ok || got.Secret != secret {
        t.Errorf("expected to retrieve secret %s, got %v", secret, got)
    }
}

func TestAddExpiredSecret(t *testing.T) {
    ctx := context.Background()
    ac := active_conns.NewActiveConnections(ctx)

    expired := &mongo_helper.ActiveSecret{
        Secret:     "expired-secret",
        ExpiresAt: time.Now().Add(-1 * time.Minute),
    }

    err := ac.Add((*mongo_helper.ActiveSecret)(expired))
    if err == nil {
        t.Error("expected error when adding expired secret, got nil")
    }
}

func TestCheck(t *testing.T) {
    ctx := context.Background()
    ac := active_conns.NewActiveConnections(ctx)

    valid := &mongo_helper.ActiveSecret{
        Secret:     "valid-secret",
        ExpiresAt: time.Now().Add(10 * time.Second),
    }
    ac.Add((*mongo_helper.ActiveSecret)(valid))

    if !ac.Check(valid.Secret) {
        t.Error("expected Check to return true for valid secret")
    }

    ac.Remove(valid.Secret)
    if ac.Check(valid.Secret) {
        t.Error("expected Check to return false after removal")
    }
}

func TestCleanup(t *testing.T) {
    ctx := context.Background()
    ac := active_conns.NewActiveConnections(ctx)

    expired := &mongo_helper.ActiveSecret{
        Secret:     "expired",
        ExpiresAt: time.Now().Add(-1 * time.Minute),
    }
    valid := &mongo_helper.ActiveSecret{
        Secret:     "valid",
        ExpiresAt: time.Now().Add(10 * time.Second),
    }

    ac.Add((*mongo_helper.ActiveSecret)(expired))
    ac.Add((*mongo_helper.ActiveSecret)(valid))

    ac.Cleanup()

    if _, ok := ac.Get("expired"); ok {
        t.Error("expected expired secret to be removed")
    }
    if _, ok := ac.Get("valid"); !ok {
        t.Error("expected valid secret to remain")
    }
}
