import { Link, Outlet } from 'react-router-dom';
import { useStore } from '../store';
import './Layout.css';

export default function Layout() {
  const connected = useStore((state) => state.connected);

  return (
    <div className="layout">
      <header className="header">
        <div className="header-content">
          <h1 className="camp-name">⚔️ Míl Party</h1>
          <nav className="nav">
            <Link to="/" className="nav-link">Camp</Link>
            <Link to="/quests" className="nav-link">Quest Board</Link>
          </nav>
          <div className="connection-status">
            <span className={`status-dot ${connected ? 'connected' : 'disconnected'}`} />
            <span className="status-text">{connected ? 'Live' : 'Offline'}</span>
          </div>
        </div>
      </header>
      <main className="main">
        <Outlet />
      </main>
    </div>
  );
}
