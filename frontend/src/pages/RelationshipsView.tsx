import { useState, useEffect } from 'react';
import { useStore } from '../store';
import LoadingSpinner from '../components/LoadingSpinner';
import ErrorDisplay from '../components/ErrorDisplay';
import './RelationshipsView.css';

interface AgentRelationship {
  id: number;
  agent1_id: number;
  agent2_id: number;
  relationship_type: 'friendship' | 'rivalry' | 'neutral';
  strength: number;
  collaboration_count: number;
  successful_collabs: number;
  competition_count: number;
  last_interaction?: string;
  agent1: Agent;
  agent2: Agent;
}

interface Agent {
  id: number;
  name: string;
  title: string;
  class: string;
  level: number;
  avatar_emoji: string;
}

interface RelationshipStats {
  agent_id: number;
  total_friends: number;
  total_rivals: number;
  best_friend?: AgentRelationship;
  biggest_rival?: AgentRelationship;
  relationships: AgentRelationship[];
}

function RelationshipsView() {
  const agents = useStore((s) => s.agents);
  const [relationships, setRelationships] = useState<AgentRelationship[]>([]);
  const [selectedAgent, setSelectedAgent] = useState<number | null>(null);
  const [selectedAgentStats, setSelectedAgentStats] = useState<RelationshipStats | null>(null);
  const [loading, setLoading] = useState(false);
  const [analyzing, setAnalyzing] = useState(false);
  const [error, setError] = useState<string>('');

  // Load all relationships on component mount
  useEffect(() => {
    loadRelationships();
  }, []);

  // Load agent-specific stats when an agent is selected
  useEffect(() => {
    if (selectedAgent) {
      loadAgentStats(selectedAgent);
    } else {
      setSelectedAgentStats(null);
    }
  }, [selectedAgent]);

  const loadRelationships = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/relationships');
      if (response.ok) {
        const data = await response.json();
        setRelationships(data || []);
      } else {
        throw new Error('Failed to load relationships');
      }
    } catch (err) {
      console.error('Error loading relationships:', err);
      setError('Failed to load relationships');
    } finally {
      setLoading(false);
    }
  };

  const loadAgentStats = async (agentId: number) => {
    try {
      const response = await fetch(`/api/relationships/stats/${agentId}`);
      if (response.ok) {
        const data = await response.json();
        setSelectedAgentStats(data);
      } else {
        throw new Error('Failed to load agent stats');
      }
    } catch (err) {
      console.error('Error loading agent stats:', err);
      setError('Failed to load agent stats');
    }
  };

  const analyzeRelationships = async () => {
    try {
      setAnalyzing(true);
      const response = await fetch('/api/relationships/analyze', { method: 'POST' });
      if (response.ok) {
        // Reload relationships after analysis
        setTimeout(() => {
          loadRelationships();
          if (selectedAgent) {
            loadAgentStats(selectedAgent);
          }
        }, 1000);
      } else {
        throw new Error('Failed to analyze relationships');
      }
    } catch (err) {
      console.error('Error analyzing relationships:', err);
      setError('Failed to analyze relationships');
    } finally {
      setAnalyzing(false);
    }
  };

  const getRelationshipIcon = (type: string) => {
    switch (type) {
      case 'friendship':
        return '👥';
      case 'rivalry':
        return '⚔️';
      case 'neutral':
        return '🤝';
      default:
        return '❓';
    }
  };

  const getStrengthColor = (strength: number) => {
    if (strength >= 80) return '#00ff88';
    if (strength >= 60) return '#88ff00';
    if (strength >= 40) return '#ffff00';
    if (strength >= 20) return '#ff8800';
    return '#ff4444';
  };

  const getRelationshipDescription = (rel: AgentRelationship) => {
    const { relationship_type, strength } = rel;
    
    if (relationship_type === 'friendship') {
      if (strength >= 80) return 'Inseparable allies';
      if (strength >= 60) return 'Close friends';
      if (strength >= 40) return 'Good friends';
      return 'Friendly acquaintances';
    } else if (relationship_type === 'rivalry') {
      if (strength >= 80) return 'Bitter enemies';
      if (strength >= 60) return 'Strong rivals';
      if (strength >= 40) return 'Competitors';
      return 'Mild tension';
    } else {
      return 'Neutral colleagues';
    }
  };

  if (loading) {
    return (
      <div className="relationships-view">
        <LoadingSpinner />
      </div>
    );
  }

  return (
    <div className="relationships-view">
      <div className="rv-header">
        <div className="rv-title-section">
          <h1>🤝 Agent Relationships</h1>
          <p>Explore the bonds, rivalries, and collaborations between agents</p>
        </div>
        <div className="rv-controls">
          <button 
            className={`rv-analyze-btn ${analyzing ? 'analyzing' : ''}`}
            onClick={analyzeRelationships}
            disabled={analyzing}
          >
            {analyzing ? '🔍 Analyzing...' : '🔍 Analyze Relationships'}
          </button>
        </div>
      </div>

      {error && <ErrorDisplay error={error} onRetry={() => setError('')} />}

      <div className="rv-content">
        {/* Agent Selector */}
        <div className="rv-agent-selector">
          <h3>Select an Agent</h3>
          <div className="rv-agent-grid">
            <div 
              className={`rv-agent-card ${selectedAgent === null ? 'active' : ''}`}
              onClick={() => setSelectedAgent(null)}
            >
              <div className="rvac-emoji">🌐</div>
              <div className="rvac-name">All Relationships</div>
            </div>
            {agents.map((agent: any) => (
              <div 
                key={agent.id}
                className={`rv-agent-card ${selectedAgent === agent.id ? 'active' : ''}`}
                onClick={() => setSelectedAgent(agent.id)}
              >
                <div className="rvac-emoji">{agent.avatar_emoji}</div>
                <div className="rvac-name">{agent.name}</div>
                <div className="rvac-level">Lv.{agent.level}</div>
              </div>
            ))}
          </div>
        </div>

        {/* Agent-Specific Stats */}
        {selectedAgent && selectedAgentStats && (
          <div className="rv-agent-stats">
            <h3>Relationship Summary</h3>
            <div className="rvas-summary">
              <div className="rvas-stat">
                <span className="rvas-emoji">👥</span>
                <span className="rvas-count">{selectedAgentStats.total_friends}</span>
                <span className="rvas-label">Friends</span>
              </div>
              <div className="rvas-stat">
                <span className="rvas-emoji">⚔️</span>
                <span className="rvas-count">{selectedAgentStats.total_rivals}</span>
                <span className="rvas-label">Rivals</span>
              </div>
              <div className="rvas-stat">
                <span className="rvas-emoji">🤝</span>
                <span className="rvas-count">{selectedAgentStats.relationships.length}</span>
                <span className="rvas-label">Total</span>
              </div>
            </div>

            {selectedAgentStats.best_friend && (
              <div className="rvas-highlight">
                <h4>👥 Best Friend</h4>
                <div className="rvas-highlight-card">
                  <span className="rvas-avatar">{selectedAgentStats.best_friend.agent2.avatar_emoji}</span>
                  <span className="rvas-name">{selectedAgentStats.best_friend.agent2.name}</span>
                  <span className="rvas-strength" style={{ color: getStrengthColor(selectedAgentStats.best_friend.strength) }}>
                    {selectedAgentStats.best_friend.strength}%
                  </span>
                </div>
              </div>
            )}

            {selectedAgentStats.biggest_rival && (
              <div className="rvas-highlight">
                <h4>⚔️ Biggest Rival</h4>
                <div className="rvas-highlight-card">
                  <span className="rvas-avatar">{selectedAgentStats.biggest_rival.agent2.avatar_emoji}</span>
                  <span className="rvas-name">{selectedAgentStats.biggest_rival.agent2.name}</span>
                  <span className="rvas-strength" style={{ color: getStrengthColor(selectedAgentStats.biggest_rival.strength) }}>
                    {selectedAgentStats.biggest_rival.strength}%
                  </span>
                </div>
              </div>
            )}
          </div>
        )}

        {/* Relationships List */}
        <div className="rv-relationships">
          <h3>
            {selectedAgent 
              ? `Relationships for ${agents.find((a: any) => a.id === selectedAgent)?.name}` 
              : 'All Relationships'
            }
          </h3>
          
          {relationships.length === 0 ? (
            <div className="rv-empty">
              <div className="rve-icon">🤷</div>
              <div className="rve-message">
                No relationships found. Try analyzing relationships first!
              </div>
            </div>
          ) : (
            <div className="rv-relationships-grid">
              {(selectedAgent && selectedAgentStats 
                ? selectedAgentStats.relationships 
                : relationships
              ).map(relationship => (
                <div key={relationship.id} className={`rv-relationship-card rrc-${relationship.relationship_type}`}>
                  <div className="rrc-header">
                    <span className="rrc-icon">{getRelationshipIcon(relationship.relationship_type)}</span>
                    <span className="rrc-type">{relationship.relationship_type}</span>
                    <span 
                      className="rrc-strength" 
                      style={{ color: getStrengthColor(relationship.strength) }}
                    >
                      {relationship.strength}%
                    </span>
                  </div>
                  
                  <div className="rrc-agents">
                    <div className="rrc-agent">
                      <span className="rrc-avatar">{relationship.agent1.avatar_emoji}</span>
                      <div className="rrc-info">
                        <div className="rrc-name">{relationship.agent1.name}</div>
                        <div className="rrc-title">{relationship.agent1.title}</div>
                      </div>
                    </div>
                    
                    <div className="rrc-connection">
                      {relationship.relationship_type === 'friendship' ? '💖' : 
                       relationship.relationship_type === 'rivalry' ? '💥' : '🤝'}
                    </div>
                    
                    <div className="rrc-agent">
                      <span className="rrc-avatar">{relationship.agent2.avatar_emoji}</span>
                      <div className="rrc-info">
                        <div className="rrc-name">{relationship.agent2.name}</div>
                        <div className="rrc-title">{relationship.agent2.title}</div>
                      </div>
                    </div>
                  </div>
                  
                  <div className="rrc-description">
                    {getRelationshipDescription(relationship)}
                  </div>
                  
                  <div className="rrc-stats">
                    <div className="rrc-stat">
                      <span className="rrc-stat-icon">🤝</span>
                      <span className="rrc-stat-value">{relationship.collaboration_count}</span>
                      <span className="rrc-stat-label">collabs</span>
                    </div>
                    <div className="rrc-stat">
                      <span className="rrc-stat-icon">✅</span>
                      <span className="rrc-stat-value">{relationship.successful_collabs}</span>
                      <span className="rrc-stat-label">success</span>
                    </div>
                    <div className="rrc-stat">
                      <span className="rrc-stat-icon">⚔️</span>
                      <span className="rrc-stat-value">{relationship.competition_count}</span>
                      <span className="rrc-stat-label">compete</span>
                    </div>
                  </div>
                  
                  {relationship.last_interaction && (
                    <div className="rrc-last-interaction">
                      Last interaction: {new Date(relationship.last_interaction).toLocaleDateString()}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default RelationshipsView;