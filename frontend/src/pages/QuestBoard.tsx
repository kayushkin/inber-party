import { useEffect, useState } from 'react';
import { useStore, api } from '../store';
import './QuestBoard.css';

export default function QuestBoard() {
  const tasks = useStore((state) => state.tasks);
  const inberQuests = useStore((state) => state.inberQuests);
  const startPolling = useStore((state) => state.startPolling);
  const stopPolling = useStore((state) => state.stopPolling);

  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [expandedQuest, setExpandedQuest] = useState<number | null>(null);

  useEffect(() => {
    // Start auto-refresh polling (every 10s)
    startPolling(10000);
    return () => stopPolling();
  }, [startPolling, stopPolling]);

  // Use inber quests if available, otherwise fall back to tasks
  const isInber = api.isInberMode();
  const questsToShow = isInber ? inberQuests : [];

  const filteredQuests = statusFilter === 'all'
    ? questsToShow
    : questsToShow.filter(q => q.status === statusFilter);

  const statusCounts = {
    all: questsToShow.length,
    completed: questsToShow.filter(q => q.status === 'completed').length,
    in_progress: questsToShow.filter(q => q.status === 'in_progress').length,
    failed: questsToShow.filter(q => q.status === 'failed').length,
    available: questsToShow.filter(q => q.status === 'available').length,
  };

  const getDifficultyStars = (difficulty: number) => '⭐'.repeat(Math.min(difficulty, 5));

  const getDifficultyColor = (difficulty: number) => {
    if (difficulty >= 4) return '#f87171';
    if (difficulty >= 3) return '#fbbf24';
    if (difficulty >= 2) return '#60a5fa';
    return '#4ade80';
  };

  const formatTokens = (tokens: number) => {
    if (tokens >= 1000000) return `${(tokens / 1000000).toFixed(1)}M`;
    if (tokens >= 1000) return `${(tokens / 1000).toFixed(1)}k`;
    return String(tokens);
  };

  const formatCost = (cost: number) => {
    if (cost <= 0) return '—';
    if (cost < 0.01) return `$${cost.toFixed(4)}`;
    return `$${cost.toFixed(2)}`;
  };

  const formatTime = (dateStr?: string) => {
    if (!dateStr) return '—';
    try {
      const d = new Date(dateStr);
      const now = new Date();
      const diffMs = now.getTime() - d.getTime();
      const diffMin = Math.floor(diffMs / 60000);
      if (diffMin < 1) return 'just now';
      if (diffMin < 60) return `${diffMin}m ago`;
      const diffHr = Math.floor(diffMin / 60);
      if (diffHr < 24) return `${diffHr}h ago`;
      const diffDay = Math.floor(diffHr / 24);
      return `${diffDay}d ago`;
    } catch {
      return dateStr;
    }
  };

  // Fallback: render legacy tasks if not in inber mode
  if (!isInber) {
    return (
      <div className="quest-board">
        <div className="quest-board-header">
          <h1>Quest Board</h1>
        </div>
        <div className="quests-grid">
          {tasks.map((task) => (
            <div key={task.id} className={`quest-card status-${task.status}`}>
              <div className="quest-card-header">
                <h3 className="quest-card-name">{task.name}</h3>
                <div className="quest-difficulty" style={{ borderColor: '#888' }}>
                  {task.difficulty}
                </div>
              </div>
              <p className="quest-card-description">{task.description}</p>
              <div className="quest-card-footer">
                <div className="quest-reward">
                  <span className="reward-icon">⭐</span>
                  +{task.xp_reward} XP
                </div>
              </div>
            </div>
          ))}
          {tasks.length === 0 && (
            <div className="no-quests"><p>No quests available.</p></div>
          )}
        </div>
      </div>
    );
  }

  return (
    <div className="quest-board">
      <div className="quest-board-header">
        <h1>Quest Board</h1>
        <div className="quest-filters">
          {(['all', 'in_progress', 'completed', 'failed'] as const).map(status => (
            <button
              key={status}
              className={`filter-btn ${statusFilter === status ? 'active' : ''} filter-${status}`}
              onClick={() => setStatusFilter(status)}
            >
              {status === 'all' ? 'All' : status === 'in_progress' ? 'Active' : status.charAt(0).toUpperCase() + status.slice(1)}
              <span className="filter-count">{statusCounts[status]}</span>
            </button>
          ))}
        </div>
      </div>

      <div className="quests-grid">
        {filteredQuests.map((quest) => (
          <div
            key={quest.id}
            className={`quest-card status-${quest.status} ${expandedQuest === quest.id ? 'expanded' : ''}`}
            onClick={() => setExpandedQuest(expandedQuest === quest.id ? null : quest.id)}
          >
            <div className="quest-card-header">
              <h3 className="quest-card-name">{quest.name}</h3>
              <div
                className="quest-difficulty"
                style={{ borderColor: getDifficultyColor(quest.difficulty) }}
              >
                {getDifficultyStars(quest.difficulty)}
              </div>
            </div>

            {quest.description && (
              <p className="quest-card-description">
                {expandedQuest === quest.id ? quest.description : quest.description.slice(0, 100) + (quest.description.length > 100 ? '...' : '')}
              </p>
            )}

            <div className="quest-meta-grid">
              <div className="quest-meta-item">
                <span className="meta-label">Agent</span>
                <span className="meta-value agent-name">{quest.assigned_agent_name || '—'}</span>
              </div>
              <div className="quest-meta-item">
                <span className="meta-label">Tokens</span>
                <span className="meta-value tokens">{formatTokens(quest.tokens_used)}</span>
              </div>
              <div className="quest-meta-item">
                <span className="meta-label">Cost</span>
                <span className="meta-value cost">{formatCost(quest.cost)}</span>
              </div>
              <div className="quest-meta-item">
                <span className="meta-label">Turns</span>
                <span className="meta-value turns">{quest.turns}</span>
              </div>
            </div>

            {(quest.children ?? 0) > 0 && (
              <div className="quest-subquests">
                ⚔️ {quest.children} sub-quest{quest.children !== 1 ? 's' : ''} spawned
              </div>
            )}

            <div className="quest-card-footer">
              <div className="quest-reward">
                <span className="reward-icon">⭐</span>
                +{quest.xp_reward} XP
              </div>
              <div className="quest-time">{formatTime(quest.started_at || quest.created_at)}</div>
            </div>

            {quest.status === 'in_progress' && (
              <div className="quest-progress-bar">
                <div className="quest-progress-fill" style={{ width: `${quest.progress}%` }} />
              </div>
            )}

            {quest.status === 'failed' && quest.error_text && expandedQuest === quest.id && (
              <div className="quest-error">
                <span className="error-label">Error:</span> {quest.error_text}
              </div>
            )}
          </div>
        ))}

        {filteredQuests.length === 0 && (
          <div className="no-quests">
            <p>No quests match this filter.</p>
          </div>
        )}
      </div>
    </div>
  );
}
