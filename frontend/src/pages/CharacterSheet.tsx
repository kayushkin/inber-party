import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { api, useStore } from '../store';
import type { AgentDetail, InberAgent, InberQuest } from '../store';
import './CharacterSheet.css';

export default function CharacterSheet() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [agent, setAgent] = useState<AgentDetail | null>(null);
  const [inberAgent, setInberAgent] = useState<InberAgent | null>(null);
  const [quests, setQuests] = useState<InberQuest[]>([]);
  const [loading, setLoading] = useState(true);
  const inberAgents = useStore((state) => state.inberAgents);
  const startPolling = useStore((state) => state.startPolling);
  const stopPolling = useStore((state) => state.stopPolling);

  useEffect(() => {
    startPolling(15000);
    return () => stopPolling();
  }, [startPolling, stopPolling]);

  useEffect(() => {
    if (id) {
      setLoading(true);
      const idx = parseInt(id);

      // Load basic agent detail
      api.getAgent(idx)
        .then(setAgent)
        .catch(() => navigate('/'));

      // Load inber-specific data
      const ia = inberAgents[idx - 1];
      if (ia) {
        setInberAgent(ia);
        api.getAgentQuests(ia.id).then(setQuests);
      }

      setLoading(false);
    }
  }, [id, navigate, inberAgents]);

  // Update inberAgent when store refreshes
  useEffect(() => {
    if (id) {
      const idx = parseInt(id);
      const ia = inberAgents[idx - 1];
      if (ia) setInberAgent(ia);
    }
  }, [inberAgents, id]);

  if (loading || !agent) {
    return <div className="loading">Loading character sheet...</div>;
  }

  const isInber = !!inberAgent;
  const xpToNextLevel = isInber ? (inberAgent!.xp + inberAgent!.xp_to_next) : (agent.level + 1) * 100;
  const currentXP = isInber ? inberAgent!.xp : agent.xp;
  const xpProgress = xpToNextLevel > 0 ? (currentXP / xpToNextLevel) * 100 : 0;

  const formatTokens = (n: number) => {
    if (n >= 1000000) return `${(n / 1000000).toFixed(1)}M`;
    if (n >= 1000) return `${(n / 1000).toFixed(1)}k`;
    return String(n);
  };

  const formatCost = (c: number) => {
    if (c <= 0) return '$0';
    if (c < 0.01) return `$${c.toFixed(4)}`;
    return `$${c.toFixed(2)}`;
  };

  const completedQuests = quests.filter(q => q.status === 'completed').length;
  const failedQuests = quests.filter(q => q.status === 'failed').length;
  const totalQuestTokens = quests.reduce((sum, q) => sum + q.tokens_used, 0);
  const totalQuestCost = quests.reduce((sum, q) => sum + q.cost, 0);

  return (
    <div className="character-sheet">
      <button className="back-button" onClick={() => navigate('/')}>
        ← Back to Camp
      </button>

      <div className="character-header">
        <div className="character-avatar">{agent.avatar_emoji}</div>
        <div className="character-title-section">
          <h1 className="character-name">{agent.name}</h1>
          <h2 className="character-title">{agent.title}</h2>
          <div className="character-class">{agent.class}</div>
        </div>
        {isInber && (
          <div className="character-live-badge">🔮 LIVE</div>
        )}
      </div>

      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-label">Level</div>
          <div className="stat-value">{agent.level}</div>
        </div>
        <div className="stat-card">
          <div className="stat-label">XP</div>
          <div className="stat-value">{currentXP}</div>
          <div className="stat-progress">
            <div className="progress-bar">
              <div className="progress-fill" style={{ width: `${xpProgress}%` }} />
            </div>
            <div className="progress-text">{currentXP} / {xpToNextLevel}</div>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Energy</div>
          <div className="stat-value">{agent.energy}%</div>
          <div className="stat-progress">
            <div className="progress-bar energy">
              <div className="progress-fill" style={{ width: `${agent.energy}%` }} />
            </div>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Status</div>
          <div className={`stat-value status-${agent.status}`}>
            {agent.status.replace('_', ' ').toUpperCase()}
          </div>
        </div>
      </div>

      {/* Inber-specific stats row */}
      {isInber && (
        <div className="stats-grid inber-stats">
          <div className="stat-card">
            <div className="stat-label">Total Tokens</div>
            <div className="stat-value tokens-color">{formatTokens(inberAgent!.total_tokens)}</div>
          </div>
          <div className="stat-card">
            <div className="stat-label">Total Cost</div>
            <div className="stat-value cost-color">{formatCost(inberAgent!.total_cost)}</div>
          </div>
          <div className="stat-card">
            <div className="stat-label">Sessions</div>
            <div className="stat-value">{inberAgent!.session_count}</div>
          </div>
          <div className="stat-card">
            <div className="stat-label">Quests</div>
            <div className="stat-value">
              <span className="success-count">{inberAgent!.quest_count}</span>
              {inberAgent!.error_count > 0 && (
                <span className="error-count"> / {inberAgent!.error_count} ✗</span>
              )}
            </div>
          </div>
        </div>
      )}

      <div className="content-grid">
        <div className="section skills-section">
          <h3>Skills</h3>
          {agent.skills && agent.skills.length > 0 ? (
            <div className="skills-list">
              {agent.skills.map((skill) => (
                <div key={skill.id} className="skill-item">
                  <div className="skill-name">{skill.skill_name}</div>
                  <div className="skill-meta">
                    <span className="skill-level">Level {skill.level}</span>
                    <span className="skill-count">{skill.task_count} uses</span>
                  </div>
                  <div className="skill-bar">
                    <div className="skill-bar-fill" style={{ width: `${Math.min(skill.level * 10, 100)}%` }} />
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p className="empty-message">No skills yet</p>
          )}
        </div>

        <div className="section quest-log-section">
          <h3>Quest History {quests.length > 0 && <span className="quest-count">({quests.length})</span>}</h3>
          {quests.length > 0 ? (
            <>
              {/* Quest summary bar */}
              <div className="quest-summary">
                <div className="summary-item">
                  <span className="summary-label">Completed</span>
                  <span className="summary-value success">{completedQuests}</span>
                </div>
                <div className="summary-item">
                  <span className="summary-label">Failed</span>
                  <span className="summary-value fail">{failedQuests}</span>
                </div>
                <div className="summary-item">
                  <span className="summary-label">Tokens</span>
                  <span className="summary-value tokens">{formatTokens(totalQuestTokens)}</span>
                </div>
                <div className="summary-item">
                  <span className="summary-label">Cost</span>
                  <span className="summary-value cost">{formatCost(totalQuestCost)}</span>
                </div>
              </div>
              <div className="quest-list">
                {quests.slice(0, 20).map((quest) => (
                  <div key={quest.id} className={`quest-item status-${quest.status}`}>
                    <div className="quest-header">
                      <div className="quest-name">{quest.name}</div>
                      <div className="quest-status">{quest.status.replace('_', ' ')}</div>
                    </div>
                    {quest.description && (
                      <div className="quest-description">
                        {quest.description.slice(0, 120)}{quest.description.length > 120 ? '...' : ''}
                      </div>
                    )}
                    <div className="quest-meta">
                      <span className="quest-tokens">🔮 {formatTokens(quest.tokens_used)}</span>
                      <span className="quest-turns">↻ {quest.turns} turns</span>
                      <span className="quest-reward">+{quest.xp_reward} XP</span>
                      {(quest.children ?? 0) > 0 && (
                        <span className="quest-children">⚔️ {quest.children} sub</span>
                      )}
                    </div>
                  </div>
                ))}
                {quests.length > 20 && (
                  <p className="empty-message">...and {quests.length - 20} more quests</p>
                )}
              </div>
            </>
          ) : (
            <p className="empty-message">No quest history</p>
          )}
        </div>

        <div className="section achievements-section">
          <h3>Achievements</h3>
          {agent.achievements && agent.achievements.length > 0 ? (
            <div className="achievements-list">
              {agent.achievements.map((achievement) => (
                <div key={achievement.id} className="achievement-item">
                  <div className="achievement-icon">🏆</div>
                  <div className="achievement-info">
                    <div className="achievement-name">{achievement.achievement_name}</div>
                    <div className="achievement-date">
                      {new Date(achievement.unlocked_at).toLocaleDateString()}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p className="empty-message">No achievements yet</p>
          )}
        </div>
      </div>
    </div>
  );
}
