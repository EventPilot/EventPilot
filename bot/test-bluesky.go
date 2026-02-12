package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	fmt.Println("🦋 Bluesky Test Script")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println()

	// Get credentials from environment
	handle := strings.TrimSpace(os.Getenv("BSKY_HANDLE"))
	password := strings.TrimSpace(os.Getenv("BSKY_PASSWORD"))

	if handle == "" {
		fmt.Println("❌ BSKY_HANDLE not set")
		fmt.Println("Run: export $(cat .env | xargs)")
		os.Exit(1)
	}

	if password == "" {
		fmt.Println("❌ BSKY_PASSWORD not set")
		fmt.Println("Run: export $(cat .env | xargs)")
		os.Exit(1)
	}

	fmt.Printf("📝 Handle: %s\n", handle)
	fmt.Printf("🔑 Password: %s\n", maskPassword(password))
	fmt.Println()

	// Step 1: Authenticate
	fmt.Println("🔐 Step 1: Authenticating...")
	session, err := authenticate(handle, password)
	if err != nil {
		fmt.Printf("❌ Authentication failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Authenticated as: %s\n", session.Handle)
	fmt.Printf("✅ DID: %s\n", session.DID)
	fmt.Println()

	// Step 2: Create test post
	fmt.Println("📤 Step 2: Creating test post...")
	testMessage := fmt.Sprintf("Test post %s", time.Now().Format("3:04 PM"))
	fmt.Printf("Message: %s\n", testMessage)
	fmt.Println()

	postURI, err := createPost(session, testMessage)
	if err != nil {
		fmt.Printf("❌ Failed to create post: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Post created successfully!\n")
	fmt.Printf("📍 URI: %s\n", postURI)
	fmt.Println()

	// Extract post ID and create URL
	parts := strings.Split(postURI, "/")
	if len(parts) >= 3 {
		postID := parts[len(parts)-1]
		postURL := fmt.Sprintf("https://bsky.app/profile/%s/post/%s", handle, postID)
		fmt.Printf("🔗 View at: %s\n", postURL)
	}

	fmt.Println()
	fmt.Println("🎉 Test completed successfully!")
	fmt.Println()
	fmt.Println("Your Bluesky credentials are working correctly!")
}

type Session struct {
	AccessJwt  string `json:"accessJwt"`
	RefreshJwt string `json:"refreshJwt"`
	Handle     string `json:"handle"`
	DID        string `json:"did"`
}

func authenticate(handle, password string) (*Session, error) {
	type AuthRequest struct {
		Identifier string `json:"identifier"`
		Password   string `json:"password"`
	}

	reqBody := AuthRequest{
		Identifier: handle,
		Password:   password,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

	resp, err := http.Post(
		"https://bsky.social/xrpc/com.atproto.server.createSession",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth failed (status %d): %s", resp.StatusCode, string(body))
	}

	var session Session
	if err := json.Unmarshal(body, &session); err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	return &session, nil
}

func createPost(session *Session, text string) (string, error) {
	type Post struct {
		Text      string `json:"text"`
		CreatedAt string `json:"createdAt"`
		Type      string `json:"$type"`
	}

	type CreateRecordRequest struct {
		Repo       string `json:"repo"`
		Collection string `json:"collection"`
		Record     Post   `json:"record"`
	}

	now := time.Now().UTC().Format(time.RFC3339)

	reqBody := CreateRecordRequest{
		Repo:       session.DID,
		Collection: "app.bsky.feed.post",
		Record: Post{
			Text:      text,
			CreatedAt: now,
			Type:      "app.bsky.feed.post",
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal error: %w", err)
	}

	req, err := http.NewRequest(
		"POST",
		"https://bsky.social/xrpc/com.atproto.repo.createRecord",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", fmt.Errorf("request creation error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+session.AccessJwt)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("post failed (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		URI string `json:"uri"`
		CID string `json:"cid"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse error: %w", err)
	}

	return result.URI, nil
}

func maskPassword(password string) string {
	if len(password) <= 8 {
		return "****"
	}
	return password[:4] + "****" + password[len(password)-4:]
}
