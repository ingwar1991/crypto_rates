package token_test

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "crypto_rates_auth/internal/config"
    active_conns "crypto_rates_auth/internal/connections"
    "crypto_rates_auth/internal/token"
    mongo_helper "crypto_rates_auth/internal/mongo"
)


func TestCreateAndVerifyJwt(t *testing.T) {
    cfg := &config.Config{JWTSecret: "mysecret"}
    _, claims, err := token.CreateJwt(cfg, "user@example.com", "apikey123")
    if err != nil {
        t.Fatalf("CreateJwt failed: %v", err)
    }

    if claims["sub"] != "user@example.com" {
        t.Errorf("expected sub claim to be 'user@example.com'")
    }
}

func TestJwtMiddleware_Success(t *testing.T) {
    cfg := &config.Config{JWTSecret: "mysecret"}
    tokenStr, _, _ := token.CreateJwt(cfg, "user@example.com", "apikey123")

    conns := active_conns.NewActiveConnections(context.Background())
    conns.Add(&mongo_helper.ActiveSecret{
        Secret:     tokenStr,
        ExpiresAt: time.Now().Add(5 * time.Minute),
    })

    handler := token.JwtMiddleware(func(w http.ResponseWriter, r *http.Request) {
        conn, err := token.GetActiveConnFromContext(r.Context())
        if err != nil || conn.Secret != tokenStr {
            t.Errorf("failed to retrieve active connection from context")
        }

        w.WriteHeader(http.StatusOK)
    }, cfg, conns)

    req := httptest.NewRequest("GET", "/", nil)
    req.Header.Set("Authorization", "Bearer "+tokenStr)
    rr := httptest.NewRecorder()

    handler(rr, req)

    if rr.Code != http.StatusOK {
        t.Errorf("expected 200 OK, got %d", rr.Code)
    }
}

func TestJwtMiddleware_InvalidToken(t *testing.T) {
    cfg := &config.Config{JWTSecret: "mysecret"}
    conns := active_conns.NewActiveConnections(context.Background())

    handler := token.JwtMiddleware(func(w http.ResponseWriter, r *http.Request) {
        t.Error("handler should not be called with invalid token")
    }, cfg, conns)

    req := httptest.NewRequest("GET", "/", nil)
    req.Header.Set("Authorization", "Bearer invalidtoken")
    rr := httptest.NewRecorder()

    handler(rr, req)

    if rr.Code != http.StatusUnauthorized {
        t.Errorf("expected 401 Unauthorized, got %d", rr.Code)
    }
}

