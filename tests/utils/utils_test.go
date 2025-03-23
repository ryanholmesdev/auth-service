package utils

import (
	"auth-service/utils"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_ValidateRedirectURIFromEnv_ValidURI(t *testing.T) {
	os.Setenv("ALLOWED_REDIRECT_DOMAINS", "example.com,localhost")
	defer os.Unsetenv("ALLOWED_REDIRECT_DOMAINS")

	err := utils.ValidateRedirectURIFromEnv("http://example.com/callback")
	assert.NoError(t, err)
}

func Test_ValidateRedirectURIFromEnv_InvalidURI(t *testing.T) {
	os.Setenv("ALLOWED_REDIRECT_DOMAINS", "example.com,localhost")
	defer os.Unsetenv("ALLOWED_REDIRECT_DOMAINS")

	err := utils.ValidateRedirectURIFromEnv("http://evil.com/callback")
	assert.Error(t, err)
}

func Test_ValidateRedirectURI_ShouldPassForSubdomains(t *testing.T) {
	allowedDomains := []string{"example.com"}

	result := utils.ValidateRedirectURI("http://sub.example.com/callback", allowedDomains)
	assert.True(t, result)
}

func Test_ValidateRedirectURI_ShouldFailForSimilarLookingDomain(t *testing.T) {
	allowedDomains := []string{"example.com"}

	result := utils.ValidateRedirectURI("http://example.com.evil.com/callback", allowedDomains)
	assert.False(t, result)
}

func Test_ValidateRedirectURI_ShouldFailForIPAddress(t *testing.T) {
	allowedDomains := []string{"127.0.0.1", "localhost"}

	result := utils.ValidateRedirectURI("http://127.0.0.1:8080/callback", allowedDomains)
	assert.True(t, result)

	result = utils.ValidateRedirectURI("http://192.168.1.1/callback", allowedDomains)
	assert.False(t, result)
}

func Test_ValidateRedirectURI_ShouldFailIfAllowedDomainsEmpty(t *testing.T) {
	allowedDomains := []string{}

	result := utils.ValidateRedirectURI("http://example.com/callback", allowedDomains)
	assert.False(t, result)
}

func Test_ValidateRedirectURI_ShouldPassForMixedCaseDomains(t *testing.T) {
	allowedDomains := []string{"Example.COM"}

	result := utils.ValidateRedirectURI("http://example.com/callback", allowedDomains)
	assert.True(t, result)

	result = utils.ValidateRedirectURI("http://EXAMPLE.com/callback", allowedDomains)
	assert.True(t, result)
}

func Test_ValidateRedirectURI_ShouldFailForMalformedURL(t *testing.T) {
	allowedDomains := []string{"example.com"}

	// Malformed URL
	result := utils.ValidateRedirectURI("http://", allowedDomains)
	assert.False(t, result)

	result = utils.ValidateRedirectURI("://missing-scheme.com", allowedDomains)
	assert.False(t, result)

	result = utils.ValidateRedirectURI("http:///missing-host", allowedDomains)
	assert.False(t, result)
}

func Test_ValidateRedirectURI_ShouldFailForEmptyURL(t *testing.T) {
	allowedDomains := []string{"example.com"}

	// Empty URL
	result := utils.ValidateRedirectURI("", allowedDomains)
	assert.False(t, result)
}

func Test_ValidateRedirectURI_ShouldFailForMissingScheme(t *testing.T) {
	allowedDomains := []string{"example.com"}

	// Missing scheme (should fail if scheme is required)
	result := utils.ValidateRedirectURI("example.com/callback", allowedDomains)
	assert.False(t, result)
}

func Test_ValidateRedirectURI_ShouldFailForMissingHost(t *testing.T) {
	allowedDomains := []string{"example.com"}

	// Missing host
	result := utils.ValidateRedirectURI("http:///callback", allowedDomains)
	assert.False(t, result)
}

// GetAllowedRedirectDomains() tests
func Test_GetAllowedRedirectDomains_ShouldReturnDomains_WhenEnvIsSet(t *testing.T) {
	os.Setenv("ALLOWED_REDIRECT_DOMAINS", "example.com,localhost")
	defer os.Unsetenv("ALLOWED_REDIRECT_DOMAINS")

	domains, err := utils.GetAllowedRedirectDomains()

	assert.NoError(t, err)
	expected := []string{"example.com", "localhost"}
	assert.Equal(t, expected, domains)
}

func Test_GetAllowedRedirectDomains_ShouldReturnError_WhenEnvIsNotSet(t *testing.T) {
	os.Unsetenv("ALLOWED_REDIRECT_DOMAINS")

	domains, err := utils.GetAllowedRedirectDomains()

	assert.Error(t, err)
	assert.Nil(t, domains)
	assert.EqualError(t, err, "ALLOWED_REDIRECT_DOMAINS is not set in the environment")
}

func Test_GetAllowedRedirectDomains_ShouldReturnError_WhenEmptyString(t *testing.T) {
	os.Setenv("ALLOWED_REDIRECT_DOMAINS", "")
	defer os.Unsetenv("ALLOWED_REDIRECT_DOMAINS")

	domains, err := utils.GetAllowedRedirectDomains()

	assert.Error(t, err)
	assert.Nil(t, domains)
	assert.EqualError(t, err, "ALLOWED_REDIRECT_DOMAINS is not set in the environment")
}
