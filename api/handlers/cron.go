package handlers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/supabase-community/supabase-go"
)

type CronHandler struct {
	SupabaseClient *supabase.Client
}

func (h *CronHandler) ProcessCompletedEvents(ctx context.Context) error {
	today := time.Now().UTC().Format("2006-01-02")

	type eventRow struct {
		ID string `json:"id"`
	}
	var events []eventRow
	_, err := h.SupabaseClient.From("events").
		Select("id", "", false).
		Lte("event_date", today).
		ExecuteTo(&events)
	if err != nil {
		return fmt.Errorf("failed to fetch events: %w", err)
	}

	type chatRow struct {
		ID string `json:"id"`
	}

	processed, skipped := 0, 0

	for _, event := range events {
		// Idempotency: skip events that already have chats created.
		var chats []chatRow
		_, err := h.SupabaseClient.From("chats").
			Select("id", "", false).
			Eq("event_id", event.ID).
			Limit(1, "").
			ExecuteTo(&chats)
		if err != nil {
			log.Printf("cron: error checking chats for event %s: %v", event.ID, err)
			continue
		}
		if len(chats) > 0 {
			skipped++
			continue
		}

		if _, err = requestInputsForEvent(ctx, h.SupabaseClient, event.ID); err != nil {
			log.Printf("cron: error processing event %s: %v", event.ID, err)
			continue
		}
		processed++
	}

	log.Printf("cron: processed=%d skipped=%d", processed, skipped)
	return nil
}
