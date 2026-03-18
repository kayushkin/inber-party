import { useEffect, useRef, useState } from 'react';
import { useLocation } from 'react-router-dom';
import './PageTransition.css';

interface PageTransitionProps {
  children: React.ReactNode;
}

export default function PageTransition({ children }: PageTransitionProps) {
  const location = useLocation();
  const [isTransitioning, setIsTransitioning] = useState(false);
  const [currentContent, setCurrentContent] = useState(children);
  const previousLocationRef = useRef(location.pathname);

  useEffect(() => {
    // Skip transition on initial load
    if (previousLocationRef.current === location.pathname) {
      return;
    }

    // Start transition
    setIsTransitioning(true);

    // Wait for exit animation, then update content
    const timer = setTimeout(() => {
      setCurrentContent(children);
      setIsTransitioning(false);
      previousLocationRef.current = location.pathname;
    }, 150); // Half of the total transition duration

    return () => clearTimeout(timer);
  }, [location.pathname, children]);

  return (
    <div className="page-transition-container">
      <div className={`page-content ${isTransitioning ? 'transitioning' : ''}`}>
        {currentContent}
      </div>
      {isTransitioning && (
        <div className="room-transition-overlay">
          <div className="transition-text">
            🚪 Moving to {getRoomName(location.pathname)}...
          </div>
          <div className="pixel-spinner">
            <div className="spinner-blade" />
            <div className="spinner-blade" />
            <div className="spinner-blade" />
            <div className="spinner-blade" />
          </div>
        </div>
      )}
    </div>
  );
}

function getRoomName(pathname: string): string {
  const roomNames: Record<string, string> = {
    '/': 'Tavern',
    '/quests': 'Quest Board',
    '/war-room': 'War Room',
    '/guild-chat': 'Guild Hall',
    '/conversations': 'Conversations',
    '/library': 'Library',
    '/training': 'Training Grounds',
    '/forge': 'Forge',
    '/stats': 'Stats Hall',
    '/compare': 'Comparison Chamber',
  };

  // Handle dynamic routes like /agent/:id and /quarters/:id
  if (pathname.startsWith('/agent/')) return 'Character Sheet';
  if (pathname.startsWith('/quarters/')) return 'Agent Quarters';

  return roomNames[pathname] || 'Unknown Room';
}