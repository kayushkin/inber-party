import { useEffect, useState } from 'react';
import { useStore, api } from '../store';
import './QuestBoard.css';

export default function QuestBoard() {
  const tasks = useStore((state) => state.tasks);
  const agents = useStore((state) => state.agents);
  const setTasks = useStore((state) => state.setTasks);
  const setAgents = useStore((state) => state.setAgents);
  
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newTask, setNewTask] = useState({
    name: '',
    description: '',
    difficulty: 'medium',
    xp_reward: 50,
  });

  useEffect(() => {
    Promise.all([
      api.getTasks().then(setTasks),
      api.getAgents().then(setAgents),
    ]);
  }, [setTasks, setAgents]);

  const handleCreateTask = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await api.createTask(newTask);
      setShowCreateModal(false);
      setNewTask({
        name: '',
        description: '',
        difficulty: 'medium',
        xp_reward: 50,
      });
      // Refresh tasks
      api.getTasks().then(setTasks);
    } catch (error) {
      console.error('Failed to create task:', error);
    }
  };

  const handleAssignTask = async (taskId: number, agentId: number) => {
    try {
      await api.updateTask(taskId, {
        assigned_agent_id: agentId,
        status: 'in_progress',
      });
      // Refresh tasks
      api.getTasks().then(setTasks);
    } catch (error) {
      console.error('Failed to assign task:', error);
    }
  };

  const handleCompleteTask = async (taskId: number) => {
    try {
      await api.updateTask(taskId, {
        status: 'completed',
        progress: 100,
      });
      // Refresh tasks
      api.getTasks().then(setTasks);
    } catch (error) {
      console.error('Failed to complete task:', error);
    }
  };

  const getAgent = (agentId?: number) => {
    return agents.find((a) => a.id === agentId);
  };

  const getDifficultyColor = (difficulty: string) => {
    switch (difficulty.toLowerCase()) {
      case 'easy':
        return '#4ade80';
      case 'medium':
        return '#fbbf24';
      case 'hard':
        return '#f87171';
      default:
        return '#888';
    }
  };

  return (
    <div className="quest-board">
      <div className="quest-board-header">
        <h1>Quest Board</h1>
        <button className="create-quest-button" onClick={() => setShowCreateModal(true)}>
          + Create Quest
        </button>
      </div>

      <div className="quests-grid">
        {tasks.map((task) => {
          const assignedAgent = getAgent(task.assigned_agent_id);
          
          return (
            <div key={task.id} className={`quest-card status-${task.status}`}>
              <div className="quest-card-header">
                <h3 className="quest-card-name">{task.name}</h3>
                <div
                  className="quest-difficulty"
                  style={{ borderColor: getDifficultyColor(task.difficulty) }}
                >
                  {task.difficulty}
                </div>
              </div>

              <p className="quest-card-description">{task.description}</p>

              <div className="quest-card-footer">
                <div className="quest-reward">
                  <span className="reward-icon">⭐</span>
                  +{task.xp_reward} XP
                </div>

                {task.status === 'available' && (
                  <div className="quest-assign">
                    <select
                      className="assign-select"
                      onChange={(e) => {
                        const agentId = parseInt(e.target.value);
                        if (agentId) {
                          handleAssignTask(task.id, agentId);
                        }
                      }}
                      defaultValue=""
                    >
                      <option value="" disabled>
                        Assign to...
                      </option>
                      {agents.map((agent) => (
                        <option key={agent.id} value={agent.id}>
                          {agent.avatar_emoji} {agent.name}
                        </option>
                      ))}
                    </select>
                  </div>
                )}

                {task.status === 'in_progress' && assignedAgent && (
                  <div className="quest-assigned">
                    <div className="assigned-agent">
                      <span className="assigned-avatar">{assignedAgent.avatar_emoji}</span>
                      <span className="assigned-name">{assignedAgent.name}</span>
                    </div>
                    <div className="quest-progress-bar">
                      <div
                        className="quest-progress-fill"
                        style={{ width: `${task.progress}%` }}
                      />
                    </div>
                    <button
                      className="complete-button"
                      onClick={() => handleCompleteTask(task.id)}
                    >
                      Complete
                    </button>
                  </div>
                )}

                {task.status === 'completed' && assignedAgent && (
                  <div className="quest-completed">
                    <span className="completed-icon">✅</span>
                    Completed by {assignedAgent.name}
                  </div>
                )}

                {task.status === 'failed' && (
                  <div className="quest-failed">
                    <span className="failed-icon">❌</span>
                    Quest Failed
                  </div>
                )}
              </div>
            </div>
          );
        })}

        {tasks.length === 0 && (
          <div className="no-quests">
            <p>No quests available. Create one to get started!</p>
          </div>
        )}
      </div>

      {showCreateModal && (
        <div className="modal-overlay" onClick={() => setShowCreateModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h2>Create New Quest</h2>
            <form onSubmit={handleCreateTask}>
              <div className="form-group">
                <label>Quest Name</label>
                <input
                  type="text"
                  value={newTask.name}
                  onChange={(e) => setNewTask({ ...newTask, name: e.target.value })}
                  required
                  placeholder="Enter quest name..."
                />
              </div>

              <div className="form-group">
                <label>Description</label>
                <textarea
                  value={newTask.description}
                  onChange={(e) => setNewTask({ ...newTask, description: e.target.value })}
                  required
                  placeholder="Describe the quest..."
                  rows={4}
                />
              </div>

              <div className="form-row">
                <div className="form-group">
                  <label>Difficulty</label>
                  <select
                    value={newTask.difficulty}
                    onChange={(e) => setNewTask({ ...newTask, difficulty: e.target.value })}
                  >
                    <option value="easy">Easy</option>
                    <option value="medium">Medium</option>
                    <option value="hard">Hard</option>
                  </select>
                </div>

                <div className="form-group">
                  <label>XP Reward</label>
                  <input
                    type="number"
                    value={newTask.xp_reward}
                    onChange={(e) =>
                      setNewTask({ ...newTask, xp_reward: parseInt(e.target.value) })
                    }
                    min="10"
                    step="10"
                  />
                </div>
              </div>

              <div className="modal-actions">
                <button type="button" onClick={() => setShowCreateModal(false)} className="cancel-button">
                  Cancel
                </button>
                <button type="submit" className="submit-button">
                  Create Quest
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
