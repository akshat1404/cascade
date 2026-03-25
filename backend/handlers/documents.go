package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/akshat1404/collaborent/backend/middleware"
	"github.com/akshat1404/collaborent/backend/models"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type DocumentHandler struct {
	documents *mongo.Collection
}

func NewDocumentHandler(documents *mongo.Collection) *DocumentHandler {
	return &DocumentHandler{documents: documents}
}

func (h *DocumentHandler) Create(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	doc := models.Document{
		Title: "Untitled Document",
		Content: map[string]interface{}{
			"type": "doc",
			"content": []map[string]interface{}{
				{"type": "paragraph"},
			},
		},
		OwnerID:   ownerID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := h.documents.InsertOne(ctx, doc)
	if err != nil {
		http.Error(w, "Failed to create document", http.StatusInternalServerError)
		return
	}

	doc.ID = result.InsertedID.(interface{})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(doc)
}
