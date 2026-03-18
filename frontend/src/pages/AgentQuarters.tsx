import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useStore, classColor, formatTokens } from '../store';
import type { RPGAgent, RPGQuest, QuestHistoryEntry } from '../store';
import Tooltip from '../components/Tooltip';
import './AgentQuarters.css';

const API_URL = import.meta.env.VITE_API_URL || '';

interface WorkSession {
  id: number;
  name: string;
  status: string;
  tokens_used: number;
  turns: number;
  created_at: string;
  description?: string;
}

interface Conversation {
  id: string;
  participant_ids: string[];
  participants: string[];
  title: string;
  messages: Message[];
  started_at: string;
  last_active: string;
  type: string;
}

interface Message {
  id: string;
  from_agent: string;
  to_agent?: string;
  content: string;
  timestamp: string;
  type: string;
}

export default function AgentQuarters() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const agents = useStore((s) => s.agents);
  const [recentWork, setRecentWork] = useState<WorkSession[]>([]);
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [questHistory, setQuestHistory] = useState<QuestHistoryEntry[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchAll = useStore((s) => s.fetchAll);
  const agent: RPGAgent | undefined = agents.find((a) => a.id === id);

  useEffect(() => {
    if (agents.length === 0) fetchAll();
  }, [agents.length, fetchAll]);

  useEffect(() => {
    if (!id) return;

    setLoading(true);
    
    // Fetch agent's recent work (quests)
    fetch(`${API_URL}/api/inber/quests?agent=${encodeURIComponent(id)}&limit=10`)
      .then((r) => r.ok ? r.json() : [])
      .then((quests: RPGQuest[]) => {
        const workSessions: WorkSession[] = quests.map(q => ({
          id: q.id,
          name: q.name,
          status: q.status,
          tokens_used: q.tokens_used,
          turns: q.turns,
          created_at: q.created_at,
          description: `${q.turns} turns, ${formatTokens(q.tokens_used)} tokens`
        }));
        setRecentWork(workSessions);
      })
      .catch(() => setRecentWork([]));

    // Fetch conversations involving this agent
    fetch(`${API_URL}/api/inber/conversations?limit=50`)
      .then((r) => r.ok ? r.json() : [])
      .then((allConvos: Conversation[]) => {
        // Filter conversations that involve this agent
        const agentConvos = allConvos.filter(c => 
          c.participant_ids.includes(id) || c.messages.some((m: Message) => m.from_agent === id)
        );
        setConversations(agentConvos.slice(0, 8)); // Latest 8 conversations
      })
      .catch(() => setConversations([]));

    // Fetch quest history for activity timeline
    fetch(`${API_URL}/api/inber/quest-history?agent=${encodeURIComponent(id)}&limit=15`)
      .then((r) => r.ok ? r.json() : [])
      .then(setQuestHistory)
      .catch(() => setQuestHistory([]));

    setLoading(false);
  }, [id]);

  if (!agent) {
    return (
      <div className="quarters-loading">
        {agents.length === 0 ? (
          <p>🏠 Preparing quarters...</p>
        ) : (
          <>
            <p>🚪 Quarters for "{id}" not found.</p>
            <button onClick={() => navigate('/')} className="quarters-back-btn">← Back to Tavern</button>
          </>
        )}
      </div>
    );
  }

  const cc = classColor(agent.class);

  return (
    <div className="agent-quarters">
      <button className="quarters-back-btn" onClick={() => navigate('/')}>
        ← Back to Tavern
      </button>

      {/* Room Header */}
      <div className="quarters-header">
        <div className="room-sign" style={{ borderColor: cc }}>
          <h1 className="room-title" style={{ color: cc }}>{agent.name}'s Quarters</h1>
          <p className="room-subtitle">Level {agent.level} {agent.class} • {agent.energy}% Energy</p>
        </div>
        <div className="quarters-avatar">
          <img
            src={`/avatars/${agent.name.toLowerCase().replace(/\s+/g, '-')}.png`}
            alt={agent.name}
            onError={(e) => {
              e.currentTarget.style.display = 'none';
              const sibling = e.currentTarget.nextElementSibling as HTMLElement;
              if (sibling) sibling.style.display = 'block';
            }}
          />
          <span style={{ display: 'none', fontSize: '3rem' }}>{agent.avatar_emoji}</span>
        </div>
      </div>

      {loading ? (
        <div className="quarters-loading">
          <p>🕯️ Organizing the room...</p>
        </div>
      ) : (
        <div className="quarters-content">
          {/* Recent Work Area */}
          <div className="quarters-section work-area">
            <h2>📋 Work Desk</h2>
            <p className="section-subtitle">Recent projects and tasks</p>
            {recentWork.length > 0 ? (
              <div className="work-list">
                {recentWork.map((work) => (
                  <div key={work.id} className={`work-item status-${work.status}`}>
                    <div className="work-header">
                      <span className="work-name">{work.name}</span>
                      <span className={`work-status status-${work.status}`}>
                        {work.status === 'completed' ? '✅' : work.status === 'active' ? '⚡' : '⏸️'}
                      </span>
                    </div>
                    <div className="work-meta">
                      <span>🔮 {formatTokens(work.tokens_used)}</span>
                      <span>↻ {work.turns} turns</span>
                      <Tooltip content={new Date(work.created_at).toLocaleString()}>
                        <span className="work-date">
                          {new Date(work.created_at).toLocaleDateString()}
                        </span>
                      </Tooltip>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="empty-area">
                <p>📄 The work desk is clean - no recent projects</p>
              </div>
            )}
          </div>

          {/* Memory Wall - Recent Conversations */}
          <div className="quarters-section memory-wall">
            <h2>💭 Memory Wall</h2>
            <p className="section-subtitle">Recent thoughts and conversations</p>
            {conversations.length > 0 ? (
              <div className="memory-list">
                {conversations.map((convo) => {
                  const agentMessage = convo.messages.find((m: Message) => m.from_agent === id);
                  return (
                    <div key={convo.id} className="memory-item">
                      <div className="memory-header">
                        <span className="memory-topic">{convo.title || 'Untitled Conversation'}</span>
                        <span className="memory-participants">
                          {convo.participants.length} participants
                        </span>
                      </div>
                      {agentMessage && (
                        <div className="memory-snippet">
                          "{agentMessage.content.slice(0, 120)}..."
                        </div>
                      )}
                      <div className="memory-meta">
                        <span>{convo.messages.length} messages</span>
                        <Tooltip content={new Date(convo.last_active || '').toLocaleString()}>
                          <span className="memory-date">
                            {new Date(convo.last_active || '').toLocaleDateString()}
                          </span>
                        </Tooltip>
                      </div>
                    </div>
                  );
                })}
              </div>
            ) : (
              <div className="empty-area">
                <p>🧠 The memory wall is blank - no recent conversations</p>
              </div>
            )}
          </div>

          {/* Activity Timeline */}
          <div className="quarters-section activity-timeline">
            <h2>📈 Activity Timeline</h2>
            <p className="section-subtitle">Work patterns over time</p>
            {questHistory.length > 0 ? (
              <div className="timeline-content">
                <div className="timeline-chart">
                  <svg viewBox="0 0 300 100" className="activity-chart">
                    {(() => {
                      const data = questHistory.slice().reverse(); // Chronological order
                      const maxTokens = Math.max(...data.map(d => d.tokens), 1);
                      return data.map((entry, i) => {
                        const x = (i / (data.length - 1)) * 280 + 10;
                        const height = (entry.tokens / maxTokens) * 60;
                        const y = 90 - height;
                        return (
                          <g key={entry.id}>
                            <rect
                              x={x - 3}
                              y={y}
                              width="6"
                              height={height}
                              fill={cc}
                              opacity="0.7"
                            />
                            <title>{`${entry.tokens} tokens • ${new Date(entry.started_at).toLocaleDateString()}`}</title>
                          </g>
                        );
                      });
                    })()}
                  </svg>
                </div>
                <div className="timeline-summary">
                  <div className="timeline-stat">
                    <strong>{questHistory.reduce((sum, h) => sum + h.tokens, 0).toLocaleString()}</strong>
                    <span>Total Tokens</span>
                  </div>
                  <div className="timeline-stat">
                    <strong>{questHistory.length}</strong>
                    <span>Recent Tasks</span>
                  </div>
                  <div className="timeline-stat">
                    <strong>{Math.round(questHistory.reduce((sum, h) => sum + h.tokens, 0) / questHistory.length).toLocaleString()}</strong>
                    <span>Avg Per Task</span>
                  </div>
                </div>
              </div>
            ) : (
              <div className="empty-area">
                <p>📊 No activity data available</p>
              </div>
            )}
          </div>

          {/* Personal Space */}
          <div className="quarters-section personal-space">
            <h2>🛏️ Personal Space</h2>
            <p className="section-subtitle">Agent's current state and environment</p>
            <div className="personal-grid">
              <div className="personal-item">
                <span className="personal-label">Current Mood</span>
                <span className="personal-value" style={{ color: cc }}>
                  {agent.energy > 80 ? '😊 Energetic' : 
                   agent.energy > 50 ? '😐 Content' : 
                   agent.energy > 20 ? '😴 Tired' : '😵 Exhausted'}
                </span>
              </div>
              <div className="personal-item">
                <span className="personal-label">Specialization</span>
                <span className="personal-value">{agent.class}</span>
              </div>
              <div className="personal-item">
                <span className="personal-label">Experience</span>
                <span className="personal-value">{agent.xp.toLocaleString()} XP</span>
              </div>
              <div className="personal-item">
                <span className="personal-label">Total Sessions</span>
                <span className="personal-value">{agent.session_count}</span>
              </div>
              <div className="personal-item">
                <span className="personal-label">Skills Mastered</span>
                <span className="personal-value">{agent.skills.length} skills</span>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Navigation to Character Sheet */}
      <div className="quarters-footer">
        <button
          onClick={() => navigate(`/agent/${id}`)}
          className="view-character-btn"
          style={{ borderColor: cc, color: cc }}
        >
          📋 View Character Sheet
        </button>
      </div>
    </div>
  );
}