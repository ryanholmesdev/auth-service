package main

import (
	"auth-service/generated" // Import the generated package
	"auth-service/handlers"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
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
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	// Create an instance of your server that implements the ServerInterface
	server := &handlers.Server{}

	// Use the HandlerFromMux function to register routes
	r.Mount("/", generated.HandlerFromMux(server, r))

	log.Println("Swagger UI available at http://localhost:8080/swagger/")
	log.Fatal(http.ListenAndServe(":8080", r))
}
