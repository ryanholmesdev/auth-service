package handlers

import (
	"auth-service/config"
	"auth-service/services"
	"encoding/json"
	"net/http"
	"time"
)

func (s *Server) GetAuthStatus(w http.ResponseWriter, r *http.Request) {
	// Retrieve session ID from the cookie
	sessionCookie, err := r.Cookie("session_id")
	if err != nil || sessionCookie.Value == "" {
		http.Error(w, "Session ID is required", http.StatusUnauthorized)
		return
	}

	connectedProviders := make(map[string]bool)

	// Loop through all configured providers
	for provider := range config.Providers {
		//
		//Check if a valid token exists for each provider
		token, found := services.GetAuthToken(sessionCookie.Value, provider)
		if found && token.Expiry.After(time.Now()) {
			connectedProviders[provider] = true
		} else {
			connectedProviders[provider] = false
		}
	}

	// Return the connection status as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(connectedProviders)
}
