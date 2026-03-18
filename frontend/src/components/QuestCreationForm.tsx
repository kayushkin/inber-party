import { useState } from 'react';
import { useStore } from '../store';
import './QuestCreationForm.css';

interface QuestCreationFormProps {
  isOpen: boolean;
  onClose: () => void;
  onQuestCreated: () => void;
}

export default function QuestCreationForm({ isOpen, onClose, onQuestCreated }: QuestCreationFormProps) {
  const agents = useStore((s) => s.agents);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    difficulty: 'Easy',
    xp_reward: 50,
    assigned_agent_id: '' as string,
    execute_immediately: false
  });

  const difficultyOptions = [
    { value: 'Easy', xp: 50 },
    { value: 'Medium', xp: 100 },
    { value: 'Hard', xp: 200 },
    { value: 'Epic', xp: 400 },
    { value: 'Legendary', xp: 800 }
  ];

  if (!isOpen) return null;

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value, type } = e.target;
    
    if (name === 'difficulty') {
      const difficulty = difficultyOptions.find(d => d.value === value);
      setFormData({
        ...formData,
        difficulty: value,
        xp_reward: difficulty?.xp || 50
      });
    } else if (type === 'checkbox') {
      setFormData({
        ...formData,
        [name]: (e.target as HTMLInputElement).checked
      });
    } else {
      setFormData({
        ...formData,
        [name]: value
      });
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);

    try {
      // Create local quest/task
      const taskPayload = {
        name: formData.name,
        description: formData.description,
        difficulty: formData.difficulty,
        xp_reward: formData.xp_reward
      };

      const taskResponse = await fetch('/api/tasks', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(taskPayload)
      });

      if (!taskResponse.ok) {
        throw new Error(`Failed to create quest: ${taskResponse.statusText}`);
      }

      const createdTask = await taskResponse.json();

      // If assigned to an agent and immediate execution requested
      if (formData.assigned_agent_id && formData.execute_immediately) {
        const agent = agents.find(a => a.id === formData.assigned_agent_id);
        
        if (agent) {
          // Update task to assign to agent and start it
          await fetch(`/api/tasks/${createdTask.id}`, {
            method: 'PATCH',
            headers: {
              'Content-Type': 'application/json'
            },
            body: JSON.stringify({
              assigned_agent_id: parseInt(formData.assigned_agent_id),
              status: 'in_progress'
            })
          });

          // Execute via inber API
          const inberPayload = {
            agentId: agent.id,
            task: formData.description,
            timeoutSeconds: 300
          };

          // Note: This is a fire-and-forget call to start the actual execution
          // The quest status will be updated through the normal inber integration
          fetch('/api/run', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json'
            },
            body: JSON.stringify(inberPayload)
          }).catch(err => {
            console.warn('Failed to trigger immediate execution:', err);
          });
        }
      }

      // Reset form
      setFormData({
        name: '',
        description: '',
        difficulty: 'Easy',
        xp_reward: 50,
        assigned_agent_id: '',
        execute_immediately: false
      });

      onQuestCreated();
      onClose();
    } catch (error) {
      console.error('Error creating quest:', error);
      alert('Failed to create quest. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="quest-form-overlay" onClick={onClose}>
      <div className="quest-form-modal" onClick={(e) => e.stopPropagation()}>
        <div className="quest-form-header">
          <h2>🎯 Create New Quest</h2>
          <button className="close-btn" onClick={onClose} aria-label="Close">×</button>
        </div>

        <form onSubmit={handleSubmit} className="quest-form">
          <div className="form-group">
            <label htmlFor="name">Quest Name *</label>
            <input
              type="text"
              id="name"
              name="name"
              value={formData.name}
              onChange={handleInputChange}
              placeholder="Enter a descriptive quest name..."
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="description">Quest Description *</label>
            <textarea
              id="description"
              name="description"
              value={formData.description}
              onChange={handleInputChange}
              placeholder="Describe what this quest should accomplish..."
              rows={4}
              required
            />
          </div>

          <div className="form-row">
            <div className="form-group">
              <label htmlFor="difficulty">Difficulty</label>
              <select
                id="difficulty"
                name="difficulty"
                value={formData.difficulty}
                onChange={handleInputChange}
              >
                {difficultyOptions.map(option => (
                  <option key={option.value} value={option.value}>
                    {option.value} ({option.xp} XP)
                  </option>
                ))}
              </select>
            </div>

            <div className="form-group">
              <label htmlFor="xp_reward">XP Reward</label>
              <input
                type="number"
                id="xp_reward"
                name="xp_reward"
                value={formData.xp_reward}
                onChange={handleInputChange}
                min="1"
                max="2000"
              />
            </div>
          </div>

          <div className="form-group">
            <label htmlFor="assigned_agent_id">Assign to Agent (Optional)</label>
            <select
              id="assigned_agent_id"
              name="assigned_agent_id"
              value={formData.assigned_agent_id}
              onChange={handleInputChange}
            >
              <option value="">Select an agent...</option>
              {agents.map(agent => (
                <option key={agent.id} value={agent.id}>
                  {agent.name} ({agent.class}, Level {agent.level})
                </option>
              ))}
            </select>
          </div>

          {formData.assigned_agent_id && (
            <div className="form-group">
              <label className="checkbox-label">
                <input
                  type="checkbox"
                  name="execute_immediately"
                  checked={formData.execute_immediately}
                  onChange={handleInputChange}
                />
                Execute quest immediately
              </label>
              <small>If checked, the assigned agent will start working on this quest right away via the Inber API.</small>
            </div>
          )}

          <div className="form-actions">
            <button type="button" className="btn btn-secondary" onClick={onClose}>
              Cancel
            </button>
            <button type="submit" className="btn btn-primary" disabled={isSubmitting}>
              {isSubmitting ? 'Creating...' : 'Create Quest'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}