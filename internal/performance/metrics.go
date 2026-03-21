package performance

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// MetricsCollector collects and tracks performance metrics
type MetricsCollector struct {
	mutex                 sync.RWMutex
	startTime            time.Time
	apiMetrics           map[string]*APIMetric
	systemMetrics        *SystemMetric
	agentMetrics         map[string]*AgentMetric
	questMetrics         map[string]*QuestMetric
	websocketMetrics     *WebSocketMetric
	databaseMetrics      *DatabaseMetric
	userExperienceMetrics *UserExperienceMetric
}

// APIMetric tracks API endpoint performance
type APIMetric struct {
	Endpoint        string        `json:"endpoint"`
	TotalRequests   int64         `json:"total_requests"`
	SuccessRequests int64         `json:"success_requests"`
	ErrorRequests   int64         `json:"error_requests"`
	AverageLatency  time.Duration `json:"average_latency"`
	MinLatency      time.Duration `json:"min_latency"`
	MaxLatency      time.Duration `json:"max_latency"`
	LastRequest     time.Time     `json:"last_request"`
	RequestTimes    []time.Duration `json:"-"` // Keep last 100 for rolling average
}

// SystemMetric tracks system-level performance
type SystemMetric struct {
	CPUUsage       float64   `json:"cpu_usage"`
	MemoryUsage    int64     `json:"memory_usage"` // bytes
	MemoryPercent  float64   `json:"memory_percent"`
	GoroutineCount int       `json:"goroutine_count"`
	GCCount        uint32    `json:"gc_count"`
	LastUpdated    time.Time `json:"last_updated"`
}

// AgentMetric tracks individual agent performance
type AgentMetric struct {
	AgentID         string        `json:"agent_id"`
	TokensPerMinute float64       `json:"tokens_per_minute"`
	SuccessRate     float64       `json:"success_rate"`
	ErrorCount      int64         `json:"error_count"`
	TotalTasks      int64         `json:"total_tasks"`
	AvgTaskTime     time.Duration `json:"avg_task_time"`
	IdleTime        time.Duration `json:"idle_time"`
	ProductivityScore float64     `json:"productivity_score"`
	LastActivity    time.Time     `json:"last_activity"`
}

// QuestMetric tracks quest performance
type QuestMetric struct {
	QuestID         string        `json:"quest_id"`
	CompletionTime  time.Duration `json:"completion_time"`
	TokensUsed      int           `json:"tokens_used"`
	Complexity      string        `json:"complexity"`
	ResourceUsage   int64         `json:"resource_usage"` // memory usage during quest
	ErrorCount      int           `json:"error_count"`
	Status          string        `json:"status"`
	StartTime       time.Time     `json:"start_time"`
	EndTime         time.Time     `json:"end_time"`
}

// WebSocketMetric tracks WebSocket performance
type WebSocketMetric struct {
	TotalConnections    int64     `json:"total_connections"`
	ActiveConnections   int64     `json:"active_connections"`
	MessagesPerSecond   float64   `json:"messages_per_second"`
	AverageLatency      time.Duration `json:"average_latency"`
	ConnectionChurn     int64     `json:"connection_churn"`
	ErrorRate           float64   `json:"error_rate"`
	LastUpdated         time.Time `json:"last_updated"`
}

// DatabaseMetric tracks database performance
type DatabaseMetric struct {
	QueryCount        int64         `json:"query_count"`
	AverageQueryTime  time.Duration `json:"average_query_time"`
	SlowQueries       int64         `json:"slow_queries"` // queries > 100ms
	ConnectionCount   int           `json:"connection_count"`
	CacheHitRate      float64       `json:"cache_hit_rate"`
	ErrorRate         float64       `json:"error_rate"`
	LastUpdated       time.Time     `json:"last_updated"`
}

// UserExperienceMetric tracks client-side performance
type UserExperienceMetric struct {
	PageLoadTime      time.Duration `json:"page_load_time"`
	BundleSize        int64         `json:"bundle_size"`
	InteractionLatency time.Duration `json:"interaction_latency"`
	ErrorCount        int64         `json:"error_count"`
	LastUpdated       time.Time     `json:"last_updated"`
}

// PerformanceSnapshot represents a point-in-time view of all metrics
type PerformanceSnapshot struct {
	Timestamp         time.Time                        `json:"timestamp"`
	Uptime            time.Duration                    `json:"uptime"`
	SystemMetrics     *SystemMetric                    `json:"system_metrics"`
	APIMetrics        map[string]*APIMetric            `json:"api_metrics"`
	AgentMetrics      map[string]*AgentMetric          `json:"agent_metrics"`
	QuestMetrics      map[string]*QuestMetric          `json:"quest_metrics"`
	WebSocketMetrics  *WebSocketMetric                 `json:"websocket_metrics"`
	DatabaseMetrics   *DatabaseMetric                  `json:"database_metrics"`
	UserExperienceMetrics *UserExperienceMetric        `json:"user_experience_metrics"`
	HealthStatus      string                           `json:"health_status"`
	Recommendations   []PerformanceRecommendation      `json:"recommendations"`
}

