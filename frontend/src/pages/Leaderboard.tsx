import { useState, useEffect } from 'react';
import { SkeletonAgentCard } from '../components/SkeletonLoader';
import './Leaderboard.css';

interface LeaderboardEntry {
  agent: {
    id: number;
    name: string;
    title: string;
    class: string;
    level: number;
    xp: number;
    energy: number;
    status: string;
    avatar_emoji: string;
    created_at: string;
  };
  quests_completed: number;
  quests_attempted: number;
  avg_quest_xp: number;
  total_quest_xp: number;
  efficiency_score: number;
  rank: number;
}

type SortBy = 'xp' | 'quests' | 'efficiency';

export default function Leaderboard() {
  const [leaderboard, setLeaderboard] = useState<LeaderboardEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [sortBy, setSortBy] = useState<SortBy>('xp');
  const [error, setError] = useState<string | null>(null);

  const fetchLeaderboard = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/leaderboard');
      if (!response.ok) throw new Error('Failed to fetch leaderboard');
      const data = await response.json();
      setLeaderboard(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
      console.error('Error fetching leaderboard:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchLeaderboard();
  }, []);

  // Sort the leaderboard based on selected criteria
  const sortedLeaderboard = [...leaderboard].sort((a, b) => {
    switch (sortBy) {
      case 'xp':
        return b.agent.xp - a.agent.xp;
      case 'quests':
        return b.quests_completed - a.quests_completed;
      case 'efficiency':
        return b.efficiency_score - a.efficiency_score;
      default:
        return a.rank - b.rank;
    }
  });

  const classColor = (agentClass: string) => {
    const colors: Record<string, string> = {
      'mage': '#9d4edd',
      'warrior': '#f77f00',
      'rogue': '#06ffa5',
      'archer': '#3a86ff',
      'healer': '#06ffa5',
      'paladin': '#ffd60a',
      'scribe': '#7209b7'
    };
    return colors[agentClass.toLowerCase()] || '#6c757d';
  };

  const getRankIcon = (rank: number) => {
    switch (rank) {
      case 1: return '🥇';
      case 2: return '🥈';
      case 3: return '🥉';
      default: return `#${rank}`;
    }
  };

  const getStatusEmoji = (status: string) => {
    switch (status) {
      case 'working': return '⚙️';
      case 'on_quest': return '⚔️';
      case 'stuck': return '❗';
      case 'idle': return '💤';
      default: return '🟢';
    }
  };

  const formatNumber = (num: number) => {
    if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`;
    if (num >= 1000) return `${(num / 1000).toFixed(1)}k`;
    return num.toString();
  };

  if (loading) {
    return (
      <div className="leaderboard">
        <div className="leaderboard-header">
          <h1>🏆 Hall of Champions</h1>
          <p>The most skilled adventurers in the realm</p>
        </div>
        <div className="leaderboard-skeleton">
          {Array.from({ length: 5 }).map((_, i) => (
            <SkeletonAgentCard key={i} />
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="leaderboard">
        <div className="leaderboard-header">
          <h1>🏆 Hall of Champions</h1>
          <div className="error-state">
            <p>⚠️ Failed to load leaderboard: {error}</p>
            <button onClick={fetchLeaderboard} className="retry-btn">
              🔄 Retry
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="leaderboard">
      <div className="leaderboard-header">
        <h1>🏆 Hall of Champions</h1>
        <p>The most skilled adventurers in the realm</p>
        
        <div className="sort-controls">
          <label>Sort by:</label>
          <div className="sort-buttons">
            <button 
              className={`sort-btn ${sortBy === 'xp' ? 'active' : ''}`}
              onClick={() => setSortBy('xp')}
            >
              ⭐ Experience
            </button>
            <button 
              className={`sort-btn ${sortBy === 'quests' ? 'active' : ''}`}
              onClick={() => setSortBy('quests')}
            >
              🎯 Quests
            </button>
            <button 
              className={`sort-btn ${sortBy === 'efficiency' ? 'active' : ''}`}
              onClick={() => setSortBy('efficiency')}
            >
              📊 Efficiency
            </button>
          </div>
        </div>
      </div>

      <div className="leaderboard-list">
        {sortedLeaderboard.length === 0 ? (
          <div className="empty-state">
            <p>🏜️ No adventurers found. Create some agents to populate the leaderboard!</p>
          </div>
        ) : (
          sortedLeaderboard.map((entry, index) => {
            const displayRank = sortBy === 'xp' ? entry.rank : index + 1;
            return (
              <div key={entry.agent.id} className={`leaderboard-entry ${displayRank <= 3 ? 'podium' : ''}`}>
                <div className="rank-section">
                  <span className="rank">{getRankIcon(displayRank)}</span>
                </div>
                
                <div className="agent-section">
                  <div className="agent-avatar">
                    {entry.agent.avatar_emoji || '🤖'}
                  </div>
                  <div className="agent-info">
                    <h3 className="agent-name">{entry.agent.name}</h3>
                    <p className="agent-title">
                      <span 
                        className="agent-class"
                        style={{ color: classColor(entry.agent.class) }}
                      >
                        {entry.agent.class}
                      </span>
                      {entry.agent.title && (
                        <span className="agent-title-text"> • {entry.agent.title}</span>
                      )}
                    </p>
                  </div>
                </div>

                <div className="stats-section">
                  <div className="stat-group">
                    <div className="stat-item">
                      <span className="stat-label">Level</span>
                      <span className="stat-value">{entry.agent.level}</span>
                    </div>
                    <div className="stat-item">
                      <span className="stat-label">XP</span>
                      <span className="stat-value">{formatNumber(entry.agent.xp)}</span>
                    </div>
                  </div>

                  <div className="stat-group">
                    <div className="stat-item">
                      <span className="stat-label">Quests</span>
                      <span className="stat-value">
                        {entry.quests_completed}
                        {entry.quests_attempted > 0 && (
                          <span className="stat-secondary">/{entry.quests_attempted}</span>
                        )}
                      </span>
                    </div>
                    <div className="stat-item">
                      <span className="stat-label">Efficiency</span>
                      <span className="stat-value">
                        {entry.efficiency_score.toFixed(0)}%
                      </span>
                    </div>
                  </div>
                </div>

                <div className="status-section">
                  <span 
                    className={`status-indicator ${entry.agent.status}`}
                    title={entry.agent.status}
                  >
                    {getStatusEmoji(entry.agent.status)}
                  </span>
                  <div className="energy-bar">
                    <div 
                      className="energy-fill" 
                      style={{ width: `${entry.agent.energy}%` }}
                    ></div>
                  </div>
                </div>
              </div>
            );
          })
        )}
      </div>

      {sortedLeaderboard.length > 0 && (
        <div className="leaderboard-footer">
          <p>📈 Rankings updated in real-time • Total champions: {sortedLeaderboard.length}</p>
        </div>
      )}
    </div>
  );
}