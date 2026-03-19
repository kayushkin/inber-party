import React from 'react';
import './AchievementToast.css';

export interface AchievementToastProps {
  id: string;
  name: string;
  description: string;
  icon: string;
  timestamp: number;
}

interface AchievementToastComponentProps {
  achievement: AchievementToastProps;
  onComplete: () => void;
}

export const AchievementToast: React.FC<AchievementToastComponentProps> = ({ achievement, onComplete }) => {
  React.useEffect(() => {
    const timer = setTimeout(onComplete, 5000); // Show for 5 seconds
    return () => clearTimeout(timer);
  }, [onComplete]);

  return (
    <div className="achievement-toast">
      <div className="achievement-toast-shine"></div>
      <div className="achievement-toast-content">
        <div className="achievement-toast-header">
          <div className="achievement-toast-icon">{achievement.icon}</div>
          <div className="achievement-toast-badge">
            <span className="achievement-badge-text">🏆 ACHIEVEMENT UNLOCKED!</span>
          </div>
        </div>
        <div className="achievement-toast-body">
          <div className="achievement-toast-title">{achievement.name}</div>
          <div className="achievement-toast-description">{achievement.description}</div>
        </div>
        <div className="achievement-toast-particles">
          {Array.from({ length: 12 }).map((_, i) => (
            <div 
              key={i}
              className="achievement-particle" 
              style={{ 
                '--particle-delay': `${i * 0.1}s`,
                '--particle-angle': `${i * 30}deg`
              } as React.CSSProperties}
            />
          ))}
        </div>
      </div>
    </div>
  );
};

// Container component for managing multiple toasts
interface AchievementToastContainerProps {
  achievements: AchievementToastProps[];
  onRemove: (timestamp: number) => void;
}

export const AchievementToastContainer: React.FC<AchievementToastContainerProps> = ({ achievements, onRemove }) => {
  return (
    <div className="achievement-toast-container">
      {achievements.map((achievement) => (
        <AchievementToast
          key={achievement.id}
          achievement={achievement}
          onComplete={() => onRemove(achievement.timestamp)}
        />
      ))}
    </div>
  );
};