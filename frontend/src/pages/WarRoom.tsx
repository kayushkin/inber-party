import { useStore, formatTokens, formatCost, timeAgo, getDifficultyStars, getDifficultyName, calculateQuestDifficulty } from '../store';
import { SkeletonQuestCard } from '../components/SkeletonLoader';
import BossBattle from '../components/BossBattle';
import Tooltip from '../components/Tooltip';
import './WarRoom.css';

export default function WarRoom() {
  const quests = useStore((s) => s.quests);
  const agents = useStore((s) => s.agents);
  const isLoadingQuests = useStore((s) => s.isLoadingQuests);
  const isLoadingAgents = useStore((s) => s.isLoadingAgents);

  const activeQuests = quests.filter((q) => q.status === 'in_progress');
  const workingAgents = agents.filter((a) => a.status === 'working');

  // Identify boss-level quests (high difficulty OR multi-agent OR high token count)
  const bossQuests = activeQuests.filter((q) => 
    q.tokens_used >= 50000 || // High token count
    (q.children && q.children > 0) || // Multi-agent task
    calculateQuestDifficulty(q.tokens_used) >= 4 // 4-5 star difficulty
  );
  
  const regularQuests = activeQuests.filter((q) => !bossQuests.includes(q));

  // Calculate war room stats
  const totalActiveTokens = activeQuests.reduce((sum, q) => sum + q.tokens_used, 0);
  const totalActiveCost = activeQuests.reduce((sum, q) => sum + q.cost, 0);
  const averageProgress = activeQuests.length > 0 
    ? activeQuests.reduce((sum, q) => sum + q.progress, 0) / activeQuests.length 
    : 0;

  return (
    <div className="war-room">
      <div className="war-room-header">
        <h1>⚔️ War Room</h1>
        <div className="war-room-subtitle">Active Operations Command Center</div>
      </div>

      {/* Strategic Overview */}
      <div className="strategic-overview">
        <div className="overview-stats">
          <div className="stat-card">
            <div className="stat-icon">⚡</div>
            <div className="stat-content">
              <span className="stat-number">{activeQuests.length}</span>
              <span className="stat-label">Active Quests</span>
            </div>
          </div>
          <div className="stat-card">
            <div className="stat-icon">🏃</div>
            <div className="stat-content">
              <span className="stat-number">{workingAgents.length}</span>
              <span className="stat-label">Agents Deployed</span>
            </div>
          </div>
          <div className="stat-card">
            <div className="stat-icon">🎯</div>
            <div className="stat-content">
              <span className="stat-number">{averageProgress.toFixed(0)}%</span>
              <span className="stat-label">Avg Progress</span>
            </div>
          </div>
          <div className="stat-card">
            <div className="stat-icon">🪙</div>
            <div className="stat-content">
              <span className="stat-number">{formatTokens(totalActiveTokens)}</span>
              <span className="stat-label">Total Tokens</span>
              <div className="stat-subtitle">{formatCost(totalActiveCost)}</div>
            </div>
          </div>
        </div>
      </div>

      {/* Agent Status Board */}
      <div className="agent-status-board">
        <h2>👥 Agent Status</h2>
        <div className="agents-grid">
          {isLoadingAgents ? (
            [...Array(3)].map((_, i) => (
              <div key={`skeleton-agent-${i}`} className="agent-status-card skeleton">
                <div className="skeleton-bar"></div>
              </div>
            ))
          ) : (
            agents.slice(0, 6).map((agent) => {
              const agentActiveQuests = activeQuests.filter(q => q.assigned_agent_id === agent.id);
              return (
                <Tooltip 
                  key={agent.id} 
                  content={`${agent.name}: ${agent.status.charAt(0).toUpperCase() + agent.status.slice(1)} • Level ${agent.level} • Energy ${agent.energy}/${agent.max_energy}`}
                >
                  <div className={`agent-status-card status-${agent.status}`}>
                    <div className="agent-avatar">{agent.avatar_emoji}</div>
                    <div className="agent-info">
                      <div className="agent-name">{agent.name}</div>
                      <div className="agent-details">
                        <span className="agent-level">Lv.{agent.level}</span>
                        <span className="agent-quests">{agentActiveQuests.length} quest{agentActiveQuests.length !== 1 ? 's' : ''}</span>
                      </div>
                    </div>
                    <div className="agent-energy-bar">
                      <div 
                        className="agent-energy-fill" 
                        style={{ width: `${(agent.energy / agent.max_energy) * 100}%` }}
                      />
                    </div>
                  </div>
                </Tooltip>
              );
            })
          )}
        </div>
      </div>

      {/* Boss Battles */}
      {bossQuests.length > 0 && (
        <div className="boss-battles-section">
          <h2>🐉 Boss Battles</h2>
          <div className="boss-battles-container">
            {bossQuests.map((quest) => {
              const assignedAgent = agents.find(a => a.id === quest.assigned_agent_id);
              return (
                <BossBattle 
                  key={quest.id} 
                  quest={quest} 
                  assignedAgent={assignedAgent}
                  isActive={true}
                />
              );
            })}
          </div>
        </div>
      )}

      {/* Active Operations */}
      <div className="active-operations">
        <h2>🎯 Active Operations</h2>
        {isLoadingQuests ? (
          <div className="operations-grid">
            {[...Array(4)].map((_, i) => (
              <SkeletonQuestCard key={`skeleton-operation-${i}`} />
            ))}
          </div>
        ) : regularQuests.length === 0 ? (
          <div className="no-operations">
            <div className="no-operations-icon">😴</div>
            <div className="no-operations-text">All quiet on the digital front</div>
            <div className="no-operations-subtitle">No regular operations at the moment{bossQuests.length > 0 ? ' (boss battles active above)' : ''}</div>
          </div>
        ) : (
          <div className="operations-grid">
            {regularQuests.map((quest) => {
              const assignedAgent = agents.find(a => a.id === quest.assigned_agent_id);
              const progressClass = quest.progress > 75 ? 'high' : quest.progress > 25 ? 'medium' : 'low';
              
              return (
                <div key={quest.id} className={`operation-card progress-${progressClass}`}>
                  <div className="operation-header">
                    <h3 className="operation-name">{quest.name}</h3>
                    <span 
                      className="operation-difficulty" 
                      title={`${getDifficultyName(quest.tokens_used)} (${calculateQuestDifficulty(quest.tokens_used)}/5 stars)`}
                    >
                      {getDifficultyStars(quest.tokens_used)}
                    </span>
                  </div>
                  
                  {quest.description && (
                    <p className="operation-description">
                      {quest.description.slice(0, 100)}{quest.description.length > 100 ? '...' : ''}
                    </p>
                  )}

                  <div className="operation-agent">
                    {assignedAgent && (
                      <>
                        <span className="agent-avatar">{assignedAgent.avatar_emoji}</span>
                        <span className="agent-name">{assignedAgent.name}</span>
                        <span className={`agent-status status-${assignedAgent.status}`}>
                          {assignedAgent.status.charAt(0).toUpperCase() + assignedAgent.status.slice(1)}
                        </span>
                      </>
                    )}
                  </div>

                  <div className="operation-progress">
                    <div className="progress-header">
                      <span className="progress-label">Progress</span>
                      <span className="progress-percentage">{quest.progress}%</span>
                    </div>
                    <div className="progress-bar">
                      <div 
                        className={`progress-fill progress-${progressClass}`}
                        style={{ width: `${quest.progress}%` }}
                      />
                    </div>
                  </div>

                  <div className="operation-metrics">
                    <div className="metric">
                      <span className="metric-label">Turns</span>
                      <span className="metric-value">{quest.turns}</span>
                    </div>
                    <div className="metric">
                      <span className="metric-label">Tokens</span>
                      <span className="metric-value">{formatTokens(quest.tokens_used)}</span>
                    </div>
                    <div className="metric">
                      <span className="metric-label">Cost</span>
                      <span className="metric-value">{formatCost(quest.cost)}</span>
                    </div>
                    <div className="metric">
                      <span className="metric-label">Duration</span>
                      <span className="metric-value">{timeAgo(quest.started_at || quest.created_at)}</span>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}