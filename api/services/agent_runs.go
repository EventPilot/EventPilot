package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"eventpilot/api/models"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/google/uuid"
	"github.com/supabase-community/supabase-go"
)

const (
	runStatusPlanning         = "planning"
	runStatusAwaitingApproval = "awaiting_approval"
	runStatusRunning          = "running"
	runStatusWaitingOnMember  = "waiting_on_member"
	runStatusCompleted        = "completed"
	runStatusFailed           = "failed"
	runStatusCancelled        = "cancelled"
)

type AgentTask struct {
	ID               string                 `json:"id"`
	RunID            string                 `json:"run_id,omitempty"`
	Position         int                    `json:"position"`
	Title            string                 `json:"title"`
	Kind             string                 `json:"kind"`
	Status           string                 `json:"status"`
	TargetUserID     *string                `json:"target_user_id,omitempty"`
	TargetChatID     *string                `json:"target_chat_id,omitempty"`
	TargetUserName   string                 `json:"target_user_name,omitempty"`
	TargetRole       string                 `json:"target_role,omitempty"`
	Instructions     string                 `json:"instructions,omitempty"`
	CompletionSignal string                 `json:"completion_signal,omitempty"`
	Result           string                 `json:"result,omitempty"`
	TaskPayload      map[string]interface{} `json:"task_payload,omitempty"`
	CreatedAt        time.Time              `json:"created_at,omitempty"`
	UpdatedAt        time.Time              `json:"updated_at,omitempty"`
	StartedAt        *time.Time             `json:"started_at,omitempty"`
	CompletedAt      *time.Time             `json:"completed_at,omitempty"`
	FailedAt         *time.Time             `json:"failed_at,omitempty"`
}

type AgentRun struct {
	ID                string                 `json:"id"`
	ChatID            string                 `json:"chat_id"`
	EventID           string                 `json:"event_id"`
	RequestedByUserID string                 `json:"requested_by_user_id"`
	Status            string                 `json:"status"`
	PlanSummary       string                 `json:"plan_summary"`
	BlockedOnChatID   string                 `json:"blocked_on_chat_id,omitempty"`
	CurrentTaskIndex  int                    `json:"current_task_index"`
	ContextSnapshot   map[string]interface{} `json:"context_snapshot,omitempty"`
	PlannerResponse   map[string]interface{} `json:"planner_response,omitempty"`
	Tasks             []AgentTask            `json:"tasks"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	StartedAt         *time.Time             `json:"started_at,omitempty"`
	CompletedAt       *time.Time             `json:"completed_at,omitempty"`
	FailedAt          *time.Time             `json:"failed_at,omitempty"`
}

type RunEvent struct {
	Type string    `json:"type"`
	Run  *AgentRun `json:"run"`
}

type QueuedChatMessage struct {
	ID        string                 `json:"id"`
	ChatID    string                 `json:"chat_id"`
	SenderID  string                 `json:"sender_id"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
}

type plannerResponse struct {
	Summary string        `json:"summary"`
	Tasks   []plannerTask `json:"tasks"`
}

type plannerTask struct {
	Title            string `json:"title"`
	Kind             string `json:"kind"`
	TargetUserID     string `json:"target_user_id,omitempty"`
	Instructions     string `json:"instructions,omitempty"`
	CompletionSignal string `json:"completion_signal,omitempty"`
}

type agentTaskInsertRecord struct {
	ID               string                 `json:"id"`
	RunID            string                 `json:"run_id"`
	Position         int                    `json:"position"`
	Title            string                 `json:"title"`
	Kind             string                 `json:"kind"`
	Status           string                 `json:"status"`
	TargetUserID     *string                `json:"target_user_id"`
	TargetChatID     *string                `json:"target_chat_id"`
	Instructions     string                 `json:"instructions"`
	CompletionSignal string                 `json:"completion_signal"`
	Result           string                 `json:"result"`
	TaskPayload      map[string]interface{} `json:"task_payload"`
}

