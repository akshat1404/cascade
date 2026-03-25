package handlers

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "strings"
    "time"

    "github.com/MicahParks/keyfunc/v3"
    "github.com/golang-jwt/jwt/v5"
    "go.mongodb.org/mongo-driver/v2/bson"
    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
)

type AuthHandler struct {
    users *mongo.Collection
    jwks  keyfunc.Keyfunc
}

func NewAuthHandler(users *mongo.Collection, jwks keyfunc.Keyfunc) *AuthHandler {
    return &AuthHandler{users: users, jwks: jwks}
}

func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
    authHeader := r.Header.Get("Authorization")
    if authHeader == "" {
        http.Error(w, "Missing token", http.StatusUnauthorized)
        return
    }
    tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

    token, err := jwt.Parse(tokenStr, h.jwks.Keyfunc)
    if err != nil || !token.Valid {
        log.Println("JWT error:", err)
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }

    claims := token.Claims.(jwt.MapClaims)
    supabaseID := claims["sub"].(string)
    email, _ := claims["email"].(string)
    metadata, _ := claims["user_metadata"].(map[string]interface{})
    name, _ := metadata["full_name"].(string)
    avatar, _ := metadata["avatar_url"].(string)

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    filter := bson.M{"supabaseId": supabaseID}
    update := bson.M{
        "$set": bson.M{
            "email":       email,
            "name":        name,
            "avatar":      avatar,
            "lastLoginAt": time.Now(),
        },
        "$setOnInsert": bson.M{
            "supabaseId": supabaseID,
            "createdAt":  time.Now(),
        },
    }

    _, err = h.users.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true))
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status": "success",
        "email":  email,
        "name":   name,
    })
}