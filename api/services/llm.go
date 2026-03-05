package services

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"eventpilot/api/models"
)

var fallbackMessages = map[string]string{
	"owner":        "Hi! I'm here to help capture the key moments from your event. Could you share what went well and any highlights you'd like to feature?",
	"photographer": "Hi! I'm collecting media for this event. Could you upload your best photos or videos along with a short caption for each?",
	"customer":     "Hi! Thanks for attending. We'd love to hear your thoughts — could you share a quick quote or highlight from your experience?",
	"partner":      "Hi! Thanks for being part of this event. We'd love to include your perspective — could you share a quick quote or key takeaway?",
}

const defaultFallback = "Hello! I'm here to help collect information and media for this event. What would you like to share?"

// GenerateInitialMessage calls Claude Haiku to produce a role-specific opening message for
// an event member's chat thread. On any error it returns a role-appropriate fallback.
func GenerateInitialMessage(ctx context.Context, member models.EventMembersWithDetails) string {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Println("ANTHROPIC_API_KEY not set, using fallback message")
		return getFallback(member.Role)
	}

	client := anthropic.NewClient(option.WithAPIKey(apiKey))

	msg, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5_20251001,
		MaxTokens: 200,
		System: []anthropic.TextBlockParam{
			{Text: buildSystemPrompt()},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(buildUserPrompt(member))),
		},
	})
	if err != nil {
		log.Printf("LLM call failed for member %s (role: %s): %v", member.UserID, member.Role, err)
		return getFallback(member.Role)
	}

	for _, block := range msg.Content {
		if block.Type == "text" && block.Text != "" {
			return block.Text
		}
	}

	log.Printf("LLM returned empty content for member %s (role: %s)", member.UserID, member.Role)
	return getFallback(member.Role)
}

func buildSystemPrompt() string {
	return `You are EventPilot, an AI agent that collects information and media from event stakeholders after an event.
Your job is to send a single warm, concise opening message to an event member to kick off a data-collection conversation.
The message should:
- Be 2-3 sentences maximum.
- Address the person's specific role directly (not generically).
- Ask for one concrete thing that is most relevant to their role.
- Sound conversational and friendly, not corporate.
- Not ask multiple questions at once.
Output only the message text — no preamble, no quotes, no role label.`
}

func buildUserPrompt(member models.EventMembersWithDetails) string {
	eventTitle := "Untitled Event"
	eventDescription := ""
	eventDate := ""
	userName := "there"

	if member.Event != nil {
		if member.Event.Title != "" {
			eventTitle = member.Event.Title
		}
		eventDescription = member.Event.Description
		eventDate = member.Event.EventDate
	}
	if member.User != nil && member.User.Name != "" {
		userName = member.User.Name
	}

	return fmt.Sprintf(`Generate an opening message for an event member with the following context:

Event name: %s
Event date: %s
Event description: %s
Member name: %s
Member role: %s

Role guidance:
- owner: Ask them about overall outcomes, key highlights, and anything they want featured in the post.
- photographer: Ask them to upload their best photos/videos with short captions.
- customer / partner / attendee: Ask them for a personal quote or standout moment from the event.
- (other roles): Ask what they would most like to share about the event.`,
		eventTitle,
		eventDate,
		eventDescription,
		userName,
		member.Role,
	)
}

func getFallback(role string) string {
	if msg, ok := fallbackMessages[role]; ok {
		return msg
	}
	return defaultFallback
}
