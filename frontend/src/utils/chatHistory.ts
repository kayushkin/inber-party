import type { ChatMessage } from '../store';

const CHAT_HISTORY_KEY = 'inber-party-chat-history';
const MAX_MESSAGES_PER_AGENT = 100; // Prevent localStorage from growing too large
const MAX_AGE_DAYS = 7; // Remove messages older than 7 days

interface ChatHistoryData {
  [agentId: string]: ChatMessage[];
}

// Load chat history from localStorage
export function loadChatHistory(): Record<string, ChatMessage[]> {
  try {
    const stored = localStorage.getItem(CHAT_HISTORY_KEY);
    if (!stored) return {};
    
    const data: ChatHistoryData = JSON.parse(stored);
    const now = Date.now();
    const maxAge = MAX_AGE_DAYS * 24 * 60 * 60 * 1000;
    
    // Clean up old messages and limit per agent
    const cleaned: ChatHistoryData = {};
    for (const [agentId, messages] of Object.entries(data)) {
      const recent = messages
        .filter(msg => {
          try {
            const msgTime = new Date(msg.timestamp).getTime();
            return (now - msgTime) < maxAge;
          } catch {
            return false; // Invalid timestamp
          }
        })
        .slice(-MAX_MESSAGES_PER_AGENT); // Keep only the most recent messages
      
      if (recent.length > 0) {
        cleaned[agentId] = recent;
      }
    }
    
    // Save back the cleaned data
    if (JSON.stringify(cleaned) !== JSON.stringify(data)) {
      saveChatHistory(cleaned);
    }
    
    return cleaned;
  } catch (error) {
    console.warn('Failed to load chat history from localStorage:', error);
    return {};
  }
}

// Save chat history to localStorage
export function saveChatHistory(chatMessages: Record<string, ChatMessage[]>) {
  try {
    // Clean before saving - remove streaming messages and old messages
    const cleaned: ChatHistoryData = {};
    const now = Date.now();
    const maxAge = MAX_AGE_DAYS * 24 * 60 * 60 * 1000;
    
    for (const [agentId, messages] of Object.entries(chatMessages)) {
      const validMessages = messages
        .filter(msg => {
          // Remove streaming messages (they're temporary)
          if (msg.streaming) return false;
          
          // Remove old messages
          try {
            const msgTime = new Date(msg.timestamp).getTime();
            return (now - msgTime) < maxAge;
          } catch {
            return false;
          }
        })
        .slice(-MAX_MESSAGES_PER_AGENT); // Keep only recent messages
      
      if (validMessages.length > 0) {
        cleaned[agentId] = validMessages;
      }
    }
    
    localStorage.setItem(CHAT_HISTORY_KEY, JSON.stringify(cleaned));
  } catch (error) {
    console.warn('Failed to save chat history to localStorage:', error);
    
    // If storage is full, try to clear old data and retry
    try {
      clearOldChatHistory();
      localStorage.setItem(CHAT_HISTORY_KEY, JSON.stringify(chatMessages));
    } catch {
      console.error('Failed to save chat history even after cleanup');
    }
  }
}

// Save a single agent's chat history (more efficient than saving all)
export function saveChatHistoryForAgent(agentId: string, messages: ChatMessage[]) {
  try {
    const existing = loadChatHistory();
    existing[agentId] = messages;
    saveChatHistory(existing);
  } catch (error) {
    console.warn('Failed to save chat history for agent:', agentId, error);
  }
}

// Clear old chat history to free up space
export function clearOldChatHistory() {
  try {
    const existing = loadChatHistory();
    const now = Date.now();
    const maxAge = 3 * 24 * 60 * 60 * 1000; // 3 days for cleanup
    
    const cleaned: ChatHistoryData = {};
    for (const [agentId, messages] of Object.entries(existing)) {
      const recent = messages
        .filter(msg => {
          try {
            const msgTime = new Date(msg.timestamp).getTime();
            return (now - msgTime) < maxAge;
          } catch {
            return false;
          }
        })
        .slice(-50); // Keep only 50 most recent messages per agent
      
      if (recent.length > 0) {
        cleaned[agentId] = recent;
      }
    }
    
    localStorage.setItem(CHAT_HISTORY_KEY, JSON.stringify(cleaned));
  } catch (error) {
    console.warn('Failed to clear old chat history:', error);
    // Last resort: clear all chat history
    try {
      localStorage.removeItem(CHAT_HISTORY_KEY);
    } catch {
      console.error('Failed to clear chat history completely');
    }
  }
}

// Get storage usage info (for debugging)
export function getChatHistoryStats() {
  try {
    const data = localStorage.getItem(CHAT_HISTORY_KEY);
    if (!data) return { agents: 0, totalMessages: 0, sizeKB: 0 };
    
    const parsed: ChatHistoryData = JSON.parse(data);
    const agents = Object.keys(parsed).length;
    const totalMessages = Object.values(parsed).reduce((sum, messages) => sum + messages.length, 0);
    const sizeKB = Math.round(data.length / 1024);
    
    return { agents, totalMessages, sizeKB };
  } catch {
    return { agents: 0, totalMessages: 0, sizeKB: 0 };
  }
}