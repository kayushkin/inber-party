import { useEffect, useState, useCallback } from 'react';
import { useStore } from '../store';

interface PayoutEntry {
  id: number;
  agent_id: number;
  amount: number;
  source: string;
  source_id?: number;
  description: string;
  transaction_type: 'credit' | 'debit' | 'adjustment';
  balance_before: number;
  balance_after: number;
  processed_by?: number;
  created_at: string;
}

interface PayoutSummary {
  agent_id: number;
  agent_name: string;
  total_earned: number;
  total_spent: number;
  current_balance: number;
  payout_count: number;
  last_payout?: string;
  sources: Record<string, number>;
}

interface ManualPayoutForm {
  agent_id: number | '';
  amount: number | '';
  description: string;
  type: 'credit' | 'debit';
}

const PayoutDashboard = () => {
  const [payouts, setPayouts] = useState<PayoutEntry[]>([]);
  const [summaries, setSummaries] = useState<PayoutSummary[]>([]);
  const [selectedAgent, setSelectedAgent] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);
  const [tab, setTab] = useState<'summary' | 'history' | 'manual'>('summary');
  const [source, setSource] = useState<string>('');
  const [manualForm, setManualForm] = useState<ManualPayoutForm>({
    agent_id: '',
    amount: '',
    description: '',
    type: 'credit'
  });
  const [submittingManual, setSubmittingManual] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');

  // Get agents from store for dropdown
  const agents = useStore((s) => s.agents);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      // Fetch summaries
      const summaryRes = await fetch('/api/payouts/summary');
      if (summaryRes.ok) {
        const summaryData = await summaryRes.json();
        setSummaries(summaryData);
      }

      // Fetch recent payouts
      const params = new URLSearchParams();
      if (selectedAgent) params.set('agent_id', selectedAgent.toString());
      if (source) params.set('source', source);
      params.set('limit', '100');
      
      const payoutRes = await fetch(`/api/payouts?${params}`);
      if (payoutRes.ok) {
        const payoutData = await payoutRes.json();
        setPayouts(payoutData);
      }
    } catch (error) {
      console.error('Failed to fetch payout data:', error);
    } finally {
      setLoading(false);
    }
  }, [selectedAgent, source]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const submitManualPayout = async () => {
    if (!manualForm.agent_id || !manualForm.amount || !manualForm.description) {
      alert('Please fill in all required fields');
      return;
    }

    setSubmittingManual(true);
    try {
      const amount = manualForm.type === 'debit' ? -Math.abs(Number(manualForm.amount)) : Math.abs(Number(manualForm.amount));
      
      const response = await fetch('/api/payouts', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          agent_id: Number(manualForm.agent_id),
          amount: amount,
          description: manualForm.description,
        }),
      });

      if (response.ok) {
        alert('Payout recorded successfully!');
        setManualForm({ agent_id: '', amount: '', description: '', type: 'credit' });
        await fetchData(); // Refresh data
      } else {
        const error = await response.text();
        alert(`Failed to record payout: ${error}`);
      }
    } catch (error) {
      console.error('Failed to submit manual payout:', error);
      alert('Failed to submit payout. Please try again.');
    } finally {
      setSubmittingManual(false);
    }
  };

  const exportPayoutData = () => {
    const csvContent = [
      ['Agent Name', 'Total Earned', 'Current Balance', 'Payouts', 'Last Payout'],
      ...summaries.map(summary => [
        summary.agent_name,
        summary.total_earned,
        summary.current_balance,
        summary.payout_count,
        summary.last_payout ? formatDate(summary.last_payout) : 'Never'
      ])
    ].map(row => row.join(',')).join('\n');

    const blob = new Blob([csvContent], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `payout-summary-${new Date().toISOString().split('T')[0]}.csv`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const getSourceIcon = (source: string) => {
    switch (source) {
      case 'bounty': return '💰';
      case 'quest': return '⚔️';
      case 'daily_quest': return '📅';
      case 'manual': return '👤';
      default: return '💎';
    }
  };

  const getTransactionIcon = (type: string, amount: number) => {
    if (type === 'credit' || (type === 'adjustment' && amount > 0)) return '📈';
    if (type === 'debit' || (type === 'adjustment' && amount < 0)) return '📉';
    return '⚖️';
  };

  if (loading) {
    return (
      <div className="loading" style={{ color: '#d4af37', fontSize: '1.2rem', padding: '4rem', textAlign: 'center' }}>
        ⏳ Loading payout data...
      </div>
    );
  }

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <h1 className="text-3xl font-bold text-yellow-300 mb-6">💳 Payout Ledger</h1>
      
      {/* Tab Navigation */}
      <div className="flex space-x-4 mb-6">
        <button
          onClick={() => setTab('summary')}
          className={`px-4 py-2 rounded-lg font-medium transition-colors ${
            tab === 'summary' 
              ? 'bg-yellow-600 text-white' 
              : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
          }`}
        >
          📊 Summary
        </button>
        <button
          onClick={() => setTab('history')}
          className={`px-4 py-2 rounded-lg font-medium transition-colors ${
            tab === 'history' 
              ? 'bg-yellow-600 text-white' 
              : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
          }`}
        >
          📋 Transaction History
        </button>
        <button
          onClick={() => setTab('manual')}
          className={`px-4 py-2 rounded-lg font-medium transition-colors ${
            tab === 'manual' 
              ? 'bg-yellow-600 text-white' 
              : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
          }`}
        >
          ✏️ Manual Entry
        </button>
        <button
          onClick={exportPayoutData}
          className="px-4 py-2 rounded-lg font-medium bg-blue-600 text-white hover:bg-blue-700 transition-colors ml-auto"
        >
          📄 Export CSV
        </button>
      </div>

      {/* Summary Tab */}
      {tab === 'summary' && (
        <div className="grid gap-6">
          {/* Search and totals bar */}
          <div className="bg-gray-800 rounded-lg p-4">
            <div className="flex flex-wrap items-center justify-between gap-4">
              <input
                type="text"
                placeholder="🔍 Search agents..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="px-3 py-2 bg-gray-700 border border-gray-600 rounded text-white placeholder-gray-400"
              />
              <div className="flex space-x-6 text-sm text-gray-300">
                <div>
                  Total Participants: <span className="text-yellow-400 font-medium">{summaries.length}</span>
                </div>
                <div>
                  Total Earned: <span className="text-green-400 font-medium">{summaries.reduce((sum, s) => sum + s.total_earned, 0).toLocaleString()}g</span>
                </div>
                <div>
                  Total Active Balance: <span className="text-blue-400 font-medium">{summaries.reduce((sum, s) => sum + s.current_balance, 0).toLocaleString()}g</span>
                </div>
              </div>
            </div>
          </div>
          
          <div className="bg-gray-800 rounded-lg p-6">
            <h2 className="text-xl font-semibold text-yellow-300 mb-4">Participant Earnings</h2>
            <div className="overflow-x-auto">
              <table className="w-full text-left">
                <thead>
                  <tr className="border-b border-gray-700">
                    <th className="pb-2 text-gray-300">Agent</th>
                    <th className="pb-2 text-gray-300 text-right">Total Earned</th>
                    <th className="pb-2 text-gray-300 text-right">Current Balance</th>
                    <th className="pb-2 text-gray-300 text-right">Payouts</th>
                    <th className="pb-2 text-gray-300">Last Payout</th>
                    <th className="pb-2 text-gray-300">Top Sources</th>
                  </tr>
                </thead>
                <tbody>
                  {summaries
                    .filter(summary => 
                      searchTerm === '' || 
                      summary.agent_name.toLowerCase().includes(searchTerm.toLowerCase())
                    )
                    .map((summary) => (
                    <tr key={summary.agent_id} className="border-b border-gray-700/50 hover:bg-gray-700/20">
                      <td className="py-3">
                        <div className="font-medium text-white">{summary.agent_name}</div>
                        <div className="text-sm text-gray-400">ID: {summary.agent_id}</div>
                      </td>
                      <td className="py-3 text-right">
                        <span className="text-yellow-400 font-medium">
                          {summary.total_earned.toLocaleString()}g
                        </span>
                      </td>
                      <td className="py-3 text-right">
                        <span className={`font-medium ${
                          summary.current_balance > 0 ? 'text-green-400' : 
                          summary.current_balance < 0 ? 'text-red-400' : 'text-gray-400'
                        }`}>
                          {summary.current_balance.toLocaleString()}g
                        </span>
                      </td>
                      <td className="py-3 text-right text-gray-300">
                        {summary.payout_count}
                      </td>
                      <td className="py-3 text-gray-300 text-sm">
                        {summary.last_payout ? formatDate(summary.last_payout) : 'Never'}
                      </td>
                      <td className="py-3">
                        <div className="flex space-x-2">
                          {Object.entries(summary.sources || {}).slice(0, 3).map(([source, amount]) => (
                            <span key={source} className="px-2 py-1 bg-gray-700 rounded text-xs text-gray-300">
                              {getSourceIcon(source)} {amount}g
                            </span>
                          ))}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              {summaries.length === 0 && (
                <div className="text-center py-8 text-gray-400">
                  No payout data available yet.
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* History Tab */}
      {tab === 'history' && (
        <div className="grid gap-6">
          {/* Filters */}
          <div className="bg-gray-800 rounded-lg p-4">
            <div className="flex flex-wrap gap-4">
              <select
                value={selectedAgent || ''}
                onChange={(e) => setSelectedAgent(e.target.value ? parseInt(e.target.value) : null)}
                className="px-3 py-2 bg-gray-700 border border-gray-600 rounded text-white"
              >
                <option value="">All Agents</option>
                {summaries.map((summary) => (
                  <option key={summary.agent_id} value={summary.agent_id}>
                    {summary.agent_name}
                  </option>
                ))}
              </select>
              
              <select
                value={source}
                onChange={(e) => setSource(e.target.value)}
                className="px-3 py-2 bg-gray-700 border border-gray-600 rounded text-white"
              >
                <option value="">All Sources</option>
                <option value="bounty">Bounties</option>
                <option value="quest">Quests</option>
                <option value="daily_quest">Daily Quests</option>
                <option value="manual">Manual Adjustments</option>
              </select>
              
              <button
                onClick={fetchData}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded text-white"
              >
                Refresh
              </button>
            </div>
          </div>

          {/* Transaction History */}
          <div className="bg-gray-800 rounded-lg p-6">
            <h2 className="text-xl font-semibold text-yellow-300 mb-4">Recent Transactions</h2>
            <div className="space-y-4">
              {payouts.map((payout) => (
                <div key={payout.id} className="bg-gray-700/50 rounded-lg p-4">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-3">
                      <span className="text-2xl">
                        {getTransactionIcon(payout.transaction_type, payout.amount)}
                      </span>
                      <div>
                        <div className="font-medium text-white">
                          {payout.description}
                        </div>
                        <div className="text-sm text-gray-400">
                          {getSourceIcon(payout.source)} {payout.source}
                          {payout.source_id && ` #${payout.source_id}`}
                        </div>
                      </div>
                    </div>
                    
                    <div className="text-right">
                      <div className={`text-lg font-bold ${
                        payout.transaction_type === 'credit' || 
                        (payout.transaction_type === 'adjustment' && payout.amount > 0)
                          ? 'text-green-400' 
                          : 'text-red-400'
                      }`}>
                        {payout.transaction_type === 'credit' || 
                         (payout.transaction_type === 'adjustment' && payout.amount > 0) 
                          ? '+' : ''}
                        {payout.amount.toLocaleString()}g
                      </div>
                      <div className="text-sm text-gray-400">
                        Balance: {payout.balance_after.toLocaleString()}g
                      </div>
                      <div className="text-xs text-gray-500">
                        {formatDate(payout.created_at)}
                      </div>
                    </div>
                  </div>
                </div>
              ))}
              
              {payouts.length === 0 && (
                <div className="text-center py-8 text-gray-400">
                  No transactions found.
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Manual Entry Tab */}
      {tab === 'manual' && (
        <div className="grid gap-6">
          <div className="bg-gray-800 rounded-lg p-6">
            <h2 className="text-xl font-semibold text-yellow-300 mb-4">🏦 Manual Payout Entry</h2>
            <p className="text-gray-300 mb-6">Create manual gold adjustments for agents. Use this for corrections, bonuses, or penalties.</p>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-2">
                  Agent *
                </label>
                <select
                  value={manualForm.agent_id}
                  onChange={(e) => setManualForm({ ...manualForm, agent_id: e.target.value ? parseInt(e.target.value) : '' })}
                  className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-white"
                  required
                >
                  <option value="">Select an agent...</option>
                  {agents.map((agent) => (
                    <option key={agent.id} value={agent.id}>
                      {agent.name} ({agent.gold?.toLocaleString() || 0}g)
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-300 mb-2">
                  Transaction Type *
                </label>
                <select
                  value={manualForm.type}
                  onChange={(e) => setManualForm({ ...manualForm, type: e.target.value as 'credit' | 'debit' })}
                  className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-white"
                >
                  <option value="credit">📈 Credit (Add Gold)</option>
                  <option value="debit">📉 Debit (Remove Gold)</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-300 mb-2">
                  Amount (Gold) *
                </label>
                <input
                  type="number"
                  min="1"
                  value={manualForm.amount}
                  onChange={(e) => setManualForm({ ...manualForm, amount: e.target.value ? parseInt(e.target.value) : '' })}
                  className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-white"
                  placeholder="Enter amount..."
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-300 mb-2">
                  Description *
                </label>
                <input
                  type="text"
                  value={manualForm.description}
                  onChange={(e) => setManualForm({ ...manualForm, description: e.target.value })}
                  className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-white"
                  placeholder="Reason for adjustment..."
                  required
                />
              </div>
            </div>

            <div className="mt-6 flex space-x-4">
              <button
                onClick={submitManualPayout}
                disabled={submittingManual || !manualForm.agent_id || !manualForm.amount || !manualForm.description}
                className="px-6 py-3 bg-yellow-600 hover:bg-yellow-700 disabled:bg-gray-600 disabled:cursor-not-allowed rounded text-white font-medium transition-colors"
              >
                {submittingManual ? '⏳ Processing...' : '💳 Record Payout'}
              </button>
              <button
                onClick={() => setManualForm({ agent_id: '', amount: '', description: '', type: 'credit' })}
                className="px-6 py-3 bg-gray-600 hover:bg-gray-700 rounded text-white font-medium transition-colors"
              >
                🔄 Clear Form
              </button>
            </div>
          </div>

          {/* Recent manual entries preview */}
          <div className="bg-gray-800 rounded-lg p-6">
            <h3 className="text-lg font-semibold text-yellow-300 mb-4">Recent Manual Entries</h3>
            <div className="space-y-3">
              {payouts
                .filter(payout => payout.source === 'manual')
                .slice(0, 5)
                .map((payout) => (
                  <div key={payout.id} className="bg-gray-700/50 rounded p-3 flex justify-between items-center">
                    <div>
                      <div className="text-white font-medium">{payout.description}</div>
                      <div className="text-sm text-gray-400">{formatDate(payout.created_at)}</div>
                    </div>
                    <div className={`text-lg font-bold ${
                      payout.amount > 0 ? 'text-green-400' : 'text-red-400'
                    }`}>
                      {payout.amount > 0 ? '+' : ''}{payout.amount.toLocaleString()}g
                    </div>
                  </div>
                ))}
              {payouts.filter(payout => payout.source === 'manual').length === 0 && (
                <div className="text-center py-4 text-gray-400">
                  No manual entries yet.
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default PayoutDashboard;