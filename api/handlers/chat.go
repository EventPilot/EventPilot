package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"eventpilot/api/models"
	"eventpilot/api/services"
	"log"
	"net/http"

	"github.com/supabase-community/supabase-go"
)

type ChatHandler struct {
	SupabaseClient *supabase.Client
}

func GetChat(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotImplemented) }
func CreateChatMessage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// requestInputsForEvent creates info-collection chats for all valid members of an event.
// Returns an empty slice (no error) if there are no valid members.
func requestInputsForEvent(ctx context.Context, client *supabase.Client, eventId string) ([]string, error) {
	results := []models.EventMembersWithDetails{}
	_, err := client.From("event_member").
		Select("*, event(title, description, event_date), user(name)", "", false).
		Eq("event_id", eventId).
		ExecuteTo(&results)
	if err != nil {
		return nil, err
	}

	var validMembers []models.EventMembersWithDetails
	for _, res := range results {
		if res.UserID != "" {
			validMembers = append(validMembers, res)
		}
	}

	if len(validMembers) == 0 {
		return []string{}, nil
	}

	type rpcResult struct {
		id  string
		err error
	}

	ch := make(chan rpcResult, len(validMembers))
	for _, m := range validMembers {
		go func(m models.EventMembersWithDetails) {
			message := services.GenerateInitialMessage(ctx, m)
			log.Printf("[requestInputsForEvent] generated initial message for member %s (role: %s): %q", m.UserID, m.Role, message)
			rpcBody := map[string]any{
				"p_event_id":        eventId,
				"p_member_user_id":  m.UserID,
				"p_chat_type":       "info_collection",
				"p_initial_message": message,
			}
			id := client.Rpc("create_chat_with_initial_message", "", rpcBody)
			if id == "" {
				ch <- rpcResult{err: errors.New("failed to create chat")}
			} else {
				ch <- rpcResult{id: id}
			}
		}(m)
	}

	chatIDs := make([]string, 0, len(validMembers))
	for range validMembers {
		r := <-ch
		if r.err != nil {
			return nil, r.err
		}
		chatIDs = append(chatIDs, r.id)
	}
	return chatIDs, nil
}

func (h *ChatHandler) RequestInputs(w http.ResponseWriter, r *http.Request) {
	eventId := r.PathValue("id")
	log.Printf("[RequestInputs] received request for eventId=%s", eventId)

	chatIDs, err := requestInputsForEvent(r.Context(), h.SupabaseClient, eventId)
	if err != nil {
		log.Printf("[RequestInputs] error creating chats for eventId=%s: %v", eventId, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(chatIDs) == 0 {
		log.Printf("[RequestInputs] no event members found for eventId=%s", eventId)
		http.Error(w, "no event members found", http.StatusNotFound)
		return
	}

	log.Printf("[RequestInputs] created %d chats for eventId=%s: %v", len(chatIDs), eventId, chatIDs)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"chat_ids": chatIDs})
}
