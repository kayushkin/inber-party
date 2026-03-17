import { useNavigate } from 'react-router-dom';
import { useStore, classColor, formatTokens, timeAgo } from '../store';
import type { RPGAgent } from '../store';
import ChatPanel from '../components/ChatPanel';
import LevelUpAnimation from '../components/LevelUpAnimation';
import { SkeletonAgentCard, SkeletonStatsBar } from '../components/SkeletonLoader';
import './CampView.css';

const GUILD_NAMES: Record<string, string> = {
  inber: '🏰 Inber Guild',
  openclaw: '🐾 OpenClaw Guild',
};

function guildLabel(orchestrator: string) {
  return GUILD_NAMES[orchestrator] || `⚔️ ${orchestrator || 'Unknown'} Guild`;
}

export default function CampView() {
  const navigate = useNavigate();
  const agents = useStore((s) => s.agents);
  const stats = useStore((s) => s.stats);
  const selectedAgent = useStore((s) => s.selectedAgent);
  const setSelectedAgent = useStore((s) => s.setSelectedAgent);
  const levelUpTriggers = useStore((s) => s.levelUpTriggers);
  const isLoadingAgents = useStore((s) => s.isLoadingAgents);
  const isLoadingStats = useStore((s) => s.isLoadingStats);
  const hasInitialLoad = useStore((s) => s.hasInitialLoad);

  const handleAgentClick = (agent: RPGAgent) => {
    setSelectedAgent(agent.id);
  };

  const getStatusClass = (status: string) => {
    switch (status) {
      case 'working': return 'status-working';
      case 'on_quest': return 'status-quest';
      case 'stuck': return 'status-stuck';
      default: return 'status-idle';
    }
  };

  // Group agents by orchestrator
  const groups = new Map<string, RPGAgent[]>();
  for (const agent of agents) {
    const key = agent.orchestrator || 'unknown';
    if (!groups.has(key)) groups.set(key, []);
    groups.get(key)!.push(agent);
  }
  // Sort: inber first, openclaw second, rest after
  const sortedKeys = [...groups.keys()].sort((a, b) => {
    const order: Record<string, number> = { inber: 0, openclaw: 1 };
    return (order[a] ?? 99) - (order[b] ?? 99);
  });

  return (
    <div className={`camp-layout ${selectedAgent ? 'chat-open' : ''}`}>
      <div className="camp-main">
        {/* Guild Stats Bar */}
        {isLoadingStats ? (
          <SkeletonStatsBar />
        ) : stats && (
          <div className="guild-bar">
            <div className="guild-stat"><span className="gs-val">{stats.total_agents}</span><span className="gs-lbl">Agents</span></div>
            <div className="guild-stat"><span className="gs-val">{stats.active_quests}</span><span className="gs-lbl">Active</span></div>
            <div className="guild-stat"><span className="gs-val">{stats.completed_quests}</span><span className="gs-lbl">Done</span></div>
            <div className="guild-stat"><span className="gs-val">{formatTokens(stats.total_tokens)}</span><span className="gs-lbl">Tokens</span></div>
            <div className="guild-stat"><span className="gs-val">{stats.total_sessions}</span><span className="gs-lbl">Sessions</span></div>
          </div>
        )}

        {/* Grouped Agent Grids */}
        {isLoadingAgents ? (
          // Show skeleton loaders while agents are loading
          <>
            <div className="guild-section">
              <h2 className="guild-header">{guildLabel('inber')}</h2>
              <div className="agents-grid">
                {[...Array(6)].map((_, i) => (
                  <SkeletonAgentCard key={`skeleton-inber-${i}`} />
                ))}
              </div>
            </div>
            <div className="guild-section">
              <h2 className="guild-header">{guildLabel('openclaw')}</h2>
              <div className="agents-grid">
                {[...Array(4)].map((_, i) => (
                  <SkeletonAgentCard key={`skeleton-openclaw-${i}`} />
                ))}
              </div>
            </div>
          </>
        ) : (
          sortedKeys.map((key) => (
            <div key={key} className="guild-section">
              <h2 className="guild-header">{guildLabel(key)}</h2>
              <div className="agents-grid">
                {groups.get(key)!.map((agent) => {
                const cc = classColor(agent.class);
                const xpPct = agent.xp_to_next > 0
                  ? (agent.xp / (agent.xp + agent.xp_to_next)) * 100
                  : 100;
                return (
                  <div
                    key={agent.id}
                    className={`agent-card ${getStatusClass(agent.status)} ${selectedAgent === agent.id ? 'selected' : ''}`}
                    style={{ '--cc': cc } as React.CSSProperties}
                    onClick={() => handleAgentClick(agent)}
                  >
                    <div className="ac-top">
                      <div className="ac-avatar">
                      <img
                        src={`/avatars/${agent.name.toLowerCase().replace(/\s+/g, '-')}.png`}
                        alt={agent.name}
                        onError={(e) => {
                          e.currentTarget.style.display = 'none';
                          const sibling = e.currentTarget.nextElementSibling as HTMLElement;
                          if (sibling) sibling.style.display = 'block';
                        }}
                        style={{ width: '40px', height: '40px', borderRadius: '50%' }}
                      />
                      <span style={{ display: 'none' }}>{agent.avatar_emoji}</span>
                    </div>
                      <div className="ac-info">
                        <div className="ac-name" style={{ color: cc }}>{agent.name}</div>
                        <div className="ac-class">{agent.class} · Lv {agent.level}</div>
                      </div>
                      <div className={`ac-status-dot ${getStatusClass(agent.status)}`} title={agent.status} />
                    </div>
                    <div className="ac-xp-bar">
                      <div className="ac-xp-fill" style={{ width: `${xpPct}%`, background: cc }} />
                    </div>
                    <div className="ac-bottom">
                      <span className="ac-tokens">🔮 {formatTokens(agent.total_tokens)}</span>
                      <span className="ac-time">{timeAgo(agent.last_active)}</span>
                    </div>
                    <span className="ac-chat-hint">💬 Click to chat</span>
                    <button
                      className="ac-detail-btn"
                      onClick={(e) => { e.stopPropagation(); navigate(`/agent/${agent.id}`); }}
                      title="Character Sheet"
                    >
                      📋
                    </button>
                  </div>
                );
                })}
              </div>
            </div>
          ))
        )}

        {!isLoadingAgents && agents.length === 0 && hasInitialLoad && (
          <div className="empty-state">
            <p>No agents found. Make sure inber is running at the configured INBER_URL.</p>
          </div>
        )}
      </div>

      {/* Chat Panel */}
      {selectedAgent && (
        <ChatPanel
          agentId={selectedAgent}
          agent={agents.find((a) => a.id === selectedAgent)}
          onClose={() => setSelectedAgent(null)}
        />
      )}

      {/* Level-up animations for all agents */}
      {agents.map((agent) => (
        <LevelUpAnimation
          key={`levelup-${agent.id}`}
          trigger={levelUpTriggers[agent.id] || 0}
          agentName={agent.name}
          level={agent.level}
          color={classColor(agent.class)}
        />
      ))}
    </div>
  );
}
