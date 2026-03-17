import { useEffect, useRef } from 'react';

export type SoundType = 'levelUp' | 'questComplete' | 'message' | 'error';

interface SoundConfig {
  enabled: boolean;
  volume: number;
  sounds: Record<SoundType, string>;
}

const DEFAULT_CONFIG: SoundConfig = {
  enabled: false, // Muted by default as requested
  volume: 0.6,
  sounds: {
    levelUp: '/sounds/level-up.wav',
    questComplete: '/sounds/quest-complete.wav', 
    message: '/sounds/message.wav',
    error: '/sounds/error.wav'
  }
};

export function useSoundManager() {
  const audioContextRef = useRef<AudioContext | null>(null);
  const configRef = useRef<SoundConfig>({ ...DEFAULT_CONFIG });
  const audioBuffersRef = useRef<Record<SoundType, AudioBuffer | null>>({
    levelUp: null,
    questComplete: null,
    message: null,
    error: null
  });

  // Initialize audio context (user interaction required for modern browsers)
  const initAudio = () => {
    if (audioContextRef.current) return audioContextRef.current;
    
    try {
      audioContextRef.current = new (window.AudioContext || (window as any).webkitAudioContext)();
      return audioContextRef.current;
    } catch {
      console.warn('Web Audio API not supported');
      return null;
    }
  };

  // Load sound config from localStorage
  const loadConfig = (): SoundConfig => {
    try {
      const saved = localStorage.getItem('soundConfig');
      if (saved) {
        return { ...DEFAULT_CONFIG, ...JSON.parse(saved) };
      }
    } catch {
      console.warn('Failed to load sound config from localStorage');
    }
    return DEFAULT_CONFIG;
  };

  // Save sound config to localStorage
  const saveConfig = (config: SoundConfig) => {
    try {
      localStorage.setItem('soundConfig', JSON.stringify(config));
      configRef.current = config;
    } catch {
      console.warn('Failed to save sound config to localStorage');
    }
  };

  // Generate a simple tone as fallback when audio files aren't available
  const generateTone = async (frequency: number, duration: number, waveType: OscillatorType = 'sine'): Promise<AudioBuffer | null> => {
    const audioContext = initAudio();
    if (!audioContext) return null;

    const sampleRate = audioContext.sampleRate;
    const length = sampleRate * duration;
    const buffer = audioContext.createBuffer(1, length, sampleRate);
    const data = buffer.getChannelData(0);

    for (let i = 0; i < length; i++) {
      const time = i / sampleRate;
      
      // Generate wave based on type
      let value: number;
      switch (waveType) {
        case 'square':
          value = Math.sign(Math.sin(2 * Math.PI * frequency * time));
          break;
        case 'sawtooth':
          value = 2 * (time * frequency - Math.floor(time * frequency + 0.5));
          break;
        case 'triangle':
          value = 2 * Math.abs(2 * (time * frequency - Math.floor(time * frequency + 0.5))) - 1;
          break;
        default: // sine
          value = Math.sin(2 * Math.PI * frequency * time);
      }
      
      // Add envelope (fade in/out)
      const fadeTime = 0.1;
      if (time < fadeTime) {
        value *= time / fadeTime;
      } else if (time > duration - fadeTime) {
        value *= (duration - time) / fadeTime;
      }
      
      data[i] = value * 0.3; // Reduce volume
    }

    return buffer;
  };

  // Load or generate audio for a sound type
  const loadSound = async (soundType: SoundType): Promise<AudioBuffer | null> => {
    const audioContext = initAudio();
    if (!audioContext) return null;

    // Try to load from URL first
    const soundUrl = configRef.current.sounds[soundType];
    try {
      const response = await fetch(soundUrl);
      if (response.ok) {
        const arrayBuffer = await response.arrayBuffer();
        return await audioContext.decodeAudioData(arrayBuffer);
      }
    } catch {
      // File not found, generate fallback tone
    }

    // Generate fallback tones for different sound types
    switch (soundType) {
      case 'levelUp':
        // Ascending arpeggio for level up
        return await generateTone(523, 0.8); // C5
      case 'questComplete':
        // Triumphant chord for quest complete
        return await generateTone(440, 1.0); // A4
      case 'message':
        // Simple notification tone
        return await generateTone(800, 0.3); // High tone
      case 'error':
        // Low warning tone
        return await generateTone(200, 0.5, 'sawtooth');
      default:
        return null;
    }
  };

  // Play a sound
  const playSound = async (soundType: SoundType, volume?: number) => {
    const config = configRef.current;
    if (!config.enabled) return;

    const audioContext = initAudio();
    if (!audioContext || audioContext.state === 'suspended') {
      try {
        await audioContext?.resume();
      } catch {
        return;
      }
    }

    // Load sound if not cached
    if (!audioBuffersRef.current[soundType]) {
      audioBuffersRef.current[soundType] = await loadSound(soundType);
    }

    const buffer = audioBuffersRef.current[soundType];
    if (!buffer || !audioContext) return;

    try {
      const source = audioContext.createBufferSource();
      const gainNode = audioContext.createGain();
      
      source.buffer = buffer;
      gainNode.gain.value = (volume ?? config.volume) * 0.5; // Additional scaling
      
      source.connect(gainNode);
      gainNode.connect(audioContext.destination);
      source.start();
    } catch (error) {
      console.warn('Failed to play sound:', error);
    }
  };

  // Initialize on mount
  useEffect(() => {
    configRef.current = loadConfig();
  }, []);

  // API for components
  return {
    get config() {
      return configRef.current;
    },
    
    playLevelUp: () => playSound('levelUp'),
    playQuestComplete: () => playSound('questComplete'),
    playMessage: () => playSound('message'),
    playError: () => playSound('error'),
    
    setEnabled: (enabled: boolean) => {
      const newConfig = { ...configRef.current, enabled };
      saveConfig(newConfig);
    },
    
    setVolume: (volume: number) => {
      const newConfig = { ...configRef.current, volume: Math.max(0, Math.min(1, volume)) };
      saveConfig(newConfig);
    },
    
    // Test function
    testSound: (soundType: SoundType) => {
      playSound(soundType, 0.8);
    }
  };
}