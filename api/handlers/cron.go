package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/supabase-community/supabase-go"
)

type CronHandler struct {
	SupabaseClient *supabase.Client
}

func (h *CronHandler) ProcessCompletedEvents(w http.ResponseWriter, r *http.Request) {
	secret := os.Getenv("CRON_SECRET")
	if secret == "" || r.Header.Get("Authorization") != "Bearer "+secret {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

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
		http.Error(w, fmt.Sprintf("failed to fetch events: %v", err), http.StatusInternalServerError)
		return
	}

	type chatRow struct {
		ID string `json:"id"`
	}

	results := map[string]any{}
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
			results[event.ID] = fmt.Sprintf("error checking chats: %v", err)
			continue
		}
		if len(chats) > 0 {
			skipped++
			continue
		}

		chatIDs, err := requestInputsForEvent(r.Context(), h.SupabaseClient, event.ID)
		if err != nil {
			results[event.ID] = fmt.Sprintf("error: %v", err)
			continue
		}
		results[event.ID] = chatIDs
		processed++
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"processed": processed,
		"skipped":   skipped,
		"results":   results,
	})
}
