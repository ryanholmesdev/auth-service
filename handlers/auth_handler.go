package handlers

import (
	"auth-service/config"
	"auth-service/generated"
	"auth-service/services"
	"auth-service/utils"
	"encoding/json"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"net/http"
	"strings"
	"time"
)

// Server implements the generated.ServerInterface
type Server struct{}

// GetAuthProviderLogin handles login requests
func (s *Server) GetAuthProviderLogin(w http.ResponseWriter, r *http.Request, provider string, params generated.GetAuthProviderLoginParams) {
	oauthConfig, exists := config.Providers[provider]
	if !exists {
		http.Error(w, "Unsupported provider", http.StatusNotFound)
		return
	}

	// Access the redirect_uri from params
	redirectURI := params.RedirectUri
	if redirectURI == "" {
		http.Error(w, "Redirect URI is required", http.StatusBadRequest)
		return
	}

	// Generate the state parameter including the redirect URI
	state := uuid.New().String() + "|" + redirectURI
	authURL := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)

	// Redirect to the OAuth provider
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (s *Server) GetAuthProviderCallback(w http.ResponseWriter, r *http.Request, provider string, params generated.GetAuthProviderCallbackParams) {
	oauthConfig, exists := config.Providers[provider]
	if !exists {
		http.Error(w, "Unsupported provider", http.StatusBadRequest)
		return
	}

	// Extract the state parameter from params
	state := params.State
	parts := strings.SplitN(state, "|", 2)
	if len(parts) != 2 {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}
	redirectURI := parts[1]

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Authorization code not provided", http.StatusBadRequest)
		return
	}

	// Exchange the code for a token
	token, err := oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate a session ID and store the token
	sessionID := uuid.New().String()
	services.StoreToken(sessionID, provider, token)

	// Redirect to the original URI
	// Set the session ID as a cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",   // Cookie applies to all routes
		HttpOnly: true,  // Prevent JavaScript access
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to the front-end without session ID in the URL
	http.Redirect(w, r, redirectURI, http.StatusTemporaryRedirect)
}

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
	token, found := services.GetToken(sessionCookie.Value, provider)
	if !found {
		http.Error(w, "Token not found", http.StatusNotFound)
		return
	}

	// Check if the token is expired
	if token.Expiry.Before(time.Now()) {
		// Refresh the token
		oauthConfig := config.Providers[provider]
		newToken, err := utils.RefreshAccessToken(oauthConfig, token.RefreshToken)
		if err != nil {
			http.Error(w, "Failed to refresh token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Update the token in storage
		err = services.StoreToken(sessionCookie.Value, provider, newToken)
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
