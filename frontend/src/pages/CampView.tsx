import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useStore, api } from '../store';
import type { InberAgent } from '../store';
import './CampView.css';

// RPG-flavored idle messages per class
const IDLE_MESSAGES: Record<string, string[]> = {
  Wizard: [
    'is studying ancient scrolls...',
    'is channeling arcane energies...',
    'is deciphering a cryptic tome...',
    'meditates on the nature of code...',
  ],
  Healer: [
    'is tending to the party\'s wounds...',
    'is brewing a restorative potion...',
    'is communing with the light...',
    'hums a soothing melody...',
  ],
  Ranger: [
    'scouts the perimeter...',
    'is sharpening their arrows...',
    'tracks movement in the shadows...',
    'studies the terrain ahead...',
  ],
  Warrior: [
    'is polishing their blade...',
    'practices combat forms...',
    'stands guard by the fire...',
    'is forging new armor...',
  ],
};

const CLASS_COLORS: Record<string, string> = {
  Wizard: '#a78bfa',
  Healer: '#4ade80',
  Ranger: '#60a5fa',
  Warrior: '#f87171',
};

function getIdleMessage(name: string, agentClass: string) {
  const msgs = IDLE_MESSAGES[agentClass] || IDLE_MESSAGES.Warrior;
  // Deterministic but varied by name
  const idx = name.charCodeAt(0) % msgs.length;
  return `${name} ${msgs[idx]}`;
}

function formatTokens(n: number) {
  if (n >= 1000000) return `${(n / 1000000).toFixed(1)}M`;
  if (n >= 1000) return `${(n / 1000).toFixed(1)}k`;
  return String(n);
}

function formatCost(c: number) {
  if (c <= 0) return '$0';
  if (c < 0.01) return `$${c.toFixed(4)}`;
  return `$${c.toFixed(2)}`;
}

export default function CampView() {
  const navigate = useNavigate();
  const agents = useStore((state) => state.agents);
  const inberAgents = useStore((state) => state.inberAgents);
  const inberStats = useStore((state) => state.inberStats);
  const startPolling = useStore((state) => state.startPolling);
  const stopPolling = useStore((state) => state.stopPolling);
  const [, setTick] = useState(0);

  useEffect(() => {
    startPolling(10000);
    return () => stopPolling();
  }, [startPolling, stopPolling]);

  // Rotate idle messages every 8s
  useEffect(() => {
    const t = setInterval(() => setTick(v => v + 1), 8000);
    return () => clearInterval(t);
  }, []);

  const getStatusClass = (status: string) => {
    switch (status) {
      case 'working': return 'status-working';
      case 'on_quest': return 'status-quest';
      case 'stuck': return 'status-stuck';
      case 'resting': return 'status-resting';
      default: return 'status-idle';
    }
  };

  const getStatusLabel = (status: string, agent: typeof agents[0]) => {
    if (status === 'idle') {
      const ia = inberAgents.find(a => a.name === agent.name);
      if (ia) return getIdleMessage(agent.name, ia.class);
      return getIdleMessage(agent.name, agent.class);
    }
    return status.replace('_', ' ').toUpperCase();
  };

  const getClassColor = (agentClass: string) => CLASS_COLORS[agentClass] || '#d4af37';

  // Find matching inber agent
  const getInber = (name: string): InberAgent | undefined =>
    inberAgents.find(a => a.name === name);

  const isInber = api.isInberMode();

  return (
    <div className="camp-view">
      {/* Guild Hall Stats */}
      {isInber && inberStats && (
        <div className="guild-hall">
          <h2 className="guild-title">⚔️ Guild Hall</h2>
          <div className="guild-stats">
            <div className="guild-stat">
              <div className="guild-stat-value">{inberStats.total_agents}</div>
              <div className="guild-stat-label">Adventurers</div>
            </div>
            <div className="guild-stat">
              <div className="guild-stat-value">{formatTokens(inberStats.total_tokens)}</div>
              <div className="guild-stat-label">Tokens Used</div>
            </div>
            <div className="guild-stat">
              <div className="guild-stat-value">{formatCost(inberStats.total_cost)}</div>
              <div className="guild-stat-label">Gold Spent</div>
            </div>
            <div className="guild-stat">
              <div className="guild-stat-value">{inberStats.completed_quests + inberStats.failed_quests + inberStats.active_quests}</div>
              <div className="guild-stat-label">Total Quests</div>
            </div>
            <div className="guild-stat">
              <div className="guild-stat-value">{inberStats.total_sessions}</div>
              <div className="guild-stat-label">Sessions</div>
            </div>
            <div className="guild-stat">
              <div className="guild-stat-value">{inberStats.uptime || '—'}</div>
              <div className="guild-stat-label">Uptime</div>
            </div>
          </div>
        </div>
      )}

      <div className="camp-scene">
        <div className="camp-fire">
          <div className="fire-emoji">🔥</div>
          <div className="fire-glow" />
        </div>

        <div className="agents-circle">
          {agents.map((agent, index) => {
            const ia = getInber(agent.name);
            const classColor = getClassColor(ia?.class || agent.class);
            return (
              <div
                key={agent.id}
                className={`agent-card ${getStatusClass(agent.status)}`}
                style={{
                  '--class-color': classColor,
                  transform: `rotate(${index * (360 / agents.length)}deg) translateY(-180px) rotate(-${index * (360 / agents.length)}deg)`,
                } as React.CSSProperties}
                onClick={() => navigate(`/agent/${agent.id}`)}
              >
                <div className="agent-avatar">{agent.avatar_emoji}</div>
                <div className="agent-info">
                  <div className="agent-name" style={{ color: classColor }}>{agent.name}</div>
                  <div className="agent-title">{agent.title}</div>
                  <div className="agent-class" style={{ color: classColor }}>
                    <span className="class-icon">{agent.avatar_emoji}</span>
                    {ia?.class || agent.class} · Lv {agent.level}
                  </div>
                  <div className={`agent-status ${getStatusClass(agent.status)} ${agent.status === 'idle' ? 'idle-flavor' : ''}`}>
                    {getStatusLabel(agent.status, agent)}
                  </div>
                  <div className="energy-bar">
                    <div className="energy-fill" style={{ width: `${agent.energy}%` }} />
                    <div className="energy-text">{agent.energy}%</div>
                  </div>
                </div>
              </div>
            );
          })}
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
