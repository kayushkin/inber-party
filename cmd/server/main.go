package main

import (
	"log"
	"net/http"
	"os"

	"github.com/kayushkin/inber-party/internal/api"
	"github.com/kayushkin/inber-party/internal/db"
	"github.com/kayushkin/inber-party/internal/ws"
)

func main() {
	// Get config from environment
	databaseURL := getEnv("DATABASE_URL", "postgres://localhost:5432/milparty?sslmode=disable")
	port := getEnv("PORT", "8080")

	// Connect to database
	database, err := db.Connect(databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Run migrations
	if err := database.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Seed database
	if err := database.Seed(); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	// Create WebSocket hub
	hub := ws.NewHub()
	go hub.Run()

	// Create API server
	apiServer := api.NewServer(database, hub)

	// Set up routes
	mux := http.NewServeMux()
	apiServer.RegisterRoutes(mux)
	mux.HandleFunc("/ws", hub.ServeWS)

	// Serve frontend static files (production)
	if _, err := os.Stat("frontend/dist"); err == nil {
		fs := http.FileServer(http.Dir("frontend/dist"))
		mux.Handle("/", fs)
		log.Println("✓ Serving frontend from frontend/dist/")
	} else {
		log.Println("⚠ Frontend build not found (frontend/dist/), serving API only")
	}

	// Start server
	addr := ":" + port
	log.Printf("🚀 Míl Party server listening on %s", addr)
	log.Printf("📡 WebSocket endpoint: ws://localhost%s/ws", addr)
	log.Printf("🔌 API endpoint: http://localhost%s/api", addr)

	if err := http.ListenAndServe(addr, corsMiddleware(mux)); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
