package redisclient

import (
	"context"
	"github.com/monzo/slog"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func InitializeRedis(redisAddr string) {
	Client = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Retry connection for a certain duration
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		err := Client.Ping(context.Background()).Err()
		if err == nil {
			log.Println("Connected to Redis successfully.")
			slog.Info(context.Background(), "Connected to Redis successfully", nil)
			return
		}
		log.Printf("Failed to connect to Redis (attempt %d/%d): %v", i+1, maxRetries, err)
		slog.Error(context.Background(), "Failed to connect to Redis", err, map[string]interface{}{
			"attempt": i + 1,
			"max":     maxRetries,
		})
		time.Sleep(2 * time.Second) // Wait before retrying
	}

	log.Fatalf("Failed to connect to Redis after %d attempts", maxRetries)
}
