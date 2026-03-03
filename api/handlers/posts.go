package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/supabase-community/supabase-go"
)

type EventDetails struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	EventDate   string `json:"event_date"`
	CreatedAt   string `json:"created_at"`
	Location    string `json:"location"`
	Status      string `json:"status"`
}

type MediaItem struct {
	ID          string `json:"id"`
	EventID     string `json:"event_id"`
	UploadedBy  string `json:"uploaded_by"`
	MediaType   string `json:"media_type"`
	StoragePath string `json:"storage_path"`
	Metadata    string `json:"metadata"`
	CreatedAt   string `json:"created_at"`
}

type Post struct {
	ID        string `json:"id"`
	EventID   string `json:"event_id"`
	Content   string `json:"content"`
	Status    string `json:"status"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at,omitempty"`
}

// ─── Bluesky AT Protocol structs ─────────────────────────────────────────────

type bskySession struct {
	AccessJwt string `json:"accessJwt"`
	DID       string `json:"did"`
}

type bskyPostRecord struct {
	Type      string    `json:"$type"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
}

type bskyCreateRecordRequest struct {
	Repo       string         `json:"repo"`
	Collection string         `json:"collection"`
	Record     bskyPostRecord `json:"record"`
}

type bskyCreateRecordResponse struct {
	CID string `json:"cid"`
	URI string `json:"uri"`
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func newSupabaseClient() (*supabase.Client, error) {
	url := os.Getenv("SUPABASE_URL")
	key := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if url == "" || key == "" {
		return nil, fmt.Errorf("SUPABASE_URL and SUPABASE_SERVICE_ROLE_KEY are required")
	}
	return supabase.NewClient(url, key, nil)
}

// generatePostContent calls Claude to draft a Bluesky post from event context.
func generatePostContent(event EventDetails, media []MediaItem) (string, error) {
	client := NewClient()

	systemPrompt := `You are a social media copywriter for an engineering student organisation.
Write a short, engaging Bluesky post (≤300 characters) celebrating a completed event.
Use an enthusiastic but professional tone. Include 1–3 relevant hashtags at the end.
Return ONLY the post text — no commentary, no quotes, no markdown.`

	userMsg := fmt.Sprintf(
		"Event: %s\nDate: %s\nDescription: %s",
		event.Title, event.EventDate, event.Description,
	)

	resp, err := client.SendMessage([]Message{{Role: "user", Content: userMsg}}, systemPrompt)
	if err != nil {
		return "", fmt.Errorf("Claude API error: %w", err)
	}
	if len(resp.Content) == 0 {
		return "", fmt.Errorf("empty response from Claude")
	}
	return resp.Content[0].Text, nil
}

// authenticateBluesky creates a Bluesky session and returns the access JWT + DID.
func authenticateBluesky() (*bskySession, error) {
	handle := os.Getenv("BSKY_HANDLE")
	password := os.Getenv("BSKY_PASSWORD")
	if handle == "" || password == "" {
		return nil, fmt.Errorf("BSKY_HANDLE and BSKY_PASSWORD env vars are required")
	}

	body, _ := json.Marshal(map[string]string{
		"identifier": handle,
		"password":   password,
	})

	resp, err := http.Post(
		"https://bsky.social/xrpc/com.atproto.server.createSession",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, fmt.Errorf("bluesky auth request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bluesky auth error (%d): %s", resp.StatusCode, string(raw))
	}

	var session bskySession
	if err := json.Unmarshal(raw, &session); err != nil {
		return nil, fmt.Errorf("failed to parse bluesky session: %w", err)
	}
	return &session, nil
}

// publishToBluesky posts content to Bluesky and returns the public post URL.
func publishToBluesky(content string) (string, error) {
	session, err := authenticateBluesky()
	if err != nil {
		return "", err
	}

	payload := bskyCreateRecordRequest{
		Repo:       session.DID,
		Collection: "app.bsky.feed.post",
		Record: bskyPostRecord{
			Type:      "app.bsky.feed.post",
			Text:      content,
			CreatedAt: time.Now().UTC(),
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(
		"POST",
		"https://bsky.social/xrpc/com.atproto.repo.createRecord",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+session.AccessJwt)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("bluesky post request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bluesky post error (%d): %s", resp.StatusCode, string(raw))
	}

	var result bskyCreateRecordResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("failed to parse bluesky response: %w", err)
	}

	// Convert AT URI (at://did:plc:xxx/app.bsky.feed.post/rkey) → public HTTPS URL
	// Format: https://bsky.app/profile/<handle>/post/<rkey>
	handle := os.Getenv("BLUESKY_HANDLE")
	rkey := result.URI[len(result.URI)-13:] // last segment after final "/"
	for i := len(result.URI) - 1; i >= 0; i-- {
		if result.URI[i] == '/' {
			rkey = result.URI[i+1:]
			break
		}
	}
	postURL := fmt.Sprintf("https://bsky.app/profile/%s/post/%s", handle, rkey)
	return postURL, nil
}

// ─── Handlers ────────────────────────────────────────────────────────────────

// GeneratePost uses Claude to draft a social media post for a completed event
func GeneratePost(w http.ResponseWriter, r *http.Request) {
	eventID := r.PathValue("id")

	sb, err := newSupabaseClient()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch event details
	var events []EventDetails
	_, err = sb.From("event").
		Select("id, title, description, event_date, created_at, location, status", "", false).
		Eq("id", eventID).
		ExecuteTo(&events)
	log.Printf("[GeneratePost] eventID=%v err=%v eventsFound=%d", eventID, err, len(events))
	if err != nil {
		http.Error(w, "supabase error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if len(events) == 0 {
		http.Error(w, "event not found", http.StatusNotFound)
		return
	}
	event := events[0]

	// 2. Fetch associated media captions (best-effort)
	var media []MediaItem
	_, _ = sb.From("media").
		Select("id, event_id, uploaded_by, media_type, storage_path, metadata, created_at", "", false).
		Eq("event_id", eventID).
		ExecuteTo(&media)

	// 3. Ask Claude to write the post
	content, err := generatePostContent(event, media)
	if err != nil {
		log.Printf("[GeneratePost] Claude error: %v", err)
		http.Error(w, "failed to generate post content", http.StatusInternalServerError)
		return
	}

	// 4. Upsert draft post into `post` table
	//    Assumes unique constraint on event_id so re-generating overwrites the draft.
	upsertData := map[string]interface{}{
		"event_id": eventID,
		"content":  content,
		"status":   "draft",
	}
	var posts []Post
	_, err = sb.From("post").
		Upsert(upsertData, "event_id", "", "").
		ExecuteTo(&posts)
	if err != nil {
		log.Printf("[GeneratePost] Supabase upsert error: %v", err)
		http.Error(w, "failed to save post", http.StatusInternalServerError)
		return
	}

	var post Post
	if len(posts) > 0 {
		post = posts[0]
	} else {
		post = Post{EventID: eventID, Content: content, Status: "draft", URL: ""}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

// GetPost retrieves the generated (draft or published) post for an event.
//
// GET /api/events/{id}/post
func GetPost(w http.ResponseWriter, r *http.Request) {
	eventID := r.PathValue("id")

	sb, err := newSupabaseClient()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var posts []Post
	_, err = sb.From("post").
		Select("*", "", false).
		Eq("event_id", eventID).
		ExecuteTo(&posts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(posts) == 0 {
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts[0])
}

// PublishPost sends the draft post to Bluesky and marks it as published.
//
// POST /api/events/{id}/post/publish
func PublishPost(w http.ResponseWriter, r *http.Request) {
	eventID := r.PathValue("id")

	sb, err := newSupabaseClient()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 1. Load the draft post
	var posts []Post
	_, err = sb.From("post").
		Select("*", "", false).
		Eq("event_id", eventID).
		ExecuteTo(&posts)
	if err != nil || len(posts) == 0 {
		http.Error(w, "post not found — generate one first", http.StatusNotFound)
		return
	}
	post := posts[0]

	if post.Status == "published" {
		http.Error(w, "post already published", http.StatusConflict)
		return
	}

	// 2. Push to Bluesky
	postURL, err := publishToBluesky(post.Content)
	if err != nil {
		log.Printf("[PublishPost] Bluesky error: %v", err)
		http.Error(w, "failed to publish to Bluesky: "+err.Error(), http.StatusBadGateway)
		return
	}

	// 3. Mark as published in Supabase
	updateData := map[string]interface{}{
		"status": "published",
		"url":    postURL,
	}
	var updated []Post
	_, err = sb.From("post").
		Update(updateData, "", "").
		Eq("event_id", eventID).
		ExecuteTo(&updated)
	if err != nil {
		// Post is live on Bluesky even if DB update fails — log but don't error out.
		log.Printf("[PublishPost] WARNING: post published to Bluesky but DB update failed: %v", err)
	}

	result := post
	result.Status = "published"
	result.URL = postURL

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
