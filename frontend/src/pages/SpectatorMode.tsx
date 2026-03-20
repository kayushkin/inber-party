import React, { useState, useEffect, useCallback } from 'react';
import { useStore } from '../store';
import { 
  FaPlay, 
  FaPause, 
  FaEye,
  FaUserNinja,
  FaBolt,
  FaCogs,
  FaTerminal,
  FaDatabase,
  FaSearch,
  FaCode
} from 'react-icons/fa';
import './SpectatorMode.css';

interface LiveSession {
  id: string;
  agent_id: string;
  agent_name: string;
  status: string;
  started_at: string;
  current_tool?: string;
  current_action?: string;
  progress: number;
  energy: number;
  max_energy: number;
  turns: number;
  tokens_used: number;
  cost: number;
  live_output?: string;
}

interface ToolCall {
  id: string;
  tool_name: string;
  parameters: Record<string, unknown>;
  timestamp: string;
  duration?: number;
  success?: boolean;
}

const SpectatorMode: React.FC = () => {
  const [selectedAgent, setSelectedAgent] = useState<string>('');
  const [isWatching, setIsWatching] = useState<boolean>(false);
  const [liveSession, setLiveSession] = useState<LiveSession | null>(null);
  const [recentToolCalls, setRecentToolCalls] = useState<ToolCall[]>([]);
  const [showAllAgents, setShowAllAgents] = useState<boolean>(true);
  const [refreshInterval, setRefreshInterval] = useState<number>(1000); // 1 second

  const { agents } = useStore();

  const activeAgents = agents.filter(agent => 
    agent.status === 'working' || agent.status === 'thinking'
  );

  const fetchLiveData = useCallback(async () => {
    if (!selectedAgent && !showAllAgents) return;

    try {
      const endpoint = showAllAgents 
        ? '/api/spectator/live-sessions'
        : `/api/spectator/live-session/${selectedAgent}`;
      
      const response = await fetch(endpoint);
      if (response.ok) {
        const data = await response.json();
        if (showAllAgents) {
          // Handle multiple sessions
          console.log('Live sessions:', data);
        } else {
          setLiveSession(data);
        }
      }

      // Fetch recent tool calls
      const toolCallsEndpoint = showAllAgents
        ? '/api/spectator/recent-tool-calls'
        : `/api/spectator/recent-tool-calls/${selectedAgent}`;
      
      const toolCallsResponse = await fetch(toolCallsEndpoint);
      if (toolCallsResponse.ok) {
        const toolCallsData = await toolCallsResponse.json();
        setRecentToolCalls(toolCallsData);
      }
    } catch (error) {
      console.error('Failed to fetch live data:', error);
    }
  }, [selectedAgent, showAllAgents]);

  useEffect(() => {
    if (isWatching) {
      const interval = setInterval(fetchLiveData, refreshInterval);
      return () => clearInterval(interval);
    }
  }, [isWatching, selectedAgent, refreshInterval, fetchLiveData]);

  const startWatching = () => {
    setIsWatching(true);
    fetchLiveData();
  };

  const stopWatching = () => {
    setIsWatching(false);
    setLiveSession(null);
    setRecentToolCalls([]);
  };

  const getToolIcon = (toolName: string) => {
    switch (toolName.toLowerCase()) {
      case 'exec': case 'process': return <FaTerminal />;
      case 'read': case 'write': case 'edit': return <FaCode />;
      case 'web_search': case 'web_fetch': return <FaSearch />;
      case 'browser': return <FaEye />;
      case 'nodes': return <FaDatabase />;
      default: return <FaCogs />;
    }
  };

  const getEnergyColor = (energy: number, maxEnergy: number) => {
    const percentage = (energy / maxEnergy) * 100;
    if (percentage > 70) return '#22c55e'; // green
    if (percentage > 30) return '#f59e0b'; // yellow
    return '#ef4444'; // red
  };

  return (
    <div className="spectator-mode">
      <div className="spectator-header">
        <h1>
          <FaEye className="spectator-icon" />
          Spectator Mode
        </h1>
        <p className="spectator-subtitle">
          Watch your agents work in real-time with RPG overlay
        </p>
      </div>

      <div className="spectator-controls">
        <div className="agent-selector">
          <label>
            <input
              type="radio"
              name="watch-mode"
              checked={showAllAgents}
              onChange={() => setShowAllAgents(true)}
            />
            Watch All Active Agents
          </label>
          <label>
            <input
              type="radio"
              name="watch-mode"
              checked={!showAllAgents}
              onChange={() => setShowAllAgents(false)}
            />
            Watch Specific Agent:
          </label>
          {!showAllAgents && (
            <select 
              value={selectedAgent} 
              onChange={(e) => setSelectedAgent(e.target.value)}
              disabled={showAllAgents}
            >
              <option value="">Select an agent...</option>
              {agents.map(agent => (
                <option key={agent.id} value={agent.id}>
                  {agent.name} ({agent.status})
                </option>
              ))}
            </select>
          )}
        </div>

        <div className="watch-controls">
          <button
            onClick={isWatching ? stopWatching : startWatching}
            className={`watch-btn ${isWatching ? 'watching' : ''}`}
            disabled={!showAllAgents && !selectedAgent}
          >
            {isWatching ? <FaPause /> : <FaPlay />}
            {isWatching ? 'Stop Watching' : 'Start Watching'}
          </button>
          
          <div className="refresh-rate">
            <label>Refresh Rate:</label>
            <select 
              value={refreshInterval} 
              onChange={(e) => setRefreshInterval(Number(e.target.value))}
            >
              <option value={500}>0.5s</option>
              <option value={1000}>1s</option>
              <option value={2000}>2s</option>
              <option value={5000}>5s</option>
            </select>
          </div>
        </div>
      </div>

      {isWatching && (
        <div className="spectator-display">
          {showAllAgents ? (
            <div className="multi-agent-view">
              <h2>Active Agents ({activeAgents.length})</h2>
              <div className="agent-grid">
                {activeAgents.map(agent => (
                  <div key={agent.id} className="agent-card-live">
                    <div className="agent-header">
                      <span className="agent-avatar">{agent.avatar_emoji}</span>
                      <div className="agent-info">
                        <h3>{agent.name}</h3>
                        <span className={`status status-${agent.status}`}>
                          {agent.status}
                        </span>
                      </div>
                    </div>
                    
                    <div className="agent-stats">
                      <div className="energy-bar">
                        <label>Energy:</label>
                        <div className="bar-container">
                          <div 
                            className="bar-fill"
                            style={{ 
                              width: `${(agent.energy / agent.max_energy) * 100}%`,
                              backgroundColor: getEnergyColor(agent.energy, agent.max_energy)
                            }}
                          />
                        </div>
                        <span>{agent.energy}/{agent.max_energy}</span>
                      </div>
                      
                      <div className="agent-activity">
                        <FaUserNinja /> Level {agent.level}
                        <FaBolt /> {agent.total_tokens} tokens
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ) : (
            liveSession && (
              <div className="single-agent-view">
                <div className="session-header">
                  <h2>
                    Watching: {liveSession.agent_name}
                    <span className={`status status-${liveSession.status}`}>
                      {liveSession.status}
                    </span>
                  </h2>
                  <div className="session-stats">
                    <span>Session ID: {liveSession.id}</span>
                    <span>Started: {new Date(liveSession.started_at).toLocaleTimeString()}</span>
                    <span>Turns: {liveSession.turns}</span>
                    <span>Tokens: {liveSession.tokens_used}</span>
                    <span>Cost: ${liveSession.cost.toFixed(4)}</span>
                  </div>
                </div>

                <div className="live-display">
                  <div className="energy-display">
                    <div className="energy-bar-large">
                      <label>Agent Energy</label>
                      <div className="bar-container-large">
                        <div 
                          className="bar-fill-large"
                          style={{ 
                            width: `${(liveSession.energy / liveSession.max_energy) * 100}%`,
                            backgroundColor: getEnergyColor(liveSession.energy, liveSession.max_energy)
                          }}
                        />
                      </div>
                      <span>{liveSession.energy}/{liveSession.max_energy}</span>
                    </div>
                    
                    {liveSession.current_tool && (
                      <div className="current-action">
                        <div className="action-icon">
                          {getToolIcon(liveSession.current_tool)}
                        </div>
                        <div className="action-text">
                          <strong>Current Tool:</strong> {liveSession.current_tool}
                          {liveSession.current_action && (
                            <div className="sub-action">
                              {liveSession.current_action}
                            </div>
                          )}
                        </div>
                      </div>
                    )}
                  </div>

                  {liveSession.live_output && (
                    <div className="live-output">
                      <h3>Live Output:</h3>
                      <pre className="output-text">{liveSession.live_output}</pre>
                    </div>
                  )}
                </div>
              </div>
            )
          )}

          {recentToolCalls.length > 0 && (
            <div className="tool-call-feed">
              <h3>Recent Tool Calls</h3>
              <div className="tool-call-list">
                {recentToolCalls.map(call => (
                  <div key={call.id} className="tool-call-item">
                    <div className="tool-call-icon">
                      {getToolIcon(call.tool_name)}
                    </div>
                    <div className="tool-call-details">
                      <strong>{call.tool_name}</strong>
                      <span className="timestamp">
                        {new Date(call.timestamp).toLocaleTimeString()}
                      </span>
                      {call.duration && (
                        <span className="duration">
                          {call.duration.toFixed(2)}s
                        </span>
                      )}
                      <div className="parameters">
                        {Object.keys(call.parameters).slice(0, 2).map(key => (
                          <span key={key} className="param">
                            {key}: {String(call.parameters[key]).substring(0, 50)}
                            {String(call.parameters[key]).length > 50 ? '...' : ''}
                          </span>
                        ))}
                      </div>
                    </div>
                    <div className={`status-indicator ${call.success ? 'success' : 'pending'}`} />
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {!isWatching && (
        <div className="spectator-help">
          <h3>How to use Spectator Mode:</h3>
          <ul>
            <li>Select an agent or watch all active agents</li>
            <li>Click "Start Watching" to begin real-time monitoring</li>
            <li>See live tool calls, energy levels, and progress</li>
            <li>Adjust refresh rate based on your needs</li>
          </ul>
          
          <div className="active-agents-summary">
            <h4>Currently Active Agents:</h4>
            {activeAgents.length === 0 ? (
              <p>No agents are currently active</p>
            ) : (
              <div className="agent-list">
                {activeAgents.map(agent => (
                  <div key={agent.id} className="agent-item">
                    <span className="agent-avatar">{agent.avatar_emoji}</span>
                    <span className="agent-name">{agent.name}</span>
                    <span className={`status status-${agent.status}`}>
                      {agent.status}
                    </span>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default SpectatorMode;