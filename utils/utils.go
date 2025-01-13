package utils

import (
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"net/url"
	"os"
	"strings"
)

func GetAllowedRedirectDomains() ([]string, error) {
	domains := os.Getenv("ALLOWED_REDIRECT_DOMAINS")
	if domains == "" {
		return nil, fmt.Errorf("ALLOWED_REDIRECT_DOMAINS is not set in the environment")
	}
	return strings.Split(domains, ","), nil
}

// ValidateRedirectURI checks if the URI is valid and belongs to an allowed domain.
func ValidateRedirectURI(uri string, allowedDomains []string) bool {
	parsedURL, err := url.Parse(uri)
	if err != nil {
		return false // Invalid URL
	}

	// Ensure the URL has a scheme and host
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return false
	}

	// Extract the hostname (without port) and convert to lowercase
	host := strings.ToLower(parsedURL.Hostname())

	// Check if the domain is in the allowed list (case-insensitive)
	for _, domain := range allowedDomains {
		if strings.HasSuffix(host, strings.ToLower(domain)) {
			return true
		}
	}

	return false
}

// Declare a variable that defaults to the actual implementation
var RefreshAccessTokenFunc = RefreshAccessToken

// Actual function to refresh token
func RefreshAccessToken(oauthConfig *oauth2.Config, refreshToken string) (*oauth2.Token, error) {
	token := &oauth2.Token{RefreshToken: refreshToken}
	newToken, err := oauthConfig.TokenSource(context.Background(), token).Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	return newToken, nil
}
