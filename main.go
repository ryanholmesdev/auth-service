package main

import (
	"auth-service/server"
	"context"
	"net/http"

	"github.com/monzo/slog"
)

func main() {
	ctx := context.Background()
	slog.Info(ctx, "Starting Auth Service", map[string]interface{}{
		"port": 8080,
	})
	http.ListenAndServe(":8080", server.InitializeServer())
}
