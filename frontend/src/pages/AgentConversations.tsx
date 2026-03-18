import { useState, useEffect } from 'react';
import { useStore, classColor } from '../store';
import './AgentConversations.css';

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

export default function AgentConversations() {
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [selectedConversation, setSelectedConversation] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const agents = useStore((s) => s.agents);

  const fetchConversations = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await fetch('/api/inber/conversations?limit=50');
      if (!response.ok) {
        throw new Error(`Failed to fetch conversations: ${response.status} ${response.statusText}`);
      }
      const data = await response.json();
      setConversations(data || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch conversations');
      console.error('Error fetching conversations:', err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchConversations();
  }, []);

  const getAgentInfo = (agentName: string) => {
    const agent = agents.find(a => a.name.toLowerCase() === agentName.toLowerCase());
    return agent;
  };

  const formatTime = (timestamp: string) => {
    try {
      return new Date(timestamp).toLocaleString([], { 
        month: 'short', 
        day: 'numeric', 
        hour: '2-digit', 
        minute: '2-digit' 
      });
    } catch {
      return timestamp;
    }
  };

  const formatMessageContent = (content: string, type: string) => {
    // Special formatting for system messages
    if (type === 'spawn' || type === 'system') {
      return content;
    }
    
    // Truncate very long messages for display
    if (content.length > 300) {
      return content.substring(0, 300) + '...';
    }
    
    return content;
  };

  const selectedConv = conversations.find(c => c.id === selectedConversation);

  if (isLoading) {
    return (
      <div className="agent-conversations">
        <div className="loading-state">
          <div className="loading-spinner">⏳</div>
          <p>Loading agent conversations...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="agent-conversations">
        <div className="error-state">
          <div className="error-icon">⚠️</div>
          <h3>Failed to Load Conversations</h3>
          <p>{error}</p>
          <button onClick={fetchConversations} className="retry-button">
            🔄 Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="agent-conversations">
      <div className="conversations-header">
        <h1 className="page-title">🗣️ Agent Conversations</h1>
        <p className="page-subtitle">Watch adventurers collaborate and spawn new quests</p>
        <button onClick={fetchConversations} className="refresh-button">
          🔄 Refresh
        </button>
      </div>

      <div className="conversations-layout">
        <div className="conversations-sidebar">
          <h3>Recent Conversations ({conversations.length})</h3>
          {conversations.length === 0 ? (
            <div className="no-conversations">
              <div className="no-content-icon">💬</div>
              <p>No inter-agent conversations found</p>
              <small>Conversations will appear when agents spawn sub-agents or collaborate on tasks.</small>
            </div>
          ) : (
            <div className="conversation-list">
              {conversations.map((conv) => (
                <div 
                  key={conv.id}
                  className={`conversation-card ${selectedConversation === conv.id ? 'selected' : ''}`}
                  onClick={() => setSelectedConversation(conv.id)}
                >
                  <div className="conversation-header">
                    <div className="conversation-title">{conv.title}</div>
                    <div className="conversation-type">
                      {conv.type === 'spawn_chain' && '🎯'}
                      {conv.type === 'session_chat' && '💬'}
                      {conv.type === 'inter_agent' && '🤝'}
                    </div>
                  </div>
                  <div className="conversation-participants">
                    {conv.participants.map((participant, idx) => {
                      const agent = getAgentInfo(participant);
                      return (
                        <span 
                          key={idx}
                          className="participant-badge"
                          style={{ color: agent ? classColor(agent.class) : '#d4af37' }}
                        >
                          {agent?.avatar_emoji || '⚔️'} {participant}
                        </span>
                      );
                    })}
                  </div>
                  <div className="conversation-meta">
                    <span>{conv.messages.length} messages</span>
                    <span>{formatTime(conv.last_active)}</span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="conversation-viewer">
          {!selectedConv ? (
            <div className="no-selection">
              <div className="no-selection-icon">👈</div>
              <h3>Select a Conversation</h3>
              <p>Choose a conversation from the list to view the message history between agents.</p>
            </div>
          ) : (
            <div className="conversation-detail">
              <div className="conversation-detail-header">
                <h2>{selectedConv.title}</h2>
                <div className="conversation-detail-meta">
                  <span>{selectedConv.messages.length} messages</span>
                  <span>Started {formatTime(selectedConv.started_at)}</span>
                  <span>Type: {selectedConv.type.replace('_', ' ')}</span>
                </div>
              </div>

              <div className="conversation-messages">
                {selectedConv.messages.map((message, idx) => {
                  const agent = getAgentInfo(message.from_agent);
                  const isSystemMessage = message.type === 'system' || message.type === 'spawn';
                  
                  return (
                    <div 
                      key={`${message.id}-${idx}`}
                      className={`message ${isSystemMessage ? 'system-message' : 'agent-message'}`}
                    >
                      <div className="message-header">
                        <span 
                          className="message-sender"
                          style={{ color: agent ? classColor(agent.class) : '#d4af37' }}
                        >
                          {agent?.avatar_emoji || '⚔️'} {message.from_agent}
                        </span>
                        {message.to_agent && (
                          <span className="message-target">
                            → {message.to_agent}
                          </span>
                        )}
                        <span className="message-time">
                          {formatTime(message.timestamp)}
                        </span>
                        <span className={`message-type ${message.type}`}>
                          {message.type}
                        </span>
                      </div>
                      <div className="message-content">
                        {formatMessageContent(message.content, message.type)}
                      </div>
                    </div>
                  );
                })}
                {selectedConv.messages.length === 0 && (
                  <div className="no-messages">
                    <div className="no-content-icon">💭</div>
                    <p>No messages in this conversation yet.</p>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}