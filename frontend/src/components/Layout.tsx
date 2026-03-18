import { Link, Outlet, useLocation } from 'react-router-dom';
import { useState } from 'react';
import { useStore } from '../store';
import ThemeToggle from './ThemeToggle';
import SoundToggle from './SoundToggle';
import TTSToggle from './TTSToggle';
import SoundEffects from './SoundEffects';
import AnimatedBackground from './AnimatedBackground';
import PageTransition from './PageTransition';
import Minimap from './Minimap';
import './Layout.css';

export default function Layout() {
  const connected = useStore((s) => s.connected);
  const location = useLocation();
  const [showMinimap, setShowMinimap] = useState(false);

  const navItems = [
    { to: '/', label: '🏕️ Tavern', match: '/' },
    { to: '/quests', label: '📜 Quests', match: '/quests' },
    { to: '/war-room', label: '⚔️ War Room', match: '/war-room' },
    { to: '/guild-chat', label: '👑 Guild Hall', match: '/guild-chat' },
    { to: '/conversations', label: '🗣️ Conversations', match: '/conversations' },
    { to: '/library', label: '📚 Library', match: '/library' },
    { to: '/training', label: '🏋️ Training', match: '/training' },
    { to: '/forge', label: '🔨 Forge', match: '/forge' },
    { to: '/stats', label: '📊 Stats', match: '/stats' },
    { to: '/compare', label: '🔍 Compare', match: '/compare' },
  ];

  // Map paths to room classes for ambient styling
  const getRoomClass = (path: string) => {
    if (path === '/') return 'room-tavern';
    if (path.startsWith('/agent/') || path.startsWith('/quarters/')) return 'room-quarters';
    if (path === '/quests') return 'room-quests';
    if (path === '/war-room') return 'room-war-room';
    if (path === '/guild-chat') return 'room-guild-hall';
    if (path === '/conversations') return 'room-conversations';
    if (path === '/library') return 'room-library';
    if (path === '/training') return 'room-training';
    if (path === '/forge') return 'room-forge';
    if (path === '/stats') return 'room-stats';
    if (path === '/compare') return 'room-compare';
    return 'room-default';
  };

  const roomClass = getRoomClass(location.pathname);

  return (
    <div className={`layout ${roomClass}`}>
      <AnimatedBackground />
      <header className="header">
        <div className="header-content">
          <Link to="/" className="camp-name">⚔️ Míl Party</Link>
          <nav className="nav">
            {navItems.map((item) => (
              <Link
                key={item.to}
                to={item.to}
                className={`nav-link ${location.pathname === item.match ? 'active' : ''}`}
              >
                {item.label}
              </Link>
            ))}
          </nav>
          <div className="header-controls">
            <button
              className={`minimap-toggle ${showMinimap ? 'active' : ''}`}
              onClick={() => setShowMinimap(!showMinimap)}
              title="Toggle Guild Map"
            >
              🗺️
            </button>
            <TTSToggle />
            <SoundToggle />
            <ThemeToggle />
            <div className="connection-status">
              <span className={`status-dot ${connected ? 'connected' : 'disconnected'}`} />
              <span className="status-text">{connected ? 'Live' : 'Offline'}</span>
            </div>
          </div>
        </div>
      </header>
      <main className="main">
        <PageTransition>
          <Outlet />
        </PageTransition>
      </main>
      
      {/* Minimap Sidebar */}
      <div className={`minimap-sidebar ${showMinimap ? 'show' : ''}`}>
        <Minimap />
      </div>
      
      {/* Backdrop for closing minimap */}
      {showMinimap && (
        <div 
          className="minimap-backdrop" 
          onClick={() => setShowMinimap(false)}
        />
      )}
      
      <SoundEffects />
    </div>
  );
}
