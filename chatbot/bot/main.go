package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"EventPilot/chatbot/internal/collector"
	"EventPilot/chatbot/internal/models"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable is required")
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
	eventCollector := collector.NewEventCollector(apiKey, event)

	// Start conversation
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("TEST BOT")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("\nPlease type your responses when ready.")
	fmt.Println("Type 'done' or 'exit' when finished.")
	fmt.Println("Type 'status' to see current progress.")
	fmt.Println("Type 'summary' to see what's been collected so far.")
	fmt.Println(strings.Repeat("-", 60) + "\n")

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
			fmt.Println("\nBot: Thank you! Let me summarize what we've collected...\n")
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
			fmt.Println("\nBot: I apologize, I had trouble processing that. Could you try rephrasing?\n")
			continue
		}

		fmt.Printf("\nBot: %s\n\n", response)

		// Show progress indicator
		progress := eventCollector.GetProgress()
		missingCount := len(eventCollector.GetMissingFields())
		fmt.Printf("[Progress: %.0f%% | %d field(s) remaining]\n\n", progress, missingCount)

		// Check if collection is complete
		if eventCollector.IsComplete() {
			fmt.Println("All information collected!")
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
	displayCollectedData(eventCollector.GetCollectedData())

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Thank you for using the Event Documentation Bot!")
	fmt.Println(strings.Repeat("=", 60))
}

func displayEventInfo(event *models.Event) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("EXISTING EVENT INFORMATION")
	fmt.Println(strings.Repeat("=", 60))
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

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("COLLECTED INFORMATION")
	fmt.Println(strings.Repeat("=", 60))

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
		fmt.Println("\nNo information collected yet.\n")
		return
	}

	fmt.Println("\n" + strings.Repeat("-", 60))
	fmt.Println("INFORMATION COLLECTED SO FAR:")
	fmt.Println(strings.Repeat("-", 60))

	for key, value := range collected {
		fmt.Printf("• %s: %s\n", formatFieldName(key), value)
	}
	fmt.Println(strings.Repeat("-", 60) + "\n")
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
