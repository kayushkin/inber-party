import { apiClient, useApiClient } from './apiClient';
import type { RPGQuest, QuestStatus } from '../store';

export interface CreateQuestData {
  title: string;
  description: string;
  assignedTo?: string;
  priority?: number;
  tags?: string[];
  estimatedTokens?: number;
}

export interface QuestFilter {
  status?: QuestStatus;
  assignedTo?: string;
  priority?: number;
  tags?: string[];
  limit?: number;
  offset?: number;
}

// Quest-related API calls
export class QuestService {
  // Get all quests with optional filtering
  static async getQuests(filter: QuestFilter = {}) {
    const params: Record<string, string> = {};
    
    if (filter.status) params.status = filter.status;
    if (filter.assignedTo) params.assignedTo = filter.assignedTo;
    if (filter.priority !== undefined) params.priority = filter.priority.toString();
    if (filter.tags) params.tags = filter.tags.join(',');
    if (filter.limit) params.limit = filter.limit.toString();
    if (filter.offset) params.offset = filter.offset.toString();

    return apiClient.get<RPGQuest[]>('/api/quests', params);
  }

  // Get specific quest
  static async getQuest(id: string) {
    return apiClient.get<RPGQuest>(`/api/quests/${encodeURIComponent(id)}`);
  }

  // Create new quest
  static async createQuest(questData: CreateQuestData) {
    return apiClient.post<RPGQuest>('/api/quests', questData);
  }

  // Update quest
  static async updateQuest(id: string, updates: Partial<RPGQuest>) {
    return apiClient.put<RPGQuest>(`/api/quests/${encodeURIComponent(id)}`, updates);
  }

  // Delete quest
  static async deleteQuest(id: string) {
    return apiClient.delete<{ success: boolean }>(`/api/quests/${encodeURIComponent(id)}`);
  }

  // Assign quest to agent
  static async assignQuest(questId: string, agentId: string) {
    return apiClient.post<RPGQuest>(`/api/quests/${encodeURIComponent(questId)}/assign`, {
      agentId
    });
  }

  // Complete quest
  static async completeQuest(questId: string, results?: any) {
    return apiClient.post<RPGQuest>(`/api/quests/${encodeURIComponent(questId)}/complete`, {
      results
    });
  }

  // Cancel quest
  static async cancelQuest(questId: string, reason?: string) {
    return apiClient.post<RPGQuest>(`/api/quests/${encodeURIComponent(questId)}/cancel`, {
      reason
    });
  }

  // Get quest logs
  static async getQuestLogs(questId: string) {
    return apiClient.get<any[]>(`/api/quests/${encodeURIComponent(questId)}/logs`);
  }
}

// React hooks for quest operations
export function useQuestOperations() {
  const apiHook = useApiClient();

  const loadQuests = async (filter?: QuestFilter) => {
    return apiHook.makeApiCall(
      () => QuestService.getQuests(filter),
      'Failed to load quests'
    );
  };

  const loadQuest = async (id: string) => {
    return apiHook.makeApiCall(
      () => QuestService.getQuest(id),
      `Failed to load quest ${id}`
    );
  };

  const createQuest = async (questData: CreateQuestData) => {
    return apiHook.makeApiCall(
      () => QuestService.createQuest(questData),
      'Failed to create quest'
    );
  };

  const updateQuest = async (id: string, updates: Partial<RPGQuest>) => {
    return apiHook.makeApiCall(
      () => QuestService.updateQuest(id, updates),
      `Failed to update quest ${id}`
    );
  };

  const deleteQuest = async (id: string) => {
    return apiHook.makeApiCall(
      () => QuestService.deleteQuest(id),
      `Failed to delete quest ${id}`
    );
  };

  const assignQuest = async (questId: string, agentId: string) => {
    return apiHook.makeApiCall(
      () => QuestService.assignQuest(questId, agentId),
      `Failed to assign quest ${questId} to ${agentId}`
    );
  };

  const completeQuest = async (questId: string, results?: any) => {
    return apiHook.makeApiCall(
      () => QuestService.completeQuest(questId, results),
      `Failed to complete quest ${questId}`
    );
  };

  const cancelQuest = async (questId: string, reason?: string) => {
    return apiHook.makeApiCall(
      () => QuestService.cancelQuest(questId, reason),
      `Failed to cancel quest ${questId}`
    );
  };

  const loadQuestLogs = async (questId: string) => {
    return apiHook.makeApiCall(
      () => QuestService.getQuestLogs(questId),
      `Failed to load logs for quest ${questId}`
    );
  };

  return {
    ...apiHook,
    loadQuests,
    loadQuest,
    createQuest,
    updateQuest,
    deleteQuest,
    assignQuest,
    completeQuest,
    cancelQuest,
    loadQuestLogs
  };
}

export default QuestService;