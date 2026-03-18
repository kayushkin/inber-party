import { Link, Outlet, useLocation } from 'react-router-dom';
import { useStore } from '../store';
import ThemeToggle from './ThemeToggle';
import SoundToggle from './SoundToggle';
import TTSToggle from './TTSToggle';
import SoundEffects from './SoundEffects';
import AnimatedBackground from './AnimatedBackground';
import './Layout.css';

export default function Layout() {
  const connected = useStore((s) => s.connected);
  const location = useLocation();

  const navItems = [
    { to: '/', label: '🏕️ Tavern', match: '/' },
    { to: '/quests', label: '📜 Quests', match: '/quests' },
    { to: '/war-room', label: '⚔️ War Room', match: '/war-room' },
    { to: '/guild-chat', label: '👑 Guild Hall', match: '/guild-chat' },
    { to: '/conversations', label: '🗣️ Conversations', match: '/conversations' },
    { to: '/library', label: '📚 Library', match: '/library' },
    { to: '/training', label: '🏋️ Training', match: '/training' },
    { to: '/stats', label: '📊 Stats', match: '/stats' },
    { to: '/compare', label: '🔍 Compare', match: '/compare' },
  ];

  return (
    <div className="layout">
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
        <Outlet />
      </main>
      <SoundEffects />
    </div>
  );
}
