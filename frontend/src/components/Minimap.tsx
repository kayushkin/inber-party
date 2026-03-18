import { useLocation } from 'react-router-dom';
import { useStore } from '../store';
import { classColor } from '../store';
import './Minimap.css';

interface Room {
  id: string;
  name: string;
  emoji: string;
  path: string;
  x: number;
  y: number;
  width: number;
  height: number;
}

const ROOMS: Room[] = [
  { id: 'tavern', name: 'Tavern', emoji: '🏕️', path: '/', x: 2, y: 3, width: 3, height: 2 },
  { id: 'war-room', name: 'War Room', emoji: '⚔️', path: '/war-room', x: 6, y: 2, width: 2, height: 2 },
  { id: 'guild-hall', name: 'Guild Hall', emoji: '👑', path: '/guild-chat', x: 1, y: 1, width: 2, height: 1 },
  { id: 'library', name: 'Library', emoji: '📚', path: '/library', x: 4, y: 1, width: 2, height: 1 },
  { id: 'training', name: 'Training', emoji: '🏋️', path: '/training', x: 7, y: 1, width: 2, height: 1 },
  { id: 'forge', name: 'Forge', emoji: '🔨', path: '/forge', x: 6, y: 4, width: 2, height: 1 },
  { id: 'quarters', name: 'Quarters', emoji: '🛏️', path: '/quarters', x: 1, y: 5, width: 4, height: 1 },
];

export default function Minimap() {
  const location = useLocation();
  const agents = useStore((s) => s.agents);

  // Determine which agents are in which rooms
  const getAgentRoom = (agent: any) => {
    // If we're viewing an agent's quarters, they're in quarters
    if (location.pathname.startsWith('/quarters/') && location.pathname.includes(agent.id)) {
      return 'quarters';
    }
    
    // Map based on agent status and activity
    switch (agent.status) {
      case 'working':
      case 'busy':
        // Working agents could be in war room if on active quests, otherwise forge
        return 'war-room';
      case 'error':
      case 'stuck':
        // Stuck agents need help - they're likely in war room for debugging
        return 'war-room';
      case 'idle':
      default:
        // Idle agents hang out in the tavern
        return 'tavern';
    }
  };

  // Group agents by room
  const agentsByRoom = agents.reduce((acc, agent) => {
    const room = getAgentRoom(agent);
    if (!acc[room]) acc[room] = [];
    acc[room].push(agent);
    return acc;
  }, {} as Record<string, any[]>);

  const currentRoom = ROOMS.find(room => {
    if (location.pathname === room.path) return true;
    if (room.id === 'quarters' && location.pathname.startsWith('/quarters/')) return true;
    if (room.id === 'quarters' && location.pathname.startsWith('/agent/')) return true;
    return false;
  })?.id;

  return (
    <div className="minimap">
      <div className="minimap-title">🗺️ Guild Map</div>
      <div className="minimap-grid">
        {ROOMS.map((room) => {
          const roomAgents = agentsByRoom[room.id] || [];
          const isCurrentRoom = currentRoom === room.id;
          
          return (
            <div
              key={room.id}
              className={`minimap-room ${isCurrentRoom ? 'current-room' : ''}`}
              style={{
                gridColumn: `${room.x} / span ${room.width}`,
                gridRow: `${room.y} / span ${room.height}`,
              }}
            >
              <div className="room-header">
                <span className="room-emoji">{room.emoji}</span>
                <span className="room-name">{room.name}</span>
                <span className="agent-count">({roomAgents.length})</span>
              </div>
              
              <div className="room-agents">
                {roomAgents.slice(0, 8).map((agent) => (
                  <div
                    key={agent.id}
                    className="minimap-agent"
                    style={{
                      backgroundColor: classColor(agent.class),
                      opacity: agent.status === 'working' ? 1 : 0.7,
                    }}
                    title={`${agent.name} (${agent.class}) - ${agent.status}`}
                  >
                    {agent.avatar_emoji}
                  </div>
                ))}
                {roomAgents.length > 8 && (
                  <div className="agent-overflow">
                    +{roomAgents.length - 8}
                  </div>
                )}
              </div>
            </div>
          );
        })}
      </div>
      
      <div className="minimap-legend">
        <div className="legend-item">
          <div className="legend-dot current"></div>
          <span>Your Location</span>
        </div>
        <div className="legend-item">
          <div className="legend-dot working"></div>
          <span>Working Agents</span>
        </div>
        <div className="legend-item">
          <div className="legend-dot idle"></div>
          <span>Idle Agents</span>
        </div>
      </div>
    </div>
  );
}