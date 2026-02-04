package claude

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	DefaultAPIURL      = "https://api.anthropic.com/v1/messages"
	DefaultModel       = "claude-3-haiku-20240307"
	DefaultMaxTokens   = 1024
	AnthropicVersion   = "2023-06-01"
)

// Client handles communication with the Claude API
type Client struct {
	apiKey     string
	apiURL     string
	model      string
	maxTokens  int
	httpClient *http.Client
}

// NewClient creates a new Claude API client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		apiURL:     DefaultAPIURL,
		model:      DefaultModel,
		maxTokens:  DefaultMaxTokens,
		httpClient: &http.Client{},
	}
}

// WithModel sets a custom model
func (c *Client) WithModel(model string) *Client {
	c.model = model
	return c
}

// WithMaxTokens sets a custom max tokens value
func (c *Client) WithMaxTokens(maxTokens int) *Client {
	c.maxTokens = maxTokens
	return c
}

// SendMessage sends a message to Claude and returns the response
func (c *Client) SendMessage(messages []Message, systemPrompt string) (*Response, error) {
	req := Request{
		Model:     c.model,
		MaxTokens: c.maxTokens,
		System:    systemPrompt,
		Messages:  messages,
	}

	return c.sendRequest(req)
}

// SendSimpleMessage sends a simple user message to Claude
func (c *Client) SendSimpleMessage(userMessage string, systemPrompt string) (string, error) {
	messages := []Message{
		{
			Role:    "user",
			Content: userMessage,
		},
	}

	resp, err := c.SendMessage(messages, systemPrompt)
	if err != nil {
		return "", err
	}

	if len(resp.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return resp.Content[0].Text, nil
}

// sendRequest sends a request to the Claude API
func (c *Client) sendRequest(req Request) (*Response, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", AnthropicVersion)

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
