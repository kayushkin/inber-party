import { useState, useEffect } from 'react';
import { useSoundManager } from '../hooks/useSoundManager';
import './SoundToggle.css';

export default function SoundToggle() {
  const soundManager = useSoundManager();
  const [isOpen, setIsOpen] = useState(false);
  const [config, setConfig] = useState(soundManager.config);

  // Sync local state with sound manager
  useEffect(() => {
    // Use microtask to avoid synchronous setState in effect
    Promise.resolve().then(() => setConfig(soundManager.config));
  }, [soundManager.config]);

  const handleToggleEnabled = () => {
    const newEnabled = !config.enabled;
    soundManager.setEnabled(newEnabled);
    setConfig({ ...config, enabled: newEnabled });
    
    // Play test sound when enabling
    if (newEnabled) {
      setTimeout(() => soundManager.testSound('message'), 100);
    }
  };

  const handleVolumeChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const newVolume = parseFloat(event.target.value);
    soundManager.setVolume(newVolume);
    setConfig({ ...config, volume: newVolume });
  };

  const handleTestSound = (soundType: 'levelUp' | 'questComplete' | 'message' | 'error') => {
    soundManager.testSound(soundType);
  };

  return (
    <div className="sound-toggle">
      {/* Main toggle button */}
      <button
        className={`sound-toggle-btn ${config.enabled ? 'enabled' : 'disabled'}`}
        onClick={handleToggleEnabled}
        title={config.enabled ? 'Sounds enabled' : 'Sounds muted'}
      >
        {config.enabled ? '🔊' : '🔇'}
      </button>

      {/* Settings panel button */}
      <button
        className="sound-settings-btn"
        onClick={() => setIsOpen(!isOpen)}
        title="Sound settings"
      >
        ⚙️
      </button>

      {/* Settings panel */}
      {isOpen && (
        <div className="sound-settings-panel">
          <div className="sound-settings-header">
            <span>Sound Settings</span>
            <button 
              className="close-btn"
              onClick={() => setIsOpen(false)}
              title="Close"
            >
              ✕
            </button>
          </div>
          
          <div className="sound-setting">
            <label>
              <input
                type="checkbox"
                checked={config.enabled}
                onChange={handleToggleEnabled}
              />
              Enable sounds
            </label>
          </div>

          <div className="sound-setting">
            <label>
              Volume: {Math.round(config.volume * 100)}%
              <input
                type="range"
                min="0"
                max="1"
                step="0.1"
                value={config.volume}
                onChange={handleVolumeChange}
                disabled={!config.enabled}
              />
            </label>
          </div>

          <div className="sound-tests">
            <div className="sound-tests-title">Test Sounds:</div>
            <div className="sound-test-buttons">
              <button
                onClick={() => handleTestSound('levelUp')}
                disabled={!config.enabled}
                title="Test level up sound"
              >
                🎊 Level Up
              </button>
              <button
                onClick={() => handleTestSound('questComplete')}
                disabled={!config.enabled}
                title="Test quest complete sound"
              >
                📜 Quest Done
              </button>
              <button
                onClick={() => handleTestSound('message')}
                disabled={!config.enabled}
                title="Test message sound"
              >
                💬 Message
              </button>
              <button
                onClick={() => handleTestSound('error')}
                disabled={!config.enabled}
                title="Test error sound"
              >
                ⚠️ Error
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}