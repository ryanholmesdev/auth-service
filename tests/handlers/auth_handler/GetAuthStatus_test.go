package auth_handler

import (
	"auth-service/services"
	"auth-service/tests"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

func Test_GetAuthStatus_ShouldReturnConnectedProviders(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	// Mock a session and token in Redis
	sessionID := uuid.New().String()
	mockToken := &oauth2.Token{
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
		Expiry:       time.Now().Add(time.Hour),
	}

	// Store spotify token in redis
	err := services.StoreAuthToken(sessionID, "spotify", mockToken)
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
	var response map[string]bool
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)

	// Assert that Spotify is connected and Tidal is not
	assert.True(t, response["spotify"], "Spotify should be connected")
	assert.False(t, response["tidal"], "Tidal should not be connected")
}
