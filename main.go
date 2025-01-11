package main

import (
	"auth-service/server"
	"log"
	"net/http"
)

func main() {
	log.Println("Starting Auth Service...")
	http.ListenAndServe(":8080", server.InitializeServer())
}
