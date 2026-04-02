package models

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type Event struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	EventDate   string                 `json:"event_date"`
	Location    string                 `json:"location,omitempty"`
	Status      string                 `json:"status,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

type EventMember struct {
	UserID  string `json:"user_id"`
	EventID string `json:"event_id"`
	Role    string `json:"role"`
}

type EventMembersWithDetails struct {
	Role   string `json:"role"`
	UserID string `json:"user_id"`
	Event  *Event `json:"event"`
	User   *User  `json:"user"`
}

type Chat struct {
	ID        string `json:"id"`
	EventID   string `json:"event_id"`
	UserID    string `json:"user_id"`
	CreatedAt string `json:"created_at"`
}

type ChatMessage struct {
	ID          string                 `json:"id"`
	ChatID      string                 `json:"chat_id"`
	SenderType  string                 `json:"sender_type"`
	SenderID    *string                `json:"sender_id"`
	AgentRunID  *string                `json:"agent_run_id,omitempty"`
	AgentTaskID *string                `json:"agent_task_id,omitempty"`
	MessageType string                 `json:"message_type"`
	Message     string                 `json:"message"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   string                 `json:"created_at"`
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
	CreatedAt         string                 `json:"created_at"`
	UpdatedAt         string                 `json:"updated_at"`
}

type AgentTask struct {
	ID               string                 `json:"id"`
	RunID            string                 `json:"run_id"`
	Position         int                    `json:"position"`
	Title            string                 `json:"title"`
	Kind             string                 `json:"kind"`
	Status           string                 `json:"status"`
	TargetUserID     *string                `json:"target_user_id,omitempty"`
	TargetChatID     *string                `json:"target_chat_id,omitempty"`
	Instructions     string                 `json:"instructions,omitempty"`
	CompletionSignal string                 `json:"completion_signal,omitempty"`
	Result           string                 `json:"result,omitempty"`
	TaskPayload      map[string]interface{} `json:"task_payload,omitempty"`
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
}
type Post struct {
	ID        string `json:"id"`
	EventID   string `json:"event_id"`
	Content   string `json:"content"`
	Status    string `json:"status"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"`
}

type Media struct {
	ID          string                 `json:"id"`
	EventID     string                 `json:"event_id"`
	UploadedBy  string                 `json:"uploaded_by"`
	MediaType   string                 `json:"media_type"`
	StoragePath string                 `json:"storage_path"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   string                 `json:"created_at"`
}
