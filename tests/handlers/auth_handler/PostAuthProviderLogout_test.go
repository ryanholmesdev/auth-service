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

// Test logging out a specific user from a provider
func Test_PostAuthProviderLogout_SingleUser_ShouldReturn200(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	sessionID := "test-session-id"
	provider := "spotify"
	userID := "mock-user-1"

	// Store two mock tokens in Redis
	err := services.StoreAuthToken(sessionID, provider, userID, createMockOAuthToken())
	assert.NoError(t, err)
	err = services.StoreAuthToken(sessionID, provider, "mock-user-2", createMockOAuthToken())
	assert.NoError(t, err)

	// Verify token exists before logout
	_, found := services.GetAuthToken(sessionID, provider, userID)
	assert.True(t, found, "Token should exist before logout")

	url := setup.Server.URL + "/auth/spotify/logout?user_id=" + userID
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
	assert.Contains(t, string(body), "Successfully logged out user")

	// Verify that only the specified user's token was deleted
	_, found = services.GetAuthToken(sessionID, provider, userID)
	assert.False(t, found, "Token for logged-out user should be removed")
	_, stillExists := services.GetAuthToken(sessionID, provider, "mock-user-2")
	assert.True(t, stillExists, "Other user's token should still exist")
}

// Test logging out all users from a provider
func Test_PostAuthProviderLogout_AllUsers_ShouldReturn200(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	sessionID := "test-session-id"
	provider := "spotify"

	// Store mock tokens for multiple users
	err := services.StoreAuthToken(sessionID, provider, "mock-user-1", createMockOAuthToken())
	assert.NoError(t, err)
	err = services.StoreAuthToken(sessionID, provider, "mock-user-2", createMockOAuthToken())
	assert.NoError(t, err)

	// Verify tokens exist before logout
	_, found1 := services.GetAuthToken(sessionID, provider, "mock-user-1")
	_, found2 := services.GetAuthToken(sessionID, provider, "mock-user-2")
	assert.True(t, found1, "User 1 token should exist before logout")
	assert.True(t, found2, "User 2 token should exist before logout")

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
	assert.Contains(t, string(body), "Successfully logged out all users")

	// Verify all tokens were deleted
	_, found1 = services.GetAuthToken(sessionID, provider, "mock-user-1")
	_, found2 = services.GetAuthToken(sessionID, provider, "mock-user-2")
	assert.False(t, found1, "User 1 token should be removed")
	assert.False(t, found2, "User 2 token should be removed")
}
