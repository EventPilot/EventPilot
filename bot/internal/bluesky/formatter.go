package bluesky

import (
	"fmt"
	"strings"
	"time"
)

// PostFormatter formats event data for Bluesky posts
type PostFormatter struct{}

// NewPostFormatter creates a new formatter
func NewPostFormatter() *PostFormatter {
	return &PostFormatter{}
}

// FormatSinglePost creates a concise single post about an event
func (f *PostFormatter) FormatSinglePost(eventName, date, location string, highlights []string) string {
	// Bluesky limit is 300 characters (vs Twitter's 280)
	var post strings.Builder

	// Add event emoji and name
	post.WriteString("📅 ")
	post.WriteString(eventName)

	// Add date if available
	if date != "" {
		post.WriteString(" | ")
		post.WriteString(formatDate(date))
	}

	// Add highlight if available
	if len(highlights) > 0 && highlights[0] != "" {
		post.WriteString("\n✨ ")
		highlight := highlights[0]
		// Ensure we don't exceed limit
		remaining := 300 - post.Len()
		if len(highlight) > remaining-10 {
			highlight = highlight[:remaining-13] + "..."
		}
		post.WriteString(highlight)
	}

	return post.String()
}

// FormatThread creates a thread of posts about an event
func (f *PostFormatter) FormatThread(eventName, date, location, description string, highlights, audience, guests []string) []string {
	var posts []string

	// Post 1: Introduction with thread indicator
	intro := fmt.Sprintf("📅 Just wrapped up %s! 🧵\n\n", eventName)
	if date != "" {
		intro += fmt.Sprintf("📆 %s", formatDate(date))
	}
	if location != "" {
		if date != "" {
			intro += "\n"
		}
		intro += fmt.Sprintf("📍 %s", location)
	}
	posts = append(posts, intro)

	// Post 2: Description and highlights
	if description != "" || len(highlights) > 0 {
		var p2 strings.Builder
		if description != "" {
			p2.WriteString(truncate(description, 200))
		}

		if len(highlights) > 0 {
			if p2.Len() > 0 {
				p2.WriteString("\n\n")
			}
			p2.WriteString("✨ Highlights:\n")
			for i, highlight := range highlights {
				if i >= 3 { // Max 3 highlights
					break
				}
				if highlight == "" {
					continue
				}
				remaining := 300 - p2.Len() - 20
				if remaining < 20 {
					break
				}
				h := truncate(highlight, remaining)
				p2.WriteString("• ")
				p2.WriteString(h)
				p2.WriteString("\n")
			}
		}

		if p2.Len() > 0 {
			posts = append(posts, p2.String())
		}
	}

	// Post 3: Audience
	if len(audience) > 0 && audience[0] != "" {
		var p3 strings.Builder
		p3.WriteString("👥 Attendees:\n")
		for i, aud := range audience {
			if i >= 3 {
				break
			}
			if aud == "" {
				continue
			}
			remaining := 300 - p3.Len() - 20
			if remaining < 20 {
				break
			}
			a := truncate(aud, remaining)
			p3.WriteString("• ")
			p3.WriteString(a)
			p3.WriteString("\n")
		}

		if p3.Len() > 0 {
			posts = append(posts, p3.String())
		}
	}

	// Post 4: Special guests
	if len(guests) > 0 && guests[0] != "" {
		var p4 strings.Builder
		p4.WriteString("🎤 Special guests:\n")
		for i, guest := range guests {
			if i >= 3 {
				break
			}
			if guest == "" {
				continue
			}
			remaining := 300 - p4.Len() - 20
			if remaining < 20 {
				break
			}
			g := truncate(guest, remaining)
			p4.WriteString("• ")
			p4.WriteString(g)
			p4.WriteString("\n")
		}

		if p4.Len() > 0 {
			posts = append(posts, p4.String())
		}
	}

	// Post 5: Closing
	closing := "Thanks to everyone who participated! 🙏"
	if len(posts) >= 2 {
		posts = append(posts, closing)
	}

	return posts
}

// PreviewPost shows what a post will look like
func (f *PostFormatter) PreviewPost(text string) string {
	border := strings.Repeat("─", 50)
	return fmt.Sprintf("\n%s\n%s\n%s\n📊 Length: %d/300 characters\n",
		border, text, border, len(text))
}

// PreviewThread shows what a thread will look like
func (f *PostFormatter) PreviewThread(posts []string) string {
	var preview strings.Builder
	border := strings.Repeat("─", 50)

	for i, post := range posts {
		preview.WriteString(fmt.Sprintf("\n%s\n📝 Post %d/%d:\n%s\n%s\n📊 Length: %d/300 characters\n",
			border, i+1, len(posts), post, border, len(post)))
	}

	return preview.String()
}

// Helper functions

func formatDate(dateStr string) string {
	// Try to parse and format nicely
	formats := []string{
		"2006-01-02",
		"01/02/2006",
		"January 2, 2006",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t.Format("Jan 2, 2006")
		}
	}

	// If parsing fails, return as-is
	return dateStr
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return "..."
	}
	return s[:maxLen-3] + "..."
}
