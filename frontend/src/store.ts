import { create } from 'zustand';
import { loadChatHistory, saveChatHistoryForAgent, saveChatHistory } from './utils/chatHistory';
import { 
  getActiveSeasonalEvent, 
  applySeasonalTheme, 
  removeSeasonalTheme,
  getSeasonalXpMultiplier,
  getSeasonalDecorations,
  type SeasonalEvent 
} from './utils/seasonalEvents';
import { wsManager } from './utils/websocket';

// ── Types ──────────────────────────────────────────────────

// WebSocket message types
interface QuestEventMessage {
  type: 'quest_event';
  data: {
    type: string;
    message: string;
  };
}

interface InberUpdateMessage {
  type: 'inber_update';
  data: {
    agents?: RPGAgent[];
    quests?: RPGQuest[];
    stats?: unknown;
  };
}

type WebSocketMessage = QuestEventMessage | InberUpdateMessage;

// Type guards for WebSocket messages
function isWebSocketMessage(data: unknown): data is WebSocketMessage {
  return typeof data === 'object' && data !== null && 
         'type' in data && 'data' in data && 
         typeof (data as WebSocketMessage).type === 'string' &&
         ['quest_event', 'inber_update'].includes((data as WebSocketMessage).type);
}

export type QuestStatus = 
  | 'pending'
  | 'in_progress'
  | 'completed' 
  | 'failed'
  | 'cancelled'
  | 'abandoned';

export interface RPGHeldItem {
  id: string;
  name: string;
  description: string;
  icon: string;
  activity_type: 'edit' | 'spawn' | 'search' | 'docs' | 'infra' | 'debug' | 'create';
  priority: number;
}

export interface RPGReputation {
  id: number;
  agent_id: number;
  domain: string;
  score: number;
  task_count: number;
  success_rate: number;
  last_update: string;
}

export interface RPGAgent {
  id: string;
  name: string;
  title: string;
  class: string;
  level: number;
  xp: number;
  xp_to_next: number;
  gold: number;
  energy: number;
  max_energy: number;
  status: string;
  orchestrator: string;
  avatar_emoji: string;
  total_tokens: number;
  total_cost: number;
  session_count: number;
  quest_count: number;
  error_count: number;
  skills: { skill_name: string; level: number; task_count: number }[];
  last_active?: string;
  held_items?: RPGHeldItem[];
  mood?: string;
  mood_score?: number;
  workload?: number;
  reputation?: RPGReputation[];
}

export interface RPGQuest {
  id: number;
  name: string;
  description: string;
  difficulty: number;
  xp_reward: number;
  gold_reward: number;
  status: string;
  assigned_agent_id?: string;
  assigned_agent_name?: string;
  progress: number;
  turns: number;
  tokens_used: number;
  cost: number;
  created_at: string;
  started_at?: string;
  completed_at?: string;
  error_text?: string;
  children?: number;
}

export interface RPGStats {
  total_agents: number;
  active_quests: number;
  completed_quests: number;
  failed_quests: number;
  total_xp: number;
  total_gold: number;
  total_tokens: number;
  total_cost: number;
  average_agent_level: number;
  total_sessions: number;
  uptime: string;
}

export interface RPGAchievement {
  id: string;
  name: string;
  description: string;
  icon: string;
  unlocked_at: string;
}

export interface QuestHistoryEntry {
  id: number;
  tokens: number;
  cost: number;
  status: string;
  started_at: string;
  completed_at?: string;
}

export interface RPGParty {
  id: number;
  name: string;
  description: string;
  leader_id: number;
  status: string;
  created_at: string;
  updated_at: string;
  leader?: {
    name: string;
    title: string;
    class: string;
  };
  member_count?: number;
}

export interface RPGPartyDetail extends RPGParty {
  members: RPGAgent[];
  tasks: RPGQuest[];
  leader: RPGAgent;
}

export interface ChatMessage {
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: string;
  streaming?: boolean;
  reactions?: Record<string, number>; // emoji -> count
}

export interface ServiceHealth {
  name: string;
  status: 'running' | 'warning' | 'down';
  health: number; // 0-100
  response_time?: number; // milliseconds
  error?: string;
  port?: number;
  version?: string;
  uptime?: string;
  last_check: string;
}

export interface HealthCheckResponse {
  timestamp: string;
  services: ServiceHealth[];
  overall: {
    status: 'healthy' | 'degraded' | 'unhealthy';
    healthy: number;
    total: number;
    score: number;
  };
}

