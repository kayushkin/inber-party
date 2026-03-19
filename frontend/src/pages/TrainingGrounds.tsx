import { useState, useEffect, useCallback } from 'react';
import { useStore } from '../store';
import './TrainingGrounds.css';

interface Benchmark {
  id: string;
  name: string;
  description: string;
  category: 'code' | 'reasoning' | 'knowledge' | 'creativity' | 'speed';
  score: number;
  maxScore: number;
  unit: string;
  lastRun: string;
  status: 'passing' | 'failing' | 'warning' | 'pending';
  agentScores?: { agentId: string; agentName: string; score: number; timestamp: string }[];
}

interface TestSuite {
  id: string;
  name: string;
  description: string;
  testsTotal: number;
  testsPassing: number;
  testsFailing: number;
  testsSkipped: number;
  lastRun: string;
  duration: number; // milliseconds
  coverage: number; // percentage
  status: 'passing' | 'failing' | 'warning';
  failedTests?: string[];
}

export default function TrainingGrounds() {
  const agents = useStore((s) => s.agents);
  const [benchmarks, setBenchmarks] = useState<Benchmark[]>([]);
  const [testSuites, setTestSuites] = useState<TestSuite[]>([]);
  const [selectedTab, setSelectedTab] = useState<'benchmarks' | 'tests'>('benchmarks');
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [loading, setLoading] = useState(true);

  const fetchBenchmarks = useCallback(async () => {
    try {
      // For now, using mock data. In the future, this would fetch from actual benchmark endpoints
      const mockBenchmarks: Benchmark[] = [
        {
          id: 'code-quality-1',
          name: 'Code Quality Analysis',
          description: 'Measures code readability, maintainability, and best practices',
          category: 'code',
          score: 87,
          maxScore: 100,
          unit: 'score',
          lastRun: new Date(Date.now() - 3600000).toISOString(), // 1 hour ago
          status: 'passing',
          agentScores: agents.slice(0, 5).map(agent => ({
            agentId: agent.id,
            agentName: agent.name,
            score: Math.floor(Math.random() * 40) + 60,
            timestamp: new Date(Date.now() - Math.random() * 86400000).toISOString()
          }))
        },
        {
          id: 'response-time-1',
          name: 'Average Response Time',
          description: 'Time taken to generate first response token',
          category: 'speed',
          score: 1.2,
          maxScore: 2.0,
          unit: 'seconds',
          lastRun: new Date(Date.now() - 1800000).toISOString(), // 30 min ago
          status: 'passing',
          agentScores: agents.slice(0, 5).map(agent => ({
            agentId: agent.id,
            agentName: agent.name,
            score: Math.random() * 1.5 + 0.5,
            timestamp: new Date(Date.now() - Math.random() * 86400000).toISOString()
          }))
        },
        {
          id: 'reasoning-test-1',
          name: 'Logical Reasoning',
          description: 'Complex multi-step reasoning and problem-solving capability',
          category: 'reasoning',
          score: 73,
          maxScore: 100,
          unit: 'score',
          lastRun: new Date(Date.now() - 7200000).toISOString(), // 2 hours ago
          status: 'warning',
          agentScores: agents.slice(0, 3).map(agent => ({
            agentId: agent.id,
            agentName: agent.name,
            score: Math.floor(Math.random() * 50) + 50,
            timestamp: new Date(Date.now() - Math.random() * 86400000).toISOString()
          }))
        },
        {
          id: 'knowledge-retention-1',
          name: 'Knowledge Retention',
          description: 'Ability to remember and apply information across sessions',
          category: 'knowledge',
          score: 92,
          maxScore: 100,
          unit: 'score',
          lastRun: new Date(Date.now() - 14400000).toISOString(), // 4 hours ago
          status: 'passing'
        },
        {
          id: 'creative-writing-1',
          name: 'Creative Writing',
          description: 'Original story generation and creative problem solving',
          category: 'creativity',
          score: 65,
          maxScore: 100,
          unit: 'score',
          lastRun: new Date(Date.now() - 21600000).toISOString(), // 6 hours ago
          status: 'failing'
        }
      ];

      setBenchmarks(mockBenchmarks);
    } catch (error) {
      console.error('Failed to fetch benchmarks:', error);
    } finally {
      setLoading(false);
    }
  }, [agents]);

  const fetchTestSuites = useCallback(async () => {
    try {
      // Mock test data - in the future this would come from actual CI/test runners
      const mockTestSuites: TestSuite[] = [
        {
          id: 'frontend-tests',
          name: 'Frontend Unit Tests',
          description: 'React component and utility function tests',
          testsTotal: 147,
          testsPassing: 142,
          testsFailing: 3,
          testsSkipped: 2,
          lastRun: new Date(Date.now() - 1800000).toISOString(),
          duration: 45000, // 45 seconds
          coverage: 87.5,
          status: 'failing',
          failedTests: [
            'AgentCard.test.tsx: should handle missing avatar gracefully',
            'QuestBoard.test.tsx: should sort quests by difficulty',
            'StatsView.test.tsx: should format large numbers correctly'
          ]
        },
        {
          id: 'backend-tests',
          name: 'Backend API Tests',
          description: 'API endpoint and integration tests',
          testsTotal: 89,
          testsPassing: 89,
          testsFailing: 0,
          testsSkipped: 0,
          lastRun: new Date(Date.now() - 3600000).toISOString(),
          duration: 32000, // 32 seconds
          coverage: 94.2,
          status: 'passing'
        },
        {
          id: 'e2e-tests',
          name: 'End-to-End Tests',
          description: 'Full user workflow automation tests',
          testsTotal: 24,
          testsPassing: 21,
          testsFailing: 1,
          testsSkipped: 2,
          lastRun: new Date(Date.now() - 7200000).toISOString(),
          duration: 180000, // 3 minutes
          coverage: 65.3,
          status: 'warning',
          failedTests: [
            'quest-creation-flow.spec.ts: Quest creation should update status in real-time'
          ]
        },
        {
          id: 'performance-tests',
          name: 'Performance Tests',
          description: 'Load testing and performance regression tests',
          testsTotal: 12,
          testsPassing: 10,
          testsFailing: 2,
          testsSkipped: 0,
          lastRun: new Date(Date.now() - 10800000).toISOString(),
          duration: 300000, // 5 minutes
          coverage: 45.0,
          status: 'failing',
          failedTests: [
            'load-test.spec.ts: Should handle 100 concurrent users',
            'memory-leak.spec.ts: Memory usage should stabilize after 1 hour'
          ]
        }
      ];

      setTestSuites(mockTestSuites);
    } catch (error) {
      console.error('Failed to fetch test suites:', error);
    }
  }, []);

  useEffect(() => {
    fetchBenchmarks();
    fetchTestSuites();
  }, [fetchBenchmarks, fetchTestSuites]);

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'passing': return '✅';
      case 'failing': return '❌';
      case 'warning': return '⚠️';
      case 'pending': return '⏳';
      default: return '❓';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'passing': return '#4ade80';
      case 'failing': return '#f87171';
      case 'warning': return '#fbbf24';
      case 'pending': return '#94a3b8';
      default: return '#d4af37';
    }
  };

  const getCategoryIcon = (category: string) => {
    switch (category) {
      case 'code': return '💻';
      case 'reasoning': return '🧠';
      case 'knowledge': return '📚';
      case 'creativity': return '🎨';
      case 'speed': return '⚡';
      default: return '📊';
    }
  };

  const formatDuration = (ms: number) => {
    const seconds = ms / 1000;
    if (seconds < 60) return `${seconds.toFixed(1)}s`;
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = Math.floor(seconds % 60);
    return `${minutes}m ${remainingSeconds}s`;
  };

  const timeAgo = (timestamp: string) => {
    const now = Date.now();
    const then = new Date(timestamp).getTime();
    const diff = now - then;
    
    if (diff < 60000) return 'just now';
    if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
    if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
    return `${Math.floor(diff / 86400000)}d ago`;
  };

  const filteredBenchmarks = selectedCategory === 'all' 
    ? benchmarks 
    : benchmarks.filter(b => b.category === selectedCategory);

  if (loading) {
    return (
      <div className="training-grounds">
        <div className="training-header">
          <h1>🏋️ Training Grounds</h1>
          <p>Agent benchmarks and test results</p>
        </div>
        <div className="loading-state">
          <div className="loading-spinner"></div>
          <p>Preparing the training equipment...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="training-grounds">
      <div className="training-header">
        <h1>🏋️ Training Grounds</h1>
        <p>Agent benchmarks and test results</p>
      </div>

      <div className="training-tabs">
        <button
          className={`tab-button ${selectedTab === 'benchmarks' ? 'active' : ''}`}
          onClick={() => setSelectedTab('benchmarks')}
        >
          📊 Benchmarks
        </button>
        <button
          className={`tab-button ${selectedTab === 'tests' ? 'active' : ''}`}
          onClick={() => setSelectedTab('tests')}
        >
          🧪 Test Suites
        </button>
      </div>

      {selectedTab === 'benchmarks' && (
        <div className="benchmarks-section">
          <div className="section-header">
            <h2>Performance Benchmarks</h2>
            <div className="category-filters">
              <button
                className={`filter-btn ${selectedCategory === 'all' ? 'active' : ''}`}
                onClick={() => setSelectedCategory('all')}
              >
                All
              </button>
              {['code', 'reasoning', 'knowledge', 'creativity', 'speed'].map(category => (
                <button
                  key={category}
                  className={`filter-btn ${selectedCategory === category ? 'active' : ''}`}
                  onClick={() => setSelectedCategory(category)}
                >
                  {getCategoryIcon(category)} {category}
                </button>
              ))}
            </div>
          </div>

          <div className="benchmarks-grid">
            {filteredBenchmarks.map(benchmark => (
              <div key={benchmark.id} className="benchmark-card">
                <div className="benchmark-header">
                  <div className="benchmark-title">
                    <span className="benchmark-icon">{getCategoryIcon(benchmark.category)}</span>
                    <h3>{benchmark.name}</h3>
                  </div>
                  <span 
                    className="benchmark-status"
                    style={{ color: getStatusColor(benchmark.status) }}
                  >
                    {getStatusIcon(benchmark.status)}
                  </span>
                </div>

                <p className="benchmark-description">{benchmark.description}</p>

                <div className="benchmark-score">
                  <div className="score-display">
                    <span className="score-value">
                      {benchmark.unit === 'seconds' ? benchmark.score.toFixed(2) : benchmark.score}
                    </span>
                    <span className="score-unit">{benchmark.unit}</span>
                  </div>
                  
                  {benchmark.maxScore && benchmark.unit !== 'seconds' && (
                    <div className="score-bar">
                      <div 
                        className="score-fill"
                        style={{ 
                          width: `${(benchmark.score / benchmark.maxScore) * 100}%`,
                          backgroundColor: getStatusColor(benchmark.status)
                        }}
                      />
                    </div>
                  )}
                </div>

                <div className="benchmark-footer">
                  <span className="last-run">Last run: {timeAgo(benchmark.lastRun)}</span>
                  {benchmark.agentScores && (
                    <span className="agent-count">
                      {benchmark.agentScores.length} agents tested
                    </span>
                  )}
                </div>

                {benchmark.agentScores && (
                  <div className="agent-scores">
                    <h4>Agent Performance</h4>
                    <div className="agent-scores-list">
                      {benchmark.agentScores.slice(0, 3).map(agentScore => (
                        <div key={agentScore.agentId} className="agent-score">
                          <span className="agent-name">{agentScore.agentName}</span>
                          <span className="agent-score-value">
                            {benchmark.unit === 'seconds' 
                              ? agentScore.score.toFixed(2) 
                              : Math.round(agentScore.score)}
                            {benchmark.unit}
                          </span>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      {selectedTab === 'tests' && (
        <div className="tests-section">
          <div className="section-header">
            <h2>Test Suites</h2>
          </div>

          <div className="test-suites-grid">
            {testSuites.map(suite => (
              <div key={suite.id} className="test-suite-card">
                <div className="suite-header">
                  <div className="suite-title">
                    <h3>{suite.name}</h3>
                    <span 
                      className="suite-status"
                      style={{ color: getStatusColor(suite.status) }}
                    >
                      {getStatusIcon(suite.status)}
                    </span>
                  </div>
                </div>

                <p className="suite-description">{suite.description}</p>

                <div className="suite-stats">
                  <div className="stat-row">
                    <span className="stat-label">Tests:</span>
                    <span className="stat-value">
                      <span className="passing">{suite.testsPassing} ✅</span>
                      <span className="failing">{suite.testsFailing} ❌</span>
                      <span className="skipped">{suite.testsSkipped} ⏭️</span>
                    </span>
                  </div>
                  
                  <div className="stat-row">
                    <span className="stat-label">Duration:</span>
                    <span className="stat-value">{formatDuration(suite.duration)}</span>
                  </div>
                  
                  <div className="stat-row">
                    <span className="stat-label">Coverage:</span>
                    <span className="stat-value">{suite.coverage}%</span>
                  </div>
                </div>

                <div className="coverage-bar">
                  <div 
                    className="coverage-fill"
                    style={{ width: `${suite.coverage}%` }}
                  />
                </div>

                {suite.failedTests && suite.failedTests.length > 0 && (
                  <div className="failed-tests">
                    <h4>Failed Tests</h4>
                    <ul className="failed-tests-list">
                      {suite.failedTests.map((test, index) => (
                        <li key={index} className="failed-test">
                          ❌ {test}
                        </li>
                      ))}
                    </ul>
                  </div>
                )}

                <div className="suite-footer">
                  <span className="last-run">Last run: {timeAgo(suite.lastRun)}</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}