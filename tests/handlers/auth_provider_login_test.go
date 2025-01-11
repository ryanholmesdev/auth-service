package handlers

import (
	"auth-service/tests"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func Test_WhenProviderIsValid_ShouldRedirectToOAuthProvider(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	resp, err := http.Get(setup.Server.URL + "/auth/spotify/login?redirect_uri=http://localhost:3000/callback")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
	location := resp.Header.Get("Location")
	assert.Contains(t, location, "https://accounts.spotify.com/authorize")
	assert.Contains(t, location, "state=")
}

func Test_WhenProviderIsInvalid_ShouldReturn404(t *testing.T) {
	setup := tests.InitializeTestEnvironment(t)
	defer setup.Cleanup()

	// invalid provider request
	resp, err := http.Get(setup.Server.URL + "/auth/invalid-provider/login?redirect_uri=http://localhost:3000/callback")
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Assert
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
