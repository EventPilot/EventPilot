package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"EventPilot/bot/internal/bluesky"
	"EventPilot/bot/internal/collector"
	"EventPilot/bot/internal/models"
)

func main() {
	log.Println("Starting Event Documentation Bot (Bluesky Edition)")

	// Load configuration from environment
	claudeAPIKey := strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY"))
	blueskyHandle := strings.TrimSpace(os.Getenv("BLUESKY_HANDLE"))
	blueskyPassword := strings.TrimSpace(os.Getenv("BLUESKY_APP_PASSWORD"))

	if claudeAPIKey == "" {
		log.Fatal("ANTHROPIC_API_KEY is required")
	}

	// Create event model
	event := &models.Event{}

	// Initialize event collector
	eventCollector := collector.NewEventCollector(claudeAPIKey, event)

	// Start conversation
	fmt.Println("\n🤖 Hi! I'll help you document your event.")
	fmt.Println("💬 Just tell me about it naturally, and I'll ask questions to fill in the details.")
	fmt.Println("📝 Type 'done' when you're finished, 'status' to see what we have, or 'summary' for a recap.")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		userInput := strings.TrimSpace(scanner.Text())
		if userInput == "" {
			continue
		}

		// Handle commands
		switch strings.ToLower(userInput) {
		case "done", "exit", "quit":
			if !event.IsComplete() {
				fmt.Println("\n⚠️  Event information is incomplete. Are you sure you want to exit? (yes/no)")
				fmt.Print("You: ")
				if scanner.Scan() {
					if strings.ToLower(strings.TrimSpace(scanner.Text())) != "yes" {
						continue
					}
				}
			}

			// Offer to post to Bluesky
			if blueskyHandle != "" && blueskyPassword != "" {
				if err := handleBlueskyPosting(event, blueskyHandle, blueskyPassword); err != nil {
					log.Printf("Bluesky posting failed: %v", err)
				}
			} else {
				fmt.Println("\n💡 To post to Bluesky, set BLUESKY_HANDLE and BLUESKY_APP_PASSWORD in .env")
			}

			fmt.Println("\n👋 Thanks for using Event Documentation Bot!")
			return

		case "status":
			fmt.Println(eventCollector.GetStatus())
			continue

		case "summary":
			summary, err := eventCollector.GetSummary()
			if err != nil {
				fmt.Printf("❌ Failed to generate summary: %v\n", err)
				continue
			}
			fmt.Printf("\n📋 Summary:\n%s\n\n", summary)
			continue
		}

		// Process the message
		response, err := eventCollector.ProcessMessage(userInput)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		fmt.Printf("\nBot: %s\n\n", response)

		// Check if collection is complete
		if event.IsComplete() {
			fmt.Println("✅ Great! I have all the information I need.")
			fmt.Println("📝 Type 'summary' to see what we collected, or 'done' to finish.")
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading input: %v", err)
	}
}

func handleBlueskyPosting(event *models.Event, handle, password string) error {
	fmt.Println("\n🦋 Would you like to post this event to Bluesky?")
	fmt.Println("Choose an option:")
	fmt.Println("  1. Single post (concise summary)")
	fmt.Println("  2. Thread (detailed, multiple posts)")
	fmt.Println("  3. Skip")
	fmt.Print("\nYour choice (1-3): ")

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return fmt.Errorf("failed to read choice")
	}

	choice := strings.TrimSpace(scanner.Text())

	if choice == "3" || choice == "skip" {
		fmt.Println("Skipping Bluesky post.")
		return nil
	}

	// Initialize Bluesky client
	fmt.Println("\n Connecting to Bluesky")
	bskyClient, err := bluesky.NewClient(bluesky.Config{
		Handle:   handle,
		Password: password,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to Bluesky: %w", err)
	}

	username := bskyClient.GetUserHandle()
	fmt.Printf("Connected as @%s\n\n", username)

	formatter := bluesky.NewPostFormatter()

	switch choice {
	case "1", "single":
		// Single post
		postText := formatter.FormatSinglePost(
			event.Name,
			event.Date,
			event.Location,
			event.Highlights,
		)

		fmt.Println("📝 Preview:")
		fmt.Println(formatter.PreviewPost(postText))
		fmt.Print("Post this? (yes/no): ")

		if !scanner.Scan() {
			return fmt.Errorf("failed to read confirmation")
		}

		if strings.ToLower(strings.TrimSpace(scanner.Text())) != "yes" {
			fmt.Println("Post cancelled.")
			return nil
		}

		fmt.Println("\n📤 Posting...")
		post, err := bskyClient.PostText(postText)
		if err != nil {
			return fmt.Errorf("failed to post: %w", err)
		}

		postURL := bskyClient.GetPostURL(post)
		fmt.Printf("✅ Posted successfully!\n")
		fmt.Printf("🔗 View at: %s\n", postURL)

	case "2", "thread":
		// Thread
		posts := formatter.FormatThread(
			event.Name,
			event.Date,
			event.Location,
			event.Description,
			event.Highlights,
			event.TargetAudience,
			event.SpecialGuests,
		)

		fmt.Println("📝 Preview:")
		fmt.Println(formatter.PreviewThread(posts))
		fmt.Printf("Post this %d-post thread? (yes/no): ", len(posts))

		if !scanner.Scan() {
			return fmt.Errorf("failed to read confirmation")
		}

		if strings.ToLower(strings.TrimSpace(scanner.Text())) != "yes" {
			fmt.Println("Thread cancelled.")
			return nil
		}

		fmt.Println("\n📤 Posting thread...")
		postedPosts, err := bskyClient.PostThread(posts)
		if err != nil {
			return fmt.Errorf("failed to post thread: %w", err)
		}

		fmt.Printf("✅ Thread posted successfully! (%d posts)\n", len(postedPosts))
		if len(postedPosts) > 0 {
			firstPostURL := bskyClient.GetPostURL(postedPosts[0])
			fmt.Printf("🔗 View thread at: %s\n", firstPostURL)
		}

	default:
		fmt.Println("Invalid choice. Skipping Bluesky post.")
	}

	return nil
}
