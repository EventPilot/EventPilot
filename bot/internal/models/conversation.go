package models

// ConversationState tracks the progress of an event documentation conversation
type ConversationState struct {
	EventID          string            `json:"event_id"`
	CollectedData    map[string]string `json:"collected_data"`
	MissingFields    []string          `json:"missing_fields"`
	CurrentQuestion  string            `json:"current_question"`
	ConversationStep int               `json:"conversation_step"`
	IsComplete       bool              `json:"is_complete"`
}

// NewConversationState creates a new conversation state
func NewConversationState(eventID string, missingFields []string) *ConversationState {
	return &ConversationState{
		EventID:          eventID,
		CollectedData:    make(map[string]string),
		MissingFields:    missingFields,
		ConversationStep: 0,
		IsComplete:       false,
	}
}

// AddCollectedData adds a field to the collected data and removes it from missing fields
func (cs *ConversationState) AddCollectedData(field, value string) {
	cs.CollectedData[field] = value
	cs.RemoveMissingField(field)
	
	// Check if conversation is complete
	if len(cs.MissingFields) == 0 {
		cs.IsComplete = true
	}
}

// RemoveMissingField removes a field from the missing fields list
func (cs *ConversationState) RemoveMissingField(field string) {
	for i, f := range cs.MissingFields {
		if f == field {
			cs.MissingFields = append(cs.MissingFields[:i], cs.MissingFields[i+1:]...)
			break
		}
	}
}

// GetProgress returns the completion percentage
func (cs *ConversationState) GetProgress() float64 {
	totalFields := len(cs.CollectedData) + len(cs.MissingFields)
	if totalFields == 0 {
		return 100.0
	}
	return float64(len(cs.CollectedData)) / float64(totalFields) * 100.0
}