// PerformanceRecommendation provides actionable insights
type PerformanceRecommendation struct {
	Type        string    `json:"type"`
	Priority    string    `json:"priority"` // low, medium, high, critical
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Impact      string    `json:"impact"`
	Action      string    `json:"action"`
	CreatedAt   time.Time `json:"created_at"`
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		startTime:             time.Now(),
		apiMetrics:           make(map[string]*APIMetric),
		agentMetrics:         make(map[string]*AgentMetric),
		questMetrics:         make(map[string]*QuestMetric),
		systemMetrics:        &SystemMetric{},
		websocketMetrics:     &WebSocketMetric{},
		databaseMetrics:      &DatabaseMetric{},
		userExperienceMetrics: &UserExperienceMetric{},
	}
}

// RecordAPICall records an API call performance metric
func (mc *MetricsCollector) RecordAPICall(endpoint string, duration time.Duration, success bool) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	if mc.apiMetrics[endpoint] == nil {
		mc.apiMetrics[endpoint] = &APIMetric{
			Endpoint:    endpoint,
			MinLatency: duration,
			MaxLatency: duration,
		}
	}
	
	metric := mc.apiMetrics[endpoint]
	metric.TotalRequests++
	metric.LastRequest = time.Now()
	
	if success {
		metric.SuccessRequests++
	} else {
		metric.ErrorRequests++
	}
	
	// Update latency stats
	if duration < metric.MinLatency {
		metric.MinLatency = duration
	}
	if duration > metric.MaxLatency {
		metric.MaxLatency = duration
	}
	
	// Keep rolling window of last 100 requests for average calculation
	metric.RequestTimes = append(metric.RequestTimes, duration)
	if len(metric.RequestTimes) > 100 {
		metric.RequestTimes = metric.RequestTimes[1:]
	}
	
	// Calculate rolling average
	var total time.Duration
	for _, t := range metric.RequestTimes {
		total += t
	}
	metric.AverageLatency = total / time.Duration(len(metric.RequestTimes))
}

// UpdateSystemMetrics updates system performance metrics
func (mc *MetricsCollector) UpdateSystemMetrics() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	mc.systemMetrics.MemoryUsage = int64(m.Alloc)
	mc.systemMetrics.MemoryPercent = float64(m.Alloc) / float64(m.Sys) * 100
	mc.systemMetrics.GoroutineCount = runtime.NumGoroutine()
	mc.systemMetrics.GCCount = m.NumGC
	mc.systemMetrics.LastUpdated = time.Now()
	
	// Note: CPU usage would require additional platform-specific code
	// For now, we'll estimate based on GC frequency
	mc.systemMetrics.CPUUsage = float64(runtime.NumGoroutine()) / 100.0
}

// RecordAgentActivity records agent performance data
func (mc *MetricsCollector) RecordAgentActivity(agentID string, tokensUsed int, taskDuration time.Duration, success bool) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	if mc.agentMetrics[agentID] == nil {
		mc.agentMetrics[agentID] = &AgentMetric{
			AgentID: agentID,
		}
	}
	
	metric := mc.agentMetrics[agentID]
	metric.TotalTasks++
	metric.LastActivity = time.Now()
	
	if !success {
		metric.ErrorCount++
	}
	
	// Calculate success rate
	metric.SuccessRate = float64(metric.TotalTasks-metric.ErrorCount) / float64(metric.TotalTasks) * 100
	
	// Update tokens per minute
	if taskDuration > 0 {
		tokensPerMinute := float64(tokensUsed) / taskDuration.Minutes()
		metric.TokensPerMinute = (metric.TokensPerMinute + tokensPerMinute) / 2 // Simple moving average
	}
	
	// Update average task time
	metric.AvgTaskTime = (metric.AvgTaskTime + taskDuration) / 2
	
	// Calculate productivity score (tokens per minute weighted by success rate)
	metric.ProductivityScore = metric.TokensPerMinute * (metric.SuccessRate / 100)
}

// RecordQuestMetrics records quest performance data
func (mc *MetricsCollector) RecordQuestMetrics(questID string, startTime, endTime time.Time, tokensUsed int, complexity string, errors int) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	completionTime := endTime.Sub(startTime)
	
	mc.questMetrics[questID] = &QuestMetric{
		QuestID:        questID,
		CompletionTime: completionTime,
		TokensUsed:     tokensUsed,
		Complexity:     complexity,
		ErrorCount:     errors,
		StartTime:      startTime,
		EndTime:        endTime,
		Status:         "completed",
	}
}

