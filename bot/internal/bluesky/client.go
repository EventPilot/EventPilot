package bluesky

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultAPIURL = "https://bsky.social/xrpc"
)

// Client handles communication with Bluesky API
type Client struct {
	apiURL     string
	handle     string
	password   string
	session    *Session
	httpClient *http.Client
}

// Config holds Bluesky configuration
type Config struct {
	Handle   string // your.handle.bsky.social
	Password string // App password (not your main password!)
}

// Session holds authentication info
type Session struct {
	AccessJwt  string `json:"accessJwt"`
	RefreshJwt string `json:"refreshJwt"`
	Handle     string `json:"handle"`
	DID        string `json:"did"`
}

// Post represents a Bluesky post
type Post struct {
	URI  string
	CID  string
	Text string
}

// NewClient creates a new Bluesky client
func NewClient(config Config) (*Client, error) {
	if config.Handle == "" || config.Password == "" {
		return nil, fmt.Errorf("handle and password are required")
	}

	client := &Client{
		apiURL:     DefaultAPIURL,
		handle:     config.Handle,
		password:   config.Password,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	// Authenticate immediately
	if err := client.authenticate(); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	return client, nil
}

// authenticate creates a session with Bluesky
func (c *Client) authenticate() error {
	type authRequest struct {
		Identifier string `json:"identifier"`
		Password   string `json:"password"`
	}

	reqBody := authRequest{
		Identifier: c.handle,
		Password:   c.password,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.apiURL+"/com.atproto.server.createSession", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed (status %d): %s", resp.StatusCode, string(body))
	}

	var session Session
	if err := json.Unmarshal(body, &session); err != nil {
		return fmt.Errorf("failed to parse session: %w", err)
	}

	c.session = &session
	return nil
}

// PostText posts a text post to Bluesky
func (c *Client) PostText(text string) (*Post, error) {
	if c.session == nil {
		return nil, fmt.Errorf("not authenticated")
	}

	// Bluesky has a 300 character limit (longer than Twitter!)
	if len(text) > 300 {
		return nil, fmt.Errorf("post exceeds 300 characters (length: %d)", len(text))
	}

	type recordPost struct {
		Text      string `json:"text"`
		CreatedAt string `json:"createdAt"`
		Type      string `json:"$type"`
	}

	type createRecordRequest struct {
		Repo       string     `json:"repo"`
		Collection string     `json:"collection"`
		Record     recordPost `json:"record"`
	}

	now := time.Now().UTC().Format(time.RFC3339)

	reqBody := createRecordRequest{
		Repo:       c.session.DID,
		Collection: "app.bsky.feed.post",
		Record: recordPost{
			Text:      text,
			CreatedAt: now,
			Type:      "app.bsky.feed.post",
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.apiURL+"/com.atproto.repo.createRecord", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.session.AccessJwt)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create post (status %d): %s", resp.StatusCode, string(body))
	}

	type createRecordResponse struct {
		URI string `json:"uri"`
		CID string `json:"cid"`
	}

	var postResp createRecordResponse
	if err := json.Unmarshal(body, &postResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &Post{
		URI:  postResp.URI,
		CID:  postResp.CID,
		Text: text,
	}, nil
}

// PostThread posts a thread of posts to Bluesky
func (c *Client) PostThread(posts []string) ([]*Post, error) {
	if len(posts) == 0 {
		return nil, fmt.Errorf("no posts to publish")
	}

	postedPosts := make([]*Post, 0, len(posts))
	var previousPost *Post

	for i, text := range posts {
		if len(text) > 300 {
			return postedPosts, fmt.Errorf("post %d exceeds 300 characters (length: %d)", i+1, len(text))
		}

		// For thread posts (except the first), we need to add reply info
		if previousPost != nil {
			post, err := c.postReply(text, previousPost)
			if err != nil {
				return postedPosts, fmt.Errorf("failed to post reply %d: %w", i+1, err)
			}
			postedPosts = append(postedPosts, post)
			previousPost = post
		} else {
			// First post in thread
			post, err := c.PostText(text)
			if err != nil {
				return postedPosts, fmt.Errorf("failed to post %d: %w", i+1, err)
			}
			postedPosts = append(postedPosts, post)
			previousPost = post
		}
	}

	return postedPosts, nil
}

// postReply posts a reply to a previous post (for threading)
func (c *Client) postReply(text string, parent *Post) (*Post, error) {
	if c.session == nil {
		return nil, fmt.Errorf("not authenticated")
	}

	type replyRef struct {
		URI string `json:"uri"`
		CID string `json:"cid"`
	}

	type reply struct {
		Root   replyRef `json:"root"`
		Parent replyRef `json:"parent"`
	}

	type recordPost struct {
		Text      string `json:"text"`
		CreatedAt string `json:"createdAt"`
		Type      string `json:"$type"`
		Reply     *reply `json:"reply,omitempty"`
	}

	type createRecordRequest struct {
		Repo       string     `json:"repo"`
		Collection string     `json:"collection"`
		Record     recordPost `json:"record"`
	}

	now := time.Now().UTC().Format(time.RFC3339)

	// For a thread, root is the first post, parent is the immediate previous post
	reqBody := createRecordRequest{
		Repo:       c.session.DID,
		Collection: "app.bsky.feed.post",
		Record: recordPost{
			Text:      text,
			CreatedAt: now,
			Type:      "app.bsky.feed.post",
			Reply: &reply{
				Root:   replyRef{URI: parent.URI, CID: parent.CID},
				Parent: replyRef{URI: parent.URI, CID: parent.CID},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.apiURL+"/com.atproto.repo.createRecord", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.session.AccessJwt)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create reply (status %d): %s", resp.StatusCode, string(body))
	}

	type createRecordResponse struct {
		URI string `json:"uri"`
		CID string `json:"cid"`
	}

	var postResp createRecordResponse
	if err := json.Unmarshal(body, &postResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &Post{
		URI:  postResp.URI,
		CID:  postResp.CID,
		Text: text,
	}, nil
}

// GetUserHandle returns the authenticated user's handle
func (c *Client) GetUserHandle() string {
	if c.session == nil {
		return ""
	}
	return c.session.Handle
}

// GetPostURL converts a post URI to a web URL
func (c *Client) GetPostURL(post *Post) string {
	if post == nil {
		return ""
	}

	// Extract the post ID from the URI
	// URI format: at://did:plc:xxxxx/app.bsky.feed.post/xxxxx
	parts := strings.Split(post.URI, "/")
	if len(parts) < 3 {
		return ""
	}

	postID := parts[len(parts)-1]
	handle := c.GetUserHandle()

	return fmt.Sprintf("https://bsky.app/profile/%s/post/%s", handle, postID)
}
