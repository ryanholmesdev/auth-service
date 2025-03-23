package auth_handler

import (
	"auth-service/services"
	"auth-service/tests"
	"auth-service/tests/mocks"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"
)

// createSessionRequest prepares an HTTP request with the session cookie attached.
func createSessionRequest(t *testing.T, method, url, sessionID string) *http.Request {
	req, err := http.NewRequest(method, url, nil)
	assert.NoError(t, err)
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: sessionID,
		Path:  "/",
	})
	return req
}

// Test_PostAuthProviderLogout_InvalidProvider_ShouldReturn400 remains the same.
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

// Test_PostAuthProviderLogout_MissingSessionID_ShouldReturn400 remains unchanged.
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

// Test_PostAuthProviderLogout_SingleUser_ShouldReturn200 refactored.
func Test_PostAuthProviderLogout_SingleUser_ShouldReturn200(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	sessionID := "test-session-id"
	provider := "spotify"
	userID := "mock-user-1"

	// Use mocks.NewMockUser to create mock user info.
	mockUser1 := mocks.NewMockUser("spotify", userID, "User One", "user1@example.com")
	mockUser2 := mocks.NewMockUser("spotify", "mock-user-2", "User Two", "user2@example.com")

	// Use mocks.NewMockOAuth2Token to create tokens.
	token := mocks.NewMockOAuth2Token("spotify", time.Hour)

	// Store two mock tokens in Redis.
	err := services.StoreAuthToken(sessionID, provider, mockUser1, token)
	assert.NoError(t, err)
	err = services.StoreAuthToken(sessionID, provider, mockUser2, token)
	assert.NoError(t, err)

	// Verify token exists before logout.
	_, found := services.GetAuthToken(sessionID, provider, userID)
	assert.True(t, found, "Token should exist before logout")

	// Use helper to create a request with the session cookie.
	baseURL := setup.Server.URL + "/auth/spotify/logout?user_id=" + userID
	jar, err := cookiejar.New(nil)
	assert.NoError(t, err)
	client := &http.Client{Jar: jar}
	req := createSessionRequest(t, "POST", baseURL, sessionID)

	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(body), "Successfully logged out user "+userID)

	// Verify that only the specified user's token was deleted.
	_, found = services.GetAuthToken(sessionID, provider, userID)
	assert.False(t, found, "Token for logged-out user should be removed")
	_, stillExists := services.GetAuthToken(sessionID, provider, "mock-user-2")
	assert.True(t, stillExists, "Other user's token should still exist")
}

// Test_PostAuthProviderLogout_AllUsers_ShouldReturn200 refactored.
func Test_PostAuthProviderLogout_AllUsers_ShouldReturn200(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	sessionID := "test-session-id"
	provider := "spotify"

	// Create mock user info using mocks.
	mockUser1 := mocks.NewMockUser("spotify", "mock-user-1", "User One", "user1@example.com")
	mockUser2 := mocks.NewMockUser("spotify", "mock-user-2", "User Two", "user2@example.com")

	// Create and store tokens using mocks.
	token := mocks.NewMockOAuth2Token("spotify", time.Hour)
	err := services.StoreAuthToken(sessionID, provider, mockUser1, token)
	assert.NoError(t, err)
	err = services.StoreAuthToken(sessionID, provider, mockUser2, token)
	assert.NoError(t, err)

	// Verify tokens exist before logout.
	_, found1 := services.GetAuthToken(sessionID, provider, "mock-user-1")
	_, found2 := services.GetAuthToken(sessionID, provider, "mock-user-2")
	assert.True(t, found1, "User 1 token should exist before logout")
	assert.True(t, found2, "User 2 token should exist before logout")

	baseURL := setup.Server.URL + "/auth/spotify/logout"
	jar, err := cookiejar.New(nil)
	assert.NoError(t, err)
	client := &http.Client{Jar: jar}
	req := createSessionRequest(t, "POST", baseURL, sessionID)

	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(body), "Successfully logged out all users from provider spotify")

	// Verify all tokens were deleted.
	_, found1 = services.GetAuthToken(sessionID, provider, "mock-user-1")
	_, found2 = services.GetAuthToken(sessionID, provider, "mock-user-2")
	assert.False(t, found1, "User 1 token should be removed")
	assert.False(t, found2, "User 2 token should be removed")
}
