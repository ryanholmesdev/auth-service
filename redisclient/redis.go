package redisclient

import (
	"context"
	"fmt"
	"time"

	"github.com/monzo/slog"
	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func InitializeRedis(redisAddr string) {
	ctx := context.Background()
	Client = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Retry connection for a certain duration
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		err := Client.Ping(ctx).Err()
		if err == nil {
			slog.Info(ctx, "Connected to Redis successfully", map[string]interface{}{
				"redis_addr": redisAddr,
			})
			return
		}
		slog.Warn(ctx, "Failed to connect to Redis", err, map[string]interface{}{
			"attempt":     i + 1,
			"max_retries": maxRetries,
			"redis_addr":  redisAddr,
		})
		time.Sleep(2 * time.Second) // Wait before retrying
	}

	slog.Error(ctx, "Failed to connect to Redis after maximum retries", fmt.Errorf("connection failed"), map[string]interface{}{
		"max_retries": maxRetries,
		"redis_addr":  redisAddr,
	})
	panic("Failed to connect to Redis after maximum retries")
}
