package handlers

import (
	"auth-service/services"
	"encoding/json"
	"net/http"
)

func (s *Server) GetAuthStatus(w http.ResponseWriter, r *http.Request) {
	// Retrieve session ID from the cookie
	sessionCookie, err := r.Cookie("session_id")
	if err != nil || sessionCookie.Value == "" {
		http.Error(w, "Session ID is required", http.StatusUnauthorized)
		return
	}

	connectedProviders, err := services.GetLoggedInProviders(sessionCookie.Value)
	if err != nil {
		http.Error(w, "Unable to get logged in providers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(connectedProviders)
}
