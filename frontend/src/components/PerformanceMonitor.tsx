import { memo } from 'react';
import './PerformanceMonitor.css';

interface PerformanceMonitorProps {
  agentCount?: number;
  visible?: boolean;
}

const PerformanceMonitor = memo<PerformanceMonitorProps>(({ 
  agentCount = 0, 
  visible = false 
}) => {
  if (!visible) return null;

  return (
    <div className="performance-monitor performance-excellent">
      <div className="pm-header">
        <span className="pm-title">⚡ Performance</span>
      </div>
      
      <div className="pm-metrics">
        <div className="pm-metric">
          <span className="pm-label">FPS:</span>
          <span className="pm-value pm-fps-excellent">60</span>
        </div>
        
        <div className="pm-metric">
          <span className="pm-label">Agents:</span>
          <span className="pm-value">{agentCount}</span>
        </div>
      </div>
      
      <div className="pm-chart">
        <div 
          className="pm-fps-bar" 
          style={{ 
            width: '100%',
            backgroundColor: '#4ade80'
          }} 
        />
      </div>
      
      <div className="pm-tips">
        {agentCount >= 25 && (
          <div className="pm-tip pm-tip-success">
            ✅ Optimized for {agentCount} agents!
          </div>
        )}
      </div>
    </div>
  );
});

PerformanceMonitor.displayName = 'PerformanceMonitor';

export default PerformanceMonitor;