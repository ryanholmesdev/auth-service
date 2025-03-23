package handlers

import (
	"auth-service/config"
	"auth-service/generated"
	"auth-service/services"
	"auth-service/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// GetAuthProviderLogin handles login requests.
func (s *Server) GetAuthProviderLogin(w http.ResponseWriter, r *http.Request, provider string, params generated.GetAuthProviderLoginParams) {
	oauthConfig, exists := config.Providers[provider]
	if !exists {
		http.Error(w, "Unsupported provider", http.StatusNotFound)
		return
	}

	// Access and validate the redirect URI.
	redirectURI := params.RedirectUri
	if err := utils.ValidateRedirectURIFromEnv(redirectURI); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate a state token (for CSRF protection) and combine it with the redirect URI.
	stateToken := uuid.New().String()
	state := stateToken + "|" + redirectURI

	// Generate the PKCE code verifier and corresponding challenge.
	verifier, err := utils.GenerateCodeVerifier()
	if err != nil {
		http.Error(w, "Server error while generating code verifier", http.StatusInternalServerError)
		return
	}
	challenge := utils.GenerateCodeChallenge(verifier)

	// Store the PKCE data (here, the code verifier) in Redis using the state token as key.
	// (services.StorePKCEData should marshal the data as needed and set a TTL.)
	if err = services.StorePKCEData(stateToken, verifier); err != nil {
		http.Error(w, "Server error while storing PKCE data", http.StatusInternalServerError)
		return
	}

	// Generate the authorization URL including the PKCE parameters.
	authURL := oauthConfig.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.AccessTypeOffline,
	)

	// Redirect the user to the OAuth provider.
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// GetAuthProviderCallback handles the OAuth callback.
func (s *Server) GetAuthProviderCallback(w http.ResponseWriter, r *http.Request, provider string, params generated.GetAuthProviderCallbackParams) {
	oauthConfig, exists := config.Providers[provider]
	if !exists {
		http.Error(w, "Unsupported provider", http.StatusBadRequest)
		return
	}

	// Extract the state token and original redirect URI from the state parameter.
	parts := strings.SplitN(params.State, "|", 2)
	if len(parts) != 2 {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}
	stateToken := parts[0]
	redirectURI := parts[1]

	// Validate the redirect URI.
	if err := utils.ValidateRedirectURIFromEnv(redirectURI); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Retrieve the code verifier previously stored with this state token.
	codeVerifier, err := services.GetCodeVerifier(stateToken)
	if err != nil || codeVerifier == "" {
		http.Error(w, "Failed to retrieve code verifier", http.StatusInternalServerError)
		return
	}

	// Optionally delete the PKCE data to prevent reuse.
	go func(token string) {
		if err := services.DeletePKCEData(token); err != nil {
			log.Printf("Failed to delete PKCE data: %v", err)
		}
	}(stateToken)

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Authorization code not provided", http.StatusBadRequest)
		return
	}

	// Exchange the authorization code for an access token, providing the code verifier.
	token, err := oauthConfig.Exchange(r.Context(), code,
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
	)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch the user information from the provider.
	user, err := services.GetUserInfo(r.Context(), provider, token)
	if err != nil {
		http.Error(w, "Failed to fetch user information: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate a session ID and store the token along with the user info.
	sessionID := uuid.New().String()
	if err = services.StoreAuthToken(sessionID, provider, user, token); err != nil {
		http.Error(w, "Failed to store token", http.StatusInternalServerError)
		return
	}

	// Set the session ID as a secure cookie.
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect the user back to the original redirect URI.
	http.Redirect(w, r, redirectURI, http.StatusTemporaryRedirect)
}

// PostAuthProviderLogout handles logout requests.
func (s *Server) PostAuthProviderLogout(w http.ResponseWriter, r *http.Request, provider string, params generated.PostAuthProviderLogoutParams) {
	if _, exists := config.Providers[provider]; !exists {
		http.Error(w, "Unsupported provider", http.StatusBadRequest)
		return
	}

	// Retrieve the session ID from cookies.
	sessionCookie, err := r.Cookie("session_id")
	if err != nil || sessionCookie.Value == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}
	sessionID := sessionCookie.Value

	// If a specific user is specified, log out that user.
	if params.UserId != nil && *params.UserId != "" {
		if err := services.DeleteAuthToken(sessionID, provider, *params.UserId); err != nil {
			http.Error(w, "Failed to log out user", http.StatusInternalServerError)
			return
		}
		log.Printf("Logged out user %s from provider %s", *params.UserId, provider)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": fmt.Sprintf("Successfully logged out user %s from provider %s", *params.UserId, provider),
		})
		return
	}

	// Otherwise, log out all users for the provider.
	if err := services.DeleteAllAuthTokensForProvider(sessionID, provider); err != nil {
		http.Error(w, "Failed to log out all users", http.StatusInternalServerError)
		return
	}
	log.Printf("Logged out all users from provider %s", provider)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Successfully logged out all users from provider %s", provider),
	})
}