export interface JournalHighlight {
  type: 'quest_completed' | 'quest_failed' | 'level_up' | 'achievement' | 'collaboration' | 'status_update';
  title: string;
  description: string;
  time?: string;
  icon?: string;
}

export interface JournalStats {
  quests_completed: number;
  quests_failed: number;
  tokens_used: number;
  cost_incurred: number;
  xp_gained: number;
  collaborations: number;
  tools_used: number;
}

export interface RPGJournal {
  agent_id: string;
  agent_name: string;
  date: string;
  title: string;
  narrative: string;
  highlights: JournalHighlight[];
  stats: JournalStats;
  generated_at: string;
}

export interface TavernTalkMessage {
  type: string;
  speaker: string;
  target: string;
  message: string;
  timestamp: string;
  mood: 'friendly' | 'competitive' | 'supportive' | 'appreciative' | 'casual';
  context: string;
}

// ── Store ──────────────────────────────────────────────────

interface StoreState {
  agents: RPGAgent[];
  quests: RPGQuest[];
  parties: RPGParty[];
  stats: RPGStats | null;
  connected: boolean;
  pollTimer: ReturnType<typeof setInterval> | null;
  wsUnsubscribe: (() => void) | null;
  
  // Loading states
  isLoadingAgents: boolean;
  isLoadingQuests: boolean;
  isLoadingParties: boolean;
  
  // Health check
  healthData: HealthCheckResponse | null;
  isLoadingHealth: boolean;
  fetchHealthData: () => Promise<void>;
  isLoadingStats: boolean;
  hasInitialLoad: boolean;

  // Chat state per agent
  chatMessages: Record<string, ChatMessage[]>;
  chatLoading: Record<string, boolean>;

  // Selected agent for chat panel
  selectedAgent: string | null;
  setSelectedAgent: (id: string | null) => void;

  // Theme state
  theme: 'light' | 'dark';
  setTheme: (theme: 'light' | 'dark') => void;
  toggleTheme: () => void;

  // Seasonal events state
  seasonalEvent: SeasonalEvent | null;
  seasonalDecorations: string[];
  seasonalXpMultiplier: number;
  seasonalEnabled: boolean;
  updateSeasonalEvent: () => void;
  toggleSeasonalEvents: () => void;

  // Level-up detection
  previousAgentLevels: Record<string, number>;
  levelUpTriggers: Record<string, number>; // timestamp of level up for animation trigger
  triggerLevelUp: (agentId: string) => void;

  // Quest completion detection
  previousQuestStatuses: Record<number, string>;
  questCompletionTriggers: Record<number, { trigger: number; questName: string; xpReward: number; agentName: string }>;
  triggerQuestCompletion: (questId: number, questName: string, xpReward: number, agentName: string) => void;

  // Achievement notifications
  achievementToasts: Array<{
    id: string;
    name: string;
    description: string;
    icon: string;
    timestamp: number;
  }>;
  previousAgentAchievements: Record<string, Set<string>>; // agentId -> Set of achievement IDs
  addAchievementToast: (achievement: { id: string; name: string; description: string; icon: string }) => void;
  removeAchievementToast: (timestamp: number) => void;
  checkForNewAchievements: () => Promise<void>;

  // Tavern Talk state
  tavernTalkMessages: TavernTalkMessage[];
  isLoadingTavernTalk: boolean;
  lastTavernTalkUpdate: number;

  // Actions
  fetchAll: () => Promise<void>;
  connectWebSocket: () => void;
  disconnectWebSocket: () => void;
  startPolling: (intervalMs?: number) => void;
  stopPolling: () => void;
  sendMessage: (agentId: string, message: string) => Promise<void>;
  addChatMessage: (agentId: string, msg: ChatMessage) => void;
  clearChatHistory: (agentId?: string) => void;
  addReaction: (agentId: string, messageIndex: number, emoji: string) => void;
  fetchTavernTalk: () => Promise<void>;
}

const API_URL = import.meta.env.VITE_API_URL || '';
const WS_URL = import.meta.env.VITE_WS_URL || `${location.protocol === 'https:' ? 'wss:' : 'ws:'}//${location.host}/ws`;

