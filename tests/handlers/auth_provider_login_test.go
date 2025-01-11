package handlers

import (
	"auth-service/tests"
	"auth-service/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"
)

func Test_WhenProviderIsInvalid_ShouldReturn404(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	// Arrange
	provider := "invalid-provider"
	baseURL := setup.Server.URL + "/auth/" + provider + "/login"

	reqURL, err := url.Parse(baseURL)
	assert.NoError(t, err)

	query := reqURL.Query()
	query.Set("redirect_uri", "http://localhost:3000/callback")
	reqURL.RawQuery = query.Encode()

	// Act
	resp, err := http.Get(reqURL.String())
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Assert
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	bodyBytes := make([]byte, resp.ContentLength)
	_, err = resp.Body.Read(bodyBytes)

	actualMessage := string(bodyBytes)
	expectedMessage := "Unsupported provider"
	assert.Contains(t, actualMessage, expectedMessage, "Unsupported provider")
}

func Test_WhenRedirectURIIsMissing_ShouldReturn400(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	provider := "spotify"
	baseURL := setup.Server.URL + "/auth/" + provider + "/login"

	resp, err := http.Get(baseURL)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	actualMessage := string(bodyBytes)
	expectedMessage := "Query argument redirect_uri is required, but not found"
	assert.Contains(t, actualMessage, expectedMessage)
}

func Test_WhenRedirectURIIsInvalid_ShouldReturn400(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	// Arrange: Set allowed domains to control the test environment
	os.Setenv("ALLOWED_REDIRECT_DOMAINS", "localhost,example.com")
	defer os.Unsetenv("ALLOWED_REDIRECT_DOMAINS") // Cleanup after test

	provider := "spotify"
	baseURL := setup.Server.URL + "/auth/" + provider + "/login"

	// Invalid domain
	reqURL, err := url.Parse(baseURL)
	assert.NoError(t, err)

	query := reqURL.Query()
	query.Set("redirect_uri", "http://malicious.com/callback") // Not in allowed domains
	reqURL.RawQuery = query.Encode()

	// Act
	resp, err := http.Get(reqURL.String())
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Assert
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	actualMessage := string(bodyBytes)
	expectedMessage := "Invalid redirect URI"
	assert.Contains(t, actualMessage, expectedMessage)
}
func Test_WhenOAuthLoginIsTriggered_ShouldRedirectToProvider(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	os.Setenv("ALLOWED_REDIRECT_DOMAINS", "localhost,example.com")
	defer os.Unsetenv("ALLOWED_REDIRECT_DOMAINS")

	provider := "spotify"
	baseURL := setup.Server.URL + "/auth/" + provider + "/login"

	// Mock OAuth Config
	mockOAuthConfig := new(mocks.MockOAuthConfig)

	// Handle any state token
	mockOAuthConfig.
		On("AuthCodeURL", mock.Anything).
		Return("https://accounts.spotify.com/authorize?state=mocked-state-token")

	// Build request
	reqURL, err := url.Parse(baseURL)
	assert.NoError(t, err)

	query := reqURL.Query()
	query.Set("redirect_uri", "http://localhost:3000/callback")
	reqURL.RawQuery = query.Encode()

	// Act
	resp, err := http.Get(reqURL.String())
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Assert: Check for Redirect (302)
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)

	// Assert: Check if the Location header is correct
	location := resp.Header.Get("Location")
	assert.Contains(t, location, "https://accounts.spotify.com/authorize")

	// Assert: Verify AuthCodeURL was called
	mockOAuthConfig.AssertCalled(t, "AuthCodeURL", mock.Anything)
}
