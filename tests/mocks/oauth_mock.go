package mocks

import (
	"auth-service/models"
	"golang.org/x/oauth2"
	"time"
)

// MockOAuthConfig mocks the oauth2.Config
type MockOAuthConfig struct {
	oauth2.Config
}

func (m *MockOAuthConfig) AuthCodeURL(state string, _ ...oauth2.AuthCodeOption) string {
	return "http://mock-oauth-provider/authorize?state=" + state
}

func NewMockOAuth2Token(provider string, duration time.Duration) *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  "mock-access-token-" + provider,
		RefreshToken: "mock-refresh-token-" + provider,
		Expiry:       time.Now().Add(duration),
	}
}

func NewMockUser(provider, userID, displayName, email string) *models.UserInfo {
	return &models.UserInfo{
		ID:          userID,
		DisplayName: displayName,
		Email:       email,
	}
}
