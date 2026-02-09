package collector

import (
	"fmt"

	"EventPilot/bot/internal/claude"
	"EventPilot/bot/internal/models"
)

// EventCollector manages the conversation flow for collecting event information
type EventCollector struct {
	claudeClient        *claude.Client
	extractor           *Extractor
	conversationState   *models.ConversationState
	conversationHistory []claude.Message
}

// NewEventCollector creates a new event collector
func NewEventCollector(apiKey string, event *models.Event) *EventCollector {
	claudeClient := claude.NewClient(apiKey)
	extractor := NewExtractor(claudeClient)

	// Get missing fields from the event
	missingFields := event.GetMissingFields()
	state := models.NewConversationState(event.ID, missingFields)

	return &EventCollector{
		claudeClient:        claudeClient,
		extractor:           extractor,
		conversationState:   state,
		conversationHistory: []claude.Message{},
	}
}

// Start initiates the conversation with a greeting
func (ec *EventCollector) Start() (string, error) {
	systemPrompt := buildSystemPrompt(ec.conversationState.MissingFields)

	// Send initial greeting request
	greeting, err := ec.claudeClient.SendSimpleMessage(
		"Hello! I'm ready to help document this event.",
		systemPrompt,
	)
	if err != nil {
		return "", fmt.Errorf("failed to start conversation: %w", err)
	}

	// Add to conversation history
	ec.conversationHistory = append(ec.conversationHistory, claude.Message{
		Role:    "user",
		Content: "Hello! I'm ready to help document this event.",
	})
	ec.conversationHistory = append(ec.conversationHistory, claude.Message{
		Role:    "assistant",
		Content: greeting,
	})

	return greeting, nil
}

// ProcessMessage handles a user message and returns Claude's response
func (ec *EventCollector) ProcessMessage(userMessage string) (string, error) {
	// Add user message to history
	ec.conversationHistory = append(ec.conversationHistory, claude.Message{
		Role:    "user",
		Content: userMessage,
	})

	// Extract information from the user's message
	if err := ec.extractor.Extract(userMessage, ec.conversationState); err != nil {
		return "", fmt.Errorf("failed to extract information: %w", err)
	}

	// Build system prompt with updated missing fields
	systemPrompt := buildSystemPrompt(ec.conversationState.MissingFields)

	// Get Claude's response
	response, err := ec.claudeClient.SendMessage(ec.conversationHistory, systemPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to get response: %w", err)
	}

	if len(response.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	assistantMessage := response.Content[0].Text

	// Add assistant response to history
	ec.conversationHistory = append(ec.conversationHistory, claude.Message{
		Role:    "assistant",
		Content: assistantMessage,
	})

	// Increment conversation step
	ec.conversationState.ConversationStep++

	return assistantMessage, nil
}

// GetSummary generates a summary of all collected information
func (ec *EventCollector) GetSummary() (string, error) {
	return ec.extractor.GenerateSummary(ec.conversationState)
}

// IsComplete checks if all required information has been collected
func (ec *EventCollector) IsComplete() bool {
	return ec.conversationState.IsComplete
}

// GetProgress returns the completion percentage
func (ec *EventCollector) GetProgress() float64 {
	return ec.conversationState.GetProgress()
}

// GetCollectedData returns all collected data
func (ec *EventCollector) GetCollectedData() map[string]string {
	return ec.conversationState.CollectedData
}

// GetMissingFields returns the list of fields still needed
func (ec *EventCollector) GetMissingFields() []string {
	return ec.conversationState.MissingFields
}

// GetState returns the current conversation state
func (ec *EventCollector) GetState() *models.ConversationState {
	return ec.conversationState
}