type EventMemberWithUser struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	User   *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"user"`
}

type RunManager struct {
	client      *supabase.Client
	mu          sync.RWMutex
	subscribers map[string]map[chan RunEvent]struct{}
}

func NewRunManager(client *supabase.Client) *RunManager {
	return &RunManager{
		client:      client,
		subscribers: make(map[string]map[chan RunEvent]struct{}),
	}
}

func (m *RunManager) Subscribe(ctx context.Context, runID string) (<-chan RunEvent, func(), error) {
	run, err := m.GetRun(ctx, runID)
	if err != nil {
		return nil, nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	ch := make(chan RunEvent, 8)
	if m.subscribers[runID] == nil {
		m.subscribers[runID] = make(map[chan RunEvent]struct{})
	}
	m.subscribers[runID][ch] = struct{}{}
	ch <- RunEvent{Type: "run_state", Run: cloneRun(run)}

	cancel := func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		if subs, ok := m.subscribers[runID]; ok {
			delete(subs, ch)
			if len(subs) == 0 {
				delete(m.subscribers, runID)
			}
		}
		close(ch)
	}

	return ch, cancel, nil
}

func (m *RunManager) CreatePlanningRun(ctx context.Context, chatID, eventID, requestedByUserID string, contextSnapshot map[string]interface{}) (*AgentRun, error) {
	now := time.Now().UTC()
	run := &AgentRun{
		ID:                uuid.NewString(),
		ChatID:            chatID,
		EventID:           eventID,
		RequestedByUserID: requestedByUserID,
		Status:            runStatusPlanning,
		PlanSummary:       "",
		CurrentTaskIndex:  0,
		ContextSnapshot:   cloneMap(contextSnapshot),
		PlannerResponse:   map[string]interface{}{},
		Tasks:             []AgentTask{},
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	insert := map[string]any{
		"id":                   run.ID,
		"chat_id":              run.ChatID,
		"event_id":             run.EventID,
		"requested_by_user_id": run.RequestedByUserID,
		"status":               run.Status,
		"plan_summary":         run.PlanSummary,
		"current_task_index":   run.CurrentTaskIndex,
		"context_snapshot":     run.ContextSnapshot,
		"planner_response":     run.PlannerResponse,
	}

	_, _, err := m.client.From("agent_run").Insert(insert, false, "", "", "").Execute()
	if err != nil {
		return nil, err
	}
	m.publish(run.ID, "run_state", run)
	return run, nil
}

func (m *RunManager) FinalizeRunPlan(ctx context.Context, runID string, planSummary string, tasks []AgentTask, plannerPayload map[string]interface{}) (*AgentRun, error) {
	for i := range tasks {
		tasks[i].RunID = runID
	}
	if len(tasks) > 0 {
		records := make([]agentTaskInsertRecord, 0, len(tasks))
		for _, task := range tasks {
			record := agentTaskInsertRecord{
				ID:               task.ID,
				RunID:            runID,
				Position:         task.Position,
				Title:            task.Title,
				Kind:             task.Kind,
				Status:           task.Status,
				TargetUserID:     task.TargetUserID,
				TargetChatID:     task.TargetChatID,
				Instructions:     task.Instructions,
				CompletionSignal: task.CompletionSignal,
				Result:           task.Result,
				TaskPayload:      emptyJSON(task.TaskPayload),
			}
			records = append(records, record)
		}
		_, _, err := m.client.From("agent_task").Insert(records, false, "", "", "").Execute()
		if err != nil {
			_ = m.markRunFailed(ctx, runID, fmt.Sprintf("failed to persist tasks: %v", err))
			return nil, err
		}
	}

	update := map[string]any{
		"status":           runStatusAwaitingApproval,
		"plan_summary":     strings.TrimSpace(planSummary),
		"planner_response": emptyJSON(plannerPayload),
	}
	if _, _, err := m.client.From("agent_run").Update(update, "", "").Eq("id", runID).Execute(); err != nil {
		return nil, err
	}

	run, err := m.GetRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	m.publish(runID, "run_state", run)
	return run, nil
}

func (m *RunManager) GetRun(ctx context.Context, runID string) (*AgentRun, error) {
	var runs []AgentRun
	_, err := m.client.From("agent_run").
		Select("id, chat_id, event_id, requested_by_user_id, status, plan_summary, blocked_on_chat_id, current_task_index, context_snapshot, planner_response, created_at, updated_at, started_at, completed_at, failed_at", "", false).
		Eq("id", runID).
		ExecuteTo(&runs)
	if err != nil {
		return nil, err
	}
	if len(runs) == 0 {
		return nil, errors.New("run not found")
	}

	run := runs[0]
	tasks, err := m.listTasksByRun(ctx, run.ID)
	if err != nil {
		return nil, err
	}
	run.Tasks = tasks
	if err := m.enrichTaskTargets(ctx, run.EventID, run.Tasks); err != nil {
		return nil, err
	}
	return &run, nil
}

func (m *RunManager) FindActiveRunByChat(ctx context.Context, chatID string) (*AgentRun, bool, error) {
	var runs []AgentRun
	_, err := m.client.From("agent_run").
		Select("id, chat_id, event_id, requested_by_user_id, status, plan_summary, blocked_on_chat_id, current_task_index, context_snapshot, planner_response, created_at, updated_at, started_at, completed_at, failed_at", "", false).
		Eq("chat_id", chatID).
		ExecuteTo(&runs)
	if err != nil {
		return nil, false, err
	}

	active := make([]AgentRun, 0, len(runs))
	for _, run := range runs {
		if isActiveRunStatus(run.Status) {
			active = append(active, run)
		}
	}
	if len(active) == 0 {
		return nil, false, nil
	}
	sort.Slice(active, func(i, j int) bool {
		return active[i].UpdatedAt.After(active[j].UpdatedAt)
	})
	run, err := m.GetRun(ctx, active[0].ID)
	if err != nil {
		return nil, false, err
	}
	return run, true, nil
}

func (m *RunManager) StartRun(ctx context.Context, runID string) error {
	run, err := m.GetRun(ctx, runID)
	if err != nil {
		return err
	}
	if run.Status != runStatusAwaitingApproval {
		return errors.New("run is not awaiting approval")
	}

	now := time.Now().UTC()
	update := map[string]any{
		"status":       runStatusRunning,
		"started_at":   now,
		"failed_at":    nil,
		"completed_at": nil,
	}
	if _, _, err := m.client.From("agent_run").Update(update, "", "").Eq("id", runID).Execute(); err != nil {
		return err
	}
	updated, err := m.GetRun(ctx, runID)
	if err != nil {
		return err
	}
	m.publish(runID, "run_state", updated)
	go m.executeRun(context.Background(), runID)
	return nil
}

func (m *RunManager) HandleIncomingMessage(ctx context.Context, chatID string, message string) (bool, error) {
	waitingRuns, err := m.loadWaitingRunsByBlockedChat(ctx, chatID)
	if err != nil {
		return false, err
	}
	if len(waitingRuns) == 0 {
		return false, nil
	}

	handled := false
	for _, run := range waitingRuns {
		task := waitingTask(run.Tasks)
		if task == nil {
			continue
		}

		complete, followup := evaluateCompletion(ctx, *task, message)
		if !complete {
			if strings.TrimSpace(followup) != "" {
				if err := m.insertChatMessage(chatID, "agent", nil, followup, "member_request", run.ID, task.ID, map[string]any{
					"follow_up": true,
				}); err != nil {
					return handled, err
				}
			}
			_ = UpdateEventContext(m.client, run.EventID, map[string]any{
				"last_agent_assessment": fmt.Sprintf("Still waiting on required member input for task: %s", task.Title),
				"missing_inputs":        []string{task.Title},
			})
			handled = true
			continue
		}

		now := time.Now().UTC()
		if _, _, err := m.client.From("agent_task").Update(map[string]any{
			"status":       "completed",
			"result":       "Required teammate input received.",
			"completed_at": now,
			"failed_at":    nil,
		}, "", "").Eq("id", task.ID).Execute(); err != nil {
			return handled, err
		}
		if _, _, err := m.client.From("agent_run").Update(map[string]any{
			"status":             runStatusRunning,
			"blocked_on_chat_id": nil,
			"current_task_index": task.Position + 1,
		}, "", "").Eq("id", run.ID).Execute(); err != nil {
			return handled, err
		}
		_ = UpdateEventContext(m.client, run.EventID, completionContextUpdate(*task))
		updated, err := m.GetRun(ctx, run.ID)
		if err != nil {
			return handled, err
		}
		m.publish(run.ID, "task_state", updated)
		go m.executeRun(context.Background(), run.ID)
		handled = true
	}

	return handled, nil
}

func (m *RunManager) QueueRequesterMessage(ctx context.Context, messageID string) error {
	return m.updateQueuedMessageMetadata(ctx, messageID, map[string]any{
		"workflow_state":        "queued",
		"queued_for_active_run": true,
	})
}

func (m *RunManager) updateQueuedMessageMetadata(ctx context.Context, messageID string, fields map[string]any) error {
	var rows []struct {
		Metadata map[string]interface{} `json:"metadata"`
	}
	_, err := m.client.From("chat_message").
		Select("metadata", "", false).
		Eq("id", messageID).
		ExecuteTo(&rows)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return errors.New("chat message not found")
	}

	metadata := cloneMap(rows[0].Metadata)
	for key, value := range fields {
		metadata[key] = value
	}

	_, _, err = m.client.From("chat_message").
		Update(map[string]any{"metadata": emptyJSON(metadata)}, "", "").
		Eq("id", messageID).
		Execute()
	return err
}

func (m *RunManager) PromoteNextQueuedMessage(ctx context.Context, chatID, eventID, requesterUserID string) (*AgentRun, bool, error) {
	queued, ok, err := m.findNextQueuedRequesterMessage(ctx, chatID)
	if err != nil || !ok {
		return nil, ok, err
	}

	event, members, err := m.loadEventContextForPlanning(ctx, eventID)
	if err != nil {
		return nil, false, err
	}
	requester, err := m.loadRequester(ctx, requesterUserID)
	if err != nil {
		return nil, false, err
	}

	planningRun, err := m.CreatePlanningRun(ctx, chatID, eventID, requesterUserID, event.Context)
	if err != nil {
		return nil, false, err
	}

	planSummary, tasks, plannerPayload, err := BuildRunPlan(ctx, event, requester, members, queued.Message)
	if err != nil {
		return nil, false, err
	}

	run, err := m.FinalizeRunPlan(ctx, planningRun.ID, planSummary, tasks, plannerPayload)
	if err != nil {
		return nil, false, err
	}

	if err := m.updateQueuedMessageMetadata(ctx, queued.ID, map[string]any{
		"workflow_state":        "consumed",
		"queued_for_active_run": false,
	}); err != nil {
		return nil, false, err
	}
	if _, _, err := m.client.From("chat_message").
		Update(map[string]any{"agent_run_id": run.ID}, "", "").
		Eq("id", queued.ID).
		Execute(); err != nil {
		return nil, false, err
	}

	if err := m.InsertApprovalRequestMessage(ctx, chatID, run); err != nil {
		return nil, false, err
	}

	return run, true, nil
}

func (m *RunManager) InsertApprovalRequestMessage(ctx context.Context, chatID string, run *AgentRun) error {
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

	return m.insertChatMessage(chatID, "agent", nil, approvalText, "approval_request", run.ID, "", map[string]any{
		"task_count": len(run.Tasks),
	})
}

func (m *RunManager) findNextQueuedRequesterMessage(ctx context.Context, chatID string) (*QueuedChatMessage, bool, error) {
	var rows []QueuedChatMessage
	_, err := m.client.From("chat_message").
		Select("id, chat_id, sender_id, message, metadata, created_at", "", false).
		Eq("chat_id", chatID).
		Eq("sender_type", "user").
		Eq("message_type", "message").
		ExecuteTo(&rows)
	if err != nil {
		return nil, false, err
	}

	queued := make([]QueuedChatMessage, 0, len(rows))
	for _, row := range rows {
		if workflowState(row.Metadata) == "queued" {
			queued = append(queued, row)
		}
	}
	if len(queued) == 0 {
		return nil, false, nil
	}

	sort.Slice(queued, func(i, j int) bool {
		return queued[i].CreatedAt.Before(queued[j].CreatedAt)
	})
	return &queued[0], true, nil
}

func (m *RunManager) loadEventContextForPlanning(ctx context.Context, eventID string) (models.Event, []EventMemberWithUser, error) {
	var events []models.Event
	_, err := m.client.From("event").
		Select("id, title, description, event_date, location, status, context", "", false).
		Eq("id", eventID).
		ExecuteTo(&events)
	if err != nil {
		return models.Event{}, nil, err
	}
	if len(events) == 0 {
		return models.Event{}, nil, errors.New("event not found")
	}

	var members []EventMemberWithUser
	_, err = m.client.From("event_member").
		Select("user_id, role, user(id, name)", "", false).
		Eq("event_id", eventID).
		ExecuteTo(&members)
	if err != nil {
		return models.Event{}, nil, err
	}
	return events[0], members, nil
}

func (m *RunManager) loadRequester(ctx context.Context, requesterUserID string) (models.User, error) {
	var users []models.User
	_, err := m.client.From("user").
		Select("id, name, role", "", false).
		Eq("id", requesterUserID).
		ExecuteTo(&users)
	if err != nil {
		return models.User{}, err
	}
	if len(users) == 0 {
		return models.User{ID: requesterUserID, Name: "Requester"}, nil
	}
	if strings.TrimSpace(users[0].Name) == "" {
		users[0].Name = "Requester"
	}
	return users[0], nil
}

func (m *RunManager) executeRun(ctx context.Context, runID string) {
	for {
		run, err := m.GetRun(ctx, runID)
		if err != nil {
			return
		}
		if run.Status != runStatusRunning {
			return
		}

		task := nextPendingTask(run.Tasks)
		if task == nil {
			if _, _, err := m.client.From("agent_run").Update(map[string]any{
				"status":             runStatusCompleted,
				"completed_at":       time.Now().UTC(),
				"blocked_on_chat_id": nil,
			}, "", "").Eq("id", run.ID).Execute(); err != nil {
				return
			}
			updated, err := m.GetRun(ctx, run.ID)
			if err == nil {
				m.publish(run.ID, "completed", updated)
			}
			_, _, _ = m.PromoteNextQueuedMessage(ctx, run.ChatID, run.EventID, run.RequestedByUserID)
			return
		}

		now := time.Now().UTC()
		if _, _, err := m.client.From("agent_task").Update(map[string]any{
			"status":     "in_progress",
			"started_at": now,
		}, "", "").Eq("id", task.ID).Execute(); err != nil {
			_ = m.markTaskAndRunFailed(ctx, run.ID, task.ID, err.Error())
			return
		}
		if _, _, err := m.client.From("agent_run").Update(map[string]any{
			"current_task_index": task.Position,
		}, "", "").Eq("id", run.ID).Execute(); err != nil {
			_ = m.markTaskAndRunFailed(ctx, run.ID, task.ID, err.Error())
			return
		}
		updated, err := m.GetRun(ctx, run.ID)
		if err == nil {
			m.publish(run.ID, "task_state", updated)
		}

		switch task.Kind {
		case "ask_member":
			if task.TargetUserID == nil || *task.TargetUserID == "" {
				_ = m.markTaskAndRunFailed(ctx, run.ID, task.ID, "No target event member available.")
				return
			}

			targetChatID, err := EnsureChatForUser(m.client, run.EventID, *task.TargetUserID)
			if err != nil {
				_ = m.markTaskAndRunFailed(ctx, run.ID, task.ID, err.Error())
				return
			}

			prompt := strings.TrimSpace(task.Instructions)
			if prompt == "" {
				prompt = fmt.Sprintf("EventPilot needs %s from you before we can continue.", task.Title)
			}

			if err := m.insertChatMessage(targetChatID, "agent", nil, prompt, "member_request", run.ID, task.ID, map[string]any{
				"source": "agent_task",
			}); err != nil {
				_ = m.markTaskAndRunFailed(ctx, run.ID, task.ID, err.Error())
				return
			}

			if _, _, err := m.client.From("agent_task").Update(map[string]any{
				"status":         "waiting",
				"target_chat_id": targetChatID,
				"result":         "Waiting for teammate response.",
			}, "", "").Eq("id", task.ID).Execute(); err != nil {
				_ = m.markTaskAndRunFailed(ctx, run.ID, task.ID, err.Error())
				return
			}
			if _, _, err := m.client.From("agent_run").Update(map[string]any{
				"status":             runStatusWaitingOnMember,
				"blocked_on_chat_id": targetChatID,
				"current_task_index": task.Position,
			}, "", "").Eq("id", run.ID).Execute(); err != nil {
				_ = m.markTaskAndRunFailed(ctx, run.ID, task.ID, err.Error())
				return
			}
			waitingRun, err := m.GetRun(ctx, run.ID)
			if err == nil {
				m.publish(run.ID, "blocked", waitingRun)
			}
			return
		default:
			result := strings.TrimSpace(task.Result)
			if result == "" {
				result = "Completed."
			}
			if _, _, err := m.client.From("agent_task").Update(map[string]any{
				"status":       "completed",
				"result":       result,
				"completed_at": time.Now().UTC(),
			}, "", "").Eq("id", task.ID).Execute(); err != nil {
				_ = m.markTaskAndRunFailed(ctx, run.ID, task.ID, err.Error())
				return
			}
			_ = UpdateEventContext(m.client, run.EventID, map[string]any{
				"last_agent_assessment": fmt.Sprintf("Completed task: %s", task.Title),
			})
			_ = m.insertChatMessage(run.ChatID, "agent", nil, fmt.Sprintf("Completed task: %s", task.Title), "task_update", run.ID, task.ID, map[string]any{
				"task_kind": task.Kind,
			})
			updated, err := m.GetRun(ctx, run.ID)
			if err == nil {
				m.publish(run.ID, "task_state", updated)
			}
		}
	}
}

func (m *RunManager) publish(runID string, eventType string, run *AgentRun) {
	if run == nil {
		return
	}
	m.mu.RLock()
	subs := m.subscribers[runID]
	m.mu.RUnlock()

	event := RunEvent{Type: eventType, Run: cloneRun(run)}
	for ch := range subs {
		select {
		case ch <- event:
		default:
		}
	}
}

func (m *RunManager) listTasksByRun(ctx context.Context, runID string) ([]AgentTask, error) {
	var tasks []AgentTask
	_, err := m.client.From("agent_task").
		Select("id, run_id, position, title, kind, status, target_user_id, target_chat_id, instructions, completion_signal, result, task_payload, created_at, updated_at, started_at, completed_at, failed_at", "", false).
		Eq("run_id", runID).
		ExecuteTo(&tasks)
	if err != nil {
		return nil, err
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Position < tasks[j].Position
	})
	return tasks, nil
}

func (m *RunManager) enrichTaskTargets(ctx context.Context, eventID string, tasks []AgentTask) error {
	targetIDs := make(map[string]struct{})
	for _, task := range tasks {
		if task.TargetUserID != nil && *task.TargetUserID != "" {
			targetIDs[*task.TargetUserID] = struct{}{}
		}
	}
	if len(targetIDs) == 0 {
		return nil
	}

	var members []EventMemberWithUser
	_, err := m.client.From("event_member").
		Select("user_id, role, user(id, name)", "", false).
		Eq("event_id", eventID).
		ExecuteTo(&members)
	if err != nil {
		return err
	}

	memberByID := make(map[string]EventMemberWithUser, len(members))
	for _, member := range members {
		memberByID[member.UserID] = member
	}
	for i := range tasks {
		if tasks[i].TargetUserID == nil {
			continue
		}
		member, ok := memberByID[*tasks[i].TargetUserID]
		if !ok {
			continue
		}
		tasks[i].TargetRole = member.Role
		if member.User != nil {
			tasks[i].TargetUserName = member.User.Name
		}
	}
	return nil
}

func (m *RunManager) loadWaitingRunsByBlockedChat(ctx context.Context, chatID string) ([]*AgentRun, error) {
	var runs []AgentRun
	_, err := m.client.From("agent_run").
		Select("id, chat_id, event_id, requested_by_user_id, status, plan_summary, blocked_on_chat_id, current_task_index, context_snapshot, planner_response, created_at, updated_at, started_at, completed_at, failed_at", "", false).
		Eq("blocked_on_chat_id", chatID).
		Eq("status", runStatusWaitingOnMember).
		ExecuteTo(&runs)
	if err != nil {
		return nil, err
	}
	out := make([]*AgentRun, 0, len(runs))
	for _, run := range runs {
		loaded, err := m.GetRun(ctx, run.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, loaded)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	return out, nil
}

func (m *RunManager) markTaskAndRunFailed(ctx context.Context, runID string, taskID string, reason string) error {
	now := time.Now().UTC()
	if taskID != "" {
		_, _, _ = m.client.From("agent_task").Update(map[string]any{
			"status":    "failed",
			"result":    reason,
			"failed_at": now,
		}, "", "").Eq("id", taskID).Execute()
	}
	return m.markRunFailed(ctx, runID, reason)
}

func (m *RunManager) markRunFailed(ctx context.Context, runID string, reason string) error {
	_, _, err := m.client.From("agent_run").Update(map[string]any{
		"status":       runStatusFailed,
		"failed_at":    time.Now().UTC(),
		"plan_summary": reason,
	}, "", "").Eq("id", runID).Execute()
	if err != nil {
		return err
	}
	run, err := m.GetRun(ctx, runID)
	if err == nil {
		m.publish(runID, "run_state", run)
	}
	return err
}

func (m *RunManager) insertChatMessage(chatID string, senderType string, senderID *string, message string, messageType string, runID string, taskID string, metadata map[string]any) error {
	record := map[string]any{
		"id":           uuid.NewString(),
		"chat_id":      chatID,
		"sender_type":  senderType,
		"message":      message,
		"message_type": messageType,
		"metadata":     emptyJSON(metadata),
	}
	if senderID != nil && *senderID != "" {
		record["sender_id"] = *senderID
	}
	if runID != "" {
		record["agent_run_id"] = runID
	}
	if taskID != "" {
		record["agent_task_id"] = taskID
	}
	_, _, err := m.client.From("chat_message").Insert(record, false, "", "", "").Execute()
	return err
}

func workflowState(metadata map[string]interface{}) string {
	if metadata == nil {
		return ""
	}
	value, ok := metadata["workflow_state"]
	if !ok {
		return ""
	}
	if state, ok := value.(string); ok {
		return strings.TrimSpace(state)
	}
	return ""
}

func BuildRunPlan(ctx context.Context, event models.Event, requester models.User, members []EventMemberWithUser, message string) (string, []AgentTask, map[string]interface{}, error) {
	resp, err := planTasks(ctx, event, requester, members, message)
	if err != nil {
		return "", nil, nil, err
	}

	tasks := make([]AgentTask, 0, len(resp.Tasks))
	for idx, task := range resp.Tasks {
		kind := normalizeTaskKind(task.Kind)
		payload := map[string]interface{}{
			"title":             strings.TrimSpace(task.Title),
			"kind":              kind,
			"instructions":      strings.TrimSpace(task.Instructions),
			"completion_signal": strings.TrimSpace(task.CompletionSignal),
		}
		t := AgentTask{
			ID:               uuid.NewString(),
			Position:         idx,
			Title:            strings.TrimSpace(task.Title),
			Kind:             kind,
			Status:           "pending",
			Instructions:     strings.TrimSpace(task.Instructions),
			CompletionSignal: strings.TrimSpace(task.CompletionSignal),
			TaskPayload:      payload,
		}
		if t.Title == "" {
			t.Title = "Continue agent workflow"
			t.TaskPayload["title"] = t.Title
		}
		if kind == "ask_member" && task.TargetUserID != "" {
			targetID := task.TargetUserID
			t.TargetUserID = &targetID
			t.TaskPayload["target_user_id"] = targetID
			for _, member := range members {
				if member.UserID == targetID {
					t.TargetRole = member.Role
					if member.User != nil {
						t.TargetUserName = member.User.Name
					}
					break
				}
			}
		}
		tasks = append(tasks, t)
	}

	if len(tasks) == 0 {
		tasks = []AgentTask{{
			ID:          uuid.NewString(),
			Position:    0,
			Title:       "Review the request and prepare the next post update",
			Kind:        "internal",
			Status:      "pending",
			TaskPayload: map[string]interface{}{"title": "Review the request and prepare the next post update", "kind": "internal"},
		}}
	}

	payload := map[string]interface{}{
		"summary": resp.Summary,
		"tasks":   tasksToPlannerPayload(tasks),
	}
	return strings.TrimSpace(resp.Summary), tasks, payload, nil
}

func normalizeTaskKind(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "ask_member", "request_member", "wait_for_member":
		return "ask_member"
	default:
		return "internal"
	}
}

func planTasks(ctx context.Context, event models.Event, requester models.User, members []EventMemberWithUser, message string) (*plannerResponse, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Printf("[agent_runs.planTasks] using fallback plan: ANTHROPIC_API_KEY is not set for event=%s requester=%s", event.ID, requester.ID)
		return fallbackPlan(members, message), nil
	}

	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	memberJSON, _ := json.Marshal(members)
	contextJSON, _ := json.Marshal(event.Context)
	userPrompt := fmt.Sprintf(`Return JSON with fields "summary" and "tasks".
