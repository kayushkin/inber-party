import { useState, useEffect, useRef } from 'react';
import './MMOChatroom.css';

interface ChatMessage {
  id: string;
  username: string;
  message: string;
  timestamp: string;
  type: 'chat' | 'system' | 'bounty' | 'quest';
  avatar?: string;
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
  const wsRef = useRef<WebSocket | null>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  // Connect to WebSocket for real-time chat
  useEffect(() => {
    if (!username) return;

    const connectChat = () => {
      const wsUrl = 'ws://localhost:8080/api/ws/chat';
      const ws = new WebSocket(wsUrl);
      
      ws.onopen = () => {
        console.log('Connected to chat server');
        setIsConnected(true);
        
        // Send join message
        ws.send(JSON.stringify({
          type: 'join',
          username: username
        }));
      };

      ws.onclose = () => {
        console.log('Disconnected from chat server');
        setIsConnected(false);
        // Attempt to reconnect after 3 seconds
        setTimeout(connectChat, 3000);
      };

      ws.onerror = (error) => {
        console.error('Chat WebSocket error:', error);
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          if (data.type === 'chat_message') {
            setMessages(prev => [...prev, data.message]);
          } else if (data.type === 'chat_history') {
            setMessages(data.messages || []);
          } else if (data.type === 'user_joined') {
            setMessages(prev => [...prev, {
              id: generateChatMessageId('user_joined'),
              username: 'System',
              message: `${data.username} joined the chatroom`,
              timestamp: new Date().toISOString(),
              type: 'system'
            }]);
          } else if (data.type === 'user_left') {
            setMessages(prev => [...prev, {
              id: generateChatMessageId('user_left'),
              username: 'System',
              message: `${data.username} left the chatroom`,
              timestamp: new Date().toISOString(),
              type: 'system'
            }]);
          }
        } catch (error) {
          console.error('Error parsing chat message:', error);
        }
      };

      wsRef.current = ws;
    };

    connectChat();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [username]);

  const sendMessage = () => {
    if (!inputMessage.trim() || !wsRef.current || !isConnected) return;

    const message = {
      type: 'chat_message',
      username: username,
      message: inputMessage.trim(),
      timestamp: new Date().toISOString(),
      avatar: `https://api.dicebear.com/7.x/pixel-art/svg?seed=${username}&size=32`
    };

    wsRef.current.send(JSON.stringify(message));
    setInputMessage('');
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  };

  const changeUsername = () => {
    const newUsername = prompt('Choose a new username:', username);
    if (newUsername && newUsername.trim() && newUsername !== username) {
      setUsername(newUsername.trim());
      localStorage.setItem('mmo_username', newUsername.trim());
      // Reconnect with new username
      if (wsRef.current) {
        wsRef.current.close();
      }
    }
  };

  const formatTime = (timestamp: string) => {
    try {
      const date = new Date(timestamp);
      return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } catch {
      return '';
    }
  };

  const getMessageTypeClass = (type: string) => {
    switch (type) {
      case 'system': return 'message-system';
      case 'bounty': return 'message-bounty';
      case 'quest': return 'message-quest';
      default: return 'message-chat';
    }
  };

  return (
    <div className="mmo-chatroom">
      <div className="chatroom-header">
        <h1>🌍 MMO Trade Chat</h1>
        <div className="chatroom-info">
          <span className={`connection-status ${isConnected ? 'connected' : 'disconnected'}`}>
            {isConnected ? '🟢 Online' : '🔴 Connecting...'}
          </span>
          <button className="username-btn" onClick={changeUsername}>
            👤 {username}
          </button>
        </div>
      </div>

      <div className="chat-container">
        <div className="messages-area">
          {messages.length === 0 ? (
            <div className="no-messages">
              <p>Welcome to the guild chatroom! 🏰</p>
              <p>This is where bounty hunters and quest givers meet to discuss tasks, negotiate terms, and share information.</p>
              <p>Be the first to break the ice!</p>
            </div>
          ) : (
            messages.map((msg) => (
              <div key={msg.id} className={`message ${getMessageTypeClass(msg.type)}`}>
                {msg.type !== 'system' && msg.avatar && (
                  <img 
                    src={msg.avatar} 
                    alt={msg.username}
                    className="message-avatar"
                  />
                )}
                <div className="message-content">
                  <div className="message-header">
                    <span className="message-username">{msg.username}</span>
                    <span className="message-time">{formatTime(msg.timestamp)}</span>
                  </div>
                  <div className="message-text">{msg.message}</div>
                </div>
              </div>
            ))
          )}
          <div ref={messagesEndRef} />
        </div>

        <div className="chat-input-area">
          <div className="chat-input-container">
            <input
              type="text"
              value={inputMessage}
              onChange={(e) => setInputMessage(e.target.value)}
              onKeyPress={handleKeyPress}
              placeholder={isConnected ? "Type your message..." : "Connecting to chat..."}
              disabled={!isConnected}
              className="chat-input"
              maxLength={500}
            />
            <button
              onClick={sendMessage}
              disabled={!isConnected || !inputMessage.trim()}
              className="send-button"
            >
              📨
            </button>
          </div>
          <div className="chat-tips">
            <p>💡 <strong>Tips:</strong> Discuss bounties, ask for clarification on tasks, negotiate terms, or just chat with fellow adventurers!</p>
          </div>
        </div>
      </div>
    </div>
  );
}