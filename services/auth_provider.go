package services

import (
	"auth-service/config"
	"context"
	"encoding/json"
	"errors"
	"golang.org/x/oauth2"
)

// GetUserID fetches the authenticated user ID from the provider
func GetUserID(provider string, token *oauth2.Token) (string, error) {
	client := config.Providers[provider].Client(context.Background(), token)
	url, err := config.GetProviderUserInfoURL(provider)
	if err != nil {
		return "", err
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var user struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", err
	}

	if user.ID == "" {
		return "", errors.New("user ID not found in response")
	}

	return user.ID, nil
}
