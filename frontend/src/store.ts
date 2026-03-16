import { create } from 'zustand';

export interface Agent {
  id: number;
  name: string;
  title: string;
  class: string;
  level: number;
  xp: number;
  energy: number;
  status: 'idle' | 'working' | 'on_quest' | 'stuck' | 'resting';
  avatar_emoji: string;
  created_at: string;
  updated_at: string;
}

export interface AgentDetail extends Agent {
  skills: Skill[];
  achievements: Achievement[];
  tasks: Task[];
}

export interface Task {
  id: number;
  name: string;
  description: string;
  difficulty: string;
  xp_reward: number;
  status: 'available' | 'in_progress' | 'completed' | 'failed';
  assigned_agent_id?: number;
  progress: number;
  created_at: string;
  started_at?: string;
  completed_at?: string;
}

export interface Skill {
  id: number;
  agent_id: number;
  skill_name: string;
  level: number;
  task_count: number;
}

export interface Achievement {
  id: number;
  agent_id: number;
  achievement_name: string;
  unlocked_at: string;
}

export interface Stats {
  total_agents: number;
  active_tasks: number;
  completed_tasks: number;
  total_xp: number;
  average_agent_level: number;
}

interface StoreState {
  agents: Agent[];
  tasks: Task[];
  inberQuests: InberQuest[];
  inberAgents: InberAgent[];
  stats: Stats | null;
  ws: WebSocket | null;
  connected: boolean;
  pollTimer: ReturnType<typeof setInterval> | null;
  
  setAgents: (agents: Agent[]) => void;
  setTasks: (tasks: Task[]) => void;
  setInberQuests: (quests: InberQuest[]) => void;
  setInberAgents: (agents: InberAgent[]) => void;
  setStats: (stats: Stats) => void;
  updateAgent: (id: number, updates: Partial<Agent>) => void;
  updateTask: (id: number, updates: Partial<Task>) => void;
  addAgent: (agent: Agent) => void;
  addTask: (task: Task) => void;
  
  connectWebSocket: () => void;
  disconnectWebSocket: () => void;
  startPolling: (intervalMs?: number) => void;
  stopPolling: () => void;
}

const API_URL = import.meta.env.VITE_API_URL || '';
const WS_URL = import.meta.env.VITE_WS_URL || `${location.protocol === 'https:' ? 'wss:' : 'ws:'}//${location.host}/ws`;

// Demo/mock data for when the backend is unavailable
const MOCK_AGENTS: Agent[] = [
  { id: 1, name: 'Bran', title: 'the Methodical', class: 'Wizard', level: 12, xp: 1200, energy: 85, status: 'working', avatar_emoji: '🧙', created_at: '2026-03-01T00:00:00Z', updated_at: '2026-03-15T12:00:00Z' },
  { id: 2, name: 'Scáthach', title: 'the Swift', class: 'Ranger', level: 15, xp: 1500, energy: 62, status: 'idle', avatar_emoji: '🏹', created_at: '2026-03-01T00:00:00Z', updated_at: '2026-03-15T10:00:00Z' },
  { id: 3, name: 'Aoife', title: 'the Bold', class: 'Warrior', level: 9, xp: 900, energy: 100, status: 'on_quest', avatar_emoji: '⚔️', created_at: '2026-03-01T00:00:00Z', updated_at: '2026-03-15T14:00:00Z' },
];

