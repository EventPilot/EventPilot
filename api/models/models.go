package models

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type Event struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	EventDate   string `json:"event_date"`
}

type EventOwner struct {
	UserID  int `json:"user_id"`
	EventID int `json:"event_id"`
}

type Chat struct {
	ID        int    `json:"id"`
	EventID   int    `json:"event_id"`
	CreatedAt string `json:"created_at"`
}

type ChatMessage struct {
	ID         int    `json:"id"`
	ChatID     int    `json:"chat_id"`
	SenderType string `json:"sender_type"`
	SenderID   *int   `json:"sender_id"`
	Message    string `json:"message"`
	CreatedAt  string `json:"created_at"`
}

type Post struct {
	ID        int    `json:"id"`
	EventID   int    `json:"event_id"`
	Content   string `json:"content"`
	Status    string `json:"status"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"`
}

type Media struct {
	ID          int                    `json:"id"`
	EventID     int                    `json:"event_id"`
	UploadedBy  int                    `json:"uploaded_by"`
	MediaType   string                 `json:"media_type"`
	StoragePath string                 `json:"storage_path"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   string                 `json:"created_at"`
}
