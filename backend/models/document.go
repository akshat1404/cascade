package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Document struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Title     string        `bson:"title"         json:"title"`
	Content   interface{}   `bson:"content"       json:"content"`
	OwnerID   string        `bson:"ownerId"       json:"owner_id"`
	CreatedAt time.Time     `bson:"createdAt"     json:"created_at"`
	UpdatedAt time.Time     `bson:"updatedAt"     json:"updated_at"`
}
