package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func signInRedirectionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"Status": "Success", "Message": "You have been signed in successfully"})
}

func main() {
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/auth/callback", signInRedirectionHandler)
	log.Println("Backend running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
