package collector

import (
	"fmt"
	"strings"

	"EventPilot/bot/internal/claude"
	"EventPilot/bot/internal/models"
)

// EventCollector manages the conversation flow for collecting event information
type EventCollector struct {
	claudeClient        *claude.Client
	event               *models.Event
	conversationHistory []claude.Message
	conversationStep    int
}

// NewEventCollector creates a new event collector
func NewEventCollector(apiKey string, event *models.Event) *EventCollector {
	claudeClient := claude.NewClient(apiKey)

	return &EventCollector{
		claudeClient:        claudeClient,
		event:               event,
		conversationHistory: []claude.Message{},
		conversationStep:    0,
	}
}

// ProcessMessage handles a user message and returns Claude's response
func (ec *EventCollector) ProcessMessage(userMessage string) (string, error) {
	// Add user message to history
	ec.conversationHistory = append(ec.conversationHistory, claude.Message{
		Role:    "user",
		Content: userMessage,
	})

	// Build system prompt with current event state
	systemPrompt := ec.buildSystemPrompt()

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

	// Try to extract information from the conversation
	ec.extractFromConversation(userMessage, assistantMessage)

	// Increment conversation step
	ec.conversationStep++

	return assistantMessage, nil
}

// extractFromConversation uses simple pattern matching to extract event data
func (ec *EventCollector) extractFromConversation(userMsg, botMsg string) {
	userLower := strings.ToLower(userMsg)

	// Simple extraction patterns (can be enhanced later)
	if ec.event.Name == "" {
		// Look for event names after common phrases
		if strings.Contains(userLower, "called") || strings.Contains(userLower, "named") {
			// Extract text after "called" or "named"
			if idx := strings.Index(userLower, "called "); idx != -1 {
				ec.event.Name = strings.TrimSpace(userMsg[idx+7:])
			} else if idx := strings.Index(userLower, "named "); idx != -1 {
				ec.event.Name = strings.TrimSpace(userMsg[idx+6:])
			}
		}
	}

	// Extract location
	if ec.event.Location == "" {
		if strings.Contains(userLower, " in ") || strings.Contains(userLower, " at ") {
			if idx := strings.Index(userLower, " in "); idx != -1 {
				location := strings.TrimSpace(userMsg[idx+4:])
				if len(location) > 0 && len(location) < 100 {
					ec.event.Location = location
				}
			} else if idx := strings.Index(userLower, " at "); idx != -1 {
				location := strings.TrimSpace(userMsg[idx+4:])
				if len(location) > 0 && len(location) < 100 {
					ec.event.Location = location
				}
			}
		}
	}

	// Extract date (simple pattern)
	if ec.event.Date == "" {
		dateKeywords := []string{"yesterday", "today", "last week", "last month", "on ", "january", "february", "march", "april", "may", "june", "july", "august", "september", "october", "november", "december"}
		for _, keyword := range dateKeywords {
			if strings.Contains(userLower, keyword) {
				// Just store the user's message as the date for now
				words := strings.Fields(userMsg)
				if len(words) > 0 {
					ec.event.Date = userMsg
				}
				break
			}
		}
	}
}

// buildSystemPrompt creates the system prompt based on current event state
func (ec *EventCollector) buildSystemPrompt() string {
	missingFields := ec.event.GetMissingFields()

	var prompt strings.Builder
	prompt.WriteString("You are a helpful assistant collecting information about an event through natural conversation. ")
	prompt.WriteString("Ask questions one at a time in a friendly, conversational way.\n\n")

	if len(missingFields) > 0 {
		prompt.WriteString("Information still needed:\n")
		for _, field := range missingFields {
			prompt.WriteString(fmt.Sprintf("- %s\n", formatSingleFieldName(field)))
		}
		prompt.WriteString("\nFocus on collecting one piece of information at a time. ")
		prompt.WriteString("Be conversational and don't list multiple questions.")
	} else {
		prompt.WriteString("You have collected all the basic information! ")
		prompt.WriteString("Ask if there's anything else they'd like to add about the event.")
	}

	return prompt.String()
}

func formatSingleFieldName(field string) string {
	switch field {
	case "name":
		return "Event name"
	case "date":
		return "When it took place"
	case "location":
		return "Where it was held"
	case "description":
		return "Brief description of the event"
	case "highlights":
		return "Key highlights or memorable moments"
	case "target_audience":
		return "Who attended (target audience)"
	case "special_guests":
		return "Special guests or speakers"
	default:
		return field
	}
}

// GetSummary generates a summary of all collected information
func (ec *EventCollector) GetSummary() (string, error) {
	var summary strings.Builder

	summary.WriteString("📋 Event Summary\n")
	summary.WriteString(strings.Repeat("=", 50) + "\n\n")

	if ec.event.Name != "" {
		summary.WriteString(fmt.Sprintf("📌 Name: %s\n", ec.event.Name))
	}
	if ec.event.Date != "" {
		summary.WriteString(fmt.Sprintf("📅 Date: %s\n", ec.event.Date))
	}
	if ec.event.Location != "" {
		summary.WriteString(fmt.Sprintf("📍 Location: %s\n", ec.event.Location))
	}
	if ec.event.Description != "" {
		summary.WriteString(fmt.Sprintf("\n📝 Description:\n%s\n", ec.event.Description))
	}

	if len(ec.event.Highlights) > 0 && ec.event.Highlights[0] != "" {
		summary.WriteString("\n✨ Highlights:\n")
		for _, h := range ec.event.Highlights {
			if h != "" {
				summary.WriteString(fmt.Sprintf("  • %s\n", h))
			}
		}
	}

	if len(ec.event.TargetAudience) > 0 && ec.event.TargetAudience[0] != "" {
		summary.WriteString("\n👥 Audience:\n")
		for _, a := range ec.event.TargetAudience {
			if a != "" {
				summary.WriteString(fmt.Sprintf("  • %s\n", a))
			}
		}
	}

	if len(ec.event.SpecialGuests) > 0 && ec.event.SpecialGuests[0] != "" {
		summary.WriteString("\n🎤 Special Guests:\n")
		for _, g := range ec.event.SpecialGuests {
			if g != "" {
				summary.WriteString(fmt.Sprintf("  • %s\n", g))
			}
		}
	}

	if summary.Len() == 0 {
		return "No information collected yet.", nil
	}

	return summary.String(), nil
}

// GetStatus returns a status message about what's been collected
func (ec *EventCollector) GetStatus() string {
	missing := ec.event.GetMissingFields()

	if len(missing) == 0 {
		return "✅ All information collected! Type 'summary' to see what we have, or 'done' to finish."
	}

	totalFields := 8 // Total fields we're collecting
	collected := totalFields - len(missing)

	var status strings.Builder
	status.WriteString(fmt.Sprintf("📊 Progress: Collected %d/%d fields\n", collected, totalFields))
	status.WriteString("\nStill needed:\n")
	for _, field := range missing {
		status.WriteString(fmt.Sprintf("  • %s\n", formatSingleFieldName(field)))
	}

	return status.String()
}
