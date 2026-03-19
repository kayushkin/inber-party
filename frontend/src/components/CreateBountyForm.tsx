import React, { useState, useEffect } from 'react';
import './CreateBountyForm.css';

interface CreateBountyFormData {
  title: string;
  description: string;
  requirements: string;
  payout_amount: number;
  creator_id: number;
  deadline: string;
  required_skills: string[];
}

interface Agent {
  id: number;
  name: string;
  title: string;
  class: string;
}

interface CreateBountyFormProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (bounty: CreateBountyFormData) => Promise<void>;
}

export default function CreateBountyForm({ isOpen, onClose, onSubmit }: CreateBountyFormProps) {
  const [form, setForm] = useState<CreateBountyFormData>({
    title: '',
    description: '',
    requirements: '',
    payout_amount: 50,
    creator_id: 0,
    deadline: '',
    required_skills: []
  });
  const [agents, setAgents] = useState<Agent[]>([]);
  const [skillInput, setSkillInput] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (isOpen) {
      fetchAgents();
    }
  }, [isOpen]);

  const fetchAgents = async () => {
    try {
      const response = await fetch('/api/agents');
      if (response.ok) {
        const agentData = await response.json();
        setAgents(agentData);
        // Auto-select first agent if none selected
        if (form.creator_id === 0 && agentData.length > 0) {
          setForm(prev => ({ ...prev, creator_id: agentData[0].id }));
        }
      }
    } catch (error) {
      console.error('Error fetching agents:', error);
    }
  };

  const handleInputChange = (field: keyof CreateBountyFormData, value: any) => {
    setForm(prev => ({ ...prev, [field]: value }));
    if (error) setError(''); // Clear error when user types
  };

  const addSkill = () => {
    if (skillInput.trim() && !form.required_skills.includes(skillInput.trim())) {
      setForm(prev => ({
        ...prev,
        required_skills: [...prev.required_skills, skillInput.trim()]
      }));
      setSkillInput('');
    }
  };

  const removeSkill = (skillToRemove: string) => {
    setForm(prev => ({
      ...prev,
      required_skills: prev.required_skills.filter(skill => skill !== skillToRemove)
    }));
  };

  const getTierForPayout = (payout: number) => {
    if (payout < 50) return 'bronze';
    if (payout < 200) return 'silver';
    if (payout < 500) return 'gold';
    return 'legendary';
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    // Validation
    if (!form.title.trim()) {
      setError('Title is required');
      return;
    }
    if (!form.description.trim()) {
      setError('Description is required');
      return;
    }
    if (form.payout_amount <= 0) {
      setError('Payout amount must be positive');
      return;
    }
    if (form.creator_id === 0) {
      setError('Please select a creator');
      return;
    }

    setIsSubmitting(true);
    setError('');

    try {
      await onSubmit(form);
      // Reset form on success
      setForm({
        title: '',
        description: '',
        requirements: '',
        payout_amount: 50,
        creator_id: agents.length > 0 ? agents[0].id : 0,
        deadline: '',
        required_skills: []
      });
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create bounty');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCancel = () => {
    setError('');
    onClose();
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      addSkill();
    }
  };

  if (!isOpen) return null;

  const tier = getTierForPayout(form.payout_amount);
  const tierIcons = {
    bronze: '🥉',
    silver: '🥈', 
    gold: '🥇',
    legendary: '💎'
  };

  return (
    <div className="modal-overlay" onClick={handleCancel}>
      <div className="modal-content create-bounty-modal" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2>💰 Create New Bounty</h2>
          <button className="modal-close" onClick={handleCancel}>×</button>
        </div>

        <form onSubmit={handleSubmit} className="create-bounty-form">
          <div className="form-section">
            <div className="form-group">
              <label className="form-label">Title *</label>
              <input
                type="text"
                className="form-input"
                value={form.title}
                onChange={(e) => handleInputChange('title', e.target.value)}
                placeholder="e.g., Fix critical bug in payment system"
                maxLength={100}
              />
            </div>

            <div className="form-group">
              <label className="form-label">Description *</label>
              <textarea
                className="form-textarea"
                value={form.description}
                onChange={(e) => handleInputChange('description', e.target.value)}
                placeholder="Detailed description of the task..."
                rows={4}
                maxLength={1000}
              />
              <div className="form-help">{form.description.length}/1000 characters</div>
            </div>

            <div className="form-group">
              <label className="form-label">Requirements / Acceptance Criteria</label>
              <textarea
                className="form-textarea"
                value={form.requirements}
                onChange={(e) => handleInputChange('requirements', e.target.value)}
                placeholder="What needs to be done for this bounty to be complete? List specific requirements..."
                rows={3}
                maxLength={500}
              />
              <div className="form-help">Optional: Specific criteria for completion</div>
            </div>
          </div>

          <div className="form-section">
            <div className="form-row">
              <div className="form-group">
                <label className="form-label">Creator *</label>
                <select
                  className="form-select"
                  value={form.creator_id}
                  onChange={(e) => handleInputChange('creator_id', parseInt(e.target.value))}
                >
                  <option value={0}>Select creator...</option>
                  {agents.map(agent => (
                    <option key={agent.id} value={agent.id}>
                      {agent.name} - {agent.title}
                    </option>
                  ))}
                </select>
              </div>

              <div className="form-group">
                <label className="form-label">
                  Payout Amount * 
                  <span className="tier-indicator">
                    {tierIcons[tier]} {tier}
                  </span>
                </label>
                <div className="payout-input">
                  <span className="currency-icon">🪙</span>
                  <input
                    type="number"
                    className="form-input"
                    value={form.payout_amount}
                    onChange={(e) => handleInputChange('payout_amount', parseInt(e.target.value) || 0)}
                    min={1}
                    max={10000}
                  />
                  <span className="currency-label">Gold</span>
                </div>
              </div>
            </div>

            <div className="form-group">
              <label className="form-label">Deadline (Optional)</label>
              <input
                type="datetime-local"
                className="form-input"
                value={form.deadline}
                onChange={(e) => handleInputChange('deadline', e.target.value)}
                min={new Date().toISOString().slice(0, 16)}
              />
              <div className="form-help">Leave empty for no deadline</div>
            </div>
          </div>

          <div className="form-section">
            <label className="form-label">Required Skills (Optional)</label>
            <div className="skills-input">
              <div className="skill-add">
                <input
                  type="text"
                  className="form-input"
                  value={skillInput}
                  onChange={(e) => setSkillInput(e.target.value)}
                  onKeyPress={handleKeyPress}
                  placeholder="Type a skill and press Enter"
                  maxLength={30}
                />
                <button type="button" className="btn-add-skill" onClick={addSkill}>
                  Add
                </button>
              </div>
              
              {form.required_skills.length > 0 && (
                <div className="skills-list">
                  {form.required_skills.map((skill, index) => (
                    <span key={index} className="skill-tag">
                      {skill}
                      <button
                        type="button"
                        className="skill-remove"
                        onClick={() => removeSkill(skill)}
                      >
                        ×
                      </button>
                    </span>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* Preview */}
          <div className="form-section">
            <h3 className="section-title">Preview</h3>
            <div className={`bounty-preview tier-${tier}`}>
              <div className="preview-header">
                <div className="preview-title-row">
                  <h4 className="preview-title">{form.title || 'Untitled Bounty'}</h4>
                  <div className="preview-icons">
                    <span className="tier-icon">{tierIcons[tier]}</span>
                  </div>
                </div>
                <div className="preview-payout">
                  <span className="payout-amount">🪙 {form.payout_amount}</span>
                  <span className="payout-label">Gold</span>
                </div>
              </div>
              <div className="preview-description">
                {form.description || 'No description provided'}
              </div>
              {form.required_skills.length > 0 && (
                <div className="preview-skills">
                  <span className="skills-label">Required Skills:</span>
                  <div className="skills-list">
                    {form.required_skills.slice(0, 3).map((skill, i) => (
                      <span key={i} className="skill-tag">{skill}</span>
                    ))}
                    {form.required_skills.length > 3 && (
                      <span className="skill-tag more">+{form.required_skills.length - 3} more</span>
                    )}
                  </div>
                </div>
              )}
            </div>
          </div>

          {error && (
            <div className="error-message">
              ❌ {error}
            </div>
          )}

          <div className="form-actions">
            <button
              type="button"
              className="btn-secondary"
              onClick={handleCancel}
              disabled={isSubmitting}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="btn-primary"
              disabled={isSubmitting || !form.title.trim() || !form.description.trim() || form.creator_id === 0}
            >
              {isSubmitting ? '⚡ Creating...' : '💰 Post Bounty'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}