// UpdateWebSocketMetrics updates WebSocket performance metrics
func (mc *MetricsCollector) UpdateWebSocketMetrics(totalConnections, activeConnections, connectionChurn int64, avgLatency time.Duration, errorRate float64) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	mc.websocketMetrics.TotalConnections = totalConnections
	mc.websocketMetrics.ActiveConnections = activeConnections
	mc.websocketMetrics.ConnectionChurn = connectionChurn
	mc.websocketMetrics.AverageLatency = avgLatency
	mc.websocketMetrics.ErrorRate = errorRate
	mc.websocketMetrics.LastUpdated = time.Now()
}

// RecordDatabaseQuery records database query performance
func (mc *MetricsCollector) RecordDatabaseQuery(duration time.Duration, success bool) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	mc.databaseMetrics.QueryCount++
	
	// Update average query time
	if mc.databaseMetrics.AverageQueryTime == 0 {
		mc.databaseMetrics.AverageQueryTime = duration
	} else {
		mc.databaseMetrics.AverageQueryTime = (mc.databaseMetrics.AverageQueryTime + duration) / 2
	}
	
	// Track slow queries
	if duration > 100*time.Millisecond {
		mc.databaseMetrics.SlowQueries++
	}
	
	if !success {
		totalQueries := float64(mc.databaseMetrics.QueryCount)
		errorQueries := float64(mc.databaseMetrics.QueryCount) * (mc.databaseMetrics.ErrorRate / 100) + 1
		mc.databaseMetrics.ErrorRate = (errorQueries / totalQueries) * 100
	}
	
	mc.databaseMetrics.LastUpdated = time.Now()
}

// GetSnapshot returns a complete performance snapshot
func (mc *MetricsCollector) GetSnapshot() *PerformanceSnapshot {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	
	// Update system metrics before snapshot
	mc.UpdateSystemMetrics()
	
	// Generate recommendations
	recommendations := mc.generateRecommendations()
	
	// Determine overall health status
	healthStatus := mc.calculateHealthStatus()
	
	return &PerformanceSnapshot{
		Timestamp:             time.Now(),
		Uptime:               time.Since(mc.startTime),
		SystemMetrics:         mc.systemMetrics,
		APIMetrics:           mc.copyAPIMetrics(),
		AgentMetrics:         mc.copyAgentMetrics(),
		QuestMetrics:         mc.copyQuestMetrics(),
		WebSocketMetrics:     mc.websocketMetrics,
		DatabaseMetrics:      mc.databaseMetrics,
		UserExperienceMetrics: mc.userExperienceMetrics,
		HealthStatus:         healthStatus,
		Recommendations:      recommendations,
	}
}

// Helper methods for copying metrics (to avoid race conditions)
func (mc *MetricsCollector) copyAPIMetrics() map[string]*APIMetric {
	copy := make(map[string]*APIMetric)
	for k, v := range mc.apiMetrics {
		copy[k] = &APIMetric{
			Endpoint:        v.Endpoint,
			TotalRequests:   v.TotalRequests,
			SuccessRequests: v.SuccessRequests,
			ErrorRequests:   v.ErrorRequests,
			AverageLatency:  v.AverageLatency,
			MinLatency:      v.MinLatency,
			MaxLatency:      v.MaxLatency,
			LastRequest:     v.LastRequest,
		}
	}
	return copy
}

func (mc *MetricsCollector) copyAgentMetrics() map[string]*AgentMetric {
	copy := make(map[string]*AgentMetric)
	for k, v := range mc.agentMetrics {
		copy[k] = &AgentMetric{
			AgentID:           v.AgentID,
			TokensPerMinute:   v.TokensPerMinute,
			SuccessRate:       v.SuccessRate,
			ErrorCount:        v.ErrorCount,
			TotalTasks:        v.TotalTasks,
			AvgTaskTime:       v.AvgTaskTime,
			IdleTime:          v.IdleTime,
			ProductivityScore: v.ProductivityScore,
			LastActivity:      v.LastActivity,
		}
	}
	return copy
}

func (mc *MetricsCollector) copyQuestMetrics() map[string]*QuestMetric {
	copy := make(map[string]*QuestMetric)
	for k, v := range mc.questMetrics {
		copy[k] = &QuestMetric{
			QuestID:        v.QuestID,
			CompletionTime: v.CompletionTime,
			TokensUsed:     v.TokensUsed,
			Complexity:     v.Complexity,
			ResourceUsage:  v.ResourceUsage,
			ErrorCount:     v.ErrorCount,
			Status:         v.Status,
			StartTime:      v.StartTime,
			EndTime:        v.EndTime,
		}
	}
	return copy
}

