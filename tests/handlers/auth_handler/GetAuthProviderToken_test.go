package auth_handler

/*
func Test_GetAuthProviderToken_InvalidProvider_ShouldReturn400(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	url := setup.Server.URL + "/auth/invalid-provider/token"
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

	url := setup.Server.URL + "/auth/spotify/token"
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

func Test_GetAuthProviderToken_TokenNotFound_ShouldReturn404(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	url := setup.Server.URL + "/auth/spotify/token"
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

func Test_GetAuthProviderToken_ExpiredToken_FailedRefresh_ShouldReturn500(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	// Arrange: Store an expired token
	expiredToken := &oauth2.Token{
		AccessToken:  "expired-access-token",
		RefreshToken: "expired-refresh-token",
		Expiry:       time.Now().Add(-1 * time.Hour),
	}

	err := services.StoreAuthToken("mock-session-id", "spotify", expiredToken)
	assert.NoError(t, err)

	// Mock RefreshAccessTokenFunc to simulate a refresh failure
	utils.RefreshAccessTokenFunc = func(config *oauth2.Config, refreshToken string) (*oauth2.Token, error) {
		return nil, fmt.Errorf("failed to refresh token")
	}
	defer func() { utils.RefreshAccessTokenFunc = utils.RefreshAccessToken }() // Reset after test

	// Act: Request the token with an expired session
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

	// Assert: Should return 500 due to failed refresh
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(body), "Failed to refresh token")
}

func Test_GetAuthProviderToken_ExpiredToken_SuccessfulRefresh_ShouldReturn200(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	expiredToken := &oauth2.Token{
		AccessToken:  "expired-access-token",
		RefreshToken: "valid-refresh-token",
		Expiry:       time.Now().Add(-1 * time.Hour),
	}

	err := services.StoreAuthToken("mock-session-id", "spotify", expiredToken)
	assert.NoError(t, err)

	// Mock the RefreshAccessToken function
	utils.RefreshAccessTokenFunc = func(config *oauth2.Config, refreshToken string) (*oauth2.Token, error) {
		return &oauth2.Token{
			AccessToken:  "refreshed-access-token",
			RefreshToken: "valid-refresh-token",
			Expiry:       time.Now().Add(1 * time.Hour),
		}, nil
	}
	defer func() { utils.RefreshAccessTokenFunc = utils.RefreshAccessToken }() // Reset after test

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

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	assert.NoError(t, err)

	assert.Equal(t, "refreshed-access-token", response["access_token"])
}

func Test_GetAuthProviderToken_ValidToken_ShouldReturn200(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	validToken := &oauth2.Token{
		AccessToken:  "valid-access-token",
		RefreshToken: "valid-refresh-token",
		Expiry:       time.Now().Add(1 * time.Hour),
	}

	err := services.StoreAuthToken("mock-session-id", "spotify", validToken)
	assert.NoError(t, err)

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

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	assert.NoError(t, err)

	assert.Equal(t, "valid-access-token", response["access_token"])
}
*/
