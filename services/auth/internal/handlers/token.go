package auth_handlers

import (
	"context"
	"net/http"
	"strings"
	"time"
	"fmt"

	"crypto_rates_auth/internal/config"
	mongo_helper "crypto_rates_auth/internal/mongo"
	"crypto_rates_auth/internal/token"
	active_conns "crypto_rates_auth/internal/connections"
)


func Token(w http.ResponseWriter, r *http.Request, cfg *config.Config, mg *mongo_helper.Client, conns *active_conns.ActiveConnections) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "[auth/token] bad request", 400)

		return
	}

	apiKey := strings.TrimSpace(r.FormValue("api_key"))
	if apiKey == "" {
		http.Error(w, "[auth/token] missing api_key", 400)

		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var user mongo_helper.User
	err := mg.UsersCol.FindOne(ctx, map[string]any{
		"api_key": apiKey,
	}).Decode(&user)
	if err != nil {
		http.Error(w, "[auth/token] invalid api_key", 401)

		return
	}

	signed, claims, err := token.CreateJwt(
		cfg,
		user.Email, 
		user.APIKey,
	)
	if err != nil {
		http.Error(w, "[auth/token] sign error", 500)

		return
	}

	activeSecret := mongo_helper.ActiveSecret{
		Email: user.Email,
		Secret: signed,
		CreatedAt: time.Unix(claims["iat"].(int64), 0), 
		ExpiresAt: time.Unix(claims["exp"].(int64), 0), 
	} 
	if err = mg.InsertActiveSecret(ctx, activeSecret); err != nil {
		http.Error(w, "[auth/token] failed to save jwt", 500)

		return
	}

	if err = conns.Add(&activeSecret); err != nil {
		http.Error(w, "[auth/token] failed to save active conn", 500)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write([]byte(fmt.Sprintf(`{"jwt":"%s"}`, signed)))
}
