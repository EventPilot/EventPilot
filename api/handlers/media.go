package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"eventpilot/api/models"
	"eventpilot/api/services"

	"github.com/google/uuid"
	storage_go "github.com/supabase-community/storage-go"
	"github.com/supabase-community/supabase-go"
)

const (
	mediaBucket      = "event-media"
	mediaMaxFileSize = 25 << 20 // 25 MB
	mediaSignedTTL   = 60 * 60 * 24 * 7
)

type MediaHandler struct {
	SupabaseClient *supabase.Client
}

type uploadMediaResponse struct {
	ID          string `json:"id"`
	EventID     string `json:"event_id"`
	StoragePath string `json:"storage_path"`
	URL         string `json:"url"`
	CreatedAt   string `json:"created_at"`
}

func (h *MediaHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	user, err := authenticatedUserFromRequest(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	eventID := strings.TrimSpace(r.PathValue("id"))
	if eventID == "" {
		http.Error(w, "event id is required", http.StatusBadRequest)
		return
	}

	if err := r.ParseMultipartForm(mediaMaxFileSize); err != nil {
		http.Error(w, "failed to parse upload", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if header.Size > mediaMaxFileSize {
		http.Error(w, "file too large", http.StatusRequestEntityTooLarge)
		return
	}

	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		http.Error(w, "only image uploads are supported", http.StatusUnsupportedMediaType)
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		ext = ".jpg"
	}
	storagePath := eventID + "/" + uuid.NewString() + ext

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "failed to read upload", http.StatusInternalServerError)
		return
	}

	if _, err := h.SupabaseClient.Storage.UploadFile(mediaBucket, storagePath, bytes.NewReader(data), storage_go.FileOptions{
		ContentType: &contentType,
	}); err != nil {
		http.Error(w, "failed to upload file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Best-effort image analysis with the lightweight Haiku vision model so
	// the chat/planner agent has semantic context about what was uploaded.
	// Failures are logged but never block the upload.
	eventTitle, eventDescription := lookupEventSummary(h.SupabaseClient, eventID)
	analysis, analysisErr := services.AnalyzeImage(r.Context(), data, contentType, eventTitle, eventDescription)
	if analysisErr != nil {
		log.Printf("[UploadMedia] image analysis failed for event=%s: %v", eventID, analysisErr)
	}

	metadata := map[string]any{}
	if analysis != nil && analysis.Description != "" {
		metadata["description"] = analysis.Description
	}

	mediaID := uuid.NewString()
	createdAt := timeNowUTC()
	_, _, err = h.SupabaseClient.From("media").Insert(map[string]any{
		"id":           mediaID,
		"event_id":     eventID,
		"uploaded_by":  user.ID.String(),
		"media_type":   "image",
		"storage_path": storagePath,
		"metadata":     metadata,
		"created_at":   createdAt,
	}, false, "", "", "").Execute()
	if err != nil {
		// Roll back the storage upload so we don't leak orphan files
		_, _ = h.SupabaseClient.Storage.RemoveFile(mediaBucket, []string{storagePath})
		http.Error(w, "failed to save media record: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Push a compact summary of the analysis into the event context so it is
	// surfaced to the planner/chat agent on the next run.
	if analysis != nil {
		if err := appendMediaAnalysisToEventContext(h.SupabaseClient, eventID, mediaID, header.Filename, analysis); err != nil {
			log.Printf("[UploadMedia] failed to update event context with media analysis (event=%s media=%s): %v", eventID, mediaID, err)
		}
	}

	// storage-go@v0.7.0 mutates the shared transport Content-Type header when
	// UploadFile is called with FileOptions.ContentType, breaking subsequent
	// JSON requests on the same client. Use a fresh storage client here so
	// CreateSignedUrl sends the correct Content-Type: application/json.
	freshStorage := storage_go.NewClient(
		os.Getenv("SUPABASE_URL")+"/storage/v1",
		os.Getenv("SUPABASE_API_KEY"),
		nil,
	)
	signed, err := freshStorage.CreateSignedUrl(mediaBucket, storagePath, mediaSignedTTL)
	if err != nil {
		http.Error(w, "failed to create signed url: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := uploadMediaResponse{
		ID:          mediaID,
		EventID:     eventID,
		StoragePath: storagePath,
		URL:         signed.SignedURL,
		CreatedAt:   createdAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

type mediaRecord struct {
	ID          string `json:"id"`
	UploadedBy  string `json:"uploaded_by"`
	StoragePath string `json:"storage_path"`
}

func (h *MediaHandler) DeleteMedia(w http.ResponseWriter, r *http.Request) {
	user, err := authenticatedUserFromRequest(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	mediaID := strings.TrimSpace(r.PathValue("mediaId"))
	if mediaID == "" {
		http.Error(w, "media id is required", http.StatusBadRequest)
		return
	}

	var records []mediaRecord
	_, err = h.SupabaseClient.From("media").
		Select("id, uploaded_by, storage_path", "", false).
		Eq("id", mediaID).
		ExecuteTo(&records)
	if err != nil {
		http.Error(w, "failed to load media record", http.StatusInternalServerError)
		return
	}
	if len(records) == 0 {
		http.Error(w, "media not found", http.StatusNotFound)
		return
	}

	record := records[0]
	if record.UploadedBy != user.ID.String() {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if _, err := h.SupabaseClient.Storage.RemoveFile(mediaBucket, []string{record.StoragePath}); err != nil {
		http.Error(w, "failed to delete file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_, _, err = h.SupabaseClient.From("media").Delete("", "").Eq("id", mediaID).Execute()
	if err != nil {
		http.Error(w, "failed to delete media record: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func ListMedia(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotImplemented) }

// lookupEventSummary fetches the event title and description to pass as
// hints to the vision model. Failures are non-fatal — empty strings are fine.
func lookupEventSummary(client *supabase.Client, eventID string) (string, string) {
	var events []models.Event
	_, err := client.From("event").
		Select("id, title, description", "", false).
		Eq("id", eventID).
		ExecuteTo(&events)
	if err != nil || len(events) == 0 {
		return "", ""
	}
	return events[0].Title, events[0].Description
}

// appendMediaAnalysisToEventContext pushes a compact analysis summary onto
// event.context.media_analyses so the planner agent sees what has been
// uploaded without having to query the media table directly. We keep the
// list bounded so the context payload stays small.
const maxMediaAnalysesInContext = 20

func appendMediaAnalysisToEventContext(client *supabase.Client, eventID, mediaID, filename string, analysis *services.ImageAnalysis) error {
	entry := map[string]any{
		"media_id": mediaID,
		"filename": filename,
	}
	if analysis.Caption != "" {
		entry["caption"] = analysis.Caption
	}
	if analysis.Description != "" {
		entry["description"] = analysis.Description
	}
	if len(analysis.Subjects) > 0 {
		entry["subjects"] = analysis.Subjects
	}
	if analysis.Scene != "" {
		entry["scene"] = analysis.Scene
	}
	if analysis.Mood != "" {
		entry["mood"] = analysis.Mood
	}
	if len(analysis.Tags) > 0 {
		entry["tags"] = analysis.Tags
	}

	// Load the current context, append, and write back via the shared helper.
	var events []models.Event
	_, err := client.From("event").
		Select("id, context", "", false).
		Eq("id", eventID).
		ExecuteTo(&events)
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return nil
	}

	existing := events[0].Context
	var list []any
	if existing != nil {
		if raw, ok := existing["media_analyses"]; ok {
			if arr, ok := raw.([]any); ok {
				list = arr
			}
		}
	}
	list = append(list, entry)
	if len(list) > maxMediaAnalysesInContext {
		list = list[len(list)-maxMediaAnalysesInContext:]
	}

	return services.UpdateEventContext(client, eventID, map[string]any{
		"media_analyses": list,
	})
}
