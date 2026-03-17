import { useEffect, useState } from 'react';
import './QuestCompletionAnimation.css';

interface QuestCompletionAnimationProps {
  trigger: number; // timestamp to trigger animation
  questName: string;
  xpReward: number;
  agentName: string;
}

export default function QuestCompletionAnimation({ trigger, questName, xpReward, agentName }: QuestCompletionAnimationProps) {
  const [isAnimating, setIsAnimating] = useState(false);
  const [lastTrigger, setLastTrigger] = useState(0);

  useEffect(() => {
    if (trigger && trigger !== lastTrigger) {
      setIsAnimating(true);
      setLastTrigger(trigger);
      
      // Reset animation after duration
      const timer = setTimeout(() => setIsAnimating(false), 4500);
      return () => clearTimeout(timer);
    }
  }, [trigger, lastTrigger]);

  if (!isAnimating) return null;

  return (
    <div className="quest-completion-overlay">
      {/* Background dimming */}
      <div className="quest-completion-backdrop" />
      
      {/* Floating particles */}
      <div className="quest-completion-particles">
        {Array.from({ length: 8 }, (_, i) => (
          <div 
            key={i}
            className="completion-particle" 
            style={{ 
              '--delay': `${i * 0.2}s`,
              '--offset': `${(i * 45)}deg`,
            } as React.CSSProperties}
          />
        ))}
      </div>
      
      {/* Main scroll element */}
      <div className="quest-scroll">
        {/* Scroll top seal/rod */}
        <div className="scroll-rod scroll-rod-top" />
        
        {/* Scroll paper/parchment */}
        <div className="scroll-paper">
          {/* Decorative scroll border */}
          <div className="scroll-border" />
          
          {/* Quest completion content */}
          <div className="scroll-content">
            <div className="completion-header">
              <div className="completion-icon">📜</div>
              <div className="completion-title">QUEST COMPLETED!</div>
            </div>
            
            <div className="quest-details">
              <div className="quest-name">"{questName}"</div>
              <div className="quest-agent">By {agentName}</div>
            </div>
            
            <div className="reward-section">
              <div className="reward-label">Reward Earned:</div>
              <div className="reward-value">⭐ +{xpReward} XP</div>
            </div>
            
            <div className="completion-stamp">✓ COMPLETE</div>
          </div>
        </div>
        
        {/* Scroll bottom rod */}
        <div className="scroll-rod scroll-rod-bottom" />
      </div>
      
      {/* Celebration sparkles */}
      <div className="celebration-sparkles">
        <div className="sparkle sparkle-1">✨</div>
        <div className="sparkle sparkle-2">⭐</div>
        <div className="sparkle sparkle-3">✨</div>
        <div className="sparkle sparkle-4">💫</div>
        <div className="sparkle sparkle-5">⭐</div>
        <div className="sparkle sparkle-6">✨</div>
      </div>
    </div>
  );
}