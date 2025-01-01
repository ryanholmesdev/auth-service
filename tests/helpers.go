package tests

import (
	"context"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"testing"
)

func SetupRedisContainer(t *testing.T) (string, func()) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "redis:7",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp"),
	}
	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start Redis container: %v", err)
	}

	// Get the Redis address
	host, _ := redisContainer.Host(ctx)
	port, _ := redisContainer.MappedPort(ctx, "6379/tcp")

	redisAddr := host + ":" + port.Port()
	log.Printf("Redis running at %s", redisAddr)

	// Teardown function to stop the container
	teardown := func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Fatalf("Failed to terminate Redis container: %v", err)
		}
	}

	return redisAddr, teardown
}
