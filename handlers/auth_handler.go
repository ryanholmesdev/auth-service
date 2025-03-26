package handlers

import (
	"auth-service/config"
	"auth-service/generated"
	"auth-service/services"
	"auth-service/utils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// GetAuthProviderLogin handles login requests
func (s *Server) GetAuthProviderLogin(w http.ResponseWriter, r *http.Request, provider string, params generated.GetAuthProviderLoginParams) {
	ctx := r.Context()

	oauthConfig, exists := config.Providers[provider]
	if !exists {
		utils.LogWarn(ctx, "Unsupported provider requested", map[string]interface{}{
			"provider": provider,
		})
		http.Error(w, "Unsupported provider", http.StatusNotFound)
		return
	}

	// Access the redirect_uri from params
	redirectURI := params.RedirectUri

	// Validate the redirect URI
	allowedDomains, err := utils.GetAllowedRedirectDomains()
	if !utils.ValidateRedirectURI(redirectURI, allowedDomains) {
		utils.LogWarn(ctx, "Invalid redirect URI", map[string]interface{}{
			"redirect_uri": redirectURI,
		})
		http.Error(w, "Invalid redirect URI", http.StatusBadRequest)
		return
	}

	// Generate the state parameter including the redirect URI
	stateToken := uuid.New().String() // Generate a unique CSRF token to this request
	state := stateToken + "|" + redirectURI

	// Store the state token in Redis with a TTL
	err = services.StoreStateToken(stateToken)
	if err != nil {
		utils.LogError(ctx, "Failed to store state token", err, map[string]interface{}{
			"provider": provider,
		})
		http.Error(w, "Server error while storing state", http.StatusInternalServerError)
		return
	}

	// Generate the authorization URL
	authURL := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)

	utils.LogInfo(ctx, "Redirecting to OAuth provider", map[string]interface{}{
		"provider": provider,
		"state":    stateToken,
	})

	// Redirect to the OAuth provider
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (s *Server) GetAuthProviderCallback(w http.ResponseWriter, r *http.Request, provider string, params generated.GetAuthProviderCallbackParams) {
	ctx := r.Context()

	oauthConfig, exists := config.Providers[provider]
	if !exists {
		utils.LogWarn(ctx, "Unsupported provider in callback", map[string]interface{}{
			"provider": provider,
		})
		http.Error(w, "Unsupported provider", http.StatusBadRequest)
		return
	}

	// Extract the state parameter and split into token and redirect URI
	parts := strings.SplitN(params.State, "|", 2)
	if len(parts) != 2 {
		utils.LogWarn(ctx, "Invalid state parameter format", map[string]interface{}{
			"state": params.State,
		})
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	stateToken := parts[0]
	redirectURI := parts[1]

	// Validate the redirect URI
	allowedDomains, err := utils.GetAllowedRedirectDomains()
	if !utils.ValidateRedirectURI(redirectURI, allowedDomains) {
		utils.LogWarn(ctx, "Invalid redirect URI in callback", map[string]interface{}{
			"redirect_uri": redirectURI,
		})
		http.Error(w, "Invalid redirect URI", http.StatusBadRequest)
		return
	}

	// Validate the state token
	isValid := services.ValidateStateToken(stateToken)
	if !isValid {
		utils.LogWarn(ctx, "Invalid or expired state token", map[string]interface{}{
			"state_token": stateToken,
		})
		http.Error(w, "Invalid or expired state token", http.StatusBadRequest)
		return
	}

	// Delete the state token to prevent reuse
	go func(stateToken string) {
		err := services.DeleteStateToken(stateToken)
		if err != nil {
			utils.LogError(context.Background(), "Failed to delete state token", err, map[string]interface{}{
				"state_token": stateToken,
			})
		}
	}(stateToken)

	code := r.URL.Query().Get("code")
	if code == "" {
		utils.LogWarn(ctx, "Authorization code not provided", map[string]interface{}{
			"provider": provider,
		})
		http.Error(w, "Authorization code not provided", http.StatusBadRequest)
		return
	}

	// Exchange the authorization code for a token
	token, err := oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		utils.LogError(ctx, "Failed to exchange token", err, map[string]interface{}{
			"provider": provider,
		})
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch the user information from the provider
	user, err := services.GetUserInfo(provider, token)
	if err != nil {
		utils.LogError(ctx, "Failed to fetch user information", err, map[string]interface{}{
			"provider": provider,
		})
		http.Error(w, "Failed to fetch user information: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate a session ID and store the token
	sessionID := uuid.New().String()
	err = services.StoreAuthToken(sessionID, provider, user, token)
	if err != nil {
		utils.LogError(ctx, "Failed to store auth token", err, map[string]interface{}{
			"provider": provider,
			"user_id":  user.ID,
		})
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

	utils.LogInfo(ctx, "Successfully authenticated user", map[string]interface{}{
		"provider": provider,
		"user_id":  user.ID,
	})

	// Redirect to the original URI
	http.Redirect(w, r, redirectURI, http.StatusTemporaryRedirect)
}

func (s *Server) PostAuthProviderLogout(w http.ResponseWriter, r *http.Request, provider string, params generated.PostAuthProviderLogoutParams) {
	ctx := r.Context()

	_, exists := config.Providers[provider]
	if !exists {
		utils.LogWarn(ctx, "Unsupported provider in logout", map[string]interface{}{
			"provider": provider,
		})
		http.Error(w, "Unsupported provider", http.StatusBadRequest)
		return
	}

	// Retrieve session ID from cookies
	sessionCookie, err := r.Cookie("session_id")
	if err != nil || sessionCookie.Value == "" {
		utils.LogWarn(ctx, "Session ID missing in logout request", map[string]interface{}{
			"provider": provider,
		})
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}
	sessionID := sessionCookie.Value

	// Determine logout mode: Single user or all users for the provider
	if params.UserId != nil && *params.UserId != "" {
		// logging out a specific user
		err := services.DeleteAuthToken(sessionID, provider, *params.UserId)
		if err != nil {
			utils.LogError(ctx, "Failed to log out user", err, map[string]interface{}{
				"provider": provider,
				"user_id":  *params.UserId,
			})
			http.Error(w, "Failed to log out user", http.StatusInternalServerError)
			return
		}

		utils.LogInfo(ctx, "Logged out user", map[string]interface{}{
			"provider": provider,
			"user_id":  *params.UserId,
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": fmt.Sprintf("Successfully logged out user %s from provider %s", *params.UserId, provider),
		})

		return
	}

	// logging out all user accounts under a given provider
	err = services.DeleteAllAuthTokensForProvider(sessionID, provider)
	if err != nil {
		utils.LogError(ctx, "Failed to log out all users", err, map[string]interface{}{
			"provider": provider,
		})
		http.Error(w, "Failed to log out all users", http.StatusInternalServerError)
		return
	}

	utils.LogInfo(ctx, "Logged out all users from provider", map[string]interface{}{
		"provider": provider,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Successfully logged out all users from provider %s", provider),
	})
}
