package services

import (
	"auth-service/redisclient"
	"context"
	"encoding/json"
	"golang.org/x/oauth2"
	"log"
	"time"
)

func constructRedisKey(sessionID, provider string) string {
	return "session:" + sessionID + "_" + provider
}

func StoreAuthToken(sessionID string, provider string, token *oauth2.Token) error {
	key := constructRedisKey(sessionID, provider)

	// Serialize the token into JSON
	tokenData, err := json.Marshal(token)
	if err != nil {
		log.Printf("Failed to serialize token: %v", err)
		return err
	}

	// Store the token in Redis, with a TTL of 24 hours
	err = redisclient.Client.Set(context.Background(), key, tokenData, time.Hour*24).Err()
	if err != nil {
		log.Printf("Failed to store token in Redis: %v", err)
		return err
	}

	log.Printf("Token stored in Redis for session %s, provider %s", sessionID, provider)
	return nil
}

func GetAuthToken(sessionID string, provider string) (*oauth2.Token, bool) {
	key := constructRedisKey(sessionID, provider)

	// Retrieve the token from Redis
	tokenData, err := redisclient.Client.Get(context.Background(), key).Result()
	if err != nil {
		log.Printf("Failed to retrieve token from Redis: %v", err)
		return nil, false
	}

	// Deserialize the token from JSON
	var token oauth2.Token
	err = json.Unmarshal([]byte(tokenData), &token)
	if err != nil {
		log.Printf("Failed to deserialize token: %v", err)
		return nil, false
	}

	log.Printf("Token retrieved from Redis for session %s, provider %s", sessionID, provider)
	return &token, true
}

func DeleteAuthToken(sessionID string, provider string) error {
	key := constructRedisKey(sessionID, provider)

	err := redisclient.Client.Del(context.Background(), key).Err()
	if err != nil {
		log.Printf("Failed to delete token from Redis: %v", err)
		return err
	}

	log.Printf("Token deleted from Redis for session %s, provider %s", sessionID, provider)
	return nil
}
