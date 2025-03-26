package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/monzo/slog"
	"golang.org/x/oauth2"
)

var Providers map[string]*oauth2.Config

func InitConfig() {
	ctx := context.Background()
	slog.Info(ctx, "Initializing OAuth providers configuration")

	Providers = map[string]*oauth2.Config{
		"spotify": {
			ClientID:     getEnv("SPOTIFY_CLIENT_ID", ""),
			ClientSecret: getEnv("SPOTIFY_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("SPOTIFY_REDIRECT_URL", ""),
			Scopes:       []string{"playlist-read-private", "playlist-modify-public", "user-read-email", "user-read-private"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.spotify.com/authorize",
				TokenURL: "https://accounts.spotify.com/api/token",
			},
		},
		"tidal": {
			ClientID:     getEnv("TIDAL_CLIENT_ID", ""),
			ClientSecret: getEnv("TIDAL_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("TIDAL_REDIRECT_URL", ""),
			Scopes:       []string{"playlists.read", "collection.read", "playlists.write", "user.read"},
			Endpoint: oauth2.Endpoint{
				AuthURL:   "https://login.tidal.com/authorize",
				TokenURL:  "https://auth.tidal.com/v1/oauth2/token",
				AuthStyle: oauth2.AuthStyleInParams,
			},
		},
	}
	validateProviders()
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func validateProviders() {
	ctx := context.Background()
	for name, config := range Providers {
		if config.ClientID == "" {
			slog.Error(ctx, "Missing environment variable", fmt.Errorf("missing client ID"), map[string]interface{}{
				"provider": name,
				"var":      fmt.Sprintf("%s_CLIENT_ID", strings.ToUpper(name)),
			})
			panic(fmt.Sprintf("Missing environment variable for: %s_CLIENT_ID", strings.ToUpper(name)))
		}
		if config.ClientSecret == "" {
			slog.Error(ctx, "Missing environment variable", fmt.Errorf("missing client secret"), map[string]interface{}{
				"provider": name,
				"var":      fmt.Sprintf("%s_CLIENT_SECRET", strings.ToUpper(name)),
			})
			panic(fmt.Sprintf("Missing environment variable for: %s_CLIENT_SECRET", strings.ToUpper(name)))
		}
		if config.RedirectURL == "" {
			slog.Error(ctx, "Missing environment variable", fmt.Errorf("missing redirect URL"), map[string]interface{}{
				"provider": name,
				"var":      fmt.Sprintf("%s_REDIRECT_URL", strings.ToUpper(name)),
			})
			panic(fmt.Sprintf("Missing environment variable for: %s_REDIRECT_URL", strings.ToUpper(name)))
		}
	}
	slog.Info(ctx, "Successfully validated all provider configurations", map[string]interface{}{
		"providers": getProviderNames(),
	})
}

func getProviderNames() []string {
	names := make([]string, 0, len(Providers))
	for name := range Providers {
		names = append(names, name)
	}
	return names
}

var GetProviderUserInfoURL = func(provider string) (string, error) {
	ctx := context.Background()
	switch provider {
	case "spotify":
		return "https://api.spotify.com/v1/me", nil
	case "tidal":
		return "https://openapi.tidal.com/v2/users/me", nil
	default:
		slog.Error(ctx, "Unsupported provider", fmt.Errorf("provider not supported"), map[string]interface{}{
			"provider": provider,
		})
		return "", fmt.Errorf("provider not supported: %s", provider)
	}
}
