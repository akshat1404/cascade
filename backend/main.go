package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/akshat1404/cascade/backend/handlers"
	"github.com/akshat1404/cascade/backend/middleware"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	jwksURL := fmt.Sprintf("%s/auth/v1/.well-known/jwks.json", os.Getenv("SUPABASE_URL"))
	jwks, err := keyfunc.NewDefault([]string{jwksURL})
	if err != nil {
		log.Fatal("Failed to fetch JWKS:", err)
	}
	log.Println("JWKS loaded")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		log.Fatal("MongoDB connection failed:", err)
	}
	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("MongoDB ping failed:", err)
	}
	log.Println("Connected to MongoDB")

	db := client.Database(os.Getenv("MONGODB_DB"))

	authHandler := handlers.NewAuthHandler(db.Collection("users"), jwks)
	documentHandler := handlers.NewDocumentHandler(db.Collection("documents"))
	aiHandler := handlers.NewAIHandler()
	settingsHandler := handlers.NewSettingsHandler(db.Collection("users"))
	linkedInHandler := handlers.NewLinkedInHandler(db.Collection("users"))

	cors := middleware.CORS
	auth := middleware.Auth(jwks)

	http.HandleFunc("/health", cors(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	http.HandleFunc("/auth/callback", cors(authHandler.Callback))

	// ── Document routes ─────────────────────────────────────────────────────
	http.HandleFunc("/documents/create", cors(auth(documentHandler.Create)))
	http.HandleFunc("/documents", cors(auth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		documentHandler.ListByUser(w, r)
	})))
	http.HandleFunc("/documents/", cors(auth(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			documentHandler.GetByID(w, r)
		case http.MethodPut:
			documentHandler.Update(w, r)
		case http.MethodDelete:
			documentHandler.Delete(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// ── AI route ─────────────────────────────────────────────────────────────
	http.HandleFunc("/ai/process", cors(auth(aiHandler.Process)))

	// ── Settings routes ───────────────────────────────────────────────────────
	http.HandleFunc("/settings/connected-accounts", cors(auth(settingsHandler.GetConnectedAccounts)))

	// Dev.to
	http.HandleFunc("/settings/devto/connect", cors(auth(settingsHandler.ConnectDevTo)))
	http.HandleFunc("/settings/devto/disconnect", cors(auth(settingsHandler.DisconnectDevTo)))

	// Medium
	http.HandleFunc("/settings/medium/connect", cors(auth(settingsHandler.ConnectMedium)))
	http.HandleFunc("/settings/medium/disconnect", cors(auth(settingsHandler.DisconnectMedium)))

	// LinkedIn (disconnect needs auth; initiate + callback are browser redirects)
	http.HandleFunc("/settings/linkedin/disconnect", cors(auth(settingsHandler.DisconnectLinkedIn)))
	http.HandleFunc("/auth/linkedin", cors(linkedInHandler.Initiate))
	http.HandleFunc("/auth/linkedin/callback", cors(linkedInHandler.Callback))

	log.Println("Backend running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
