package handlers

import (
	"auth-service/config"
	"auth-service/services"
	"auth-service/utils"
	"encoding/json"
	"net/http"
	"time"
)

func (s *Server) GetAuthProviderToken(w http.ResponseWriter, r *http.Request, provider string) {
	_, exists := config.Providers[provider]
	if !exists {
		http.Error(w, "Unsupported provider", http.StatusBadRequest)
		return
	}

	sessionCookie, err := r.Cookie("session_id")
	if err != nil || sessionCookie.Value == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	// Retrieve the token
	token, found := services.GetAuthToken(sessionCookie.Value, provider)
	if !found {
		http.Error(w, "Token not found", http.StatusNotFound)
		return
	}

	// Check if the token is expired
	if token.Expiry.Before(time.Now()) {
		// Refresh the token
		oauthConfig := config.Providers[provider]
		newToken, err := utils.RefreshAccessTokenFunc(oauthConfig, token.RefreshToken)
		if err != nil {
			http.Error(w, "Failed to refresh token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Update the token in storage
		err = services.StoreAuthToken(sessionCookie.Value, provider, newToken)
		if err != nil {
			http.Error(w, "Failed to store refreshed token", http.StatusInternalServerError)
			return
		}

		token = newToken
	}

	// Return the token
	response := map[string]interface{}{
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
		"expires_in":    token.Expiry.Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
