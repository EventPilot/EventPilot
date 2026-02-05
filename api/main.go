package main

import (
	"log"
	"net/http"
	"os"

	"eventpilot/api/handlers"
	"eventpilot/api/middleware"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/events", handlers.ListEvents)
	mux.HandleFunc("POST /api/events", handlers.CreateEvent)
	mux.HandleFunc("GET /api/events/{id}", handlers.GetEvent)
	mux.HandleFunc("PATCH /api/events/{id}", handlers.UpdateEvent)
	mux.HandleFunc("GET /api/events/{id}/chat", handlers.GetChat)
	mux.HandleFunc("POST /api/events/{id}/chat/messages", handlers.CreateChatMessage)
	mux.HandleFunc("POST /api/events/{id}/chat/request-inputs", handlers.RequestInputs)
	mux.HandleFunc("POST /api/events/{id}/media", handlers.UploadMedia)
	mux.HandleFunc("GET /api/events/{id}/media", handlers.ListMedia)
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
