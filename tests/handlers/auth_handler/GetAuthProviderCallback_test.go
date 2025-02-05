package auth_handler

import (
	"auth-service/config"
	"auth-service/services"
	"auth-service/tests"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"
)

// Helper to build callback URL with query params
func buildCallbackURL(baseURL, code, state string) (*url.URL, error) {
	reqURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	query := reqURL.Query()
	query.Set("code", code)
	query.Set("state", state)
	reqURL.RawQuery = query.Encode()
	return reqURL, nil
}

// Test: Invalid Provider
func Test_Callback_InvalidProvider_ShouldReturn400(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	provider := "invalid-provider"
	baseURL := setup.Server.URL + "/auth/" + provider + "/callback"

	reqURL, err := buildCallbackURL(baseURL, "mock-code", "mock-state|http://localhost:3000/callback")
	assert.NoError(t, err)

	resp, err := http.Get(reqURL.String())
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(bodyBytes), "Unsupported provider")
}

// Test: Invalid State Format
func Test_Callback_InvalidStateFormat_ShouldReturn400(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	provider := "spotify"
	baseURL := setup.Server.URL + "/auth/" + provider + "/callback"

	reqURL, err := buildCallbackURL(baseURL, "mock-code", "invalid-state-format")
	assert.NoError(t, err)

	resp, err := http.Get(reqURL.String())
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(bodyBytes), "Invalid state parameter")
}

// Test: Invalid Redirect URI
func Test_Callback_InvalidRedirectURI_ShouldReturn400(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	os.Setenv("ALLOWED_REDIRECT_DOMAINS", "localhost,example.com")
	defer os.Unsetenv("ALLOWED_REDIRECT_DOMAINS")

	provider := "spotify"
	baseURL := setup.Server.URL + "/auth/" + provider + "/callback"

	reqURL, err := buildCallbackURL(baseURL, "mock-code", "mock-state|http://malicious.com/callback")
	assert.NoError(t, err)

	resp, err := http.Get(reqURL.String())
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(bodyBytes), "Invalid redirect URI")
}

// Test: Invalid State Token
func Test_Callback_InvalidStateToken_ShouldReturn400(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	provider := "spotify"
	baseURL := setup.Server.URL + "/auth/" + provider + "/callback"

	reqURL, err := buildCallbackURL(baseURL, "mock-code", "invalid-state|http://localhost:3000/callback")
	assert.NoError(t, err)

	resp, err := http.Get(reqURL.String())
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(bodyBytes), "Invalid or expired state token")
}

// Test: Missing Authorization Code
func Test_Callback_MissingAuthorizationCode_ShouldReturn400(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	provider := "spotify"
	baseURL := setup.Server.URL + "/auth/" + provider + "/callback"

	// Store a valid state token in Redis
	err := services.StoreStateToken("mock-state")
	assert.NoError(t, err)

	// Build callback URL WITHOUT the "code" parameter
	reqURL, err := buildCallbackURL(baseURL, "", "mock-state|http://localhost:3000/callback")
	assert.NoError(t, err)

	// Act: Call the callback endpoint
	resp, err := http.Get(reqURL.String())
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Assert: Should return 400 due to missing authorization code
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(bodyBytes), "Authorization code not provided")
}

// Test: Successful Callback Flow
func Test_Callback_SuccessfulFlow_ShouldRedirectAndStoreToken(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	mockAuthURL := setup.Server.URL + "/mock-oauth/authorize"
	mockTokenURL := setup.Server.URL + "/mock-oauth/token"
	mockRedirectURI := setup.Server.URL + "/mock-callback"
	mockUserInfoURL := setup.Server.URL + "/mock-oauth/me"

	os.Setenv("ALLOWED_REDIRECT_DOMAINS", "localhost,127.0.0.1")
	defer os.Unsetenv("ALLOWED_REDIRECT_DOMAINS")

	// Mock OAuth Config
	originalConfig := config.Providers["spotify"]
	mockConfig := *originalConfig
	mockConfig.RedirectURL = mockRedirectURI
	mockConfig.Endpoint = oauth2.Endpoint{
		AuthURL:  mockAuthURL,
		TokenURL: mockTokenURL,
	}
	config.Providers["spotify"] = &mockConfig

	// Mock GetProviderUserInfoURL
	originalGetProviderUserInfoURL := config.GetProviderUserInfoURL
	config.GetProviderUserInfoURL = func(provider string) (string, error) {
		if provider == "spotify" {
			return mockUserInfoURL, nil
		}
		return "", fmt.Errorf("provider not supported")
	}
	defer func() { config.GetProviderUserInfoURL = originalGetProviderUserInfoURL }()

	// Store a valid state token in Redis
	err := services.StoreStateToken("mock-state")
	assert.NoError(t, err)

	// Mock token exchange response
	router := setup.Server.Config.Handler.(*chi.Mux)
	router.Post("/mock-oauth/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"access_token": "mocked-access-token",
			"refresh_token": "mocked-refresh-token",
			"expires_in": 3600,
			"token_type": "Bearer"
		}`))
	})

	// Mock Spotify `/me` API response to return a user ID
	router.Get("/mock-oauth/me", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "mock-user-id",
			"display_name": "Mock User",
			"email": "mockuser@googlemail.com",
		}`))
	})

	// Build callback URL with valid state and code
	reqURL, err := buildCallbackURL(setup.Server.URL+"/auth/spotify/callback", "mock-auth-code", "mock-state|"+mockRedirectURI)
	assert.NoError(t, err)

	// Custom HTTP client to prevent automatic redirects
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Prevents following redirects
		},
	}

	// Act: Call the callback endpoint
	resp, err := client.Get(reqURL.String())
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Assert: Should redirect (307) to the frontend
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)

	// Validate the redirect URL
	redirectLocation, err := resp.Location()
	assert.NoError(t, err)
	assert.Equal(t, mockRedirectURI, redirectLocation.String())

	// Validate session cookie
	sessionCookie := resp.Cookies()
	assert.NotEmpty(t, sessionCookie)
	assert.Equal(t, "session_id", sessionCookie[0].Name)

	// Validate token storage in Redis with user ID
	token, found := services.GetAuthToken(sessionCookie[0].Value, "spotify", "mock-user-id")
	assert.True(t, found)
	assert.Equal(t, "mocked-access-token", token.Token.AccessToken)
}
