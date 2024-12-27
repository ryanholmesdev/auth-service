package config

import "golang.org/x/oauth2"

var Providers = map[string]*oauth2.Config{
	"spotify": {
		ClientID:     "your-spotify-client-id",
		ClientSecret: "your-spotify-client-secret",
		RedirectURL:  "http://localhost:8080/auth/spotify/callback",
		Scopes:       []string{"playlist-read-private", "playlist-modify-public"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.spotify.com/authorize",
			TokenURL: "https://accounts.spotify.com/api/token",
		},
	},
}
