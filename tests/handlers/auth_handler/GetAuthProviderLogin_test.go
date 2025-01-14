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

func buildRequestURL(baseURL, redirectURI string) (*url.URL, error) {
	reqURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	query := reqURL.Query()
	query.Set("redirect_uri", redirectURI)
	reqURL.RawQuery = query.Encode()
	return reqURL, nil
}

func Test_WhenProviderIsInvalid_ShouldReturn404(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	// Arrange
	provider := "invalid-provider"
	baseURL := setup.Server.URL + "/auth/" + provider + "/login"

	reqURL, err := buildRequestURL(baseURL, "http://localhost:3000/callback")
	assert.NoError(t, err)

	// Act
	resp, err := http.Get(reqURL.String())
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Assert
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(bodyBytes), "Unsupported provider")
}

func Test_WhenRedirectURIIsMissing_ShouldReturn400(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	// Arrange
	provider := "spotify"
	baseURL := setup.Server.URL + "/auth/" + provider + "/login"

	// Act
	resp, err := http.Get(baseURL)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Assert
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(bodyBytes), "Query argument redirect_uri is required, but not found")
}

func Test_WhenRedirectURIIsInvalid_ShouldReturn400(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	// Arrange
	_ = os.Setenv("ALLOWED_REDIRECT_DOMAINS", "localhost,example.com")
	defer func() {
		_ = os.Unsetenv("ALLOWED_REDIRECT_DOMAINS")
	}()

	provider := "spotify"
	baseURL := setup.Server.URL + "/auth/" + provider + "/login"

	reqURL, err := buildRequestURL(baseURL, "http://malicious.com/callback")
	assert.NoError(t, err)

	// Act
	resp, err := http.Get(reqURL.String())
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Assert
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(bodyBytes), "Invalid redirect URI")
}

func Test_WhenStateTokenStorageFails_ShouldReturn500(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	// Mock StoreStateToken to simulate a failure
	originalStoreStateToken := services.StoreStateToken
	services.StoreStateToken = func(stateToken string) error {
		return fmt.Errorf("redis internal failed")
	}
	defer func() {
		// Restore original function
		services.StoreStateToken = originalStoreStateToken
	}()

	// Arrange
	provider := "spotify"
	baseURL := setup.Server.URL + "/auth/" + provider + "/login"

	reqURL, err := buildRequestURL(baseURL, "http://localhost:3000/callback")
	assert.NoError(t, err)

	// Act
	resp, err := http.Get(reqURL.String())
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Assert: Should return 500 due to Redis failure
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(bodyBytes), "Server error while storing state")
}

func Test_WhenOAuthLoginIsTriggered_ShouldRedirectToProvider(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	mockAuthURL := setup.Server.URL + "/mock-oauth/authorize"
	mockTokenURL := setup.Server.URL + "/mock-oauth/token"
	mockRedirectURI := setup.Server.URL + "/mock-callback"

	_ = os.Setenv("ALLOWED_REDIRECT_DOMAINS", "localhost,127.0.0.1")
	defer func() {
		_ = os.Unsetenv("ALLOWED_REDIRECT_DOMAINS")
	}()

	// Arrange
	originalConfig := config.Providers["spotify"]
	mockConfig := *originalConfig
	mockConfig.RedirectURL = mockRedirectURI
	mockConfig.Endpoint = oauth2.Endpoint{
		AuthURL:  mockAuthURL,
		TokenURL: mockTokenURL,
	}
	config.Providers["spotify"] = &mockConfig

	router := setup.Server.Config.Handler.(*chi.Mux)
	router.Get("/mock-oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, mockRedirectURI+"?code=mocked-auth-code&state="+r.URL.Query().Get("state"), http.StatusFound)
	})

	router.Get("/mock-callback", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Mock callback received"))
	})

	reqURL, err := buildRequestURL(setup.Server.URL+"/auth/spotify/login", mockRedirectURI)
	assert.NoError(t, err)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil // Follow all redirects
		},
	}

	// Act
	resp, err := client.Get(reqURL.String())
	assert.NoError(t, err)
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	// Assert
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(bodyBytes), "Mock callback received")
}