const MOCK_TASKS: Task[] = [
  { id: 1, name: 'Refactor the Ancient Scrolls', description: 'Restructure the legacy codebase module', difficulty: '3', xp_reward: 50, status: 'in_progress', assigned_agent_id: 1, progress: 67, created_at: '2026-03-14T09:00:00Z', started_at: '2026-03-14T09:30:00Z' },
  { id: 2, name: 'Scout the Dark Forest', description: 'Explore and document the new API endpoints', difficulty: '2', xp_reward: 30, status: 'in_progress', assigned_agent_id: 3, progress: 42, created_at: '2026-03-14T11:00:00Z', started_at: '2026-03-14T11:15:00Z' },
  { id: 3, name: 'Forge the Iron Tests', description: 'Write comprehensive test suite for auth module', difficulty: '4', xp_reward: 80, status: 'completed', assigned_agent_id: 2, progress: 100, created_at: '2026-03-13T08:00:00Z', started_at: '2026-03-13T08:30:00Z', completed_at: '2026-03-13T16:00:00Z' },
  { id: 4, name: 'Enchant the Build Pipeline', description: 'Set up CI/CD with proper caching', difficulty: '2', xp_reward: 25, status: 'available', progress: 0, created_at: '2026-03-15T07:00:00Z' },
];

const MOCK_AGENT_DETAILS: Record<number, AgentDetail> = {
  1: { ...MOCK_AGENTS[0], skills: [{ id: 1, agent_id: 1, skill_name: 'Code Weaving', level: 8, task_count: 24 }, { id: 2, agent_id: 1, skill_name: 'Arcane Debugging', level: 6, task_count: 15 }, { id: 3, agent_id: 1, skill_name: 'Scroll Reading', level: 5, task_count: 12 }], achievements: [{ id: 1, agent_id: 1, achievement_name: 'First Quest', unlocked_at: '2026-03-02T10:00:00Z' }, { id: 2, agent_id: 1, achievement_name: '100 Spells Cast', unlocked_at: '2026-03-10T15:00:00Z' }], tasks: [MOCK_TASKS[0], MOCK_TASKS[2]] },
  2: { ...MOCK_AGENTS[1], skills: [{ id: 4, agent_id: 2, skill_name: 'Swift Execution', level: 10, task_count: 30 }, { id: 5, agent_id: 2, skill_name: 'Path Finding', level: 7, task_count: 18 }], achievements: [{ id: 3, agent_id: 2, achievement_name: 'First Quest', unlocked_at: '2026-03-02T08:00:00Z' }, { id: 4, agent_id: 2, achievement_name: 'Speed Demon', unlocked_at: '2026-03-08T12:00:00Z' }, { id: 5, agent_id: 2, achievement_name: 'Perfectionist', unlocked_at: '2026-03-12T18:00:00Z' }], tasks: [MOCK_TASKS[2]] },
  3: { ...MOCK_AGENTS[2], skills: [{ id: 6, agent_id: 3, skill_name: 'Brute Force', level: 4, task_count: 10 }, { id: 7, agent_id: 3, skill_name: 'Shield Bash', level: 3, task_count: 7 }], achievements: [{ id: 6, agent_id: 3, achievement_name: 'First Quest', unlocked_at: '2026-03-03T09:00:00Z' }], tasks: [MOCK_TASKS[1]] },
};

const MOCK_STATS: Stats = { total_agents: 3, active_tasks: 2, completed_tasks: 1, total_xp: 3600, average_agent_level: 12 };

let demoMode = false;

