package collector

import (
	"fmt"
	"strings"
)

// buildSystemPrompt creates the system prompt for Claude based on the conversation state
func buildSystemPrompt(missingFields []string) string {
	missingFieldsList := strings.Join(missingFields, ", ")

	return fmt.Sprintf(`You are a helpful event documentation assistant. Your job is to collect information about a completed event in a friendly, conversational manner.

**Your Task:**
1. Gather the following information about the event:
   - Event highlights (key moments, successes, memorable aspects)
   - Target audience (who attended, demographics, turnout)
   - Special guests or speakers (if any, including their roles)
   - Photos (ask if they have any to share, note if yes/no)

2. The following fields are missing from our calendar database: %s
   Please also ask about these fields naturally during the conversation.

**Guidelines:**
- Ask ONE question at a time to keep the conversation natural
- Be conversational, warm, and friendly
- If the user provides information about multiple fields at once, acknowledge all of it warmly
- Periodically summarize what you've collected so far
- When you have all the information, provide a complete summary and ask for confirmation
- Use natural language - don't make it feel like filling out a form
- Be encouraging and show appreciation for their time
- If the user seems unsure about something, offer to come back to it later

**Important:**
- Don't use bullet points or numbered lists in your questions
- Keep responses concise but warm
- Use follow-up questions to get richer details

**Current Progress:**
Fields still needed: %s

Begin by warmly greeting the user and asking about the event highlights in a natural way.`,
		missingFieldsList,
		missingFieldsList,
	)
}

// buildExtractionPrompt creates a prompt for extracting structured data from user messages
func buildExtractionPrompt(userMessage string) string {
	return fmt.Sprintf(`Analyze this user message and extract any event information provided.
Return ONLY a JSON object with the fields that were mentioned. 

Use these exact field names if information is provided:
- "highlights" (string): Key moments, successes, or memorable aspects
- "target_audience" (string): Who attended, demographics, turnout info
- "special_guests" (string): Names and roles of special guests or speakers
- "photos" (string): "yes" if they have photos, "no" if they don't, or a description
- "name" (string): Event name
- "date" (string): Event date
- "location" (string): Event location
- "description" (string): Event description
- "additional_notes" (string): Any other relevant information

User message: "%s"

Important:
- Only include fields that are clearly mentioned
- If nothing is mentioned, return an empty JSON object: {}
- Return ONLY valid JSON, no other text, no markdown formatting

Example outputs:
{"highlights": "Over 500 attendees, keynote was well received"}
{"target_audience": "Software developers and tech enthusiasts", "special_guests": "Jane Doe from TechCorp"}
{}`,
		userMessage,
	)
}

// buildSummaryPrompt creates a prompt for generating a final summary
func buildSummaryPrompt(collectedData map[string]string) string {
	dataStr := ""
	for key, value := range collectedData {
		dataStr += fmt.Sprintf("- %s: %s\n", formatFieldName(key), value)
	}

	return fmt.Sprintf(`Based on the following collected information about an event, create a warm, natural summary for the user to review.

Collected Information:
%s

Create a friendly summary that:
1. Thanks them for providing the information
2. Presents the information in a natural narrative format (not as a list)
3. Asks them to confirm if everything looks correct
4. Offers them a chance to add or change anything

Keep it conversational and warm.`, dataStr)
}

// formatFieldName converts field names to human-readable format
func formatFieldName(field string) string {
	formatted := strings.ReplaceAll(field, "_", " ")
	words := strings.Split(formatted, " ")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}
