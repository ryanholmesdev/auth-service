package mocks

import (
	"golang.org/x/oauth2"
)

// MockOAuthConfig mocks the oauth2.Config
type MockOAuthConfig struct {
	oauth2.Config
}

func (m *MockOAuthConfig) AuthCodeURL(state string, _ ...oauth2.AuthCodeOption) string {
	return "http://mock-oauth-provider/authorize?state=" + state
}
