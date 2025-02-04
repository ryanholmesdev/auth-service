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

	"github.com/stretchr/testify/assert"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

func Test_GetAuthStatus_ShouldReturnConnectedProviders(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	// Mock a session and token in Redis
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

	// Store spotify token in redis
	err := services.StoreAuthToken(sessionID, "spotify", "mock-user1-id", mockToken1)
	assert.NoError(t, err)

	err = services.StoreAuthToken(sessionID, "spotify", "mock-user2-id", mockToken2)
	assert.NoError(t, err)

	// Create a client with a cookie jar to store session cookies
	jar, err := cookiejar.New(nil)
	assert.NoError(t, err)

	client := &http.Client{
		Jar: jar,
	}

	// Add the session cookie to the request
	baseURL := setup.Server.URL + "/auth/status"
	req, err := http.NewRequest("GET", baseURL, nil)
	assert.NoError(t, err)

	// Attach the session cookie
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: sessionID,
	})

	// Perform the request
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Result
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	// Parse the response
	var response []services.LoggedInProvider
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)

	assert.Len(t, response, 2, "Should return 2 logged-in Spotify accounts")

	// Validate both accounts are listed and tokens are correct
	assert.Contains(t, response, services.LoggedInProvider{
		Provider: "spotify",
		User:     "mock-user1-id",
		LoggedIn: true,
	})
	assert.Contains(t, response, services.LoggedInProvider{
		Provider: "spotify",
		User:     "mock-user2-id",
		LoggedIn: true,
	})
}

func Test_GetAuthStatus_ShouldReturnNonConnectedProviders(t *testing.T) {
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

	// Result
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
