package models

import (
	"net/mail"
)

type UserInfo struct {
	ID          string
	DisplayName string
	Email       string
}

// ProviderResponse defines the behavior required to convert a provider's response to a UserInfo.
type ProviderResponse interface {
	ToUserInfo() (*UserInfo, error)
}

// validateEmail checks if an email is valid using net/mail.
func validateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
