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
  stats: Stats | null;
  ws: WebSocket | null;
  connected: boolean;
  
  setAgents: (agents: Agent[]) => void;
  setTasks: (tasks: Task[]) => void;
  setStats: (stats: Stats) => void;
  updateAgent: (id: number, updates: Partial<Agent>) => void;
  updateTask: (id: number, updates: Partial<Task>) => void;
  addAgent: (agent: Agent) => void;
  addTask: (task: Task) => void;
  
  connectWebSocket: () => void;
  disconnectWebSocket: () => void;
}

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws';

export const useStore = create<StoreState>((set, get) => ({
  agents: [],
  tasks: [],
  stats: null,
  ws: null,
  connected: false,

  setAgents: (agents) => set({ agents }),
  setTasks: (tasks) => set({ tasks }),
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
    const ws = new WebSocket(WS_URL);
    
    ws.onopen = () => {
      console.log('✅ WebSocket connected');
      set({ connected: true });
    };
    
    ws.onclose = () => {
      console.log('❌ WebSocket disconnected, reconnecting in 3s...');
      set({ connected: false });
      setTimeout(() => {
        get().connectWebSocket();
      }, 3000);
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
}));

// API helper functions
export const api = {
  async getAgents(): Promise<Agent[]> {
    const res = await fetch(`${API_URL}/api/agents`);
    return res.json();
  },
  
  async getAgent(id: number): Promise<AgentDetail> {
    const res = await fetch(`${API_URL}/api/agents/${id}`);
    return res.json();
  },
  
  async createAgent(agent: Partial<Agent>): Promise<Agent> {
    const res = await fetch(`${API_URL}/api/agents`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(agent),
    });
    return res.json();
  },
  
  async updateAgent(id: number, updates: Partial<Agent>): Promise<void> {
    await fetch(`${API_URL}/api/agents/${id}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(updates),
    });
  },
  
  async getTasks(): Promise<Task[]> {
    const res = await fetch(`${API_URL}/api/tasks`);
    return res.json();
  },
  
  async createTask(task: Partial<Task>): Promise<Task> {
    const res = await fetch(`${API_URL}/api/tasks`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(task),
    });
    return res.json();
  },
  
  async updateTask(id: number, updates: Partial<Task>): Promise<void> {
    await fetch(`${API_URL}/api/tasks/${id}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(updates),
    });
  },
  
  async getStats(): Promise<Stats> {
    const res = await fetch(`${API_URL}/api/stats`);
    return res.json();
  },
};
