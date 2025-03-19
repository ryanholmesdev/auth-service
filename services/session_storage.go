package services

import (
	"auth-service/redisclient"
	"context"
	"encoding/json"
	"log"
	"time"
)

type PKCEData struct {
	CodeVerifier string `json:"code_verifier"`
}

// StorePKCEData stores the PKCE data (including the code verifier) in Redis,
// using the state token as the key.
func StorePKCEData(stateToken, codeVerifier string) error {
	key := "pkce:" + stateToken
	data := PKCEData{
		CodeVerifier: codeVerifier,
	}
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = redisclient.Client.Set(context.Background(), key, b, time.Minute*5).Err()
	if err != nil {
		log.Printf("Failed to store PKCE data in Redis: %v", err)
		return err
	}
	return nil
}

// GetCodeVerifier retrieves the code verifier from Redis for the given state token.
func GetCodeVerifier(stateToken string) (string, error) {
	key := "pkce:" + stateToken
	result, err := redisclient.Client.Get(context.Background(), key).Result()
	if err != nil {
		log.Printf("Failed to retrieve PKCE data from Redis: %v", err)
		return "", err
	}
	var data PKCEData
	if err = json.Unmarshal([]byte(result), &data); err != nil {
		return "", err
	}
	return data.CodeVerifier, nil
}

// DeletePKCEData removes the PKCE data from Redis for the given state token.
func DeletePKCEData(stateToken string) error {
	key := "pkce:" + stateToken
	err := redisclient.Client.Del(context.Background(), key).Err()
	if err != nil {
		log.Printf("Failed to delete PKCE data from Redis: %v", err)
		return err
	}
	return nil
}
