import { useState, useEffect } from 'react';
import './SessionReplay.css';

interface SessionReplay {
  session_id: string;
  agent_id: string;
  status: string;
  started_at: string;
  completed_at?: string;
  initial_task: string;
  total_turns: number;
  total_tokens: number;
  total_cost: number;
  turns: ReplayTurn[];
}

interface ReplayTurn {
  turn_number: number;
  timestamp: string;
  input: string;
  output: string;
  tool_calls: ReplayToolCall[];
  input_tokens: number;
  output_tokens: number;
  cost: number;
  duration: number;
}

interface ReplayToolCall {
  name: string;
  parameters: any;
  result: string;
  success: boolean;
  duration: number;
  timestamp: string;
}

interface SessionReplayProps {
  sessionId: string;
  onClose: () => void;
}

export default function SessionReplay({ sessionId, onClose }: SessionReplayProps) {
  const [replay, setReplay] = useState<SessionReplay | null>(null);
  const [currentTurnIndex, setCurrentTurnIndex] = useState(0);
  const [currentToolIndex, setCurrentToolIndex] = useState(0);
  const [isPlaying, setIsPlaying] = useState(false);
  const [playbackSpeed, setPlaybackSpeed] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchReplayData();
  }, [sessionId]);

  const fetchReplayData = async () => {
    try {
      setLoading(true);
      const response = await fetch(`/api/inber/session-replay?session=${encodeURIComponent(sessionId)}`);
      if (!response.ok) {
        throw new Error(`Failed to fetch replay data: ${response.statusText}`);
      }
      const data = await response.json();
      setReplay(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (!isPlaying || !replay) return;

    const currentTurn = replay.turns[currentTurnIndex];
    if (!currentTurn) {
      setIsPlaying(false);
      return;
    }

    let timeout: number;
    
    if (currentToolIndex < currentTurn.tool_calls.length) {
      // Play tool calls
      const currentTool = currentTurn.tool_calls[currentToolIndex];
      const duration = (currentTool.duration * 1000) / playbackSpeed;
      
      timeout = setTimeout(() => {
        setCurrentToolIndex(prev => prev + 1);
      }, duration);
    } else {
      // Move to next turn
      const duration = Math.max(1000, currentTurn.duration * 1000) / playbackSpeed;
      
      timeout = setTimeout(() => {
        if (currentTurnIndex < replay.turns.length - 1) {
          setCurrentTurnIndex(prev => prev + 1);
          setCurrentToolIndex(0);
        } else {
          setIsPlaying(false);
        }
      }, duration);
    }

    return () => clearTimeout(timeout);
  }, [isPlaying, currentTurnIndex, currentToolIndex, playbackSpeed, replay]);

  const handlePlay = () => {
    setIsPlaying(!isPlaying);
  };

  const handleReset = () => {
    setIsPlaying(false);
    setCurrentTurnIndex(0);
    setCurrentToolIndex(0);
  };

  const handleStepForward = () => {
    if (!replay) return;
    
    const currentTurn = replay.turns[currentTurnIndex];
    if (currentToolIndex < currentTurn.tool_calls.length) {
      setCurrentToolIndex(prev => prev + 1);
    } else if (currentTurnIndex < replay.turns.length - 1) {
      setCurrentTurnIndex(prev => prev + 1);
      setCurrentToolIndex(0);
    }
  };

  const handleStepBackward = () => {
    if (currentToolIndex > 0) {
      setCurrentToolIndex(prev => prev - 1);
    } else if (currentTurnIndex > 0) {
      setCurrentTurnIndex(prev => prev - 1);
      const prevTurn = replay?.turns[currentTurnIndex - 1];
      setCurrentToolIndex(prevTurn ? prevTurn.tool_calls.length : 0);
    }
  };

  const formatDuration = (seconds: number) => {
    if (seconds < 60) return `${seconds.toFixed(1)}s`;
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}m ${secs.toFixed(1)}s`;
  };

  const getToolIcon = (toolName: string) => {
    const icons: { [key: string]: string } = {
      read: '📖',
      write: '✏️',
      edit: '✂️',
      exec: '⚙️',
      process: '🔄',
      web_search: '🔍',
      web_fetch: '🌐',
      image: '🖼️',
      tts: '🔊',
      browser: '🌏',
    };
    return icons[toolName] || '🔧';
  };

  if (loading) {
    return (
      <div className="session-replay-modal">
        <div className="session-replay-content">
          <div className="loading">Loading replay data...</div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="session-replay-modal">
        <div className="session-replay-content">
          <div className="error">Error: {error}</div>
          <button onClick={onClose}>Close</button>
        </div>
      </div>
    );
  }

  if (!replay) {
    return null;
  }

  const currentTurn = replay.turns[currentTurnIndex];
  const progress = (currentTurnIndex + 1) / replay.turns.length;

  return (
    <div className="session-replay-modal" onClick={(e) => e.target === e.currentTarget && onClose()}>
      <div className="session-replay-content">
        <div className="replay-header">
          <h2>🎬 Quest Replay: {replay.agent_id}</h2>
          <button className="close-button" onClick={onClose}>×</button>
        </div>

        <div className="replay-info">
          <div className="quest-title">{replay.initial_task}</div>
          <div className="quest-stats">
            <span>🎯 {replay.total_turns} turns</span>
            <span>🪙 {replay.total_tokens.toLocaleString()} tokens</span>
            <span>💰 ${replay.total_cost.toFixed(4)}</span>
            <span>⏱️ {formatDuration(replay.turns.reduce((sum, turn) => sum + turn.duration, 0))}</span>
          </div>
        </div>

        <div className="replay-controls">
          <button onClick={handleReset} title="Reset">⏮️</button>
          <button onClick={handleStepBackward} title="Step Back">⏪</button>
          <button onClick={handlePlay} title={isPlaying ? 'Pause' : 'Play'}>
            {isPlaying ? '⏸️' : '▶️'}
          </button>
          <button onClick={handleStepForward} title="Step Forward">⏩</button>
          
          <select 
            value={playbackSpeed} 
            onChange={(e) => setPlaybackSpeed(Number(e.target.value))}
            className="speed-control"
          >
            <option value={0.5}>0.5x</option>
            <option value={1}>1x</option>
            <option value={2}>2x</option>
            <option value={4}>4x</option>
            <option value={8}>8x</option>
          </select>
        </div>

        <div className="replay-progress">
          <div className="progress-bar">
            <div 
              className="progress-fill" 
              style={{ width: `${progress * 100}%` }}
            />
          </div>
          <div className="progress-text">
            Turn {currentTurnIndex + 1} of {replay.turns.length}
          </div>
        </div>

        {currentTurn && (
          <div className="current-turn">
            <div className="turn-header">
              <h3>Turn {currentTurn.turn_number}</h3>
              <span className="turn-time">{new Date(currentTurn.timestamp).toLocaleTimeString()}</span>
            </div>

            <div className="turn-content">
              <div className="turn-input">
                <h4>💬 Input:</h4>
                <div className="content-box">{currentTurn.input}</div>
              </div>

              {currentTurn.tool_calls.length > 0 && (
                <div className="tool-calls">
                  <h4>🔧 Tool Calls:</h4>
                  <div className="tools-timeline">
                    {currentTurn.tool_calls.map((tool, index) => (
                      <div 
                        key={index} 
                        className={`tool-call ${index <= currentToolIndex ? 'completed' : 'pending'} ${!tool.success ? 'failed' : ''}`}
                      >
                        <div className="tool-header">
                          <span className="tool-icon">{getToolIcon(tool.name)}</span>
                          <span className="tool-name">{tool.name}</span>
                          <span className="tool-duration">{formatDuration(tool.duration)}</span>
                          {!tool.success && <span className="tool-status">❌</span>}
                        </div>
                        
                        {index <= currentToolIndex && (
                          <div className="tool-details">
                            <div className="tool-params">
                              <strong>Parameters:</strong>
                              <pre>{JSON.stringify(tool.parameters, null, 2)}</pre>
                            </div>
                            <div className="tool-result">
                              <strong>Result:</strong>
                              <div className="result-content">{tool.result}</div>
                            </div>
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              )}

              <div className="turn-output">
                <h4>🤖 Output:</h4>
                <div className="content-box">{currentTurn.output}</div>
              </div>

              <div className="turn-stats">
                <span>📥 {currentTurn.input_tokens} tokens in</span>
                <span>📤 {currentTurn.output_tokens} tokens out</span>
                <span>💰 ${currentTurn.cost.toFixed(4)}</span>
                <span>⏱️ {formatDuration(currentTurn.duration)}</span>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}