export const useStore = create<StoreState>((set, get) => ({
  agents: [],
  tasks: [],
  inberQuests: [],
  inberAgents: [],
  stats: null,
  ws: null,
  connected: false,
  pollTimer: null,

  setAgents: (agents) => set({ agents }),
  setTasks: (tasks) => set({ tasks }),
  setInberQuests: (quests) => set({ inberQuests: quests }),
  setInberAgents: (agents) => set({ inberAgents: agents }),
  setStats: (stats) => set({ stats }),
  
  updateAgent: (id, updates) => set((state) => ({
    agents: state.agents.map((a) => (a.id === id ? { ...a, ...updates } : a)),
  })),
  
  updateTask: (id, updates) => set((state) => ({
    tasks: state.tasks.map((t) => (t.id === id ? { ...t, ...updates } : t)),
  })),
  
  addAgent: (agent) => set((state) => ({
    agents: [...state.agents, agent],
  })),
  
  addTask: (task) => set((state) => ({
    tasks: [task, ...state.tasks],
  })),

  connectWebSocket: () => {
    if (demoMode) return;
    const ws = new WebSocket(WS_URL);
    
    ws.onopen = () => {
      console.log('✅ WebSocket connected');
      set({ connected: true });
    };
    
    ws.onclose = () => {
      console.log('❌ WebSocket disconnected');
      set({ connected: false });
      if (!demoMode) {
        console.log('Reconnecting in 3s...');
        setTimeout(() => {
          get().connectWebSocket();
        }, 3000);
      }
    };
    
    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
    
    ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        console.log('📡 WebSocket message:', message);
        
        switch (message.type) {
          case 'agent_created':
            get().addAgent(message.data);
            break;
          case 'agent_updated':
            get().updateAgent(message.data.id, message.data.updates);
            break;
          case 'task_created':
            get().addTask(message.data);
            break;
          case 'task_updated':
            get().updateTask(message.data.id, message.data.updates);
            break;
        }
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    };
    
    set({ ws });
  },
  
  disconnectWebSocket: () => {
    const { ws } = get();
    if (ws) {
      ws.close();
      set({ ws: null, connected: false });
    }
  },

  startPolling: (intervalMs = 10000) => {
    const { pollTimer } = get();
    if (pollTimer) return; // already polling

    const poll = async () => {
      try {
        const [agents, tasks, stats] = await Promise.all([
          api.getAgents(),
          api.getTasks(),
          api.getStats(),
        ]);
        get().setAgents(agents);
        get().setTasks(tasks);
        get().setStats(stats);
      } catch {
        // silent — will retry next interval
      }
    };

    // Do an initial poll immediately
    poll();
    const timer = setInterval(poll, intervalMs);
    set({ pollTimer: timer });
  },

  stopPolling: () => {
    const { pollTimer } = get();
    if (pollTimer) {
      clearInterval(pollTimer);
      set({ pollTimer: null });
    }
  },
}));

// Helper: fetch with demo-mode fallback
async function fetchOrDemo<T>(url: string, fallback: T, init?: RequestInit): Promise<T> {
  if (demoMode) return fallback;
  try {
    const res = await fetch(url, init);
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    if (res.status === 204) return fallback;
    return res.json();
  } catch {
    console.warn('⚡ Backend unavailable — switching to demo mode');
    demoMode = true;
    return fallback;
  }
}

// Inber RPG types (from real agent data)
export interface InberAgent {
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
  avatar_emoji: string;
  total_tokens: number;
  total_cost: number;
  session_count: number;
  quest_count: number;
  error_count: number;
  skills: { skill_name: string; level: number; task_count: number }[];
  last_active?: string;
}

