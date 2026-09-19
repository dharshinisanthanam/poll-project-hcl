import React, { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { api } from '../services/api';
import confetti from 'canvas-confetti';
import { Vote, CheckCircle2, AlertCircle, BarChart3, Clock, AlertTriangle } from 'lucide-react';

export default function PublicPoll({ showToast }) {
  const { shareCode } = useParams();
  const navigate = useNavigate();

  const [poll, setPoll] = useState(null);
  const [selectedOption, setSelectedOption] = useState('');
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [alreadyVoted, setAlreadyVoted] = useState(false);
  const [isExpired, setIsExpired] = useState(false);

  useEffect(() => {
    async function loadPoll() {
      try {
        setLoading(true);
        // First check results endpoint to see if user has already voted
        const results = await api.getResults(shareCode).catch(() => null);
        if (results && results.has_voted) {
          setAlreadyVoted(true);
        }

        const res = await api.getPollByShareCode(shareCode);
        setPoll(res.poll);

        if (res.poll.status !== 'active') {
          setIsExpired(true);
        }
      } catch (err) {
        setError(err.message || 'Poll not found or inactive');
      } finally {
        setLoading(false);
      }
    }

    loadPoll();
  }, [shareCode]);

  const handleVoteSubmit = async (e) => {
    e.preventDefault();
    if (!selectedOption) {
      setError('Please select an option to vote');
      return;
    }

    setSubmitting(true);
    setError('');

    try {
      await api.castVote(shareCode, selectedOption);
      
      // Fire celebratory confetti!
      try {
        confetti({
          particleCount: 80,
          spread: 70,
          origin: { y: 0.6 }
        });
      } catch {
        // ignore confetti errors
      }

      showToast('Your vote has been cast successfully!', 'success');
      setAlreadyVoted(true);

      // Redirect to live results after a brief moment so user sees vote reflected
      setTimeout(() => {
        navigate(`/poll/${shareCode}/results`);
      }, 1200);
    } catch (err) {
      if (err.status === 409) {
        setAlreadyVoted(true);
        showToast('You have already voted in this poll', 'info');
      } else {
        setError(err.message || 'Failed to submit vote');
      }
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', padding: '5rem 0' }}>
        <div className="spinner" style={{ width: '38px', height: '38px' }} />
      </div>
    );
  }

  if (error && !poll) {
    return (
      <div style={{ maxWidth: '520px', margin: '4rem auto 0' }}>
        <div className="glass-card" style={{ padding: '3rem 2rem', textAlign: 'center' }}>
          <div style={{ width: '56px', height: '56px', borderRadius: '50%', background: 'rgba(239, 68, 68, 0.15)', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 1.5rem', color: '#ef4444' }}>
            <AlertCircle size={32} />
          </div>
          <h3>Poll Not Found</h3>
          <p style={{ color: 'var(--text-muted)', margin: '0.5rem 0 1.75rem', fontSize: '0.95rem' }}>
            {error || 'This poll link is invalid or may have been removed by the creator.'}
          </p>
          <Link to="/" className="btn btn-primary">
            <span>Return to Home</span>
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div style={{ maxWidth: '580px', margin: '2rem auto 0' }}>
      <div className="glass-card" style={{ padding: '2.5rem 2rem' }}>
        {/* Header Badges */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem', flexWrap: 'wrap', gap: '0.5rem' }}>
          <div className="badge-live">
            <div className="pulse-dot" />
            <span>LIVE AUDIENCE POLL</span>
          </div>

          <span style={{ fontSize: '0.8rem', color: 'var(--text-dim)', background: 'rgba(255,255,255,0.05)', padding: '0.2rem 0.6rem', borderRadius: '6px' }}>
            Code: <strong>{poll?.share_code}</strong>
          </span>
        </div>

        {/* Question */}
        <h2 style={{ fontSize: '1.45rem', marginBottom: '1.5rem', lineHeight: '1.4' }}>
          {poll?.question}
        </h2>

        {/* Status Alerts */}
        {isExpired && (
          <div className="alert alert-danger" style={{ marginBottom: '1.5rem' }}>
            <AlertTriangle size={18} style={{ flexShrink: 0, marginTop: '2px' }} />
            <div>
              <strong>This poll is closed.</strong> No more votes can be accepted.
            </div>
          </div>
        )}

        {alreadyVoted ? (
          <div style={{ textAlign: 'center', padding: '1.5rem 0' }}>
            <div style={{ width: '56px', height: '56px', borderRadius: '50%', background: 'rgba(16, 185, 129, 0.15)', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 1rem', color: '#10b981' }}>
              <CheckCircle2 size={32} />
            </div>
            <h3>Thank You for Voting!</h3>
            <p style={{ color: 'var(--text-muted)', fontSize: '0.95rem', margin: '0.5rem 0 1.75rem' }}>
              Your vote has been counted in real time. Watch live results update across all devices without refreshing!
            </p>
            <Link to={`/poll/${shareCode}/results`} className="btn btn-primary btn-lg" style={{ width: '100%' }}>
              <BarChart3 size={18} />
              <span>View Real-Time Results</span>
            </Link>
          </div>
        ) : (
          <form onSubmit={handleVoteSubmit}>
            {error && (
              <div className="alert alert-danger">
                <AlertCircle size={18} style={{ flexShrink: 0, marginTop: '2px' }} />
                <span>{error}</span>
              </div>
            )}

            {/* Voting Options */}
            <div style={{ marginBottom: '1.75rem' }}>
              {poll?.options?.map((opt) => (
                <div
                  key={opt.id}
                  id={`option-${opt.id}`}
                  className={`vote-option-card ${selectedOption === opt.id ? 'selected' : ''}`}
                  onClick={() => !isExpired && setSelectedOption(opt.id)}
                >
                  <div className="vote-radio-circle">
                    <div className="vote-radio-dot" />
                  </div>
                  <span style={{ fontSize: '1rem', fontWeight: 600, color: 'var(--text-main)' }}>
                    {opt.text}
                  </span>
                </div>
              ))}
            </div>

            {/* Actions */}
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.85rem' }}>
              <button
                id="submit-vote-btn"
                type="submit"
                disabled={submitting || !selectedOption || isExpired}
                className="btn btn-primary btn-lg"
                style={{ width: '100%' }}
              >
                {submitting ? (
                  <div className="spinner" />
                ) : (
                  <>
                    <Vote size={18} />
                    <span>Cast Your Vote</span>
                  </>
                )}
              </button>

              <Link
                to={`/poll/${shareCode}/results`}
                className="btn btn-outline"
                style={{ width: '100%' }}
              >
                <BarChart3 size={18} />
                <span>View Live Results (without voting)</span>
              </Link>
            </div>
          </form>
        )}
      </div>
    </div>
  );
}
