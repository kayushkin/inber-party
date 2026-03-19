package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kayushkin/inber-party/internal/api"
	"github.com/kayushkin/inber-party/internal/dailyquests"
	"github.com/kayushkin/inber-party/internal/db"
	"github.com/kayushkin/inber-party/internal/events"
	"github.com/kayushkin/inber-party/internal/inber"
	"github.com/kayushkin/inber-party/internal/logger"
	"github.com/kayushkin/inber-party/internal/mood"
	"github.com/kayushkin/inber-party/internal/questgiver"
	"github.com/kayushkin/inber-party/internal/version"
	"github.com/kayushkin/inber-party/internal/sync"
	"github.com/kayushkin/inber-party/internal/ws"
)

func main() {
	// Initialize structured logging
	logger.SetLevelFromEnv()
	
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
			logger.Warn("PostgreSQL unavailable, continuing with inber-only mode", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			defer database.Close()
			if err := database.Migrate(); err != nil {
				logger.Fatal("Failed to run migrations", map[string]interface{}{
					"error": err.Error(),
				})
			}
			if err := database.Seed(); err != nil {
				logger.Fatal("Failed to seed database", map[string]interface{}{
					"error": err.Error(),
				})
			}
			
			// Initialize mood calculator and update all agent moods
			moodCalc := mood.NewMoodCalculator(database)
			if err := moodCalc.UpdateAllAgentMoods(); err != nil {
				logger.Warn("Failed to initialize agent moods", map[string]interface{}{
					"error": err.Error(),
				})
			} else {
				logger.Info("Agent moods initialized")
			}
		}
	} else {
		logger.Warn("No DATABASE_URL set, running in inber-only mode")
	}

	// Initialize Daily Quest Manager
	var dailyQuestMgr *dailyquests.DailyQuestManager
	if database != nil {
		dailyQuestMgr = dailyquests.NewDailyQuestManager(database)
		// Start the daily quest scheduler
		dailyQuestMgr.ScheduleDailyQuestGeneration()
		logger.Info("Daily Quest Manager initialized and scheduled")
	}

	apiServer := api.NewServer(database, hub, nil, dailyQuestMgr)

	mux := http.NewServeMux()
	apiServer.RegisterRoutes(mux)
	mux.HandleFunc("/ws", hub.ServeWS)
	mux.HandleFunc("/api/ws/chat", hub.ServeWS) // Chat-specific WebSocket endpoint

	startTime := time.Now() // Track server startup time for health check

	inberURL := getEnv("INBER_URL", "http://127.0.0.1:8200")

	// Mount inber integration — try SQLite first, fall back to HTTP API
	var inberSource inber.DataSource
	var inberStore *inber.Store // Keep reference for cleanup

	sessDBPath, gwDBPath := inber.DefaultDBPaths()
	if envPath := os.Getenv("INBER_SESSIONS_DB"); envPath != "" {
		sessDBPath = envPath
	}
	if envPath := os.Getenv("INBER_GATEWAY_DB"); envPath != "" {
		gwDBPath = envPath
	}

	store, err := inber.NewStore(sessDBPath, gwDBPath, inberURL)
	if err != nil {
		logger.Warn("Inber SQLite unavailable", map[string]interface{}{
			"error": err.Error(),
		})
	} else if store.HasData() {
		inberStore = store
		inberSource = store
		logger.Info("Inber integration active (SQLite, read-only)")
	} else {
		store.Close()
	}

	// Fall back to HTTP API if no SQLite data
	if inberSource == nil {
		httpClient := inber.NewHTTPClient(inberURL)
		// Quick check if the API is reachable
		if _, err := httpClient.GetAgents(); err != nil {
			logger.Warn("Inber HTTP API unavailable", map[string]interface{}{
				"url":   inberURL,
				"error": err.Error(),
			})
			logger.Warn("Inber integration disabled - no data source available")
		} else {
			inberSource = httpClient
			logger.Info("Inber integration active (HTTP API)", map[string]interface{}{
				"url": inberURL,
			})
		}
	}

	if inberSource != nil {
		inberHandler := inber.NewHandler(inberSource)
		inberHandler.RegisterRoutes(mux)

		// Start real-time quest monitoring for live updates
		questMonitor := events.NewQuestMonitor(inberSource, hub)
		questMonitor.Start()
		
		// Fallback: poll inber data every 30s for full refresh (less frequent since we have real-time events)
		go func() {
			ticker := time.NewTicker(30 * time.Second)
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
		
		// Initialize quest-giver now that we have both database and inber source
		if database != nil {
			qg := questgiver.NewQuestGiver(database, inberSource)
			// Update the API server to use the quest-giver
			apiServer.QuestGiver = qg
			logger.Info("Quest Giver initialized - ready to assign tasks automatically")
			
			// Initialize agent registry sync
			agentSync := sync.NewAgentRegistrySync(inberSource, database)
			apiServer.AgentSync = agentSync
			// Start periodic sync every 5 minutes
			agentSync.StartPeriodicSync(5)
			logger.Info("Agent Registry Sync initialized", map[string]interface{}{
				"sync_interval_minutes": 5,
			})
		}
	}

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		// Check database connectivity
		dbStatus := "disconnected"
		dbError := ""
		if database != nil {
			if err := database.Ping(); err != nil {
				dbStatus = "error"
				dbError = err.Error()
			} else {
				dbStatus = "connected"
			}
		}
		
		// Check inber connectivity
		inberStatus := "disconnected"
		inberError := ""
		if inberSource != nil {
			if _, err := inberSource.GetAgents(); err != nil {
				inberStatus = "error" 
				inberError = err.Error()
			} else {
				inberStatus = "connected"
			}
		}
		
		// Check WebSocket hub
		wsStatus := "running"
		if hub == nil {
			wsStatus = "error"
		}
		
		// Overall health status
		overallStatus := "healthy"
		if dbStatus == "error" || inberStatus == "error" || wsStatus == "error" {
			overallStatus = "degraded"
		}
		if dbStatus == "disconnected" && inberStatus == "disconnected" {
			overallStatus = "unhealthy"
		}
		
		response := map[string]interface{}{
			"status": overallStatus,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"uptime": time.Since(startTime).String(),
			"services": map[string]interface{}{
				"database": map[string]interface{}{
					"status": dbStatus,
					"error":  dbError,
				},
				"inber": map[string]interface{}{
					"status": inberStatus,
					"error":  inberError,
				},
				"websocket": map[string]interface{}{
					"status": wsStatus,
				},
			},
			"version": version.Short(),
		}
		
		// Set appropriate HTTP status code
		statusCode := http.StatusOK
		if overallStatus == "degraded" {
			statusCode = http.StatusPartialContent // 206
		} else if overallStatus == "unhealthy" {
			statusCode = http.StatusServiceUnavailable // 503
		}
		
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(response)
	})

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
		logger.Info("Serving frontend from frontend/dist/")
	} else {
		logger.Warn("Frontend build not found, serving API only", map[string]interface{}{
			"expected_path": "frontend/dist/",
		})
	}

	addr := ":" + port
	logger.Info("Míl Party server starting", map[string]interface{}{
		"addr":            addr,
		"websocket_path":  "/ws",
		"api_path":        "/api",
	})

	// Configure HTTP server with proper timeouts for better error handling
	server := &http.Server{
		Addr:         addr,
		Handler:      corsMiddleware(mux),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Set up graceful shutdown handling
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	serverErrCh := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrCh <- err
		}
	}()

	// Block until we receive a shutdown signal or server error
	select {
	case err := <-serverErrCh:
		logger.Fatal("Server failed to start", map[string]interface{}{
			"error": err.Error(),
		})
	case sig := <-shutdownCh:
		logger.Info("Received signal, initiating graceful shutdown", map[string]interface{}{
			"signal": sig.String(),
		})
		
		// Create shutdown context with timeout
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		
		// Shutdown HTTP server
		logger.Info("Shutting down HTTP server")
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("HTTP server shutdown error", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			logger.Info("HTTP server shutdown complete")
		}
		
		// Close database connections
		if database != nil {
			logger.Info("Closing database connections")
			database.Close()
			logger.Info("Database connections closed")
		}
		
		// Close inber store if it was used
		if inberStore != nil {
			logger.Info("Closing inber store")
			inberStore.Close()
			logger.Info("Inber store closed")
		}
		
		logger.Info("Graceful shutdown complete")
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
