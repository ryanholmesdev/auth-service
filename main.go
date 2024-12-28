package main

import (
	"auth-service/generated" // Import the generated package
	"auth-service/handlers"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	// Log the detected HOST environment variable
	log.Println("Environment variable HOST:", getHost())

	r := chi.NewRouter()

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
		host := getHost() // Use the detected host
		handler := httpSwagger.Handler(
			httpSwagger.URL("http://" + host + "/swagger/doc.json"),
		)
		handler(w, r)
	})

	// Debug endpoint to check the current host
	r.Get("/debug/host", func(w http.ResponseWriter, r *http.Request) {
		host := getHost()
		log.Println("Current Host:", host)
		w.Write([]byte("Current Host: " + host))
	})

	// Create an instance of your server that implements the ServerInterface
	server := &handlers.Server{}

	// Use the HandlerFromMux function to register routes
	r.Mount("/", generated.HandlerFromMux(server, r))

	// Log the Swagger UI URL
	log.Println("Swagger UI available at http://" + getHost() + "/swagger/")

	// Start the server
	log.Fatal(http.ListenAndServe(":8080", r))
}

// getHost retrieves the HOST environment variable or falls back to localhost
func getHost() string {
	if host := os.Getenv("HOST"); host != "" {
		return host
	}
	return "localhost:8080" // Default to localhost
}
