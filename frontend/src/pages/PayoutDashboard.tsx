import { useEffect, useState } from 'react';

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

const PayoutDashboard = () => {
  const [payouts, setPayouts] = useState<PayoutEntry[]>([]);
  const [summaries, setSummaries] = useState<PayoutSummary[]>([]);
  const [selectedAgent, setSelectedAgent] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);
  const [tab, setTab] = useState<'summary' | 'history'>('summary');
  const [source, setSource] = useState<string>('');

  useEffect(() => {
    fetchData();
  }, [selectedAgent, source]);

  const fetchData = async () => {
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
      </div>

      {/* Summary Tab */}
      {tab === 'summary' && (
        <div className="grid gap-6">
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
                  {summaries.map((summary) => (
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
    </div>
  );
};

export default PayoutDashboard;