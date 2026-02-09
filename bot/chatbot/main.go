package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"EventPilot/bot/config"
	"EventPilot/bot/internal/collector"
	"EventPilot/bot/internal/models"
	"EventPilot/bot/internal/xapi"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Hardcoded event for testing
	event := &models.Event{
		ID:          "evt-001",
		Name:        "Tech Conference 2024",
		Date:        time.Now().AddDate(0, 0, -7), // 7 days ago
		Location:    "San Francisco Convention Center",
		Description: "Annual technology conference focusing on emerging trends",
		// Missing fields that will be collected: Highlights, TargetAudience, SpecialGuests, Photos
	}

	// Display event information
	displayEventInfo(event)

	// Initialize event collector
	eventCollector := collector.NewEventCollector(cfg.ClaudeAPIKey, event)

	// Start the conversation
	fmt.Println("EVENTPILOT BOT")
	fmt.Println("\nType your responses naturally.")
	fmt.Println("Type 'done' or 'exit' when finished.")
	fmt.Println("Type 'status' to see current progress.")
	fmt.Println("Type 'summary' to see what's been collected so far.")

	greeting, err := eventCollector.Start()
	if err != nil {
		log.Fatalf("Failed to start conversation: %v", err)
	}

	fmt.Printf("Bot: %s\n\n", greeting)

	// Main conversation loop
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		userInput := strings.TrimSpace(scanner.Text())

		// Skip empty input
		if userInput == "" {
			continue
		}

		// Handle special commands
		switch strings.ToLower(userInput) {
		case "done", "exit", "quit":
			fmt.Println("\nBot: Thank you! Let me summarize what we've collected...")
			goto finalize

		case "status":
			showStatus(eventCollector)
			continue

		case "summary":
			showSummary(eventCollector)
			continue
		}

		// Process the message
		response, err := eventCollector.ProcessMessage(userInput)
		if err != nil {
			log.Printf("Error processing message: %v", err)
			fmt.Println("\nBot: I apologize, I had trouble processing that. Could you try rephrasing?")
			continue
		}

		fmt.Printf("\nBot: %s\n\n", response)

		// Show progress indicator
		progress := eventCollector.GetProgress()
		missingCount := len(eventCollector.GetMissingFields())
		fmt.Printf("[Progress: %.0f%% | %d field(s) remaining]\n\n", progress, missingCount)

		// Check if collection is complete
		if eventCollector.IsComplete() {
			fmt.Println("✓ All information collected!")
			goto finalize
		}
	}

finalize:
	// Generate and display summary
	summary, err := eventCollector.GetSummary()
	if err != nil {
		log.Printf("Failed to generate summary: %v", err)
	} else {
		fmt.Printf("\n%s\n\n", summary)
	}

	// Display collected data
	collectedData := eventCollector.GetCollectedData()
	displayCollectedData(collectedData)

	// Offer to post to X
	if cfg.HasXAPICredentials() {
		fmt.Println("POST TO X (TWITTER)")
		fmt.Print("\nWould you like to post this to X? (yes/no): ")
		scanner.Scan()
		postConfirmation := strings.ToLower(strings.TrimSpace(scanner.Text()))

		if postConfirmation == "yes" || postConfirmation == "y" {
			handleXPosting(cfg, event, collectedData, scanner)
		}
	} else {
		fmt.Println("\n💡 Tip: Configure X API credentials to post directly to Twitter!")
		fmt.Println("Add X_API_KEY, X_API_SECRET, X_ACCESS_TOKEN, and X_ACCESS_TOKEN_SECRET to your .env file")
	}

	fmt.Println("Thank you for using EventPilot!")
}

