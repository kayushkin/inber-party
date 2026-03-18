import { useNavigate } from 'react-router-dom';
import { useStore, classColor, formatTokens, timeAgo } from '../store';
import type { RPGAgent } from '../store';
import ChatPanel from '../components/ChatPanel';
import LevelUpAnimation from '../components/LevelUpAnimation';
import { SkeletonAgentCard, SkeletonStatsBar } from '../components/SkeletonLoader';
import './TavernView.css';

const GUILD_NAMES: Record<string, string> = {
  inber: '🏰 Inber Guild',
  openclaw: '🐾 OpenClaw Guild',
};

function guildLabel(orchestrator: string) {
  return GUILD_NAMES[orchestrator] || `⚔️ ${orchestrator || 'Unknown'} Guild`;
}

interface ActivityEvent {
  id: string;
  type: 'quest_completed' | 'level_up' | 'agent_active' | 'quest_started';
  timestamp: string;
  content: string;
  agentName?: string;
  agentClass?: string;
  icon: string;
}

export default function TavernView() {
  const navigate = useNavigate();
  const agents = useStore((s) => s.agents);
  const quests = useStore((s) => s.quests);
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

  // Generate activity feed
  const generateActivityFeed = (): ActivityEvent[] => {
    const activities: ActivityEvent[] = [];

    // Recent quest completions
    const recentCompletedQuests = quests
      .filter(q => q.status === 'completed' && q.completed_at)
      .sort((a, b) => new Date(b.completed_at!).getTime() - new Date(a.completed_at!).getTime())
      .slice(0, 5);

    recentCompletedQuests.forEach(quest => {
      activities.push({
        id: `quest-${quest.id}`,
        type: 'quest_completed',
        timestamp: quest.completed_at!,
        content: `${quest.assigned_agent_name || 'Unknown Adventurer'} completed quest "${quest.name}"`,
        agentName: quest.assigned_agent_name,
        icon: '🎯'
      });
    });

    // Recent quest starts
    const recentStartedQuests = quests
      .filter(q => q.status === 'active' && q.started_at)
      .sort((a, b) => new Date(b.started_at!).getTime() - new Date(a.started_at!).getTime())
      .slice(0, 3);

    recentStartedQuests.forEach(quest => {
      activities.push({
        id: `quest-start-${quest.id}`,
        type: 'quest_started',
        timestamp: quest.started_at!,
        content: `${quest.assigned_agent_name || 'Unknown Adventurer'} began quest "${quest.name}"`,
        agentName: quest.assigned_agent_name,
        icon: '⚔️'
      });
    });

    // Recent agent activity (last active)
    const recentActiveAgents = agents
      .filter(a => a.last_active)
      .sort((a, b) => new Date(b.last_active!).getTime() - new Date(a.last_active!).getTime())
      .slice(0, 8);

    recentActiveAgents.forEach(agent => {
      const timeDiff = Date.now() - new Date(agent.last_active!).getTime();
      // Only show activity from last 2 hours
      if (timeDiff < 2 * 60 * 60 * 1000) {
        activities.push({
          id: `activity-${agent.id}`,
          type: 'agent_active',
          timestamp: agent.last_active!,
          content: `${agent.name} was active`,
          agentName: agent.name,
          agentClass: agent.class,
          icon: '👁️'
        });
      }
    });

    // Sort by timestamp (newest first) and limit to last 15 events
    return activities
      .sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
      .slice(0, 15);
  };

  const activityFeed = generateActivityFeed();

  // Group agents by orchestrator for the main area
  const groups = new Map<string, RPGAgent[]>();
  for (const agent of agents) {
    const key = agent.orchestrator || 'unknown';
    if (!groups.has(key)) groups.set(key, []);
    groups.get(key)!.push(agent);
  }
  const sortedKeys = [...groups.keys()].sort((a, b) => {
    const order: Record<string, number> = { inber: 0, openclaw: 1 };
    return (order[a] ?? 99) - (order[b] ?? 99);
  });

  // Get most active agents for tavern spotlight
  const activeAgents = agents
    .filter(a => a.status === 'working' || a.status === 'on_quest')
    .slice(0, 6);

  return (
    <div className={`tavern-layout ${selectedAgent ? 'chat-open' : ''}`}>
      <div className="tavern-main">
        {/* Tavern Header */}
        <div className="tavern-header">
          <h1 className="tavern-title">🏰 The Adventurer's Tavern</h1>
          <p className="tavern-subtitle">Where brave agents gather and share tales of their quests</p>
          <button 
            className="create-adventurer-btn"
            onClick={() => navigate('/create-adventurer')}
            title="Recruit a new adventurer"
          >
            🌟 Create New Adventurer
          </button>
        </div>

        {/* Guild Stats Bar */}
        {isLoadingStats ? (
          <SkeletonStatsBar />
        ) : stats && (
          <div className="guild-bar tavern-stats">
            <div className="guild-stat"><span className="gs-val">{stats.total_agents}</span><span className="gs-lbl">Adventurers</span></div>
            <div className="guild-stat"><span className="gs-val">{stats.active_quests}</span><span className="gs-lbl">Active Quests</span></div>
            <div className="guild-stat"><span className="gs-val">{stats.completed_quests}</span><span className="gs-lbl">Completed</span></div>
            <div className="guild-stat"><span className="gs-val">{formatTokens(stats.total_tokens)}</span><span className="gs-lbl">Experience</span></div>
            <div className="guild-stat"><span className="gs-val">{stats.total_sessions}</span><span className="gs-lbl">Adventures</span></div>
          </div>
        )}

        <div className="tavern-content">
          {/* Activity Feed */}
          <div className="tavern-activity">
            <h2 className="section-title">📜 Recent Tales & Events</h2>
            <div className="activity-feed">
              {activityFeed.length > 0 ? activityFeed.map(event => (
                <div key={event.id} className={`activity-item activity-${event.type}`}>
                  <span className="activity-icon">{event.icon}</span>
                  <div className="activity-content">
                    <span className="activity-text">{event.content}</span>
                    <span className="activity-time">{timeAgo(event.timestamp)}</span>
                  </div>
                </div>
              )) : (
                <div className="activity-empty">
                  <span className="activity-icon">🌙</span>
                  <p>The tavern is quiet... waiting for adventurers to begin their quests.</p>
                </div>
              )}
            </div>
          </div>

          {/* Currently Active Agents */}
          {activeAgents.length > 0 && (
            <div className="tavern-spotlight">
              <h2 className="section-title">⚔️ Adventurers on Duty</h2>
              <div className="active-agents">
                {activeAgents.map(agent => (
                  <div
                    key={agent.id}
                    className={`active-agent ${getStatusClass(agent.status)}`}
                    style={{ '--cc': classColor(agent.class) } as React.CSSProperties}
                    onClick={() => handleAgentClick(agent)}
                  >
                    <div className="aa-avatar">
                      <img
                        src={`/avatars/${agent.name.toLowerCase().replace(/\s+/g, '-')}.png`}
                        alt={agent.name}
                        onError={(e) => {
                          e.currentTarget.style.display = 'none';
                          const sibling = e.currentTarget.nextElementSibling as HTMLElement;
                          if (sibling) sibling.style.display = 'block';
                        }}
                        style={{ width: '32px', height: '32px', borderRadius: '50%' }}
                      />
                      <span style={{ display: 'none', fontSize: '24px' }}>{agent.avatar_emoji}</span>
                    </div>
                    <div className="aa-info">
                      <span className="aa-name" style={{ color: classColor(agent.class) }}>{agent.name}</span>
                      <span className="aa-status">Lv {agent.level} {agent.class}</span>
                    </div>
                    <div className={`aa-status-dot ${getStatusClass(agent.status)}`} />
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* All Guilds Section */}
        <div className="tavern-guilds">
          <h2 className="section-title">🏰 Guild Halls</h2>
          
          {isLoadingAgents ? (
            <>
              <div className="guild-section">
                <h3 className="guild-header">{guildLabel('inber')}</h3>
                <div className="agents-grid compact">
                  {[...Array(6)].map((_, i) => (
                    <SkeletonAgentCard key={`skeleton-inber-${i}`} />
                  ))}
                </div>
              </div>
              <div className="guild-section">
                <h3 className="guild-header">{guildLabel('openclaw')}</h3>
                <div className="agents-grid compact">
                  {[...Array(4)].map((_, i) => (
                    <SkeletonAgentCard key={`skeleton-openclaw-${i}`} />
                  ))}
                </div>
              </div>
            </>
          ) : (
            sortedKeys.map((key) => (
              <div key={key} className="guild-section">
                <h3 className="guild-header">{guildLabel(key)}</h3>
                <div className="agents-grid compact">
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
              <p>🍃 The tavern is empty. No adventurers have arrived yet.</p>
              <p>Make sure inber is running at the configured INBER_URL.</p>
            </div>
          )}
        </div>
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