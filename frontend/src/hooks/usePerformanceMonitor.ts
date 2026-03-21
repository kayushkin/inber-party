import { useCallback, useEffect, useRef, useState } from 'react';
import { useOptimizedWebSocket } from '../utils/websocket';

export interface RealTimePerformanceData {
  timestamp: string;
  health_status: 'excellent' | 'good' | 'fair' | 'poor' | 'critical';
  system: {
    memory_usage: number;
    memory_percent: number;
    goroutine_count: number;
    cpu_usage: number;
    gc_count: number;
  };
  api_summary: {
    total_endpoints: number;
    total_requests: number;
    error_rate: number;
  };
  agents: {
    active_count: number;
    avg_success_rate: number;
    total_tasks: number;
  };
  websocket: {
    active_connections: number;
    connection_churn: number;
    average_latency: string;
  };
  database: {
    query_count: number;
    average_query_time: string;
    slow_queries: number;
    cache_hit_rate: number;
  };
  recommendations_count: number;
  uptime: string;
}

interface PerformanceAlert {
  alert_type: string;
  title: string;
  description: string;
  priority: 'low' | 'medium' | 'high' | 'critical';
  timestamp: string;
}

export interface PerformanceUpdate {
  type: 'performance_update' | 'performance_alert';
  data: RealTimePerformanceData | PerformanceAlert;
}

interface UsePerformanceMonitorOptions {
  fallbackToPolling?: boolean;
  pollingInterval?: number;
  onAlert?: (alert: PerformanceAlert) => void;
}

export function usePerformanceMonitor(options: UsePerformanceMonitorOptions = {}) {
  const { fallbackToPolling = true, pollingInterval = 30000, onAlert } = options;
  
  const [performanceData, setPerformanceData] = useState<RealTimePerformanceData | null>(null);
  const [alerts, setAlerts] = useState<PerformanceAlert[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdateTime, setLastUpdateTime] = useState<Date | null>(null);
  const pollingTimeoutRef = useRef<number | undefined>(undefined);
  const wsSubId = useRef(`performance-monitor-${Math.random().toString(36).substring(2, 11)}`);
  
  // Handle WebSocket messages
  const handleMessage = useCallback((data: unknown) => {
    try {
      const message = data as PerformanceUpdate;
      
      if (message.type === 'performance_update') {
        const perfData = message.data as RealTimePerformanceData;
        setPerformanceData(perfData);
        setLastUpdateTime(new Date());
        setError(null);
        
        console.log('📊 Real-time performance update received:', {
          health: perfData.health_status,
          memory: perfData.system.memory_percent.toFixed(1) + '%',
          agents: perfData.agents.active_count
        });
      } else if (message.type === 'performance_alert') {
        const alertData = message.data as PerformanceAlert;
        setAlerts(prev => [alertData, ...prev.slice(0, 9)]); // Keep last 10 alerts
        onAlert?.(alertData);
        
        console.log(`🚨 Performance alert: ${alertData.title} (${alertData.priority})`);
      }
    } catch (err) {
      console.warn('Failed to process performance WebSocket message:', err);
    }
  }, [onAlert]);
  
  // Handle WebSocket connection state changes
  const handleConnectionState = useCallback((connected: boolean) => {
    setIsConnected(connected);
    
    if (connected) {
      console.log('✅ Performance WebSocket connected - real-time updates enabled');
      setError(null);
      // Clear polling timeout since we have real-time connection
      if (pollingTimeoutRef.current) {
        window.clearTimeout(pollingTimeoutRef.current);
        pollingTimeoutRef.current = undefined;
      }
    } else {
      console.log('❌ Performance WebSocket disconnected - falling back to polling if enabled');
      setError('Real-time connection lost');
      
      // Start polling if enabled
      if (fallbackToPolling) {
        startPolling();
      }
    }
  }, [fallbackToPolling, pollingInterval]);
  
  // Polling fallback function
  const fetchPerformanceData = useCallback(async () => {
    try {
      const response = await fetch('/api/performance/snapshot');
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      
      const data = await response.json();
      
      // Convert API response to RealTimePerformanceData format
      const perfData: RealTimePerformanceData = {
        timestamp: data.timestamp,
        health_status: data.health_status,
        system: data.system,
        api_summary: data.api_summary,
        agents: {
          active_count: data.agents_active,
          avg_success_rate: 0, // Not available in snapshot
          total_tasks: 0       // Not available in snapshot
        },
        websocket: {
          active_connections: 0,
          connection_churn: 0,
          average_latency: '0ms'
        },
        database: {
          query_count: 0,
          average_query_time: '0ms',
          slow_queries: 0,
          cache_hit_rate: 0
        },
        recommendations_count: data.recommendations_count,
        uptime: data.uptime
      };
      
      setPerformanceData(perfData);
      setLastUpdateTime(new Date());
      setError(null);
      
      console.log('📊 Performance data fetched via polling fallback');
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error occurred';
      setError(errorMessage);
      console.warn('Failed to fetch performance data:', errorMessage);
    }
  }, []);
  
  // Start polling fallback
  const startPolling = useCallback(() => {
    if (pollingTimeoutRef.current) {
      window.clearTimeout(pollingTimeoutRef.current);
    }
    
    console.log(`🔄 Starting performance data polling (${pollingInterval}ms interval)`);
    
    // Fetch immediately
    fetchPerformanceData();
    
    // Schedule recurring polling
    const scheduleNext = () => {
      pollingTimeoutRef.current = window.setTimeout(() => {
        fetchPerformanceData().finally(() => {
          if (fallbackToPolling && !isConnected) {
            scheduleNext();
          }
        });
      }, pollingInterval);
    };
    
    scheduleNext();
  }, [fetchPerformanceData, pollingInterval, isConnected, fallbackToPolling]);
  
  // Initial data fetch
  useEffect(() => {
    // Fetch initial data immediately
    fetchPerformanceData();
  }, [fetchPerformanceData]);
  
  // Setup WebSocket connection
  useOptimizedWebSocket(
    'ws://localhost:8080/ws',
    wsSubId.current,
    handleMessage,
    handleConnectionState
  );
  
  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (pollingTimeoutRef.current) {
        window.clearTimeout(pollingTimeoutRef.current);
        pollingTimeoutRef.current = undefined;
      }
    };
  }, []);
  
  // Manual refresh function
  const refreshData = useCallback(async () => {
    if (isConnected) {
      console.log('🔄 Manual refresh requested - triggering immediate update');
      // For WebSocket connection, we could send a refresh request
      // For now, just log that it's handled by real-time updates
      return Promise.resolve();
    } else {
      console.log('🔄 Manual refresh requested - fetching via API');
      return fetchPerformanceData();
    }
  }, [isConnected, fetchPerformanceData]);
  
  // Clear alerts function
  const clearAlerts = useCallback(() => {
    setAlerts([]);
  }, []);
  
  // Connection status
  const connectionInfo = {
    connected: isConnected,
    lastUpdate: lastUpdateTime,
    method: isConnected ? 'WebSocket (Real-time)' : fallbackToPolling ? 'HTTP Polling' : 'Manual Only',
    error
  };
  
  return {
    data: performanceData,
    alerts,
    connectionInfo,
    refreshData,
    clearAlerts,
    isLoading: !performanceData && !error
  };
}