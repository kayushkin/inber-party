import { useState, useEffect } from 'react';
import { formatCost, formatTokens, timeAgo } from '../store';
import './CostDashboard.css';

interface CostEntry {
  id: number;
  agent_id?: number;
  task_id?: number;
  session_id: string;
  tokens_used: number;
  cost_usd: number;
  model_name: string;
  operation_type: string;
  date: string;
  created_at: string;
  metadata?: string;
}

interface CostSummary {
  period: string;
  agent_id?: number;
  agent_name?: string;
  total_cost: number;
  total_tokens: number;
  task_count: number;
  avg_cost_per_task: number;
  date_range: string;
}

interface Agent {
  id: number;
  name: string;
  title: string;
  class: string;
  level: number;
  avatar_emoji: string;
}

export default function CostDashboard() {
  const [costSummaries, setCostSummaries] = useState<CostSummary[]>([]);
  const [costEntries, setCostEntries] = useState<CostEntry[]>([]);
  const [agents, setAgents] = useState<Agent[]>([]);
  const [loading, setLoading] = useState(true);
  const [period, setPeriod] = useState<'daily' | 'weekly' | 'monthly'>('daily');
  const [selectedAgent, setSelectedAgent] = useState<string>('');
  const [days, setDays] = useState<string>('7');

  const API_URL = import.meta.env.VITE_API_URL || '';

  useEffect(() => {
    fetchData();
  }, [period, selectedAgent, days]);

  const fetchData = async () => {
    setLoading(true);
    try {
      // Fetch agents for the dropdown
      const agentsRes = await fetch(`${API_URL}/api/agents`);
      if (agentsRes.ok) {
        setAgents(await agentsRes.json());
      }

      // Build query params for cost summary
      const params = new URLSearchParams({
        period,
        days,
      });
      if (selectedAgent) {
        params.append('agent_id', selectedAgent);
      }

      // Fetch cost summaries
      const summaryRes = await fetch(`${API_URL}/api/costs/summary?${params}`);
      if (summaryRes.ok) {
        setCostSummaries(await summaryRes.json());
      }

      // Fetch recent cost entries for the table
      const entryParams = new URLSearchParams({
        limit: '20',
      });
      if (selectedAgent) {
        entryParams.append('agent_id', selectedAgent);
      }
      const entriesRes = await fetch(`${API_URL}/api/costs?${entryParams}`);
      if (entriesRes.ok) {
        setCostEntries(await entriesRes.json());
      }
    } catch (error) {
      console.error('Failed to fetch cost data:', error);
    } finally {
      setLoading(false);
    }
  };

  const totalCost = costSummaries.reduce((sum, s) => sum + s.total_cost, 0);
  const totalTokens = costSummaries.reduce((sum, s) => sum + s.total_tokens, 0);
  const totalTasks = costSummaries.reduce((sum, s) => sum + s.task_count, 0);
  const avgCostPerTask = totalTasks > 0 ? totalCost / totalTasks : 0;

  if (loading) {
    return (
      <div className="cost-dashboard">
        <h1>💰 Cost Tracking Dashboard</h1>
        <div className="loading">Loading cost data...</div>
      </div>
    );
  }

  return (
    <div className="cost-dashboard">
      <h1>💰 Cost Tracking Dashboard</h1>
      
      {/* Controls */}
      <div className="cost-controls">
        <div className="control-group">
          <label>Period:</label>
          <select value={period} onChange={(e) => setPeriod(e.target.value as 'daily' | 'weekly' | 'monthly')}>
            <option value="daily">Daily</option>
            <option value="weekly">Weekly</option>
            <option value="monthly">Monthly</option>
          </select>
        </div>
        
        <div className="control-group">
          <label>Agent:</label>
          <select value={selectedAgent} onChange={(e) => setSelectedAgent(e.target.value)}>
            <option value="">All Agents</option>
            {agents.map(agent => (
              <option key={agent.id} value={agent.id.toString()}>
                {agent.avatar_emoji} {agent.name}
              </option>
            ))}
          </select>
        </div>
        
        <div className="control-group">
          <label>Last {period === 'daily' ? 'days' : period === 'weekly' ? 'weeks' : 'months'}:</label>
          <select value={days} onChange={(e) => setDays(e.target.value)}>
            <option value="7">7</option>
            <option value="14">14</option>
            <option value="30">30</option>
          </select>
        </div>
      </div>

      {/* Overview Stats */}
      <div className="cost-overview">
        <div className="cost-card">
          <div className="cost-value">{formatCost(totalCost)}</div>
          <div className="cost-label">Total Cost</div>
        </div>
        <div className="cost-card">
          <div className="cost-value">{formatTokens(totalTokens)}</div>
          <div className="cost-label">Total Tokens</div>
        </div>
        <div className="cost-card">
          <div className="cost-value">{totalTasks}</div>
          <div className="cost-label">Total Tasks</div>
        </div>
        <div className="cost-card">
          <div className="cost-value">{formatCost(avgCostPerTask)}</div>
          <div className="cost-label">Avg Cost/Task</div>
        </div>
      </div>

      {/* Summary Chart/Table */}
      <div className="cost-summary">
        <h2>Cost Summary by {period.charAt(0).toUpperCase() + period.slice(1)}</h2>
        <div className="summary-table">
          <table>
            <thead>
              <tr>
                <th>Date</th>
                {!selectedAgent && <th>Agent</th>}
                <th>Cost</th>
                <th>Tokens</th>
                <th>Tasks</th>
                <th>Avg/Task</th>
              </tr>
            </thead>
            <tbody>
              {costSummaries.map((summary, i) => (
                <tr key={i}>
                  <td>{summary.date_range}</td>
                  {!selectedAgent && (
                    <td>
                      {summary.agent_name || 'Unknown'}
                    </td>
                  )}
                  <td className="cost-cell">{formatCost(summary.total_cost)}</td>
                  <td>{formatTokens(summary.total_tokens)}</td>
                  <td>{summary.task_count}</td>
                  <td className="cost-cell">{formatCost(summary.avg_cost_per_task)}</td>
                </tr>
              ))}
              {costSummaries.length === 0 && (
                <tr>
                  <td colSpan={selectedAgent ? 5 : 6} className="no-data">
                    No cost data found for the selected period
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Recent Entries */}
      <div className="cost-entries">
        <h2>Recent Cost Entries</h2>
        <div className="entries-table">
          <table>
            <thead>
              <tr>
                <th>Date</th>
                <th>Agent</th>
                <th>Operation</th>
                <th>Model</th>
                <th>Tokens</th>
                <th>Cost</th>
              </tr>
            </thead>
            <tbody>
              {costEntries.map((entry) => (
                <tr key={entry.id}>
                  <td>{timeAgo(entry.created_at)}</td>
                  <td>
                    {entry.agent_id ? 
                      agents.find(a => a.id === entry.agent_id)?.name || `Agent ${entry.agent_id}` 
                      : '—'
                    }
                  </td>
                  <td className="operation-type">{entry.operation_type}</td>
                  <td className="model-name">{entry.model_name}</td>
                  <td>{formatTokens(entry.tokens_used)}</td>
                  <td className="cost-cell">{formatCost(entry.cost_usd)}</td>
                </tr>
              ))}
              {costEntries.length === 0 && (
                <tr>
                  <td colSpan={6} className="no-data">
                    No recent cost entries found
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}