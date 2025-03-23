package handlers

import (
	"auth-service/config"
	"auth-service/generated"
	"auth-service/models"
	"auth-service/services"
	"auth-service/utils"
	"encoding/json"
	"net/http"
	"time"
)

func (s *Server) GetAuthProviderToken(w http.ResponseWriter, r *http.Request, provider string, params generated.GetAuthProviderTokenParams) {
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

	if params.UserId == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}
	userID := params.UserId

	// Retrieve the token
	token, found := services.GetAuthToken(sessionCookie.Value, provider, userID)
	if !found {
		http.Error(w, "Token not found", http.StatusNotFound)
		return
	}

	// Check if the token is expired
	if token.Token.Expiry.Before(time.Now()) {
		// Refresh the token
		oauthConfig := config.Providers[provider]
		newToken, err := utils.RefreshAccessTokenFunc(oauthConfig, token.Token.RefreshToken)
		if err != nil {
			http.Error(w, "Failed to refresh token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Update the token in storage
		// Extract user info from `token`
		userInfo := &services.UserInfo{
			ID:          token.UserID,
			DisplayName: token.DisplayName,
			Email:       token.Email,
		}

		// Update the token in storage
		err = services.StoreAuthToken(sessionCookie.Value, provider, (*models.UserInfo)(userInfo), newToken)

		token.Token = newToken
	}

	// Return the token
	response := map[string]interface{}{
		"access_token":  token.Token.AccessToken,
		"refresh_token": token.Token.RefreshToken,
		"expires_in":    token.Token.Expiry.Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
