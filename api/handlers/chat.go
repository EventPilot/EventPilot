package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"eventpilot/api/models"
	"eventpilot/api/services"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/supabase-community/supabase-go"
)

type ChatHandler struct {
	SupabaseClient *supabase.Client
	RunManager     *services.RunManager
}

func GetChat(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotImplemented) }

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
				"p_initial_message": message,
			}
			id := client.Rpc("create_chat_with_initial_message", "", rpcBody)
			log.Printf("[requestInputsForEvent] RPC response for member %s: %q", m.UserID, id)
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

type createChatMessageRequest struct {
	Message string `json:"message"`
}

type createChatMessageResponse struct {
	ChatID string             `json:"chat_id"`
	Run    *services.AgentRun `json:"run"`
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

func (h *ChatHandler) CreateChatMessage(w http.ResponseWriter, r *http.Request) {
	user, err := authenticatedUserFromRequest(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	eventID := r.PathValue("id")
	if strings.TrimSpace(eventID) == "" {
		http.Error(w, "event id is required", http.StatusBadRequest)
		return
	}

	var req createChatMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.Message = strings.TrimSpace(req.Message)
	if req.Message == "" {
		http.Error(w, "message is required", http.StatusBadRequest)
		return
	}

	chatID, err := services.EnsureChatForUser(h.SupabaseClient, eventID, user.ID.String())
	if err != nil {
		http.Error(w, "failed to resolve chat", http.StatusInternalServerError)
		return
	}

	senderID := user.ID.String()
	_, _, err = h.SupabaseClient.From("chat_message").Insert(map[string]any{
		"id":           uuid.NewString(),
		"chat_id":      chatID,
		"sender_type":  "user",
		"sender_id":    senderID,
		"message":      req.Message,
		"message_type": "message",
		"metadata":     map[string]any{},
	}, false, "", "", "").Execute()
	if err != nil {
		http.Error(w, "failed to save message", http.StatusInternalServerError)
		return
	}

	handledBlockedRun, err := h.RunManager.HandleIncomingMessage(r.Context(), chatID, req.Message)
	if err != nil {
		http.Error(w, "failed to process workflow state", http.StatusInternalServerError)
		return
	}

	if active, ok, err := h.RunManager.FindActiveRunByChat(r.Context(), chatID); err != nil {
		http.Error(w, "failed to load active run", http.StatusInternalServerError)
		return
	} else if ok {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(createChatMessageResponse{
			ChatID: chatID,
			Run:    active,
		})
		return
	}
	if handledBlockedRun {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(createChatMessageResponse{
			ChatID: chatID,
			Run:    nil,
		})
		return
	}

	event, members, err := loadEventContext(h.SupabaseClient, eventID)
	if err != nil {
		http.Error(w, "failed to load event context", http.StatusInternalServerError)
		return
	}

	_ = services.UpdateEventContext(h.SupabaseClient, eventID, map[string]any{
		"last_agent_assessment": "Received a new workflow request from the requester chat.",
	})

	planningRun, err := h.RunManager.CreatePlanningRun(r.Context(), chatID, eventID, user.ID.String(), event.Context)
	if err != nil {
		http.Error(w, "failed to create run", http.StatusInternalServerError)
		return
	}

	planSummary, tasks, plannerPayload, err := services.BuildRunPlan(r.Context(), event, models.User{
		ID:   user.ID.String(),
		Name: user.Email,
	}, members, req.Message)
	if err != nil {
		http.Error(w, "failed to plan run", http.StatusInternalServerError)
		return
	}

	run, err := h.RunManager.FinalizeRunPlan(r.Context(), planningRun.ID, planSummary, tasks, plannerPayload)
	if err != nil {
		http.Error(w, "failed to persist plan", http.StatusInternalServerError)
		return
	}

	var lines []string
	if run.PlanSummary != "" {
		lines = append(lines, run.PlanSummary)
	}
	for _, task := range run.Tasks {
		lines = append(lines, strings.TrimSpace(task.Title))
	}
	approvalText := "Plan ready for approval."
	if len(lines) > 0 {
		approvalText = "Plan ready for approval:\n- " + strings.Join(lines, "\n- ")
	}
	_, _, _ = h.SupabaseClient.From("chat_message").Insert(map[string]any{
		"id":           uuid.NewString(),
		"chat_id":      chatID,
		"sender_type":  "agent",
		"message":      approvalText,
		"agent_run_id": run.ID,
		"message_type": "approval_request",
		"metadata": map[string]any{
			"task_count": len(run.Tasks),
		},
	}, false, "", "", "").Execute()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createChatMessageResponse{
		ChatID: chatID,
		Run:    run,
	})
}

func (h *ChatHandler) ApproveRun(w http.ResponseWriter, r *http.Request) {
	if _, err := authenticatedUserFromRequest(r); err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	runID := r.PathValue("runId")
	if err := h.RunManager.StartRun(r.Context(), runID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	run, err := h.RunManager.GetRun(r.Context(), runID)
	if err != nil {
		http.Error(w, "run not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(run)
}

func (h *ChatHandler) GetActiveRun(w http.ResponseWriter, r *http.Request) {
	user, err := authenticatedUserFromRequest(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	eventID := r.PathValue("id")
	var chats []models.Chat
	_, err = h.SupabaseClient.From("chat").
		Select("id, event_id, user_id, created_at", "", false).
		Eq("event_id", eventID).
		Eq("user_id", user.ID.String()).
		ExecuteTo(&chats)
	if err != nil {
		http.Error(w, "failed to resolve chat", http.StatusInternalServerError)
		return
	}
	if len(chats) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	chatID := chats[0].ID

	run, ok, err := h.RunManager.FindActiveRunByChat(r.Context(), chatID)
	if err != nil {
		http.Error(w, "failed to load active run", http.StatusInternalServerError)
		return
	}
	if !ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(run)
}

func (h *ChatHandler) StreamRun(w http.ResponseWriter, r *http.Request) {
	if _, err := authenticatedUserFromRequest(r); err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	runID := r.PathValue("runId")
	events, cancel, err := h.RunManager.Subscribe(r.Context(), runID)
	if err != nil {
		http.Error(w, "run not found", http.StatusNotFound)
		return
	}
	defer cancel()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	for {
		select {
		case <-r.Context().Done():
			return
		case event, ok := <-events:
			if !ok {
				return
			}
			payload, _ := json.Marshal(event)
			_, _ = w.Write([]byte("event: " + event.Type + "\n"))
			_, _ = w.Write([]byte("data: " + string(payload) + "\n\n"))
			flusher.Flush()
		}
	}
}

func loadEventContext(client *supabase.Client, eventID string) (models.Event, []services.EventMemberWithUser, error) {
	var events []models.Event
	_, err := client.From("event").
		Select("id, title, description, event_date, location, status, context", "", false).
		Eq("id", eventID).
		ExecuteTo(&events)
	if err != nil {
		return models.Event{}, nil, err
	}
	if len(events) == 0 {
		return models.Event{}, nil, errors.New("event not found")
	}

	var members []services.EventMemberWithUser
	_, err = client.From("event_member").
		Select("user_id, role, user(id, name)", "", false).
		Eq("event_id", eventID).
		ExecuteTo(&members)
	if err != nil {
		return models.Event{}, nil, err
	}

	return events[0], members, nil
}
