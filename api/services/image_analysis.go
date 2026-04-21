package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// ImageAnalysis is the structured description of an uploaded image produced
// by the lightweight Haiku vision model. It is stored alongside the media
// record and surfaced to the planner/chat agent as additional context.
type ImageAnalysis struct {
	Description string `json:"description"`
}

// supported content types for Claude vision
var supportedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// AnalyzeImage sends the image bytes to Claude Haiku and returns a structured
// analysis. Returns an error if the API key is missing, the content type is
// unsupported, or the model call fails. Callers should treat analysis as a
// best-effort enrichment — persist the media record even if this fails.
func AnalyzeImage(ctx context.Context, data []byte, contentType, eventTitle, eventDescription string) (*ImageAnalysis, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, errors.New("ANTHROPIC_API_KEY not set")
	}
	if len(data) == 0 {
		return nil, errors.New("empty image data")
	}

	normalized := strings.ToLower(strings.TrimSpace(contentType))
	// normalize jpg -> jpeg for the Anthropic API
	if normalized == "image/jpg" {
		normalized = "image/jpeg"
	}
	if !supportedImageTypes[normalized] {
		return nil, fmt.Errorf("unsupported image content type: %s", contentType)
	}

	client := anthropic.NewClient(option.WithAPIKey(apiKey))

	encoded := base64.StdEncoding.EncodeToString(data)

	userPrompt := fmt.Sprintf(`Analyze this image uploaded for an event and return STRICT JSON matching this schema:
{
  "description": "2-3 sentence objective description of what is in the image, including setting, subjects, and mood"
}

Event title: %s
Event description: %s

Rules:
- Output ONLY the JSON object. No markdown, no preamble, no code fences.
- Do not invent identities of specific people. Describe them generically.`, eventTitle, eventDescription)

	msg, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5_20251001,
		MaxTokens: 400,
		System: []anthropic.TextBlockParam{{
			Text: "You are an image analyst for EventPilot. You look at images uploaded to an event and produce concise, factual metadata that a planner agent will use to decide how to feature the image in a post. Return only valid JSON.",
		}},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				anthropic.NewImageBlockBase64(normalized, encoded),
				anthropic.NewTextBlock(userPrompt),
			),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("claude vision call failed: %w", err)
	}

	var text strings.Builder
	for _, block := range msg.Content {
		if block.Type == "text" {
			text.WriteString(block.Text)
		}
	}
	raw := strings.TrimSpace(text.String())
	if raw == "" {
		return nil, errors.New("empty response from vision model")
	}

	payload := extractJSONPayload(raw)
	var analysis ImageAnalysis
	if err := json.Unmarshal([]byte(payload), &analysis); err != nil {
		log.Printf("[AnalyzeImage] failed to parse JSON: %v; raw=%q", err, raw)
		// Fall back to using the whole response as the description so the
		// caller still gets something useful.
		return &ImageAnalysis{Description: raw}, nil
	}

	return &analysis, nil
}
