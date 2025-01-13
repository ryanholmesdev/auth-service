package server

import (
	"auth-service/config"
	"auth-service/generated"
	"auth-service/handlers"
	"auth-service/redisclient"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// ✅ InitializeServer returns a fully configured router
func InitializeServer() http.Handler {
	// Load environment variables
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "local"
	}

	envPath := filepath.Join(".env." + appEnv)
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("Warning: No %s file found. Using system environment variables.", envPath)
	}

	// Initialize Redis
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		log.Fatal("REDIS_ADDR environment variable is not set")
	}

	redisclient.InitializeRedis(redisAddr)

	config.InitConfig()

	// Setup Router
	r := chi.NewRouter()

	// CORS Middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Swagger setup
	setupSwagger(r)

	// Register Handlers
	server := &handlers.Server{}
	r.Mount("/", generated.HandlerFromMux(server, r))

	return r
}

// ✅ Swagger setup function
func setupSwagger(r *chi.Mux) {
	r.Get("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		swagger, err := generated.GetSwagger()
		if err != nil {
			log.Println("Failed to load OpenAPI spec:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		jsonBytes, err := json.Marshal(swagger)
		if err != nil {
			log.Println("Failed to marshal OpenAPI spec:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))
}
