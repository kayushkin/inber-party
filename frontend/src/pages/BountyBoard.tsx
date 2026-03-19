import { useState, useEffect } from 'react';
import { SkeletonQuestCard } from '../components/SkeletonLoader';
import CreateBountyForm from '../components/CreateBountyForm';
import { useNotifications } from '../contexts/NotificationContext';
import './BountyBoard.css';

interface Bounty {
  id: number;
  title: string;
  description: string;
  requirements: string;
  payout_amount: number;
  status: 'open' | 'claimed' | 'submitted' | 'completed' | 'rejected' | 'paid' | 'disputed';
  deadline?: string;
  creator_id: number;
  claimer_id?: number;
  work_submission?: string;
  verification_notes?: string;
  rating?: {
    rating: number;
    comment: string;
    created_at: string;
  };
  required_skills: string[];
  tier: 'bronze' | 'silver' | 'gold' | 'legendary';
  claimed_at?: string;
  submitted_at?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
}

interface Agent {
  id: number;
  name: string;
  title: string;
  class: string;
  avatar_emoji: string;
  level: number;
}

export default function BountyBoard() {
  const { showSuccess, showError } = useNotifications();
  const [bounties, setBounties] = useState<Bounty[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [filter, setFilter] = useState<string>('all');
  const [sortBy, setSortBy] = useState<'newest' | 'oldest' | 'payout-high' | 'payout-low' | 'deadline'>('newest');
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [agents, setAgents] = useState<Agent[]>([]);
  const [showClaimModal, setShowClaimModal] = useState(false);
  const [selectedBountyId, setSelectedBountyId] = useState<number | null>(null);
  const [claimingBounty, setClaimingBounty] = useState(false);
  const [showVerificationModal, setShowVerificationModal] = useState(false);
  const [selectedBounty, setSelectedBounty] = useState<Bounty | null>(null);
  const [verificationNotes, setVerificationNotes] = useState('');
  const [verifying, setVerifying] = useState(false);
  const [showRatingModal, setShowRatingModal] = useState(false);
  const [selectedBountyForRating, setSelectedBountyForRating] = useState<Bounty | null>(null);
  const [rating, setRating] = useState(5);
  const [ratingComment, setRatingComment] = useState('');
  const [submittingRating, setSubmittingRating] = useState(false);
  const [showDisputeModal, setShowDisputeModal] = useState(false);
  const [selectedBountyForDispute, setSelectedBountyForDispute] = useState<Bounty | null>(null);
  const [disputeReason, setDisputeReason] = useState('');
  const [disputeEvidence, setDisputeEvidence] = useState('');
  const [submittingDispute, setSubmittingDispute] = useState(false);
  
  useEffect(() => {
    fetchBounties();
    fetchAgents();
  }, []);

  const fetchBounties = async () => {
    try {
      setIsLoading(true);
      const response = await fetch('/api/bounties');
      if (response.ok) {
        const data = await response.json();
        setBounties(data);
      } else {
        showError('Failed to fetch bounties', `Server responded with status: ${response.status}`);
      }
    } catch (error) {
      showError('Error fetching bounties', `${error}`);
    } finally {
      setIsLoading(false);
    }
  };

  const fetchAgents = async () => {
    try {
      const response = await fetch('/api/agents');
      if (response.ok) {
        const data = await response.json();
        setAgents(data);
      } else {
        showError('Failed to fetch agents', `Server responded with status: ${response.status}`);
      }
    } catch (error) {
      showError('Error fetching agents', `${error}`);
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
      showSuccess('Bounty Created!', `Successfully created bounty: ${newBounty.title}`);
    } catch (error) {
      throw error; // Re-throw to let the form handle the error
    }
  };

  const handleClaimClick = (bountyId: number) => {
    setSelectedBountyId(bountyId);
    setShowClaimModal(true);
  };

  const handleClaimBounty = async (agentId: number) => {
    if (!selectedBountyId) return;

    try {
      setClaimingBounty(true);
      const response = await fetch(`/api/bounties/${selectedBountyId}/claim`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          claimer_id: agentId,
        }),
      });

      if (!response.ok) {
        const errorData = await response.text();
        throw new Error(`Failed to claim bounty: ${errorData}`);
      }

      // Refresh bounties to show updated status
      showSuccess('Bounty Claimed!', 'You have successfully claimed this bounty.');
      await fetchBounties();
      setShowClaimModal(false);
      setSelectedBountyId(null);
    } catch (error) {
      showError('Error claiming bounty', `Failed to claim bounty: ${error}`);
    } finally {
      setClaimingBounty(false);
    }
  };

  const handleVerifyClick = (bounty: Bounty) => {
    setSelectedBounty(bounty);
    setVerificationNotes('');
    setShowVerificationModal(true);
  };

  const handleVerifyBounty = async (approved: boolean) => {
    if (!selectedBounty) return;

    try {
      setVerifying(true);
      const response = await fetch(`/api/bounties/${selectedBounty.id}/verify`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          approved: approved,
          notes: verificationNotes,
        }),
      });

      if (!response.ok) {
        const errorData = await response.text();
        throw new Error(`Failed to verify bounty: ${errorData}`);
      }

      // Refresh bounties to show updated status
      showSuccess('Bounty Verified!', `Bounty has been ${approved ? 'approved' : 'rejected'} successfully.`);
      await fetchBounties();
      setShowVerificationModal(false);
      setSelectedBounty(null);
      setVerificationNotes('');
    } catch (error) {
      showError('Error verifying bounty', `Failed to verify bounty: ${error}`);
    } finally {
      setVerifying(false);
    }
  };

  const handleRateClick = (bounty: Bounty) => {
    setSelectedBountyForRating(bounty);
    setRating(5);
    setRatingComment('');
    setShowRatingModal(true);
  };

  const handleSubmitRating = async () => {
    if (!selectedBountyForRating) return;

    try {
      setSubmittingRating(true);
      const response = await fetch('/api/ratings', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          bounty_id: selectedBountyForRating.id,
          rater_id: selectedBountyForRating.creator_id,
          rated_id: selectedBountyForRating.claimer_id,
          rating: rating,
          comment: ratingComment,
          categories: {
            quality: rating,
            timeliness: rating,
            communication: rating
          }
        }),
      });

      if (!response.ok) {
        const errorData = await response.text();
        throw new Error(`Failed to submit rating: ${errorData}`);
      }

      // Refresh bounties to show updated status
      showSuccess('Rating Submitted!', 'Your rating has been submitted successfully.');
      await fetchBounties();
      setShowRatingModal(false);
      setSelectedBountyForRating(null);
      setRating(5);
      setRatingComment('');
    } catch (error) {
      showError('Error submitting rating', `Failed to submit rating: ${error}`);
    } finally {
      setSubmittingRating(false);
    }
  };

  const handleDisputeClick = (bounty: Bounty) => {
    setSelectedBountyForDispute(bounty);
    setDisputeReason('');
    setDisputeEvidence('');
    setShowDisputeModal(true);
  };

  const handleSubmitDispute = async () => {
    if (!selectedBountyForDispute || !disputeReason.trim()) return;

    try {
      setSubmittingDispute(true);
      const response = await fetch('/api/disputes', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          bounty_id: selectedBountyForDispute.id,
          claimer_id: selectedBountyForDispute.claimer_id,
          reason: disputeReason,
          evidence: disputeEvidence,
        }),
      });

      if (!response.ok) {
        const errorData = await response.text();
        throw new Error(`Failed to submit dispute: ${errorData}`);
      }

      // Refresh bounties to show updated status
      await fetchBounties();
      setShowDisputeModal(false);
      setSelectedBountyForDispute(null);
      setDisputeReason('');
      setDisputeEvidence('');
      
      showSuccess('Dispute Submitted!', 'Your dispute has been submitted successfully. The bounty is now under review.');
    } catch (error) {
      showError('Error submitting dispute', `Failed to submit dispute: ${error}`);
    } finally {
      setSubmittingDispute(false);
    }
  };

  const filteredAndSortedBounties = (() => {
    let filtered = filter === 'all' 
      ? bounties 
      : bounties.filter(bounty => bounty.status === filter);

    // Apply sorting
    switch (sortBy) {
      case 'newest':
        filtered = [...filtered].sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
        break;
      case 'oldest':
        filtered = [...filtered].sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime());
        break;
      case 'payout-high':
        filtered = [...filtered].sort((a, b) => b.payout_amount - a.payout_amount);
        break;
      case 'payout-low':
        filtered = [...filtered].sort((a, b) => a.payout_amount - b.payout_amount);
        break;
      case 'deadline':
        filtered = [...filtered].sort((a, b) => {
          // Bounties with deadlines come first, sorted by deadline
          if (!a.deadline && !b.deadline) return 0;
          if (!a.deadline) return 1;
          if (!b.deadline) return -1;
          return new Date(a.deadline).getTime() - new Date(b.deadline).getTime();
        });
        break;
      default:
        break;
    }

    return filtered;
  })();

  const counts = {
    all: bounties.length,
    open: bounties.filter(b => b.status === 'open').length,
    claimed: bounties.filter(b => b.status === 'claimed').length,
    submitted: bounties.filter(b => b.status === 'submitted').length,
    completed: bounties.filter(b => b.status === 'completed').length,
    disputed: bounties.filter(b => b.status === 'disputed').length,
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
      case 'submitted': return '📋';
      case 'completed': return '✅';
      case 'rejected': return '❌';
      case 'paid': return '💰';
      case 'disputed': return '⚖️';
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

        <div className="bounty-controls">
          <div className="bounty-filters">
            {(['all', 'open', 'claimed', 'submitted', 'completed', 'disputed'] as const).map((status) => (
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
          
          <div className="bounty-sort">
            <label htmlFor="sort-select">Sort by:</label>
            <select 
              id="sort-select"
              className="sort-select"
              value={sortBy}
              onChange={(e) => setSortBy(e.target.value as typeof sortBy)}
            >
              <option value="newest">🕐 Newest First</option>
              <option value="oldest">🕑 Oldest First</option>
              <option value="payout-high">💰 Highest Payout</option>
              <option value="payout-low">🪙 Lowest Payout</option>
              <option value="deadline">⏰ Deadline Soon</option>
            </select>
          </div>
        </div>
      </div>

      <div className="bounties-grid">
        {isLoading ? (
          // Show skeleton loaders while bounties are loading
          [...Array(8)].map((_, i) => (
            <SkeletonQuestCard key={`skeleton-bounty-${i}`} />
          ))
        ) : (
          filteredAndSortedBounties.map((bounty) => (
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
                  <div className={`meta-item deadline ${formatTimeRemaining(bounty.deadline) === 'Expired' ? 'expired' : ''}`}>
                    <span className="meta-label">Deadline:</span>
                    <span className="meta-value">{formatTimeRemaining(bounty.deadline)}</span>
                  </div>
                )}
                {bounty.claimer_id && (
                  <div className="meta-item claimer-info">
                    <span className="meta-label">
                      {bounty.status === 'claimed' && 'Claimed by:'}
                      {bounty.status === 'submitted' && 'Submitted by:'}
                      {bounty.status === 'completed' && 'Completed by:'}
                      {bounty.status === 'rejected' && 'Worked on by:'}
                      {bounty.status === 'paid' && 'Completed by:'}
                      {bounty.status === 'disputed' && 'Disputed by:'}
                    </span>
                    <span className="meta-value claimer-name">
                      {(() => {
                        const claimer = agents.find(agent => agent.id === bounty.claimer_id);
                        if (claimer) {
                          return (
                            <span className="claimer-details">
                              <span className="claimer-avatar">{claimer.avatar_emoji}</span>
                              <span className="claimer-info-text">
                                {claimer.name} 
                                <span className="claimer-level">L{claimer.level}</span>
                              </span>
                            </span>
                          );
                        }
                        return `Agent #${bounty.claimer_id}`;
                      })()}
                    </span>
                  </div>
                )}
                {bounty.claimed_at && bounty.status !== 'open' && (
                  <div className="meta-item claimed-time">
                    <span className="meta-label">
                      {bounty.status === 'claimed' && 'Claimed:'}
                      {bounty.status === 'submitted' && 'Submitted:'}
                      {bounty.status === 'completed' && 'Completed:'}
                      {bounty.status === 'rejected' && 'Rejected:'}
                      {bounty.status === 'paid' && 'Paid:'}
                      {bounty.status === 'disputed' && 'Disputed:'}
                    </span>
                    <span className="meta-value">
                      {(() => {
                        const targetDate = bounty.status === 'submitted' && bounty.submitted_at 
                          ? bounty.submitted_at 
                          : bounty.status === 'completed' && bounty.completed_at 
                          ? bounty.completed_at 
                          : bounty.claimed_at;
                        return targetDate ? new Date(targetDate).toLocaleDateString() : 'Unknown';
                      })()}
                    </span>
                  </div>
                )}
              </div>

              <div className="bounty-actions">
                {bounty.status === 'open' && (
                  <button 
                    className="claim-btn"
                    onClick={() => handleClaimClick(bounty.id)}
                  >
                    🔒 Claim Bounty
                  </button>
                )}
                {bounty.status === 'claimed' && (
                  <button className="submit-btn" disabled>
                    🔄 In Progress
                  </button>
                )}
                {bounty.status === 'submitted' && (
                  <button 
                    className="verify-btn"
                    onClick={() => handleVerifyClick(bounty)}
                  >
                    📋 Review Work
                  </button>
                )}
                {bounty.status === 'completed' && !bounty.rating && (
                  <button 
                    className="rate-btn"
                    onClick={() => handleRateClick(bounty)}
                  >
                    ⭐ Rate Work
                  </button>
                )}
                {bounty.status === 'completed' && bounty.rating && (
                  <div className="rating-display">
                    <span className="rating-stars">
                      {'⭐'.repeat(bounty.rating.rating)}
                    </span>
                    <span className="rating-text">Rated {bounty.rating.rating}/5</span>
                  </div>
                )}
                {bounty.status === 'rejected' && (
                  <div className="rejected-actions">
                    <button className="rejected-btn" disabled>
                      ❌ Rejected
                    </button>
                    <button 
                      className="dispute-btn"
                      onClick={() => handleDisputeClick(bounty)}
                      title="File a dispute if you believe the rejection was unfair"
                    >
                      ⚖️ Dispute
                    </button>
                  </div>
                )}
                {bounty.status === 'disputed' && (
                  <button className="disputed-btn" disabled>
                    ⚖️ Under Dispute Review
                  </button>
                )}
                {bounty.status === 'paid' && bounty.rating && (
                  <div className="bounty-completed">
                    <button className="completed-btn" disabled>
                      ✅ Paid & Rated
                    </button>
                    <div className="rating-display">
                      <span className="rating-stars">
                        {'⭐'.repeat(bounty.rating.rating)}
                      </span>
                    </div>
                  </div>
                )}
              </div>
            </div>
          ))
        )}
        
        {!isLoading && filteredAndSortedBounties.length === 0 && (
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

      {/* Agent Selection Modal for Claiming */}
      {showClaimModal && (
        <div className="modal-overlay">
          <div className="modal-content">
            <div className="modal-header">
              <h2>🔒 Claim Bounty</h2>
              <button 
                className="modal-close"
                onClick={() => setShowClaimModal(false)}
                disabled={claimingBounty}
              >
                ×
              </button>
            </div>
            
            <div className="modal-body">
              <p>Select which agent should claim this bounty:</p>
              
              <div className="agents-list">
                {agents.map((agent) => (
                  <button
                    key={agent.id}
                    className="agent-option"
                    onClick={() => handleClaimBounty(agent.id)}
                    disabled={claimingBounty}
                  >
                    <span className="agent-avatar">{agent.avatar_emoji}</span>
                    <div className="agent-info">
                      <div className="agent-name">{agent.name}</div>
                      <div className="agent-title">{agent.title}</div>
                      <div className="agent-meta">
                        <span className="agent-class">{agent.class}</span>
                        <span className="agent-level">Level {agent.level}</span>
                      </div>
                    </div>
                  </button>
                ))}
              </div>
              
              {agents.length === 0 && (
                <div className="no-agents">
                  <p>No agents available. Create an agent first to claim bounties.</p>
                </div>
              )}
            </div>
            
            {claimingBounty && (
              <div className="modal-footer">
                <div className="claiming-status">
                  🔄 Claiming bounty...
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Verification Modal for Reviewing Submitted Work */}
      {showVerificationModal && selectedBounty && (
        <div className="modal-overlay">
          <div className="modal-content verification-modal">
            <div className="modal-header">
              <h2>📋 Review Submitted Work</h2>
              <button 
                className="modal-close"
                onClick={() => setShowVerificationModal(false)}
                disabled={verifying}
              >
                ×
              </button>
            </div>
            
            <div className="modal-body">
              <div className="bounty-summary">
                <h3>{selectedBounty.title}</h3>
                <p><strong>Payout:</strong> 🪙 {selectedBounty.payout_amount} Gold</p>
                <p><strong>Requirements:</strong> {selectedBounty.requirements}</p>
              </div>

              <div className="work-submission">
                <h4>📝 Submitted Work</h4>
                <div className="submission-content">
                  {selectedBounty.work_submission ? (
                    <pre className="submission-text">{selectedBounty.work_submission}</pre>
                  ) : (
                    <p className="no-submission">No work submitted yet.</p>
                  )}
                </div>
              </div>

              <div className="verification-notes-section">
                <h4>💬 Verification Notes</h4>
                <textarea
                  className="verification-notes-input"
                  value={verificationNotes}
                  onChange={(e) => setVerificationNotes(e.target.value)}
                  placeholder="Add notes about the work quality, any issues, or feedback..."
                  rows={4}
                />
              </div>
            </div>
            
            <div className="modal-footer">
              {verifying ? (
                <div className="verifying-status">
                  🔄 Processing verification...
                </div>
              ) : (
                <div className="verification-actions">
                  <button 
                    className="verify-reject-btn"
                    onClick={() => handleVerifyBounty(false)}
                  >
                    ❌ Reject Work
                  </button>
                  <button 
                    className="verify-approve-btn"
                    onClick={() => handleVerifyBounty(true)}
                  >
                    ✅ Approve Work
                  </button>
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Rating Modal for Rating Completed Work */}
      {showRatingModal && selectedBountyForRating && (
        <div className="modal-overlay">
          <div className="modal-content rating-modal">
            <div className="modal-header">
              <h2>⭐ Rate Completed Work</h2>
              <button 
                className="modal-close"
                onClick={() => setShowRatingModal(false)}
                disabled={submittingRating}
              >
                ×
              </button>
            </div>
            
            <div className="modal-body">
              <div className="bounty-summary">
                <h3>{selectedBountyForRating.title}</h3>
                <p><strong>Completed by:</strong> Agent #{selectedBountyForRating.claimer_id}</p>
                <p><strong>Payout:</strong> 🪙 {selectedBountyForRating.payout_amount} Gold</p>
              </div>

              <div className="rating-section">
                <h4>⭐ Overall Rating</h4>
                <div className="star-rating">
                  {[1, 2, 3, 4, 5].map((star) => (
                    <button
                      key={star}
                      className={`star-btn ${star <= rating ? 'active' : ''}`}
                      onClick={() => setRating(star)}
                      disabled={submittingRating}
                    >
                      ⭐
                    </button>
                  ))}
                  <span className="rating-label">
                    {rating === 1 && 'Poor'}
                    {rating === 2 && 'Fair'}
                    {rating === 3 && 'Good'}
                    {rating === 4 && 'Very Good'}
                    {rating === 5 && 'Excellent'}
                  </span>
                </div>
              </div>

              <div className="comment-section">
                <h4>💬 Feedback (Optional)</h4>
                <textarea
                  className="rating-comment-input"
                  value={ratingComment}
                  onChange={(e) => setRatingComment(e.target.value)}
                  placeholder="Share feedback about the work quality, timeliness, communication..."
                  rows={4}
                  disabled={submittingRating}
                />
              </div>
            </div>
            
            <div className="modal-footer">
              {submittingRating ? (
                <div className="submitting-status">
                  🔄 Submitting rating...
                </div>
              ) : (
                <div className="rating-actions">
                  <button 
                    className="cancel-btn"
                    onClick={() => setShowRatingModal(false)}
                  >
                    Cancel
                  </button>
                  <button 
                    className="submit-rating-btn"
                    onClick={handleSubmitRating}
                  >
                    ⭐ Submit Rating
                  </button>
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Dispute Modal for Filing Disputes */}
      {showDisputeModal && selectedBountyForDispute && (
        <div className="modal-overlay">
          <div className="modal-content dispute-modal">
            <div className="modal-header">
              <h2>⚖️ File Dispute</h2>
              <button 
                className="modal-close"
                onClick={() => setShowDisputeModal(false)}
                disabled={submittingDispute}
              >
                ×
              </button>
            </div>
            
            <div className="modal-body">
              <div className="dispute-explanation">
                <p><strong>📋 Filing a Dispute</strong></p>
                <p>If you believe your work was unfairly rejected, you can file a dispute. 
                   Please provide a clear explanation and any supporting evidence.</p>
              </div>

              <div className="bounty-summary">
                <h3>{selectedBountyForDispute.title}</h3>
                <p><strong>Payout:</strong> 🪙 {selectedBountyForDispute.payout_amount} Gold</p>
                {selectedBountyForDispute.verification_notes && (
                  <div className="rejection-reason">
                    <p><strong>Rejection reason:</strong></p>
                    <div className="rejection-notes">
                      {selectedBountyForDispute.verification_notes}
                    </div>
                  </div>
                )}
              </div>

              <div className="dispute-reason-section">
                <h4>📝 Why do you believe the rejection was unfair? *</h4>
                <textarea
                  className="dispute-reason-input"
                  value={disputeReason}
                  onChange={(e) => setDisputeReason(e.target.value)}
                  placeholder="Explain why you believe your work meets the requirements and should be accepted..."
                  rows={4}
                  disabled={submittingDispute}
                  required
                />
              </div>

              <div className="dispute-evidence-section">
                <h4>🔗 Supporting Evidence (Optional)</h4>
                <textarea
                  className="dispute-evidence-input"
                  value={disputeEvidence}
                  onChange={(e) => setDisputeEvidence(e.target.value)}
                  placeholder="Provide links to your work, screenshots, documentation, test results, or any other evidence that supports your case..."
                  rows={4}
                  disabled={submittingDispute}
                />
                <p className="evidence-help">
                  💡 Include links to GitHub commits, deployed demos, test results, 
                  screenshots, or any documentation that proves your work is complete.
                </p>
              </div>
            </div>
            
            <div className="modal-footer">
              {submittingDispute ? (
                <div className="submitting-status">
                  🔄 Submitting dispute...
                </div>
              ) : (
                <div className="dispute-actions">
                  <button 
                    className="cancel-btn"
                    onClick={() => setShowDisputeModal(false)}
                  >
                    Cancel
                  </button>
                  <button 
                    className="submit-dispute-btn"
                    onClick={handleSubmitDispute}
                    disabled={!disputeReason.trim()}
                  >
                    ⚖️ Submit Dispute
                  </button>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}