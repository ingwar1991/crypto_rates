package auth_handlers

import (
	"context"
	"html/template"
	"net/http"
	"strings"
	"time"
	"fmt"

	mongo_helper "crypto_rates_auth/internal/mongo"
	"crypto_rates_auth/internal/token"
)

func Verify(w http.ResponseWriter, r *http.Request, tmpl *template.Template, mg *mongo_helper.Client) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "[auth/verify] bad request", 400)

		return
	}

	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	code := strings.TrimSpace(r.FormValue("code"))
	if email == "" || code == "" {
		http.Error(w, "[auth/verify] missing fields", 400)

		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Validate code (and consume it)
	codeHash := token.Sha256Hex(code)
	if _, err := mg.ValidateAndConsumeCode(ctx, email, codeHash); err != nil {
		http.Error(w, "[auth/verify] invalid or expired OTP", 400)

		return 
	}

	user, err := mg.FindUser(ctx, email)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		http.Error(w, "[auth/verify] error during accessing users data", 400)

		return 
	}

	if user == nil {
		apiKey, err := token.GenAPIKey() 
		if err != nil {
			http.Error(w, "[auth/verify] failed to generate apiKey", 400)

			return 
		}
		
		user = &mongo_helper.User{
			Email: email, 
			APIKey: apiKey, 
			CreatedAt: time.Now(),
		}

		mg.InsertUser(ctx, *user)
	}

	_ = tmpl.Execute(w, map[string]any{"APIKey": user.APIKey})
}
