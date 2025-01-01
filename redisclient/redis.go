package redisclient

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
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
			return
		}
		log.Printf("Failed to connect to Redis (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(2 * time.Second) // Wait before retrying
	}

	log.Fatalf("Failed to connect to Redis after %d attempts", maxRetries)
}
