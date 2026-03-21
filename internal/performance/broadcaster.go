package performance

import (
	"context"
	"log"
	"sync"
	"time"
)

// PerformanceBroadcaster handles real-time performance metrics broadcasting via WebSocket
type PerformanceBroadcaster struct {
	collector    *MetricsCollector
	broadcastCh  chan<- interface{} // Channel to send WebSocket messages
	subscribers  map[chan PerformanceSnapshot]bool
	mu           sync.RWMutex
	interval     time.Duration
	running      bool
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewPerformanceBroadcaster creates a new performance broadcaster
func NewPerformanceBroadcaster(collector *MetricsCollector, broadcastCh chan<- interface{}) *PerformanceBroadcaster {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &PerformanceBroadcaster{
		collector:   collector,
		broadcastCh: broadcastCh,
		subscribers: make(map[chan PerformanceSnapshot]bool),
		interval:    5 * time.Second, // Broadcast every 5 seconds
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start begins broadcasting performance metrics
func (pb *PerformanceBroadcaster) Start() {
	pb.mu.Lock()
	if pb.running {
		pb.mu.Unlock()
		return
	}
	pb.running = true
	pb.mu.Unlock()
	
	log.Println("🚀 Started real-time performance metrics broadcaster")
	
	go pb.broadcastLoop()
}

// Stop stops the performance broadcaster
func (pb *PerformanceBroadcaster) Stop() {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	
	if !pb.running {
		return
	}
	
	pb.running = false
	pb.cancel()
	
	// Close all subscriber channels
	for ch := range pb.subscribers {
		close(ch)
	}
	pb.subscribers = make(map[chan PerformanceSnapshot]bool)
	
	log.Println("🛑 Stopped performance metrics broadcaster")
}

// Subscribe adds a subscriber for performance updates
func (pb *PerformanceBroadcaster) Subscribe() (<-chan PerformanceSnapshot, func()) {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	
	ch := make(chan PerformanceSnapshot, 10) // Buffer to prevent blocking
	pb.subscribers[ch] = true
	
	// Return the channel and an unsubscribe function
	unsubscribe := func() {
		pb.mu.Lock()
		defer pb.mu.Unlock()
		
		if pb.subscribers[ch] {
			delete(pb.subscribers, ch)
			close(ch)
		}
	}
	
	return ch, unsubscribe
}

// SetInterval changes the broadcast interval
func (pb *PerformanceBroadcaster) SetInterval(interval time.Duration) {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	
	if interval < time.Second {
		interval = time.Second
	}
	if interval > time.Minute {
		interval = time.Minute
	}
	
	pb.interval = interval
	log.Printf("📊 Performance broadcast interval updated to %v", interval)
}

// GetActiveSubscribers returns the number of active subscribers
func (pb *PerformanceBroadcaster) GetActiveSubscribers() int {
	pb.mu.RLock()
	defer pb.mu.RUnlock()
	return len(pb.subscribers)
}

// broadcastLoop is the main loop that broadcasts performance data
func (pb *PerformanceBroadcaster) broadcastLoop() {
	ticker := time.NewTicker(pb.interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-pb.ctx.Done():
			return
			
		case <-ticker.C:
			pb.broadcastPerformanceUpdate()
			
			// Update ticker interval if it changed
			pb.mu.RLock()
			currentInterval := pb.interval
			pb.mu.RUnlock()
			
			if ticker.C != time.NewTicker(currentInterval).C {
				ticker.Stop()
				ticker = time.NewTicker(currentInterval)
			}
		}
	}
}

// broadcastPerformanceUpdate sends current performance metrics to all subscribers
func (pb *PerformanceBroadcaster) broadcastPerformanceUpdate() {
	// Get current performance snapshot
	snapshot := pb.collector.GetSnapshot()
	
	// Create WebSocket message
	wsMessage := map[string]interface{}{
		"type": "performance_update",
		"data": map[string]interface{}{
			"timestamp":    snapshot.Timestamp,
			"health_status": snapshot.HealthStatus,
			"system": map[string]interface{}{
				"memory_usage":    snapshot.SystemMetrics.MemoryUsage,
				"memory_percent":  snapshot.SystemMetrics.MemoryPercent,
				"goroutine_count": snapshot.SystemMetrics.GoroutineCount,
				"cpu_usage":       snapshot.SystemMetrics.CPUUsage,
				"gc_count":        snapshot.SystemMetrics.GCCount,
			},
			"api_summary": map[string]interface{}{
				"total_endpoints": len(snapshot.APIMetrics),
				"total_requests":  pb.getTotalRequests(snapshot.APIMetrics),
				"error_rate":      pb.calculateAPIErrorRate(snapshot.APIMetrics),
			},
			"agents": map[string]interface{}{
				"active_count": len(snapshot.AgentMetrics),
				"avg_success_rate": pb.calculateAvgAgentSuccessRate(snapshot.AgentMetrics),
				"total_tasks": pb.getTotalAgentTasks(snapshot.AgentMetrics),
			},
			"websocket": map[string]interface{}{
				"active_connections": snapshot.WebSocketMetrics.ActiveConnections,
				"connection_churn":   snapshot.WebSocketMetrics.ConnectionChurn,
				"average_latency":    snapshot.WebSocketMetrics.AverageLatency.String(),
			},
			"database": map[string]interface{}{
				"query_count":        snapshot.DatabaseMetrics.QueryCount,
				"average_query_time": snapshot.DatabaseMetrics.AverageQueryTime.String(),
				"slow_queries":       snapshot.DatabaseMetrics.SlowQueries,
				"cache_hit_rate":     snapshot.DatabaseMetrics.CacheHitRate,
			},
			"recommendations_count": len(snapshot.Recommendations),
			"uptime": snapshot.Uptime.String(),
		},
	}
	
	// Send to WebSocket broadcaster if available
	if pb.broadcastCh != nil {
		select {
		case pb.broadcastCh <- wsMessage:
		default:
			log.Println("⚠️ Performance broadcast channel full, dropping update")
		}
	}
	
	// Send to direct subscribers
	pb.mu.RLock()
	subscribers := make([]chan PerformanceSnapshot, 0, len(pb.subscribers))
	for ch := range pb.subscribers {
		subscribers = append(subscribers, ch)
	}
	pb.mu.RUnlock()
	
	// Send to subscribers (non-blocking)
	for _, ch := range subscribers {
		select {
		case ch <- *snapshot:
		default:
			// Subscriber channel full, skip this update
		}
	}
}

// Helper functions for calculating aggregated metrics

func (pb *PerformanceBroadcaster) getTotalRequests(apiMetrics map[string]*APIMetric) int64 {
	var total int64
	for _, metric := range apiMetrics {
		total += metric.TotalRequests
	}
	return total
}

func (pb *PerformanceBroadcaster) calculateAPIErrorRate(apiMetrics map[string]*APIMetric) float64 {
	var totalRequests, errorRequests int64
	
	for _, metric := range apiMetrics {
		totalRequests += metric.TotalRequests
		errorRequests += metric.ErrorRequests
	}
	
	if totalRequests == 0 {
		return 0.0
	}
	
	return float64(errorRequests) / float64(totalRequests) * 100
}

func (pb *PerformanceBroadcaster) calculateAvgAgentSuccessRate(agentMetrics map[string]*AgentMetric) float64 {
	if len(agentMetrics) == 0 {
		return 0.0
	}
	
	var total float64
	for _, metric := range agentMetrics {
		total += metric.SuccessRate
	}
	
	return total / float64(len(agentMetrics))
}

func (pb *PerformanceBroadcaster) getTotalAgentTasks(agentMetrics map[string]*AgentMetric) int64 {
	var total int64
	for _, metric := range agentMetrics {
		total += metric.TotalTasks
	}
	return total
}

// BroadcastAlert sends a performance alert immediately
func (pb *PerformanceBroadcaster) BroadcastAlert(alertType, title, description, priority string) {
	alert := map[string]interface{}{
		"type": "performance_alert",
		"data": map[string]interface{}{
			"alert_type":  alertType,
			"title":       title,
			"description": description,
			"priority":    priority,
			"timestamp":   time.Now(),
		},
	}
	
	if pb.broadcastCh != nil {
		select {
		case pb.broadcastCh <- alert:
			log.Printf("🚨 Performance alert broadcasted: %s (%s)", title, priority)
		default:
			log.Println("⚠️ Performance alert broadcast channel full, dropping alert")
		}
	}
}