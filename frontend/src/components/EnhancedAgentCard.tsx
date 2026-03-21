import React, { memo, useMemo, useCallback, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { classColor, formatTokens, timeAgo } from '../store';
import type { RPGAgent } from '../store';
import { 
  InteractiveHover, 
  Tooltip, 
  ProgressBar, 
  StateIndicator, 
  FadeIn
} from './MicroInteractions';
import { useInViewAnimation } from '../hooks/useInViewAnimation';
import EnhancedButton from './EnhancedButton';
import './AgentCard.css';
import './MicroInteractions.css';

interface EnhancedAgentCardProps {
  agent: RPGAgent;
  isSelected: boolean;
  onClick: (agent: RPGAgent) => void;
  compact?: boolean;
  animateIn?: boolean;
}

// Mood helper function with enhanced emojis and descriptions
const getMoodData = (mood?: string): { emoji: string; description: string; state: 'success' | 'error' | 'warning' | 'info' | 'default' } => {
  switch (mood) {
    case 'exhausted': 
      return { emoji: '😫', description: 'Needs rest urgently', state: 'error' };
    case 'stressed': 
      return { emoji: '😰', description: 'Under pressure', state: 'warning' };
    case 'neutral': 
      return { emoji: '😐', description: 'Feeling balanced', state: 'default' };
    case 'content': 
      return { emoji: '😊', description: 'Comfortable and ready', state: 'info' };
    case 'happy': 
      return { emoji: '😄', description: 'Energized and motivated', state: 'success' };
    default: 
      return { emoji: '😐', description: 'Feeling balanced', state: 'default' };
  }
};

// Enhanced status information with better descriptions
const getStatusData = (status: string): { 
  class: string; 
  emoji: string; 
  description: string; 
  pulse: boolean;
} => {
  switch (status) {
    case 'working':
      return { 
        class: 'status-working', 
        emoji: '⚡', 
        description: 'Currently working on a task', 
        pulse: true 
      };
    case 'on_quest':
      return { 
        class: 'status-quest', 
        emoji: '⚔️', 
        description: 'On an active quest', 
        pulse: true 
      };
    case 'stuck':
      return { 
        class: 'status-stuck', 
        emoji: '🚫', 
        description: 'Needs assistance or guidance', 
        pulse: true 
      };
    default:
      return { 
        class: 'status-idle', 
        emoji: '💤', 
        description: 'Available and waiting', 
        pulse: false 
      };
  }
};

// Enhanced avatar component with loading state and fallback
const EnhancedAgentAvatar = memo<{ 
  agent: RPGAgent; 
  size?: number; 
}>(({ agent, size = 40 }) => {
  const [imageLoaded, setImageLoaded] = useState(false);
  const [imageError, setImageError] = useState(false);

  const handleImageLoad = useCallback(() => {
    setImageLoaded(true);
  }, []);

  const handleImageError = useCallback(() => {
    setImageError(true);
    setImageLoaded(true);
  }, []);

  return (
    <div className="ac-avatar">
      {!imageLoaded && (
        <div 
          className="skeleton"
          style={{ 
            width: `${size}px`, 
            height: `${size}px`, 
            borderRadius: '50%' 
          }}
        />
      )}
      
      {!imageError ? (
        <img
          src={`/avatars/${agent.name.toLowerCase().replace(/\s+/g, '-')}.png`}
          alt={agent.name}
          onLoad={handleImageLoad}
          onError={handleImageError}
          style={{ 
            width: `${size}px`, 
            height: `${size}px`, 
            borderRadius: '50%',
            display: imageLoaded && !imageError ? 'block' : 'none'
          }}
        />
      ) : (
        <div
          style={{ 
            width: `${size}px`, 
            height: `${size}px`, 
            borderRadius: '50%',
            background: 'var(--bg-light)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontSize: `${size * 0.6}px`,
            border: '2px solid var(--border-color)'
          }}
        >
          {agent.avatar_emoji}
        </div>
      )}
    </div>
  );
});

EnhancedAgentAvatar.displayName = 'EnhancedAgentAvatar';

const EnhancedAgentCard = memo<EnhancedAgentCardProps>(({ 
  agent, 
  isSelected, 
  onClick, 
  compact = true, 
  animateIn = true 
}) => {
  const navigate = useNavigate();
  const { ref, inView } = useInViewAnimation(0.1);
  
  // Enhanced calculations with more data
  const { 
    classColorValue, 
    xpPercentage, 
    statusData, 
    moodData,
    isHighPerformer,
    needsAttention
  } = useMemo(() => {
    const classColorValue = classColor(agent.class);
    const xpPercentage = agent.xp_to_next > 0 
      ? (agent.xp / (agent.xp + agent.xp_to_next)) * 100 
      : 100;
    const statusData = getStatusData(agent.status);
    const moodData = getMoodData(agent.mood);
    const isLowHealth = (agent.mood_score || 75) < 30;
    const isHighPerformer = (agent.mood_score || 75) > 90 && agent.level > 5;
    const needsAttention = agent.status === 'stuck' || isLowHealth;
    
    return {
      classColorValue,
      xpPercentage,
      statusData,
      moodData,
      isLowHealth,
      isHighPerformer,
      needsAttention
    };
  }, [agent]);

  // Memoize formatted values
  const formattedTokens = useMemo(() => formatTokens(agent.total_tokens), [agent.total_tokens]);
  const lastActiveTime = useMemo(() => timeAgo(agent.last_active), [agent.last_active]);

  // Enhanced event handlers
  const handleCardClick = useCallback(() => {
    onClick(agent);
  }, [agent, onClick]);

  const handleDetailClick = useCallback((e: React.MouseEvent) => {
    e.stopPropagation();
    navigate(`/agent/${agent.id}`);
  }, [agent.id, navigate]);

  // Card styling with enhanced states
  const cardClasses = useMemo(() => [
    'agent-card',
    statusData.class,
    isSelected ? 'selected' : '',
    compact ? 'compact' : '',
    needsAttention ? 'pulse-attention' : '',
    isHighPerformer ? 'glow-on-hover' : ''
  ].filter(Boolean).join(' '), [statusData.class, isSelected, compact, needsAttention, isHighPerformer]);

  const cardStyle = useMemo(() => ({ 
    '--cc': classColorValue 
  } as React.CSSProperties), [classColorValue]);

  const cardContent = (
    <InteractiveHover 
      scale={!needsAttention} 
      glow={isHighPerformer}
      className={cardClasses}
    >
      <div
        style={cardStyle}
        onClick={handleCardClick}
        role="button"
        tabIndex={0}
        aria-label={`${agent.name}, ${agent.class}, Level ${agent.level}, ${statusData.description}`}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            handleCardClick();
          }
        }}
      >
        <div className="ac-top">
          <EnhancedAgentAvatar agent={agent} size={compact ? 40 : 60} />
          
          <div className="ac-info">
            <div className="ac-name" style={{ color: classColorValue }}>
              {agent.name}
              {isHighPerformer && (
                <Tooltip content="High performer! This agent is excelling.">
                  <span style={{ marginLeft: '4px' }}>⭐</span>
                </Tooltip>
              )}
            </div>
            <div className="ac-class">
              {agent.class} · Lv {agent.level}
              {needsAttention && (
                <Tooltip content="This agent needs attention">
                  <span style={{ marginLeft: '4px', color: '#f87171' }}>⚠️</span>
                </Tooltip>
              )}
            </div>
          </div>
          
          <Tooltip content={statusData.description}>
            <div className={`ac-status-dot ${statusData.class} ${statusData.pulse ? 'pulse-soft' : ''}`}>
              <span style={{ fontSize: '12px' }}>{statusData.emoji}</span>
            </div>
          </Tooltip>
        </div>
        
        <Tooltip content={`XP: ${agent.xp} / ${agent.xp + agent.xp_to_next} (${Math.round(xpPercentage)}%)`}>
          <div className="ac-xp-bar">
            <ProgressBar progress={xpPercentage} animated={true} />
          </div>
        </Tooltip>
        
        <div className="ac-bottom">
          <Tooltip content={`Total tokens used: ${agent.total_tokens.toLocaleString()}`}>
            <span className="ac-tokens">🔮 {formattedTokens}</span>
          </Tooltip>
          
          <Tooltip content={`Gold coins earned`}>
            <span className="ac-gold">🪙 {agent.gold || 0}</span>
          </Tooltip>
          
          <StateIndicator state={moodData.state} animate={false}>
            <Tooltip content={`${moodData.description} (${agent.mood_score || 75}/100)`}>
              <span className="ac-mood">
                {moodData.emoji} {agent.mood || 'neutral'}
              </span>
            </Tooltip>
          </StateIndicator>
          
          <Tooltip content={`Last active: ${agent.last_active}`}>
            <span className="ac-time">{lastActiveTime}</span>
          </Tooltip>
        </div>
        
        <span className="ac-chat-hint">💬 Click to start conversation</span>
        
        <Tooltip content="View detailed character sheet">
          <EnhancedButton
            variant="ghost"
            size="icon"
            className="ac-detail-btn"
            onClick={handleDetailClick}
            icon="📋"
            aria-label="View character sheet"
          />
        </Tooltip>
      </div>
    </InteractiveHover>
  );

  if (animateIn && inView) {
    return (
      <div ref={ref}>
        <FadeIn>
          {cardContent}
        </FadeIn>
      </div>
    );
  }

  return (
    <div ref={ref}>
      {cardContent}
    </div>
  );
});

EnhancedAgentCard.displayName = 'EnhancedAgentCard';

export default EnhancedAgentCard;