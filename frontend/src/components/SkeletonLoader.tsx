import './SkeletonLoader.css';

interface SkeletonProps {
  width?: string;
  height?: string;
  borderRadius?: string;
  className?: string;
}

export function Skeleton({ width = '100%', height = '1rem', borderRadius = '4px', className = '' }: SkeletonProps) {
  return (
    <div 
      className={`skeleton ${className}`}
      style={{ width, height, borderRadius }}
    />
  );
}

interface SkeletonAgentCardProps {
  className?: string;
}

export function SkeletonAgentCard({ className = '' }: SkeletonAgentCardProps) {
  return (
    <div className={`agent-card skeleton-card ${className}`}>
      <div className="ac-top">
        <div className="ac-avatar">
          <Skeleton width="40px" height="40px" borderRadius="50%" />
        </div>
        <div className="ac-info">
          <Skeleton width="80px" height="14px" className="skeleton-name" />
          <Skeleton width="60px" height="12px" className="skeleton-class" />
        </div>
        <Skeleton width="8px" height="8px" borderRadius="50%" className="skeleton-status-dot" />
      </div>
      <div className="ac-xp-bar">
        <Skeleton width="100%" height="4px" borderRadius="2px" />
      </div>
      <div className="ac-bottom">
        <Skeleton width="50px" height="12px" />
        <Skeleton width="40px" height="12px" />
      </div>
    </div>
  );
}

interface SkeletonQuestCardProps {
  className?: string;
}

export function SkeletonQuestCard({ className = '' }: SkeletonQuestCardProps) {
  return (
    <div className={`quest-item skeleton-card ${className}`}>
      <div className="qi-header">
        <Skeleton width="70%" height="16px" className="skeleton-quest-name" />
        <Skeleton width="60px" height="20px" borderRadius="10px" className="skeleton-status-badge" />
      </div>
      <Skeleton width="90%" height="12px" className="skeleton-quest-desc" />
      <Skeleton width="50%" height="12px" className="skeleton-quest-desc-2" />
      <div className="qi-footer">
        <Skeleton width="80px" height="12px" />
        <Skeleton width="60px" height="12px" />
      </div>
    </div>
  );
}

interface SkeletonStatsBarProps {
  className?: string;
}

export function SkeletonStatsBar({ className = '' }: SkeletonStatsBarProps) {
  return (
    <div className={`guild-bar skeleton-stats-bar ${className}`}>
      {[...Array(5)].map((_, i) => (
        <div key={i} className="guild-stat skeleton-stat">
          <Skeleton width="24px" height="16px" className="skeleton-stat-val" />
          <Skeleton width="40px" height="12px" className="skeleton-stat-lbl" />
        </div>
      ))}
    </div>
  );
}

export default { Skeleton, SkeletonAgentCard, SkeletonQuestCard, SkeletonStatsBar };