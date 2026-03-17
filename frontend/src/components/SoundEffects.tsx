import { useEffect, useRef } from 'react';
import { useStore } from '../store';
import { useSoundManager } from '../hooks/useSoundManager';

/**
 * SoundEffects component - handles playing sounds when events occur
 * This component listens to store triggers and plays appropriate sounds
 */
export default function SoundEffects() {
  const soundManager = useSoundManager();
  
  // Store selectors
  const levelUpTriggers = useStore((s) => s.levelUpTriggers);
  const questCompletionTriggers = useStore((s) => s.questCompletionTriggers);
  const chatMessages = useStore((s) => s.chatMessages);
  
  // Track last processed triggers to avoid duplicates
  const lastLevelUpTriggers = useRef<Record<string, number>>({});
  const lastQuestCompletionTriggers = useRef<Record<number, number>>({});
  const lastMessageCounts = useRef<Record<string, number>>({});

  // Handle level up sounds
  useEffect(() => {
    Object.entries(levelUpTriggers).forEach(([agentId, trigger]) => {
      const lastTrigger = lastLevelUpTriggers.current[agentId];
      if (trigger && trigger !== lastTrigger) {
        soundManager.playLevelUp();
        lastLevelUpTriggers.current[agentId] = trigger;
      }
    });
  }, [levelUpTriggers, soundManager]);

  // Handle quest completion sounds
  useEffect(() => {
    Object.entries(questCompletionTriggers).forEach(([questIdStr, data]) => {
      const questId = parseInt(questIdStr, 10);
      const lastTrigger = lastQuestCompletionTriggers.current[questId];
      if (data?.trigger && data.trigger !== lastTrigger) {
        soundManager.playQuestComplete();
        lastQuestCompletionTriggers.current[questId] = data.trigger;
      }
    });
  }, [questCompletionTriggers, soundManager]);

  // Handle new message sounds (optional - might be too noisy)
  useEffect(() => {
    Object.entries(chatMessages).forEach(([agentId, messages]) => {
      const messageCount = messages.length;
      const lastCount = lastMessageCounts.current[agentId] || 0;
      
      if (messageCount > lastCount) {
        // Only play sound for assistant messages (not user messages)
        const newMessages = messages.slice(lastCount);
        const hasNewAssistantMessage = newMessages.some(msg => msg.role === 'assistant' && !msg.streaming);
        
        if (hasNewAssistantMessage) {
          // Add a small delay to avoid playing sound during streaming
          setTimeout(() => soundManager.playMessage(), 500);
        }
      }
      
      lastMessageCounts.current[agentId] = messageCount;
    });
  }, [chatMessages, soundManager]);

  // This component doesn't render anything
  return null;
}