import { useStore } from '../store';
import type { SeasonalEvent } from '../utils/seasonalEvents';
import './SeasonalEventBanner.css';

const SeasonalEventBanner: React.FC = () => {
  const seasonalEvent = useStore((s) => s.seasonalEvent);
  const seasonalEnabled = useStore((s) => s.seasonalEnabled);
  const toggleSeasonalEvents = useStore((s) => s.toggleSeasonalEvents);
  const seasonalXpMultiplier = useStore((s) => s.seasonalXpMultiplier);

  if (!seasonalEvent) {
    return null;
  }

  return (
    <div className={`seasonal-event-banner ${seasonalEnabled ? 'enabled' : 'disabled'}`}>
      <div className="seasonal-content">
        <div className="seasonal-info">
          <div className="seasonal-icon">
            {seasonalEvent.theme.decorations[0] || '✨'}
          </div>
          <div className="seasonal-text">
            <span className="seasonal-name">{seasonalEvent.name}</span>
            <span className="seasonal-desc">{seasonalEvent.description}</span>
          </div>
        </div>
        
        {seasonalEnabled && seasonalXpMultiplier > 1 && (
          <div className="seasonal-bonus">
            <span className="bonus-label">XP Bonus:</span>
            <span className="bonus-value">+{Math.round((seasonalXpMultiplier - 1) * 100)}%</span>
          </div>
        )}
        
        <button
          className={`seasonal-toggle ${seasonalEnabled ? 'enabled' : 'disabled'}`}
          onClick={toggleSeasonalEvents}
          title={`${seasonalEnabled ? 'Disable' : 'Enable'} seasonal events`}
        >
          {seasonalEnabled ? '🎉' : '⚪'}
        </button>
      </div>
      
      {seasonalEnabled && (
        <div className="seasonal-progress">
          <div className="progress-bar">
            <div 
              className="progress-fill"
              style={{
                width: `${getEventProgress(seasonalEvent)}%`
              }}
            />
          </div>
          <div className="progress-text">
            {getDaysRemaining(seasonalEvent)} days remaining
          </div>
        </div>
      )}
    </div>
  );
};

// Helper function to calculate event progress
function getEventProgress(event: SeasonalEvent): number {
  const now = new Date().getTime();
  const start = event.startDate.getTime();
  const end = event.endDate.getTime();
  
  const total = end - start;
  const elapsed = now - start;
  
  return Math.max(0, Math.min(100, (elapsed / total) * 100));
}

// Helper function to get days remaining
function getDaysRemaining(event: SeasonalEvent): number {
  const now = new Date().getTime();
  const end = event.endDate.getTime();
  
  const msPerDay = 24 * 60 * 60 * 1000;
  return Math.max(0, Math.ceil((end - now) / msPerDay));
}

export default SeasonalEventBanner;