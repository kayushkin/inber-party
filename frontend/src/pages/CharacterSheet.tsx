import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useStore, classColor, formatTokens, formatCost } from '../store';
import type { RPGAgent, RPGQuest, RPGAchievement, QuestHistoryEntry } from '../store';
import ChatPanel from '../components/ChatPanel';
import LevelUpAnimation from '../components/LevelUpAnimation';
import Tooltip from '../components/Tooltip';
import SkillTree from '../components/SkillTree';
import EquipmentComponent from '../components/Equipment';
import Reputation from '../components/Reputation';
import { STAT_TOOLTIPS, ACHIEVEMENT_TOOLTIPS, getSkillTooltip } from '../constants/tooltips';
import { getAgentEquipment, inferAvailableTools } from '../constants/equipment';
import { useAgentOperations } from '../services/agentService';
import './CharacterSheet.css';

// Mood helper functions
function getMoodColor(score: number): string {
  if (score <= 20) return '#ef4444'; // exhausted - red
  if (score <= 40) return '#f97316'; // stressed - orange  
  if (score <= 60) return '#eab308'; // neutral - yellow
  if (score <= 80) return '#22c55e'; // content - green
  return '#3b82f6'; // happy - blue
}

function getMoodEmoji(mood: string): string {
  switch (mood) {
    case 'exhausted': return '😫';
    case 'stressed': return '😰';
    case 'neutral': return '😐';
    case 'content': return '😊';
    case 'happy': return '😄';
    default: return '😐';
  }
}

function capitalize(str: string): string {
  return str.charAt(0).toUpperCase() + str.slice(1);
}

