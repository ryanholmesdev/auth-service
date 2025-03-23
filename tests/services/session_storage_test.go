package services

import (
	"auth-service/redisclient"
	"auth-service/services"
	"auth-service/tests"
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// setupTestRedis initializes a Redis client for testing.
func setupTestRedis(t *testing.T) (*redis.Client, func()) {
	client, cleanup := tests.StartRedisTestContainer(t)
	redisclient.Client = client
	return client, cleanup
}

func TestStorePKCEData_Isolated(t *testing.T) {
	_, cleanup := setupTestRedis(t)
	defer cleanup()

	stateToken := "isolated-store-token"
	codeVerifier := "test-code-verifier"

	// Explicitly call the function from the services package.
	err := services.StorePKCEData(stateToken, codeVerifier)
	assert.NoError(t, err, "Should store PKCE data without error")

	// Retrieve the stored code verifier.
	retrievedVerifier, err := services.GetCodeVerifier(stateToken)
	assert.NoError(t, err)
	assert.Equal(t, codeVerifier, retrievedVerifier, "Stored code verifier should match")
}

func TestDeletePKCEData_Isolated(t *testing.T) {
	_, cleanup := setupTestRedis(t)
	defer cleanup()

	stateToken := "isolated-delete-token"
	codeVerifier := "test-code-verifier"

	// Store the PKCE data.
	err := services.StorePKCEData(stateToken, codeVerifier)
	assert.NoError(t, err)

	// Delete the stored PKCE data.
	err = services.DeletePKCEData(stateToken)
	assert.NoError(t, err)

	// Attempt to retrieve the data after deletion; should result in an error.
	_, err = services.GetCodeVerifier(stateToken)
	assert.Error(t, err, "Expected error when retrieving deleted PKCE data")
}

func TestStorePKCEData_Expires_Isolated(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	stateToken := "isolated-expire-token"
	codeVerifier := "test-code-verifier"
	key := "pkce:" + stateToken

	// Manually marshal the data and set it with a short TTL (2 seconds).
	data := services.PKCEData{CodeVerifier: codeVerifier}
	b, err := json.Marshal(data)
	assert.NoError(t, err)
	err = client.Set(context.Background(), key, b, 2*time.Second).Err()
	assert.NoError(t, err)

	// Wait for the key to expire.
	time.Sleep(3 * time.Second)

	_, err = services.GetCodeVerifier(stateToken)
	assert.Error(t, err, "Expired PKCE data should not be retrievable")
}
