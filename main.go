package main

import (
	"auth-service/server"
	"auth-service/utils"
	"context"
	"net/http"
)

func main() {
	// Initialize the structured logger
	utils.InitLogger()

	// Create a root context for logging
	ctx := context.Background()
	utils.LogInfo(ctx, "Starting Auth Service...", map[string]interface{}{
		"service": "auth-service",
	})

	http.ListenAndServe(":8080", server.InitializeServer())
}
