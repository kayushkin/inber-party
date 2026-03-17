import { useState } from 'react';
import { useStore, formatTokens, formatCost, timeAgo, getDifficultyStars, calculateQuestDifficulty, getDifficultyName } from '../store';
import QuestCompletionAnimation from '../components/QuestCompletionAnimation';
import { SkeletonQuestCard } from '../components/SkeletonLoader';
import './QuestBoard.css';

export default function QuestBoard() {
  const quests = useStore((s) => s.quests);
  const questCompletionTriggers = useStore((s) => s.questCompletionTriggers);
  const isLoadingQuests = useStore((s) => s.isLoadingQuests);
  const hasInitialLoad = useStore((s) => s.hasInitialLoad);
  const [filter, setFilter] = useState<string>('all');

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
        <h1>📜 Quest Board</h1>
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
              <div className="qm"><span className="qm-l">Tokens</span><span className="qm-v tokens">{formatTokens(q.tokens_used)}</span></div>
              <div className="qm"><span className="qm-l">Cost</span><span className="qm-v cost">{formatCost(q.cost)}</span></div>
              <div className="qm"><span className="qm-l">Turns</span><span className="qm-v">{q.turns}</span></div>
            </div>
            <div className="quest-card-footer">
              <span className="quest-reward">⭐ +{q.xp_reward} XP</span>
              <span className="quest-time">{timeAgo(q.started_at || q.created_at)}</span>
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
    </div>
  );
}
