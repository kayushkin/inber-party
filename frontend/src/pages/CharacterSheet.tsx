import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { api, useStore } from '../store';
import type { AgentDetail, InberAgent, InberQuest, InberAchievement, QuestHistoryEntry } from '../store';
import './CharacterSheet.css';

const CLASS_COLORS: Record<string, string> = {
  Wizard: '#a78bfa',
  Healer: '#4ade80',
  Ranger: '#60a5fa',
  Warrior: '#f87171',
};

// SVG Sparkline component
function Sparkline({ data, color, height = 40, width = '100%' }: {
  data: number[];
  color: string;
  height?: number;
  width?: string | number;
}) {
  if (data.length < 2) return null;
  const max = Math.max(...data, 1);
  const min = Math.min(...data, 0);
  const range = max - min || 1;
  const w = 200;
  const h = height;
  const points = data.map((v, i) => {
    const x = (i / (data.length - 1)) * w;
    const y = h - ((v - min) / range) * (h - 4) - 2;
    return `${x},${y}`;
  }).join(' ');
  const areaPoints = `0,${h} ${points} ${w},${h}`;

  return (
    <svg viewBox={`0 0 ${w} ${h}`} width={width} height={height} preserveAspectRatio="none" style={{ display: 'block' }}>
      <polygon points={areaPoints} fill={color} opacity="0.15" />
      <polyline points={points} fill="none" stroke={color} strokeWidth="2" />
    </svg>
  );
}

// Bar chart
function BarChart({ data, color, height = 50, width = '100%' }: {
  data: { value: number; label?: string; status?: string }[];
  color: string;
  height?: number;
  width?: string | number;
}) {
  if (data.length === 0) return null;
  const max = Math.max(...data.map(d => d.value), 1);
  const w = 200;
  const h = height;
  const barW = Math.max(2, (w / data.length) - 1);

  return (
    <svg viewBox={`0 0 ${w} ${h}`} width={width} height={height} preserveAspectRatio="none" style={{ display: 'block' }}>
      {data.map((d, i) => {
        const barH = (d.value / max) * (h - 4);
        const x = (i / data.length) * w;
        const barColor = d.status === 'failed' ? '#ef4444' : color;
        return (
          <rect key={i} x={x} y={h - barH - 2} width={barW} height={barH}
            fill={barColor} rx="1" opacity="0.8" />
        );
      })}
    </svg>
  );
}

export default function CharacterSheet() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [agent, setAgent] = useState<AgentDetail | null>(null);
  const [inberAgent, setInberAgent] = useState<InberAgent | null>(null);
  const [quests, setQuests] = useState<InberQuest[]>([]);
  const [achievements, setAchievements] = useState<InberAchievement[]>([]);
  const [questHistory, setQuestHistory] = useState<QuestHistoryEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [levelUpAnim, setLevelUpAnim] = useState(false);
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

      api.getAgent(idx)
        .then(setAgent)
        .catch(() => navigate('/'));

      const ia = inberAgents[idx - 1];
      if (ia) {
        setInberAgent(ia);
        api.getAgentQuests(ia.id).then(setQuests);
        api.getAchievements(ia.id).then(setAchievements);
        api.getQuestHistory(ia.id, 20).then(setQuestHistory);
      }

      setLoading(false);
    }
  }, [id, navigate, inberAgents]);

  // Detect level changes for animation
  const prevLevel = useState(0);
  useEffect(() => {
    if (id) {
      const idx = parseInt(id);
      const ia = inberAgents[idx - 1];
      if (ia) {
        setInberAgent(ia);
        if (prevLevel[0] > 0 && ia.level > prevLevel[0]) {
          setLevelUpAnim(true);
          setTimeout(() => setLevelUpAnim(false), 2000);
        }
        prevLevel[1](ia.level);
      }
    }
  }, [inberAgents, id]);

  if (loading || !agent) {
    return <div className="loading">Loading character sheet...</div>;
  }

  const isInber = !!inberAgent;
  const xpToNextLevel = isInber ? (inberAgent!.xp + inberAgent!.xp_to_next) : (agent.level + 1) * 100;
  const currentXP = isInber ? inberAgent!.xp : agent.xp;
  const xpProgress = xpToNextLevel > 0 ? (currentXP / xpToNextLevel) * 100 : 0;
  const classColor = CLASS_COLORS[inberAgent?.class || agent.class] || '#d4af37';

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

      <div className="character-header" style={{ borderColor: classColor }}>
        <div className="character-avatar">
          {agent.avatar_emoji}
          {levelUpAnim && <div className="level-up-effect">⬆️ LEVEL UP!</div>}
        </div>
        <div className="character-title-section">
          <h1 className="character-name" style={{ color: classColor }}>{agent.name}</h1>
          <h2 className="character-title">{agent.title}</h2>
          <div className="character-class" style={{ borderColor: classColor, color: classColor }}>
            {inberAgent?.class || agent.class}
          </div>
        </div>
        {isInber && (
          <div className="character-live-badge">🔮 LIVE</div>
        )}
      </div>

      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-label">Level</div>
          <div className="stat-value" style={{ color: classColor }}>{agent.level}</div>
        </div>
        <div className="stat-card">
          <div className="stat-label">XP</div>
          <div className="stat-value">{currentXP}</div>
          <div className="stat-progress">
            <div className="progress-bar">
              <div className="progress-fill" style={{ width: `${xpProgress}%`, background: `linear-gradient(90deg, ${classColor}, ${classColor}dd)` }} />
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

      {/* Token & Cost Charts */}
      {questHistory.length > 1 && (
        <div className="charts-section">
          <h3>Activity</h3>
          <div className="charts-grid">
            <div className="chart-card">
              <div className="chart-label">Tokens per Quest (last {questHistory.length})</div>
              <BarChart
                data={questHistory.map(q => ({ value: q.tokens, status: q.status }))}
                color={classColor}
                height={60}
              />
            </div>
            <div className="chart-card">
              <div className="chart-label">Token Trend</div>
              <Sparkline
                data={questHistory.map(q => q.tokens)}
                color={classColor}
                height={60}
              />
            </div>
            <div className="chart-card">
              <div className="chart-label">Cost per Quest</div>
              <Sparkline
                data={questHistory.map(q => q.cost)}
                color="#4ade80"
                height={60}
              />
            </div>
          </div>
        </div>
      )}

      <div className="content-grid">
        {/* Achievements */}
        <div className="section achievements-section">
          <h3>Achievements {achievements.length > 0 && <span className="badge-count">{achievements.length}</span>}</h3>
          {achievements.length > 0 ? (
            <div className="achievements-grid">
              {achievements.map((a) => (
                <div key={a.id} className="achievement-badge" title={a.description}>
                  <div className="badge-icon">{a.icon}</div>
                  <div className="badge-name">{a.name}</div>
                </div>
              ))}
            </div>
          ) : (
            <p className="empty-message">No achievements yet</p>
          )}
        </div>

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
                    <div className="skill-bar-fill" style={{ width: `${Math.min(skill.level * 10, 100)}%`, background: `linear-gradient(90deg, ${classColor}, ${classColor}dd)` }} />
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
      </div>
    </div>
  );
}
