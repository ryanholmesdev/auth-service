package utils

import (
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"net/url"
	"strings"
)

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

	// Check if the domain is in the allowed list
	for _, domain := range allowedDomains {
		if strings.HasSuffix(parsedURL.Host, domain) {
			return true
		}
	}

	return false
}

func RefreshAccessToken(oauthConfig *oauth2.Config, refreshToken string) (*oauth2.Token, error) {
	token := &oauth2.Token{RefreshToken: refreshToken}
	newToken, err := oauthConfig.TokenSource(context.Background(), token).Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	return newToken, nil
}
