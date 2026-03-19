import { useCallback, useEffect, useRef, useState } from 'react';

export interface TTSConfig {
  enabled: boolean;
  voice: string | null;
  rate: number;
  pitch: number;
  volume: number;
}

const DEFAULT_TTS_CONFIG: TTSConfig = {
  enabled: true,
  voice: null, // Will use system default or preference
  rate: 1.0,
  pitch: 1.0,
  volume: 0.8,
};

export function useTTS() {
  const [config, setConfig] = useState<TTSConfig>(() => {
    try {
      const saved = localStorage.getItem('ttsConfig');
      if (saved) {
        const parsedConfig = JSON.parse(saved);
        return { ...DEFAULT_TTS_CONFIG, ...parsedConfig };
      }
    } catch {
      console.warn('Failed to load TTS config from localStorage');
    }
    return DEFAULT_TTS_CONFIG;
  });
  const [voices, setVoices] = useState<SpeechSynthesisVoice[]>([]);
  const [speaking, setSpeaking] = useState(false);
  const [supported] = useState(() => 'speechSynthesis' in window);
  const utteranceRef = useRef<SpeechSynthesisUtterance | null>(null);

  const saveConfig = useCallback((newConfig: TTSConfig): void => {
    try {
      localStorage.setItem('ttsConfig', JSON.stringify(newConfig));
      setConfig(newConfig);
    } catch {
      console.warn('Failed to save TTS config to localStorage');
    }
  }, []);

  const loadVoices = useCallback((): void => {
    const availableVoices = speechSynthesis.getVoices();
    setVoices(availableVoices);
    
    // If no voice is set, try to find a good default
    if (!config.voice && availableVoices.length > 0) {
      // Prefer English voices, then any voice that sounds natural
      const preferredVoice = availableVoices.find(voice => 
        voice.lang.startsWith('en') && (
          voice.name.toLowerCase().includes('nova') ||
          voice.name.toLowerCase().includes('natural') ||
          voice.name.toLowerCase().includes('neural')
        )
      ) || availableVoices.find(voice => voice.lang.startsWith('en')) || availableVoices[0];
      
      if (preferredVoice) {
        const newConfig = { ...config, voice: preferredVoice.name };
        saveConfig(newConfig);
      }
    }
  }, [config, saveConfig]);

  const speak = (text: string, options?: Partial<TTSConfig>): void => {
    if (!supported || !config.enabled || !text.trim()) return;
    
    // Cancel any ongoing speech
    stop();
    
    const utterance = new SpeechSynthesisUtterance(text);
    const effectiveConfig = { ...config, ...options };
    
    // Find the selected voice
    if (effectiveConfig.voice) {
      const selectedVoice = voices.find(voice => voice.name === effectiveConfig.voice);
      if (selectedVoice) {
        utterance.voice = selectedVoice;
      }
    }
    
    utterance.rate = effectiveConfig.rate;
    utterance.pitch = effectiveConfig.pitch;
    utterance.volume = effectiveConfig.volume;
    
    utterance.onstart = () => setSpeaking(true);
    utterance.onend = () => {
      setSpeaking(false);
      utteranceRef.current = null;
    };
    utterance.onerror = (event) => {
      console.warn('TTS error:', event.error);
      setSpeaking(false);
      utteranceRef.current = null;
    };
    
    utteranceRef.current = utterance;
    speechSynthesis.speak(utterance);
  };

  const stop = (): void => {
    if (speaking) {
      speechSynthesis.cancel();
      setSpeaking(false);
      utteranceRef.current = null;
    }
  };

  const pause = (): void => {
    if (speaking && !speechSynthesis.paused) {
      speechSynthesis.pause();
    }
  };

  const resume = (): void => {
    if (speaking && speechSynthesis.paused) {
      speechSynthesis.resume();
    }
  };

  const setEnabled = (enabled: boolean): void => {
    const newConfig = { ...config, enabled };
    saveConfig(newConfig);
    if (!enabled) {
      stop();
    }
  };

  const setVoice = (voiceName: string | null): void => {
    const newConfig = { ...config, voice: voiceName };
    saveConfig(newConfig);
  };

  const setRate = (rate: number): void => {
    const newConfig = { ...config, rate: Math.max(0.1, Math.min(10, rate)) };
    saveConfig(newConfig);
  };

  const setPitch = (pitch: number): void => {
    const newConfig = { ...config, pitch: Math.max(0, Math.min(2, pitch)) };
    saveConfig(newConfig);
  };

  const setVolume = (volume: number): void => {
    const newConfig = { ...config, volume: Math.max(0, Math.min(1, volume)) };
    saveConfig(newConfig);
  };

  // Initialize TTS support and load voices
  useEffect(() => {
    if (supported) {
      loadVoices();
      
      // Listen for voice changes (async loading)
      if (speechSynthesis.onvoiceschanged !== undefined) {
        speechSynthesis.onvoiceschanged = loadVoices;
      }
    }
  }, [supported, loadVoices]);

  // Helper to clean text for better speech
  const cleanTextForSpeech = (text: string): string => {
    return text
      // Remove markdown
      .replace(/\*\*(.*?)\*\*/g, '$1') // Bold
      .replace(/\*(.*?)\*/g, '$1') // Italic
      .replace(/`(.*?)`/g, '$1') // Inline code
      .replace(/```[\s\S]*?```/g, 'code block') // Code blocks
      .replace(/#{1,6}\s/g, '') // Headers
      .replace(/\[([^\]]*)\]\([^)]*\)/g, '$1') // Links
      // Remove special characters that don't read well
      .replace(/[📊🎉🔥⚡️🎯🏆💡🚀]/gu, '')
      // Clean up whitespace
      .replace(/\s+/g, ' ')
      .trim();
  };

  return {
    supported,
    config,
    voices,
    speaking,
    speak: (text: string, options?: Partial<TTSConfig>) => speak(cleanTextForSpeech(text), options),
    stop,
    pause,
    resume,
    setEnabled,
    setVoice,
    setRate,
    setPitch,
    setVolume,
    cleanTextForSpeech
  };
}