export const useStore = create<StoreState>((set, get) => ({
  agents: [],
  quests: [],
  parties: [],
  stats: null,
  connected: false,
  pollTimer: null,
  wsUnsubscribe: null,
  chatMessages: loadChatHistory(), // Load persisted chat history on initialization
  chatLoading: {},
  selectedAgent: null,
  previousAgentLevels: {},
  levelUpTriggers: {},
  previousQuestStatuses: {},
  questCompletionTriggers: {},
  
  // Achievement notification state defaults
  achievementToasts: [],
  previousAgentAchievements: {},
  
  // Loading state defaults
  isLoadingAgents: false,
  isLoadingQuests: false,
  isLoadingParties: false,
  isLoadingStats: false,
  hasInitialLoad: false,
  
  // Health check defaults
  healthData: null,
  isLoadingHealth: false,

  // Tavern Talk defaults
  tavernTalkMessages: [],
  isLoadingTavernTalk: false,
  lastTavernTalkUpdate: 0,

  // Initialize theme from localStorage or default to dark
  theme: (localStorage.getItem('theme') as 'light' | 'dark') || 'dark',

  setSelectedAgent: (id) => set({ selectedAgent: id }),

  setTheme: (theme) => {
    localStorage.setItem('theme', theme);
    document.documentElement.setAttribute('data-theme', theme);
    set({ theme });
  },

  toggleTheme: () => {
    const currentTheme = get().theme;
    const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
    get().setTheme(newTheme);
  },

  // Initialize seasonal events
  seasonalEvent: null,
  seasonalDecorations: [],
  seasonalXpMultiplier: 1,
  seasonalEnabled: localStorage.getItem('seasonalEnabled') !== 'false', // Default to enabled

  updateSeasonalEvent: () => {
    const { seasonalEnabled } = get();
    if (!seasonalEnabled) {
      set({
        seasonalEvent: null,
        seasonalDecorations: [],
        seasonalXpMultiplier: 1
      });
      removeSeasonalTheme();
      return;
    }

    const activeEvent = getActiveSeasonalEvent();
    const decorations = getSeasonalDecorations();
    const xpMultiplier = getSeasonalXpMultiplier();
    
    set({ 
      seasonalEvent: activeEvent,
      seasonalDecorations: decorations,
      seasonalXpMultiplier: xpMultiplier
    });
    
    // Apply theme changes
    if (activeEvent) {
      applySeasonalTheme(activeEvent);
    } else {
      removeSeasonalTheme();
    }
  },

  toggleSeasonalEvents: () => {
    const { seasonalEnabled } = get();
    const newEnabled = !seasonalEnabled;
    
    localStorage.setItem('seasonalEnabled', newEnabled.toString());
    set({ seasonalEnabled: newEnabled });
    
    // Update seasonal event state immediately
    get().updateSeasonalEvent();
  },

  triggerLevelUp: (agentId) => set((state) => ({
    levelUpTriggers: { ...state.levelUpTriggers, [agentId]: Date.now() }
  })),

  triggerQuestCompletion: (questId, questName, xpReward, agentName) => set((state) => ({
    questCompletionTriggers: { 
      ...state.questCompletionTriggers, 
      [questId]: { trigger: Date.now(), questName, xpReward, agentName }
    }
  })),

  addAchievementToast: (achievement) => set((state) => {
    const timestamp = Date.now();
    const uniqueId = `${achievement.id}-${timestamp}-${Math.random().toString(36).substr(2, 9)}`;
    return {
      achievementToasts: [...state.achievementToasts, {
        ...achievement,
        id: uniqueId, // Override with guaranteed unique ID
        timestamp
      }]
    };
  }),

  removeAchievementToast: (timestamp) => set((state) => ({
    achievementToasts: state.achievementToasts.filter(toast => toast.timestamp !== timestamp)
  })),

  checkForNewAchievements: async () => {
    const { agents, previousAgentAchievements, addAchievementToast } = get();
    
    // Check achievements for each agent
    for (const agent of agents) {
      try {
        const achievementsRes = await fetch(`${API_URL}/api/inber/achievements?agent=${encodeURIComponent(agent.id)}`);
        if (!achievementsRes.ok) continue;
        
        const currentAchievements: RPGAchievement[] = await achievementsRes.json();
        const currentAchievementIds = new Set(currentAchievements.map(a => a.id));
        const previousAchievementIds = previousAgentAchievements[agent.id] || new Set();
        
        // Find newly unlocked achievements
        const newAchievements = currentAchievements.filter(achievement => 
          !previousAchievementIds.has(achievement.id)
        );
        
        // Show toast for each new achievement
        newAchievements.forEach(achievement => {
          console.log(`${agent.name} unlocked achievement: ${achievement.name}`);
          addAchievementToast({
            id: `${agent.id}_${achievement.id}`,  // Make ID unique per agent
            name: achievement.name,
            description: achievement.description,
            icon: achievement.icon
          });
        });
        
        // Update the stored achievements for this agent
        set((state) => ({
          previousAgentAchievements: {
            ...state.previousAgentAchievements,
            [agent.id]: currentAchievementIds
          }
        }));
        
      } catch (error) {
        console.warn(`Failed to check achievements for agent ${agent.id}:`, error);
      }
    }
  },

  fetchAll: async () => {
    const { hasInitialLoad } = get();
    
    // Only show loading states on initial load, not on subsequent polls
    if (!hasInitialLoad) {
      set({ 
        isLoadingAgents: true, 
        isLoadingQuests: true, 
        isLoadingParties: true,
        isLoadingStats: true 
      });
    }

    try {
      const [agentsRes, questsRes, partiesRes, statsRes] = await Promise.all([
        fetch(`${API_URL}/api/inber/agents`),
        fetch(`${API_URL}/api/inber/quests?limit=100`),
        fetch(`${API_URL}/api/parties`),
        fetch(`${API_URL}/api/inber/stats`),
      ]);
      
      if (agentsRes.ok) {
        const newAgents = await agentsRes.json();
        const { previousAgentLevels, triggerLevelUp } = get();
        
        // Check for level ups
        newAgents.forEach((agent: RPGAgent) => {
          const prevLevel = previousAgentLevels[agent.id];
          if (prevLevel && agent.level > prevLevel) {
            console.log(`${agent.name} leveled up from ${prevLevel} to ${agent.level}!`);
            triggerLevelUp(agent.id);
          }
        });
        
        // Update agents and track their levels
        const newLevels = newAgents.reduce((acc: Record<string, number>, agent: RPGAgent) => {
          acc[agent.id] = agent.level;
          return acc;
        }, {});
        
        set({ 
          agents: newAgents,
          previousAgentLevels: newLevels,
          isLoadingAgents: false
        });
        
        // Check for new achievements after agents are updated
        get().checkForNewAchievements();
      } else {
        set({ isLoadingAgents: false });
      }
      
      if (questsRes.ok) {
        const newQuests = await questsRes.json();
        const { previousQuestStatuses, triggerQuestCompletion } = get();
        
        // Check for quest completions
        newQuests.forEach((quest: RPGQuest) => {
          const prevStatus = previousQuestStatuses[quest.id];
          if (prevStatus && prevStatus !== 'completed' && quest.status === 'completed') {
            console.log(`Quest "${quest.name}" completed by ${quest.assigned_agent_name || 'Unknown Agent'}!`);
            triggerQuestCompletion(quest.id, quest.name, quest.xp_reward, quest.assigned_agent_name || 'Unknown Agent');
          }
        });
        
        // Update quests and track their statuses
        const newStatuses = newQuests.reduce((acc: Record<number, string>, quest: RPGQuest) => {
          acc[quest.id] = quest.status;
          return acc;
        }, {});
        
        set({ 
          quests: newQuests,
          previousQuestStatuses: newStatuses,
          isLoadingQuests: false
        });
      } else {
        set({ isLoadingQuests: false });
      }
      
      if (partiesRes.ok) {
        const newParties = await partiesRes.json();
        set({ 
          parties: newParties,
          isLoadingParties: false
        });
      } else {
        set({ isLoadingParties: false });
      }
      
      if (statsRes.ok) {
        set({ stats: await statsRes.json(), isLoadingStats: false });
      } else {
        set({ isLoadingStats: false });
      }
      
      // Mark initial load as complete
      if (!hasInitialLoad) {
        set({ hasInitialLoad: true });
      }
    } catch {
      console.warn('Failed to fetch data');
      set({ 
        isLoadingAgents: false, 
        isLoadingQuests: false, 
        isLoadingParties: false,
        isLoadingStats: false,
        hasInitialLoad: true 
      });
    }
  },

  fetchHealthData: async () => {
    const { isLoadingHealth } = get();
    if (isLoadingHealth) return; // Prevent multiple concurrent requests
    
    set({ isLoadingHealth: true });
    
    try {
      const healthRes = await fetch(`${API_URL}/api/health`);
      if (healthRes.ok) {
        const healthData: HealthCheckResponse = await healthRes.json();
        set({ 
          healthData,
          isLoadingHealth: false 
        });
      } else {
        console.warn('Failed to fetch health data');
        set({ isLoadingHealth: false });
      }
    } catch (error) {
      console.warn('Health check failed:', error);
      set({ isLoadingHealth: false });
    }
  },

  connectWebSocket: () => {
    // Disconnect existing connection if any
    get().disconnectWebSocket();
    
    const handleMessage = (data: unknown) => {
      if (!isWebSocketMessage(data)) {
        console.warn('Received invalid WebSocket message:', data);
        return;
      }
      
      const msg = data as WebSocketMessage;
      if (msg.type === 'quest_event') {
        // Handle real-time quest events
        const questEvent = msg.data;
        console.log(`🎯 Quest Event: ${questEvent.message}`);
        
        // Play sound effects for important events
        switch (questEvent.type) {
          case 'quest_completed':
            console.log(`🎉 ${questEvent.message}`);
            // The quest completion sound/animation is already handled by the existing triggerQuestCompletion logic
            break;
          case 'quest_failed':
            console.log(`💀 ${questEvent.message}`);
            break;
          case 'quest_started':
            console.log(`⚔️ ${questEvent.message}`);
            break;
          case 'agent_level_up':
            console.log(`✨ ${questEvent.message}`);
            // The level up animation is already handled by the existing triggerLevelUp logic
            break;
          case 'agent_status_changed':
            console.log(`📍 ${questEvent.message}`);
            break;
        }
        
        // Trigger a fresh data fetch to get the latest state after the event
        // This ensures the UI stays in sync with the real-time events
        get().fetchAll();
      } else if (msg.type === 'inber_update') {
        const { agents, quests, stats } = msg.data;
        if (agents) {
          const { previousAgentLevels, triggerLevelUp } = get();
          
          // Check for level ups
          agents.forEach((agent: RPGAgent) => {
            const prevLevel = previousAgentLevels[agent.id];
            if (prevLevel && agent.level > prevLevel) {
              console.log(`${agent.name} leveled up from ${prevLevel} to ${agent.level}!`);
              triggerLevelUp(agent.id);
            }
          });
          
          // Update agents and track their levels
          const newLevels = agents.reduce((acc: Record<string, number>, agent: RPGAgent) => {
            acc[agent.id] = agent.level;
            return acc;
          }, {});
          
          set({ 
            agents,
            previousAgentLevels: newLevels
          });
          
          // Check for new achievements after agents are updated via WebSocket
          get().checkForNewAchievements();
        }
        if (quests) {
          const { previousQuestStatuses, triggerQuestCompletion } = get();
          
          // Check for quest completions
          quests.forEach((quest: RPGQuest) => {
            const prevStatus = previousQuestStatuses[quest.id];
            if (prevStatus && prevStatus !== 'completed' && quest.status === 'completed') {
              console.log(`Quest "${quest.name}" completed by ${quest.assigned_agent_name || 'Unknown Agent'}!`);
              triggerQuestCompletion(quest.id, quest.name, quest.xp_reward, quest.assigned_agent_name || 'Unknown Agent');
            }
          });
          
          // Update quest statuses
          const newStatuses = quests.reduce((acc: Record<number, string>, quest: RPGQuest) => {
            acc[quest.id] = quest.status;
            return acc;
          }, {});
          
          set({ 
            quests,
            previousQuestStatuses: newStatuses
          });
        }
        if (stats && typeof stats === 'object' && stats !== null) {
          set({ stats: stats as RPGStats });
        }
      }
    };

    const handleStateChange = (connected: boolean) => {
      set({ connected });
    };

    // Use the optimized WebSocket manager
    const unsubscribe = wsManager.subscribe(
      WS_URL,
      'main-store',
      handleMessage,
      handleStateChange
    );
    
    set({ wsUnsubscribe: unsubscribe });
  },

  disconnectWebSocket: () => {
    const { wsUnsubscribe } = get();
    if (wsUnsubscribe) {
      wsUnsubscribe();
      set({ wsUnsubscribe: null, connected: false });
    }
  },

  startPolling: (intervalMs = 10000) => {
    if (get().pollTimer) return;
    get().fetchAll();
    const timer = setInterval(() => get().fetchAll(), intervalMs);
    set({ pollTimer: timer });
  },

  stopPolling: () => {
    const t = get().pollTimer;
    if (t) { clearInterval(t); set({ pollTimer: null }); }
  },

  addChatMessage: (agentId, msg) => set((state) => {
    const newMessages = [...(state.chatMessages[agentId] || []), msg];
    const updatedChatMessages = {
      ...state.chatMessages,
      [agentId]: newMessages,
    };
    
    // Persist to localStorage (skip streaming messages as they're temporary)
    if (!msg.streaming) {
      saveChatHistoryForAgent(agentId, newMessages);
    }
    
    return { chatMessages: updatedChatMessages };
  }),

  sendMessage: async (agentId, message) => {
    const { addChatMessage } = get();
    addChatMessage(agentId, {
      role: 'user',
      content: message,
      timestamp: new Date().toISOString(),
    });

    set((s) => ({ chatLoading: { ...s.chatLoading, [agentId]: true } }));

    try {
      const res = await fetch(`${API_URL}/api/run`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Accept: 'text/event-stream' },
        body: JSON.stringify({ agent: agentId, message }),
      });

      if (!res.ok) {
        addChatMessage(agentId, {
          role: 'system',
          content: `Error: ${res.status} ${res.statusText}`,
          timestamp: new Date().toISOString(),
        });
        return;
      }

      // Handle SSE streaming
      const reader = res.body?.getReader();
      if (!reader) return;
      const decoder = new TextDecoder();
      let buffer = '';
      let assistantMsg = '';

      // Add placeholder message
      const msgIdx = (get().chatMessages[agentId] || []).length;
      addChatMessage(agentId, {
        role: 'assistant',
        content: '',
        timestamp: new Date().toISOString(),
        streaming: true,
      });

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        buffer += decoder.decode(value, { stream: true });

        // Parse SSE lines
        const lines = buffer.split('\n');
        buffer = lines.pop() || '';
        for (const line of lines) {
          if (line.startsWith('data: ')) {
            const data = line.slice(6);
            if (data === '[DONE]') continue;
            try {
              const parsed = JSON.parse(data);
              if (parsed.content) {
                assistantMsg += parsed.content;
              } else if (parsed.text) {
                assistantMsg += parsed.text;
              } else if (typeof parsed === 'string') {
                assistantMsg += parsed;
              }
            } catch {
              // Plain text SSE
              assistantMsg += data;
            }
            // Update the streaming message
            set((s) => {
              const msgs = [...(s.chatMessages[agentId] || [])];
              if (msgs[msgIdx]) {
                msgs[msgIdx] = { ...msgs[msgIdx], content: assistantMsg };
              }
              return { chatMessages: { ...s.chatMessages, [agentId]: msgs } };
            });
          }
        }
      }

      // Mark streaming complete and persist
      set((s) => {
        const msgs = [...(s.chatMessages[agentId] || [])];
        if (msgs[msgIdx]) {
          msgs[msgIdx] = { ...msgs[msgIdx], streaming: false, content: assistantMsg || '(no response)' };
          
          // Persist the completed message
          saveChatHistoryForAgent(agentId, msgs);
        }
        return { chatMessages: { ...s.chatMessages, [agentId]: msgs } };
      });
    } catch (err) {
      addChatMessage(agentId, {
        role: 'system',
        content: `Connection error: ${err}`,
        timestamp: new Date().toISOString(),
      });
    } finally {
      set((s) => ({ chatLoading: { ...s.chatLoading, [agentId]: false } }));
    }
  },

  clearChatHistory: (agentId) => set((state) => {
    if (agentId) {
      // Clear history for specific agent
      const updatedChatMessages = { ...state.chatMessages };
      delete updatedChatMessages[agentId];
      saveChatHistory(updatedChatMessages);
      return { chatMessages: updatedChatMessages };
    } else {
      // Clear all chat history
      saveChatHistory({});
      return { chatMessages: {} };
    }
  }),

  addReaction: (agentId, messageIndex, emoji) => set((state) => {
    const messages = [...(state.chatMessages[agentId] || [])];
    const message = messages[messageIndex];
    
    if (message && !message.streaming) {
      const reactions = message.reactions || {};
      const newCount = (reactions[emoji] || 0) + 1;
      
      // Toggle reaction: if count is 1, remove it; otherwise increment
      const newReactions = { ...reactions };
      if (newCount === 1) {
        newReactions[emoji] = 1;
      } else {
        // Remove reaction when clicking again
        delete newReactions[emoji];
      }
      
      messages[messageIndex] = { 
        ...message, 
        reactions: Object.keys(newReactions).length > 0 ? newReactions : undefined
      };
      
      const updatedChatMessages = {
        ...state.chatMessages,
        [agentId]: messages,
      };
      
      // Persist to localStorage
      saveChatHistoryForAgent(agentId, messages);
      
      return { chatMessages: updatedChatMessages };
    }
    
    return state;
  }),

  fetchTavernTalk: async () => {
    const { isLoadingTavernTalk, lastTavernTalkUpdate } = get();
    
    // Throttle requests - don't fetch more than once every 30 seconds
    const now = Date.now();
    if (isLoadingTavernTalk || (now - lastTavernTalkUpdate < 30000)) {
      return;
    }
    
    set({ isLoadingTavernTalk: true });
    
    try {
      const response = await fetch(`${API_URL}/api/tavern-talk`);
      if (response.ok) {
        const tavernTalkMessages: TavernTalkMessage[] = await response.json();
        set({ 
          tavernTalkMessages,
          isLoadingTavernTalk: false,
          lastTavernTalkUpdate: now
        });
      } else {
        console.warn('Failed to fetch tavern talk');
        set({ isLoadingTavernTalk: false });
      }
    } catch (error) {
      console.warn('Tavern talk fetch failed:', error);
      set({ 
        tavernTalkMessages: [], // Clear any existing messages on error
        isLoadingTavernTalk: false 
      });
    }
  },
}));

