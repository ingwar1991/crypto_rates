package auth_handlers

import (
	"context"
	"net/http"
	"strings"
	"html/template"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"crypto_rates_auth/internal/config"
	"crypto_rates_auth/internal/email"
	"crypto_rates_auth/internal/mongo"
	"crypto_rates_auth/internal/token"
)

func Register(w http.ResponseWriter, r *http.Request, cfg *config.Config, tmpl *template.Template, mg *mongo_helper.Client) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "[auth/start] bad request", 400)

		return
	}

	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	if email == "" {
		http.Error(w, "[auth/start] email required", 400)

		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Upsert user with API key if not exists
	apiKey := ""
	res := mg.UsersCol.FindOne(ctx, bson.M{
		"email": email,
	})

	var user mongo_helper.User 
	if err := res.Decode(&user); err == mongo.ErrNoDocuments {
		apiKey = primitive.NewObjectID().Hex()

		_, err := mg.UsersCol.InsertOne(ctx, mongo_helper.User{
			Email: email, 
			APIKey: apiKey, 
			CreatedAt: time.Now(),
		})
		if err != nil {
			http.Error(w, "[auth/start] db error: failed to insert new user", 500)

			return
		}
	} else if err != nil {
		http.Error(w, "[auth/start] db error: failed to decode user", 500)

		return
	} else {
		apiKey = user.APIKey
	}

	// Create OTP code
	code, err := token.GenOTP() 
	if err != nil {
		http.Error(w, "[auth/start] gen error", 500)

		return
	}

	codeHash := token.Sha256Hex(code)
	exp := time.Now().Add(5 * time.Minute)

	_, err = mg.CodesCol.InsertOne(ctx, mongo_helper.LoginCode{
		Email: email, 
		CodeHash: codeHash, 
		CreatedAt: time.Now(),
		ExpiresAt: exp, 
	})
	if err != nil {
		http.Error(w, "[auth/start] db error: failed ot insert new code", 500)

		return
	}

	// Try to email, else reveal html page 
	emailed := true
	if err := email_helper.SendOTP(cfg, email, code); err != nil {
		emailed = false
	}

	data := map[string]any{
		"ShowVerify": true,
		"Email":      email,
		"Emailed":    emailed,
	}
	if !emailed {
		data["RevealCode"] = code
	}

	_ = tmpl.Execute(w, data)
	// _ = apiKey // not shown on page, but available for later use if needed
}
