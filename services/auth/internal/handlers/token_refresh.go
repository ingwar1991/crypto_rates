package auth_handlers

import (
	"context"
	"net/http"
	"time"
	"fmt"

	"crypto_rates_auth/internal/config"
	mongo_helper "crypto_rates_auth/internal/mongo"
	"crypto_rates_auth/internal/token"
	active_conns "crypto_rates_auth/internal/connections"
)


func TokenRefresh(cfg *config.Config, mg *mongo_helper.Client, conns *active_conns.ActiveConnections) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
		curConn, err := token.GetActiveConnFromContext(r.Context())
		if err != nil {
			http.Error(w, "[auth/log/rest] invalid cur conn", http.StatusBadRequest)

			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		var user mongo_helper.User
		err = mg.UsersCol.FindOne(ctx, map[string]any{
			"email": curConn.Email,
		}).Decode(&user)
		if err != nil {
			http.Error(w, "[auth/token/refresh] user not found", 401)

			return
		}

		signed, claims, err := token.CreateJwt(
			cfg,
			curConn.Email, 
			user.APIKey,
		) 
		if err != nil {
			http.Error(w, "[auth/token] sign error", 500)

			return
		}

		activeSecret := mongo_helper.ActiveSecret{
			Email: curConn.Email,
			Secret: signed,
			CreatedAt: time.Unix(claims["iat"].(int64), 0), 
			ExpiresAt: time.Unix(claims["exp"].(int64), 0), 
		} 
		if err = mg.InsertActiveSecret(ctx, activeSecret); err != nil {
			http.Error(w, "[auth/token/refresh] failed to save jwt", 500)

			return
		}

		// remove old active
		mg.RemoveActiveSecret(ctx, *curConn)
		conns.Remove(curConn.Secret)

		if err = conns.Add(&activeSecret); err != nil {
			http.Error(w, "[auth/token] failed to save active conn", 500)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte(fmt.Sprintf(`{"jwt":"%s"}`, signed)))
	}
}
