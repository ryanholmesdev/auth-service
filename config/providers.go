package config

import (
	"fmt"
	"golang.org/x/oauth2"
	"log"
	"os"
	"strings"
)

var Providers map[string]*oauth2.Config

func InitConfig() {
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
	for name, config := range Providers {
		if config.ClientID == "" {
			log.Fatalf("Missing environment variable for: %s_CLIENT_ID", strings.ToUpper(name))
		}
		if config.ClientSecret == "" {
			log.Fatalf("Missing environment variable for: %s_CLIENT_SECRET", strings.ToUpper(name))
		}
		if config.RedirectURL == "" {
			log.Fatalf("Missing environment variable for: %s_REDIRECT_URL", strings.ToUpper(name))
		}
	}
}

var GetProviderUserInfoURL = func(provider string) (string, error) {
	switch provider {
	case "spotify":
		return "https://api.spotify.com/v1/me", nil
	case "tidal":
		return "https://openapi.tidal.com/v2/users/me", nil
	default:
		return "", fmt.Errorf("provider not supported: %s", provider)
	}
}
