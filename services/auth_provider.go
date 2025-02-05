package services

import (
	"auth-service/config"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/oauth2"
	"net/http"
	"regexp"
	"strings"
)

// UserInfo holds the user details retrieved from the provider
type UserInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}

// validateEmail checks if an email is valid
func validateEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

// GetUserInfo retrieves the user's unique ID, name, and email from the provider's API
func GetUserInfo(provider string, token *oauth2.Token) (*UserInfo, error) {
	// Get the provider's user info endpoint URL
	url, err := config.GetProviderUserInfoURL(provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider user info URL: %w", err)
	}

	// Create an HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Attach Authorization header
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	// Make the request
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("provider returned non-200 status: %d", resp.StatusCode)
	}

	// Parse the JSON response
	var user UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if strings.TrimSpace(user.ID) == "" {
		return nil, errors.New("user ID is missing in response")
	}

	if strings.TrimSpace(user.DisplayName) == "" {
		return nil, errors.New("display name is missing in response")
	}

	if strings.TrimSpace(user.Email) == "" {
		return nil, errors.New("email is missing in response")
	}
	if !validateEmail(user.Email) {
		return nil, errors.New("invalid email format")
	}

	return &user, nil
}
