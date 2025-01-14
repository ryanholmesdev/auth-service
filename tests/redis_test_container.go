package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func StartRedisTestContainer(t *testing.T) (redisClient *redis.Client, terminate func()) {
	t.Helper()

	ctx := context.Background()

	// Start Redis container
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "redis:latest",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForLog("Ready to accept connections").WithStartupTimeout(10 * time.Second),
		},
		Started: true,
	})
	require.NoError(t, err, "Failed to start Redis container")

	// Get dynamic port
	port, err := container.MappedPort(ctx, "6379")
	require.NoError(t, err)

	// Connect Redis client
	addr := fmt.Sprintf("localhost:%s", port.Port())
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	// Verify Redis is ready
	_, err = client.Ping(ctx).Result()
	require.NoError(t, err, "Redis is not ready")

	// Cleanup function
	cleanup := func() {
		_ = container.Terminate(ctx)
	}

	return client, cleanup
}
