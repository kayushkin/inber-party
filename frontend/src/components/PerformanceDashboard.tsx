import { useState, useEffect } from 'react';
import { usePerformanceMonitor } from '../hooks/usePerformanceMonitor';
import './PerformanceDashboard.css';

interface SystemMetrics {
  cpu_usage: number;
  memory_usage: number;
  memory_percent: number;
  goroutine_count: number;
  gc_count: number;
  last_updated: string;
}

interface APIMetric {
  endpoint: string;
  total_requests: number;
  success_requests: number;
  error_requests: number;
  average_latency: string;
  min_latency: string;
  max_latency: string;
  last_request: string;
}

interface AgentMetric {
  agent_id: string;
  tokens_per_minute: number;
  success_rate: number;
  error_count: number;
  total_tasks: number;
  avg_task_time: string;
  productivity_score: number;
  last_activity: string;
}

interface PerformanceRecommendation {
  type: string;
  priority: 'low' | 'medium' | 'high' | 'critical';
  title: string;
  description: string;
  impact: string;
  action: string;
  created_at: string;
}

interface PerformanceSnapshot {
  timestamp: string;
  uptime: string;
  health_status: 'excellent' | 'good' | 'fair' | 'poor' | 'critical';
  system_metrics: SystemMetrics;
  api_metrics: Record<string, APIMetric>;
  agent_metrics: Record<string, AgentMetric>;
  recommendations: PerformanceRecommendation[];
}