// ── Helpers ────────────────────────────────────────────────

export function formatTokens(n: number) {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}k`;
  return String(n);
}

export function formatCost(c: number) {
  if (c <= 0) return '$0';
  if (c < 0.01) return `$${c.toFixed(4)}`;
  return `$${c.toFixed(2)}`;
}

export function timeAgo(dateStr?: string) {
  if (!dateStr) return '—';
  try {
    const diffMin = Math.floor((Date.now() - new Date(dateStr).getTime()) / 60000);
    if (diffMin < 1) return 'just now';
    if (diffMin < 60) return `${diffMin}m ago`;
    const h = Math.floor(diffMin / 60);
    if (h < 24) return `${h}h ago`;
    return `${Math.floor(h / 24)}d ago`;
  } catch { return dateStr; }
}

// Class-specific colors
export const CLASS_COLORS: Record<string, string> = {
  Overlord: '#ff6b6b',
  Artificer: '#a78bfa',
  Engineer: '#60a5fa',
  Bard: '#fbbf24',
  Ranger: '#34d399',
  Scribe: '#f0abfc',
  Shadow: '#94a3b8',
  Smith: '#fb923c',
  Gladiator: '#f87171',
  Sovereign: '#ffd700',
  Adventurer: '#d4af37',
  Wizard: '#a78bfa',
  Healer: '#4ade80',
  Warrior: '#f87171',
  Scout: '#8b5cf6',
  Cleric: '#4ade80',
  Jester: '#f472b6',
  Sage: '#818cf8',
  Tinker: '#a78bfa',
  Sentinel: '#38bdf8',
};

export function classColor(cls: string) {
  return CLASS_COLORS[cls] || '#d4af37';
}

// Calculate quest difficulty based on token usage
export function calculateQuestDifficulty(tokens: number): number {
  if (tokens >= 100000) return 5; // ★★★★★ Very Hard
  if (tokens >= 20000) return 4;  // ★★★★ Hard  
  if (tokens >= 5000) return 3;   // ★★★ Medium
  if (tokens >= 1000) return 2;   // ★★ Easy
  return 1;                       // ★ Very Easy
}

// Get difficulty stars string
export function getDifficultyStars(tokens: number): string {
  const difficulty = calculateQuestDifficulty(tokens);
  return '★'.repeat(difficulty);
}

// Get difficulty name for tooltip
export function getDifficultyName(tokens: number): string {
  const difficulty = calculateQuestDifficulty(tokens);
  const names = ['Very Easy', 'Easy', 'Medium', 'Hard', 'Very Hard'];
  return names[difficulty - 1] || 'Unknown';
}
