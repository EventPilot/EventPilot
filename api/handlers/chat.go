package handlers

import (
	"bytes"
	"encoding/json"
	"eventpilot/api/models"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/supabase-community/supabase-go"
)

type ChatHandler struct {
	SupabaseClient *supabase.Client
}

type Client struct {
	apiKey           string
	apiUrl           string
	model            string
	maxTokens        int
	anthropicVersion string
	httpClient       *http.Client
}

func NewClient() *Client {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable is required")
	}

	return &Client{
		apiKey:           apiKey,
		apiUrl:           "https://api.anthropic.com/v1/messages",
		model:            "claude-haiku-4-5-20251001",
		maxTokens:        1024,
		anthropicVersion: "2023-06-01",
		httpClient:       &http.Client{},
	}
}

func GetChat(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotImplemented) }
func CreateChatMessage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *ChatHandler) RequestInputs(w http.ResponseWriter, r *http.Request) {
	/*
		Take the event id from the request
		Get the event from the database
		Generate chat messages for the event
		Store the chat messages in the database
		Return the chat messages
	*/

	eventId := r.PathValue("id")

	results := []models.EventMembersWithDetails{}
	_, err := h.SupabaseClient.From("event_member").
		Select("*, event(title, description, event_date), user(name)", "", false).
		Eq("event_id", eventId).
		ExecuteTo(&results)

	if err != nil || len(results) == 0 {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Debug: log what Supabase returned (after unmarshal). Remove once done.
	if b, _ := json.MarshalIndent(results[0], "", "  "); len(b) > 0 {
		log.Printf("[RequestInputs] first result from Supabase: %s", string(b))
	}

	systemPrompt := `You are a helpful marketing assistant that initiates chats with event members to collect information and media for the event`
	msg := fmt.Sprintf(
		`Event: %s
		Event Description: %s
		Event Date: %s
		Event Members:
		`, results[0].Event.Title, results[0].Event.Description, results[0].Event.EventDate)

	for _, res := range results {
		msg += fmt.Sprintf(`
		Name: %s, Role: %s
		`, res.User.Name, res.Role)
	}
	_, _ = systemPrompt, msg // reserved for future LLM use

	initialMessage := "Hello! I'm here to help collect information and media for this event. What would you like to share?"
	chatIDs := make([]string, 0, len(results))

	for _, res := range results {
		if res.UserID == "" {
			continue
		}
		rpcBody := map[string]interface{}{
			"p_event_id":        eventId,
			"p_member_user_id":  res.UserID,
			"p_chat_type":       "info_collection",
			"p_initial_message": initialMessage,
		}
		result := h.SupabaseClient.Rpc("create_chat_with_initial_message", "", rpcBody)
		if result == "" {
			http.Error(w, "failed to create chat", http.StatusInternalServerError)
			return
		}
		chatIDs = append(chatIDs, result)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"chat_ids": chatIDs})
}

func (c *Client) SendMessage(messages []Message, systemPrompt string) (*Response, error) {
	req := Request{
		Model:     c.model,
		MaxTokens: c.maxTokens,
		System:    systemPrompt,
		Messages:  messages,
	}

	return c.sendRequest(req)
}

func (c *Client) sendRequest(req Request) (*Response, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.apiUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", c.anthropicVersion)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for error response
	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("API error: %s - %s", errResp.Error.Type, errResp.Error.Message)
	}

	var claudeResp Response
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &claudeResp, nil
}
