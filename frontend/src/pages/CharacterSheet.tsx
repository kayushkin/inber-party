import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { api } from '../store';
import type { AgentDetail } from '../store';
import './CharacterSheet.css';

export default function CharacterSheet() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [agent, setAgent] = useState<AgentDetail | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (id) {
      setLoading(true);
      api.getAgent(parseInt(id))
        .then(setAgent)
        .catch((err) => {
          console.error('Failed to load agent:', err);
          navigate('/');
        })
        .finally(() => setLoading(false));
    }
  }, [id, navigate]);

  if (loading) {
    return <div className="loading">Loading character sheet...</div>;
  }

  if (!agent) {
    return <div className="error">Agent not found</div>;
  }

  const xpToNextLevel = (agent.level + 1) * 100;
  const xpProgress = (agent.xp / xpToNextLevel) * 100;

  return (
    <div className="character-sheet">
      <button className="back-button" onClick={() => navigate('/')}>
        ← Back to Camp
      </button>

      <div className="character-header">
        <div className="character-avatar">{agent.avatar_emoji}</div>
        <div className="character-title-section">
          <h1 className="character-name">{agent.name}</h1>
          <h2 className="character-title">{agent.title}</h2>
          <div className="character-class">{agent.class}</div>
        </div>
      </div>

      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-label">Level</div>
          <div className="stat-value">{agent.level}</div>
        </div>
        <div className="stat-card">
          <div className="stat-label">XP</div>
          <div className="stat-value">{agent.xp}</div>
          <div className="stat-progress">
            <div className="progress-bar">
              <div className="progress-fill" style={{ width: `${xpProgress}%` }} />
            </div>
            <div className="progress-text">{agent.xp} / {xpToNextLevel}</div>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Energy</div>
          <div className="stat-value">{agent.energy}%</div>
          <div className="stat-progress">
            <div className="progress-bar energy">
              <div className="progress-fill" style={{ width: `${agent.energy}%` }} />
            </div>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Status</div>
          <div className={`stat-value status-${agent.status}`}>
            {agent.status.replace('_', ' ').toUpperCase()}
          </div>
        </div>
      </div>

      <div className="content-grid">
        <div className="section skills-section">
          <h3>Skills</h3>
          {agent.skills && agent.skills.length > 0 ? (
            <div className="skills-list">
              {agent.skills.map((skill) => (
                <div key={skill.id} className="skill-item">
                  <div className="skill-name">{skill.skill_name}</div>
                  <div className="skill-meta">
                    <span className="skill-level">Level {skill.level}</span>
                    <span className="skill-count">{skill.task_count} tasks</span>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p className="empty-message">No skills yet</p>
          )}
        </div>

        <div className="section achievements-section">
          <h3>Achievements</h3>
          {agent.achievements && agent.achievements.length > 0 ? (
            <div className="achievements-list">
              {agent.achievements.map((achievement) => (
                <div key={achievement.id} className="achievement-item">
                  <div className="achievement-icon">🏆</div>
                  <div className="achievement-info">
                    <div className="achievement-name">{achievement.achievement_name}</div>
                    <div className="achievement-date">
                      {new Date(achievement.unlocked_at).toLocaleDateString()}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p className="empty-message">No achievements yet</p>
          )}
        </div>

        <div className="section quest-log-section">
          <h3>Quest Log</h3>
          {agent.tasks && agent.tasks.length > 0 ? (
            <div className="quest-list">
              {agent.tasks.map((task) => (
                <div key={task.id} className={`quest-item status-${task.status}`}>
                  <div className="quest-header">
                    <div className="quest-name">{task.name}</div>
                    <div className="quest-status">{task.status.replace('_', ' ')}</div>
                  </div>
                  <div className="quest-description">{task.description}</div>
                  <div className="quest-meta">
                    <span className="quest-difficulty">Difficulty: {task.difficulty}</span>
                    <span className="quest-reward">+{task.xp_reward} XP</span>
                  </div>
                  {task.status === 'in_progress' && (
                    <div className="quest-progress">
                      <div className="progress-bar">
                        <div className="progress-fill" style={{ width: `${task.progress}%` }} />
                      </div>
                      <div className="progress-text">{task.progress}%</div>
                    </div>
                  )}
                </div>
              ))}
            </div>
          ) : (
            <p className="empty-message">No active quests</p>
          )}
        </div>
      </div>
    </div>
  );
}
