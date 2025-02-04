package services

import (
	"auth-service/redisclient"
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"log"
	"strings"
	"time"
)

// Constructs a Redis key for storing OAuth tokens per session, provider, and user ID
func constructRedisKey(sessionID, provider, userID string) string {
	return fmt.Sprintf("session:%s_%s_%s", sessionID, provider, userID)
}

// StoreAuthToken stores an OAuth token for a specific user under a provider
func StoreAuthToken(sessionID, provider, userID string, token *oauth2.Token) error {
	key := constructRedisKey(sessionID, provider, userID)

	// Serialize token into JSON
	tokenData, err := json.Marshal(token)
	if err != nil {
		log.Printf("Failed to serialize token: %v", err)
		return err
	}

	// Store token in Redis with expiration based on the token's expiry time
	err = redisclient.Client.Set(context.Background(), key, tokenData, time.Until(token.Expiry)).Err()
	if err != nil {
		log.Printf("Failed to store token in Redis: %v", err)
		return err
	}

	log.Printf("Token stored in Redis for session %s, provider %s, user %s", sessionID, provider, userID)
	return nil
}

// GetAuthToken retrieves an OAuth token for a specific provider and user account
func GetAuthToken(sessionID, provider, userID string) (*oauth2.Token, bool) {
	key := constructRedisKey(sessionID, provider, userID)

	// Retrieve token from Redis
	tokenData, err := redisclient.Client.Get(context.Background(), key).Result()
	if err != nil {
		log.Printf("Failed to retrieve token from Redis: %v", err)
		return nil, false
	}

	// Deserialize token from JSON
	var token oauth2.Token
	err = json.Unmarshal([]byte(tokenData), &token)
	if err != nil {
		log.Printf("Failed to deserialize token: %v", err)
		return nil, false
	}

	log.Printf("Token retrieved from Redis for session %s, provider %s, user %s", sessionID, provider, userID)
	return &token, true
}

type LoggedInProvider struct {
	Provider string `json:"provider"`
	User     string `json:"user"`
	LoggedIn bool   `json:"logged_in"`
}

// GetLoggedInProviders returns all logged-in providers and user IDs for a session
func GetLoggedInProviders(sessionID string) ([]LoggedInProvider, error) {
	// Pattern to search for all providers under the session
	pattern := fmt.Sprintf("session:%s_*", sessionID)
	keys, err := redisclient.Client.Keys(context.Background(), pattern).Result()
	if err != nil {
		log.Printf("Failed to fetch keys from Redis: %v", err)
		return nil, err
	}

	var loggedInProviders []LoggedInProvider

	for _, key := range keys {
		// Extract provider and user ID from Redis key
		parts := strings.Split(key, "_")
		if len(parts) < 3 {
			continue // Skip invalid keys
		}

		provider := parts[1] // Extracts provider from `session:<session_id>_<provider>_<user_id>`
		userID := parts[2]   // Extracts user ID

		// Fetch stored token (ignoring expiration logic)
		tokenData, err := redisclient.Client.Get(context.Background(), key).Result()
		if err != nil {
			continue // Skip if token retrieval fails
		}

		var token oauth2.Token
		if err := json.Unmarshal([]byte(tokenData), &token); err == nil {
			// Always mark as logged in (ignoring expiry check)
			loggedInProviders = append(loggedInProviders, LoggedInProvider{
				Provider: provider,
				User:     userID,
				LoggedIn: true,
			})
		}
	}

	return loggedInProviders, nil
}

// DeleteAuthToken removes an OAuth token for a specific provider and user account
func DeleteAuthToken(sessionID, provider, userID string) error {
	key := constructRedisKey(sessionID, provider, userID)

	err := redisclient.Client.Del(context.Background(), key).Err()
	if err != nil {
		log.Printf("Failed to delete token from Redis: %v", err)
		return err
	}

	log.Printf("Token deleted from Redis for session %s, provider %s, user %s", sessionID, provider, userID)
	return nil
}
