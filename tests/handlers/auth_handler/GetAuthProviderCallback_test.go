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
	assert.Contains(t, string(bodyBytes), "invalid redirect URI")
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

	// Instead of storing an auth token, store the PKCE data using the state token.
	// Here, we use "mock-code-verifier" as the code verifier.
	err := services.StorePKCEData("mock-state", "mock-code-verifier")
	assert.NoError(t, err)

	// Build callback URL WITHOUT the "code" parameter.
	// The state parameter now is in the format "stateToken|redirectURI".
	reqURL, err := buildCallbackURL(baseURL, "", "mock-state|http://localhost:3000/callback")
	assert.NoError(t, err)

	// Act: Call the callback endpoint.
	resp, err := http.Get(reqURL.String())
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Assert: Should return 400 due to missing authorization code.
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(bodyBytes), "Authorization code not provided")
}

// Test: Successful Callback Flow
func Test_Callback_Spotify_SuccessfulFlow_ShouldRedirectAndStoreToken(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	mockAuthURL := setup.Server.URL + "/mock-oauth/authorize"
	mockTokenURL := setup.Server.URL + "/mock-oauth/token"
	mockRedirectURI := setup.Server.URL + "/mock-callback"
	mockUserInfoURL := setup.Server.URL + "/mock-oauth/me"

	os.Setenv("ALLOWED_REDIRECT_DOMAINS", "localhost,127.0.0.1")
	defer os.Unsetenv("ALLOWED_REDIRECT_DOMAINS")

	// Mock OAuth Config for Spotify.
	originalConfig := config.Providers["spotify"]
	mockConfig := *originalConfig
	mockConfig.RedirectURL = mockRedirectURI
	mockConfig.Endpoint = oauth2.Endpoint{
		AuthURL:  mockAuthURL,
		TokenURL: mockTokenURL,
	}
	config.Providers["spotify"] = &mockConfig

	// Mock GetProviderUserInfoURL.
	originalGetProviderUserInfoURL := config.GetProviderUserInfoURL
	config.GetProviderUserInfoURL = func(provider string) (string, error) {
		if provider == "spotify" {
			return mockUserInfoURL, nil
		}
		return "", fmt.Errorf("provider not supported")
	}
	defer func() { config.GetProviderUserInfoURL = originalGetProviderUserInfoURL }()

	// Instead of pre-storing an auth token, store PKCE data with the state token.
	stateToken := "mock-state"
	codeVerifier := "mock-code-verifier"
	err := services.StorePKCEData(stateToken, codeVerifier)
	assert.NoError(t, err)

	// The callback endpoint will call the provider’s /me endpoint to get user info.
	// We'll mock that endpoint to return a specific user.
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
	router.Get("/mock-oauth/me", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Return user info matching what the callback will later use.
		w.Write([]byte(`{
			"id": "mock-user-id",
			"display_name": "Mock User",
			"email": "mockuser@googlemail.com"
		}`))
	})

	// Build callback URL with valid state and code.
	// Notice the state parameter is now "mock-state|{redirectURI}".
	reqURL, err := buildCallbackURL(setup.Server.URL+"/auth/spotify/callback", "mock-auth-code", stateToken+"|"+mockRedirectURI)
	assert.NoError(t, err)

	// Use a custom HTTP client that prevents automatic redirects.
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Act: Call the callback endpoint.
	resp, err := client.Get(reqURL.String())
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Assert: The callback should redirect (307) to the original redirect URI.
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
	redirectLocation, err := resp.Location()
	assert.NoError(t, err)
	assert.Equal(t, mockRedirectURI, redirectLocation.String())

	// Validate that a session cookie was set.
	sessionCookie := resp.Cookies()
	assert.NotEmpty(t, sessionCookie)
	var cookie *http.Cookie
	for _, c := range sessionCookie {
		if c.Name == "session_id" {
			cookie = c
			break
		}
	}
	assert.NotNil(t, cookie)

	// Validate token storage in Redis.
	// The callback should have stored the token under the new session cookie value.
	token, found := services.GetAuthToken(cookie.Value, "spotify", "mock-user-id")
	assert.True(t, found)
	assert.Equal(t, "mocked-access-token", token.Token.AccessToken)
}

func Test_Callback_Tidal_SuccessfulFlow_ShouldRedirectAndStoreToken(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	// Define mock URLs.
	mockAuthURL := setup.Server.URL + "/mock-oauth/authorize"
	mockTokenURL := setup.Server.URL + "/mock-oauth/token"
	mockRedirectURI := setup.Server.URL + "/mock-callback"
	// Use a dedicated endpoint for tidal's user info.
	mockUserInfoURL := setup.Server.URL + "/mock-oauth/tidal-me"

	os.Setenv("ALLOWED_REDIRECT_DOMAINS", "localhost,127.0.0.1")
	defer os.Unsetenv("ALLOWED_REDIRECT_DOMAINS")

	// Mock OAuth Config for Tidal.
	originalConfig := config.Providers["tidal"]
	mockConfig := *originalConfig
	mockConfig.RedirectURL = mockRedirectURI
	mockConfig.Endpoint = oauth2.Endpoint{
		AuthURL:  mockAuthURL,
		TokenURL: mockTokenURL,
	}
	config.Providers["tidal"] = &mockConfig

	// Mock GetProviderUserInfoURL for tidal.
	originalGetProviderUserInfoURL := config.GetProviderUserInfoURL
	config.GetProviderUserInfoURL = func(provider string) (string, error) {
		if provider == "tidal" {
			return mockUserInfoURL, nil
		}
		return "", fmt.Errorf("provider not supported")
	}
	defer func() { config.GetProviderUserInfoURL = originalGetProviderUserInfoURL }()

	// Store PKCE data using a state token.
	stateToken := "mock-state"
	codeVerifier := "mock-code-verifier"
	err := services.StorePKCEData(stateToken, codeVerifier)
	assert.NoError(t, err)

	// Set up mock endpoints.
	router := setup.Server.Config.Handler.(*chi.Mux)
	// Mock token exchange endpoint.
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
	// Mock tidal user info endpoint, returning the tidal response format.
	router.Get("/mock-oauth/tidal-me", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"id": "mock-user-id",
				"attributes": {
					"username": "TidalUser",
					"email": "tidaluser@example.com",
					"emailVerified": true,
					"country": "US"
				}
			}
		}`))
	})

	// Build callback URL with valid state and auth code.
	reqURL, err := buildCallbackURL(setup.Server.URL+"/auth/tidal/callback", "mock-auth-code", stateToken+"|"+mockRedirectURI)
	assert.NoError(t, err)

	// Use a custom HTTP client that prevents following redirects.
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Act: Call the callback endpoint.
	resp, err := client.Get(reqURL.String())
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Assert: Verify a temporary redirect to the frontend.
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
	redirectLocation, err := resp.Location()
	assert.NoError(t, err)
	assert.Equal(t, mockRedirectURI, redirectLocation.String())

	// Verify that a session cookie was set.
	sessionCookie := resp.Cookies()
	assert.NotEmpty(t, sessionCookie)
	var cookie *http.Cookie
	for _, c := range sessionCookie {
		if c.Name == "session_id" {
			cookie = c
			break
		}
	}
	assert.NotNil(t, cookie)

	// Validate token storage in Redis for the tidal provider.
	token, found := services.GetAuthToken(cookie.Value, "tidal", "mock-user-id")
	assert.True(t, found)
	assert.Equal(t, "mocked-access-token", token.Token.AccessToken)
}
