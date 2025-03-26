package server

import (
	"auth-service/config"
	"auth-service/generated"
	"auth-service/handlers"
	"auth-service/redisclient"
	"context"
	"encoding/json"
	"fmt"
	"github.com/monzo/slog"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
)

func InitializeServer() http.Handler {
	// Load environment variables
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "local"
	}

	envPath := filepath.Join(".env." + appEnv)
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("Warning: No %s file found. Using system environment variables.", envPath)
		slog.Warn(context.Background(), "No .env file found", map[string]interface{}{
			"env_path": envPath,
		})
	}

	// Initialize Redis
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		log.Fatal("REDIS_ADDR environment variable is not set")
		slog.Error(context.Background(), "REDIS_ADDR environment variable is not set", fmt.Errorf("missing environment variable"), nil)
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

	log.Println("Server started successfully")
	slog.Info(context.Background(), "Server started successfully", nil)

	return r
}

func setupSwagger(r *chi.Mux) {
	r.Get("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		swagger, err := generated.GetSwagger()
		if err != nil {
			log.Println("Failed to load OpenAPI spec:", err)
			slog.Error(r.Context(), "Failed to load OpenAPI spec", err, nil)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		jsonBytes, err := json.Marshal(swagger)
		if err != nil {
			log.Println("Failed to marshal OpenAPI spec:", err)
			slog.Error(r.Context(), "Failed to marshal OpenAPI spec", err, nil)
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
