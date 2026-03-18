import { useNetworkStatus } from '../services/apiClient';
import './OfflineIndicator.css';

export default function OfflineIndicator() {
  const isOnline = useNetworkStatus();

  if (isOnline) return null;

  return (
    <div className="offline-indicator">
      <div className="offline-content">
        <span className="offline-icon">📡</span>
        <span className="offline-text">You're offline</span>
        <span className="offline-subtext">Some features may be limited</span>
      </div>
    </div>
  );
}