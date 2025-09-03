package auth_handlers

import (
	"encoding/json"
    "net/http"
	"context"
	"time"

	mongo_helper "crypto_rates_auth/internal/mongo"
	"crypto_rates_auth/internal/token"
)


func LogStream(mg *mongo_helper.Client) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "[auth/log/stream] method not allowed", http.StatusMethodNotAllowed)

            return
        }

		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "[auth/log/rest] Content-Type must be application/json", http.StatusUnsupportedMediaType)

			return
		}

        var entry mongo_helper.LogEntry 
        if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
            http.Error(w, "[auth/log/stream] invalid json", http.StatusBadRequest)

            return
        }

		curConn, err := token.GetActiveConnFromContext(r.Context())
		if err != nil {
            http.Error(w, "[auth/log/rest] invalid cur conn", http.StatusBadRequest)

            return
		}

		entry.Email = curConn.Email 
		entry.Secret = curConn.Secret

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

        if err := mg.InsertStreamLog(ctx, entry); err != nil {
            http.Error(w, "[auth/log/stream] failed to save rest log", http.StatusInternalServerError)

            return
        }

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

        _, _ = w.Write([]byte(`{"status":"success"}`))
    }
}

