package models

import "time"

type User struct {
	Email       string    `bson:"email"`
	Name        string    `bson:"name"`
	Avatar      string    `bson:"avatar"`
	SupabaseID  string    `bson:"supabaseId"`
	CreatedAt   time.Time `bson:"createdAt"`
	LastLoginAt time.Time `bson:"lastLoginAt"`
}
