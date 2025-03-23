package models

import (
	"errors"
	"strings"
)

// TidalUserResponse models the nested JSON returned by Tidal.
type TidalUserResponse struct {
	Data struct {
		ID         string `json:"id"`
		Attributes struct {
			Username      string `json:"username"`
			Email         string `json:"email"`
			EmailVerified bool   `json:"emailVerified"`
			Country       string `json:"country"`
		} `json:"attributes"`
	} `json:"data"`
}

// ToUserInfo converts the TidalUserResponse to a unified UserInfo.
func (t *TidalUserResponse) ToUserInfo() (*UserInfo, error) {
	data := t.Data
	if strings.TrimSpace(data.ID) == "" {
		return nil, errors.New("user ID is missing in Tidal response")
	}
	if strings.TrimSpace(data.Attributes.Username) == "" {
		return nil, errors.New("username is missing in Tidal response")
	}
	if strings.TrimSpace(data.Attributes.Email) == "" || !validateEmail(data.Attributes.Email) {
		return nil, errors.New("invalid or missing email in Tidal response")
	}
	return &UserInfo{
		ID:          data.ID,
		DisplayName: data.Attributes.Username,
		Email:       data.Attributes.Email,
	}, nil
}
