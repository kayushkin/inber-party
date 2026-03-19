import { useState, useEffect } from 'react';
import { SkeletonQuestCard } from '../components/SkeletonLoader';
import CreateBountyForm from '../components/CreateBountyForm';
import './BountyBoard.css';

interface Bounty {
  id: number;
  title: string;
  description: string;
  requirements: string;
  payout_amount: number;
  status: 'open' | 'claimed' | 'completed' | 'rejected' | 'paid';
  deadline?: string;
  creator_id: number;
  claimer_id?: number;
  required_skills: string[];
  tier: 'bronze' | 'silver' | 'gold' | 'legendary';
  claimed_at?: string;
  submitted_at?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
}

export default function BountyBoard() {
  const [bounties, setBounties] = useState<Bounty[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [filter, setFilter] = useState<string>('all');
  const [showCreateForm, setShowCreateForm] = useState(false);
  
  useEffect(() => {
    fetchBounties();
  }, []);

  const fetchBounties = async () => {
    try {
      setIsLoading(true);
      const response = await fetch('/api/bounties');
      if (response.ok) {
        const data = await response.json();
        setBounties(data);
      } else {
        console.error('Failed to fetch bounties:', response.status);
      }
    } catch (error) {
      console.error('Error fetching bounties:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleCreateBounty = async (bountyData: any) => {
    try {
      const response = await fetch('/api/bounties', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          ...bountyData,
          // Convert deadline to ISO format if provided
          deadline: bountyData.deadline ? new Date(bountyData.deadline).toISOString() : null,
        }),
      });

      if (!response.ok) {
        const errorData = await response.text();
        throw new Error(`Failed to create bounty: ${errorData}`);
      }

      const newBounty = await response.json();
      setBounties(prev => [newBounty, ...prev]);
    } catch (error) {
      throw error; // Re-throw to let the form handle the error
    }
  };

  const filteredBounties = filter === 'all' 
    ? bounties 
    : bounties.filter(bounty => bounty.status === filter);

  const counts = {
    all: bounties.length,
    open: bounties.filter(b => b.status === 'open').length,
    claimed: bounties.filter(b => b.status === 'claimed').length,
    completed: bounties.filter(b => b.status === 'completed').length,
  };

  const getTierIcon = (tier: string) => {
    switch (tier) {
      case 'bronze': return '🥉';
      case 'silver': return '🥈';
      case 'gold': return '🥇';
      case 'legendary': return '💎';
      default: return '⭐';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'open': return '🔓';
      case 'claimed': return '🔒';
      case 'completed': return '✅';
      case 'rejected': return '❌';
      case 'paid': return '💰';
      default: return '❓';
    }
  };

  const formatTimeRemaining = (deadline?: string) => {
    if (!deadline) return null;
    
    const now = new Date();
    const deadlineDate = new Date(deadline);
    const diffMs = deadlineDate.getTime() - now.getTime();
    
    if (diffMs <= 0) return 'Expired';
    
    const days = Math.floor(diffMs / (1000 * 60 * 60 * 24));
    const hours = Math.floor((diffMs % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
    
    if (days > 0) return `${days}d ${hours}h`;
    return `${hours}h`;
  };

  return (
    <div className="bounty-board">
      <div className="bounty-board-header">
        <div className="bounty-board-title">
          <h1>💰 Bounty Board</h1>
          <p className="bounty-board-subtitle">Task marketplace for adventurers</p>
        </div>
        
        <div className="bounty-board-actions">
          <button 
            className="create-bounty-btn"
            onClick={() => setShowCreateForm(true)}
          >
            ➕ Create Bounty
          </button>
        </div>

        <div className="bounty-filters">
          {(['all', 'open', 'claimed', 'completed'] as const).map((status) => (
            <button
              key={status}
              className={`filter-btn ${filter === status ? 'active' : ''} filter-${status}`}
              onClick={() => setFilter(status)}
            >
              {status === 'all' ? 'All' : status.charAt(0).toUpperCase() + status.slice(1)}
              <span className="filter-count">{counts[status as keyof typeof counts]}</span>
            </button>
          ))}
        </div>
      </div>

      <div className="bounties-grid">
        {isLoading ? (
          // Show skeleton loaders while bounties are loading
          [...Array(8)].map((_, i) => (
            <SkeletonQuestCard key={`skeleton-bounty-${i}`} />
          ))
        ) : (
          filteredBounties.map((bounty) => (
            <div key={bounty.id} className={`bounty-card tier-${bounty.tier} status-${bounty.status}`}>
              <div className="bounty-card-header">
                <div className="bounty-title-row">
                  <h3 className="bounty-title">{bounty.title}</h3>
                  <div className="bounty-icons">
                    <span className="tier-icon" title={`${bounty.tier} tier`}>
                      {getTierIcon(bounty.tier)}
                    </span>
                    <span className="status-icon" title={bounty.status}>
                      {getStatusIcon(bounty.status)}
                    </span>
                  </div>
                </div>
                <div className="bounty-payout">
                  <span className="payout-amount">🪙 {bounty.payout_amount}</span>
                  <span className="payout-label">Gold</span>
                </div>
              </div>

              <div className="bounty-description">
                <p>{bounty.description.slice(0, 150)}{bounty.description.length > 150 ? '...' : ''}</p>
              </div>

              {bounty.required_skills.length > 0 && (
                <div className="required-skills">
                  <span className="skills-label">Required Skills:</span>
                  <div className="skills-list">
                    {bounty.required_skills.slice(0, 3).map((skill, i) => (
                      <span key={i} className="skill-tag">{skill}</span>
                    ))}
                    {bounty.required_skills.length > 3 && (
                      <span className="skill-tag more">+{bounty.required_skills.length - 3} more</span>
                    )}
                  </div>
                </div>
              )}

              <div className="bounty-meta">
                <div className="meta-item">
                  <span className="meta-label">Posted:</span>
                  <span className="meta-value">{new Date(bounty.created_at).toLocaleDateString()}</span>
                </div>
                {bounty.deadline && (
                  <div className="meta-item deadline">
                    <span className="meta-label">Deadline:</span>
                    <span className="meta-value">{formatTimeRemaining(bounty.deadline)}</span>
                  </div>
                )}
              </div>

              <div className="bounty-actions">
                {bounty.status === 'open' && (
                  <button className="claim-btn">
                    🔒 Claim Bounty
                  </button>
                )}
                {bounty.status === 'claimed' && (
                  <button className="submit-btn" disabled>
                    🔄 In Progress
                  </button>
                )}
                {bounty.status === 'completed' && (
                  <button className="completed-btn" disabled>
                    ✅ Completed
                  </button>
                )}
              </div>
            </div>
          ))
        )}
        
        {!isLoading && filteredBounties.length === 0 && (
          <div className="no-bounties">
            <div className="no-bounties-icon">💰</div>
            <h3>No bounties found</h3>
            <p>No bounties match the current filter.</p>
          </div>
        )}
      </div>

      <CreateBountyForm
        isOpen={showCreateForm}
        onClose={() => setShowCreateForm(false)}
        onSubmit={handleCreateBounty}
      />
    </div>
  );
}