import { useState } from 'react';
import { useStore, formatTokens, formatCost, timeAgo, getDifficultyStars, calculateQuestDifficulty, getDifficultyName } from '../store';
import QuestCompletionAnimation from '../components/QuestCompletionAnimation';
import QuestCreationForm from '../components/QuestCreationForm';
import { SkeletonQuestCard } from '../components/SkeletonLoader';
import './QuestBoard.css';

// Helper function to calculate duration between two timestamps
function calculateDuration(startTime: string, endTime: string): string {
  const start = new Date(startTime);
  const end = new Date(endTime);
  const diffMs = end.getTime() - start.getTime();
  
  const minutes = Math.floor(diffMs / (1000 * 60));
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);
  
  if (days > 0) {
    return `${days}d ${hours % 24}h`;
  } else if (hours > 0) {
    return `${hours}h ${minutes % 60}m`;
  } else if (minutes > 0) {
    return `${minutes}m`;
  } else {
    return '<1m';
  }
}

export default function QuestBoard() {
  const quests = useStore((s) => s.quests);
  const questCompletionTriggers = useStore((s) => s.questCompletionTriggers);
  const isLoadingQuests = useStore((s) => s.isLoadingQuests);
  const hasInitialLoad = useStore((s) => s.hasInitialLoad);
  const fetchAll = useStore((s) => s.fetchAll);
  const [filter, setFilter] = useState<string>('all');
  const [showCreateForm, setShowCreateForm] = useState(false);

  const filtered = filter === 'all' ? quests : quests.filter((q) => q.status === filter);

  const counts = {
    all: quests.length,
    in_progress: quests.filter((q) => q.status === 'in_progress').length,
    completed: quests.filter((q) => q.status === 'completed').length,
    failed: quests.filter((q) => q.status === 'failed').length,
  };

  return (
    <div className="quest-board">
      <div className="quest-board-header">
        <div className="quest-board-title">
          <h1>📜 Quest Board</h1>
          <button 
            className="create-quest-btn"
            onClick={() => setShowCreateForm(true)}
            title="Create a new quest"
          >
            ➕ Create Quest
          </button>
        </div>
        <div className="quest-filters">
          {(['all', 'in_progress', 'completed', 'failed'] as const).map((s) => (
            <button
              key={s}
              className={`filter-btn ${filter === s ? 'active' : ''} filter-${s}`}
              onClick={() => setFilter(s)}
            >
              {s === 'all' ? 'All' : s === 'in_progress' ? 'Active' : s.charAt(0).toUpperCase() + s.slice(1)}
              <span className="filter-count">{counts[s]}</span>
            </button>
          ))}
        </div>
      </div>

      <div className="quests-grid">
        {isLoadingQuests ? (
          // Show skeleton loaders while quests are loading
          [...Array(8)].map((_, i) => (
            <SkeletonQuestCard key={`skeleton-quest-${i}`} />
          ))
        ) : (
          filtered.map((q) => (
          <div key={q.id} className={`quest-card status-${q.status}`}>
            <div className="quest-card-header">
              <h3 className="quest-card-name">{q.name}</h3>
              <span 
                className="quest-difficulty" 
                title={`${getDifficultyName(q.tokens_used)} (${calculateQuestDifficulty(q.tokens_used)}/5 stars) - ${formatTokens(q.tokens_used)} tokens`}
              >
                {getDifficultyStars(q.tokens_used)}
              </span>
            </div>
            {q.description && (
              <p className="quest-card-desc">{q.description.slice(0, 120)}{q.description.length > 120 ? '...' : ''}</p>
            )}
            <div className="quest-meta-grid">
              <div className="qm"><span className="qm-l">Agent</span><span className="qm-v agent">{q.assigned_agent_name || '—'}</span></div>
              <div className="qm">
                <span className="qm-l">Tokens</span>
                <span className="qm-v tokens" title={`${q.tokens_used} tokens consumed`}>
                  {formatTokens(q.tokens_used)}
                  {q.tokens_used > 10000 && <span className="token-warning">⚠️</span>}
                </span>
              </div>
              <div className="qm">
                <span className="qm-l">Cost</span>
                <span className="qm-v cost" title={`$${q.cost.toFixed(4)} total cost`}>
                  {formatCost(q.cost)}
                  {q.cost > 1.0 && <span className="cost-warning">💰</span>}
                </span>
              </div>
              <div className="qm">
                <span className="qm-l">Turns</span>
                <span className="qm-v" title={`${q.turns} conversation turns`}>
                  {q.turns}
                  {q.turns > 20 && <span className="turn-warning">🔄</span>}
                </span>
              </div>
              {q.children && q.children > 0 && (
                <div className="qm">
                  <span className="qm-l">Sub-quests</span>
                  <span className="qm-v sub-quests" title={`${q.children} sub-quests spawned`}>
                    {q.children} 🔗
                  </span>
                </div>
              )}
            </div>
            
            {/* Enhanced quest details section */}
            {(q.status === 'failed' && q.error_text) && (
              <div className="quest-error-section">
                <span className="error-label">❌ Error:</span>
                <span className="error-text" title={q.error_text}>
                  {q.error_text.length > 60 ? q.error_text.slice(0, 60) + '...' : q.error_text}
                </span>
              </div>
            )}
            
            {q.status === 'completed' && q.completed_at && (
              <div className="quest-timing-section">
                <span className="timing-label">⏱️ Duration:</span>
                <span className="timing-value">
                  {calculateDuration(q.started_at || q.created_at, q.completed_at)}
                </span>
              </div>
            )}

            <div className="quest-card-footer">
              <span className="quest-reward">⭐ +{q.xp_reward} XP</span>
              <span className="quest-reward">🪙 +{q.gold_reward || 0} Gold</span>
              <span className="quest-time">{timeAgo(q.started_at || q.created_at)}</span>
              {q.status === 'in_progress' && q.started_at && (
                <span className="quest-runtime" title="Time since started">
                  🕐 {calculateDuration(q.started_at, new Date().toISOString())}
                </span>
              )}
            </div>
            {q.status === 'in_progress' && (
              <div className="quest-progress-bar">
                <div className="quest-progress-fill" style={{ width: `${q.progress}%` }} />
              </div>
            )}
            </div>
          ))
        )}
        {!isLoadingQuests && filtered.length === 0 && hasInitialLoad && (
          <div className="no-quests">No quests match this filter.</div>
        )}
      </div>

      {/* Quest completion animations */}
      {Object.entries(questCompletionTriggers).map(([questId, data]) => (
        <QuestCompletionAnimation
          key={questId}
          trigger={data.trigger}
          questName={data.questName}
          xpReward={data.xpReward}
          agentName={data.agentName}
        />
      ))}

      {/* Quest creation form */}
      <QuestCreationForm
        isOpen={showCreateForm}
        onClose={() => setShowCreateForm(false)}
        onQuestCreated={() => {
          fetchAll(); // Refresh the quest list
        }}
      />
    </div>
  );
}
