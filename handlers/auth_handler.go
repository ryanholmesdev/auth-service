package handlers

import (
	"auth-service/config"
	"auth-service/generated"
	"auth-service/services"
	"auth-service/utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/monzo/slog"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// GetAuthProviderLogin handles login requests.
func (s *Server) GetAuthProviderLogin(w http.ResponseWriter, r *http.Request, provider string, params generated.GetAuthProviderLoginParams) {
	ctx := r.Context()

	slog.Info(ctx, "Starting auth provider login", map[string]interface{}{
		"provider":     provider,
		"redirect_uri": params.RedirectUri,
	})

	oauthConfig, exists := config.Providers[provider]
	if !exists {
		slog.Error(ctx, "Unsupported provider", fmt.Errorf("provider not found"), map[string]interface{}{
			"provider": provider,
		})
		http.Error(w, "Unsupported provider", http.StatusNotFound)
		return
	}

	// Access and validate the redirect URI.
	redirectURI := params.RedirectUri
	if err := utils.ValidateRedirectURIFromEnv(redirectURI); err != nil {
		slog.Error(ctx, "Invalid redirect URI", err, map[string]interface{}{
			"redirect_uri": redirectURI,
		})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate a state token (for CSRF protection) and combine it with the redirect URI.
	stateToken := uuid.New().String()
	state := stateToken + "|" + redirectURI

	// Generate the PKCE code verifier and corresponding challenge.
	verifier, err := utils.GenerateCodeVerifier()
	if err != nil {
		slog.Error(ctx, "Failed to generate code verifier", err, nil)
		http.Error(w, "Server error while generating code verifier", http.StatusInternalServerError)
		return
	}
	challenge := utils.GenerateCodeChallenge(verifier)

	// Store the PKCE data
	if err = services.StorePKCEData(stateToken, verifier); err != nil {
		slog.Error(ctx, "Failed to store PKCE data", err, map[string]interface{}{
			"state_token": stateToken,
		})
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

	slog.Info(ctx, "Redirecting to auth provider", map[string]interface{}{
		"provider": provider,
		"auth_url": authURL,
	})

	// Redirect the user to the OAuth provider.
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// GetAuthProviderCallback handles the OAuth callback.
func (s *Server) GetAuthProviderCallback(w http.ResponseWriter, r *http.Request, provider string, params generated.GetAuthProviderCallbackParams) {
	ctx := r.Context()
	slog.Info(ctx, "Received auth provider callback", map[string]interface{}{
		"provider": provider,
		"state":    params.State,
	})

	oauthConfig, exists := config.Providers[provider]
	if !exists {
		slog.Error(ctx, "Unsupported provider", fmt.Errorf("provider not found"), map[string]interface{}{
			"provider": provider,
		})
		http.Error(w, "Unsupported provider", http.StatusBadRequest)
		return
	}

	// Extract the state token and original redirect URI from the state parameter.
	parts := strings.SplitN(params.State, "|", 2)
	if len(parts) != 2 {
		slog.Error(ctx, "Invalid state parameter", fmt.Errorf("invalid state format"), map[string]interface{}{
			"state": params.State,
		})
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}
	stateToken := parts[0]
	redirectURI := parts[1]

	// Validate the redirect URI.
	if err := utils.ValidateRedirectURIFromEnv(redirectURI); err != nil {
		slog.Error(ctx, "Invalid redirect URI", err, map[string]interface{}{
			"redirect_uri": redirectURI,
		})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Retrieve the code verifier previously stored with this state token.
	codeVerifier, err := services.GetCodeVerifier(stateToken)
	if err != nil || codeVerifier == "" {
		slog.Error(ctx, "Failed to retrieve code verifier", err, map[string]interface{}{
			"state_token": stateToken,
		})
		http.Error(w, "Failed to retrieve code verifier", http.StatusInternalServerError)
		return
	}

	// Optionally delete the PKCE data to prevent reuse.
	go func(token string) {
		if err := services.DeletePKCEData(token); err != nil {
			log.Printf("Failed to delete PKCE data: %v", err)
			slog.Error(context.Background(), "Failed to delete PKCE data", err, map[string]interface{}{
				"state_token": token,
			})
		}
	}(stateToken)

	code := r.URL.Query().Get("code")
	if code == "" {
		slog.Error(ctx, "Authorization code not provided", fmt.Errorf("missing code"), nil)
		http.Error(w, "Authorization code not provided", http.StatusBadRequest)
		return
	}

	// Exchange the authorization code for an access token
	token, err := oauthConfig.Exchange(r.Context(), code,
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
	)
	if err != nil {
		slog.Error(ctx, "Failed to exchange token", err, map[string]interface{}{
			"code": code,
		})
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch the user information from the provider.
	user, err := services.GetUserInfo(r.Context(), provider, token)
	if err != nil {
		slog.Error(ctx, "Failed to fetch user information", err, map[string]interface{}{
			"provider": provider,
		})
		http.Error(w, "Failed to fetch user information: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate a session ID and store the token
	sessionID := uuid.New().String()
	if err = services.StoreAuthToken(sessionID, provider, user, token); err != nil {
		slog.Error(ctx, "Failed to store token", err, map[string]interface{}{
			"session_id": sessionID,
			"provider":   provider,
			"user_id":    user.ID,
		})
		http.Error(w, "Failed to store token", http.StatusInternalServerError)
		return
	}

	slog.Info(ctx, "Successfully authenticated user", map[string]interface{}{
		"session_id":   sessionID,
		"provider":     provider,
		"user_id":      user.ID,
		"redirect_uri": redirectURI,
	})

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
	ctx := r.Context()
	slog.Info(ctx, "Starting auth provider logout", map[string]interface{}{
		"provider": provider,
		"user_id":  params.UserId,
	})

	if _, exists := config.Providers[provider]; !exists {
		slog.Error(ctx, "Unsupported provider", fmt.Errorf("provider not found"), map[string]interface{}{
			"provider": provider,
		})
		http.Error(w, "Unsupported provider", http.StatusBadRequest)
		return
	}

	// Retrieve the session ID from cookies.
	sessionCookie, err := r.Cookie("session_id")
	if err != nil || sessionCookie.Value == "" {
		slog.Error(ctx, "Session ID is required", fmt.Errorf("missing session cookie"), nil)
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}
	sessionID := sessionCookie.Value

	// If a specific user is specified, log out that user.
	if params.UserId != nil && *params.UserId != "" {
		if err := services.DeleteAuthToken(sessionID, provider, *params.UserId); err != nil {
			slog.Error(ctx, "Failed to log out user", err, map[string]interface{}{
				"session_id": sessionID,
				"provider":   provider,
				"user_id":    *params.UserId,
			})
			http.Error(w, "Failed to log out user", http.StatusInternalServerError)
			return
		}
		log.Printf("Logged out user %s from provider %s", *params.UserId, provider)
		slog.Info(ctx, "Successfully logged out user", map[string]interface{}{
			"session_id": sessionID,
			"provider":   provider,
			"user_id":    *params.UserId,
		})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": fmt.Sprintf("Successfully logged out user %s from provider %s", *params.UserId, provider),
		})
		return
	}

	// Otherwise, log out all users for the provider.
	if err := services.DeleteAllAuthTokensForProvider(sessionID, provider); err != nil {
		slog.Error(ctx, "Failed to log out all users", err, map[string]interface{}{
			"session_id": sessionID,
			"provider":   provider,
		})
		http.Error(w, "Failed to log out all users", http.StatusInternalServerError)
		return
	}
	log.Printf("Logged out all users from provider %s", provider)
	slog.Info(ctx, "Successfully logged out all users", map[string]interface{}{
		"session_id": sessionID,
		"provider":   provider,
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Successfully logged out all users from provider %s", provider),
	})
}
