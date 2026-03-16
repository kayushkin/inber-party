import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useStore } from '../store';
import './CampView.css';

export default function CampView() {
  const navigate = useNavigate();
  const agents = useStore((state) => state.agents);

  const startPolling = useStore((state) => state.startPolling);
  const stopPolling = useStore((state) => state.stopPolling);

  useEffect(() => {
    startPolling(10000);
    return () => stopPolling();
  }, [startPolling, stopPolling]);

  const getStatusClass = (status: string) => {
    switch (status) {
      case 'working':
        return 'status-working';
      case 'on_quest':
        return 'status-quest';
      case 'stuck':
        return 'status-stuck';
      case 'resting':
        return 'status-resting';
      default:
        return 'status-idle';
    }
  };

  const getStatusLabel = (status: string) => {
    return status.replace('_', ' ').toUpperCase();
  };

  return (
    <div className="camp-view">
      <div className="camp-scene">
        <div className="camp-fire">
          <div className="fire-emoji">🔥</div>
          <div className="fire-glow" />
        </div>
        
        <div className="agents-circle">
          {agents.map((agent, index) => (
            <div
              key={agent.id}
              className={`agent-card ${getStatusClass(agent.status)}`}
              style={{
                transform: `rotate(${index * (360 / agents.length)}deg) translateY(-180px) rotate(-${index * (360 / agents.length)}deg)`,
              }}
              onClick={() => navigate(`/agent/${agent.id}`)}
            >
              <div className="agent-avatar">{agent.avatar_emoji}</div>
              <div className="agent-info">
                <div className="agent-name">{agent.name}</div>
                <div className="agent-title">{agent.title}</div>
                <div className="agent-class">
                  <span className="class-icon">
                    {agent.class === 'Wizard' && '🧙'}
                    {agent.class === 'Ranger' && '🏹'}
                    {agent.class === 'Warrior' && '⚔️'}
                  </span>
                  Level {agent.level}
                </div>
                <div className={`agent-status ${getStatusClass(agent.status)}`}>
                  {getStatusLabel(agent.status)}
                </div>
                <div className="energy-bar">
                  <div className="energy-fill" style={{ width: `${agent.energy}%` }} />
                  <div className="energy-text">{agent.energy}%</div>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
      
      <div className="camp-info">
        <h2>The Camp</h2>
        <p>Your party of {agents.length} adventurers rests by the fire, ready for their next quest.</p>
        <p className="hint">Click on an agent to view their character sheet.</p>
      </div>
    </div>
  );
}
