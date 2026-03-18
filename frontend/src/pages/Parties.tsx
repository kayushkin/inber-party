import { useState } from 'react';
import { useStore } from '../store';

interface CreatePartyForm {
  name: string;
  description: string;
  leaderId: number;
  memberIds: number[];
}

export default function Parties() {
  const { parties, agents, isLoadingParties } = useStore();
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [createForm, setCreateForm] = useState<CreatePartyForm>({
    name: '',
    description: '',
    leaderId: 0,
    memberIds: []
  });

  const handleCreateParty = async (e: React.FormEvent) => {
    e.preventDefault();
    
    try {
      const response = await fetch('/api/parties', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: createForm.name,
          description: createForm.description,
          leader_id: createForm.leaderId,
          member_ids: createForm.memberIds
        })
      });

      if (response.ok) {
        setShowCreateForm(false);
        setCreateForm({ name: '', description: '', leaderId: 0, memberIds: [] });
        // Refresh data - the store polling should pick it up
      } else {
        console.error('Failed to create party');
      }
    } catch (error) {
      console.error('Error creating party:', error);
    }
  };

  const handleMemberToggle = (agentId: number) => {
    setCreateForm(prev => ({
      ...prev,
      memberIds: prev.memberIds.includes(agentId)
        ? prev.memberIds.filter(id => id !== agentId)
        : [...prev.memberIds, agentId]
    }));
  };

  // Task assignment functionality to be implemented

  if (isLoadingParties && parties.length === 0) {
    return (
      <div className="p-6">
        <div className="animate-pulse">
          <div className="h-8 bg-gray-300 rounded w-1/4 mb-6"></div>
          <div className="grid gap-4">
            {[...Array(3)].map((_, i) => (
              <div key={i} className="h-32 bg-gray-300 rounded"></div>
            ))}
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">🏰 Guild Parties</h1>
        <button
          onClick={() => setShowCreateForm(true)}
          className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg transition-colors"
        >
          ⚔️ Form New Party
        </button>
      </div>

      {/* Create Party Modal */}
      {showCreateForm && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white dark:bg-gray-800 p-6 rounded-lg max-w-md w-full mx-4">
            <h2 className="text-xl font-bold mb-4">Create New Party</h2>
            <form onSubmit={handleCreateParty}>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium mb-1">Party Name</label>
                  <input
                    type="text"
                    value={createForm.name}
                    onChange={(e) => setCreateForm(prev => ({ ...prev, name: e.target.value }))}
                    className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600"
                    placeholder="The Fellowship"
                    required
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium mb-1">Description</label>
                  <textarea
                    value={createForm.description}
                    onChange={(e) => setCreateForm(prev => ({ ...prev, description: e.target.value }))}
                    className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600"
                    placeholder="A party of brave adventurers..."
                    rows={3}
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium mb-1">Leader</label>
                  <select
                    value={createForm.leaderId}
                    onChange={(e) => setCreateForm(prev => ({ ...prev, leaderId: parseInt(e.target.value) }))}
                    className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600"
                    required
                  >
                    <option value={0}>Select a leader...</option>
                    {agents.map(agent => (
                      <option key={agent.id} value={parseInt(agent.id)}>
                        {agent.name} - Level {agent.level} {agent.class}
                      </option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium mb-1">Members</label>
                  <div className="space-y-2 max-h-40 overflow-y-auto">
                    {agents.map(agent => (
                      <label key={agent.id} className="flex items-center space-x-2">
                        <input
                          type="checkbox"
                          checked={createForm.memberIds.includes(parseInt(agent.id)) || parseInt(agent.id) === createForm.leaderId}
                          onChange={() => handleMemberToggle(parseInt(agent.id))}
                          disabled={parseInt(agent.id) === createForm.leaderId}
                          className="rounded"
                        />
                        <span className="text-sm">
                          {agent.avatar_emoji} {agent.name} - Level {agent.level} {agent.class}
                        </span>
                      </label>
                    ))}
                  </div>
                </div>
              </div>

              <div className="flex space-x-3 mt-6">
                <button
                  type="submit"
                  className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg transition-colors"
                >
                  Create Party
                </button>
                <button
                  type="button"
                  onClick={() => setShowCreateForm(false)}
                  className="bg-gray-600 hover:bg-gray-700 text-white px-4 py-2 rounded-lg transition-colors"
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Parties List */}
      <div className="grid gap-6">
        {parties.length === 0 ? (
          <div className="text-center py-12">
            <div className="text-6xl mb-4">🏰</div>
            <h2 className="text-xl font-semibold mb-2">No Parties Formed Yet</h2>
            <p className="text-gray-600 dark:text-gray-400">
              Create your first party to tackle multi-agent quests together!
            </p>
          </div>
        ) : (
          parties.map(party => (
            <div key={party.id} className="bg-white dark:bg-gray-800 rounded-lg p-6 border">
              <div className="flex items-center justify-between mb-4">
                <div>
                  <h3 className="text-xl font-bold">{party.name}</h3>
                  <p className="text-gray-600 dark:text-gray-400">{party.description}</p>
                </div>
                <div className="flex items-center space-x-2">
                  <span className={`px-2 py-1 rounded-full text-xs font-medium ${
                    party.status === 'active' ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200' :
                    party.status === 'on_quest' ? 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200' :
                    'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200'
                  }`}>
                    {party.status === 'on_quest' ? '🗡️ On Quest' : 
                     party.status === 'active' ? '🏰 Active' : '💤 Inactive'}
                  </span>
                  <span className="text-sm text-gray-500">
                    {party.member_count} members
                  </span>
                </div>
              </div>

              <div className="flex items-center justify-between">
                <div>
                  <span className="text-sm text-gray-600 dark:text-gray-400">Leader:</span>
                  <span className="ml-2 font-medium">
                    {party.leader?.name || 'Unknown'} ({party.leader?.class || 'Unknown Class'})
                  </span>
                </div>
                <div className="space-x-2">
                  <button
                    onClick={() => {/* Navigate to party detail */}}
                    className="bg-gray-600 hover:bg-gray-700 text-white px-3 py-1 rounded text-sm transition-colors"
                  >
                    👥 View Details
                  </button>
                  <button
                    onClick={() => {/* Show task assignment modal */}}
                    className="bg-blue-600 hover:bg-blue-700 text-white px-3 py-1 rounded text-sm transition-colors"
                  >
                    📜 Assign Quest
                  </button>
                </div>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}