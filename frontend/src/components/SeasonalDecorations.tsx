import { useStore } from '../store';
import './SeasonalDecorations.css';

interface SeasonalDecorationsProps {
  className?: string;
}

const SeasonalDecorations: React.FC<SeasonalDecorationsProps> = ({ className = '' }) => {
  const seasonalEvent = useStore((s) => s.seasonalEvent);
  const seasonalDecorations = useStore((s) => s.seasonalDecorations);
  const seasonalEnabled = useStore((s) => s.seasonalEnabled);

  if (!seasonalEnabled || !seasonalEvent || seasonalDecorations.length === 0) {
    return null;
  }

  return (
    <div className={`seasonal-decorations ${className}`}>
      {/* Floating decorations */}
      <div className="decoration-layer">
        {seasonalDecorations.slice(0, 6).map((decoration, index) => (
          <div
            key={`decoration-${index}`}
            className={`floating-decoration decoration-${index + 1}`}
            style={{
              animationDelay: `${index * 0.5}s`,
              left: `${10 + (index * 15)}%`,
            }}
          >
            {decoration}
          </div>
        ))}
      </div>

      {/* Corner decorations */}
      <div className="corner-decorations">
        <div className="corner-decoration top-left">
          {seasonalDecorations[0]}
        </div>
        <div className="corner-decoration top-right">
          {seasonalDecorations[1] || seasonalDecorations[0]}
        </div>
        <div className="corner-decoration bottom-left">
          {seasonalDecorations[2] || seasonalDecorations[0]}
        </div>
        <div className="corner-decoration bottom-right">
          {seasonalDecorations[3] || seasonalDecorations[0]}
        </div>
      </div>

      {/* Particle effects for specific events */}
      {seasonalEvent.theme.particles && (
        <div className={`particle-effect ${seasonalEvent.theme.particles}`}>
          {Array.from({ length: 20 }).map((_, i) => (
            <div 
              key={`particle-${i}`}
              className="particle"
              style={{
                left: `${Math.random() * 100}%`,
                animationDelay: `${Math.random() * 2}s`,
                animationDuration: `${3 + Math.random() * 2}s`
              }}
            />
          ))}
        </div>
      )}
    </div>
  );
};

export default SeasonalDecorations;