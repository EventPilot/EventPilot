package xapi

import (
	"fmt"
	"strings"

	"EventPilot/bot/internal/models"
)

// PostFormatter formats event data into X posts
type PostFormatter struct{}

// NewPostFormatter creates a new post formatter
func NewPostFormatter() *PostFormatter {
	return &PostFormatter{}
}

// FormatSinglePost formats event data into a single post (max 280 chars)
func (f *PostFormatter) FormatSinglePost(event *models.Event, collectedData map[string]string) (string, error) {
	// Build the post
	var post strings.Builder

	// Event name and date
	post.WriteString(fmt.Sprintf("📅 %s", event.Name))
	if !event.Date.IsZero() {
		post.WriteString(fmt.Sprintf(" | %s", event.Date.Format("Jan 2, 2006")))
	}
	post.WriteString("\n\n")

	// Highlights (most important)
	if highlights, ok := collectedData["highlights"]; ok && highlights != "" {
		post.WriteString(fmt.Sprintf("✨ %s", highlights))
	}

	postText := post.String()

	// Check length
	if len(postText) > 280 {
		// Truncate if too long
		postText = postText[:277] + "..."
	}

	return postText, nil
}

// FormatThread formats event data into a thread of posts
func (f *PostFormatter) FormatThread(event *models.Event, collectedData map[string]string) ([]string, error) {
	var tweets []string

	// Tweet 1: Introduction
	tweet1 := fmt.Sprintf("📅 Just wrapped up %s! Here's what happened: 🧵", event.Name)
	if !event.Date.IsZero() {
		tweet1 = fmt.Sprintf("📅 Just wrapped up %s on %s! Here's what happened: 🧵",
			event.Name,
			event.Date.Format("Jan 2"))
	}
	tweets = append(tweets, tweet1)

	// Tweet 2: Highlights
	if highlights, ok := collectedData["highlights"]; ok && highlights != "" {
		tweet2 := fmt.Sprintf("✨ Highlights:\n%s", f.truncateText(highlights, 250))
		tweets = append(tweets, tweet2)
	}

	// Tweet 3: Audience
	if audience, ok := collectedData["target_audience"]; ok && audience != "" {
		tweet3 := fmt.Sprintf("👥 Attendees:\n%s", f.truncateText(audience, 250))
		tweets = append(tweets, tweet3)
	}

	// Tweet 4: Special guests
	if guests, ok := collectedData["special_guests"]; ok && guests != "" {
		tweet4 := fmt.Sprintf("🎤 Special guests:\n%s", f.truncateText(guests, 250))
		tweets = append(tweets, tweet4)
	}

	// Tweet 5: Location (if available)
	if event.Location != "" {
		tweet5 := fmt.Sprintf("📍 Location: %s", event.Location)
		tweets = append(tweets, tweet5)
	}

	// Tweet 6: Closing
	if len(tweets) > 1 {
		closing := "Thanks to everyone who attended! 🙏"
		if photos, ok := collectedData["photos"]; ok && strings.Contains(strings.ToLower(photos), "yes") {
			closing += " Photos coming soon! 📸"
		}
		tweets = append(tweets, closing)
	}

	return tweets, nil
}

// FormatCustomPost allows custom formatting with Claude AI
func (f *PostFormatter) FormatCustomPost(event *models.Event, collectedData map[string]string, tone string) string {
	// This will be enhanced with Claude later
	// For now, use a simple format
	var post strings.Builder

	post.WriteString(fmt.Sprintf("%s was incredible! ", event.Name))

	if highlights, ok := collectedData["highlights"]; ok {
		post.WriteString(fmt.Sprintf("%s ", f.truncateText(highlights, 200)))
	}

	return f.truncateText(post.String(), 280)
}

// GenerateHashtags generates relevant hashtags from event data
func (f *PostFormatter) GenerateHashtags(event *models.Event, collectedData map[string]string) []string {
	hashtags := []string{}

	// Event name as hashtag (remove spaces)
	eventHash := strings.ReplaceAll(event.Name, " ", "")
	if len(eventHash) > 0 && len(eventHash) < 30 {
		hashtags = append(hashtags, "#"+eventHash)
	}

	// Year hashtag
	if !event.Date.IsZero() {
		year := event.Date.Year()
		hashtags = append(hashtags, fmt.Sprintf("#%d", year))
	}

	// Generic event hashtags
	hashtags = append(hashtags, "#Event", "#EventRecap")

	return hashtags
}

// PreviewPost shows what the post will look like with character count
func (f *PostFormatter) PreviewPost(postText string) string {
	charCount := len(postText)
	preview := strings.Builder{}

	preview.WriteString("=" + strings.Repeat("=", 60) + "\n")
	preview.WriteString("POST PREVIEW\n")
	preview.WriteString("=" + strings.Repeat("=", 60) + "\n\n")
	preview.WriteString(postText)
	preview.WriteString("\n\n")
	preview.WriteString(strings.Repeat("-", 60) + "\n")
	preview.WriteString(fmt.Sprintf("Character count: %d/280", charCount))

	if charCount > 280 {
		preview.WriteString(" ⚠️  TOO LONG")
	} else {
		preview.WriteString(" ✅")
	}

	preview.WriteString("\n" + strings.Repeat("=", 60))

	return preview.String()
}

// PreviewThread shows what the thread will look like
func (f *PostFormatter) PreviewThread(tweets []string) string {
	preview := strings.Builder{}

	preview.WriteString("=" + strings.Repeat("=", 60) + "\n")
	preview.WriteString(fmt.Sprintf("THREAD PREVIEW (%d tweets)\n", len(tweets)))
	preview.WriteString("=" + strings.Repeat("=", 60) + "\n\n")

	for i, tweet := range tweets {
		preview.WriteString(fmt.Sprintf("Tweet %d/%d (%d chars):\n", i+1, len(tweets), len(tweet)))
		preview.WriteString(strings.Repeat("-", 60) + "\n")
		preview.WriteString(tweet)
		preview.WriteString("\n" + strings.Repeat("-", 60) + "\n\n")
	}

	preview.WriteString(strings.Repeat("=", 60))

	return preview.String()
}

// truncateText truncates text to fit within maxLength
func (f *PostFormatter) truncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength-3] + "..."
}

// SplitIntoTweets splits long text into multiple tweets
func (f *PostFormatter) SplitIntoTweets(text string, maxLength int) []string {
	if maxLength <= 0 {
		maxLength = 280
	}

	// If text fits in one tweet
	if len(text) <= maxLength {
		return []string{text}
	}

	var tweets []string
	words := strings.Fields(text)
	currentTweet := strings.Builder{}

	for _, word := range words {
		// Check if adding this word would exceed limit
		testLength := currentTweet.Len()
		if testLength > 0 {
			testLength++ // for space
		}
		testLength += len(word)

		if testLength > maxLength {
			// Save current tweet and start new one
			if currentTweet.Len() > 0 {
				tweets = append(tweets, strings.TrimSpace(currentTweet.String()))
				currentTweet.Reset()
			}
		}

		if currentTweet.Len() > 0 {
			currentTweet.WriteString(" ")
		}
		currentTweet.WriteString(word)
	}

	// Add the last tweet
	if currentTweet.Len() > 0 {
		tweets = append(tweets, strings.TrimSpace(currentTweet.String()))
	}

	return tweets
}
