package config

import (
	"fmt"
	"golang.org/x/oauth2"
	"log"
	"os"
)

var Providers map[string]*oauth2.Config

func InitConfig() {
	// Initialize the Providers map
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
			log.Fatalf("Missing environment variable for %s: SPOTIFY_CLIENT_ID", name)
		}
		if config.ClientSecret == "" {
			log.Fatalf("Missing environment variable for %s: SPOTIFY_CLIENT_SECRET", name)
		}
		if config.RedirectURL == "" {
			log.Fatalf("Missing environment variable for %s: SPOTIFY_REDIRECT_URL", name)
		}
	}
	fmt.Println("All providers are configured correctly.")
}
