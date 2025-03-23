package auth_handler

import (
	"auth-service/services"
	"auth-service/tests"
	"auth-service/tests/mocks"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_GetAuthStatus_ShouldReturnConnectedProviders(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()
	sessionID := uuid.New().String()

	mockSpotifyToken1 := mocks.NewMockOAuth2Token("spotify-1", time.Hour)
	mockSpotifyToken2 := mocks.NewMockOAuth2Token("spotify-2", time.Hour)
	mockTidalToken := mocks.NewMockOAuth2Token("tidal", time.Hour)

	mockSpotifyUser1 := mocks.NewMockUser("spotify", "mock-spotify-user1-id", "Spotify User One", "spotify1@example.com")
	mockSpotifyUser2 := mocks.NewMockUser("spotify", "mock-spotify-user2-id", "Spotify User Two", "spotify2@example.com")
	mockTidalUser := mocks.NewMockUser("tidal", "mock-tidal-user-id", "Tidal User", "tidal@example.com")

	// Store tokens in Redis.
	err := services.StoreAuthToken(sessionID, "spotify", mockSpotifyUser1, mockSpotifyToken1)
	assert.NoError(t, err)
	err = services.StoreAuthToken(sessionID, "spotify", mockSpotifyUser2, mockSpotifyToken2)
	assert.NoError(t, err)
	err = services.StoreAuthToken(sessionID, "tidal", mockTidalUser, mockTidalToken)
	assert.NoError(t, err)

	// Create an HTTP client with a cookie jar.
	jar, err := cookiejar.New(nil)
	assert.NoError(t, err)
	client := &http.Client{Jar: jar}

	// Prepare the request with the session cookie.
	baseURL := setup.Server.URL + "/auth/status"
	req, err := http.NewRequest("GET", baseURL, nil)
	assert.NoError(t, err)
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: sessionID,
	})

	// Perform the request.
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Validate the status code.
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Parse the response.
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	var response []services.LoggedInProvider
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)

	// Expect three logged-in providers.
	assert.Len(t, response, 3)

	// Validate that both Spotify accounts and the Tidal account are present.
	var foundSpotify1, foundSpotify2, foundTidal bool
	for _, provider := range response {
		if provider.Provider == "spotify" && provider.UserID == mockSpotifyUser1.ID {
			foundSpotify1 = true
			assert.Equal(t, mockSpotifyUser1.DisplayName, provider.DisplayName)
			assert.Equal(t, mockSpotifyUser1.Email, provider.Email)
			assert.True(t, provider.LoggedIn)
		}
		if provider.Provider == "spotify" && provider.UserID == mockSpotifyUser2.ID {
			foundSpotify2 = true
			assert.Equal(t, mockSpotifyUser2.DisplayName, provider.DisplayName)
			assert.Equal(t, mockSpotifyUser2.Email, provider.Email)
			assert.True(t, provider.LoggedIn)
		}
		if provider.Provider == "tidal" && provider.UserID == mockTidalUser.ID {
			foundTidal = true
			assert.Equal(t, mockTidalUser.DisplayName, provider.DisplayName)
			assert.Equal(t, mockTidalUser.Email, provider.Email)
			assert.True(t, provider.LoggedIn)
		}
	}
	assert.True(t, foundSpotify1, "Spotify User One should be in the response")
	assert.True(t, foundSpotify2, "Spotify User Two should be in the response")
	assert.True(t, foundTidal, "Tidal User should be in the response")
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