export default function CharacterSheet() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const agents = useStore((s) => s.agents);
  const [quests, setQuests] = useState<RPGQuest[]>([]);
  const [achievements, setAchievements] = useState<RPGAchievement[]>([]);
  const [questHistory, setQuestHistory] = useState<QuestHistoryEntry[]>([]);
  const [showAnimations, setShowAnimations] = useState(false);
  const [relationships, setRelationships] = useState<any[]>([]);
  const selectedAgent = useStore((s) => s.selectedAgent);
  const setSelectedAgent = useStore((s) => s.setSelectedAgent);
  const levelUpTriggers = useStore((s) => s.levelUpTriggers);

  const fetchAll = useStore((s) => s.fetchAll);
  const agent: RPGAgent | undefined = agents.find((a) => a.id === id);
  
  const {
    loadAgentQuests,
    loadAgentAchievements,
    loadAgentQuestHistory,
    error,
    hasError,
    canRetry,
    retry,
    isOnline
  } = useAgentOperations();

  useEffect(() => {
    if (agents.length === 0) fetchAll();
  }, [agents.length, fetchAll]);

  useEffect(() => {
    if (!id) return;
    
    const loadAgentData = async () => {
      // Load all agent-related data
      const [questsResult, achievementsResult, historyResult, relationshipsResult] = await Promise.all([
        loadAgentQuests(id, 100),
        loadAgentAchievements(id),
        loadAgentQuestHistory(id, 20),
        fetch(`/api/relationships?agent_id=${id}`).then(r => r.ok ? r.json() : []).catch(() => [])
      ]);

      // Update state with successful results
      if (questsResult?.data) setQuests(questsResult.data);
      if (achievementsResult?.data) setAchievements(achievementsResult.data);
      if (historyResult?.data) setQuestHistory(historyResult.data);
      if (relationshipsResult) setRelationships(relationshipsResult);
    };

    loadAgentData();
  }, [id, loadAgentQuests, loadAgentAchievements, loadAgentQuestHistory]);

  // Trigger animations when agent data is loaded
  useEffect(() => {
    if (agent) {
      // Reset and trigger animation after a short delay to ensure DOM is ready
      const resetTimer = setTimeout(() => setShowAnimations(false), 0);
      const animateTimer = setTimeout(() => setShowAnimations(true), 100);
      return () => {
        clearTimeout(resetTimer);
        clearTimeout(animateTimer);
      };
    }
  }, [agent]);

  if (!agent) {
    return (
      <div className="loading" style={{ color: '#d4af37', fontSize: '1.2rem', padding: '4rem', textAlign: 'center' }}>
        {agents.length === 0 ? '⏳ Loading adventurer...' : (
          <>
            <p>🗡️ Adventurer "{id}" not found in the guild roster.</p>
            <button onClick={() => navigate('/')} style={{ marginTop: '1rem', padding: '0.5rem 1rem', background: 'rgba(212,175,55,0.2)', border: '2px solid #d4af37', color: '#d4af37', borderRadius: '6px', cursor: 'pointer', fontFamily: 'Courier New, monospace' }}>← Back to Tavern</button>
          </>
        )}
      </div>
    );
  }

  // Show error state with retry option
  if (hasError && error) {
    return (
      <div className="loading" style={{ color: '#dc2626', fontSize: '1.2rem', padding: '4rem', textAlign: 'center' }}>
        <p>⚠️ Failed to load adventurer data</p>
        <p style={{ fontSize: '1rem', color: '#6b7280', marginBottom: '1rem' }}>
          {error.message}
        </p>
        {!isOnline && (
          <p style={{ fontSize: '0.9rem', color: '#f97316', marginBottom: '1rem' }}>
            📡 You appear to be offline
          </p>
        )}
        <div style={{ display: 'flex', gap: '1rem', justifyContent: 'center', flexWrap: 'wrap' }}>
          {canRetry && (
            <button
              onClick={retry}
              style={{
                padding: '0.5rem 1rem',
                background: '#059669',
                border: 'none',
                color: 'white',
                borderRadius: '6px',
                cursor: 'pointer',
                fontFamily: 'Courier New, monospace'
              }}
            >
              🔄 Retry
            </button>
          )}
          <button
            onClick={() => navigate('/')}
            style={{
              padding: '0.5rem 1rem',
              background: 'rgba(212,175,55,0.2)',
              border: '2px solid #d4af37',
              color: '#d4af37',
              borderRadius: '6px',
              cursor: 'pointer',
              fontFamily: 'Courier New, monospace'
            }}
          >
            ← Back to Tavern
          </button>
        </div>
      </div>
    );
  }

  const cc = classColor(agent.class);
  const xpPct = agent.xp_to_next > 0 ? (agent.xp / (agent.xp + agent.xp_to_next)) * 100 : 100;

  return (
    <div className="cs-layout">
      <div className="character-sheet">
        <button className="back-button" onClick={() => navigate('/')}>← Back to Tavern</button>

        {/* Header */}
        <div className="character-header" style={{ borderColor: cc }}>
          <div className="character-avatar">
            <img
              src={`/avatars/${agent.name.toLowerCase().replace(/\s+/g, '-')}.png`}
              alt={agent.name}
              onError={(e) => {
                e.currentTarget.style.display = 'none';
                const sibling = e.currentTarget.nextElementSibling as HTMLElement;
                if (sibling) sibling.style.display = 'block';
              }}
              style={{ width: '60px', height: '60px', borderRadius: '50%' }}
            />
            <span style={{ display: 'none' }}>{agent.avatar_emoji}</span>
          </div>
          <div className="character-title-section">
            <h1 className="character-name" style={{ color: cc }}>{agent.name}</h1>
            <h2 className="character-title">{agent.title}</h2>
            <div className="character-class" style={{ borderColor: cc, color: cc }}>
              {agent.class}
            </div>
          </div>
          <div className="character-actions">
            <button className="chat-with-btn" onClick={() => setSelectedAgent(agent.id)}>
              💬 Chat
            </button>
            <button 
              className="quarters-btn" 
              onClick={() => navigate(`/quarters/${id}`)}
              style={{ borderColor: cc, color: cc }}
            >
              🏠 Quarters
            </button>
          </div>
        </div>

        {/* Stats */}
        <div className="stats-grid">
          <div className="stat-card">
            <Tooltip content={STAT_TOOLTIPS.level}>
              <div className="stat-label">Level</div>
            </Tooltip>
            <div className="stat-value" style={{ color: cc }}>{agent.level}</div>
          </div>
          <div className="stat-card">
            <Tooltip content={STAT_TOOLTIPS.xp}>
              <div className="stat-label">XP</div>
            </Tooltip>
            <div className="stat-value">{agent.xp}</div>
            <div className="progress-bar"><div className="progress-fill animated" style={{ width: showAnimations ? `${xpPct}%` : '0%', background: cc }} /></div>
          </div>
          <div className="stat-card">
            <Tooltip content="Gold earned from completing quests. Used as currency in the RPG world.">
              <div className="stat-label">Gold</div>
            </Tooltip>
            <div className="stat-value" style={{ color: '#fbbf24' }}>🪙 {agent.gold || 0}</div>
          </div>
          <div className="stat-card">
            <Tooltip content={STAT_TOOLTIPS.energy}>
              <div className="stat-label">Energy</div>
            </Tooltip>
            <div className="stat-value">{agent.energy}%</div>
            <div className="progress-bar energy"><div className="progress-fill animated" style={{ width: showAnimations ? `${agent.energy}%` : '0%' }} /></div>
          </div>
          <div className="stat-card">
            <Tooltip content={`${STAT_TOOLTIPS.mood} Current score: ${agent.mood_score || 75}/100`}>
              <div className="stat-label">Mood</div>
            </Tooltip>
            <div className="stat-value" style={{ color: getMoodColor(agent.mood_score || 75) }}>
              {getMoodEmoji(agent.mood || 'neutral')} {capitalize(agent.mood || 'neutral')}
            </div>
            <div className="progress-bar mood">
              <div 
                className="progress-fill animated" 
                style={{ 
                  width: showAnimations ? `${agent.mood_score || 75}%` : '0%',
                  background: getMoodColor(agent.mood_score || 75)
                }} 
              />
            </div>
          </div>
          <div className="stat-card">
            <Tooltip content={STAT_TOOLTIPS.tokens}>
              <div className="stat-label">Tokens</div>
            </Tooltip>
            <div className="stat-value" style={{ color: '#60a5fa' }}>{formatTokens(agent.total_tokens)}</div>
          </div>
          <div className="stat-card">
            <Tooltip content={STAT_TOOLTIPS.cost}>
              <div className="stat-label">Cost</div>
            </Tooltip>
            <div className="stat-value" style={{ color: '#4ade80' }}>{formatCost(agent.total_cost)}</div>
          </div>
          <div className="stat-card">
            <Tooltip content={STAT_TOOLTIPS.sessions}>
              <div className="stat-label">Sessions</div>
            </Tooltip>
            <div className="stat-value">{agent.session_count}</div>
          </div>
        </div>

        {/* Quick Actions */}
        <div className="section">
          <h3>Quick Actions</h3>
          <div className="quick-actions-grid">
            <button 
              className="quick-action-btn" 
              onClick={() => {
                setSelectedAgent(agent.id);
                const sendMessage = useStore.getState().sendMessage;
                sendMessage(agent.id, "Scout this repo - analyze the codebase structure, recent changes, and identify any immediate issues or opportunities for improvement.");
              }}
              style={{ borderColor: cc }}
            >
              🔍 Scout Repo
            </button>
            <button 
              className="quick-action-btn" 
              onClick={() => {
                setSelectedAgent(agent.id);
                const sendMessage = useStore.getState().sendMessage;
                sendMessage(agent.id, "Fix bugs - scan the codebase for any obvious bugs, linting errors, or issues that need immediate attention and fix them.");
              }}
              style={{ borderColor: cc }}
            >
              🐛 Fix Bugs
            </button>
            <button 
              className="quick-action-btn" 
              onClick={() => {
                setSelectedAgent(agent.id);
                const sendMessage = useStore.getState().sendMessage;
                sendMessage(agent.id, "Write tests - identify areas of the codebase that need better test coverage and write comprehensive unit tests.");
              }}
              style={{ borderColor: cc }}
            >
              🧪 Write Tests
            </button>
            <button 
              className="quick-action-btn" 
              onClick={() => {
                setSelectedAgent(agent.id);
                const sendMessage = useStore.getState().sendMessage;
                sendMessage(agent.id, "Review documentation - check if README, docs, and code comments are up to date and comprehensive. Improve where needed.");
              }}
              style={{ borderColor: cc }}
            >
              📚 Review Docs
            </button>
            <button 
              className="quick-action-btn" 
              onClick={() => {
                setSelectedAgent(agent.id);
                const sendMessage = useStore.getState().sendMessage;
                sendMessage(agent.id, "Refactor code - identify areas where code can be simplified, optimized, or made more maintainable without changing functionality.");
              }}
              style={{ borderColor: cc }}
            >
              🔧 Refactor
            </button>
            <button 
              className="quick-action-btn" 
              onClick={() => {
                setSelectedAgent(agent.id);
                const sendMessage = useStore.getState().sendMessage;
                sendMessage(agent.id, "Security audit - scan for potential security vulnerabilities, unsafe practices, or areas that need security hardening.");
              }}
              style={{ borderColor: cc }}
            >
              🛡️ Security Check
            </button>
          </div>
        </div>

        {/* Achievements */}
        {achievements.length > 0 && (
          <div className="section">
            <h3>Achievements <span className="badge-count">{achievements.length}</span></h3>
            <div className="achievements-grid">
              {achievements.map((a) => {
                const achievementKey = a.name.toLowerCase().replace(/\s+/g, '_') as keyof typeof ACHIEVEMENT_TOOLTIPS;
                const tooltip = a.description || ACHIEVEMENT_TOOLTIPS[achievementKey] || 'Achievement unlocked!';
                return (
                  <Tooltip key={a.id} content={tooltip}>
                    <div className="achievement-badge">
                      <span className="badge-icon">{a.icon}</span>
                      <span className="badge-name">{a.name}</span>
                    </div>
                  </Tooltip>
                );
              })}
            </div>
          </div>
        )}

        {/* Skills */}
        {agent.skills.length > 0 && (
          <div className="section">
            <h3>Skills</h3>
            <div className="skills-list">
              {agent.skills.map((skill, i) => (
                <div key={i} className="skill-item">
                  <div className="skill-header">
                    <Tooltip content={getSkillTooltip(skill.skill_name)}>
                      <span className="skill-name">{skill.skill_name}</span>
                    </Tooltip>
                    <span className="skill-level">Lv {skill.level}</span>
                  </div>
                  <div className="skill-bar">
                    <div className="skill-bar-fill animated" style={{ width: showAnimations ? `${Math.min(skill.level * 10, 100)}%` : '0%', background: cc }} />
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Reputation */}
        {agent.reputation && (
          <div className="section">
            <Reputation reputation={agent.reputation} />
          </div>
        )}

        {/* Skill Tree */}
        <div className="section">
          <h3>Skill Tree</h3>
          <SkillTree 
            agentClass={agent.class}
            agentLevel={agent.level}
            agentSkills={agent.skills}
          />
        </div>

        {/* Equipment */}
        <div className="section">
          <h3>Equipment & Gear</h3>
          <EquipmentComponent 
            equipment={getAgentEquipment(agent, inferAvailableTools(agent))}
            heldItems={agent.held_items}
          />
        </div>

        {/* Quest History */}
        {quests.length > 0 && (
          <div className="section">
            <h3>Quest Log ({quests.length})</h3>
            <div className="quest-list">
              {quests.slice(0, 20).map((q) => (
                <div key={q.id} className={`quest-item status-${q.status}`}>
                  <div className="quest-header">
                    <span className="quest-name">{q.name}</span>
                    <span className="quest-status">{q.status}</span>
                  </div>
                  <div className="quest-meta">
                    <span>🔮 {formatTokens(q.tokens_used)}</span>
                    <span>↻ {q.turns}</span>
                    <span>+{q.xp_reward} XP</span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Enhanced Analytics Dashboard */}
        {questHistory.length > 2 && (
          <div className="section">
            <h3>📊 Analytics Dashboard</h3>
            
            {/* Performance Metrics */}
            <div className="analytics-metrics">
              {(() => {
                const completedQuests = questHistory.filter(q => q.status === 'completed').length;
                const successRate = questHistory.length > 0 ? ((completedQuests / questHistory.length) * 100) : 0;
                const avgTokens = questHistory.reduce((sum, q) => sum + q.tokens, 0) / questHistory.length;
                const avgCost = questHistory.reduce((sum, q) => sum + q.cost, 0) / questHistory.length;
                
                return (
                  <div className="metrics-grid" style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(120px, 1fr))', gap: '1rem', marginBottom: '1.5rem' }}>
                    <div className="metric-card" style={{ padding: '1rem', background: 'rgba(212,175,55,0.1)', borderRadius: '8px', textAlign: 'center' }}>
                      <div style={{ color: '#22c55e', fontSize: '1.5rem', fontWeight: 'bold' }}>{successRate.toFixed(1)}%</div>
                      <div style={{ fontSize: '0.8rem', opacity: 0.8 }}>Success Rate</div>
                    </div>
                    <div className="metric-card" style={{ padding: '1rem', background: 'rgba(212,175,55,0.1)', borderRadius: '8px', textAlign: 'center' }}>
                      <div style={{ color: '#60a5fa', fontSize: '1.5rem', fontWeight: 'bold' }}>{Math.round(avgTokens)}</div>
                      <div style={{ fontSize: '0.8rem', opacity: 0.8 }}>Avg Tokens/Quest</div>
                    </div>
                    <div className="metric-card" style={{ padding: '1rem', background: 'rgba(212,175,55,0.1)', borderRadius: '8px', textAlign: 'center' }}>
                      <div style={{ color: '#4ade80', fontSize: '1.5rem', fontWeight: 'bold' }}>${avgCost.toFixed(3)}</div>
                      <div style={{ fontSize: '0.8rem', opacity: 0.8 }}>Avg Cost/Quest</div>
                    </div>
                    <div className="metric-card" style={{ padding: '1rem', background: 'rgba(212,175,55,0.1)', borderRadius: '8px', textAlign: 'center' }}>
                      <div style={{ color: cc, fontSize: '1.5rem', fontWeight: 'bold' }}>{questHistory.length}</div>
                      <div style={{ fontSize: '0.8rem', opacity: 0.8 }}>Total Quests</div>
                    </div>
                  </div>
                );
              })()}
            </div>

            {/* Token Usage Trend */}
            <div className="chart-container" style={{ marginBottom: '1.5rem' }}>
              <h4 style={{ margin: '0 0 0.5rem 0', fontSize: '1rem', color: cc }}>🔮 Token Usage Trend</h4>
              <svg viewBox={`0 0 200 40`} width="100%" height="60" preserveAspectRatio="none" style={{ display: 'block' }}>
                {(() => {
                  const data = questHistory.map((q) => q.tokens);
                  const max = Math.max(...data, 1);
                  const points = data.map((v, i) => `${(i / (data.length - 1)) * 200},${40 - (v / max) * 36 - 2}`).join(' ');
                  return <>
                    <polygon points={`0,40 ${points} 200,40`} fill={cc} opacity="0.15" />
                    <polyline points={points} fill="none" stroke={cc} strokeWidth="2" />
                  </>;
                })()}
              </svg>
            </div>

            {/* Cost Trend */}
            <div className="chart-container" style={{ marginBottom: '1.5rem' }}>
              <h4 style={{ margin: '0 0 0.5rem 0', fontSize: '1rem', color: '#4ade80' }}>💰 Cost Trend</h4>
              <svg viewBox={`0 0 200 40`} width="100%" height="60" preserveAspectRatio="none" style={{ display: 'block' }}>
                {(() => {
                  const data = questHistory.map((q) => q.cost);
                  const max = Math.max(...data, 0.001);
                  const points = data.map((v, i) => `${(i / (data.length - 1)) * 200},${40 - (v / max) * 36 - 2}`).join(' ');
                  return <>
                    <polygon points={`0,40 ${points} 200,40`} fill="#4ade80" opacity="0.15" />
                    <polyline points={points} fill="none" stroke="#4ade80" strokeWidth="2" />
                  </>;
                })()}
              </svg>
            </div>

            {/* Quest Status Distribution */}
            <div className="status-distribution" style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap' }}>
              {(() => {
                const statusCounts = questHistory.reduce((acc, q) => {
                  acc[q.status] = (acc[q.status] || 0) + 1;
                  return acc;
                }, {} as Record<string, number>);
                
                const statusColors = {
                  completed: '#22c55e',
                  failed: '#ef4444',
                  in_progress: '#f59e0b'
                };

                return Object.entries(statusCounts).map(([status, count]) => (
                  <div key={status} style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                    <div 
                      style={{ 
                        width: '12px', 
                        height: '12px', 
                        borderRadius: '50%', 
                        background: statusColors[status as keyof typeof statusColors] || '#6b7280' 
                      }}
                    />
                    <span style={{ fontSize: '0.9rem' }}>
                      {status}: {count}
                    </span>
                  </div>
                ));
              })()}
            </div>
          </div>
        )}

        {/* Relationships Section */}
        {relationships.length > 0 && (
          <div className="section">
            <h3>🤝 Relationships</h3>
            <div className="relationships-grid" style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(250px, 1fr))', gap: '1rem' }}>
              {relationships.map((rel: any) => {
                const otherAgent = rel.agent1.id === agent.id ? rel.agent2 : rel.agent1;
                const getRelationshipIcon = (type: string) => {
                  switch (type) {
                    case 'friendship': return '👥';
                    case 'rivalry': return '⚔️';
                    case 'neutral': return '🤝';
                    default: return '❓';
                  }
                };
                const getStrengthColor = (strength: number) => {
                  if (strength >= 80) return '#00ff88';
                  if (strength >= 60) return '#88ff00';
                  if (strength >= 40) return '#ffff00';
                  if (strength >= 20) return '#ff8800';
                  return '#ff4444';
                };
                
                return (
                  <div 
                    key={rel.id} 
                    className="relationship-card" 
                    style={{ 
                      background: 'rgba(212,175,55,0.1)', 
                      border: `2px solid ${rel.relationship_type === 'friendship' ? '#66bb6a' : rel.relationship_type === 'rivalry' ? '#ef5350' : '#78909c'}`,
                      borderRadius: '8px', 
                      padding: '1rem',
                      cursor: 'pointer'
                    }}
                    onClick={() => navigate(`/agent/${otherAgent.id}`)}
                  >
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.5rem' }}>
                      <span style={{ fontSize: '1.5rem' }}>{getRelationshipIcon(rel.relationship_type)}</span>
                      <span style={{ fontSize: '1.5rem' }}>{otherAgent.avatar_emoji}</span>
                      <div style={{ flex: 1 }}>
                        <div style={{ fontWeight: '600', color: cc }}>{otherAgent.name}</div>
                        <div style={{ fontSize: '0.8rem', opacity: 0.8 }}>{otherAgent.title}</div>
                      </div>
                      <span style={{ color: getStrengthColor(rel.strength), fontWeight: '700' }}>
                        {rel.strength}%
                      </span>
                    </div>
                    <div style={{ fontSize: '0.9rem', color: cc, textTransform: 'capitalize', marginBottom: '0.5rem' }}>
                      {rel.relationship_type}
                    </div>
                    <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.8rem', opacity: 0.8 }}>
                      <span>🤝 {rel.collaboration_count}</span>
                      <span>✅ {rel.successful_collabs}</span>
                      <span>⚔️ {rel.competition_count}</span>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        )}
      </div>

      {selectedAgent === agent.id && (
        <ChatPanel agentId={agent.id} agent={agent} onClose={() => setSelectedAgent(null)} />
      )}
      
      {/* Level-up animation */}
      {agent && (
        <LevelUpAnimation
          trigger={levelUpTriggers[agent.id] || 0}
          agentName={agent.name}
          level={agent.level}
          color={cc}
        />
      )}
    </div>
  );
}
