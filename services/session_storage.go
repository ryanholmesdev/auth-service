package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/monzo/slog"
)

type PKCEData struct {
	CodeVerifier string `json:"code_verifier"`
}

// StorePKCEData stores the PKCE data (including the code verifier) in Redis,
// using the state token as the key.
func StorePKCEData(stateToken, codeVerifier string) error {
	ctx := context.Background()
	key := "pkce:" + stateToken
	data := PKCEData{
		CodeVerifier: codeVerifier,
	}
	b, err := json.Marshal(data)
	if err != nil {
		slog.Error(ctx, "Failed to marshal PKCE data", err, map[string]interface{}{
			"state_token": stateToken,
		})
		return fmt.Errorf("failed to marshal PKCE data: %w", err)
	}
	err = redisClient.Set(ctx, key, b, time.Minute*5).Err()
	if err != nil {
		slog.Error(ctx, "Failed to store PKCE data in Redis", err, map[string]interface{}{
			"state_token": stateToken,
		})
		return fmt.Errorf("failed to store PKCE data in Redis: %w", err)
	}
	slog.Info(ctx, "Successfully stored PKCE data", map[string]interface{}{
		"state_token": stateToken,
	})
	return nil
}

// GetCodeVerifier retrieves the code verifier from Redis for the given state token.
func GetCodeVerifier(stateToken string) (string, error) {
	ctx := context.Background()
	key := "pkce:" + stateToken
	result, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			slog.Info(ctx, "PKCE data not found in Redis", map[string]interface{}{
				"state_token": stateToken,
			})
			return "", fmt.Errorf("PKCE data not found")
		}
		slog.Error(ctx, "Failed to retrieve PKCE data from Redis", err, map[string]interface{}{
			"state_token": stateToken,
		})
		return "", fmt.Errorf("failed to retrieve PKCE data from Redis: %w", err)
	}
	var data PKCEData
	if err = json.Unmarshal([]byte(result), &data); err != nil {
		slog.Error(ctx, "Failed to unmarshal PKCE data", err, map[string]interface{}{
			"state_token": stateToken,
		})
		return "", fmt.Errorf("failed to unmarshal PKCE data: %w", err)
	}
	slog.Info(ctx, "Successfully retrieved PKCE data", map[string]interface{}{
		"state_token": stateToken,
	})
	return data.CodeVerifier, nil
}

// DeletePKCEData removes the PKCE data from Redis for the given state token.
func DeletePKCEData(stateToken string) error {
	ctx := context.Background()
	key := "pkce:" + stateToken
	err := redisClient.Del(ctx, key).Err()
	if err != nil {
		slog.Error(ctx, "Failed to delete PKCE data from Redis", err, map[string]interface{}{
			"state_token": stateToken,
		})
		return fmt.Errorf("failed to delete PKCE data from Redis: %w", err)
	}
	slog.Info(ctx, "Successfully deleted PKCE data", map[string]interface{}{
		"state_token": stateToken,
	})
	return nil
}
