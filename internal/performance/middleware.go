package performance

import (
	"net/http"
	"strconv"
	"time"
)

// PerformanceMiddleware wraps HTTP handlers to track performance metrics
type PerformanceMiddleware struct {
	collector *MetricsCollector
}

// NewPerformanceMiddleware creates a new performance middleware
func NewPerformanceMiddleware(collector *MetricsCollector) *PerformanceMiddleware {
	return &PerformanceMiddleware{
		collector: collector,
	}
}

// ResponseWriter wrapper to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.written = true
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	if !rw.written {
		rw.statusCode = http.StatusOK
		rw.written = true
	}
	return rw.ResponseWriter.Write(data)
}

// WrapHandler wraps an HTTP handler with performance monitoring
func (pm *PerformanceMiddleware) WrapHandler(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		
		// Wrap the response writer to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		
		// Call the original handler
		handler(wrapped, r)
		
		// Record the performance metrics
		duration := time.Since(startTime)
		success := wrapped.statusCode >= 200 && wrapped.statusCode < 400
		
		// Use the URL path as the endpoint identifier
		endpoint := r.URL.Path
		
		// Record the API call
		pm.collector.RecordAPICall(endpoint, duration, success)
		
		// Add performance headers for debugging
		w.Header().Set("X-Response-Time", duration.String())
		w.Header().Set("X-Status", strconv.Itoa(wrapped.statusCode))
	}
}

// DatabaseQueryWrapper wraps database query functions to track performance
func (pm *PerformanceMiddleware) DatabaseQueryWrapper(queryFunc func() error) error {
	startTime := time.Now()
	err := queryFunc()
	duration := time.Since(startTime)
	
	pm.collector.RecordDatabaseQuery(duration, err == nil)
	return err
}

// AgentActivityWrapper wraps agent activity functions to track performance
func (pm *PerformanceMiddleware) AgentActivityWrapper(agentID string, tokensUsed int, activity func() error) error {
	startTime := time.Now()
	err := activity()
	duration := time.Since(startTime)
	
	pm.collector.RecordAgentActivity(agentID, tokensUsed, duration, err == nil)
	return err
}