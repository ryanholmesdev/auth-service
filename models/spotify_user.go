package models

import (
	"errors"
	"strings"
)

// SpotifyUserResponse models the JSON returned by Spotify.
type SpotifyUserResponse struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}

// ToUserInfo converts the SpotifyUserResponse to a unified UserInfo.
func (s *SpotifyUserResponse) ToUserInfo() (*UserInfo, error) {
	if strings.TrimSpace(s.ID) == "" {
		return nil, errors.New("user ID is missing in Spotify response")
	}
	if strings.TrimSpace(s.DisplayName) == "" {
		return nil, errors.New("display name is missing in Spotify response")
	}
	if strings.TrimSpace(s.Email) == "" || !validateEmail(s.Email) {
		return nil, errors.New("invalid or missing email in Spotify response")
	}
	return &UserInfo{
		ID:          s.ID,
		DisplayName: s.DisplayName,
		Email:       s.Email,
	}, nil
}
