import { Link, Outlet, useLocation } from 'react-router-dom';
import { useStore } from '../store';
import ThemeToggle from './ThemeToggle';
import './Layout.css';

export default function Layout() {
  const connected = useStore((s) => s.connected);
  const location = useLocation();

  const navItems = [
    { to: '/', label: '🏕️ Tavern', match: '/' },
    { to: '/quests', label: '📜 Quests', match: '/quests' },
    { to: '/stats', label: '📊 Stats', match: '/stats' },
  ];

  return (
    <div className="layout">
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
    </div>
  );
}
