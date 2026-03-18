import { apiClient, useApiClient } from './apiClient';
import type { RPGAgent, RPGQuest, RPGAchievement, QuestHistoryEntry } from '../store';

// Agent-related API calls
export class AgentService {
  // Get all agents
  static async getAgents() {
    return apiClient.get<RPGAgent[]>('/api/agents');
  }

  // Get specific agent
  static async getAgent(id: string) {
    return apiClient.get<RPGAgent>(`/api/agents/${encodeURIComponent(id)}`);
  }

  // Get agent quests
  static async getAgentQuests(agentId: string, limit: number = 100) {
    return apiClient.get<RPGQuest[]>('/api/inber/quests', { 
      agent: agentId, 
      limit: limit.toString() 
    });
  }

  // Get agent achievements
  static async getAgentAchievements(agentId: string) {
    return apiClient.get<RPGAchievement[]>('/api/inber/achievements', { 
      agent: agentId 
    });
  }

  // Get agent quest history
  static async getAgentQuestHistory(agentId: string, limit: number = 20) {
    return apiClient.get<QuestHistoryEntry[]>('/api/inber/quest-history', { 
      agent: agentId, 
      limit: limit.toString() 
    });
  }

  // Create new agent
  static async createAgent(agentData: Partial<RPGAgent>) {
    return apiClient.post<RPGAgent>('/api/agents', agentData);
  }

  // Update agent
  static async updateAgent(id: string, updates: Partial<RPGAgent>) {
    return apiClient.put<RPGAgent>(`/api/agents/${encodeURIComponent(id)}`, updates);
  }

  // Delete agent
  static async deleteAgent(id: string) {
    return apiClient.delete<{ success: boolean }>(`/api/agents/${encodeURIComponent(id)}`);
  }
}

// React hooks for agent operations
export function useAgentOperations() {
  const apiHook = useApiClient();

  const loadAgents = async () => {
    return apiHook.makeApiCall(
      () => AgentService.getAgents(),
      'Failed to load agents'
    );
  };

  const loadAgent = async (id: string) => {
    return apiHook.makeApiCall(
      () => AgentService.getAgent(id),
      `Failed to load agent ${id}`
    );
  };

  const loadAgentQuests = async (agentId: string, limit?: number) => {
    return apiHook.makeApiCall(
      () => AgentService.getAgentQuests(agentId, limit),
      `Failed to load quests for ${agentId}`
    );
  };

  const loadAgentAchievements = async (agentId: string) => {
    return apiHook.makeApiCall(
      () => AgentService.getAgentAchievements(agentId),
      `Failed to load achievements for ${agentId}`
    );
  };

  const loadAgentQuestHistory = async (agentId: string, limit?: number) => {
    return apiHook.makeApiCall(
      () => AgentService.getAgentQuestHistory(agentId, limit),
      `Failed to load quest history for ${agentId}`
    );
  };

  const createAgent = async (agentData: Partial<RPGAgent>) => {
    return apiHook.makeApiCall(
      () => AgentService.createAgent(agentData),
      'Failed to create agent'
    );
  };

  const updateAgent = async (id: string, updates: Partial<RPGAgent>) => {
    return apiHook.makeApiCall(
      () => AgentService.updateAgent(id, updates),
      `Failed to update agent ${id}`
    );
  };

  const deleteAgent = async (id: string) => {
    return apiHook.makeApiCall(
      () => AgentService.deleteAgent(id),
      `Failed to delete agent ${id}`
    );
  };

  return {
    ...apiHook,
    loadAgents,
    loadAgent,
    loadAgentQuests,
    loadAgentAchievements,
    loadAgentQuestHistory,
    createAgent,
    updateAgent,
    deleteAgent
  };
}

export default AgentService;