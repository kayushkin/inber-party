import { useState, useRef, useEffect } from 'react';
import { useStore, classColor } from '../store';
import type { RPGAgent } from '../store';
import './ChatPanel.css';

interface Props {
  agentId: string;
  agent?: RPGAgent;
  onClose: () => void;
}

export default function ChatPanel({ agentId, agent, onClose }: Props) {
  const [input, setInput] = useState('');
  const messages = useStore((s) => s.chatMessages[agentId] || []);
  const loading = useStore((s) => s.chatLoading[agentId] || false);
  const sendMessage = useStore((s) => s.sendMessage);
  const clearChatHistory = useStore((s) => s.clearChatHistory);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSend = () => {
    const msg = input.trim();
    if (!msg || loading) return;
    setInput('');
    sendMessage(agentId, msg);
  };

  const handleClearHistory = () => {
    if (confirm(`Clear chat history with ${agent?.name || agentId}?`)) {
      clearChatHistory(agentId);
    }
  };

  const cc = agent ? classColor(agent.class) : '#d4af37';

  return (
    <div className="chat-panel" style={{ '--cc': cc } as React.CSSProperties}>
      <div className="chat-header">
        <div className="chat-agent-info">
          <span className="chat-avatar">{agent?.avatar_emoji || '⚔️'}</span>
          <div>
            <div className="chat-agent-name" style={{ color: cc }}>{agent?.name || agentId}</div>
            <div className="chat-agent-class">{agent?.class} · Lv {agent?.level}</div>
          </div>
        </div>
        <button className="chat-close" onClick={onClose}>✕</button>
      </div>

      {messages.length > 0 && (
        <div className="chat-history-info">
          <div className="chat-history-indicator">
            <span>💾</span>
            <span>{messages.length} messages saved</span>
          </div>
          <button className="chat-clear-btn" onClick={handleClearHistory}>
            Clear History
          </button>
        </div>
      )}

      <div className="chat-messages">
        {messages.length === 0 && (
          <div className="chat-empty">
            <p>Send a quest to <strong>{agent?.name || agentId}</strong></p>
            <p className="chat-hint">Messages are sent as tasks via inber's /api/run</p>
          </div>
        )}
        {messages.map((msg, i) => (
          <div key={i} className={`chat-msg chat-msg-${msg.role}`}>
            <div className="chat-bubble">
              {msg.content || (msg.streaming ? '...' : '')}
              {msg.streaming && <span className="chat-typing">▋</span>}
            </div>
          </div>
        ))}
        <div ref={messagesEndRef} />
      </div>

      <div className="chat-input-area">
        <input
          className="chat-input"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleSend()}
          placeholder={`Send quest to ${agent?.name || agentId}...`}
          disabled={loading}
        />
        <button
          className="chat-send"
          onClick={handleSend}
          disabled={loading || !input.trim()}
        >
          {loading ? '⏳' : '⚔️'}
        </button>
      </div>
    </div>
  );
}
