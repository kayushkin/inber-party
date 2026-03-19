import { useState, useEffect } from 'react';
import { useStore, classColor } from '../store';
import './AgentConversations.css';

interface Conversation {
  id: string;
  participant_ids?: string[];
  participants?: string[];
  agent_id?: string;
  agent_name?: string;
  title: string;
  messages: Message[];
  started_at?: string;
  start_time?: string;
  last_active?: string;
  end_time?: string;
  type?: string;
  message_count?: number;
  source: 'inber' | 'logstack';
}

interface LogstackConversationResponse {
  id: string;
  agent_name?: string;
  title: string;
  messages: Message[];
  started_at?: string;
  last_active?: string;
  type?: string;
  message_count?: number;
}

interface InberConversationResponse {
  id: string;
  participant_ids?: string[];
  agent_id?: string;
  title: string;
  messages: Message[];
  start_time?: string;
  end_time?: string;
}

interface Message {
  id: string;
  from_agent?: string;
  to_agent?: string;
  role?: string;
  content: string;
  timestamp: string;
  type?: string;
  has_tools?: boolean;
}

export default function AgentConversations() {
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [selectedConversation, setSelectedConversation] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'all' | 'logstack' | 'inber'>('all');
  const agents = useStore((s) => s.agents);

  const fetchConversations = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const [logstackRes, inberRes] = await Promise.allSettled([
        fetch('/api/conversations?limit=30'),
        fetch('/api/inber/conversations?limit=30')
      ]);

      const allConversations: Conversation[] = [];

      // Process logstack conversations
      if (logstackRes.status === 'fulfilled' && logstackRes.value.ok) {
        const logstackData = await logstackRes.value.json();
        const logstackConvs = (logstackData || []).map((conv: LogstackConversationResponse) => ({
          ...conv,
          source: 'logstack' as const,
          participants: conv.agent_name ? [conv.agent_name] : [],
          started_at: conv.started_at,
          last_active: conv.last_active,
          type: 'session_chat'
        }));
        allConversations.push(...logstackConvs);
      }

      // Process inber conversations  
      if (inberRes.status === 'fulfilled' && inberRes.value.ok) {
        const inberData = await inberRes.value.json();
        const inberConvs = (inberData || []).map((conv: InberConversationResponse) => ({
          ...conv,
          source: 'inber' as const
        }));
        allConversations.push(...inberConvs);
      }

      // Sort by timestamp (newest first)
      allConversations.sort((a, b) => {
        const timeA = new Date(a.last_active || a.end_time || a.started_at || a.start_time || 0).getTime();
        const timeB = new Date(b.last_active || b.end_time || b.started_at || b.start_time || 0).getTime();
        return timeB - timeA;
      });

      setConversations(allConversations);
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

  const formatMessageContent = (content: string, type?: string) => {
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

  const getFilteredConversations = () => {
    return conversations.filter(conv => {
      if (activeTab === 'all') return true;
      if (activeTab === 'logstack') return conv.source === 'logstack';
      if (activeTab === 'inber') return conv.source === 'inber';
      return true;
    });
  };

  const filteredConversations = getFilteredConversations();
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
        <p className="page-subtitle">Watch adventurers collaborate and view their complete conversation history</p>
        <button onClick={fetchConversations} className="refresh-button">
          🔄 Refresh
        </button>
      </div>

      <div className="conversation-tabs">
        <button 
          className={`tab-button ${activeTab === 'all' ? 'active' : ''}`}
          onClick={() => setActiveTab('all')}
        >
          📚 All ({conversations.length})
        </button>
        <button 
          className={`tab-button ${activeTab === 'logstack' ? 'active' : ''}`}
          onClick={() => setActiveTab('logstack')}
        >
          💬 Chat History ({conversations.filter(c => c.source === 'logstack').length})
        </button>
        <button 
          className={`tab-button ${activeTab === 'inber' ? 'active' : ''}`}
          onClick={() => setActiveTab('inber')}
        >
          🤝 Inter-Agent ({conversations.filter(c => c.source === 'inber').length})
        </button>
      </div>

      <div className="conversations-layout">
        <div className="conversations-sidebar">
          <h3>Recent Conversations ({filteredConversations.length})</h3>
          {filteredConversations.length === 0 ? (
            <div className="no-conversations">
              <div className="no-content-icon">💬</div>
              <p>No conversations found</p>
              <small>
                {activeTab === 'logstack' && 'Chat history will appear from OpenClaw session logs.'}
                {activeTab === 'inber' && 'Inter-agent conversations will appear when agents spawn sub-agents or collaborate.'}
                {activeTab === 'all' && 'Conversations will appear from both chat history and agent collaboration.'}
              </small>
            </div>
          ) : (
            <div className="conversation-list">
              {filteredConversations.map((conv) => {
                const participants = conv.participants || (conv.agent_name ? [conv.agent_name] : []);
                const messageCount = conv.message_count || conv.messages?.length || 0;
                const lastActive = conv.last_active || conv.end_time;
                
                return (
                  <div 
                    key={`${conv.source}-${conv.id}`}
                    className={`conversation-card ${selectedConversation === conv.id ? 'selected' : ''}`}
                    onClick={() => setSelectedConversation(conv.id)}
                  >
                    <div className="conversation-header">
                      <div className="conversation-title">{conv.title}</div>
                      <div className="conversation-type">
                        <span className={`source-badge ${conv.source}`}>
                          {conv.source === 'logstack' ? '💬' : '🤝'}
                        </span>
                        {conv.type === 'spawn_chain' && '🎯'}
                        {conv.type === 'session_chat' && '💬'}
                        {conv.type === 'inter_agent' && '🤝'}
                      </div>
                    </div>
                    <div className="conversation-participants">
                      {participants.map((participant, idx) => {
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
                      <span>{messageCount} messages</span>
                      {lastActive && <span>{formatTime(lastActive)}</span>}
                      <span className={`source-label ${conv.source}`}>
                        {conv.source === 'logstack' ? 'Session' : 'Inter-agent'}
                      </span>
                    </div>
                  </div>
                );
              })}
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
                  <span>{selectedConv.messages?.length || selectedConv.message_count || 0} messages</span>
                  <span>Started {formatTime(selectedConv.started_at || selectedConv.start_time || '')}</span>
                  <span>Source: {selectedConv.source}</span>
                  {selectedConv.type && <span>Type: {selectedConv.type.replace('_', ' ')}</span>}
                </div>
              </div>

              <div className="conversation-messages">
                {(selectedConv.messages || []).map((message, idx) => {
                  // Handle both logstack and inber message formats
                  const senderName = message.from_agent || selectedConv.agent_name || 'Agent';
                  const agent = getAgentInfo(senderName);
                  const isSystemMessage = message.type === 'system' || message.type === 'spawn';
                  const isUserMessage = message.role === 'user';
                  const hasTools = message.has_tools;
                  
                  return (
                    <div 
                      key={`${message.id}-${idx}`}
                      className={`message ${isSystemMessage ? 'system-message' : isUserMessage ? 'user-message' : 'agent-message'}`}
                    >
                      <div className="message-header">
                        <span 
                          className="message-sender"
                          style={{ color: agent ? classColor(agent.class) : (isUserMessage ? '#4a7c2e' : '#d4af37') }}
                        >
                          {isUserMessage ? '👤 User' : `${agent?.avatar_emoji || '⚔️'} ${senderName}`}
                        </span>
                        {message.to_agent && (
                          <span className="message-target">
                            → {message.to_agent}
                          </span>
                        )}
                        <span className="message-time">
                          {formatTime(message.timestamp)}
                        </span>
                        {message.type && (
                          <span className={`message-type ${message.type}`}>
                            {message.type}
                          </span>
                        )}
                        {hasTools && (
                          <span className="tools-indicator" title="Used tools">
                            🔧
                          </span>
                        )}
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