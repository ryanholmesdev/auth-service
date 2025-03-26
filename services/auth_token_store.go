package services

import (
	"auth-service/models"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/go-redis/redis/v8"
	"github.com/monzo/slog"
)

var redisClient *redis.Client

func init() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}

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

// AuthToken represents a stored OAuth token with associated user information.
type AuthToken struct {
	Token       *oauth2.Token
	UserID      string
	DisplayName string
	Email       string
}

// StoreAuthToken stores an OAuth token with associated user information in Redis.
func StoreAuthToken(sessionID, provider string, user *models.UserInfo, token *oauth2.Token) error {
	ctx := context.Background()

	// Create the token data structure
	tokenData := AuthToken{
		Token:       token,
		UserID:      user.ID,
		DisplayName: user.DisplayName,
		Email:       user.Email,
	}

	// Marshal the token data to JSON
	tokenJSON, err := json.Marshal(tokenData)
	if err != nil {
		slog.Error(ctx, "Failed to marshal token data", err, map[string]interface{}{
			"provider":   provider,
			"session_id": sessionID,
			"user_id":    user.ID,
		})
		return fmt.Errorf("failed to marshal token data: %w", err)
	}

	// Store the token in Redis with a key that includes the session ID, provider, and user ID
	key := fmt.Sprintf("auth_token:%s:%s:%s", sessionID, provider, user.ID)
	if err := redisClient.Set(ctx, key, tokenJSON, 24*time.Hour).Err(); err != nil {
		slog.Error(ctx, "Failed to store token in Redis", err, map[string]interface{}{
			"provider":   provider,
			"session_id": sessionID,
			"user_id":    user.ID,
		})
		return fmt.Errorf("failed to store token in Redis: %w", err)
	}

	slog.Info(ctx, "Successfully stored auth token", map[string]interface{}{
		"provider":     provider,
		"session_id":   sessionID,
		"user_id":      user.ID,
		"display_name": user.DisplayName,
		"email":        user.Email,
	})
	return nil
}

// GetAuthToken retrieves an OAuth token from Redis.
func GetAuthToken(sessionID, provider, userID string) (*AuthToken, bool) {
	ctx := context.Background()

	// Construct the Redis key
	key := fmt.Sprintf("auth_token:%s:%s:%s", sessionID, provider, userID)

	// Retrieve the token data from Redis
	tokenJSON, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			slog.Info(ctx, "Token not found in Redis", map[string]interface{}{
				"provider":   provider,
				"session_id": sessionID,
				"user_id":    userID,
			})
			return nil, false
		}
		slog.Error(ctx, "Failed to retrieve token from Redis", err, map[string]interface{}{
			"provider":   provider,
			"session_id": sessionID,
			"user_id":    userID,
		})
		return nil, false
	}

	// Unmarshal the token data
	var tokenData AuthToken
	if err := json.Unmarshal([]byte(tokenJSON), &tokenData); err != nil {
		slog.Error(ctx, "Failed to unmarshal token data", err, map[string]interface{}{
			"provider":   provider,
			"session_id": sessionID,
			"user_id":    userID,
		})
		return nil, false
	}

	slog.Info(ctx, "Successfully retrieved auth token", map[string]interface{}{
		"provider":     provider,
		"session_id":   sessionID,
		"user_id":      userID,
		"display_name": tokenData.DisplayName,
		"email":        tokenData.Email,
	})
	return &tokenData, true
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
	ctx := context.Background()
	// Pattern to search for all providers under the session
	pattern := fmt.Sprintf("session:%s_*", sessionID)
	keys, err := redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		slog.Error(ctx, "Failed to fetch keys from Redis", err, map[string]interface{}{
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
		authDataJSON, err := redisClient.Get(ctx, key).Result()
		if err != nil {
			slog.Error(ctx, "Failed to retrieve auth data from Redis", err, map[string]interface{}{
				"key":      key,
				"provider": provider,
			})
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
		} else {
			slog.Error(ctx, "Failed to unmarshal auth data", err, map[string]interface{}{
				"key":      key,
				"provider": provider,
			})
		}
	}

	slog.Info(ctx, "Successfully retrieved logged in providers", map[string]interface{}{
		"session_id": sessionID,
		"count":      len(loggedInProviders),
	})
	return loggedInProviders, nil
}

// DeleteAuthToken deletes an OAuth token from Redis.
func DeleteAuthToken(sessionID, provider, userID string) error {
	ctx := context.Background()

	// Construct the Redis key
	key := fmt.Sprintf("auth_token:%s:%s:%s", sessionID, provider, userID)

	// Delete the token from Redis
	if err := redisClient.Del(ctx, key).Err(); err != nil {
		slog.Error(ctx, "Failed to delete token from Redis", err, map[string]interface{}{
			"provider":   provider,
			"session_id": sessionID,
			"user_id":    userID,
		})
		return fmt.Errorf("failed to delete token from Redis: %w", err)
	}

	slog.Info(ctx, "Successfully deleted auth token", map[string]interface{}{
		"provider":   provider,
		"session_id": sessionID,
		"user_id":    userID,
	})
	return nil
}

// DeleteAllAuthTokensForProvider deletes all OAuth tokens for a specific provider and session.
func DeleteAllAuthTokensForProvider(sessionID, provider string) error {
	ctx := context.Background()

	// Construct the pattern to match all keys for this provider and session
	pattern := fmt.Sprintf("auth_token:%s:%s:*", sessionID, provider)

	// Get all keys matching the pattern
	keys, err := redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		slog.Error(ctx, "Failed to get keys from Redis", err, map[string]interface{}{
			"provider":   provider,
			"session_id": sessionID,
		})
		return fmt.Errorf("failed to get keys from Redis: %w", err)
	}

	if len(keys) == 0 {
		slog.Info(ctx, "No tokens found to delete", map[string]interface{}{
			"provider":   provider,
			"session_id": sessionID,
		})
		return nil
	}

	// Delete all matching keys
	if err := redisClient.Del(ctx, keys...).Err(); err != nil {
		slog.Error(ctx, "Failed to delete tokens from Redis", err, map[string]interface{}{
			"provider":   provider,
			"session_id": sessionID,
			"key_count":  len(keys),
		})
		return fmt.Errorf("failed to delete tokens from Redis: %w", err)
	}

	slog.Info(ctx, "Successfully deleted all tokens for provider", map[string]interface{}{
		"provider":   provider,
		"session_id": sessionID,
		"key_count":  len(keys),
	})
	return nil
}
