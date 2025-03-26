package services

import (
	"context"
	"fmt"
	"net/http"

	"auth-service/config"
	"auth-service/models"

	"github.com/go-resty/resty/v2"
	"github.com/monzo/slog"
	"golang.org/x/oauth2"
)

func getClient() *resty.Client {
	return resty.New()
}

// makeAuthenticatedRequest sends an authenticated GET request using the provided OAuth token.
// It uses generics to decode the response into the specific provider type.
func makeAuthenticatedRequest[T any, P interface {
	*T
	models.ProviderResponse
}](ctx context.Context, url string, token *oauth2.Token) (P, error) {
	var result T
	client := getClient()

	slog.Info(ctx, "Making authenticated request to provider", map[string]interface{}{
		"url": url,
	})

	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token.AccessToken).
		SetResult(&result).
		Get(url)
	if err != nil {
		slog.Error(ctx, "Failed to make authenticated request", err, map[string]interface{}{
			"url": url,
		})
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		slog.Error(ctx, "Provider returned non-OK status", fmt.Errorf("status code: %d", resp.StatusCode()), map[string]interface{}{
			"url":         url,
			"status_code": resp.StatusCode(),
		})
		return nil, fmt.Errorf("provider returned non-OK status: %d", resp.StatusCode())
	}
	return P(&result), nil
}

// GetUserInfo retrieves the user info from the given provider by delegating to
// the appropriate models conversion logic.
func GetUserInfo(ctx context.Context, provider string, token *oauth2.Token) (*models.UserInfo, error) {
	url, err := config.GetProviderUserInfoURL(provider)
	if err != nil {
		slog.Error(ctx, "Failed to get provider user info URL", err, map[string]interface{}{
			"provider": provider,
		})
		return nil, fmt.Errorf("failed to get provider user info URL: %w", err)
	}

	slog.Info(ctx, "Fetching user info from provider", map[string]interface{}{
		"provider": provider,
		"url":      url,
	})

	var response models.ProviderResponse

	switch provider {
	case "spotify":
		spotifyUser, err := makeAuthenticatedRequest[models.SpotifyUserResponse, *models.SpotifyUserResponse](ctx, url, token)
		if err != nil {
			return nil, err
		}
		response = spotifyUser

	case "tidal":
		tidalUser, err := makeAuthenticatedRequest[models.TidalUserResponse, *models.TidalUserResponse](ctx, url, token)
		if err != nil {
			return nil, err
		}
		response = tidalUser

	default:
		err := fmt.Errorf("unsupported provider: %s", provider)
		slog.Error(ctx, "Unsupported provider", err, map[string]interface{}{
			"provider": provider,
		})
		return nil, err
	}

	userInfo, err := response.ToUserInfo()
	if err != nil {
		slog.Error(ctx, "Failed to convert provider response to user info", err, map[string]interface{}{
			"provider": provider,
		})
		return nil, err
	}

	slog.Info(ctx, "Successfully retrieved user info", map[string]interface{}{
		"provider":     provider,
		"user_id":      userInfo.ID,
		"display_name": userInfo.DisplayName,
		"email":        userInfo.Email,
	})

	return userInfo, nil
}

// UserInfo is our normalized user info structure
type UserInfo struct {
	ID          string
	DisplayName string
	Email       string
}
