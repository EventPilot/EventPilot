package collector

import (
	"encoding/json"
	"fmt"
	"strings"

	"EventPilot/bot/internal/claude"
	"EventPilot/bot/internal/models"
)

// Handles extracting structured information from user messages
type Extractor struct {
	claudeClient *claude.Client
}

// Creates a new information extractor
func NewExtractor(claudeClient *claude.Client) *Extractor {
	return &Extractor{
		claudeClient: claudeClient,
	}
}

// Analyzes a user message and extracts structured event information
func (e *Extractor) Extract(userMessage string, state *models.ConversationState) error {
	prompt := buildExtractionPrompt(userMessage)

	response, err := e.claudeClient.SendSimpleMessage(prompt, "")
	if err != nil {
		return fmt.Errorf("failed to extract information: %w", err)
	}

	// Clean up the response (remove markdown code blocks if present)
	cleanedJSON := cleanJSONResponse(response)

	// Parse the extracted data
	var extractedData map[string]interface{}
	if err := json.Unmarshal([]byte(cleanedJSON), &extractedData); err != nil {
		// If JSON parsing fails, it's okay - user might not have provided extractable info
		return nil
	}

	// Add extracted data to conversation state
	for key, value := range extractedData {
		if strValue, ok := value.(string); ok && strValue != "" {
			state.AddCollectedData(key, strValue)
		}
	}

	return nil
}

// Removes markdown formatting and extra whitespace from JSON responses
func cleanJSONResponse(response string) string {
	// Remove markdown code blocks
	cleaned := strings.TrimSpace(response)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	return cleaned
}

// Creates a summary of all collected information
func (e *Extractor) GenerateSummary(state *models.ConversationState) (string, error) {
	if len(state.CollectedData) == 0 {
		return "No information has been collected yet.", nil
	}

	prompt := buildSummaryPrompt(state.CollectedData)

	summary, err := e.claudeClient.SendSimpleMessage(prompt, "")
	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %w", err)
	}

	return summary, nil
}

// Checks if all required fields have been collected
func (e *Extractor) ValidateCompleteness(state *models.ConversationState) (bool, []string) {
	if len(state.MissingFields) == 0 {
		return true, nil
	}
	return false, state.MissingFields
}
