import { useEffect, useState, useCallback } from 'react';
import { useStore } from '../store';
import './MapView.css';

interface CodebaseNode {
  id: string;
  name: string;
  type: 'directory' | 'file';
  path: string;
  x: number;
  y: number;
  children?: CodebaseNode[];
  size?: number; // file size for visual scaling
}

interface AgentPosition {
  agentId: string;
  x: number;
  y: number;
  currentPath?: string;
}

export default function MapView() {
  const agents = useStore((s) => s.agents);
  const [codebaseNodes, setCodebaseNodes] = useState<CodebaseNode[]>([]);
  const [agentPositions, setAgentPositions] = useState<AgentPosition[]>([]);
  const [selectedNode, setSelectedNode] = useState<CodebaseNode | null>(null);
  const [loading, setLoading] = useState(true);
  const [viewBox, setViewBox] = useState({ x: 0, y: 0, width: 1200, height: 800 });

  const fetchCodebaseStructure = async () => {
    setLoading(true);
    try {
      // Mock data for now - in a real implementation, this would fetch from an API
      // that analyzes the git repository structure
      const mockNodes: CodebaseNode[] = [
        {
          id: 'root',
          name: 'inber-party',
          type: 'directory',
          path: '/',
          x: 600,
          y: 400,
          children: [
            {
              id: 'frontend',
              name: 'frontend',
              type: 'directory',
              path: '/frontend',
              x: 300,
              y: 200,
              children: [
                { id: 'src', name: 'src', type: 'directory', path: '/frontend/src', x: 200, y: 150 },
                { id: 'components', name: 'components', type: 'directory', path: '/frontend/src/components', x: 100, y: 100 },
                { id: 'pages', name: 'pages', type: 'directory', path: '/frontend/src/pages', x: 300, y: 100 }
              ]
            },
            {
              id: 'backend',
              name: 'cmd/server',
              type: 'directory',
              path: '/cmd/server',
              x: 900,
              y: 200,
              children: [
                { id: 'main.go', name: 'main.go', type: 'file', path: '/cmd/server/main.go', x: 850, y: 150, size: 500 },
                { id: 'handlers.go', name: 'handlers.go', type: 'file', path: '/cmd/server/handlers.go', x: 950, y: 150, size: 1200 }
              ]
            },
            {
              id: 'internal',
              name: 'internal',
              type: 'directory',
              path: '/internal',
              x: 600,
              y: 600,
              children: [
                { id: 'models', name: 'models', type: 'directory', path: '/internal/models', x: 500, y: 700 },
                { id: 'store', name: 'store', type: 'directory', path: '/internal/store', x: 700, y: 700 }
              ]
            }
          ]
        }
      ];
      
      setCodebaseNodes(mockNodes);
    } catch (error) {
      console.error('Failed to fetch codebase structure:', error);
    } finally {
      setLoading(false);
    }
  };

  const initializeAgentPositions = useCallback(() => {
    // Initialize agents at random positions on the map
    const positions = agents.map((agent, index) => ({
      agentId: agent.id,
      x: 200 + (index * 150) % 800,
      y: 300 + Math.floor(index / 5) * 100,
      currentPath: undefined
    }));
    setAgentPositions(positions);
  }, [agents]);

  useEffect(() => {
    // Fetch codebase structure
    fetchCodebaseStructure();
    // Initialize agent positions
    initializeAgentPositions();
  }, [initializeAgentPositions]);

  const renderCodebaseNode = (node: CodebaseNode) => {
    const isDirectory = node.type === 'directory';
    const radius = isDirectory ? 20 : Math.max(8, Math.min(15, (node.size || 100) / 100));
    const fill = isDirectory ? '#4CAF50' : '#2196F3';
    const stroke = selectedNode?.id === node.id ? '#FF9800' : '#333';
    
    return (
      <g key={node.id}>
        <circle
          cx={node.x}
          cy={node.y}
          r={radius}
          fill={fill}
          stroke={stroke}
          strokeWidth={selectedNode?.id === node.id ? 3 : 1}
          className="codebase-node"
          onClick={() => setSelectedNode(node)}
          style={{ cursor: 'pointer' }}
        />
        <text
          x={node.x}
          y={node.y - radius - 5}
          textAnchor="middle"
          className="node-label"
          fontSize="12"
          fill="#333"
        >
          {node.name}
        </text>
        
        {/* Render connections to children */}
        {node.children?.map(child => (
          <line
            key={`${node.id}-${child.id}`}
            x1={node.x}
            y1={node.y}
            x2={child.x}
            y2={child.y}
            stroke="#666"
            strokeWidth={1}
            className="connection-line"
          />
        ))}
        
        {/* Recursively render children */}
        {node.children?.map(child => renderCodebaseNode(child))}
      </g>
    );
  };

  const renderAgent = (position: AgentPosition) => {
    const agent = agents.find(a => a.id === position.agentId);
    if (!agent) return null;

    // Agent status colors
    const getAgentColor = () => {
      if (agent.status === 'working') return '#FF5722';
      if (agent.status === 'idle') return '#4CAF50';
      if (agent.status === 'stuck') return '#F44336';
      return '#9E9E9E';
    };

    return (
      <g key={agent.id} className="agent-marker">
        <circle
          cx={position.x}
          cy={position.y}
          r={12}
          fill={getAgentColor()}
          stroke="#FFF"
          strokeWidth={2}
          className="agent-avatar"
        />
        <text
          x={position.x}
          y={position.y + 25}
          textAnchor="middle"
          fontSize="10"
          fill="#333"
          className="agent-name"
        >
          {agent.name}
        </text>
        
        {/* Status indicator */}
        <circle
          cx={position.x + 8}
          cy={position.y - 8}
          r={3}
          fill={agent.status === 'working' ? '#FF9800' : 'transparent'}
          className="status-indicator"
        >
          {agent.status === 'working' && (
            <animate
              attributeName="r"
              values="3;6;3"
              dur="1.5s"
              repeatCount="indefinite"
            />
          )}
        </circle>
      </g>
    );
  };

  const handleZoom = (factor: number) => {
    setViewBox(prev => ({
      ...prev,
      width: prev.width * factor,
      height: prev.height * factor
    }));
  };

  if (loading) {
    return (
      <div className="map-view loading">
        <div className="loading-spinner"></div>
        <p>Generating world map...</p>
      </div>
    );
  }

  return (
    <div className="map-view">
      <div className="map-header">
        <h2>🗺️ Codebase Map</h2>
        <div className="map-controls">
          <button onClick={() => handleZoom(0.8)} className="zoom-btn">🔍−</button>
          <button onClick={() => handleZoom(1.25)} className="zoom-btn">🔍+</button>
          <button onClick={() => setViewBox({ x: 0, y: 0, width: 1200, height: 800 })} className="reset-btn">
            🎯 Reset View
          </button>
        </div>
      </div>

      <div className="map-container">
        <svg
          width="100%"
          height="600"
          viewBox={`${viewBox.x} ${viewBox.y} ${viewBox.width} ${viewBox.height}`}
          className="world-map"
        >
          {/* Background grid */}
          <defs>
            <pattern id="grid" width="50" height="50" patternUnits="userSpaceOnUse">
              <path d="M 50 0 L 0 0 0 50" fill="none" stroke="#e0e0e0" strokeWidth="1" />
            </pattern>
          </defs>
          <rect width="100%" height="100%" fill="url(#grid)" />
          
          {/* Render codebase structure */}
          {codebaseNodes.map(node => renderCodebaseNode(node))}
          
          {/* Render agents */}
          {agentPositions.map(position => renderAgent(position))}
        </svg>

        {/* Legend */}
        <div className="map-legend">
          <h4>Legend</h4>
          <div className="legend-item">
            <div className="legend-icon directory"></div>
            <span>Directory</span>
          </div>
          <div className="legend-item">
            <div className="legend-icon file"></div>
            <span>File</span>
          </div>
          <div className="legend-item">
            <div className="legend-icon agent working"></div>
            <span>Working Agent</span>
          </div>
          <div className="legend-item">
            <div className="legend-icon agent idle"></div>
            <span>Idle Agent</span>
          </div>
        </div>
      </div>

      {/* Selected node info */}
      {selectedNode && (
        <div className="node-info">
          <h3>{selectedNode.name}</h3>
          <p><strong>Type:</strong> {selectedNode.type}</p>
          <p><strong>Path:</strong> {selectedNode.path}</p>
          {selectedNode.size && (
            <p><strong>Size:</strong> {selectedNode.size} lines</p>
          )}
          <button onClick={() => setSelectedNode(null)}>Close</button>
        </div>
      )}
    </div>
  );
}