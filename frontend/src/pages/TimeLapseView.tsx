import { useState, useEffect, useRef } from 'react';
import { useStore } from '../store';

interface ActivityEvent {
  timestamp: string;
  type: string;
  agent_id?: number;
  agent_name?: string;
  task_id?: number;
  task_name?: string;
  status?: string;
  description: string;
  metadata?: Record<string, any>;
}

interface TimelineResponse {
  events: ActivityEvent[];
  start_time: string;
  end_time: string;
  total: number;
}

interface TimelineState {
  events: ActivityEvent[];
  currentIndex: number;
  isPlaying: boolean;
  speed: number;
}

const TimeLapseView: React.FC = () => {
  const agents = useStore((s) => s.agents);
  const [timeline, setTimeline] = useState<TimelineState>({
    events: [],
    currentIndex: 0,
    isPlaying: false,
    speed: 1000, // milliseconds between events
  });
  
  const [timeRange, setTimeRange] = useState({
    start: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString().split('T')[0], // yesterday
    end: new Date().toISOString().split('T')[0] // today
  });
  
  const [selectedAgent, setSelectedAgent] = useState<string>('');
  const [loading, setLoading] = useState(false);
  const intervalRef = useRef<number | null>(null);

  // Fetch timeline data
  const fetchTimeline = async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams({
        start: new Date(timeRange.start + 'T00:00:00.000Z').toISOString(),
        end: new Date(timeRange.end + 'T23:59:59.999Z').toISOString(),
        limit: '1000'
      });
      
      if (selectedAgent) {
        params.append('agent_id', selectedAgent);
      }

      const response = await fetch(`/api/activity/timeline?${params}`);
      if (!response.ok) throw new Error('Failed to fetch timeline');
      
      const data: TimelineResponse = await response.json();
      setTimeline(prev => ({
        ...prev,
        events: data.events,
        currentIndex: 0,
        isPlaying: false
      }));
    } catch (error) {
      console.error('Error fetching timeline:', error);
    } finally {
      setLoading(false);
    }
  };

  // Play/pause controls
  const togglePlayback = () => {
    if (timeline.isPlaying) {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
      setTimeline(prev => ({ ...prev, isPlaying: false }));
    } else {
      const interval = setInterval(() => {
        setTimeline(prev => {
          if (prev.currentIndex >= prev.events.length - 1) {
            return { ...prev, isPlaying: false, currentIndex: 0 };
          }
          return { ...prev, currentIndex: prev.currentIndex + 1 };
        });
      }, timeline.speed);
      
      intervalRef.current = interval;
      setTimeline(prev => ({ ...prev, isPlaying: true }));
    }
  };

  const resetPlayback = () => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }
    setTimeline(prev => ({ ...prev, currentIndex: 0, isPlaying: false }));
  };

  const changeSpeed = (newSpeed: number) => {
    setTimeline(prev => ({ ...prev, speed: newSpeed }));
    
    // If playing, restart with new speed
    if (timeline.isPlaying && intervalRef.current) {
      clearInterval(intervalRef.current);
      const interval = setInterval(() => {
        setTimeline(prev => {
          if (prev.currentIndex >= prev.events.length - 1) {
            return { ...prev, isPlaying: false, currentIndex: 0 };
          }
          return { ...prev, currentIndex: prev.currentIndex + 1 };
        });
      }, newSpeed);
      intervalRef.current = interval;
    }
  };

  // Cleanup interval on unmount
  useEffect(() => {
    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    };
  }, []);

  // Get current visible events (current event + recent history)
  const getVisibleEvents = () => {
    const historyLength = 10;
    const startIndex = Math.max(0, timeline.currentIndex - historyLength);
    const endIndex = timeline.currentIndex + 1;
    
    return timeline.events.slice(startIndex, endIndex);
  };

  // Get agent activity visualization
  const getAgentActivity = () => {
    const activeAgents = new Map<number, ActivityEvent>();
    const recentEvents = getVisibleEvents();
    
    // Track the most recent activity for each agent
    recentEvents.forEach(event => {
      if (event.agent_id) {
        activeAgents.set(event.agent_id, event);
      }
    });
    
    return activeAgents;
  };

  const getEventTypeColor = (type: string): string => {
    switch (type) {
      case 'task_created': return '#3b82f6'; // blue
      case 'task_started': return '#10b981'; // green
      case 'task_completed': return '#8b5cf6'; // purple
      case 'agent_updated': return '#f59e0b'; // yellow
      case 'agent_active': return '#ef4444'; // red
      default: return '#6b7280'; // gray
    }
  };

  const formatTimestamp = (timestamp: string): string => {
    return new Date(timestamp).toLocaleTimeString();
  };

  return (
    <div className="time-lapse-view">
      <div className="bg-gradient-to-br from-purple-900 via-blue-900 to-indigo-900 min-h-screen">
        <div className="container mx-auto px-4 py-8">
          <div className="mb-8">
            <h1 className="text-4xl font-bold text-white mb-4 text-center">
              🎬 Adventure Time-Lapse
            </h1>
            <p className="text-purple-200 text-center text-lg">
              Watch your agents' daily activities compressed into an animated timeline
            </p>
          </div>

          {/* Controls */}
          <div className="bg-black/30 backdrop-blur-sm rounded-xl p-6 mb-8">
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-6">
              {/* Time Range */}
              <div>
                <label className="block text-purple-200 text-sm font-medium mb-2">
                  Time Range
                </label>
                <div className="flex gap-2">
                  <input
                    type="date"
                    value={timeRange.start}
                    onChange={(e) => setTimeRange(prev => ({ ...prev, start: e.target.value }))}
                    className="flex-1 bg-gray-800 border border-gray-600 rounded-lg px-3 py-2 text-white"
                  />
                  <input
                    type="date"
                    value={timeRange.end}
                    onChange={(e) => setTimeRange(prev => ({ ...prev, end: e.target.value }))}
                    className="flex-1 bg-gray-800 border border-gray-600 rounded-lg px-3 py-2 text-white"
                  />
                </div>
              </div>

              {/* Agent Filter */}
              <div>
                <label className="block text-purple-200 text-sm font-medium mb-2">
                  Agent Filter
                </label>
                <select
                  value={selectedAgent}
                  onChange={(e) => setSelectedAgent(e.target.value)}
                  className="w-full bg-gray-800 border border-gray-600 rounded-lg px-3 py-2 text-white"
                >
                  <option value="">All Agents</option>
                  {agents.map(agent => (
                    <option key={agent.id} value={agent.id}>
                      {agent.name}
                    </option>
                  ))}
                </select>
              </div>

              {/* Load Button */}
              <div className="flex items-end">
                <button
                  onClick={fetchTimeline}
                  disabled={loading}
                  className="w-full bg-purple-600 hover:bg-purple-700 disabled:bg-gray-600 text-white font-medium py-2 px-4 rounded-lg transition-colors"
                >
                  {loading ? 'Loading...' : 'Load Timeline'}
                </button>
              </div>
            </div>

            {/* Playback Controls */}
            {timeline.events.length > 0 && (
              <div className="border-t border-gray-600 pt-6">
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center gap-4">
                    <button
                      onClick={togglePlayback}
                      className="bg-green-600 hover:bg-green-700 text-white px-4 py-2 rounded-lg transition-colors"
                    >
                      {timeline.isPlaying ? '⏸️ Pause' : '▶️ Play'}
                    </button>
                    <button
                      onClick={resetPlayback}
                      className="bg-gray-600 hover:bg-gray-700 text-white px-4 py-2 rounded-lg transition-colors"
                    >
                      ⏮️ Reset
                    </button>
                  </div>

                  <div className="flex items-center gap-4">
                    <span className="text-purple-200 text-sm">Speed:</span>
                    <button
                      onClick={() => changeSpeed(2000)}
                      className={`px-3 py-1 rounded ${timeline.speed === 2000 ? 'bg-purple-600' : 'bg-gray-600'} text-white text-sm`}
                    >
                      0.5x
                    </button>
                    <button
                      onClick={() => changeSpeed(1000)}
                      className={`px-3 py-1 rounded ${timeline.speed === 1000 ? 'bg-purple-600' : 'bg-gray-600'} text-white text-sm`}
                    >
                      1x
                    </button>
                    <button
                      onClick={() => changeSpeed(500)}
                      className={`px-3 py-1 rounded ${timeline.speed === 500 ? 'bg-purple-600' : 'bg-gray-600'} text-white text-sm`}
                    >
                      2x
                    </button>
                    <button
                      onClick={() => changeSpeed(200)}
                      className={`px-3 py-1 rounded ${timeline.speed === 200 ? 'bg-purple-600' : 'bg-gray-600'} text-white text-sm`}
                    >
                      5x
                    </button>
                  </div>
                </div>

                {/* Progress Bar */}
                <div className="bg-gray-700 rounded-full h-2 mb-2">
                  <div
                    className="bg-purple-600 h-2 rounded-full transition-all duration-200"
                    style={{ width: `${(timeline.currentIndex / Math.max(timeline.events.length - 1, 1)) * 100}%` }}
                  />
                </div>
                <div className="text-purple-200 text-sm text-center">
                  Event {timeline.currentIndex + 1} of {timeline.events.length}
                </div>
              </div>
            )}
          </div>

          {/* Main Timeline Visualization */}
          {timeline.events.length > 0 && (
            <div className="grid grid-cols-1 xl:grid-cols-3 gap-8">
              {/* Agent Activity Panel */}
              <div className="xl:col-span-1">
                <div className="bg-black/30 backdrop-blur-sm rounded-xl p-6">
                  <h3 className="text-xl font-bold text-white mb-4">🏃‍♀️ Active Agents</h3>
                  <div className="space-y-3">
                    {Array.from(getAgentActivity().entries()).map(([agentId, event]) => (
                      <div
                        key={agentId}
                        className="bg-gray-800/50 rounded-lg p-3 border-l-4 transition-all"
                        style={{ borderLeftColor: getEventTypeColor(event.type) }}
                      >
                        <div className="flex items-center gap-3">
                          <div className="w-8 h-8 bg-purple-600 rounded-full flex items-center justify-center text-white font-bold text-sm">
                            {event.agent_name?.charAt(0).toUpperCase()}
                          </div>
                          <div className="flex-1 min-w-0">
                            <div className="text-white font-medium truncate">
                              {event.agent_name}
                            </div>
                            <div className="text-purple-300 text-sm truncate">
                              {event.description}
                            </div>
                            <div className="text-gray-400 text-xs">
                              {formatTimestamp(event.timestamp)}
                            </div>
                          </div>
                        </div>
                      </div>
                    ))}
                    {getAgentActivity().size === 0 && (
                      <div className="text-gray-400 text-center py-8">
                        No active agents at this time
                      </div>
                    )}
                  </div>
                </div>
              </div>

              {/* Event Stream */}
              <div className="xl:col-span-2">
                <div className="bg-black/30 backdrop-blur-sm rounded-xl p-6">
                  <h3 className="text-xl font-bold text-white mb-4">📜 Event Stream</h3>
                  <div className="space-y-3 max-h-96 overflow-y-auto">
                    {getVisibleEvents().reverse().map((event, index) => {
                      const isCurrentEvent = index === getVisibleEvents().length - 1;
                      return (
                        <div
                          key={`${event.timestamp}-${index}`}
                          className={`p-4 rounded-lg border-l-4 transition-all ${
                            isCurrentEvent 
                              ? 'bg-purple-900/50 border-purple-400 scale-105' 
                              : 'bg-gray-800/30 border-gray-600'
                          }`}
                          style={{ borderLeftColor: getEventTypeColor(event.type) }}
                        >
                          <div className="flex items-start justify-between">
                            <div className="flex-1">
                              <div className="flex items-center gap-2 mb-1">
                                <span className="text-white font-medium">
                                  {event.agent_name || 'System'}
                                </span>
                                <span className="text-xs px-2 py-1 rounded" style={{ 
                                  backgroundColor: getEventTypeColor(event.type) + '20',
                                  color: getEventTypeColor(event.type)
                                }}>
                                  {event.type.replace('_', ' ').toUpperCase()}
                                </span>
                              </div>
                              <div className="text-purple-200 mb-1">
                                {event.description}
                              </div>
                              {event.task_name && (
                                <div className="text-gray-400 text-sm">
                                  Task: {event.task_name}
                                </div>
                              )}
                            </div>
                            <div className="text-gray-400 text-xs ml-4">
                              {formatTimestamp(event.timestamp)}
                            </div>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Summary Stats */}
          {timeline.events.length > 0 && (
            <div className="mt-8 bg-black/30 backdrop-blur-sm rounded-xl p-6">
              <h3 className="text-xl font-bold text-white mb-4">📊 Timeline Summary</h3>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <div className="text-center">
                  <div className="text-2xl font-bold text-purple-400">
                    {timeline.events.filter(e => e.type === 'task_created').length}
                  </div>
                  <div className="text-purple-200 text-sm">Tasks Created</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-green-400">
                    {timeline.events.filter(e => e.type === 'task_completed').length}
                  </div>
                  <div className="text-purple-200 text-sm">Tasks Completed</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-blue-400">
                    {new Set(timeline.events.filter(e => e.agent_id).map(e => e.agent_id)).size}
                  </div>
                  <div className="text-purple-200 text-sm">Active Agents</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-yellow-400">
                    {timeline.events.length}
                  </div>
                  <div className="text-purple-200 text-sm">Total Events</div>
                </div>
              </div>
            </div>
          )}

          {/* Empty State */}
          {timeline.events.length === 0 && !loading && (
            <div className="text-center py-16">
              <div className="text-6xl mb-4">🎬</div>
              <h3 className="text-xl font-bold text-white mb-2">No Activity Found</h3>
              <p className="text-purple-200">
                Try adjusting your time range or agent filter and loading the timeline again.
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default TimeLapseView;