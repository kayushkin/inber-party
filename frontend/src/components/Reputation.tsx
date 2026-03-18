import type { RPGReputation } from '../store';
import './Reputation.css';

interface ReputationProps {
  reputation: RPGReputation[];
}

function getReputationLevel(score: number): string {
  if (score >= 900) return 'Legendary';
  if (score >= 800) return 'Renowned';
  if (score >= 700) return 'Respected';
  if (score >= 600) return 'Trusted';
  if (score >= 500) return 'Competent';
  if (score >= 400) return 'Adequate';
  if (score >= 300) return 'Developing';
  if (score >= 200) return 'Novice';
  if (score >= 100) return 'Beginner';
  return 'Unknown';
}

function getReputationColor(score: number): string {
  if (score >= 900) return '#fbbf24'; // legendary - gold
  if (score >= 800) return '#a78bfa'; // renowned - purple
  if (score >= 700) return '#3b82f6'; // respected - blue
  if (score >= 600) return '#22c55e'; // trusted - green
  if (score >= 500) return '#06b6d4'; // competent - cyan
  if (score >= 400) return '#eab308'; // adequate - yellow
  if (score >= 300) return '#f97316'; // developing - orange
  if (score >= 200) return '#ef4444'; // novice - red
  if (score >= 100) return '#64748b'; // beginner - gray
  return '#374151'; // unknown - dark gray
}

function getDomainIcon(domain: string): string {
  const icons: Record<string, string> = {
    coding: '💻',
    testing: '🧪',
    documentation: '📚',
    devops: '⚙️',
    design: '🎨',
    security: '🔒',
    database: '📊',
    maintenance: '🔧',
    general: '⭐',
  };
  return icons[domain] || '⭐';
}

export default function Reputation({ reputation }: ReputationProps) {
  if (!reputation || reputation.length === 0) {
    return (
      <div className="reputation-empty">
        <div className="reputation-empty-icon">🌟</div>
        <div className="reputation-empty-text">No reputation yet</div>
        <div className="reputation-empty-subtext">Complete tasks to build reputation in different domains</div>
      </div>
    );
  }

  // Sort by score descending
  const sortedReputation = [...reputation].sort((a, b) => b.score - a.score);

  return (
    <div className="reputation-container">
      <div className="reputation-header">
        <h3>Domain Reputation</h3>
        <div className="reputation-summary">
          Strong in {sortedReputation.filter(r => r.score >= 600).length} domains
        </div>
      </div>
      
      <div className="reputation-list">
        {sortedReputation.map((rep) => (
          <div key={`${rep.domain}-${rep.id}`} className="reputation-item">
            <div className="reputation-item-header">
              <div className="reputation-domain">
                <span className="reputation-icon">{getDomainIcon(rep.domain)}</span>
                <span className="reputation-domain-name">{rep.domain}</span>
              </div>
              <div className="reputation-level" style={{ color: getReputationColor(rep.score) }}>
                {getReputationLevel(rep.score)}
              </div>
            </div>
            
            <div className="reputation-bar">
              <div 
                className="reputation-bar-fill" 
                style={{ 
                  width: `${(rep.score / 1000) * 100}%`,
                  backgroundColor: getReputationColor(rep.score)
                }}
              />
              <span className="reputation-score">{rep.score}/1000</span>
            </div>
            
            <div className="reputation-stats">
              <div className="reputation-stat">
                <span className="reputation-stat-label">Tasks:</span>
                <span className="reputation-stat-value">{rep.task_count}</span>
              </div>
              <div className="reputation-stat">
                <span className="reputation-stat-label">Success Rate:</span>
                <span className="reputation-stat-value">{Math.round(rep.success_rate * 100)}%</span>
              </div>
            </div>
          </div>
        ))}
      </div>
      
      <div className="reputation-footer">
        <div className="reputation-legend">
          <div className="reputation-legend-item">
            <div className="reputation-legend-color" style={{ backgroundColor: '#22c55e' }}></div>
            <span>Trusted (600+)</span>
          </div>
          <div className="reputation-legend-item">
            <div className="reputation-legend-color" style={{ backgroundColor: '#3b82f6' }}></div>
            <span>Respected (700+)</span>
          </div>
          <div className="reputation-legend-item">
            <div className="reputation-legend-color" style={{ backgroundColor: '#fbbf24' }}></div>
            <span>Legendary (900+)</span>
          </div>
        </div>
      </div>
    </div>
  );
}