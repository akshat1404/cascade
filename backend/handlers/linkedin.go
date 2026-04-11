package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type LinkedInHandler struct {
	users *mongo.Collection
}

func NewLinkedInHandler(users *mongo.Collection) *LinkedInHandler {
	return &LinkedInHandler{users: users}
}

// GET /auth/linkedin
// Redirects the browser to LinkedIn's OAuth consent screen.
// The state param encodes the supabase user ID from the Authorization header.
func (h *LinkedInHandler) Initiate(w http.ResponseWriter, r *http.Request) {
	// The supabaseID is passed via query param because this is a browser redirect
	// (no Authorization header in that flow). We read it from ?user_id=
	supabaseID := r.URL.Query().Get("user_id")
	if supabaseID == "" {
		http.Error(w, "Missing user_id", http.StatusBadRequest)
		return
	}

	clientID := os.Getenv("LINKEDIN_CLIENT_ID")
	redirectURI := os.Getenv("LINKEDIN_REDIRECT_URI")

	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("state", supabaseID)
	params.Set("scope", "openid profile email w_member_social")

	authURL := "https://www.linkedin.com/oauth/v2/authorization?" + params.Encode()
	http.Redirect(w, r, authURL, http.StatusFound)
}

// GET /auth/linkedin/callback
// Receives the OAuth callback from LinkedIn.
func (h *LinkedInHandler) Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	supabaseID := r.URL.Query().Get("state")
	frontendURL := os.Getenv("FRONTEND_URL")

	if code == "" || supabaseID == "" {
		http.Redirect(w, r, frontendURL+"/dashboard/settings?error=linkedin_failed", http.StatusFound)
		return
	}

	// Exchange code for access token
	clientID := os.Getenv("LINKEDIN_CLIENT_ID")
	clientSecret := os.Getenv("LINKEDIN_CLIENT_SECRET")
	redirectURI := os.Getenv("LINKEDIN_REDIRECT_URI")

	tokenBody := url.Values{}
	tokenBody.Set("grant_type", "authorization_code")
	tokenBody.Set("code", code)
	tokenBody.Set("redirect_uri", redirectURI)
	tokenBody.Set("client_id", clientID)
	tokenBody.Set("client_secret", clientSecret)

	tokenResp, err := http.Post(
		"https://www.linkedin.com/oauth/v2/accessToken",
		"application/x-www-form-urlencoded",
		strings.NewReader(tokenBody.Encode()),
	)
	if err != nil || tokenResp.StatusCode != http.StatusOK {
		http.Redirect(w, r, frontendURL+"/dashboard/settings?error=linkedin_token_failed", http.StatusFound)
		return
	}
	defer tokenResp.Body.Close()

	var tokenData struct {
		AccessToken string `json:"access_token"`
	}
	tokenBytes, _ := io.ReadAll(tokenResp.Body)
	json.Unmarshal(tokenBytes, &tokenData)

	if tokenData.AccessToken == "" {
		http.Redirect(w, r, frontendURL+"/dashboard/settings?error=linkedin_no_token", http.StatusFound)
		return
	}

	// Fetch profile from LinkedIn userinfo endpoint
	profileReq, _ := http.NewRequest(http.MethodGet, "https://api.linkedin.com/v2/userinfo", nil)
	profileReq.Header.Set("Authorization", "Bearer "+tokenData.AccessToken)
	profileResp, err := http.DefaultClient.Do(profileReq)
	if err != nil || profileResp.StatusCode != http.StatusOK {
		http.Redirect(w, r, frontendURL+"/dashboard/settings?error=linkedin_profile_failed", http.StatusFound)
		return
	}
	defer profileResp.Body.Close()

	var profile struct {
		Sub  string `json:"sub"`
		Name string `json:"name"`
	}
	json.NewDecoder(profileResp.Body).Decode(&profile)

	// Encrypt access token
	encrypted, err := encrypt(tokenData.AccessToken)
	if err != nil {
		http.Redirect(w, r, frontendURL+"/dashboard/settings?error=encryption_failed", http.StatusFound)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"supabaseId": supabaseID}
	update := bson.M{
		"$set": bson.M{
			"connectedAccounts.linkedin": bson.M{
				"accessToken": encrypted,
				"profileId":   profile.Sub,
				"profileName": profile.Name,
				"connectedAt": time.Now(),
			},
		},
		"$setOnInsert": bson.M{"supabaseId": supabaseID},
	}
	_, err = h.users.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true))
	if err != nil {
		fmt.Println("DB error:", err)
		http.Redirect(w, r, frontendURL+"/dashboard/settings?error=db_failed", http.StatusFound)
		return
	}

	http.Redirect(w, r, frontendURL+"/dashboard/settings?linkedin=connected", http.StatusFound)
}