Each task must have: "title", "kind", optional "target_user_id", optional "instructions", optional "completion_signal".
Only use kind values "internal" or "ask_member".
Ask another member only if the current request clearly requires information or assets from them.

Event:
- id: %s
- title: %s
- description: %s
- event_date: %s
- location: %s
- context: %s

Requester:
- id: %s
- name: %s

Available event members:
%s

Latest user message:
%s`, event.ID, event.Title, event.Description, event.EventDate, event.Location, string(contextJSON), requester.ID, requester.Name, string(memberJSON), message)

	msg, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5_20251001,
		MaxTokens: 600,
		System: []anthropic.TextBlockParam{{
			Text: `You are planning the next actions for EventPilot.
Return only valid JSON. Keep the plan short and executable.`,
		}},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
		},
	})
	if err != nil {
		log.Printf("[agent_runs.planTasks] Claude planner failed for event=%s requester=%s: %v", event.ID, requester.ID, err)
		log.Printf("[agent_runs.planTasks] using fallback plan after Claude error for event=%s requester=%s", event.ID, requester.ID)
		return fallbackPlan(members, message), nil
	}

	var text strings.Builder
	for _, block := range msg.Content {
		if block.Type == "text" {
			text.WriteString(block.Text)
		}
	}
	raw := strings.TrimSpace(text.String())
	normalized := extractJSONPayload(raw)

	var resp plannerResponse
	if err := json.Unmarshal([]byte(normalized), &resp); err != nil {
		log.Printf("[agent_runs.planTasks] failed to parse Claude planner response for event=%s requester=%s: %v; raw=%q", event.ID, requester.ID, err, raw)
		log.Printf("[agent_runs.planTasks] using fallback plan after invalid Claude JSON for event=%s requester=%s", event.ID, requester.ID)
		return fallbackPlan(members, message), nil
	}
	if strings.TrimSpace(resp.Summary) == "" {
		log.Printf("[agent_runs.planTasks] Claude planner returned empty summary for event=%s requester=%s; using default summary", event.ID, requester.ID)
		resp.Summary = "Review the request and work through the next event tasks."
	}
	log.Printf("[agent_runs.planTasks] Claude planner produced %d task(s) for event=%s requester=%s", len(resp.Tasks), event.ID, requester.ID)
	return &resp, nil
}

func extractJSONPayload(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "```") {
		lines := strings.Split(trimmed, "\n")
		if len(lines) >= 2 {
			lines = lines[1:]
		}
		if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "```" {
			lines = lines[:len(lines)-1]
		}
		trimmed = strings.TrimSpace(strings.Join(lines, "\n"))
	}
	return trimmed
}