export interface InberQuest {
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

export interface InberStats {
  total_agents: number;
  active_quests: number;
  completed_quests: number;
  failed_quests: number;
  total_xp: number;
  total_tokens: number;
  total_cost: number;
  average_agent_level: number;
}

// Convert inber agent to our Agent type
function inberAgentToAgent(ia: InberAgent, idx: number): Agent {
  return {
    id: idx + 1,
    name: ia.name,
    title: ia.title,
    class: ia.class,
    level: ia.level,
    xp: ia.xp,
    energy: ia.energy,
    status: (ia.status as Agent['status']) || 'idle',
    avatar_emoji: ia.avatar_emoji,
    created_at: ia.last_active || new Date().toISOString(),
    updated_at: ia.last_active || new Date().toISOString(),
  };
}

function inberAgentToDetail(ia: InberAgent, idx: number): AgentDetail {
  return {
    ...inberAgentToAgent(ia, idx),
    skills: (ia.skills || []).map((s, i) => ({
      id: i + 1,
      agent_id: idx + 1,
      skill_name: s.skill_name,
      level: s.level,
      task_count: s.task_count,
    })),
    achievements: [],
    tasks: [],
  };
}

function inberQuestToTask(iq: InberQuest): Task {
  return {
    id: iq.id,
    name: iq.name,
    description: iq.description,
    difficulty: String(iq.difficulty),
    xp_reward: iq.xp_reward,
    status: (iq.status as Task['status']) || 'available',
    assigned_agent_id: undefined, // string IDs don't map directly
    progress: iq.progress,
    created_at: iq.created_at,
    started_at: iq.started_at,
    completed_at: iq.completed_at,
  };
}

let inberMode = false;
let inberAgents: InberAgent[] = [];

// API helper functions — tries inber endpoints first, then legacy, then demo
export const api = {
  async getAgents(): Promise<Agent[]> {
    // Try inber endpoint first
    if (!demoMode) {
      try {
        const res = await fetch(`${API_URL}/api/inber/agents`);
        if (res.ok) {
          inberAgents = await res.json();
          inberMode = true;
          // Also store raw inber agents in zustand
          useStore.getState().setInberAgents(inberAgents);
          return inberAgents.map(inberAgentToAgent);
        }
      } catch { /* fall through */ }
    }
    return fetchOrDemo(`${API_URL}/api/agents`, MOCK_AGENTS);
  },

  async getAgent(id: number): Promise<AgentDetail> {
    if (inberMode && inberAgents.length > 0) {
      const ia = inberAgents[id - 1] || inberAgents.find(a => a.name.toLowerCase() === String(id));
      if (ia) return inberAgentToDetail(ia, id - 1);
    }
    return fetchOrDemo(`${API_URL}/api/agents/${id}`, MOCK_AGENT_DETAILS[id] || { ...MOCK_AGENTS[0], skills: [], achievements: [], tasks: [] });
  },

  async createAgent(agent: Partial<Agent>): Promise<Agent> {
    return fetchOrDemo(`${API_URL}/api/agents`, { ...MOCK_AGENTS[0], ...agent, id: Date.now() } as Agent, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(agent),
    });
  },

  async updateAgent(id: number, updates: Partial<Agent>): Promise<void> {
    await fetchOrDemo(`${API_URL}/api/agents/${id}`, undefined, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(updates),
    });
  },

  async getTasks(): Promise<Task[]> {
    // Try inber quests endpoint
    if (!demoMode) {
      try {
        const res = await fetch(`${API_URL}/api/inber/quests`);
        if (res.ok) {
          const quests: InberQuest[] = await res.json();
          inberMode = true;
          // Store raw inber quests in zustand for rich display
          useStore.getState().setInberQuests(quests);
          return quests.map(inberQuestToTask);
        }
      } catch { /* fall through */ }
    }
    return fetchOrDemo(`${API_URL}/api/tasks`, MOCK_TASKS);
  },

  async createTask(task: Partial<Task>): Promise<Task> {
    return fetchOrDemo(`${API_URL}/api/tasks`, { ...MOCK_TASKS[0], ...task, id: Date.now() } as Task, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(task),
    });
  },

  async updateTask(id: number, updates: Partial<Task>): Promise<void> {
    await fetchOrDemo(`${API_URL}/api/tasks/${id}`, undefined, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(updates),
    });
  },

  async getStats(): Promise<Stats> {
    // Try inber stats endpoint
    if (!demoMode) {
      try {
        const res = await fetch(`${API_URL}/api/inber/stats`);
        if (res.ok) {
          const s: InberStats = await res.json();
          inberMode = true;
          return {
            total_agents: s.total_agents,
            active_tasks: s.active_quests,
            completed_tasks: s.completed_quests,
            total_xp: s.total_xp,
            average_agent_level: s.average_agent_level,
          };
        }
      } catch { /* fall through */ }
    }
    return fetchOrDemo(`${API_URL}/api/stats`, MOCK_STATS);
  },

  async getAgentQuests(agentId: string): Promise<InberQuest[]> {
    if (!demoMode) {
      try {
        const res = await fetch(`${API_URL}/api/inber/quests?agent=${encodeURIComponent(agentId)}&limit=100`);
        if (res.ok) return res.json();
      } catch { /* fall through */ }
    }
    return [];
  },

  isDemoMode(): boolean {
    return demoMode;
  },

  isInberMode(): boolean {
    return inberMode;
  },
};
