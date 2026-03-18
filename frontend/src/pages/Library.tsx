import { useState, useEffect } from 'react';
import { timeAgo, formatTokens, formatCost } from '../store';
import SessionReplay from '../components/SessionReplay';
import './Library.css';

interface SessionLog {
  sessionKey: string;
  agentId: string;
  agentName?: string;
  status: string;
  lastActive: string;
  messageCount: number;
  totalTokens: number;
  totalCost: number;
  created: string;
  summary?: string;
}

interface SessionMessage {
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: string;
  tokens?: number;
}

interface SessionDetail {
  sessionKey: string;
  messages: SessionMessage[];
  agentId: string;
  agentName: string;
  status: string;
  totalTokens: number;
  totalCost: number;
  created: string;
  lastActive: string;
}

export default function Library() {
  const [sessions, setSessions] = useState<SessionLog[]>([]);
  const [filteredSessions, setFilteredSessions] = useState<SessionLog[]>([]);
  const [selectedSession, setSelectedSession] = useState<SessionDetail | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [loading, setLoading] = useState(true);
  const [loadingDetail, setLoadingDetail] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [replaySessionId, setReplaySessionId] = useState<string | null>(null);

  // Fetch sessions list
  useEffect(() => {
    fetchSessions();
  }, []);

  // Filter sessions based on search and status
  useEffect(() => {
    let filtered = sessions;

    if (searchTerm) {
      const term = searchTerm.toLowerCase();
      filtered = sessions.filter(session => 
        session.agentName?.toLowerCase().includes(term) ||
        session.sessionKey.toLowerCase().includes(term) ||
        session.summary?.toLowerCase().includes(term)
      );
    }

    if (statusFilter !== 'all') {
      filtered = filtered.filter(session => session.status === statusFilter);
    }

    setFilteredSessions(filtered);
  }, [sessions, searchTerm, statusFilter]);

  const fetchSessions = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/inber-proxy/sessions');
      if (!response.ok) {
        throw new Error('Failed to fetch sessions');
      }
      const data = await response.json();
      
      // Transform the data to match our interface
      const transformedSessions: SessionLog[] = data.map((session: any) => ({
        sessionKey: session.sessionKey || session.key || session.id,
        agentId: session.agentId || session.agent_id || 'unknown',
        agentName: session.agentName || session.agent_name || 'Unknown Agent',
        status: session.status || 'unknown',
        lastActive: session.lastActive || session.last_active || session.updated_at || new Date().toISOString(),
        messageCount: session.messageCount || session.message_count || 0,
        totalTokens: session.totalTokens || session.total_tokens || 0,
        totalCost: session.totalCost || session.total_cost || 0,
        created: session.created || session.created_at || new Date().toISOString(),
        summary: session.summary || session.title || ''
      }));

      setSessions(transformedSessions);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
      console.error('Failed to fetch sessions:', err);
    } finally {
      setLoading(false);
    }
  };

  const fetchSessionDetail = async (sessionKey: string) => {
    try {
      setLoadingDetail(true);
      const response = await fetch(`/api/inber-proxy/sessions/${sessionKey}/history`);
      if (!response.ok) {
        throw new Error('Failed to fetch session details');
      }
      const data = await response.json();
      
      const session = sessions.find(s => s.sessionKey === sessionKey);
      if (!session) return;

      const detail: SessionDetail = {
        sessionKey,
        messages: data.messages || [],
        agentId: session.agentId,
        agentName: session.agentName || 'Unknown Agent',
        status: session.status,
        totalTokens: session.totalTokens,
        totalCost: session.totalCost,
        created: session.created,
        lastActive: session.lastActive
      };

      setSelectedSession(detail);
    } catch (err) {
      console.error('Failed to fetch session detail:', err);
    } finally {
      setLoadingDetail(false);
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status.toLowerCase()) {
      case 'active': return '🔥';
      case 'completed': return '✅';
      case 'failed': return '❌';
      case 'idle': return '😴';
      case 'paused': return '⏸️';
      default: return '❓';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'active': return '#ff6b6b';
      case 'completed': return '#4ade80';
      case 'failed': return '#f87171';
      case 'idle': return '#94a3b8';
      case 'paused': return '#fbbf24';
      default: return '#d4af37';
    }
  };

  if (loading) {
    return (
      <div className="library">
        <div className="library-header">
          <h1>📚 Library</h1>
          <p>Session logs and history archive</p>
        </div>
        <div className="loading-state">
          <div className="loading-spinner"></div>
          <p>Searching the ancient scrolls...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="library">
        <div className="library-header">
          <h1>📚 Library</h1>
          <p>Session logs and history archive</p>
        </div>
        <div className="error-state">
          <h3>❌ Error accessing the archives</h3>
          <p>{error}</p>
          <button onClick={fetchSessions} className="retry-button">
            🔄 Try Again
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="library">
      <div className="library-header">
        <h1>📚 Library</h1>
        <p>Session logs and history archive</p>
      </div>

      <div className="library-controls">
        <div className="search-box">
          <input
            type="text"
            placeholder="🔍 Search sessions by agent, key, or content..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="search-input"
          />
        </div>
        
        <div className="filter-box">
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="status-filter"
          >
            <option value="all">All Status</option>
            <option value="active">Active</option>
            <option value="completed">Completed</option>
            <option value="failed">Failed</option>
            <option value="idle">Idle</option>
            <option value="paused">Paused</option>
          </select>
        </div>
      </div>

      <div className="library-content">
        <div className="sessions-list">
          <div className="sessions-header">
            <h3>Session Archives ({filteredSessions.length})</h3>
            <button onClick={fetchSessions} className="refresh-button" disabled={loading}>
              🔄 Refresh
            </button>
          </div>

          {filteredSessions.length === 0 ? (
            <div className="empty-state">
              <p>📜 No sessions found matching your search.</p>
            </div>
          ) : (
            <div className="sessions-grid">
              {filteredSessions.map((session) => (
                <div
                  key={session.sessionKey}
                  className={`session-card ${selectedSession?.sessionKey === session.sessionKey ? 'selected' : ''}`}
                  onClick={() => fetchSessionDetail(session.sessionKey)}
                >
                  <div className="session-card-header">
                    <span 
                      className="session-status"
                      style={{ color: getStatusColor(session.status) }}
                    >
                      {getStatusIcon(session.status)} {session.status}
                    </span>
                    <span className="session-time">{timeAgo(session.lastActive)}</span>
                  </div>
                  
                  <div className="session-card-body">
                    <h4 className="agent-name">{session.agentName}</h4>
                    <p className="session-key">{session.sessionKey}</p>
                    {session.summary && (
                      <p className="session-summary">{session.summary}</p>
                    )}
                  </div>
                  
                  <div className="session-card-footer">
                    <div className="session-stats">
                      <span className="session-messages">💬 {session.messageCount}</span>
                      <span className="session-tokens">🔮 {formatTokens(session.totalTokens)}</span>
                      <span className="session-cost">{formatCost(session.totalCost)}</span>
                    </div>
                    <button 
                      className="replay-button"
                      onClick={(e) => {
                        e.stopPropagation(); // Prevent card selection
                        setReplaySessionId(session.sessionKey);
                      }}
                      title="View Session Replay"
                    >
                      🎬 Replay
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="session-detail">
          {selectedSession ? (
            <>
              <div className="detail-header">
                <h3>Session Details</h3>
                <button
                  onClick={() => setSelectedSession(null)}
                  className="close-button"
                >
                  ✕
                </button>
              </div>

              {loadingDetail ? (
                <div className="loading-detail">
                  <div className="loading-spinner"></div>
                  <p>Loading session history...</p>
                </div>
              ) : (
                <>
                  <div className="detail-info">
                    <div className="detail-stat">
                      <strong>Agent:</strong> {selectedSession.agentName}
                    </div>
                    <div className="detail-stat">
                      <strong>Session Key:</strong> {selectedSession.sessionKey}
                    </div>
                    <div className="detail-stat">
                      <strong>Status:</strong>
                      <span 
                        style={{ color: getStatusColor(selectedSession.status) }}
                      >
                        {getStatusIcon(selectedSession.status)} {selectedSession.status}
                      </span>
                    </div>
                    <div className="detail-stat">
                      <strong>Created:</strong> {timeAgo(selectedSession.created)}
                    </div>
                    <div className="detail-stat">
                      <strong>Tokens:</strong> {formatTokens(selectedSession.totalTokens)}
                    </div>
                    <div className="detail-stat">
                      <strong>Cost:</strong> {formatCost(selectedSession.totalCost)}
                    </div>
                  </div>

                  <div className="messages-container">
                    <h4>Message History</h4>
                    {selectedSession.messages.length === 0 ? (
                      <p className="no-messages">No messages found in this session.</p>
                    ) : (
                      <div className="messages-list">
                        {selectedSession.messages.map((message, index) => (
                          <div key={index} className={`message message-${message.role}`}>
                            <div className="message-header">
                              <span className="message-role">
                                {message.role === 'user' ? '👤' : message.role === 'assistant' ? '🤖' : '⚙️'}
                                {message.role}
                              </span>
                              <span className="message-time">
                                {timeAgo(message.timestamp)}
                              </span>
                              {message.tokens && (
                                <span className="message-tokens">
                                  🔮 {formatTokens(message.tokens)}
                                </span>
                              )}
                            </div>
                            <div className="message-content">
                              <pre>{message.content}</pre>
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                </>
              )}
            </>
          ) : (
            <div className="detail-placeholder">
              <h3>📜 Select a Session</h3>
              <p>Choose a session from the list to view its detailed history and message log.</p>
            </div>
          )}
        </div>
      </div>

      {replaySessionId && (
        <SessionReplay
          sessionId={replaySessionId}
          onClose={() => setReplaySessionId(null)}
        />
      )}
    </div>
  );
}