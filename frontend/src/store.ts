import { create } from 'zustand';

// ── Types ──────────────────────────────────────────────────

export interface RPGAgent {
  id: string;
  name: string;
  title: string;
  class: string;
  level: number;
  xp: number;
  xp_to_next: number;
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
}

export interface RPGQuest {
  id: number;
  name: string;
  description: string;
  difficulty: number;
  xp_reward: number;
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

export interface ChatMessage {
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: string;
  streaming?: boolean;
}

// ── Store ──────────────────────────────────────────────────

interface StoreState {
  agents: RPGAgent[];
  quests: RPGQuest[];
  stats: RPGStats | null;
  connected: boolean;
  ws: WebSocket | null;
  pollTimer: ReturnType<typeof setInterval> | null;

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

  // Level-up detection
  previousAgentLevels: Record<string, number>;
  levelUpTriggers: Record<string, number>; // timestamp of level up for animation trigger
  triggerLevelUp: (agentId: string) => void;

  // Quest completion detection
  previousQuestStatuses: Record<number, string>;
  questCompletionTriggers: Record<number, { trigger: number; questName: string; xpReward: number; agentName: string }>;
  triggerQuestCompletion: (questId: number, questName: string, xpReward: number, agentName: string) => void;

  // Actions
  fetchAll: () => Promise<void>;
  connectWebSocket: () => void;
  disconnectWebSocket: () => void;
  startPolling: (intervalMs?: number) => void;
  stopPolling: () => void;
  sendMessage: (agentId: string, message: string) => Promise<void>;
  addChatMessage: (agentId: string, msg: ChatMessage) => void;
}

const API_URL = import.meta.env.VITE_API_URL || '';
const WS_URL = import.meta.env.VITE_WS_URL || `${location.protocol === 'https:' ? 'wss:' : 'ws:'}//${location.host}/ws`;

export const useStore = create<StoreState>((set, get) => ({
  agents: [],
  quests: [],
  stats: null,
  connected: false,
  ws: null,
  pollTimer: null,
  chatMessages: {},
  chatLoading: {},
  selectedAgent: null,
  previousAgentLevels: {},
  levelUpTriggers: {},
  previousQuestStatuses: {},
  questCompletionTriggers: {},

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

  triggerLevelUp: (agentId) => set((state) => ({
    levelUpTriggers: { ...state.levelUpTriggers, [agentId]: Date.now() }
  })),

  triggerQuestCompletion: (questId, questName, xpReward, agentName) => set((state) => ({
    questCompletionTriggers: { 
      ...state.questCompletionTriggers, 
      [questId]: { trigger: Date.now(), questName, xpReward, agentName }
    }
  })),

  fetchAll: async () => {
    try {
      const [agentsRes, questsRes, statsRes] = await Promise.all([
        fetch(`${API_URL}/api/inber/agents`),
        fetch(`${API_URL}/api/inber/quests?limit=100`),
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
          previousAgentLevels: newLevels
        });
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
          previousQuestStatuses: newStatuses
        });
      }
      if (statsRes.ok) set({ stats: await statsRes.json() });
    } catch {
      console.warn('Failed to fetch data');
    }
  },

  connectWebSocket: () => {
    const ws = new WebSocket(WS_URL);
    ws.onopen = () => set({ connected: true });
    ws.onclose = () => {
      set({ connected: false });
      setTimeout(() => get().connectWebSocket(), 3000);
    };
    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        if (msg.type === 'inber_update') {
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
          if (stats) set({ stats });
        }
      } catch { /* ignore */ }
    };
    set({ ws });
  },

  disconnectWebSocket: () => {
    get().ws?.close();
    set({ ws: null, connected: false });
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

  addChatMessage: (agentId, msg) => set((state) => ({
    chatMessages: {
      ...state.chatMessages,
      [agentId]: [...(state.chatMessages[agentId] || []), msg],
    },
  })),

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

      // Mark streaming complete
      set((s) => {
        const msgs = [...(s.chatMessages[agentId] || [])];
        if (msgs[msgIdx]) {
          msgs[msgIdx] = { ...msgs[msgIdx], streaming: false, content: assistantMsg || '(no response)' };
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