export default function PerformanceDashboard() {
  const [fullSnapshot, setFullSnapshot] = useState<PerformanceSnapshot | null>(null);
  const [selectedTab, setSelectedTab] = useState<'overview' | 'api' | 'agents' | 'recommendations'>('overview');
  const [showRealtimeIndicator, setShowRealtimeIndicator] = useState(false);

  // Use the new real-time performance monitor
  const {
    data: realtimeData,
    alerts,
    connectionInfo,
    refreshData,
    clearAlerts,
    isLoading
  } = usePerformanceMonitor({
    fallbackToPolling: true,
    pollingInterval: 30000,
    onAlert: (alert) => {
      console.log(`📢 Performance Alert: ${alert.title} (${alert.priority})`);
      // Optional: Show toast notification for critical alerts
      if (alert.priority === 'critical') {
        // You could integrate with a notification system here
      }
    }
  });

  // Flash indicator when real-time data updates
  useEffect(() => {
    if (realtimeData && connectionInfo.connected) {
      setShowRealtimeIndicator(true);
      const timer = setTimeout(() => setShowRealtimeIndicator(false), 1000);
      return () => clearTimeout(timer);
    }
  }, [realtimeData, connectionInfo.connected]);

  // Fetch full snapshot for detailed view (API, Agents, Recommendations)
  const fetchFullSnapshot = async () => {
    try {
      const response = await fetch('/api/performance/metrics');
      if (!response.ok) {
        throw new Error('Failed to fetch full performance data');
      }
      const data = await response.json();
      setFullSnapshot(data);
    } catch (err) {
      console.warn('Failed to fetch full snapshot:', err);
    }
  };

  // Fetch full snapshot on mount and when tab changes to detailed views
  useEffect(() => {
    if (selectedTab !== 'overview') {
      fetchFullSnapshot();
    }
  }, [selectedTab]);

  // Manual refresh function
  const handleRefresh = async () => {
    await refreshData();
    if (selectedTab !== 'overview') {
      await fetchFullSnapshot();
    }
  };

  const formatLatency = (latencyStr: string) => {
    if (latencyStr.includes('ms')) {
      const ms = parseFloat(latencyStr);
      return ms < 1000 ? `${ms.toFixed(1)}ms` : `${(ms / 1000).toFixed(2)}s`;
    }
    return latencyStr;
  };

  const getHealthColor = (status: string) => {
    switch (status) {
      case 'excellent': return '#4ade80'; // green
      case 'good': return '#22d3ee'; // cyan
      case 'fair': return '#fbbf24'; // yellow
      case 'poor': return '#fb7185'; // pink
      case 'critical': return '#ef4444'; // red
      default: return '#6b7280'; // gray
    }
  };

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'critical': return '#ef4444';
      case 'high': return '#f97316';
      case 'medium': return '#eab308';
      case 'low': return '#22c55e';
      default: return '#6b7280';
    }
  };

  const formatBytes = (bytes: number) => {
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    if (bytes === 0) return '0 Bytes';
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
  };

  const formatUptime = (uptime: string) => {
    // Parse nanosecond string like "2h45m30.123456789s"
    const match = uptime.match(/(\d+)h(\d+)m([\d.]+)s/);
    if (match) {
      const [, hours, minutes, seconds] = match;
      const secs = Math.floor(parseFloat(seconds));
      return `${hours}h ${minutes}m ${secs}s`;
    }
    return uptime;
  };

  if (isLoading && !realtimeData) {
    return (
      <div className="performance-dashboard">
        <div className="loading-state">
          <div className="loading-spinner"></div>
          <p>Loading performance metrics...</p>
        </div>
      </div>
    );
  }

  if (connectionInfo.error && !realtimeData) {
    return (
      <div className="performance-dashboard">
        <div className="error-state">
          <h3>⚠️ Error Loading Performance Data</h3>
          <p>{connectionInfo.error}</p>
          <button onClick={handleRefresh} className="retry-button">
            🔄 Retry
          </button>
        </div>
      </div>
    );
  }

  if (!realtimeData) {
    return <div className="performance-dashboard">No performance data available</div>;
  }

  // Use full snapshot data for detailed views, real-time data for overview
  const apiMetrics = fullSnapshot ? Object.values(fullSnapshot.api_metrics) : [];
  const agentMetrics = fullSnapshot ? Object.values(fullSnapshot.agent_metrics) : [];
  const recommendations = fullSnapshot ? fullSnapshot.recommendations : [];

  // Health status is now derived directly from real-time data

  return (
    <div className="performance-dashboard">
      <div className="dashboard-header">
        <div className="header-info">
          <h2>🚀 Performance Metrics Dashboard</h2>
          <div className="header-stats">
            <div className="health-indicator">
              <span 
                className="health-dot" 
                style={{ backgroundColor: getHealthColor(realtimeData.health_status) }}
              ></span>
              <span className="health-text">
                {realtimeData.health_status.charAt(0).toUpperCase() + realtimeData.health_status.slice(1)}
              </span>
            </div>
            <div className="uptime-info">
              ⏱️ Uptime: {formatUptime(realtimeData.uptime)}
            </div>
            <div className="connection-status">
              <span 
                className={`connection-dot ${connectionInfo.connected ? 'connected' : 'disconnected'} ${showRealtimeIndicator ? 'flash' : ''}`}
                style={{ 
                  backgroundColor: connectionInfo.connected ? '#4ade80' : '#ef4444',
                  animation: showRealtimeIndicator ? 'flash 0.5s ease-out' : 'none'
                }}
              ></span>
              <span className="connection-text" title={connectionInfo.method}>
                {connectionInfo.connected ? '🟢 Real-time' : '🔶 Polling'}
              </span>
              {connectionInfo.lastUpdate && (
                <span className="last-update">
                  Last: {connectionInfo.lastUpdate.toLocaleTimeString()}
                </span>
              )}
            </div>
          </div>
        </div>
        <div className="dashboard-controls">
          <div className="connection-method">
            <small>{connectionInfo.method}</small>
          </div>
          {alerts.length > 0 && (
            <button onClick={clearAlerts} className="clear-alerts-button">
              🔔 Clear Alerts ({alerts.length})
            </button>
          )}
          <button onClick={handleRefresh} className="refresh-button" disabled={isLoading}>
            {isLoading ? '🔄' : '↻'} Refresh
          </button>
        </div>
      </div>

      <div className="dashboard-tabs">
        <button 
          className={`tab-button ${selectedTab === 'overview' ? 'active' : ''}`}
          onClick={() => setSelectedTab('overview')}
        >
          📊 Overview
        </button>
        <button 
          className={`tab-button ${selectedTab === 'api' ? 'active' : ''}`}
          onClick={() => setSelectedTab('api')}
        >
          🌐 API Performance
        </button>
        <button 
          className={`tab-button ${selectedTab === 'agents' ? 'active' : ''}`}
          onClick={() => setSelectedTab('agents')}
        >
          🤖 Agent Metrics
        </button>
        <button 
          className={`tab-button ${selectedTab === 'recommendations' ? 'active' : ''}`}
          onClick={() => setSelectedTab('recommendations')}
        >
          💡 Recommendations ({recommendations.length})
        </button>
      </div>

      <div className="dashboard-content">
        {selectedTab === 'overview' && (
          <div className="overview-tab">
            <div className="metrics-grid">
              <div className="metric-card">
                <h3>🖥️ System Resources</h3>
                <div className="metric-item">
                  <span className="metric-label">Memory Usage:</span>
                  <span className="metric-value">
                    {formatBytes(realtimeData.system.memory_usage)} 
                    ({realtimeData.system.memory_percent.toFixed(1)}%)
                  </span>
                </div>
                <div className="metric-item">
                  <span className="metric-label">Goroutines:</span>
                  <span className="metric-value">{realtimeData.system.goroutine_count}</span>
                </div>
                <div className="metric-item">
                  <span className="metric-label">CPU Usage:</span>
                  <span className="metric-value">{realtimeData.system.cpu_usage.toFixed(1)}%</span>
                </div>
                <div className="metric-item">
                  <span className="metric-label">GC Count:</span>
                  <span className="metric-value">{realtimeData.system.gc_count}</span>
                </div>
              </div>

              <div className="metric-card">
                <h3>🌐 API Performance</h3>
                <div className="metric-item">
                  <span className="metric-label">Total Endpoints:</span>
                  <span className="metric-value">{realtimeData.api_summary.total_endpoints}</span>
                </div>
                <div className="metric-item">
                  <span className="metric-label">Total Requests:</span>
                  <span className="metric-value">
                    {realtimeData.api_summary.total_requests.toLocaleString()}
                  </span>
                </div>
                <div className="metric-item">
                  <span className="metric-label">Error Rate:</span>
                  <span className="metric-value">{realtimeData.api_summary.error_rate.toFixed(2)}%</span>
                </div>
                <div className="metric-item">
                  <span className="metric-label">DB Query Time:</span>
                  <span className="metric-value">{realtimeData.database.average_query_time}</span>
                </div>
              </div>

              <div className="metric-card">
                <h3>🤖 Agent Activity</h3>
                <div className="metric-item">
                  <span className="metric-label">Active Agents:</span>
                  <span className="metric-value">{realtimeData.agents.active_count}</span>
                </div>
                <div className="metric-item">
                  <span className="metric-label">Total Tasks:</span>
                  <span className="metric-value">
                    {realtimeData.agents.total_tasks.toLocaleString()}
                  </span>
                </div>
                <div className="metric-item">
                  <span className="metric-label">Avg Success Rate:</span>
                  <span className="metric-value">
                    {realtimeData.agents.avg_success_rate > 0 
                      ? realtimeData.agents.avg_success_rate.toFixed(1) + '%'
                      : 'N/A'
                    }
                  </span>
                </div>
                <div className="metric-item">
                  <span className="metric-label">WebSocket Connections:</span>
                  <span className="metric-value">{realtimeData.websocket.active_connections}</span>
                </div>
              </div>

              <div className="metric-card">
                <h3>🔥 Performance Alerts & Recommendations</h3>
                <div className="recommendations-summary">
                  {alerts.filter(a => a.priority === 'critical').length > 0 && (
                    <div className="alert-item critical">
                      🔴 {alerts.filter(a => a.priority === 'critical').length} Critical Alerts
                    </div>
                  )}
                  {alerts.filter(a => a.priority === 'high').length > 0 && (
                    <div className="alert-item high">
                      🟠 {alerts.filter(a => a.priority === 'high').length} High Priority Alerts
                    </div>
                  )}
                  {realtimeData.recommendations_count > 0 ? (
                    <div className="alert-item medium">
                      💡 {realtimeData.recommendations_count} Recommendations Available
                    </div>
                  ) : (
                    <div className="alert-item success">
                      ✅ No performance issues detected
                    </div>
                  )}
                  {alerts.length > 0 && (
                    <div className="alert-item info">
                      🔔 {alerts.length} Recent Alerts (Real-time)
                    </div>
                  )}
                </div>
              </div>

              <div className="metric-card">
                <h3>📊 Real-time Metrics</h3>
                <div className="metric-item">
                  <span className="metric-label">Connection:</span>
                  <span className="metric-value">
                    {connectionInfo.connected ? '🟢 Live' : '🔶 Polling'}
                  </span>
                </div>
                <div className="metric-item">
                  <span className="metric-label">Database Queries:</span>
                  <span className="metric-value">{realtimeData.database.query_count.toLocaleString()}</span>
                </div>
                <div className="metric-item">
                  <span className="metric-label">Slow Queries:</span>
                  <span className="metric-value">{realtimeData.database.slow_queries}</span>
                </div>
                <div className="metric-item">
                  <span className="metric-label">Cache Hit Rate:</span>
                  <span className="metric-value">{realtimeData.database.cache_hit_rate.toFixed(1)}%</span>
                </div>
              </div>

              <div className="metric-card">
                <h3>🌐 WebSocket Performance</h3>
                <div className="metric-item">
                  <span className="metric-label">Active Connections:</span>
                  <span className="metric-value">{realtimeData.websocket.active_connections}</span>
                </div>
                <div className="metric-item">
                  <span className="metric-label">Connection Churn:</span>
                  <span className="metric-value">{realtimeData.websocket.connection_churn}</span>
                </div>
                <div className="metric-item">
                  <span className="metric-label">Avg Latency:</span>
                  <span className="metric-value">{realtimeData.websocket.average_latency}</span>
                </div>
                <div className="metric-item">
                  <span className="metric-label">Update Method:</span>
                  <span className="metric-value">{connectionInfo.method}</span>
                </div>
              </div>
            </div>
          </div>
        )}

        {selectedTab === 'api' && (
          <div className="api-tab">
            <h3>🌐 API Endpoint Performance</h3>
            <div className="api-metrics-table">
              <div className="table-header">
                <span>Endpoint</span>
                <span>Requests</span>
                <span>Success Rate</span>
                <span>Avg Latency</span>
                <span>Min/Max</span>
                <span>Last Request</span>
              </div>
              {apiMetrics.map((api, index) => (
                <div key={index} className="table-row">
                  <span className="endpoint-path">{api.endpoint}</span>
                  <span className="requests-count">{api.total_requests.toLocaleString()}</span>
                  <span className="success-rate">
                    {api.total_requests > 0 ? 
                      ((api.success_requests / api.total_requests) * 100).toFixed(1) + '%'
                      : 'N/A'
                    }
                  </span>
                  <span className="latency">{formatLatency(api.average_latency)}</span>
                  <span className="latency-range">
                    {formatLatency(api.min_latency)} - {formatLatency(api.max_latency)}
                  </span>
                  <span className="last-request">
                    {new Date(api.last_request).toLocaleString()}
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}

        {selectedTab === 'agents' && (
          <div className="agents-tab">
            <h3>🤖 Agent Performance Metrics</h3>
            <div className="agents-grid">
              {agentMetrics.map((agent, index) => (
                <div key={index} className="agent-card">
                  <div className="agent-header">
                    <h4>🤖 {agent.agent_id}</h4>
                    <div className="productivity-score">
                      Score: {agent.productivity_score.toFixed(1)}
                    </div>
                  </div>
                  <div className="agent-stats">
                    <div className="stat-item">
                      <span className="stat-label">Success Rate:</span>
                      <span className="stat-value">{agent.success_rate.toFixed(1)}%</span>
                    </div>
                    <div className="stat-item">
                      <span className="stat-label">Tokens/Min:</span>
                      <span className="stat-value">{agent.tokens_per_minute.toFixed(1)}</span>
                    </div>
                    <div className="stat-item">
                      <span className="stat-label">Total Tasks:</span>
                      <span className="stat-value">{agent.total_tasks}</span>
                    </div>
                    <div className="stat-item">
                      <span className="stat-label">Errors:</span>
                      <span className="stat-value">{agent.error_count}</span>
                    </div>
                    <div className="stat-item">
                      <span className="stat-label">Avg Task Time:</span>
                      <span className="stat-value">{agent.avg_task_time}</span>
                    </div>
                    <div className="stat-item">
                      <span className="stat-label">Last Activity:</span>
                      <span className="stat-value">
                        {new Date(agent.last_activity).toLocaleString()}
                      </span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {selectedTab === 'recommendations' && (
          <div className="recommendations-tab">
            <h3>💡 Performance Recommendations</h3>
            {recommendations.length === 0 ? (
              <div className="no-recommendations">
                ✅ Excellent! No performance recommendations at this time.
              </div>
            ) : (
              <div className="recommendations-list">
                {recommendations.map((rec, index) => (
                  <div key={index} className={`recommendation-card priority-${rec.priority}`}>
                    <div className="recommendation-header">
                      <span 
                        className="priority-badge"
                        style={{ backgroundColor: getPriorityColor(rec.priority) }}
                      >
                        {rec.priority.toUpperCase()}
                      </span>
                      <h4>{rec.title}</h4>
                    </div>
                    <div className="recommendation-body">
                      <p className="description">{rec.description}</p>
                      <div className="impact">
                        <strong>Impact:</strong> {rec.impact}
                      </div>
                      <div className="action">
                        <strong>Recommended Action:</strong> {rec.action}
                      </div>
                    </div>
                    <div className="recommendation-footer">
                      <small>Type: {rec.type} • {new Date(rec.created_at).toLocaleString()}</small>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>

      <div className="dashboard-footer">
        <small>
          Last updated: {new Date(realtimeData.timestamp).toLocaleString()} • 
          Mode: {connectionInfo.method} • 
          Status: {connectionInfo.connected ? 'Connected' : connectionInfo.error || 'Disconnected'}
          {connectionInfo.connected && ' • Real-time updates active'}
        </small>
      </div>
    </div>
  );
}