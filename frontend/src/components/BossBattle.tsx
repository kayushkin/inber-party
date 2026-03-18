import { useEffect, useState } from 'react';
import type { RPGQuest, RPGAgent } from '../store';
import { formatTokens, formatCost, timeAgo, getDifficultyStars } from '../store';
import './BossBattle.css';

interface BossBattleProps {
  quest: RPGQuest;
  assignedAgent?: RPGAgent;
  isActive?: boolean;
}

export default function BossBattle({ quest, assignedAgent, isActive = true }: BossBattleProps) {
  const [healthBarAnimation, setHealthBarAnimation] = useState(false);

  useEffect(() => {
    // Animate health bar on mount
    const timer = setTimeout(() => setHealthBarAnimation(true), 100);
    return () => clearTimeout(timer);
  }, []);

  // Calculate boss "health" based on progress (100% - progress = remaining health)
  const bossHealth = Math.max(0, 100 - quest.progress);
  const healthPercentage = bossHealth;
  
  // Determine boss type based on quest characteristics
  const getBossType = () => {
    if (quest.tokens_used >= 200000) return { name: '🐉 Ancient Dragon', title: 'Legendary Boss' };
    if (quest.tokens_used >= 100000) return { name: '👹 Shadow Lord', title: 'Epic Boss' };
    if (quest.tokens_used >= 50000) return { name: '🗿 Stone Titan', title: 'Elite Boss' };
    if (quest.children && quest.children > 0) return { name: '🕸️ Hive Queen', title: 'Multi-Agent Boss' };
    return { name: '⚔️ War Chief', title: 'Raid Boss' };
  };

  const bossType = getBossType();
  
  // Get health bar color based on remaining health
  const getHealthColor = () => {
    if (healthPercentage > 75) return '#dc3545'; // Red (high health)
    if (healthPercentage > 50) return '#fd7e14'; // Orange (medium health)
    if (healthPercentage > 25) return '#ffc107'; // Yellow (low health)
    return '#28a745'; // Green (very low health/almost defeated)
  };

  const getStatusEffect = () => {
    if (quest.error_text) return '💀 Enraged';
    if (quest.progress > 90) return '🩸 Bloodied';
    if (quest.progress > 75) return '⚔️ Wounded';
    if (quest.progress > 50) return '🛡️ Defending';
    if (quest.progress > 25) return '👁️ Alert';
    return '💪 Full Power';
  };

  return (
    <div className={`boss-battle ${isActive ? 'active' : 'completed'} ${quest.error_text ? 'enraged' : ''}`}>
      {/* Boss Header */}
      <div className="boss-header">
        <div className="boss-info">
          <h3 className="boss-name">{bossType.name}</h3>
          <div className="boss-title">{bossType.title}</div>
          <div className="boss-quest-name">"{quest.name}"</div>
        </div>
        <div className="boss-difficulty">
          <span className="difficulty-stars">{getDifficultyStars(quest.tokens_used)}</span>
          <span className="boss-status">{getStatusEffect()}</span>
        </div>
      </div>

      {/* Boss Health Bar */}
      <div className="boss-health-container">
        <div className="boss-health-label">
          <span>Boss Health</span>
          <span>{Math.round(healthPercentage)}%</span>
        </div>
        <div className="boss-health-bar">
          <div className="boss-health-bg">
            <div 
              className={`boss-health-fill ${healthBarAnimation ? 'animate' : ''}`}
              style={{ 
                width: `${healthPercentage}%`,
                backgroundColor: getHealthColor(),
              }}
            >
              <div className="health-shine"></div>
            </div>
          </div>
          {healthPercentage < 20 && (
            <div className="low-health-effect">⚡</div>
          )}
        </div>
      </div>

      {/* Champion (Assigned Agent) */}
      {assignedAgent && (
        <div className="champion-section">
          <div className="champion-header">
            <span className="champion-label">🏆 Champion</span>
            <span className={`champion-status status-${assignedAgent.status}`}>
              {assignedAgent.status.charAt(0).toUpperCase() + assignedAgent.status.slice(1)}
            </span>
          </div>
          <div className="champion-info">
            <span className="champion-avatar">{assignedAgent.avatar_emoji}</span>
            <div className="champion-details">
              <span className="champion-name">{assignedAgent.name}</span>
              <span className="champion-level">Level {assignedAgent.level} {assignedAgent.class}</span>
            </div>
            <div className="champion-energy-bar">
              <div 
                className="champion-energy-fill" 
                style={{ width: `${(assignedAgent.energy / assignedAgent.max_energy) * 100}%` }}
              />
            </div>
          </div>
        </div>
      )}

      {/* Battle Metrics */}
      <div className="battle-metrics">
        <div className="metric-group">
          <div className="metric">
            <span className="metric-icon">⚔️</span>
            <span className="metric-label">Turns</span>
            <span className="metric-value">{quest.turns}</span>
          </div>
          <div className="metric">
            <span className="metric-icon">🔮</span>
            <span className="metric-label">Tokens</span>
            <span className="metric-value">{formatTokens(quest.tokens_used)}</span>
          </div>
          <div className="metric">
            <span className="metric-icon">💎</span>
            <span className="metric-label">Cost</span>
            <span className="metric-value">{formatCost(quest.cost)}</span>
          </div>
          <div className="metric">
            <span className="metric-icon">⏳</span>
            <span className="metric-label">Duration</span>
            <span className="metric-value">{timeAgo(quest.started_at || quest.created_at)}</span>
          </div>
          {quest.children && quest.children > 0 && (
            <div className="metric">
              <span className="metric-icon">👥</span>
              <span className="metric-label">Minions</span>
              <span className="metric-value">{quest.children}</span>
            </div>
          )}
        </div>
      </div>

      {/* Boss Description */}
      {quest.description && (
        <div className="boss-description">
          <div className="description-label">📜 Battle Report</div>
          <div className="description-text">{quest.description}</div>
        </div>
      )}

      {/* Error state */}
      {quest.error_text && (
        <div className="boss-error">
          <span className="error-icon">💀</span>
          <span>Boss has entered Enraged state due to battle complications</span>
        </div>
      )}
    </div>
  );
}