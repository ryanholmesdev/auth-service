package services

import (
	"auth-service/redisclient"
	"context"
	"encoding/json"
	"golang.org/x/oauth2"
	"log"
	"time"
)

func StoreToken(sessionID string, provider string, token *oauth2.Token) error {
	// Construct the Redis key
	key := "session:" + sessionID + "_" + provider

	// Serialize the token into JSON
	tokenData, err := json.Marshal(token)
	if err != nil {
		log.Printf("Failed to serialize token: %v", err)
		return err
	}

	// Store the token in Redis
	err = redisclient.Client.Set(context.Background(), key, tokenData, time.Hour*24).Err() // TTL of 24 hours
	if err != nil {
		log.Printf("Failed to store token in Redis: %v", err)
		return err
	}

	log.Printf("Token stored in Redis for session %s, provider %s", sessionID, provider)
	return nil
}

func GetToken(sessionID string, provider string) (*oauth2.Token, bool) {
	// Construct the Redis key
	key := "session:" + sessionID + "_" + provider

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