func handleXPosting(cfg *config.Config, event *models.Event, collectedData map[string]string, scanner *bufio.Scanner) {
	// Initialize X API client
	xClient, err := xapi.NewClient(xapi.Config{
		APIKey:            cfg.XAPIKey,
		APISecret:         cfg.XAPISecret,
		AccessToken:       cfg.XAccessToken,
		AccessTokenSecret: cfg.XAccessTokenSecret,
	})
	if err != nil {
		log.Printf("Failed to initialize X API client: %v", err)
		return
	}

	// Get user info
	user, err := xClient.GetUserInfo()
	if err != nil {
		log.Printf("Failed to get X user info: %v", err)
	} else {
		fmt.Printf("✓ Connected to X as @%s\n\n", user)
	}

	// Create formatter
	formatter := xapi.NewPostFormatter()

	// Ask for post type
	fmt.Println("Choose post type:")
	fmt.Println("  1. Single post (concise)")
	fmt.Println("  2. Thread (detailed, multiple tweets)")
	fmt.Print("\nEnter choice (1 or 2): ")
	scanner.Scan()
	choice := strings.TrimSpace(scanner.Text())

	switch choice {
	case "1":
		// Single post
		postText, err := formatter.FormatSinglePost(event, collectedData)
		if err != nil {
			log.Printf("Error formatting post: %v", err)
			return
		}

		// Preview
		fmt.Println("\n" + formatter.PreviewPost(postText))

		// Confirm
		fmt.Print("\nPost this to X? (yes/no): ")
		scanner.Scan()
		if strings.ToLower(strings.TrimSpace(scanner.Text())) != "yes" {
			fmt.Println("Post cancelled.")
			return
		}

		// Post
		tweet, err := xClient.PostTweet(postText)
		if err != nil {
			log.Printf("Failed to post tweet: %v", err)
			return
		}

		fmt.Printf("\n✓ Successfully posted to X!\n")
		fmt.Printf("View at: https://twitter.com/%s/status/%s\n", user, tweet.ID)

	case "2":
		// Thread
		tweets, err := formatter.FormatThread(event, collectedData)
		if err != nil {
			log.Printf("Error formatting thread: %v", err)
			return
		}

		// Preview
		fmt.Println("\n" + formatter.PreviewThread(tweets))

		// Confirm
		fmt.Print("\nPost this thread to X? (yes/no): ")
		scanner.Scan()
		if strings.ToLower(strings.TrimSpace(scanner.Text())) != "yes" {
			fmt.Println("Thread cancelled.")
			return
		}

		// Post thread
		postedTweets, err := xClient.PostThread(tweets)
		if err != nil {
			log.Printf("Failed to post thread: %v", err)
			return
		}

		fmt.Printf("\n✓ Successfully posted thread to X! (%d tweets)\n", len(postedTweets))
		fmt.Printf("View at: https://twitter.com/%s/status/%s\n",
			user,
			postedTweets[0].ID)

	default:
		fmt.Println("Invalid choice. Post cancelled.")
	}
}

func displayEventInfo(event *models.Event) {
	fmt.Println("EXISTING EVENT INFORMATION")
	fmt.Printf("Event ID: %s\n", event.ID)
	fmt.Printf("Event Name: %s\n", event.Name)
	fmt.Printf("Date: %s\n", event.Date.Format("January 2, 2006"))
	fmt.Printf("Location: %s\n", event.Location)
	if event.Description != "" {
		fmt.Printf("Description: %s\n", event.Description)
	}

	missingFields := event.GetMissingFields()
	if len(missingFields) > 0 {
		fmt.Printf("\nFields to collect: %s\n", strings.Join(missingFields, ", "))
	}
}

func displayCollectedData(data map[string]string) {
	if len(data) == 0 {
		fmt.Println("\nNo information has been collected yet.")
		return
	}

	fmt.Println("COLLECTED INFORMATION")

	for key, value := range data {
		fmt.Printf("• %s: %s\n", formatFieldName(key), value)
	}
}

func showStatus(collector *collector.EventCollector) {
	progress := collector.GetProgress()
	missing := collector.GetMissingFields()
	collected := collector.GetCollectedData()

	fmt.Println("\n" + strings.Repeat("-", 60))
	fmt.Printf("Progress: %.0f%%\n", progress)
	fmt.Printf("Collected: %d field(s)\n", len(collected))
	fmt.Printf("Remaining: %d field(s)\n", len(missing))

	if len(missing) > 0 {
		fmt.Printf("Still needed: %s\n", strings.Join(missing, ", "))
	}
	fmt.Println(strings.Repeat("-", 60) + "\n")
}

func showSummary(collector *collector.EventCollector) {
	collected := collector.GetCollectedData()

	if len(collected) == 0 {
		fmt.Println("\nNo information collected yet.")
		return
	}

	fmt.Println("INFORMATION COLLECTED SO FAR:")

	for key, value := range collected {
		fmt.Printf("• %s: %s\n", formatFieldName(key), value)
	}
}

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
