import { useState, useEffect } from 'react';
import { wsManager } from '../utils/websocket';

// WebSocket connection statistics interface (matching utils/websocket.ts)
interface WSConnectionStats {
  state: 'CONNECTING' | 'OPEN' | 'CLOSING' | 'CLOSED' | 'DISCONNECTED' | 'PENDING';
  subscribers: number;
  reconnectAttempts: number;
  hasReconnectTimer: boolean;
  hasConnectionTimer: boolean;
  isTestEnvironment: boolean;
}

interface WSDebugPanelProps {
  className?: string;
}

interface TestOptimizations {
  isTestEnvironment: boolean;
  testConnectionDelay: number;
  testConnectionPersistence: number;
  testReconnectCooldown: number;
  persistentConnections: string[];
  maxReconnectAttempts: number;
  maxReconnectDelay: number;
}

export default function WSDebugPanel({ className }: WSDebugPanelProps) {
  const [stats, setStats] = useState<Record<string, WSConnectionStats>>({});
  const [testOptimizations, setTestOptimizations] = useState<TestOptimizations | null>(null);
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    const interval = setInterval(() => {
      setStats(wsManager.getStats());
      setTestOptimizations(wsManager.getTestOptimizationStatus());
    }, 1000);

    return () => clearInterval(interval);
  }, []);

  if (!isVisible) {
    return (
      <button 
        onClick={() => setIsVisible(true)}
        className="ws-debug-toggle"
        style={{
          position: 'fixed',
          bottom: '20px',
          right: '20px',
          padding: '8px',
          backgroundColor: '#333',
          color: 'white',
          border: 'none',
          borderRadius: '4px',
          fontSize: '12px',
          zIndex: 9999,
          cursor: 'pointer',
        }}
        title="Show WebSocket Debug Panel"
      >
        WS Debug
      </button>
    );
  }

  return (
    <div 
      className={className}
      style={{
        position: 'fixed',
        bottom: '20px',
        right: '20px',
        width: '400px',
        backgroundColor: 'rgba(0, 0, 0, 0.9)',
        color: 'white',
        padding: '16px',
        borderRadius: '8px',
        fontSize: '12px',
        fontFamily: 'monospace',
        zIndex: 9999,
        maxHeight: '300px',
        overflowY: 'auto',
      }}
    >
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
        <h4 style={{ margin: 0, fontSize: '14px' }}>WebSocket Debug Panel</h4>
        <button 
          onClick={() => setIsVisible(false)}
          style={{
            background: 'none',
            border: 'none',
            color: 'white',
            fontSize: '16px',
            cursor: 'pointer',
            padding: '0',
          }}
        >
          ×
        </button>
      </div>
      
      <div style={{ marginBottom: '8px' }}>
        <strong>Active Connections: {Object.keys(stats).length}</strong>
      </div>
      
      {testOptimizations && (
        <div style={{ marginBottom: '12px', padding: '8px', backgroundColor: 'rgba(34, 197, 94, 0.1)', border: '1px solid #22c55e', borderRadius: '4px' }}>
          <div style={{ color: '#22c55e', fontWeight: 'bold', marginBottom: '4px' }}>🧪 Test Environment Optimizations</div>
          <div style={{ fontSize: '10px', color: '#a3a3a3' }}>
            <div>Connection Delay: {testOptimizations.testConnectionDelay}ms</div>
            <div>Persistence: {testOptimizations.testConnectionPersistence}ms</div>
            <div>Reconnect Cooldown: {testOptimizations.testReconnectCooldown}ms</div>
            <div>Max Attempts: {testOptimizations.maxReconnectAttempts}</div>
            <div>Persistent URLs: {testOptimizations.persistentConnections.length}</div>
          </div>
        </div>
      )}
      
      {Object.entries(stats).map(([url, data]: [string, WSConnectionStats]) => (
        <div key={url} style={{ marginBottom: '12px', paddingBottom: '8px', borderBottom: '1px solid #444' }}>
          <div style={{ marginBottom: '4px' }}>
            <strong>{url.replace('ws://localhost:8080', '').replace('ws://', '') || 'Main WebSocket'}</strong>
          </div>
          <div style={{ paddingLeft: '8px' }}>
            <div>State: <span style={{ color: getStateColor(data.state) }}>{data.state}</span></div>
            <div>Subscribers: {data.subscribers}</div>
            <div>Reconnect Attempts: {data.reconnectAttempts}</div>
            <div>Has Reconnect Timer: {data.hasReconnectTimer ? 'Yes' : 'No'}</div>
          </div>
        </div>
      ))}
      
      {Object.keys(stats).length === 0 && (
        <div style={{ color: '#888' }}>No active WebSocket connections</div>
      )}
      
      <div style={{ marginTop: '12px', fontSize: '10px', color: '#888' }}>
        Updates every second. Use this to monitor connection health and identify connection issues during E2E tests.
      </div>
    </div>
  );
}

function getStateColor(state: string): string {
  switch (state) {
    case 'OPEN': return '#4ade80';
    case 'CONNECTING': return '#fbbf24';
    case 'CLOSING': return '#f87171';
    case 'CLOSED': return '#ef4444';
    case 'DISCONNECTED': return '#6b7280';
    default: return '#d1d5db';
  }
}