package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kayushkin/inber-party/internal/api"
	"github.com/kayushkin/inber-party/internal/db"
	"github.com/kayushkin/inber-party/internal/inber"
	"github.com/kayushkin/inber-party/internal/ws"
)

func main() {
	databaseURL := getEnv("DATABASE_URL", "")
	port := getEnv("PORT", "8080")

	hub := ws.NewHub()
	go hub.Run()

	// Connect to PostgreSQL (optional)
	var database *db.DB
	if databaseURL != "" {
		var err error
		database, err = db.Connect(databaseURL)
		if err != nil {
			log.Printf("⚠ PostgreSQL unavailable: %v (continuing with inber-only mode)", err)
		} else {
			defer database.Close()
			if err := database.Migrate(); err != nil {
				log.Fatalf("Failed to run migrations: %v", err)
			}
			if err := database.Seed(); err != nil {
				log.Fatalf("Failed to seed database: %v", err)
			}
		}
	} else {
		log.Println("⚠ No DATABASE_URL set — running in inber-only mode")
	}

	apiServer := api.NewServer(database, hub)

	mux := http.NewServeMux()
	apiServer.RegisterRoutes(mux)
	mux.HandleFunc("/ws", hub.ServeWS)

	// Health check endpoint
	startTime := time.Now()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"uptime": time.Since(startTime).String(),
		})
	})

	inberURL := getEnv("INBER_URL", "http://127.0.0.1:8200")

	// Mount inber integration — try SQLite first, fall back to HTTP API
	var inberSource inber.DataSource

	sessDBPath, gwDBPath := inber.DefaultDBPaths()
	if envPath := os.Getenv("INBER_SESSIONS_DB"); envPath != "" {
		sessDBPath = envPath
	}
	if envPath := os.Getenv("INBER_GATEWAY_DB"); envPath != "" {
		gwDBPath = envPath
	}

	inberStore, err := inber.NewStore(sessDBPath, gwDBPath, inberURL)
	if err != nil {
		log.Printf("⚠ Inber SQLite unavailable: %v", err)
	} else if inberStore.HasData() {
		defer inberStore.Close()
		inberSource = inberStore
		log.Println("✓ Inber integration active (SQLite, read-only)")
	} else {
		inberStore.Close()
	}

	// Fall back to HTTP API if no SQLite data
	if inberSource == nil {
		httpClient := inber.NewHTTPClient(inberURL)
		// Quick check if the API is reachable
		if _, err := httpClient.GetAgents(); err != nil {
			log.Printf("⚠ Inber HTTP API at %s also unavailable: %v", inberURL, err)
			log.Println("⚠ Inber integration disabled — no data source available")
		} else {
			inberSource = httpClient
			log.Printf("✓ Inber integration active (HTTP API at %s)", inberURL)
		}
	}

	if inberSource != nil {
		inberHandler := inber.NewHandler(inberSource)
		inberHandler.RegisterRoutes(mux)

		// Auto-refresh: poll inber data every 10s and broadcast changes via WebSocket
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				agents, err := inberSource.GetAgents()
				if err != nil {
					continue
				}
				stats, err := inberSource.GetStats()
				if err != nil {
					continue
				}
				hub.Broadcast(ws.Message{
					Type: "inber_update",
					Data: map[string]interface{}{
						"agents": agents,
						"stats":  stats,
					},
				})
			}
		}()
	}

	// Proxy /api/run to inber for chat/task functionality
	mux.HandleFunc("/api/run", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// Forward the request to inber
		proxyReq, err := http.NewRequest(http.MethodPost, inberURL+"/api/run", r.Body)
		if err != nil {
			http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
			return
		}
		proxyReq.Header.Set("Content-Type", "application/json")
		proxyReq.Header.Set("Accept", "text/event-stream")

		client := &http.Client{Timeout: 0} // no timeout for streaming
		resp, err := client.Do(proxyReq)
		if err != nil {
			http.Error(w, "Failed to reach inber: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Stream the response back
		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(resp.StatusCode)

		flusher, ok := w.(http.Flusher)
		buf := make([]byte, 4096)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				w.Write(buf[:n])
				if ok {
					flusher.Flush()
				}
			}
			if err != nil {
				break
			}
		}
	})

	// Proxy /api/agents and /api/sessions to inber for direct access
	for _, path := range []string{"/api/agents", "/api/sessions"} {
		p := path
		mux.HandleFunc("/api/inber-proxy"+p, func(w http.ResponseWriter, r *http.Request) {
			resp, err := http.Get(inberURL + p)
			if err != nil {
				http.Error(w, "Failed to reach inber", http.StatusBadGateway)
				return
			}
			defer resp.Body.Close()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(resp.StatusCode)
			io.Copy(w, resp.Body)
		})
	}

	// Serve frontend static files
	if _, err := os.Stat("frontend/dist"); err == nil {
		fs := http.FileServer(http.Dir("frontend/dist"))
		mux.Handle("/", fs)
		log.Println("✓ Serving frontend from frontend/dist/")
	} else {
		log.Println("⚠ Frontend build not found (frontend/dist/), serving API only")
	}

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
