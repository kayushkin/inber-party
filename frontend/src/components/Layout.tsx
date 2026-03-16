import { Link, Outlet } from 'react-router-dom';
import { useStore, api } from '../store';
import './Layout.css';

export default function Layout() {
  const connected = useStore((state) => state.connected);
  const isDemo = api.isDemoMode();
  const isInber = api.isInberMode();

  return (
    <div className="layout">
      {isInber && (
        <div className="demo-banner" style={{ background: 'linear-gradient(90deg, #002a00, #005500, #002a00)', color: '#00ff88', textAlign: 'center', padding: '4px 0', fontSize: '12px', fontFamily: 'monospace', letterSpacing: '2px' }}>
          🔮 LIVE — Real agent activity from inber 🔮
        </div>
      )}
      {isDemo && !isInber && (
        <div className="demo-banner" style={{ background: 'linear-gradient(90deg, #4a2800, #7a4400, #4a2800)', color: '#ffd700', textAlign: 'center', padding: '4px 0', fontSize: '12px', fontFamily: 'monospace', letterSpacing: '2px' }}>
          ⚡ DEMO MODE — Backend unavailable, showing sample data ⚡
        </div>
      )}
      <header className="header">
        <div className="header-content">
          <h1 className="camp-name">⚔️ Míl Party</h1>
          <nav className="nav">
            <Link to="/" className="nav-link">Camp</Link>
            <Link to="/quests" className="nav-link">Quest Board</Link>
          </nav>
          <div className="connection-status">
            <span className={`status-dot ${connected ? 'connected' : 'disconnected'}`} />
            <span className="status-text">{connected ? 'Live' : isDemo ? 'Demo' : 'Offline'}</span>
          </div>
        </div>
      </header>
      <main className="main">
        <Outlet />
      </main>
    </div>
  );
}