func fallbackPlan(members []EventMemberWithUser, message string) *plannerResponse {
	response := &plannerResponse{
		Summary: "Review the latest request, gather any missing event inputs, and update the draft workflow.",
		Tasks: []plannerTask{
			{Title: "Review the latest request and update the working plan", Kind: "internal"},
		},
	}
	lower := strings.ToLower(message)
	if strings.Contains(lower, "photo") || strings.Contains(lower, "image") || strings.Contains(lower, "video") {
		for _, member := range members {
			role := strings.ToLower(member.Role)
			if strings.Contains(role, "photo") {
				response.Tasks = append(response.Tasks, plannerTask{
					Title:            "Request media from the event photographer",
					Kind:             "ask_member",
					TargetUserID:     member.UserID,
					Instructions:     "Please share the event photos or videos we should use for this post.",
					CompletionSignal: "The response includes the requested media or confirms where it was uploaded.",
				})
				break
			}
		}
	}
	response.Tasks = append(response.Tasks, plannerTask{
		Title: "Prepare the next post update once required inputs are available",
		Kind:  "internal",
	})
	return response
}

func evaluateCompletion(_ context.Context, task AgentTask, message string) (bool, string) {
	lower := strings.ToLower(message)
	if task.CompletionSignal != "" && strings.Contains(strings.ToLower(task.CompletionSignal), "media") {
		if strings.Contains(lower, "photo") || strings.Contains(lower, "image") || strings.Contains(lower, "video") || strings.Contains(lower, "upload") {
			return true, ""
		}
		return false, "I still need the actual media or a clear upload confirmation before the workflow can continue."
	}
	if len(strings.Fields(strings.TrimSpace(message))) >= 5 {
		return true, ""
	}
	return false, "I need a bit more detail before I can continue the workflow."
}

