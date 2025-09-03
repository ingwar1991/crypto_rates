package token

import (
	"strings"
    "crypto/sha256"
    "encoding/hex"
	"math/big"
	"crypto/rand"
	"context"
	"fmt"
	"time"
	"net/http"
	"errors"

	jwt "github.com/golang-jwt/jwt/v5"

	"crypto_rates_auth/internal/config"
	active_conns "crypto_rates_auth/internal/connections"
	mongo_helper "crypto_rates_auth/internal/mongo"
)

var ctxActiveConnKey string = "active_conn"

func Sha256Hex(s string) string {
    sum := sha256.Sum256([]byte(s))

    return hex.EncodeToString(sum[:])
}

func RandomDigits(n int) (string, error) {
	var b strings.Builder
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil { 
			return "", err 
		}

		b.WriteByte(byte('0' + num.Int64()))
	}

	return b.String(), nil
}

func GenOTP() (string, error) {
	return RandomDigits(6)
}

func GenAPIKey() (_ string, err error) {
	b := make([]byte, 32) // 256-bit key
	if _, err = rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), err
}

func getJwtFromRequest(r *http.Request) (string, error) {
    authHeader := r.Header.Get("Authorization")
    if authHeader == "" {
        return "", fmt.Errorf("[getJwtFromRequest] missing Authorization header")
    }

    parts := strings.Fields(authHeader)
    if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
        return "", fmt.Errorf("[getJwtFromRequest] invalid Authorization header format")
    }

    return parts[1], nil
}

func verifyJwt(tokenString, jwtSecret string) error {
    token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
        if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("[verifyJwt] unexpected signing method")
        }

        return []byte(jwtSecret), nil
    })

    if err != nil || !token.Valid {
        return fmt.Errorf("[verifyJwt] invalid token")
    }

    return nil
}

func JwtMiddleware(handler http.HandlerFunc, cfg *config.Config, conns *active_conns.ActiveConnections) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        tokenString, err := getJwtFromRequest(r)
        if err != nil {
            http.Error(w, err.Error(), http.StatusUnauthorized)

            return
        }

        if err := verifyJwt(tokenString, cfg.JWTSecret); err != nil {
            http.Error(w, err.Error(), http.StatusUnauthorized)

            return
        }

		curConn, ok := conns.Get(tokenString)
		if !ok {
			http.Error(w, fmt.Sprintf("[auth/middleware] Failed to obtain active conn: %v, %v", curConn, ok), http.StatusUnauthorized)

            return
		}

		ctx := context.WithValue(r.Context(), ctxActiveConnKey, curConn)
        handler(w, r.WithContext(ctx))
    }
}

func GetActiveConnFromContext(ctx context.Context) (*mongo_helper.ActiveSecret, error) {
    val := ctx.Value(ctxActiveConnKey)
    if val == nil {
        return nil, errors.New("[GetActiveConnFromContext] active secret missing in context")
    }

    conn, ok := val.(*mongo_helper.ActiveSecret)
    if !ok {
        return nil, errors.New("[GetActiveConnFromContext] invalid active secret type")
    }

    return conn, nil
}

func CreateJwt(cfg *config.Config, email, apiKey string) (string, jwt.MapClaims, error) {
	tNow := time.Now()
	claims := jwt.MapClaims{
		"sub": email, 
		"api": apiKey, 
		"iat": tNow.Unix(),
		"exp": tNow.Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(cfg.JWTSecret))

	return signed, claims, err
}
