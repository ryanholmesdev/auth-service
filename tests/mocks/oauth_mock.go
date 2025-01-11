package mocks

import (
	"github.com/stretchr/testify/mock"
)

// MockOAuthConfig mocks the oauth2.Config
type MockOAuthConfig struct {
	mock.Mock
}

func (m *MockOAuthConfig) AuthCodeURL(state string, _ ...interface{}) string {
	args := m.Called(state)
	return args.String(0)
}
