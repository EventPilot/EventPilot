package main

import (
	"log"
	"net/http"
	"os"
	"eventpilot/api/handlers"
	"eventpilot/api/middleware"
	"github.com/supabase-community/supabase-go"
)

func main() {
	mux := http.NewServeMux()

	supabaseClient, err := supabase.NewClient(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_SERVICE_ROLE_KEY"), nil)
	if err != nil {
		log.Fatalf("Error creating supabase client: %v", err)
	}

	chatHandler := &handlers.ChatHandler{SupabaseClient: supabaseClient}

	mux.HandleFunc("POST /api/events", handlers.CreateEvent)
	mux.HandleFunc("PATCH /api/events/{id}", handlers.UpdateEvent)
	mux.HandleFunc("POST /api/events/{id}/chat/messages", handlers.CreateChatMessage)
	mux.HandleFunc("POST /api/events/{id}/chat/request-inputs", chatHandler.RequestInputs)
	mux.HandleFunc("POST /api/events/{id}/media", handlers.UploadMedia)
	mux.HandleFunc("POST /api/events/{id}/generate-post", handlers.GeneratePost)
	mux.HandleFunc("GET /api/events/{id}/post", handlers.GetPost)
	mux.HandleFunc("POST /api/events/{id}/post/publish", handlers.PublishPost)
	mux.HandleFunc("GET /api/users", handlers.ListUsers)
	mux.HandleFunc("GET /api/users/{id}", handlers.GetUser)

	handler := middleware.Auth(mux)
	handler = middleware.CORS(handler)
	handler = middleware.Logger(handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
