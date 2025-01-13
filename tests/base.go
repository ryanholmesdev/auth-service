package tests

import (
	"auth-service/server"
	"context"
	_ "github.com/flashlabs/rootpath" // Set's the directory to the root to load the .envs
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"net/http/httptest"
	"os"
	"testing"
)

type TestSetup struct {
	RedisContainer testcontainers.Container
	Server         *httptest.Server
	Cleanup        func()
}

func InitializeTestEnvironment(t *testing.T) *TestSetup {
	ctx := context.Background()

	// Start Redis Testcontainer
	redisReq := testcontainers.ContainerRequest{
		Image:        "redis:7",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp"),
	}
	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: redisReq,
		Started:          true,
	})
	assert.NoError(t, err)

	host, err := redisContainer.Host(ctx)
	assert.NoError(t, err)

	port, err := redisContainer.MappedPort(ctx, "6379")
	assert.NoError(t, err)

	redisAddr := host + ":" + port.Port()

	// Inject Redis address into environment
	os.Setenv("REDIS_ADDR", redisAddr)
	os.Setenv("APP_ENV", "test")

	// Start server with httptest
	testServer := httptest.NewServer(server.InitializeServer())

	// Cleanup function
	cleanup := func() {
		_ = redisContainer.Terminate(ctx)
		testServer.Close()
	}

	return &TestSetup{
		RedisContainer: redisContainer,
		Server:         testServer,
		Cleanup:        cleanup,
	}
}
