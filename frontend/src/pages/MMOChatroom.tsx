import React, { useState, useEffect, useRef } from 'react';
import { useOptimizedWebSocket } from '../utils/websocket';
import './MMOChatroom.css';

interface ChatMessage {
  id: string;
  username: string;
  message: string;
  timestamp: string;
  type: 'chat' | 'system' | 'bounty' | 'quest';
  avatar?: string;
}

// WebSocket message types for chat
interface ChatWebSocketMessage {
  type: 'chat_message' | 'chat_history' | 'user_joined' | 'user_left';
  message?: ChatMessage;
  messages?: ChatMessage[];
  username?: string;
}

// Type guard for chat WebSocket messages
function isChatWebSocketMessage(data: unknown): data is ChatWebSocketMessage {
  return typeof data === 'object' && data !== null && 
         'type' in data && typeof (data as ChatWebSocketMessage).type === 'string' &&
         ['chat_message', 'chat_history', 'user_joined', 'user_left'].includes((data as ChatWebSocketMessage).type);
}

// Generate unique IDs to avoid React key conflicts
let chatMessageIdCounter = 0;
const generateChatMessageId = (suffix?: string): string => {
  chatMessageIdCounter += 1;
  // Use counter as primary identifier with component prefix for uniqueness
  const uniqueId = `mmo_${chatMessageIdCounter}_${Date.now() % 10000}`;
  return suffix ? `${uniqueId}_${suffix}` : uniqueId;
};

export default function MMOChatroom() {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [inputMessage, setInputMessage] = useState('');
  const [username, setUsername] = useState(() => {
    const savedUsername = localStorage.getItem('mmo_username');
    if (savedUsername) {
      return savedUsername;
    } else {
      const newUsername = prompt('Choose your username for the chatroom:') || `Player${Math.floor(Math.random() * 1000)}`;
      localStorage.setItem('mmo_username', newUsername);
      return newUsername;
    }
  });
  const [isConnected, setIsConnected] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const wsUrl = 'ws://localhost:8080/api/ws/chat';
  
  // Use optimized WebSocket connection
  const { send } = useOptimizedWebSocket(
    wsUrl,
    `mmo-chat-${username}`,
    (data) => {
      // Handle incoming messages
      if (!isChatWebSocketMessage(data)) {
        console.warn('Received invalid chat WebSocket message:', data);
        return;
      }
      
      if (data.type === 'chat_message' && data.message) {
        setMessages(prev => [...prev, data.message!]);
      } else if (data.type === 'chat_history') {
        setMessages(data.messages || []);
      } else if (data.type === 'user_joined' && data.username) {
        setMessages(prev => [...prev, {
          id: generateChatMessageId('user_joined'),
          username: 'System',
          message: `${data.username} joined the chatroom`,
          timestamp: new Date().toISOString(),
          type: 'system'
        }]);
      } else if (data.type === 'user_left' && data.username) {
        setMessages(prev => [...prev, {
          id: generateChatMessageId('user_left'),
          username: 'System',
          message: `${data.username} left the chatroom`,
          timestamp: new Date().toISOString(),
          type: 'system'
        }]);
      }
    },
    (connected) => {
      setIsConnected(connected);
    }
  );

  // Send join message when connected and username is available
  React.useEffect(() => {
    if (isConnected && username && send) {
      send({
        type: 'join',
        username: username
      });
    }
  }, [isConnected, username, send]);

  const sendMessage = () => {
    if (!inputMessage.trim() || !isConnected) return;

    const message = {
      type: 'chat_message',
      username: username,
      message: inputMessage.trim(),
      timestamp: new Date().toISOString(),
      avatar: `https://api.dicebear.com/7.x/pixel-art/svg?seed=${username}&size=32`
    };

    const success = send(message);
    if (success) {
      setInputMessage('');
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  };

  const formatTimestamp = (timestamp: string) => {
    try {
      const date = new Date(timestamp);
      return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } catch {
      return '';
    }
  };

  const changeUsername = () => {
    const newUsername = prompt('Enter new username:', username);
    if (newUsername && newUsername !== username) {
      localStorage.setItem('mmo_username', newUsername);
      setUsername(newUsername);
      
      // Send leave message for old username and join for new one
      if (isConnected) {
        send({
          type: 'username_changed',
          oldUsername: username,
          newUsername: newUsername
        });
      }
    }
  };

  return (
    <div className="mmo-chatroom">
      <div className="chat-header">
        <div className="chat-title">
          <h2>🏰 Tavern Chat</h2>
          <div className={`connection-status ${isConnected ? 'connected' : 'disconnected'}`}>
            {isConnected ? '🟢 Connected' : '🔴 Disconnected'}
          </div>
        </div>
        <div className="user-info">
          <span>👤 {username}</span>
          <button onClick={changeUsername} className="change-username-btn" title="Change username">
            ✏️
          </button>
        </div>
      </div>

      <div className="messages-container">
        <div className="messages-scroll">
          {messages.length === 0 ? (
            <div className="no-messages">
              <p>Welcome to the tavern! Start chatting with fellow adventurers...</p>
            </div>
          ) : (
            messages.map((msg) => (
              <div key={msg.id} className={`message ${msg.type === 'system' ? 'system-message' : 'chat-message'}`}>
                {msg.type !== 'system' && (
                  <div className="message-avatar">
                    {msg.avatar ? (
                      <img src={msg.avatar} alt={`${msg.username} avatar`} />
                    ) : (
                      <div className="default-avatar">
                        {msg.username.charAt(0).toUpperCase()}
                      </div>
                    )}
                  </div>
                )}
                <div className="message-content">
                  <div className="message-header">
                    <span className="username">{msg.username}</span>
                    <span className="timestamp">{formatTimestamp(msg.timestamp)}</span>
                  </div>
                  <div className="message-text">{msg.message}</div>
                </div>
              </div>
            ))
          )}
          <div ref={messagesEndRef} />
        </div>
      </div>

      <div className="message-input-container">
        <div className="input-wrapper">
          <input
            type="text"
            value={inputMessage}
            onChange={(e) => setInputMessage(e.target.value)}
            onKeyPress={handleKeyPress}
            placeholder={isConnected ? "Type a message..." : "Connecting..."}
            disabled={!isConnected}
            maxLength={500}
          />
          <button
            onClick={sendMessage}
            disabled={!inputMessage.trim() || !isConnected}
            className="send-button"
          >
            📤
          </button>
        </div>
        <div className="input-info">
          <span className="char-count">{inputMessage.length}/500</span>
          {!isConnected && (
            <span className="connection-warning">⚠️ Reconnecting...</span>
          )}
        </div>
      </div>
    </div>
  );
}