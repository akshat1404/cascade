package models

import "time"

type LinkedInAccount struct {
	AccessToken string    `bson:"accessToken"`
	ProfileID   string    `bson:"profileId"`
	ProfileName string    `bson:"profileName"`
	ConnectedAt time.Time `bson:"connectedAt"`
}

type DevToAccount struct {
	APIKey      string    `bson:"apiKey"`
	Username    string    `bson:"username"`
	ConnectedAt time.Time `bson:"connectedAt"`
}

type MediumAccount struct {
	IntegrationToken string    `bson:"integrationToken"`
	AuthorID         string    `bson:"authorId"`
	Username         string    `bson:"username"`
	ConnectedAt      time.Time `bson:"connectedAt"`
}

type ConnectedAccounts struct {
	LinkedIn *LinkedInAccount `bson:"linkedin,omitempty"`
	DevTo    *DevToAccount    `bson:"devto,omitempty"`
	Medium   *MediumAccount   `bson:"medium,omitempty"`
}

type User struct {
	Email             string            `bson:"email"`
	Name              string            `bson:"name"`
	Avatar            string            `bson:"avatar"`
	SupabaseID        string            `bson:"supabaseId"`
	CreatedAt         time.Time         `bson:"createdAt"`
	LastLoginAt       time.Time         `bson:"lastLoginAt"`
	ConnectedAccounts ConnectedAccounts `bson:"connectedAccounts,omitempty"`
}
