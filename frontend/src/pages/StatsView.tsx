import { useStore, formatTokens, formatCost, classColor } from '../store';
import { Skeleton } from '../components/SkeletonLoader';
import Tooltip from '../components/Tooltip';
import { STAT_TOOLTIPS } from '../constants/tooltips';
import './StatsView.css';

// Export utility functions
function exportAsJSON(agents: any[], stats: any) {
  const exportData = {
    exported_at: new Date().toISOString(),
    guild_stats: stats,
    agent_stats: agents.map(agent => ({
      id: agent.id,
      name: agent.name,
      title: agent.title,
      class: agent.class,
      level: agent.level,
      xp: agent.xp,
      gold: agent.gold,
      total_tokens: agent.total_tokens,
      total_cost: agent.total_cost,
      session_count: agent.session_count,
      quest_count: agent.quest_count,
      error_count: agent.error_count,
      last_active: agent.last_active,
      skills: agent.skills,
      reputation: agent.reputation,
      mood: agent.mood,
      mood_score: agent.mood_score,
      workload: agent.workload
    }))
  };
  
  const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `guild-stats-${new Date().toISOString().split('T')[0]}.json`;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}

function exportAsCSV(agents: any[], stats: any) {
  // Guild stats header
  const guildData = [
    'Guild Statistics',
    '',
    'Total Adventurers,' + stats?.total_agents,
    'Total Sessions,' + stats?.total_sessions, 
    'Total Tokens,' + stats?.total_tokens,
    'Total Cost,$' + stats?.total_cost?.toFixed(4),
    'Completed Quests,' + stats?.completed_quests,
    'Failed Quests,' + stats?.failed_quests,
    'Active Quests,' + stats?.active_quests,
    'Uptime,' + stats?.uptime,
    '',
    'Agent Statistics',
    ''
  ];
  
  // CSV headers for agents
  const headers = [
    'Name',
    'Title', 
    'Class',
    'Level',
    'XP',
    'Gold',
    'Total Tokens',
    'Total Cost',
    'Sessions',
    'Quests',
    'Errors',
    'Last Active',
    'Mood',
    'Mood Score',
    'Workload',
    'Top Skills'
  ];
  
  // Agent data rows
  const rows = agents.map(agent => [
    agent.name,
    agent.title,
    agent.class,
    agent.level,
    agent.xp,
    agent.gold,
    agent.total_tokens,
    '$' + agent.total_cost?.toFixed(4),
    agent.session_count,
    agent.quest_count,
    agent.error_count,
    agent.last_active || '',
    agent.mood || '',
    agent.mood_score || '',
    agent.workload || '',
    agent.skills?.slice(0, 3).map((s: any) => `${s.skill_name}(${s.level})`).join('; ') || ''
  ]);
  
  const csvContent = [...guildData, headers.join(','), ...rows.map(row => row.join(','))].join('\n');
  
  const blob = new Blob([csvContent], { type: 'text/csv' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `guild-stats-${new Date().toISOString().split('T')[0]}.csv`;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}

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
        <div className="ov-card"><div className="ov-val">{stats.total_agents}</div><Tooltip content={STAT_TOOLTIPS.adventurers}><div className="ov-lbl">Adventurers</div></Tooltip></div>
        <div className="ov-card"><div className="ov-val">{stats.total_sessions}</div><Tooltip content={STAT_TOOLTIPS.total_sessions}><div className="ov-lbl">Sessions</div></Tooltip></div>
        <div className="ov-card"><div className="ov-val">{formatTokens(stats.total_tokens)}</div><Tooltip content={STAT_TOOLTIPS.total_tokens}><div className="ov-lbl">Total Tokens</div></Tooltip></div>
        <div className="ov-card"><div className="ov-val">{formatCost(stats.total_cost)}</div><Tooltip content={STAT_TOOLTIPS.total_cost}><div className="ov-lbl">Total Cost</div></Tooltip></div>
        <div className="ov-card"><div className="ov-val">{stats.completed_quests}</div><Tooltip content={STAT_TOOLTIPS.quests_done}><div className="ov-lbl">Quests Done</div></Tooltip></div>
        <div className="ov-card"><div className="ov-val">{stats.failed_quests}</div><Tooltip content={STAT_TOOLTIPS.quests_failed}><div className="ov-lbl">Quests Failed</div></Tooltip></div>
        <div className="ov-card"><div className="ov-val">{stats.active_quests}</div><Tooltip content={STAT_TOOLTIPS.active}><div className="ov-lbl">Active</div></Tooltip></div>
        <div className="ov-card"><div className="ov-val">{stats.uptime || '—'}</div><Tooltip content={STAT_TOOLTIPS.uptime}><div className="ov-lbl">Uptime</div></Tooltip></div>
      </div>

      <div className="export-controls">
        <h3>📁 Export Guild Data</h3>
        <div className="export-buttons">
          <button 
            className="btn-export btn-json" 
            onClick={() => exportAsJSON(agents, stats)}
            title="Download comprehensive guild and agent data as JSON"
          >
            📋 Export JSON
          </button>
          <button 
            className="btn-export btn-csv" 
            onClick={() => exportAsCSV(agents, stats)}
            title="Download agent statistics as CSV spreadsheet"
          >
            📊 Export CSV
          </button>
        </div>
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
