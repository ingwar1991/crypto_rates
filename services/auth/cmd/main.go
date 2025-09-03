package main

import (
	"context"
	"html/template"
	"log"
	"net/http"

	active_conns "crypto_rates_auth/internal/connections"
	handlers "crypto_rates_auth/internal/handlers"
	"crypto_rates_auth/internal/token"
	"crypto_rates_auth/internal/config"
	"crypto_rates_auth/internal/mongo"
	"crypto_rates_auth/internal/template"
)

var (
	mg *mongo_helper.Client
	tmpl *template.Template
	conns *active_conns.ActiveConnections
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Printf("[auth server] Can't get the config: %v\n", err)

        return
    }

	mg, err := mongo_helper.NewClient(cfg)
	if err != nil {
        log.Printf("[auth server] Can't access mongo: %v\n", err)

        return
	}

	tmpl, err = template_helper.Load()
	if err != nil {
        log.Printf("[auth server] Can't read template: %v\n", err)

		return
	}

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

	conns = active_conns.NewActiveConnections(ctx)

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.Index(w, r, tmpl)
	})
    http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		handlers.Register(w, r, cfg, tmpl, mg)
	})
    http.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
		handlers.Verify(w, r, tmpl, mg)
	})
    http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		handlers.Token(w, r, cfg, mg, conns)
	})

    http.HandleFunc("/token/refresh", token.JwtMiddleware(handlers.TokenRefresh(cfg, mg, conns), cfg, conns))

    http.HandleFunc("/check_authentication", token.JwtMiddleware(handlers.CheckAuthentication(), cfg, conns)) 

	http.HandleFunc("/log/rest", token.JwtMiddleware(handlers.LogRest(mg), cfg, conns)) 
    http.HandleFunc("/log/websocket", token.JwtMiddleware(handlers.LogStream(mg), cfg, conns))

	log.Printf("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
