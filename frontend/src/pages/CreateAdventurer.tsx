import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { CLASS_COLORS, classColor } from '../store';
import './CreateAdventurer.css';

interface CreateAdventurerForm {
  name: string;
  title: string;
  class: string;
  domain: string;
  avatar_emoji: string;
}

export default function CreateAdventurer() {
  const navigate = useNavigate();
  const [form, setForm] = useState<CreateAdventurerForm>({
    name: '',
    title: '',
    class: 'Adventurer',
    domain: '',
    avatar_emoji: '🧙‍♂️'
  });
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState('');

  const availableClasses = Object.keys(CLASS_COLORS);
  const avatarOptions = ['🧙‍♂️', '⚔️', '🏹', '🛡️', '🔮', '⚒️', '📜', '🎭', '🌟', '🔥', '⚡', '❄️', '🌱', '💀', '👑'];

  const handleInputChange = (field: keyof CreateAdventurerForm, value: string) => {
    setForm(prev => ({ ...prev, [field]: value }));
    if (error) setError(''); // Clear error when user types
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    // Basic validation
    if (!form.name.trim()) {
      setError('Name is required');
      return;
    }
    
    if (!form.title.trim()) {
      setError('Title is required');
      return;
    }

    setIsSubmitting(true);
    setError('');

    try {
      const response = await fetch('/api/agents', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name: form.name.trim(),
          title: form.title.trim(),
          class: form.class,
          level: 1,
          xp: 0,
          avatar_emoji: form.avatar_emoji,
        }),
      });

      if (!response.ok) {
        throw new Error(`Failed to create adventurer: ${response.statusText}`);
      }

      const newAgent = await response.json();
      
      // Navigate to the new agent's character sheet
      navigate(`/agent/${newAgent.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create adventurer');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCancel = () => {
    navigate('/');
  };

  return (
    <div className="create-adventurer">
      <div className="create-adventurer-header">
        <button className="back-btn" onClick={handleCancel} title="Back to Tavern">
          ← Back to Tavern
        </button>
        <h1 className="page-title">🌟 Create New Adventurer</h1>
        <p className="page-subtitle">Forge a new hero to join your guild</p>
      </div>

      <div className="create-adventurer-content">
        <form onSubmit={handleSubmit} className="adventurer-form">
          <div className="form-section">
            <h2 className="section-title">Basic Information</h2>
            
            <div className="form-group">
              <label className="form-label">Name *</label>
              <input
                type="text"
                className="form-input"
                value={form.name}
                onChange={(e) => handleInputChange('name', e.target.value)}
                placeholder="Enter the adventurer's name"
                maxLength={50}
              />
            </div>

            <div className="form-group">
              <label className="form-label">Title *</label>
              <input
                type="text"
                className="form-input"
                value={form.title}
                onChange={(e) => handleInputChange('title', e.target.value)}
                placeholder="e.g., The Code Whisperer, Keeper of Secrets"
                maxLength={100}
              />
            </div>

            <div className="form-group">
              <label className="form-label">Domain/Repository</label>
              <input
                type="text"
                className="form-input"
                value={form.domain}
                onChange={(e) => handleInputChange('domain', e.target.value)}
                placeholder="e.g., frontend, backend, docs (optional)"
                maxLength={50}
              />
              <div className="form-help">
                Specify the domain or repository this adventurer will specialize in
              </div>
            </div>
          </div>

          <div className="form-section">
            <h2 className="section-title">Class & Appearance</h2>
            
            <div className="form-group">
              <label className="form-label">Class</label>
              <div className="class-selector">
                {availableClasses.map((cls) => (
                  <div
                    key={cls}
                    className={`class-option ${form.class === cls ? 'selected' : ''}`}
                    style={{ '--class-color': classColor(cls) } as React.CSSProperties}
                    onClick={() => handleInputChange('class', cls)}
                  >
                    <div className="class-name" style={{ color: classColor(cls) }}>
                      {cls}
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <div className="form-group">
              <label className="form-label">Avatar</label>
              <div className="avatar-selector">
                {avatarOptions.map((emoji) => (
                  <div
                    key={emoji}
                    className={`avatar-option ${form.avatar_emoji === emoji ? 'selected' : ''}`}
                    onClick={() => handleInputChange('avatar_emoji', emoji)}
                  >
                    {emoji}
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Preview */}
          <div className="form-section">
            <h2 className="section-title">Preview</h2>
            <div className="adventurer-preview">
              <div 
                className="preview-card"
                style={{ '--class-color': classColor(form.class) } as React.CSSProperties}
              >
                <div className="preview-avatar">{form.avatar_emoji}</div>
                <div className="preview-info">
                  <div className="preview-name" style={{ color: classColor(form.class) }}>
                    {form.name || 'Unnamed Adventurer'}
                  </div>
                  <div className="preview-title">
                    {form.title || 'No Title'}
                  </div>
                  <div className="preview-class">
                    {form.class} · Lv 1
                  </div>
                  {form.domain && (
                    <div className="preview-domain">
                      Domain: {form.domain}
                    </div>
                  )}
                </div>
              </div>
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
              disabled={isSubmitting || !form.name.trim() || !form.title.trim()}
            >
              {isSubmitting ? '⚡ Creating...' : '🌟 Create Adventurer'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}