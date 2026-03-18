import { useEffect, useState } from 'react';
import { useTTS } from '../hooks/useTTS';
import './TTSSettings.css';

interface Props {
  onClose: () => void;
}

export default function TTSSettings({ onClose }: Props) {
  const {
    supported,
    config,
    voices,
    speaking,
    setEnabled,
    setVoice,
    setRate,
    setPitch,
    setVolume,
    speak,
    stop
  } = useTTS();

  const [localConfig, setLocalConfig] = useState(config);

  useEffect(() => {
    setLocalConfig(config);
  }, [config]);

  const handleToggle = () => {
    const newEnabled = !localConfig.enabled;
    setEnabled(newEnabled);
    setLocalConfig(prev => ({ ...prev, enabled: newEnabled }));
    
    if (!newEnabled && speaking) {
      stop();
    }
  };

  const handleVoiceChange = (voiceName: string) => {
    setVoice(voiceName);
    setLocalConfig(prev => ({ ...prev, voice: voiceName }));
  };

  const handleRateChange = (rate: number) => {
    setRate(rate);
    setLocalConfig(prev => ({ ...prev, rate }));
  };

  const handlePitchChange = (pitch: number) => {
    setPitch(pitch);
    setLocalConfig(prev => ({ ...prev, pitch }));
  };

  const handleVolumeChange = (volume: number) => {
    setVolume(volume);
    setLocalConfig(prev => ({ ...prev, volume }));
  };

  const handleTestVoice = (voice?: string) => {
    if (!localConfig.enabled) return;
    
    const testText = voice 
      ? `This is how ${voice} sounds as your adventure guide.`
      : "This is a test of text-to-speech for your adventure messages.";
    
    speak(testText, voice ? { voice } : undefined);
  };

  if (!supported) {
    return (
      <div className="tts-settings">
        <div className="tts-settings-header">
          <h3>🔊 Text-to-Speech</h3>
          <button className="tts-close" onClick={onClose}>✕</button>
        </div>
        <div className="tts-not-supported">
          <p>😔 Text-to-speech is not supported in your browser.</p>
          <p>Try using a modern browser like Chrome, Firefox, or Safari.</p>
        </div>
      </div>
    );
  }

  // Group voices by language for better organization
  const voicesByLang = voices.reduce((acc, voice) => {
    const lang = voice.lang.split('-')[0] || 'other';
    if (!acc[lang]) acc[lang] = [];
    acc[lang].push(voice);
    return acc;
  }, {} as Record<string, SpeechSynthesisVoice[]>);

  // Prioritize English voices
  const sortedLangs = Object.keys(voicesByLang).sort((a, b) => {
    if (a === 'en') return -1;
    if (b === 'en') return 1;
    return a.localeCompare(b);
  });

  return (
    <div className="tts-settings">
      <div className="tts-settings-header">
        <h3>🔊 Text-to-Speech</h3>
        <button className="tts-close" onClick={onClose}>✕</button>
      </div>
      
      <div className="tts-settings-content">
        <div className="tts-setting-group">
          <div className="tts-setting">
            <label>
              <input
                type="checkbox"
                checked={localConfig.enabled}
                onChange={handleToggle}
              />
              Enable voice messages for agent responses
            </label>
          </div>
        </div>

        {localConfig.enabled && (
          <>
            <div className="tts-setting-group">
              <label className="tts-label">Voice</label>
              <select
                value={localConfig.voice || ''}
                onChange={(e) => handleVoiceChange(e.target.value)}
                className="tts-select"
              >
                <option value="">System Default</option>
                {sortedLangs.map(lang => (
                  <optgroup key={lang} label={lang.toUpperCase()}>
                    {voicesByLang[lang].map(voice => (
                      <option key={voice.name} value={voice.name}>
                        {voice.name} {voice.localService ? '(local)' : '(remote)'}
                      </option>
                    ))}
                  </optgroup>
                ))}
              </select>
              {localConfig.voice && (
                <button 
                  className="tts-test-btn" 
                  onClick={() => handleTestVoice(localConfig.voice || undefined)}
                  disabled={speaking}
                >
                  {speaking ? '⏸️ Stop' : '▶️ Test'}
                </button>
              )}
            </div>

            <div className="tts-setting-group">
              <label className="tts-label">
                Speech Rate: {localConfig.rate.toFixed(1)}x
              </label>
              <input
                type="range"
                min="0.5"
                max="2"
                step="0.1"
                value={localConfig.rate}
                onChange={(e) => handleRateChange(parseFloat(e.target.value))}
                className="tts-slider"
              />
              <div className="tts-slider-labels">
                <span>Slow</span>
                <span>Normal</span>
                <span>Fast</span>
              </div>
            </div>

            <div className="tts-setting-group">
              <label className="tts-label">
                Pitch: {localConfig.pitch.toFixed(1)}
              </label>
              <input
                type="range"
                min="0.5"
                max="2"
                step="0.1"
                value={localConfig.pitch}
                onChange={(e) => handlePitchChange(parseFloat(e.target.value))}
                className="tts-slider"
              />
              <div className="tts-slider-labels">
                <span>Low</span>
                <span>Normal</span>
                <span>High</span>
              </div>
            </div>

            <div className="tts-setting-group">
              <label className="tts-label">
                Volume: {Math.round(localConfig.volume * 100)}%
              </label>
              <input
                type="range"
                min="0"
                max="1"
                step="0.1"
                value={localConfig.volume}
                onChange={(e) => handleVolumeChange(parseFloat(e.target.value))}
                className="tts-slider"
              />
              <div className="tts-slider-labels">
                <span>Quiet</span>
                <span>Loud</span>
              </div>
            </div>

            <div className="tts-test-section">
              <button 
                className="tts-test-full-btn"
                onClick={() => handleTestVoice()}
                disabled={speaking}
              >
                {speaking ? '⏸️ Stop Test' : '🎙️ Test Current Settings'}
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}