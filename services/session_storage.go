package services

import (
	"auth-service/redisclient"
	"golang.org/x/net/context"
	"log"
	"time"
)

// Construct a Redis key for session state
func constructSessionKey(stateToken string) string {
	return "state:" + stateToken
}

// StoreStateToken stores the CSRF state token in Redis with a TTL
func StoreStateToken(stateToken string) error {
	key := constructSessionKey(stateToken)
	err := redisclient.Client.Set(context.Background(), key, "valid", time.Minute*5).Err()
	if err != nil {
		log.Printf("Failed to store state token in Redis: %v", err)
		return err
	}
	log.Printf("State token stored in Redis: %s", stateToken)
	return nil
}

// ValidateStateToken checks if the CSRF state token exists in Redis
func ValidateStateToken(stateToken string) bool {
	key := constructSessionKey(stateToken)
	_, err := redisclient.Client.Get(context.Background(), key).Result()
	if err != nil {
		log.Printf("State token validation failed: %v", err)
		return false
	}
	return true
}

// DeleteStateToken removes the CSRF state token from Redis
func DeleteStateToken(stateToken string) error {
	key := constructSessionKey(stateToken)
	err := redisclient.Client.Del(context.Background(), key).Err()
	if err != nil {
		log.Printf("Failed to delete state token from Redis: %v", err)
		return err
	}
	log.Printf("State token deleted from Redis: %s", stateToken)
	return nil
}
