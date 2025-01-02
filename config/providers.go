package config

import (
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
			Scopes:       []string{"playlist-read-private", "playlist-modify-public"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.spotify.com/authorize",
				TokenURL: "https://accounts.spotify.com/api/token",
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
