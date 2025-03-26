package handlers

import (
	"auth-service/services"
	"encoding/json"
	"fmt"
	"github.com/monzo/slog"
	"net/http"
)

func (s *Server) GetAuthStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slog.Info(ctx, "Getting auth status", nil)

	// Retrieve session ID from the cookie
	sessionCookie, err := r.Cookie("session_id")
	if err != nil || sessionCookie.Value == "" {
		slog.Error(ctx, "Session ID is required", fmt.Errorf("missing session cookie"), nil)
		http.Error(w, "Session ID is required", http.StatusUnauthorized)
		return
	}

	connectedProviders, err := services.GetLoggedInProviders(sessionCookie.Value)
	if err != nil {
		slog.Error(ctx, "Unable to get logged in providers", err, map[string]interface{}{
			"session_id": sessionCookie.Value,
		})
		http.Error(w, "Unable to get logged in providers", http.StatusInternalServerError)
		return
	}

	slog.Info(ctx, "Successfully retrieved auth status", map[string]interface{}{
		"session_id": sessionCookie.Value,
		"providers":  len(connectedProviders),
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(connectedProviders)
}
