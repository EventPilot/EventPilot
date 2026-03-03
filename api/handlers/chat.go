package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"eventpilot/api/models"
	"eventpilot/api/services"
	"net/http"

	"github.com/supabase-community/supabase-go"
)

type ChatHandler struct {
	SupabaseClient *supabase.Client
}

func GetChat(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotImplemented) }

type createChatMessageRequest struct {
	ChatID     string  `json:"chat_id"`
	SenderType string  `json:"sender_type"`
	SenderID   *string `json:"sender_id,omitempty"`
	Message    string  `json:"message"`
}

func (h *ChatHandler) CreateChatMessage(w http.ResponseWriter, r *http.Request) {
	eventId := r.PathValue("id")
	if eventId == "" {
		http.Error(w, "missing event id", http.StatusBadRequest)
		return
	}

	var req createChatMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.ChatID == "" {
		http.Error(w, "chat_id is required", http.StatusBadRequest)
		return
	}
	if req.Message == "" {
		http.Error(w, "message is required", http.StatusBadRequest)
		return
	}
	if req.SenderType == "" {
		req.SenderType = "user"
	}

	// Ensure the chat belongs to the given event.
	type chatRow struct {
		ID      string `json:"id"`
		EventID string `json:"event_id"`
	}
	var chats []chatRow
	_, err := h.SupabaseClient.
		From("chat").
		Select("id, event_id", "", false).
		Eq("id", req.ChatID).
		Eq("event_id", eventId).
		Limit(1, "").
		ExecuteTo(&chats)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to verify chat: %v", err), http.StatusInternalServerError)
		return
	}
	if len(chats) == 0 {
		http.Error(w, "chat not found for event", http.StatusNotFound)
		return
	}

	// Insert the new chat message.
	insertData := map[string]any{
		"chat_id":     req.ChatID,
		"sender_type": req.SenderType,
		"message":     req.Message,
	}
	if req.SenderID != nil {
		insertData["sender_id"] = req.SenderID
	}

	var created []map[string]any
	_, err = h.SupabaseClient.
		From("chat_message").
		Insert(insertData, false, "", "representation", "").
		ExecuteTo(&created)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create chat message: %v", err), http.StatusInternalServerError)
		return
	}
	if len(created) == 0 {
		http.Error(w, "failed to create chat message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(created[0])
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

	chatIDs, err := requestInputsForEvent(r.Context(), h.SupabaseClient, eventId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(chatIDs) == 0 {
		http.Error(w, "no event members found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"chat_ids": chatIDs})
}
