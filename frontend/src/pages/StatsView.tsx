import { useStore, formatTokens, formatCost, classColor } from '../store';
import { Skeleton } from '../components/SkeletonLoader';
import './StatsView.css';

export default function StatsView() {
  const agents = useStore((s) => s.agents);
  const stats = useStore((s) => s.stats);
  const isLoadingStats = useStore((s) => s.isLoadingStats);
  const isLoadingAgents = useStore((s) => s.isLoadingAgents);
  const hasInitialLoad = useStore((s) => s.hasInitialLoad);

  if (isLoadingStats || !hasInitialLoad) {
    return (
      <div className="stats-view">
        <h1>📊 Guild Statistics</h1>
        
        <div className="overview-grid">
          {[...Array(8)].map((_, i) => (
            <div key={i} className="ov-card">
              <div className="ov-val"><Skeleton width="60px" height="24px" /></div>
              <div className="ov-lbl"><Skeleton width="80px" height="16px" /></div>
            </div>
          ))}
        </div>

        <h2>Agent Leaderboard</h2>
        <div className="leaderboard">
          {[...Array(6)].map((_, i) => (
            <div key={i} className="lb-item skeleton-card">
              <div className="lb-rank"><Skeleton width="20px" height="20px" /></div>
              <div className="lb-info">
                <Skeleton width="120px" height="18px" />
                <Skeleton width="80px" height="14px" />
              </div>
              <div className="lb-stats">
                <Skeleton width="60px" height="16px" />
                <Skeleton width="50px" height="14px" />
              </div>
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (!stats) return <div className="loading">No stats available.</div>;

  // Sort agents by total tokens descending
  const sorted = [...agents].sort((a, b) => b.total_tokens - a.total_tokens);

  return (
    <div className="stats-view">
      <h1>📊 Guild Statistics</h1>

      <div className="overview-grid">
        <div className="ov-card"><div className="ov-val">{stats.total_agents}</div><div className="ov-lbl">Adventurers</div></div>
        <div className="ov-card"><div className="ov-val">{stats.total_sessions}</div><div className="ov-lbl">Sessions</div></div>
        <div className="ov-card"><div className="ov-val">{formatTokens(stats.total_tokens)}</div><div className="ov-lbl">Total Tokens</div></div>
        <div className="ov-card"><div className="ov-val">{formatCost(stats.total_cost)}</div><div className="ov-lbl">Total Cost</div></div>
        <div className="ov-card"><div className="ov-val">{stats.completed_quests}</div><div className="ov-lbl">Quests Done</div></div>
        <div className="ov-card"><div className="ov-val">{stats.failed_quests}</div><div className="ov-lbl">Quests Failed</div></div>
        <div className="ov-card"><div className="ov-val">{stats.active_quests}</div><div className="ov-lbl">Active</div></div>
        <div className="ov-card"><div className="ov-val">{stats.uptime || '—'}</div><div className="ov-lbl">Uptime</div></div>
      </div>

      <h2>Agent Leaderboard</h2>
      <div className="leaderboard">
        {isLoadingAgents ? (
          [...Array(6)].map((_, i) => (
            <div key={i} className="lb-row skeleton-card">
              <div className="lb-rank"><Skeleton width="20px" height="20px" /></div>
              <div className="lb-avatar"><Skeleton width="24px" height="24px" borderRadius="50%" /></div>
              <div className="lb-info">
                <Skeleton width="120px" height="18px" />
                <Skeleton width="80px" height="14px" />
              </div>
              <div className="lb-bar-container">
                <Skeleton width="100%" height="8px" borderRadius="4px" />
              </div>
              <div className="lb-stats">
                <Skeleton width="60px" height="16px" />
                <Skeleton width="50px" height="14px" />
              </div>
            </div>
          ))
        ) : (
          sorted.map((agent, i) => {
          const cc = classColor(agent.class);
          const maxTokens = sorted[0]?.total_tokens || 1;
          const pct = (agent.total_tokens / maxTokens) * 100;
          return (
            <div key={agent.id} className="lb-row">
              <div className="lb-rank">#{i + 1}</div>
              <div className="lb-avatar">{agent.avatar_emoji}</div>
              <div className="lb-info">
                <div className="lb-name" style={{ color: cc }}>{agent.name}</div>
                <div className="lb-class">{agent.class} · Lv {agent.level}</div>
              </div>
              <div className="lb-bar-container">
                <div className="lb-bar" style={{ width: `${pct}%`, background: cc }} />
              </div>
              <div className="lb-stats">
                <span className="lb-tokens">{formatTokens(agent.total_tokens)}</span>
                <span className="lb-cost">{formatCost(agent.total_cost)}</span>
              </div>
            </div>
          );
          })
        )}
      </div>
    </div>
  );
}
