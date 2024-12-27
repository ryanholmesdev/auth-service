package handlers

import (
	"auth-service/config"
	"auth-service/generated"
	"auth-service/services"
	"golang.org/x/oauth2"
	"log"
	"net/http"
)

// Server implements the generated.ServerInterface
type Server struct{}

// GetAuthProviderLogin handles login requests
func (s *Server) GetAuthProviderLogin(w http.ResponseWriter, r *http.Request, provider string) {
	oauthConfig, exists := config.Providers[provider]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	authURL := oauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// GetAuthProviderCallback handles OAuth callback
func (s *Server) GetAuthProviderCallback(w http.ResponseWriter, r *http.Request, provider string) {
	oauthConfig, exists := config.Providers[provider]
	if !exists {
		http.Error(w, "Unsupported provider", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Authorization code not provided", http.StatusBadRequest)
		return
	}
	token, err := oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sessionID := "unique-session-id" // Generate or retrieve a real session ID here
	services.StoreToken(sessionID, provider, token)

	w.Write([]byte("Authentication successful!"))
}

// GetAuthToken retrieves the token for a session
func (s *Server) GetAuthToken(w http.ResponseWriter, r *http.Request, params generated.GetAuthTokenParams) {
	token, found := services.GetToken(params.SessionId, "provider") // Replace "provider" with actual logic
	if !found {
		http.Error(w, "Token not found", http.StatusNotFound)
		return
	}

	log.Println(token)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Token retrieved successfully!"}`))
}
