package auth_handler

import (
	"auth-service/services"
	"auth-service/tests"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"testing"
	"time"
)

func Test_GetAuthProviderToken_InvalidProvider_ShouldReturn400(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	url := setup.Server.URL + "/auth/invalid-provider/token?user_id=mock-user-id"
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(body), "Unsupported provider")
}

func Test_GetAuthProviderToken_MissingSessionCookie_ShouldReturn400(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	url := setup.Server.URL + "/auth/spotify/token?user_id=mock-user-id"
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(body), "Session ID is required")
}

func Test_GetAuthProviderToken_MissingUserID_ShouldReturn400(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	url := setup.Server.URL + "/auth/spotify/token"
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)

	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: "mock-session-id",
	})

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(body), "Query parameter user_id is required")
}

func Test_GetAuthProviderToken_TokenNotFound_ShouldReturn404(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	url := setup.Server.URL + "/auth/spotify/token?user_id=mock-user-id"
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)

	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: "invalid-session",
	})

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(body), "Token not found")
}

func Test_GetAuthProviderToken_ValidToken_ShouldReturn200(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	// Create mock token
	validToken := &oauth2.Token{
		AccessToken:  "valid-access-token",
		RefreshToken: "valid-refresh-token",
		Expiry:       time.Now().Add(1 * time.Hour),
	}

	// Create user info
	mockUser := &services.UserInfo{
		ID:          "mock-user-id",
		DisplayName: "John Doe",
		Email:       "john@example.com",
	}

	// Store both token and user info in Redis
	err := services.StoreAuthToken("mock-session-id", "spotify", mockUser, validToken)
	assert.NoError(t, err)

	url := setup.Server.URL + "/auth/spotify/token?user_id=mock-user-id"
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)

	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: "mock-session-id",
	})

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	assert.NoError(t, err)

	assert.Equal(t, "valid-access-token", response["access_token"])
}
