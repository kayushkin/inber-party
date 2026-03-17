import { useNavigate } from 'react-router-dom';
import { useStore, classColor, formatTokens, timeAgo } from '../store';
import type { RPGAgent } from '../store';
import ChatPanel from '../components/ChatPanel';
import './CampView.css';

export default function CampView() {
  const navigate = useNavigate();
  const agents = useStore((s) => s.agents);
  const stats = useStore((s) => s.stats);
  const selectedAgent = useStore((s) => s.selectedAgent);
  const setSelectedAgent = useStore((s) => s.setSelectedAgent);

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

  return (
    <div className={`camp-layout ${selectedAgent ? 'chat-open' : ''}`}>
      <div className="camp-main">
        {/* Guild Stats Bar */}
        {stats && (
          <div className="guild-bar">
            <div className="guild-stat"><span className="gs-val">{stats.total_agents}</span><span className="gs-lbl">Agents</span></div>
            <div className="guild-stat"><span className="gs-val">{stats.active_quests}</span><span className="gs-lbl">Active</span></div>
            <div className="guild-stat"><span className="gs-val">{stats.completed_quests}</span><span className="gs-lbl">Done</span></div>
            <div className="guild-stat"><span className="gs-val">{formatTokens(stats.total_tokens)}</span><span className="gs-lbl">Tokens</span></div>
            <div className="guild-stat"><span className="gs-val">{stats.total_sessions}</span><span className="gs-lbl">Sessions</span></div>
          </div>
        )}

        {/* Agent Grid */}
        <div className="agents-grid">
          {agents.map((agent) => {
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
                  <div className="ac-avatar">{agent.avatar_emoji}</div>
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

        {agents.length === 0 && (
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
    </div>
  );
}
