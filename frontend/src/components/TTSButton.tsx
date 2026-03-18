import { useState, useEffect } from 'react';
import { useTTS } from '../hooks/useTTS';
import './TTSButton.css';

interface Props {
  text: string;
  agentName?: string;
  className?: string;
  size?: 'small' | 'medium';
}

export default function TTSButton({ text, agentName, className = '', size = 'small' }: Props) {
  const { speak, stop, speaking, supported, config } = useTTS();
  const [isThisMessageSpeaking, setIsThisMessageSpeaking] = useState(false);

  useEffect(() => {
    // Reset speaking state when global speaking stops
    if (!speaking) {
      setIsThisMessageSpeaking(false);
    }
  }, [speaking]);

  const handleClick = () => {
    if (!supported || !config.enabled) return;

    if (isThisMessageSpeaking) {
      stop();
      setIsThisMessageSpeaking(false);
    } else {
      // Stop any other message that might be speaking
      stop();
      
      // Prepare text for speech
      let speechText = text;
      if (agentName) {
        speechText = `${agentName} says: ${text}`;
      }
      
      setIsThisMessageSpeaking(true);
      speak(speechText);
    }
  };

  // Don't render if TTS is not supported or disabled
  if (!supported || !config.enabled) {
    return null;
  }

  // Don't render for empty or very short text
  if (!text || text.trim().length < 3) {
    return null;
  }

  const iconSize = size === 'small' ? '12' : '16';
  const buttonClass = `tts-button ${className} tts-button-${size}`;

  return (
    <button
      className={buttonClass}
      onClick={handleClick}
      title={isThisMessageSpeaking ? 'Stop speech' : 'Read aloud'}
      disabled={!text.trim()}
    >
      {isThisMessageSpeaking ? (
        <svg width={iconSize} height={iconSize} viewBox="0 0 24 24" fill="currentColor">
          <path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z"/>
        </svg>
      ) : (
        <svg width={iconSize} height={iconSize} viewBox="0 0 24 24" fill="currentColor">
          <path d="M3 9v6h4l5 5V4L7 9H3zm13.5 3c0-1.77-1.02-3.29-2.5-4.03v8.05c1.48-.73 2.5-2.25 2.5-4.02zM14 3.23v2.06c2.89.86 5 3.54 5 6.71s-2.11 5.85-5 6.71v2.06c4.01-.91 7-4.49 7-8.77s-2.99-7.86-7-8.77z"/>
        </svg>
      )}
    </button>
  );
}