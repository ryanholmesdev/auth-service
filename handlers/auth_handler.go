package handlers

import (
	"auth-service/config"
	"auth-service/generated"
	"auth-service/services"
	"auth-service/utils"
	"encoding/json"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"log"
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

	// Validate the redirect URI
	allowedDomains, err := utils.GetAllowedRedirectDomains()
	if !utils.ValidateRedirectURI(redirectURI, allowedDomains) {
		http.Error(w, "Invalid redirect URI", http.StatusBadRequest)
		return
	}

	// Generate the state parameter including the redirect URI
	stateToken := uuid.New().String() // Generate a unique CSRF token to this request
	state := stateToken + "|" + redirectURI

	// Store the state token in Redis with a TTL
	err = services.StoreStateToken(stateToken)
	if err != nil {
		http.Error(w, "Server error while storing state", http.StatusInternalServerError)
		return
	}

	// Generate the authorization URL
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

	// Extract the state parameter and split into token and redirect URI
	parts := strings.SplitN(params.State, "|", 2)
	if len(parts) != 2 {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	stateToken := parts[0]
	redirectURI := parts[1]

	// Validate the redirect URI
	allowedDomains, err := utils.GetAllowedRedirectDomains()
	if !utils.ValidateRedirectURI(redirectURI, allowedDomains) {
		http.Error(w, "Invalid redirect URI", http.StatusBadRequest)
		return
	}

	// Validate the state token
	isValid := services.ValidateStateToken(stateToken)
	if !isValid {
		http.Error(w, "Invalid or expired state token", http.StatusBadRequest)
		return
	}

	// Delete the state token to prevent reuse
	go func(stateToken string) {
		err := services.DeleteStateToken(stateToken)
		if err != nil {
			log.Printf("Failed to delete state token: %v", err)
		}
	}(stateToken)

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Authorization code not provided", http.StatusBadRequest)
		return
	}

	// Exchange the authorization code for a token
	token, err := oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate a session ID and store the token
	sessionID := uuid.New().String()
	err = services.StoreAuthToken(sessionID, provider, token)
	if err != nil {
		http.Error(w, "Failed to store token", http.StatusInternalServerError)
		return
	}

	// Set the session ID as a cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to the original URI
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
