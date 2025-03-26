package main

import (
	"auth-service/server"
	"context"
	"github.com/monzo/slog"
	"log"
	"net/http"
)

func main() {
	log.Println("Starting Auth Service...")

	slog.Info(context.Background(), "Starting Auth Service", map[string]interface{}{
		"port": "8080",
	})

	http.ListenAndServe(":8080", server.InitializeServer())
}
