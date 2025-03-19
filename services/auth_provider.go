package services

import (
	"auth-service/config"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/go-resty/resty/v2"
	"golang.org/x/oauth2"
)

type SpotifyUserInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}

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

type UserInfo struct {
	ID          string
	DisplayName string
	Email       string
}

type ProviderResponse interface {
	ToUserInfo() (*UserInfo, error)
}

func (s *SpotifyUserInfo) ToUserInfo() (*UserInfo, error) {
	if strings.TrimSpace(s.ID) == "" {
		return nil, errors.New("user ID is missing in response")
	}
	if strings.TrimSpace(s.DisplayName) == "" {
		return nil, errors.New("display name is missing in response")
	}
	if strings.TrimSpace(s.Email) == "" || !validateEmail(s.Email) {
		return nil, errors.New("invalid email format or missing email")
	}
	return &UserInfo{
		ID:          s.ID,
		DisplayName: s.DisplayName,
		Email:       s.Email,
	}, nil
}

func (t *TidalUserResponse) ToUserInfo() (*UserInfo, error) {
	data := t.Data
	if strings.TrimSpace(data.ID) == "" {
		return nil, errors.New("user ID is missing in response")
	}
	if strings.TrimSpace(data.Attributes.Username) == "" {
		return nil, errors.New("username is missing in response")
	}
	if strings.TrimSpace(data.Attributes.Email) == "" || !validateEmail(data.Attributes.Email) {
		return nil, errors.New("invalid email format or missing email")
	}
	return &UserInfo{
		ID:          data.ID,
		DisplayName: data.Attributes.Username,
		Email:       data.Attributes.Email,
	}, nil
}

func validateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func makeAuthenticatedRequest[T any, P interface {
	*T
	ProviderResponse
}](url string, token *oauth2.Token) (P, error) {
	var result T
	client := resty.New()

	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+token.AccessToken).
		SetResult(&result).
		Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, errors.New("provider returned non-200 status")
	}
	return P(&result), nil
}

func GetUserInfo(provider string, token *oauth2.Token) (*UserInfo, error) {
	url, err := config.GetProviderUserInfoURL(provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider user info URL: %w", err)
	}

	var response ProviderResponse
	switch provider {
	case "spotify":
		spotifyUser, err := makeAuthenticatedRequest[SpotifyUserInfo, *SpotifyUserInfo](url, token)
		if err != nil {
			return nil, err
		}
		response = spotifyUser

	case "tidal":
		tidalUser, err := makeAuthenticatedRequest[TidalUserResponse, *TidalUserResponse](url, token)
		if err != nil {
			return nil, err
		}
		response = tidalUser

	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	return response.ToUserInfo()
}
