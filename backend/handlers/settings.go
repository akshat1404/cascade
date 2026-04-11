package handlers

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/akshat1404/cascade/backend/middleware"
	"github.com/akshat1404/cascade/backend/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ─── AES-256-GCM helpers ────────────────────────────────────────────────────

func encrypt(plaintext string) (string, error) {
	key := []byte(os.Getenv("ENCRYPTION_KEY")) // must be 32 bytes
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decrypt(encoded string) (string, error) {
	key := []byte(os.Getenv("ENCRYPTION_KEY"))
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", err
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// ─── Handler struct ─────────────────────────────────────────────────────────

type SettingsHandler struct {
	users *mongo.Collection
}

func NewSettingsHandler(users *mongo.Collection) *SettingsHandler {
	return &SettingsHandler{users: users}
}

// ─── GET /settings/connected-accounts ────────────────────────────────────────

func (h *SettingsHandler) GetConnectedAccounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	supabaseID := r.Context().Value(middleware.UserIDKey).(string)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err := h.users.FindOne(ctx, bson.M{"supabaseId": supabaseID}).Decode(&user)
	if err != nil {
		// No user doc yet — return all disconnected
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"linkedin": map[string]interface{}{"connected": false},
			"devto":    map[string]interface{}{"connected": false},
			"medium":   map[string]interface{}{"connected": false},
		})
		return
	}

	resp := map[string]interface{}{
		"linkedin": map[string]interface{}{"connected": false},
		"devto":    map[string]interface{}{"connected": false},
		"medium":   map[string]interface{}{"connected": false},
	}

	if user.ConnectedAccounts.LinkedIn != nil {
		resp["linkedin"] = map[string]interface{}{
			"connected": true,
			"name":      user.ConnectedAccounts.LinkedIn.ProfileName,
		}
	}
	if user.ConnectedAccounts.DevTo != nil {
		resp["devto"] = map[string]interface{}{
			"connected": true,
			"username":  user.ConnectedAccounts.DevTo.Username,
		}
	}
	if user.ConnectedAccounts.Medium != nil {
		resp["medium"] = map[string]interface{}{
			"connected": true,
			"username":  user.ConnectedAccounts.Medium.Username,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ─── POST /settings/devto/connect ────────────────────────────────────────────

func (h *SettingsHandler) ConnectDevTo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	supabaseID := r.Context().Value(middleware.UserIDKey).(string)

	var body struct {
		APIKey string `json:"apiKey"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.APIKey == "" {
		http.Error(w, "Missing apiKey", http.StatusBadRequest)
		return
	}

	// Verify the key
	req, _ := http.NewRequest(http.MethodGet, "https://dev.to/api/users/me", nil)
	req.Header.Set("api-key", body.APIKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid API key"})
		return
	}
	defer resp.Body.Close()

	var devtoUser struct {
		Username string `json:"username"`
	}
	json.NewDecoder(resp.Body).Decode(&devtoUser)

	// Encrypt
	encrypted, err := encrypt(body.APIKey)
	if err != nil {
		http.Error(w, "Encryption error", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"supabaseId": supabaseID}
	update := bson.M{
		"$set": bson.M{
			"connectedAccounts.devto": bson.M{
				"apiKey":      encrypted,
				"username":    devtoUser.Username,
				"connectedAt": time.Now(),
			},
		},
		"$setOnInsert": bson.M{"supabaseId": supabaseID},
	}
	_, err = h.users.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true))
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"username": devtoUser.Username})
}

// ─── DELETE /settings/devto/disconnect ───────────────────────────────────────

func (h *SettingsHandler) DisconnectDevTo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	supabaseID := r.Context().Value(middleware.UserIDKey).(string)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"supabaseId": supabaseID}
	update := bson.M{"$unset": bson.M{"connectedAccounts.devto": ""}}
	h.users.UpdateOne(ctx, filter, update)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "disconnected"})
}

// ─── POST /settings/medium/connect ───────────────────────────────────────────

func (h *SettingsHandler) ConnectMedium(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	supabaseID := r.Context().Value(middleware.UserIDKey).(string)

	var body struct {
		IntegrationToken string `json:"integrationToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.IntegrationToken == "" {
		http.Error(w, "Missing integrationToken", http.StatusBadRequest)
		return
	}

	// Verify the token
	req, _ := http.NewRequest(http.MethodGet, "https://api.medium.com/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+body.IntegrationToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid integration token"})
		return
	}
	defer resp.Body.Close()

	var mediumResp struct {
		Data struct {
			ID       string `json:"id"`
			Username string `json:"username"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&mediumResp)

	encrypted, err := encrypt(body.IntegrationToken)
	if err != nil {
		http.Error(w, "Encryption error", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"supabaseId": supabaseID}
	update := bson.M{
		"$set": bson.M{
			"connectedAccounts.medium": bson.M{
				"integrationToken": encrypted,
				"authorId":         mediumResp.Data.ID,
				"username":         mediumResp.Data.Username,
				"connectedAt":      time.Now(),
			},
		},
		"$setOnInsert": bson.M{"supabaseId": supabaseID},
	}
	_, err = h.users.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true))
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"username": mediumResp.Data.Username,
		"authorId": mediumResp.Data.ID,
	})
}

// ─── DELETE /settings/medium/disconnect ──────────────────────────────────────

func (h *SettingsHandler) DisconnectMedium(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	supabaseID := r.Context().Value(middleware.UserIDKey).(string)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"supabaseId": supabaseID}
	update := bson.M{"$unset": bson.M{"connectedAccounts.medium": ""}}
	h.users.UpdateOne(ctx, filter, update)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "disconnected"})
}

// ─── DELETE /settings/linkedin/disconnect ────────────────────────────────────

func (h *SettingsHandler) DisconnectLinkedIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	supabaseID := r.Context().Value(middleware.UserIDKey).(string)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"supabaseId": supabaseID}
	update := bson.M{"$unset": bson.M{"connectedAccounts.linkedin": ""}}
	h.users.UpdateOne(ctx, filter, update)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "disconnected"})
}
