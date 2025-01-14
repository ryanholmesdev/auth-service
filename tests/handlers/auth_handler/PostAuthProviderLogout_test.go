package auth_handler

import (
	"auth-service/services"
	"auth-service/tests"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

// Helper to create a mock OAuth token
func createMockOAuthToken() *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
		Expiry:       time.Now().Add(time.Hour),
	}
}

// Test invalid provider handling
func Test_PostAuthProviderLogout_InvalidProvider_ShouldReturn400(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	url := setup.Server.URL + "/auth/invalid-provider/logout"
	req, err := http.NewRequest("POST", url, nil)
	assert.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(body), "Unsupported provider")
}

// Test missing session cookie
func Test_PostAuthProviderLogout_MissingSessionID_ShouldReturn400(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	url := setup.Server.URL + "/auth/spotify/logout"
	req, err := http.NewRequest("POST", url, nil)
	assert.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(body), "Session ID is required")
}

// Test successful logout
func Test_PostAuthProviderLogout_Success_ShouldReturn200(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	sessionID := "test-session-id"
	provider := "spotify"

	// Store a mock token in Redis
	err := services.StoreAuthToken(sessionID, provider, createMockOAuthToken())
	assert.NoError(t, err)

	url := setup.Server.URL + "/auth/spotify/logout"
	req, err := http.NewRequest("POST", url, nil)
	assert.NoError(t, err)

	// Attach session cookie
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}
	cookie := &http.Cookie{
		Name:  "session_id",
		Value: sessionID,
		Path:  "/",
	}
	jar.SetCookies(req.URL, []*http.Cookie{cookie})

	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(body), "Logged out successfully")
}
