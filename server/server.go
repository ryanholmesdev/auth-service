package server

import (
	"auth-service/config"
	"auth-service/generated"
	"auth-service/handlers"
	"auth-service/redisclient"
	"auth-service/utils"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
)

func InitializeServer() http.Handler {
	ctx := context.Background()

	// Load environment variables
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "local"
	}

	envPath := filepath.Join(".env." + appEnv)
	if err := godotenv.Load(envPath); err != nil {
		utils.LogWarn(ctx, "No environment file found", map[string]interface{}{
			"env_path": envPath,
			"error":    err.Error(),
		})
	}

	// Initialize Redis
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		utils.LogError(ctx, "Redis address not configured", nil, map[string]interface{}{
			"error": "REDIS_ADDR environment variable is not set",
		})
		os.Exit(1)
	}

	redisclient.InitializeRedis(redisAddr)

	config.InitConfig()

	// Setup Router
	r := chi.NewRouter()

	// CORS Middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Swagger setup
	setupSwagger(r)

	// Register Handlers
	server := &handlers.Server{}
	r.Mount("/", generated.HandlerFromMux(server, r))

	utils.LogInfo(ctx, "Server started successfully", map[string]interface{}{
		"env": appEnv,
	})

	return r
}

func setupSwagger(r *chi.Mux) {
	ctx := context.Background()

	r.Get("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		swagger, err := generated.GetSwagger()
		if err != nil {
			utils.LogError(ctx, "Failed to load OpenAPI spec", err, nil)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		jsonBytes, err := json.Marshal(swagger)
		if err != nil {
			utils.LogError(ctx, "Failed to marshal OpenAPI spec", err, nil)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))
}
