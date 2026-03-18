import { useState, useRef, useEffect } from 'react';
import { useStore, classColor } from '../store';
import type { RPGAgent } from '../store';
import './GuildMasterChat.css';

interface GuildMessage {
  id: string;
  role: 'guildmaster' | 'adventurer' | 'system';
  content: string;
  timestamp: string;
  agent?: RPGAgent;
  agentId?: string;
}

export default function GuildMasterChat() {
  const [input, setInput] = useState('');
  const [selectedTarget, setSelectedTarget] = useState<string>('');
  const [messages, setMessages] = useState<GuildMessage[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const agents = useStore((s) => s.agents);
  const sendMessage = useStore((s) => s.sendMessage);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  // Add welcome message on mount
  useEffect(() => {
    if (messages.length === 0) {
      setMessages([{
        id: 'welcome',
        role: 'system',
        content: 'Welcome to the Guild Hall, Guild Master! Select an adventurer and give them orders, or broadcast a message to your entire party.',
        timestamp: new Date().toISOString(),
      }]);
    }
  }, [messages.length]);

  const activeAgents = agents.filter(agent => agent.status !== 'offline');
  const targetAgent = selectedTarget && selectedTarget !== 'broadcast' 
    ? agents.find(a => a.id === selectedTarget) 
    : null;

  const handleSend = async () => {
    const msg = input.trim();
    if (!msg || isLoading) return;

    if (!selectedTarget) {
      alert('Please select an adventurer or choose "Broadcast to All"');
      return;
    }

    const newMessage: GuildMessage = {
      id: Date.now().toString(),
      role: 'guildmaster',
      content: msg,
      timestamp: new Date().toISOString(),
      agentId: selectedTarget === 'broadcast' ? undefined : selectedTarget,
    };

    setMessages(prev => [...prev, newMessage]);
    setInput('');
    setIsLoading(true);

    try {
      if (selectedTarget === 'broadcast') {
        // Send to all active agents
        await Promise.allSettled(
          activeAgents.map(agent => sendMessage(agent.id, `[Guild Master broadcast]: ${msg}`))
        );
        
        // Add system message about broadcast
        setMessages(prev => [...prev, {
          id: Date.now().toString() + '_broadcast',
          role: 'system',
          content: `📢 Message broadcasted to ${activeAgents.length} adventurers`,
          timestamp: new Date().toISOString(),
        }]);
        
      } else {
        // Send to specific agent
        await sendMessage(selectedTarget, `[Guild Master order]: ${msg}`);
        
        // Add response placeholder (this is simplified - in a real implementation, 
        // you'd want to hook into the chat system or use WebSocket updates)
        setTimeout(() => {
          setMessages(prev => [...prev, {
            id: Date.now().toString() + '_response',
            role: 'adventurer',
            content: `Order received and acknowledged, Guild Master! I'll get right on it.`,
            timestamp: new Date().toISOString(),
            agent: targetAgent || undefined,
            agentId: selectedTarget,
          }]);
        }, 1000);
      }
    } catch (error) {
      setMessages(prev => [...prev, {
        id: Date.now().toString() + '_error',
        role: 'system',
        content: `Failed to send message: ${error}`,
        timestamp: new Date().toISOString(),
      }]);
    } finally {
      setIsLoading(false);
    }
  };

  const formatTime = (timestamp: string) => {
    return new Date(timestamp).toLocaleTimeString([], { 
      hour: '2-digit', 
      minute: '2-digit' 
    });
  };

  return (
    <div className="guild-master-chat">
      <div className="chat-header">
        <div className="guild-master-info">
          <span className="guild-master-avatar">👑</span>
          <div>
            <h1 className="guild-master-title">Guild Master's Hall</h1>
            <p className="guild-master-subtitle">Command your adventurers</p>
          </div>
        </div>
        
        <div className="target-selector">
          <label htmlFor="target-select">Send to:</label>
          <select 
            id="target-select"
            value={selectedTarget} 
            onChange={(e) => setSelectedTarget(e.target.value)}
            className="target-dropdown"
          >
            <option value="">Choose adventurer...</option>
            <option value="broadcast">📢 Broadcast to All</option>
            {activeAgents.map(agent => (
              <option key={agent.id} value={agent.id}>
                {agent.avatar_emoji} {agent.name} (Lv {agent.level} {agent.class})
              </option>
            ))}
          </select>
        </div>
      </div>

      <div className="chat-messages">
        {messages.map((msg) => (
          <div key={msg.id} className={`guild-message guild-message-${msg.role}`}>
            <div className="message-header">
              <span className="message-sender">
                {msg.role === 'guildmaster' && '👑 Guild Master'}
                {msg.role === 'adventurer' && msg.agent && (
                  <span style={{ color: classColor(msg.agent.class) }}>
                    {msg.agent.avatar_emoji} {msg.agent.name}
                  </span>
                )}
                {msg.role === 'system' && '⚡ System'}
              </span>
              <span className="message-time">{formatTime(msg.timestamp)}</span>
            </div>
            <div className="message-content">
              {msg.content}
              {msg.agentId && msg.agentId !== 'broadcast' && (
                <span className="message-target">
                  → {targetAgent?.name || msg.agentId}
                </span>
              )}
            </div>
          </div>
        ))}
        {isLoading && (
          <div className="guild-message guild-message-system">
            <div className="message-content">
              <span className="loading-indicator">⏳ Sending orders...</span>
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      <div className="chat-input-area">
        <input
          className="guild-input"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleSend()}
          placeholder={
            selectedTarget === 'broadcast' 
              ? "Send orders to all adventurers..." 
              : selectedTarget 
                ? `Give orders to ${targetAgent?.name || 'adventurer'}...`
                : "Select an adventurer first..."
          }
          disabled={isLoading || !selectedTarget}
        />
        <button
          className="guild-send"
          onClick={handleSend}
          disabled={isLoading || !input.trim() || !selectedTarget}
        >
          {isLoading ? '⏳' : selectedTarget === 'broadcast' ? '📢' : '⚔️'}
        </button>
      </div>

      <div className="active-adventurers">
        <h3>Active Adventurers ({activeAgents.length})</h3>
        <div className="adventurer-list">
          {activeAgents.map(agent => (
            <div 
              key={agent.id} 
              className={`adventurer-card ${selectedTarget === agent.id ? 'selected' : ''}`}
              onClick={() => setSelectedTarget(agent.id)}
            >
              <span className="adventurer-avatar">{agent.avatar_emoji}</span>
              <div className="adventurer-info">
                <div className="adventurer-name" style={{ color: classColor(agent.class) }}>
                  {agent.name}
                </div>
                <div className="adventurer-details">
                  Lv {agent.level} {agent.class} • {agent.status}
                </div>
              </div>
              <div className="adventurer-stats">
                <div className="stat">⚡ {agent.energy}/{agent.max_energy}</div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}