package handlers

import (
	"auth-service/services"
	"encoding/json"
	"net/http"

	"github.com/monzo/slog"
)

func (s *Server) GetAuthStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Retrieve session ID from the cookie
	sessionCookie, err := r.Cookie("session_id")
	if err != nil || sessionCookie.Value == "" {
		slog.Error(ctx, "Session ID is required", err, map[string]interface{}{
			"status_code": http.StatusUnauthorized,
		})
		http.Error(w, "Session ID is required", http.StatusUnauthorized)
		return
	}

	connectedProviders, err := services.GetLoggedInProviders(sessionCookie.Value)
	if err != nil {
		slog.Error(ctx, "Failed to get logged in providers", err, map[string]interface{}{
			"session_id":  sessionCookie.Value,
			"status_code": http.StatusInternalServerError,
		})
		http.Error(w, "Unable to get logged in providers", http.StatusInternalServerError)
		return
	}

	slog.Info(ctx, "Successfully retrieved auth status", map[string]interface{}{
		"session_id":          sessionCookie.Value,
		"connected_providers": len(connectedProviders),
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(connectedProviders)
}
