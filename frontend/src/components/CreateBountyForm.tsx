import React, { useState, useEffect, useCallback } from 'react';
import './CreateBountyForm.css';

interface Verifier {
  verifier_type: string;
  config: Record<string, unknown>;
  required: boolean;
  weight: number;
}

interface CreateBountyFormData {
  title: string;
  description: string;
  requirements: string;
  payout_amount: number;
  creator_id: number;
  deadline: string;
  required_skills: string[];
  verifiers: Verifier[];
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
    required_skills: [],
    verifiers: []
  });
  const [agents, setAgents] = useState<Agent[]>([]);
  const [skillInput, setSkillInput] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [verifierTypes, setVerifierTypes] = useState<Record<string, unknown>>({});
  const [showVerifierConfig, setShowVerifierConfig] = useState(false);

  const fetchAgents = useCallback(async () => {
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
  }, [form.creator_id]);

  const fetchVerifierTypes = useCallback(async () => {
    try {
      const response = await fetch('/api/verifiers/types');
      if (response.ok) {
        const data = await response.json();
        setVerifierTypes(data.schemas || {});
      }
    } catch (error) {
      console.error('Error fetching verifier types:', error);
    }
  }, []);

  useEffect(() => {
    if (isOpen) {
      fetchAgents();
      fetchVerifierTypes();
    }
  }, [isOpen, fetchAgents, fetchVerifierTypes]);

  const handleInputChange = (field: keyof CreateBountyFormData, value: string | number | string[] | Verifier[]) => {
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

  const addVerifier = (verifierType: string) => {
    const newVerifier: Verifier = {
      verifier_type: verifierType,
      config: {},
      required: true,
      weight: 1.0
    };
    setForm(prev => ({
      ...prev,
      verifiers: [...prev.verifiers, newVerifier]
    }));
  };

  const removeVerifier = (index: number) => {
    setForm(prev => ({
      ...prev,
      verifiers: prev.verifiers.filter((_, i) => i !== index)
    }));
  };

  const updateVerifier = (index: number, updates: Partial<Verifier>) => {
    setForm(prev => ({
      ...prev,
      verifiers: prev.verifiers.map((v, i) => i === index ? { ...v, ...updates } : v)
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
        required_skills: [],
        verifiers: []
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
                  <span className={`tier-indicator tier-${tier}`}>
                    {tierIcons[tier]} {tier.charAt(0).toUpperCase() + tier.slice(1)} Tier
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
                <div className="tier-info">
                  <small className="tier-thresholds">
                    Tiers: Bronze (&lt; 50) • Silver (50-199) • Gold (200-499) • Legendary (500+)
                  </small>
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
            <div className="section-header">
              <h3 className="section-title">🔍 Verification Requirements</h3>
              <button
                type="button"
                className="btn-toggle"
                onClick={() => setShowVerifierConfig(!showVerifierConfig)}
              >
                {showVerifierConfig ? 'Hide' : 'Configure'} Verifiers
              </button>
            </div>
            
            {showVerifierConfig && (
              <div className="verifiers-config">
                <p className="form-help">
                  Add verification requirements that must be met before the bounty is approved.
                  Automated verifiers can check files, run tests, or validate other criteria.
                </p>

                {form.verifiers.map((verifier, index) => (
                  <div key={index} className="verifier-config">
                    <div className="verifier-header">
                      <span className="verifier-type">{verifier.verifier_type}</span>
                      <div className="verifier-controls">
                        <label className="checkbox-label">
                          <input
                            type="checkbox"
                            checked={verifier.required}
                            onChange={(e) => updateVerifier(index, { required: e.target.checked })}
                          />
                          Required
                        </label>
                        <button
                          type="button"
                          className="btn-remove"
                          onClick={() => removeVerifier(index)}
                        >
                          Remove
                        </button>
                      </div>
                    </div>

                    <div className="verifier-config-fields">
                      {verifier.verifier_type === 'file_exists' && (
                        <>
                          <div className="config-field">
                            <label>Files to check (one per line):</label>
                            <textarea
                              placeholder="README.md&#10;tests/test_*.py&#10;package.json"
                              onChange={(e) => {
                                const files = e.target.value.split('\n').filter(f => f.trim());
                                updateVerifier(index, { config: { ...verifier.config, files } });
                              }}
                              rows={3}
                            />
                          </div>
                          <div className="config-field">
                            <label>Base path (optional):</label>
                            <input
                              type="text"
                              placeholder="/path/to/project"
                              onChange={(e) => updateVerifier(index, { 
                                config: { ...verifier.config, base_path: e.target.value } 
                              })}
                            />
                          </div>
                        </>
                      )}

                      {verifier.verifier_type === 'test_suite' && (
                        <>
                          <div className="config-field">
                            <label>Command to run:</label>
                            <input
                              type="text"
                              placeholder="npm test"
                              onChange={(e) => updateVerifier(index, { 
                                config: { ...verifier.config, command: e.target.value } 
                              })}
                            />
                          </div>
                          <div className="config-field">
                            <label>Working directory (optional):</label>
                            <input
                              type="text"
                              placeholder="/path/to/project"
                              onChange={(e) => updateVerifier(index, { 
                                config: { ...verifier.config, working_dir: e.target.value } 
                              })}
                            />
                          </div>
                          <div className="config-field">
                            <label>Timeout (seconds):</label>
                            <input
                              type="number"
                              placeholder="300"
                              onChange={(e) => updateVerifier(index, { 
                                config: { ...verifier.config, timeout_sec: parseInt(e.target.value) || 300 } 
                              })}
                            />
                          </div>
                        </>
                      )}

                      {verifier.verifier_type === 'manual' && (
                        <>
                          <div className="config-field">
                            <label>Review instructions:</label>
                            <textarea
                              placeholder="What should the reviewer check?"
                              onChange={(e) => updateVerifier(index, { 
                                config: { ...verifier.config, instructions: e.target.value } 
                              })}
                              rows={3}
                            />
                          </div>
                          <div className="config-field">
                            <label>Assign to (optional):</label>
                            <input
                              type="text"
                              placeholder="username or role"
                              onChange={(e) => updateVerifier(index, { 
                                config: { ...verifier.config, assignee: e.target.value } 
                              })}
                            />
                          </div>
                        </>
                      )}
                    </div>
                  </div>
                ))}

                <div className="add-verifier">
                  <label>Add verifier:</label>
                  <div className="verifier-buttons">
                    {Object.keys(verifierTypes).map(type => (
                      <button
                        key={type}
                        type="button"
                        className="btn-add-verifier"
                        onClick={() => addVerifier(type)}
                      >
                        {type.replace('_', ' ')}
                      </button>
                    ))}
                  </div>
                </div>
              </div>
            )}
          </div>

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