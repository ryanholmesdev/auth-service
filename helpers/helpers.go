package helpers

import (
	"fmt"
	"net/url"
	"strings"
)

func ConstructSessionKey(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
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

	// Check if the domain is in the allowed list
	for _, domain := range allowedDomains {
		if strings.HasSuffix(parsedURL.Host, domain) {
			return true
		}
	}

	return false
}
