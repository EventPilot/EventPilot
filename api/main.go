package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"eventpilot/api/handlers"
	"eventpilot/api/middleware"

	"github.com/joho/godotenv"
	"github.com/supabase-community/supabase-go"
)

func startCronWorker(h *handlers.CronHandler, interval time.Duration) {
	run := func() {
		if err := h.ProcessCompletedEvents(context.Background()); err != nil {
			log.Printf("cron worker error: %v", err)
		}
	}
	run() // run immediately on startup
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			run()
		}
	}()
}

func main() {
	_ = godotenv.Load("./.env")
	_ = godotenv.Load("./api/.env")

	mux := http.NewServeMux()

	supabaseClient, err := supabase.NewClient(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_API_KEY"), nil)
	if err != nil {
		log.Fatalf("Error creating supabase client: %v", err)
	}

	chatHandler := &handlers.ChatHandler{SupabaseClient: supabaseClient}
	cronHandler := &handlers.CronHandler{SupabaseClient: supabaseClient}

	startCronWorker(cronHandler, time.Hour)

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
