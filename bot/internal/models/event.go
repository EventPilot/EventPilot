package models

import (
	"strings"
	"time"
)

type Event struct {
	ID              string    `json:"id" db:"id"`
	Name            string    `json:"name" db:"name"`
	Date            string    `json:"date" db:"date"`
	Location        string    `json:"location" db:"location"`
	Description     string    `json:"description" db:"description"`
	Highlights      []string  `json:"highlights" db:"highlights"`
	TargetAudience  []string  `json:"target_audience" db:"target_audience"`
	SpecialGuests   []string  `json:"special_guests" db:"special_guests"`
	Photos          []string  `json:"photos" db:"photos"`
	AdditionalNotes string    `json:"additional_notes" db:"additional_notes"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// Checks if a specific field is empty
func (e *Event) HasMissingField(field string) bool {
	switch field {
	case "name":
		return e.Name == ""
	case "date":
		return e.Date == ""
	case "location":
		return e.Location == ""
	case "description":
		return e.Description == ""
	case "highlights":
		return len(e.Highlights) == 0 || (len(e.Highlights) == 1 && e.Highlights[0] == "")
	case "target_audience":
		return len(e.TargetAudience) == 0 || (len(e.TargetAudience) == 1 && e.TargetAudience[0] == "")
	case "special_guests":
		return len(e.SpecialGuests) == 0 || (len(e.SpecialGuests) == 1 && e.SpecialGuests[0] == "")
	case "photos":
		return len(e.Photos) == 0
	default:
		return false
	}
}

// Returns a list of fields that are empty
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
	postEventFields := []string{"highlights", "target_audience", "special_guests"}
	for _, field := range postEventFields {
		if e.HasMissingField(field) {
			missing = append(missing, field)
		}
	}

	return missing
}

// Checks if all required fields are filled
func (e *Event) IsComplete() bool {
	// Basic required fields
	if e.Name == "" || e.Date == "" || e.Location == "" {
		return false
	}

	// At least some content
	if e.Description == "" && len(e.Highlights) == 0 {
		return false
	}

	return true
}

// GetHighlightsString returns highlights as a comma-separated string
func (e *Event) GetHighlightsString() string {
	return strings.Join(e.Highlights, ", ")
}

// GetTargetAudienceString returns target audience as a comma-separated string
func (e *Event) GetTargetAudienceString() string {
	return strings.Join(e.TargetAudience, ", ")
}

// GetSpecialGuestsString returns special guests as a comma-separated string
func (e *Event) GetSpecialGuestsString() string {
	return strings.Join(e.SpecialGuests, ", ")
}
