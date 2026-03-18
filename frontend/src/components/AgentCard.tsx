import React, { memo, useMemo, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { classColor, formatTokens, timeAgo } from '../store';
import type { RPGAgent } from '../store';
import './AgentCard.css';

interface AgentCardProps {
  agent: RPGAgent;
  isSelected: boolean;
  onClick: (agent: RPGAgent) => void;
  compact?: boolean;
}

// Mood helper function moved outside component to avoid recreation
const getMoodEmoji = (mood?: string): string => {
  switch (mood) {
    case 'exhausted': return '😫';
    case 'stressed': return '😰';
    case 'neutral': return '😐';
    case 'content': return '😊';
    case 'happy': return '😄';
    default: return '😐';
  }
};

// Status class helper moved outside component
const getStatusClass = (status: string): string => {
  switch (status) {
    case 'working': return 'status-working';
    case 'on_quest': return 'status-quest';
    case 'stuck': return 'status-stuck';
    default: return 'status-idle';
  }
};

// Memoized avatar component for better performance
const AgentAvatar = memo<{ agent: RPGAgent; size?: number }>(({ agent, size = 40 }) => {
  const handleImageError = useCallback((e: React.SyntheticEvent<HTMLImageElement>) => {
    e.currentTarget.style.display = 'none';
    const sibling = e.currentTarget.nextElementSibling as HTMLElement;
    if (sibling) sibling.style.display = 'block';
  }, []);

  return (
    <div className="ac-avatar">
      <img
        src={`/avatars/${agent.name.toLowerCase().replace(/\s+/g, '-')}.png`}
        alt={agent.name}
        onError={handleImageError}
        style={{ width: `${size}px`, height: `${size}px`, borderRadius: '50%' }}
      />
      <span style={{ display: 'none', fontSize: `${size * 0.6}px` }}>
        {agent.avatar_emoji}
      </span>
    </div>
  );
});

AgentAvatar.displayName = 'AgentAvatar';

const AgentCard = memo<AgentCardProps>(({ agent, isSelected, onClick, compact = true }) => {
  const navigate = useNavigate();
  
  // Memoize expensive calculations
  const { classColorValue, xpPercentage, statusClassName } = useMemo(() => ({
    classColorValue: classColor(agent.class),
    xpPercentage: agent.xp_to_next > 0 
      ? (agent.xp / (agent.xp + agent.xp_to_next)) * 100 
      : 100,
    statusClassName: getStatusClass(agent.status)
  }), [agent.class, agent.xp, agent.xp_to_next, agent.status]);

  // Memoize formatted values to avoid recalculation
  const formattedTokens = useMemo(() => formatTokens(agent.total_tokens), [agent.total_tokens]);
  const lastActiveTime = useMemo(() => timeAgo(agent.last_active), [agent.last_active]);
  const moodEmoji = useMemo(() => getMoodEmoji(agent.mood), [agent.mood]);

  // Use useCallback for event handlers to prevent child re-renders
  const handleCardClick = useCallback(() => {
    onClick(agent);
  }, [agent, onClick]);

  const handleDetailClick = useCallback((e: React.MouseEvent) => {
    e.stopPropagation();
    navigate(`/agent/${agent.id}`);
  }, [agent.id, navigate]);

  // Avoid creating new objects in render
  const cardStyle = useMemo(() => ({ '--cc': classColorValue } as React.CSSProperties), [classColorValue]);
  const xpFillStyle = useMemo(() => ({ 
    width: `${xpPercentage}%`, 
    background: classColorValue 
  }), [xpPercentage, classColorValue]);

  return (
    <div
      className={`agent-card ${statusClassName} ${isSelected ? 'selected' : ''} ${compact ? 'compact' : ''}`}
      style={cardStyle}
      onClick={handleCardClick}
    >
      <div className="ac-top">
        <AgentAvatar agent={agent} size={compact ? 40 : 60} />
        <div className="ac-info">
          <div className="ac-name" style={{ color: classColorValue }}>
            {agent.name}
          </div>
          <div className="ac-class">
            {agent.class} · Lv {agent.level}
          </div>
        </div>
        <div 
          className={`ac-status-dot ${statusClassName}`} 
          title={agent.status} 
        />
      </div>
      
      <div className="ac-xp-bar">
        <div className="ac-xp-fill" style={xpFillStyle} />
      </div>
      
      <div className="ac-bottom">
        <span className="ac-tokens">🔮 {formattedTokens}</span>
        <span className="ac-gold">🪙 {agent.gold || 0}</span>
        <span 
          className="ac-mood" 
          title={`Mood: ${agent.mood || 'neutral'} (${agent.mood_score || 75}/100)`}
        >
          {moodEmoji} {agent.mood || 'neutral'}
        </span>
        <span className="ac-time">{lastActiveTime}</span>
      </div>
      
      <span className="ac-chat-hint">💬 Click to chat</span>
      <button
        className="ac-detail-btn"
        onClick={handleDetailClick}
        title="Character Sheet"
      >
        📋
      </button>
    </div>
  );
});

AgentCard.displayName = 'AgentCard';

export default AgentCard;