package auth_handler

import (
	"auth-service/services"
	"auth-service/tests"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func Test_GetAuthStatus_ShouldReturnConnectedProviders(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	// Mock session and tokens
	sessionID := uuid.New().String()
	mockToken1 := &oauth2.Token{
		AccessToken:  "mock-access-token-1",
		RefreshToken: "mock-refresh-token-1",
		Expiry:       time.Now().Add(time.Hour),
	}
	mockToken2 := &oauth2.Token{
		AccessToken:  "mock-access-token-2",
		RefreshToken: "mock-refresh-token-2",
		Expiry:       time.Now().Add(time.Hour),
	}

	// Create user info for mock users
	mockUser1 := &services.UserInfo{
		ID:          "mock-user1-id",
		DisplayName: "User One",
		Email:       "user1@example.com",
	}
	mockUser2 := &services.UserInfo{
		ID:          "mock-user2-id",
		DisplayName: "User Two",
		Email:       "user2@example.com",
	}

	// Store tokens in Redis
	err := services.StoreAuthToken(sessionID, "spotify", mockUser1, mockToken1)
	assert.NoError(t, err)
	err = services.StoreAuthToken(sessionID, "spotify", mockUser2, mockToken2)
	assert.NoError(t, err)

	// Create an HTTP client with a cookie jar
	jar, err := cookiejar.New(nil)
	assert.NoError(t, err)

	client := &http.Client{
		Jar: jar,
	}

	// Add the session cookie to the request
	baseURL := setup.Server.URL + "/auth/status"
	req, err := http.NewRequest("GET", baseURL, nil)
	assert.NoError(t, err)
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: sessionID,
	})

	// Perform the request
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Validate status code
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Parse the response
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	var response []services.LoggedInProvider
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)

	// Should return 2 logged-in providers
	assert.Len(t, response, 2)

	// Validate both accounts are present
	var foundUser1, foundUser2 bool
	for _, provider := range response {
		if provider.Provider == "spotify" && provider.UserID == mockUser1.ID {
			foundUser1 = true
			assert.Equal(t, mockUser1.DisplayName, provider.DisplayName)
			assert.Equal(t, mockUser1.Email, provider.Email)
			assert.True(t, provider.LoggedIn)
		}
		if provider.Provider == "spotify" && provider.UserID == mockUser2.ID {
			foundUser2 = true
			assert.Equal(t, mockUser2.DisplayName, provider.DisplayName)
			assert.Equal(t, mockUser2.Email, provider.Email)
			assert.True(t, provider.LoggedIn)
		}
	}
	assert.True(t, foundUser1, "User 1 should be in the response")
	assert.True(t, foundUser2, "User 2 should be in the response")
}

func Test_GetAuthStatus_ShouldReturnUnauthorised(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	client := &http.Client{}

	baseURL := setup.Server.URL + "/auth/status"
	req, err := http.NewRequest("GET", baseURL, nil)
	assert.NoError(t, err)

	// Perform the request
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Unauthorized response
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
