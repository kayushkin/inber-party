import { useEffect, useState } from 'react';
import './LevelUpAnimation.css';

interface LevelUpAnimationProps {
  trigger: number; // timestamp to trigger animation
  agentName: string;
  level: number;
  color: string;
}

export default function LevelUpAnimation({ trigger, agentName, level, color }: LevelUpAnimationProps) {
  const [isAnimating, setIsAnimating] = useState(false);
  const [lastTrigger, setLastTrigger] = useState(0);

  useEffect(() => {
    if (trigger && trigger !== lastTrigger) {
      // Use a microtask to avoid synchronous setState in effect
      Promise.resolve().then(() => {
        setIsAnimating(true);
        setLastTrigger(trigger);
      });
      
      // Reset animation after duration
      const timer = setTimeout(() => setIsAnimating(false), 3000);
      return () => clearTimeout(timer);
    }
  }, [trigger, lastTrigger]);

  if (!isAnimating) return null;

  return (
    <div className="level-up-overlay">
      {/* Background glow effect */}
      <div className="level-up-glow" style={{ borderColor: color, boxShadow: `0 0 40px ${color}` }} />
      
      {/* Particle burst */}
      <div className="level-up-particles">
        {Array.from({ length: 12 }, (_, i) => (
          <div 
            key={i}
            className="particle" 
            style={{ 
              '--angle': `${(i * 30)}deg`,
              '--color': color,
            } as React.CSSProperties}
          />
        ))}
      </div>
      
      {/* Central burst effect */}
      <div className="level-up-burst" style={{ background: `radial-gradient(circle, ${color}, transparent)` }} />
      
      {/* Level up text */}
      <div className="level-up-text">
        <div className="level-up-title" style={{ color }}>LEVEL UP!</div>
        <div className="level-up-subtitle">{agentName} reached Level {level}</div>
        <div className="level-up-sparkles">✨⭐✨</div>
      </div>
    </div>
  );
}