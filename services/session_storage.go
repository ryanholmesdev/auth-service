package services

import (
	"auth-service/redisclient"
	"golang.org/x/net/context"
	"log"
	"time"
)

func ConstructSessionKey(stateToken string) string {
	return "state:" + stateToken
}

var StoreStateToken = storeStateToken

// StoreStateToken stores the CSRF state token in Redis with a TTL
func storeStateToken(stateToken string) error {
	key := ConstructSessionKey(stateToken)
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
	key := ConstructSessionKey(stateToken)
	_, err := redisclient.Client.Get(context.Background(), key).Result()
	if err != nil {
		log.Printf("State token validation failed: %v", err)
		return false
	}
	return true
}

// DeleteStateToken removes the CSRF state token from Redis
func DeleteStateToken(stateToken string) error {
	key := ConstructSessionKey(stateToken)
	err := redisclient.Client.Del(context.Background(), key).Err()
	if err != nil {
		log.Printf("Failed to delete state token from Redis: %v", err)
		return err
	}
	log.Printf("State token deleted from Redis: %s", stateToken)
	return nil
}
