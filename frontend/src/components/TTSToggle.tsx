import { useState } from 'react';
import { useTTS } from '../hooks/useTTS';
import TTSSettings from './TTSSettings';
import './TTSToggle.css';

export default function TTSToggle() {
  const { supported, config, speaking, stop, setEnabled, speak } = useTTS();
  const [isSettingsOpen, setIsSettingsOpen] = useState(false);

  // Don't render if TTS is not supported
  if (!supported) return null;

  const handleToggleEnabled = () => {
    const newEnabled = !config.enabled;
    setEnabled(newEnabled);
    
    // Stop any ongoing speech when disabling
    if (!newEnabled && speaking) {
      stop();
    }
    
    // Play test speech when enabling
    if (newEnabled) {
      setTimeout(() => {
        speak("Voice messages enabled for your adventure party!");
      }, 100);
    }
  };

  const handleSettingsToggle = () => {
    setIsSettingsOpen(!isSettingsOpen);
  };

  const handleCloseSettings = () => {
    setIsSettingsOpen(false);
  };

  return (
    <>
      <div className="tts-toggle">
        {/* Main TTS toggle button */}
        <button
          className={`tts-toggle-btn ${config.enabled ? 'enabled' : 'disabled'} ${speaking ? 'speaking' : ''}`}
          onClick={handleToggleEnabled}
          title={config.enabled ? 'Voice messages enabled' : 'Voice messages disabled'}
        >
          {speaking ? '🎙️' : config.enabled ? '🗣️' : '🔇'}
        </button>

        {/* Settings button */}
        <button
          className="tts-settings-toggle-btn"
          onClick={handleSettingsToggle}
          title="Voice settings"
        >
          ⚙️
        </button>
      </div>

      {/* Settings modal */}
      {isSettingsOpen && (
        <div className="tts-settings-overlay">
          <TTSSettings onClose={handleCloseSettings} />
        </div>
      )}
    </>
  );
}