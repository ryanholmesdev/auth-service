package services

import (
	"auth-service/redisclient"
	"auth-service/services"
	"auth-service/tests"
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func setupTestRedis(t *testing.T) (*redis.Client, func()) {
	client, cleanup := tests.StartRedisTestContainer(t)
	redisclient.Client = client
	return client, cleanup
}

// Test storing a state token
func TestStoreStateToken_Isolated(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	stateToken := "isolated-store-token"

	err := services.StoreStateToken(stateToken)
	assert.NoError(t, err, "Should store state token without error")

	key := services.ConstructSessionKey(stateToken)
	val, err := client.Get(context.Background(), key).Result()
	assert.NoError(t, err)
	assert.Equal(t, "valid", val, "Stored token should be 'valid'")
}

// Test validating a state token
func TestValidateStateToken_Isolated(t *testing.T) {
	_, cleanup := setupTestRedis(t)
	defer cleanup()

	stateToken := "isolated-validate-token"

	_ = services.StoreStateToken(stateToken)

	isValid := services.ValidateStateToken(stateToken)
	assert.True(t, isValid, "State token should be valid")
}

// Test deleting a state token
func TestDeleteStateToken_Isolated(t *testing.T) {
	_, cleanup := setupTestRedis(t)
	defer cleanup()

	stateToken := "isolated-delete-token"

	_ = services.StoreStateToken(stateToken)

	err := services.DeleteStateToken(stateToken)
	assert.NoError(t, err)

	isValid := services.ValidateStateToken(stateToken)
	assert.False(t, isValid, "Deleted token should not be valid")
}

// Test expired state token
func TestValidateStateToken_Expires_Isolated(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	stateToken := "isolated-expire-token"

	// Store with short TTL
	key := services.ConstructSessionKey(stateToken)
	err := client.Set(context.Background(), key, "valid", 2*time.Second).Err()
	assert.NoError(t, err)

	time.Sleep(3 * time.Second)

	isValid := services.ValidateStateToken(stateToken)
	assert.False(t, isValid, "Expired state token should not be valid")
}
