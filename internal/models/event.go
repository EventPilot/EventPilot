package models

import "time"

// Event represents an event with all its details
type Event struct {
	ID              string    `json:"id" db:"id"`
	Name            string    `json:"name" db:"name"`
	Date            time.Time `json:"date" db:"date"`
	Location        string    `json:"location" db:"location"`
	Description     string    `json:"description" db:"description"`
	Highlights      string    `json:"highlights" db:"highlights"`
	TargetAudience  string    `json:"target_audience" db:"target_audience"`
	SpecialGuests   []string  `json:"special_guests" db:"special_guests"`
	Photos          []string  `json:"photos" db:"photos"`
	AdditionalNotes string    `json:"additional_notes" db:"additional_notes"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// HasMissingField checks if a specific field is empty
func (e *Event) HasMissingField(field string) bool {
	switch field {
	case "name":
		return e.Name == ""
	case "date":
		return e.Date.IsZero()
	case "location":
		return e.Location == ""
	case "description":
		return e.Description == ""
	case "highlights":
		return e.Highlights == ""
	case "target_audience":
		return e.TargetAudience == ""
	case "special_guests":
		return len(e.SpecialGuests) == 0
	case "photos":
		return len(e.Photos) == 0
	default:
		return false
	}
}

// GetMissingFields returns a list of fields that are empty
func (e *Event) GetMissingFields() []string {
	var missing []string
	
	// Check basic calendar fields
	basicFields := []string{"name", "date", "location", "description"}
	for _, field := range basicFields {
		if e.HasMissingField(field) {
			missing = append(missing, field)
		}
	}
	
	// Always collect post-event details
	postEventFields := []string{"highlights", "target_audience", "special_guests", "photos"}
	for _, field := range postEventFields {
		if e.HasMissingField(field) {
			missing = append(missing, field)
		}
	}
	
	return missing
}
