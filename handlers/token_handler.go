package handlers

import (
	"auth-service/config"
	"auth-service/generated"
	"auth-service/models"
	"auth-service/services"
	"auth-service/utils"
	"encoding/json"
	"fmt"
	"github.com/monzo/slog"
	"net/http"
	"time"
)

func (s *Server) GetAuthProviderToken(w http.ResponseWriter, r *http.Request, provider string, params generated.GetAuthProviderTokenParams) {
	ctx := r.Context()
	slog.Info(ctx, "Getting auth provider token", map[string]interface{}{
		"provider": provider,
		"user_id":  params.UserId,
	})

	_, exists := config.Providers[provider]
	if !exists {
		slog.Error(ctx, "Unsupported provider", fmt.Errorf("provider not found"), map[string]interface{}{
			"provider": provider,
		})
		http.Error(w, "Unsupported provider", http.StatusBadRequest)
		return
	}

	sessionCookie, err := r.Cookie("session_id")
	if err != nil || sessionCookie.Value == "" {
		slog.Error(ctx, "Session ID is required", fmt.Errorf("missing session cookie"), nil)
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	if params.UserId == "" {
		slog.Error(ctx, "User ID is required", fmt.Errorf("missing user ID"), nil)
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}
	userID := params.UserId

	// Retrieve the token
	token, found := services.GetAuthToken(sessionCookie.Value, provider, userID)
	if !found {
		slog.Error(ctx, "Token not found", fmt.Errorf("token not found"), map[string]interface{}{
			"session_id": sessionCookie.Value,
			"provider":   provider,
			"user_id":    userID,
		})
		http.Error(w, "Token not found", http.StatusNotFound)
		return
	}

	// Check if the token is expired
	if token.Token.Expiry.Before(time.Now()) {
		slog.Info(ctx, "Token expired, refreshing", map[string]interface{}{
			"session_id": sessionCookie.Value,
			"provider":   provider,
			"user_id":    userID,
			"expired_at": token.Token.Expiry,
		})

		// Refresh the token
		oauthConfig := config.Providers[provider]
		newToken, err := utils.RefreshAccessTokenFunc(oauthConfig, token.Token.RefreshToken)
		if err != nil {
			slog.Error(ctx, "Failed to refresh token", err, map[string]interface{}{
				"session_id": sessionCookie.Value,
				"provider":   provider,
				"user_id":    userID,
			})
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
		if err != nil {
			slog.Error(ctx, "Failed to store refreshed token", err, map[string]interface{}{
				"session_id": sessionCookie.Value,
				"provider":   provider,
				"user_id":    userID,
			})
			http.Error(w, "Failed to store refreshed token", http.StatusInternalServerError)
			return
		}

		slog.Info(ctx, "Successfully refreshed token", map[string]interface{}{
			"session_id": sessionCookie.Value,
			"provider":   provider,
			"user_id":    userID,
			"expires_at": newToken.Expiry,
		})

		token.Token = newToken
	}

	// Return the token
	response := map[string]interface{}{
		"access_token":  token.Token.AccessToken,
		"refresh_token": token.Token.RefreshToken,
		"expires_in":    token.Token.Expiry.Unix(),
	}

	slog.Info(ctx, "Successfully retrieved token", map[string]interface{}{
		"session_id": sessionCookie.Value,
		"provider":   provider,
		"user_id":    userID,
		"expires_at": token.Token.Expiry,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
