package services

import (
	"auth-service/models"
	"auth-service/redisclient"
	"context"
	"encoding/json"
	"fmt"
	"github.com/monzo/slog"
	"log"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// Constructs a Redis key for storing OAuth tokens per session, provider, and user ID
func constructRedisKey(sessionID, provider, userID string) string {
	return fmt.Sprintf("session:%s_%s_%s", sessionID, provider, userID)
}

// AuthData represents stored auth info in Redis
type AuthData struct {
	Token       *oauth2.Token `json:"token"`
	UserID      string        `json:"user_id"`
	DisplayName string        `json:"display_name"`
	Email       string        `json:"email"`
}

// StoreAuthToken stores OAuth token and user info in Redis
func StoreAuthToken(sessionID, provider string, userInfo *models.UserInfo, token *oauth2.Token) error {
	key := constructRedisKey(sessionID, provider, userInfo.ID)

	authData := AuthData{
		Token:       token,
		UserID:      userInfo.ID,
		DisplayName: userInfo.DisplayName,
		Email:       userInfo.Email,
	}

	// Serialize auth data into JSON
	authDataJSON, err := json.Marshal(authData)
	if err != nil {
		log.Printf("Failed to serialize auth data: %v", err)
		slog.Error(context.Background(), "Failed to serialize auth data", err, map[string]interface{}{
			"session_id": sessionID,
			"provider":   provider,
			"user_id":    userInfo.ID,
		})
		return err
	}

	// Store in Redis with expiration based on token expiry
	err = redisclient.Client.Set(context.Background(), key, authDataJSON, time.Until(token.Expiry)).Err()
	if err != nil {
		log.Printf("Failed to store auth data in Redis: %v", err)
		slog.Error(context.Background(), "Failed to store auth data in Redis", err, map[string]interface{}{
			"session_id": sessionID,
			"provider":   provider,
			"user_id":    userInfo.ID,
		})
		return err
	}

	log.Printf("Stored auth data in Redis for session %s, provider %s, user %s", sessionID, provider, userInfo.ID)
	slog.Info(context.Background(), "Stored auth data in Redis", map[string]interface{}{
		"session_id": sessionID,
		"provider":   provider,
		"user_id":    userInfo.ID,
	})
	return nil
}

// GetAuthToken retrieves OAuth token and user info from Redis
func GetAuthToken(sessionID, provider, userID string) (*AuthData, bool) {
	key := constructRedisKey(sessionID, provider, userID)

	// Retrieve from Redis
	authDataJSON, err := redisclient.Client.Get(context.Background(), key).Result()
	if err != nil {
		log.Printf("Failed to retrieve auth data from Redis: %v", err)
		slog.Error(context.Background(), "Failed to retrieve auth data from Redis", err, map[string]interface{}{
			"session_id": sessionID,
			"provider":   provider,
			"user_id":    userID,
		})
		return nil, false
	}

	// Deserialize JSON
	var authData AuthData
	err = json.Unmarshal([]byte(authDataJSON), &authData)
	if err != nil {
		log.Printf("Failed to deserialize auth data: %v", err)
		slog.Error(context.Background(), "Failed to deserialize auth data", err, map[string]interface{}{
			"session_id": sessionID,
			"provider":   provider,
			"user_id":    userID,
		})
		return nil, false
	}

	return &authData, true
}

type LoggedInProvider struct {
	Provider    string `json:"provider"`
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	LoggedIn    bool   `json:"logged_in"`
}

// GetLoggedInProviders returns all logged-in providers with user details
func GetLoggedInProviders(sessionID string) ([]LoggedInProvider, error) {
	// Pattern to search for all providers under the session
	pattern := fmt.Sprintf("session:%s_*", sessionID)
	keys, err := redisclient.Client.Keys(context.Background(), pattern).Result()
	if err != nil {
		log.Printf("Failed to fetch keys from Redis: %v", err)
		slog.Error(context.Background(), "Failed to fetch keys from Redis", err, map[string]interface{}{
			"session_id": sessionID,
		})
		return nil, err
	}

	var loggedInProviders []LoggedInProvider

	for _, key := range keys {
		// Extract provider and user ID from Redis key
		parts := strings.Split(key, "_")
		if len(parts) < 3 {
			continue // Skip invalid keys
		}

		provider := parts[1] // Extract provider from `session:<session_id>_<provider>_<user_id>`

		// Fetch stored auth data (including token & user details)
		authDataJSON, err := redisclient.Client.Get(context.Background(), key).Result()
		if err != nil {
			continue // Skip if retrieval fails
		}

		var authData AuthData
		if err := json.Unmarshal([]byte(authDataJSON), &authData); err == nil {
			// Append user info to the list
			loggedInProviders = append(loggedInProviders, LoggedInProvider{
				Provider:    provider,
				UserID:      authData.UserID,
				DisplayName: authData.DisplayName,
				Email:       authData.Email,
				LoggedIn:    true,
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
		slog.Error(context.Background(), "Failed to delete token from Redis", err, map[string]interface{}{
			"session_id": sessionID,
			"provider":   provider,
			"user_id":    userID,
		})
		return err
	}

	log.Printf("Token deleted from Redis for session %s, provider %s, user %s", sessionID, provider, userID)
	slog.Info(context.Background(), "Token deleted from Redis", map[string]interface{}{
		"session_id": sessionID,
		"provider":   provider,
		"user_id":    userID,
	})
	return nil
}

func DeleteAllAuthTokensForProvider(sessionID, provider string) error {
	pattern := fmt.Sprintf("session:%s_%s_*", sessionID, provider)
	keys, err := redisclient.Client.Keys(context.Background(), pattern).Result()
	if err != nil {
		log.Printf("Failed to fetch keys for logout: %v", err)
		slog.Error(context.Background(), "Failed to fetch keys for logout", err, map[string]interface{}{
			"session_id": sessionID,
			"provider":   provider,
		})
		return err
	}

	if len(keys) == 0 {
		return nil // Nothing to delete
	}

	err = redisclient.Client.Del(context.Background(), keys...).Err()
	if err != nil {
		log.Printf("Failed to delete keys for provider %s: %v", provider, err)
		slog.Error(context.Background(), "Failed to delete keys for provider", err, map[string]interface{}{
			"session_id": sessionID,
			"provider":   provider,
		})
		return err
	}

	log.Printf("Successfully deleted all tokens for provider %s", provider)
	slog.Info(context.Background(), "Successfully deleted all tokens for provider", map[string]interface{}{
		"provider": provider,
	})
	return nil
}
