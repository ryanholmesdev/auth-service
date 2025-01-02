package main

import (
	"auth-service/config"
	"auth-service/generated"
	"auth-service/handlers"
	"auth-service/redisclient"
	"encoding/json"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {

	// Determine the environment
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "local" // Default to "local" if APP_ENV is not set
	}

	// Load environment variables from the appropriate file
	envFile := ".env." + appEnv
	err := godotenv.Load(envFile)
	if err != nil {
		log.Fatalf("Error loading %s file: %v", envFile, err)
	}

	log.Printf("Environment: %s (loaded %s)", appEnv, envFile)

	// Get Redis address from environment variable
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		log.Fatal("REDIS_ADDR environment variable is not set")
	}

	config.InitConfig()

	redisclient.InitializeRedis(redisAddr)

	r := chi.NewRouter()

	// Enable CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Setup Swagger
	setupSwagger(r)

	// Create an instance of server that implements the ServerInterface
	server := &handlers.Server{}

	// Use the HandlerFromMux function to register routes
	r.Mount("/", generated.HandlerFromMux(server, r))

	log.Println("Swagger UI available at http://localhost:8080/swagger/")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func setupSwagger(r *chi.Mux) {
	// Serve the OpenAPI document
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

	// Serve the Swagger UI
	r.Get("/swagger/*", func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		swaggerURL := "http://" + host + "/swagger/doc.json"

		httpSwagger.Handler(
			httpSwagger.URL(swaggerURL),
		)(w, r)
	})
}
