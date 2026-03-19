import { useState, useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useStore, classColor, formatTokens, formatCost, timeAgo } from '../store';
import type { RPGAgent } from '../store';
import './ComparisonView.css';

export default function ComparisonView() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const agents = useStore((s) => s.agents);
  const fetchAll = useStore((s) => s.fetchAll);
  // Parse agent IDs from URL params
  const [selectedAgentIds, setSelectedAgentIds] = useState<string[]>(() => {
    return searchParams.getAll('agent');
  });

  // Update selected agents when URL params change
  useEffect(() => {
    const agentIds = searchParams.getAll('agent');
    // Use a ref to track if we need to update to avoid setState in effect
    const shouldUpdate = selectedAgentIds.join(',') !== agentIds.join(',');
    if (shouldUpdate) {
      const timer = setTimeout(() => {
        setSelectedAgentIds(agentIds);
      }, 0);
      return () => clearTimeout(timer);
    }
  }, [searchParams, selectedAgentIds]);

  useEffect(() => {
    if (agents.length === 0) fetchAll();
  }, [agents.length, fetchAll]);

  const selectedAgents = selectedAgentIds.map(id => 
    agents.find(agent => agent.id === id)
  ).filter(Boolean) as RPGAgent[];

  const handleAgentSelection = (agentId: string, selected: boolean) => {
    const newIds = selected 
      ? [...selectedAgentIds, agentId]
      : selectedAgentIds.filter(id => id !== agentId);
    
    // Limit to 3 agents
    if (newIds.length <= 3) {
      setSelectedAgentIds(newIds);
      const newParams = new URLSearchParams();
      newIds.forEach(id => newParams.append('agent', id));
      setSearchParams(newParams);
    }
  };

  const clearSelection = () => {
    setSelectedAgentIds([]);
    setSearchParams({});
  };

  return (
    <div className="comparison-layout">
      <div className="comparison-header">
        <button className="back-button" onClick={() => navigate('/')}>
          ← Back to Tavern
        </button>
        <h1>⚔️ Agent Comparison</h1>
        <button 
          className="clear-button" 
          onClick={clearSelection}
          disabled={selectedAgentIds.length === 0}
        >
          Clear Selection
        </button>
      </div>

      {/* Agent Selection Grid */}
      {selectedAgents.length < 3 && (
        <div className="agent-selection">
          <h2>Select agents to compare (max 3)</h2>
          <div className="selection-grid">
            {agents.map((agent) => {
              const isSelected = selectedAgentIds.includes(agent.id);
              const isDisabled = !isSelected && selectedAgentIds.length >= 3;
              
              return (
                <label 
                  key={agent.id} 
                  className={`agent-selector ${isSelected ? 'selected' : ''} ${isDisabled ? 'disabled' : ''}`}
                >
                  <input
                    type="checkbox"
                    checked={isSelected}
                    onChange={(e) => handleAgentSelection(agent.id, e.target.checked)}
                    disabled={isDisabled}
                  />
                  <div className="selector-content">
                    <div className="selector-avatar">
                      <img
                        src={`/avatars/${agent.name.toLowerCase().replace(/\s+/g, '-')}.png`}
                        alt={agent.name}
                        onError={(e) => {
                          e.currentTarget.style.display = 'none';
                          const sibling = e.currentTarget.nextElementSibling as HTMLElement;
                          if (sibling) sibling.style.display = 'block';
                        }}
                      />
                      <span style={{ display: 'none' }}>{agent.avatar_emoji}</span>
                    </div>
                    <div className="selector-info">
                      <div className="selector-name" style={{ color: classColor(agent.class) }}>
                        {agent.name}
                      </div>
                      <div className="selector-meta">
                        {agent.class} · Lv {agent.level}
                      </div>
                    </div>
                  </div>
                </label>
              );
            })}
          </div>
        </div>
      )}

      {/* Comparison Cards */}
      {selectedAgents.length > 1 && (
        <div className="comparison-section">
          <h2>Comparison</h2>
          <div className="comparison-cards">
            {selectedAgents.map((agent) => (
              <ComparisonCard key={agent.id} agent={agent} />
            ))}
          </div>
        </div>
      )}

      {selectedAgents.length === 1 && (
        <div className="single-agent-notice">
          <p>Select at least one more agent to compare</p>
        </div>
      )}
    </div>
  );
}

interface ComparisonCardProps {
  agent: RPGAgent;
}

function ComparisonCard({ agent }: ComparisonCardProps) {
  const navigate = useNavigate();
  const cc = classColor(agent.class);
  const xpPct = agent.xp_to_next > 0 
    ? (agent.xp / (agent.xp + agent.xp_to_next)) * 100 
    : 100;

  return (
    <div className="comparison-card" style={{ '--cc': cc } as React.CSSProperties}>
      {/* Header */}
      <div className="comp-header" style={{ borderColor: cc }}>
        <div className="comp-avatar">
          <img
            src={`/avatars/${agent.name.toLowerCase().replace(/\s+/g, '-')}.png`}
            alt={agent.name}
            onError={(e) => {
              e.currentTarget.style.display = 'none';
              const sibling = e.currentTarget.nextElementSibling as HTMLElement;
              if (sibling) sibling.style.display = 'block';
            }}
          />
          <span style={{ display: 'none' }}>{agent.avatar_emoji}</span>
        </div>
        <div className="comp-title">
          <h3 className="comp-name" style={{ color: cc }}>{agent.name}</h3>
          <div className="comp-class" style={{ color: cc }}>{agent.class}</div>
        </div>
      </div>

      {/* Core Stats */}
      <div className="comp-stats">
        <div className="comp-stat">
          <div className="stat-label">Level</div>
          <div className="stat-value" style={{ color: cc }}>{agent.level}</div>
        </div>
        <div className="comp-stat">
          <div className="stat-label">XP</div>
          <div className="stat-value">{agent.xp}</div>
          <div className="progress-bar">
            <div className="progress-fill" style={{ width: `${xpPct}%`, background: cc }} />
          </div>
        </div>
        <div className="comp-stat">
          <div className="stat-label">Energy</div>
          <div className="stat-value">{agent.energy}%</div>
          <div className="progress-bar energy">
            <div className="progress-fill" style={{ width: `${agent.energy}%` }} />
          </div>
        </div>
        <div className="comp-stat">
          <div className="stat-label">Tokens</div>
          <div className="stat-value">{formatTokens(agent.total_tokens)}</div>
        </div>
        <div className="comp-stat">
          <div className="stat-label">Cost</div>
          <div className="stat-value">{formatCost(agent.total_cost)}</div>
        </div>
        <div className="comp-stat">
          <div className="stat-label">Sessions</div>
          <div className="stat-value">{agent.session_count}</div>
        </div>
      </div>

      {/* Top Skills */}
      {agent.skills.length > 0 && (
        <div className="comp-skills">
          <div className="skills-header">Top Skills</div>
          {agent.skills.slice(0, 3).map((skill, i) => (
            <div key={i} className="skill-item">
              <span className="skill-name">{skill.skill_name}</span>
              <span className="skill-level">Lv {skill.level}</span>
            </div>
          ))}
        </div>
      )}

      {/* Last Active */}
      <div className="comp-footer">
        <div className="last-active">
          Last active: {timeAgo(agent.last_active)}
        </div>
        <button 
          className="view-details-btn"
          onClick={() => navigate(`/agent/${agent.id}`)}
        >
          View Details
        </button>
      </div>
    </div>
  );
}