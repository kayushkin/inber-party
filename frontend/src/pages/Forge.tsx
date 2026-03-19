import { useEffect, useMemo } from 'react';
import { useStore, timeAgo } from '../store';
import Tooltip from '../components/Tooltip';
import './Forge.css';

export default function Forge() {
  const { healthData, isLoadingHealth, fetchHealthData } = useStore();

  useEffect(() => {
    // Fetch health data on component mount
    fetchHealthData();
    
    // Set up polling for health data every 30 seconds
    const healthInterval = setInterval(fetchHealthData, 30000);
    return () => clearInterval(healthInterval);
  }, [fetchHealthData]);

  // Use real health data or fallback to mock data for development
  const infrastructureServices = healthData?.services || [];

  // Use useMemo to create build logs with stable timestamps
  const buildLogs = useMemo(() => {
    const now = Date.now();
    return [
      {
        id: 1,
        service: 'inber-party-frontend',
        status: 'success',
        duration: '2m 14s',
        timestamp: new Date(now - 1000 * 60 * 60 * 2),
        commit: '7f3a9b2',
        message: 'Add Forge room implementation'
      },
      {
        id: 2,
        service: 'inber-party-backend',
        status: 'success',
        duration: '1m 45s',
        timestamp: new Date(now - 1000 * 60 * 60 * 6),
        commit: '4e8c1d5',
        message: 'Update quest API endpoints'
      },
      {
        id: 3,
        service: 'openclaw-gateway',
        status: 'warning',
        duration: '3m 22s',
        timestamp: new Date(now - 1000 * 60 * 60 * 24 * 2),
        commit: 'a9f2c7e',
        message: 'Fix memory leak in WebSocket handler'
      }
    ];
  }, []); // Empty dependency array means this will only run once

  const systemMetrics = {
    totalUptime: '99.2%',
    avgResponseTime: '247ms',
    requestsToday: 8547,
    errorsToday: 23,
    diskUsage: 68,
    memoryUsage: 72,
    cpuUsage: 45
  };

  const healthyServices = healthData?.overall.healthy || 0;
  const totalServices = healthData?.overall.total || infrastructureServices.length;
  const avgHealth = healthData?.overall.score || 0;

  return (
    <div className="forge">
      <div className="forge-header">
        <h1>🔨 The Forge</h1>
        <div className="forge-subtitle">Infrastructure & Deployment Command</div>
      </div>

      {/* System Overview */}
      <div className="system-overview">
        <div className="overview-stats">
          <div className="stat-card">
            <div className="stat-icon">⚡</div>
            <div className="stat-content">
              <span className="stat-number">{healthyServices}/{totalServices}</span>
              <span className="stat-label">Services Online</span>
            </div>
          </div>
          <div className="stat-card">
            <div className="stat-icon">❤️</div>
            <div className="stat-content">
              <span className="stat-number">{avgHealth.toFixed(1)}%</span>
              <span className="stat-label">Avg Health</span>
            </div>
          </div>
          <div className="stat-card">
            <div className="stat-icon">⏱️</div>
            <div className="stat-content">
              <span className="stat-number">{systemMetrics.avgResponseTime}</span>
              <span className="stat-label">Avg Response</span>
            </div>
          </div>
          <div className="stat-card">
            <div className="stat-icon">📈</div>
            <div className="stat-content">
              <span className="stat-number">{systemMetrics.totalUptime}</span>
              <span className="stat-label">Uptime</span>
              <div className="stat-subtitle">Last 30 days</div>
            </div>
          </div>
        </div>
      </div>

      {/* Infrastructure Status */}
      <div className="infrastructure-status">
        <h2>🏗️ Infrastructure Status</h2>
        <div className="services-grid">
          {isLoadingHealth ? (
            <div className="loading-services">
              <div className="loading-text">🔍 Checking service health...</div>
            </div>
          ) : infrastructureServices.length === 0 ? (
            <div className="no-services">
              <div className="no-services-icon">⚠️</div>
              <div className="no-services-text">No health data available</div>
              <div className="no-services-subtitle">Health check service may be down</div>
            </div>
          ) : (
            infrastructureServices.map((service) => (
              <Tooltip 
                key={service.name}
                content={`${service.name}${service.version ? ` v${service.version}` : ''}${service.uptime ? ` • Uptime: ${service.uptime}` : ''} • Health: ${service.health}%${service.response_time ? ` • Response: ${service.response_time}ms` : ''}${service.error ? ` • Error: ${service.error}` : ''}`}
              >
                <div className={`service-card status-${service.status}`}>
                  <div className="service-header">
                    <div className="service-name">{service.name}</div>
                    <div className="service-status">
                      <span className={`status-indicator status-${service.status}`}>
                        {service.status === 'running' ? '🟢' : service.status === 'warning' ? '🟡' : '🔴'}
                      </span>
                      <span className="status-text">{service.status}</span>
                    </div>
                  </div>
                  
                  <div className="service-details">
                    {service.port && (
                      <div className="detail-row">
                        <span className="detail-label">Port:</span>
                        <span className="detail-value">{service.port}</span>
                      </div>
                    )}
                    {service.version && (
                      <div className="detail-row">
                        <span className="detail-label">Version:</span>
                        <span className="detail-value">{service.version}</span>
                      </div>
                    )}
                    {service.uptime && (
                      <div className="detail-row">
                        <span className="detail-label">Uptime:</span>
                        <span className="detail-value">{service.uptime}</span>
                      </div>
                    )}
                    {service.response_time && (
                      <div className="detail-row">
                        <span className="detail-label">Response:</span>
                        <span className="detail-value">{service.response_time}ms</span>
                      </div>
                    )}
                    <div className="detail-row">
                      <span className="detail-label">Last Check:</span>
                      <span className="detail-value">{timeAgo(service.last_check)}</span>
                    </div>
                  </div>

                  <div className="service-health">
                    <div className="health-header">
                      <span className="health-label">Health</span>
                      <span className="health-percentage">{service.health}%</span>
                    </div>
                    <div className="health-bar">
                      <div 
                        className={`health-fill ${service.health < 90 ? 'health-warning' : 'health-good'}`}
                        style={{ width: `${service.health}%` }}
                      />
                    </div>
                    {service.error && (
                      <div className="service-error">
                        <span className="error-text">{service.error}</span>
                      </div>
                    )}
                  </div>
                </div>
              </Tooltip>
            ))
          )}
        </div>
      </div>

      {/* System Metrics */}
      <div className="system-metrics">
        <h2>📊 System Metrics</h2>
        <div className="metrics-grid">
          <div className="metric-card">
            <div className="metric-header">
              <span className="metric-label">CPU Usage</span>
              <span className="metric-value">{systemMetrics.cpuUsage}%</span>
            </div>
            <div className="metric-bar">
              <div 
                className={`metric-fill ${systemMetrics.cpuUsage > 80 ? 'metric-critical' : systemMetrics.cpuUsage > 60 ? 'metric-warning' : 'metric-good'}`}
                style={{ width: `${systemMetrics.cpuUsage}%` }}
              />
            </div>
          </div>
          
          <div className="metric-card">
            <div className="metric-header">
              <span className="metric-label">Memory Usage</span>
              <span className="metric-value">{systemMetrics.memoryUsage}%</span>
            </div>
            <div className="metric-bar">
              <div 
                className={`metric-fill ${systemMetrics.memoryUsage > 80 ? 'metric-critical' : systemMetrics.memoryUsage > 60 ? 'metric-warning' : 'metric-good'}`}
                style={{ width: `${systemMetrics.memoryUsage}%` }}
              />
            </div>
          </div>

          <div className="metric-card">
            <div className="metric-header">
              <span className="metric-label">Disk Usage</span>
              <span className="metric-value">{systemMetrics.diskUsage}%</span>
            </div>
            <div className="metric-bar">
              <div 
                className={`metric-fill ${systemMetrics.diskUsage > 80 ? 'metric-critical' : systemMetrics.diskUsage > 60 ? 'metric-warning' : 'metric-good'}`}
                style={{ width: `${systemMetrics.diskUsage}%` }}
              />
            </div>
          </div>

          <div className="metric-summary">
            <div className="summary-stat">
              <span className="summary-number">{systemMetrics.requestsToday.toLocaleString()}</span>
              <span className="summary-label">Requests Today</span>
            </div>
            <div className="summary-stat">
              <span className={`summary-number ${systemMetrics.errorsToday > 50 ? 'error-high' : 'error-low'}`}>
                {systemMetrics.errorsToday}
              </span>
              <span className="summary-label">Errors Today</span>
            </div>
          </div>
        </div>
      </div>

      {/* Recent Deployments */}
      <div className="recent-deployments">
        <h2>🚀 Recent Deployments</h2>
        {buildLogs.length === 0 ? (
          <div className="no-deployments">
            <div className="no-deployments-icon">🏭</div>
            <div className="no-deployments-text">No recent deployments</div>
            <div className="no-deployments-subtitle">The forge is quiet for now</div>
          </div>
        ) : (
          <div className="deployments-list">
            {buildLogs.map((log) => (
              <div key={log.id} className={`deployment-card status-${log.status}`}>
                <div className="deployment-header">
                  <div className="deployment-service">{log.service}</div>
                  <div className={`deployment-status status-${log.status}`}>
                    {log.status === 'success' ? '✅' : log.status === 'warning' ? '⚠️' : '❌'}
                    <span className="status-text">{log.status}</span>
                  </div>
                  <div className="deployment-time">{timeAgo(log.timestamp.toISOString())}</div>
                </div>
                
                <div className="deployment-details">
                  <div className="deployment-commit">
                    <span className="commit-hash">#{log.commit}</span>
                    <span className="commit-message">{log.message}</span>
                  </div>
                  
                  <div className="deployment-metrics">
                    <span className="duration">⏱️ {log.duration}</span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}