func EnsureChatForUser(client *supabase.Client, eventID string, userID string) (string, error) {
	var chats []models.Chat
	_, err := client.From("chat").
		Select("id, event_id, user_id, created_at", "", false).
		Eq("event_id", eventID).
		Eq("user_id", userID).
		ExecuteTo(&chats)
	if err != nil {
		return "", err
	}
	if len(chats) > 0 {
		return chats[0].ID, nil
	}

	newChat := map[string]any{
		"id":       uuid.NewString(),
		"event_id": eventID,
		"user_id":  userID,
	}
	var created []models.Chat
	_, err = client.From("chat").
		Insert(newChat, false, "", "", "").
		ExecuteTo(&created)
	if err != nil {
		return "", err
	}
	if len(created) == 0 {
		return newChat["id"].(string), nil
	}
	return created[0].ID, nil
}

func UpdateEventContext(client *supabase.Client, eventID string, delta map[string]any) error {
	var events []models.Event
	_, err := client.From("event").
		Select("id, context", "", false).
		Eq("id", eventID).
		ExecuteTo(&events)
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return errors.New("event not found")
	}

	contextMap := cloneMap(events[0].Context)
	for key, value := range delta {
		contextMap[key] = value
	}
	update := map[string]any{
		"context":            contextMap,
		"context_updated_at": time.Now().UTC(),
	}
	_, _, err = client.From("event").Update(update, "", "").Eq("id", eventID).Execute()
	return err
}