// generateRecommendations analyzes metrics and provides actionable recommendations
func (mc *MetricsCollector) generateRecommendations() []PerformanceRecommendation {
	var recommendations []PerformanceRecommendation
	now := time.Now()
	
	// Check API performance
	for _, metric := range mc.apiMetrics {
		if metric.AverageLatency > 1000*time.Millisecond {
			recommendations = append(recommendations, PerformanceRecommendation{
				Type:        "api_performance",
				Priority:    "high",
				Title:       "High API Latency Detected",
				Description: fmt.Sprintf("Endpoint %s has average latency of %v", metric.Endpoint, metric.AverageLatency),
				Impact:      "Reduced user experience and system throughput",
				Action:      "Optimize endpoint implementation, add caching, or scale resources",
				CreatedAt:   now,
			})
		}
		
		errorRate := float64(metric.ErrorRequests) / float64(metric.TotalRequests) * 100
		if errorRate > 5.0 {
			recommendations = append(recommendations, PerformanceRecommendation{
				Type:        "api_reliability",
				Priority:    "critical",
				Title:       "High Error Rate Detected",
				Description: fmt.Sprintf("Endpoint %s has error rate of %.1f%%", metric.Endpoint, errorRate),
				Impact:      "Service reliability issues affecting user experience",
				Action:      "Investigate error causes and implement error handling improvements",
				CreatedAt:   now,
			})
		}
	}
	
	// Check system metrics
	if mc.systemMetrics.MemoryPercent > 80.0 {
		recommendations = append(recommendations, PerformanceRecommendation{
			Type:        "memory_usage",
			Priority:    "high",
			Title:       "High Memory Usage",
			Description: fmt.Sprintf("Memory usage is at %.1f%% of available system memory", mc.systemMetrics.MemoryPercent),
			Impact:      "Risk of out-of-memory errors and degraded performance",
			Action:      "Investigate memory leaks or consider increasing available memory",
			CreatedAt:   now,
		})
	}
	
	if mc.systemMetrics.GoroutineCount > 1000 {
		recommendations = append(recommendations, PerformanceRecommendation{
			Type:        "goroutine_count",
			Priority:    "medium",
			Title:       "High Goroutine Count",
			Description: fmt.Sprintf("Current goroutine count is %d", mc.systemMetrics.GoroutineCount),
			Impact:      "Potential goroutine leak affecting system resources",
			Action:      "Review goroutine lifecycle management and fix potential leaks",
			CreatedAt:   now,
		})
	}
	
	// Check agent performance
	for _, metric := range mc.agentMetrics {
		if metric.SuccessRate < 80.0 {
			recommendations = append(recommendations, PerformanceRecommendation{
				Type:        "agent_performance",
				Priority:    "medium",
				Title:       "Low Agent Success Rate",
				Description: fmt.Sprintf("Agent %s has success rate of %.1f%%", metric.AgentID, metric.SuccessRate),
				Impact:      "Reduced agent productivity and increased resource waste",
				Action:      "Review agent configuration and error patterns",
				CreatedAt:   now,
			})
		}
	}
	
	// Check database performance
	if mc.databaseMetrics.AverageQueryTime > 100*time.Millisecond {
		recommendations = append(recommendations, PerformanceRecommendation{
			Type:        "database_performance",
			Priority:    "medium",
			Title:       "Slow Database Queries",
			Description: fmt.Sprintf("Average query time is %v", mc.databaseMetrics.AverageQueryTime),
			Impact:      "Reduced application responsiveness",
			Action:      "Add database indexes, optimize queries, or consider query caching",
			CreatedAt:   now,
		})
	}
	
	return recommendations
}

// calculateHealthStatus determines overall system health
func (mc *MetricsCollector) calculateHealthStatus() string {
	score := 100.0
	
	// Deduct points for various issues
	if mc.systemMetrics.MemoryPercent > 90 {
		score -= 30
	} else if mc.systemMetrics.MemoryPercent > 80 {
		score -= 15
	}
	
	if mc.systemMetrics.GoroutineCount > 1000 {
		score -= 20
	}
	
	// Check API error rates
	for _, metric := range mc.apiMetrics {
		if metric.TotalRequests > 0 {
			errorRate := float64(metric.ErrorRequests) / float64(metric.TotalRequests) * 100
			if errorRate > 10 {
				score -= 25
			} else if errorRate > 5 {
				score -= 10
			}
		}
	}
	
	// Check database performance
	if mc.databaseMetrics.AverageQueryTime > 500*time.Millisecond {
		score -= 20
	} else if mc.databaseMetrics.AverageQueryTime > 100*time.Millisecond {
		score -= 10
	}
	
	// Determine status based on score
	if score >= 90 {
		return "excellent"
	} else if score >= 70 {
		return "good"
	} else if score >= 50 {
		return "fair"
	} else if score >= 30 {
		return "poor"
	} else {
		return "critical"
	}
}