func waitingTask(tasks []AgentTask) *AgentTask {
	for i := range tasks {
		if tasks[i].Status == "waiting" {
			return &tasks[i]
		}
	}
	return nil
}

func nextPendingTask(tasks []AgentTask) *AgentTask {
	for i := range tasks {
		if tasks[i].Status == "pending" {
			return &tasks[i]
		}
	}
	return nil
}

func isActiveRunStatus(status string) bool {
	switch status {
	case runStatusPlanning, runStatusAwaitingApproval, runStatusRunning, runStatusWaitingOnMember:
		return true
	default:
		return false
	}
}

func completionContextUpdate(task AgentTask) map[string]any {
	update := map[string]any{
		"last_agent_assessment": fmt.Sprintf("Completed waiting task: %s", task.Title),
	}
	if task.CompletionSignal != "" && strings.Contains(strings.ToLower(task.CompletionSignal), "media") {
		update["media_status"] = "member_provided_media"
	}
	return update
}

func tasksToPlannerPayload(tasks []AgentTask) []map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(tasks))
	for _, task := range tasks {
		item := map[string]interface{}{
			"id":                task.ID,
			"position":          task.Position,
			"title":             task.Title,
			"kind":              task.Kind,
			"status":            task.Status,
			"instructions":      task.Instructions,
			"completion_signal": task.CompletionSignal,
			"task_payload":      emptyJSON(task.TaskPayload),
		}
		if task.TargetUserID != nil {
			item["target_user_id"] = *task.TargetUserID
		}
		items = append(items, item)
	}
	return items
}

func emptyJSON(value map[string]interface{}) map[string]interface{} {
	if value == nil {
		return map[string]interface{}{}
	}
	return value
}

func cloneMap(value map[string]interface{}) map[string]interface{} {
	if value == nil {
		return map[string]interface{}{}
	}
	clone := make(map[string]interface{}, len(value))
	for key, item := range value {
		clone[key] = item
	}
	return clone
}

func cloneRun(run *AgentRun) *AgentRun {
	if run == nil {
		return nil
	}
	clone := *run
	if run.ContextSnapshot != nil {
		clone.ContextSnapshot = cloneMap(run.ContextSnapshot)
	}
	if run.PlannerResponse != nil {
		clone.PlannerResponse = cloneMap(run.PlannerResponse)
	}
	clone.Tasks = make([]AgentTask, len(run.Tasks))
	for i, task := range run.Tasks {
		clone.Tasks[i] = task
		if task.TaskPayload != nil {
			clone.Tasks[i].TaskPayload = cloneMap(task.TaskPayload)
		}
	}
	return &clone
}
