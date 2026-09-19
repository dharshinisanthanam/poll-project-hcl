import React, { useState, useEffect, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api } from '../services/api';
import { usePollLive } from '../hooks/usePollLive';
import { 
  BarChart3, 
  Copy, 
  Check, 
  Vote, 
  Clock, 
  ToggleLeft, 
  ToggleRight, 
  Share2, 
  Users, 
  Wifi, 
  WifiOff, 
  AlertCircle 
} from 'lucide-react';

export default function PollResults({ showToast }) {
  const { id, shareCode } = useParams();
  const pollParam = id || shareCode;

  const [results, setResults] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [copied, setCopied] = useState(false);
  const [lastVoteAnimation, setLastVoteAnimation] = useState(false);

  const fetchResults = useCallback(async () => {
    try {
      const data = await api.getResults(pollParam);
      setResults(data);
    } catch (err) {
      setError(err.message || 'Failed to load poll results');
    } finally {
      setLoading(false);
    }
  }, [pollParam]);

  useEffect(() => {
    fetchResults();
  }, [fetchResults]);

  // Real-time WebSocket live updates handler
  const handleLiveUpdate = useCallback((payload) => {
    setResults((prev) => {
      if (!prev) return payload;
      // Trigger subtle flash animation on new incoming votes
      setLastVoteAnimation(true);
      setTimeout(() => setLastVoteAnimation(false), 800);
      return {
        ...prev,
        ...payload,
        has_voted: prev.has_voted, // preserve local voter state
        is_owner: prev.is_owner,
      };
    });
  }, []);

  // Hook connecting to Redis Pub/Sub WebSocket broadcast
  const { isConnected } = usePollLive(results?.poll_id || pollParam, handleLiveUpdate);

  const handleCopyLink = () => {
    if (!results) return;
    const shareUrl = `${window.location.origin}/poll/${results.share_code}`;
    navigator.clipboard.writeText(shareUrl).then(() => {
      setCopied(true);
      showToast('Share link copied to clipboard!', 'success');
      setTimeout(() => setCopied(false), 2000);
    });
  };

  const handleToggleStatus = async () => {
    if (!results) return;
    const nextStatus = results.status === 'active' ? 'closed' : 'active';
    try {
      await api.updatePollStatus(results.poll_id, nextStatus);
      setResults({ ...results, status: nextStatus });
      showToast(`Poll ${nextStatus === 'active' ? 'reopened' : 'closed'} successfully`, 'info');
    } catch (err) {
      showToast(err.message || 'Failed to update poll status', 'error');
    }
  };

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', padding: '5rem 0' }}>
        <div className="spinner" style={{ width: '38px', height: '38px' }} />
      </div>
    );
  }

  if (error || !results) {
    return (
      <div style={{ maxWidth: '520px', margin: '4rem auto 0' }}>
        <div className="glass-card" style={{ padding: '3rem 2rem', textAlign: 'center' }}>
          <div style={{ width: '56px', height: '56px', borderRadius: '50%', background: 'rgba(239, 68, 68, 0.15)', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 1.5rem', color: '#ef4444' }}>
            <AlertCircle size={32} />
          </div>
          <h3>Results Not Found</h3>
          <p style={{ color: 'var(--text-muted)', margin: '0.5rem 0 1.75rem', fontSize: '0.95rem' }}>
            {error || 'Unable to retrieve live results for this poll.'}
          </p>
          <Link to="/" className="btn btn-primary">
            <span>Return to Home</span>
          </Link>
        </div>
      </div>
    );
  }

  // Find max votes to highlight the leading option
  const maxVotes = Math.max(...(results.options?.map(o => o.votes) || [0]));

  return (
    <div style={{ maxWidth: '720px', margin: '1.5rem auto 0' }}>
      <div className="glass-card" style={{ padding: '2.5rem 2rem' }}>
        {/* Top bar with real-time indicators */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem', flexWrap: 'wrap', gap: '0.75rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
            <div className="badge-live">
              <div className="pulse-dot" />
              <span>LIVE UPDATING</span>
            </div>

            <div style={{ display: 'flex', alignItems: 'center', gap: '0.35rem', fontSize: '0.78rem', color: isConnected ? '#34d399' : '#f87171' }}>
              {isConnected ? <Wifi size={14} /> : <WifiOff size={14} />}
              <span>{isConnected ? 'WebSocket Sync Connected' : 'Reconnecting...'}</span>
            </div>
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <span style={{ fontSize: '0.8rem', color: 'var(--text-dim)', background: 'rgba(255,255,255,0.05)', padding: '0.2rem 0.6rem', borderRadius: '6px' }}>
              Code: <strong>{results.share_code}</strong>
            </span>
            {results.status === 'active' ? (
              <span className="badge-active">Active</span>
            ) : (
              <span className="badge-closed">Closed</span>
            )}
          </div>
        </div>

        {/* Poll Question */}
        <h2 style={{ fontSize: '1.6rem', marginBottom: '1rem', lineHeight: '1.35' }}>
          {results.question}
        </h2>

        {/* Total Votes Banner */}
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', background: 'rgba(255,255,255,0.03)', padding: '0.85rem 1.25rem', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-subtle)', marginBottom: '2rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem', color: 'var(--text-muted)', fontSize: '0.92rem' }}>
            <Users size={18} />
            <span>Total Audience Votes:</span>
            <strong 
              id="total-votes-count"
              style={{ 
                color: 'var(--text-main)', 
                fontSize: '1.25rem', 
                fontWeight: 800,
                transition: 'transform 0.2s ease',
                display: 'inline-block',
                transform: lastVoteAnimation ? 'scale(1.2)' : 'scale(1)'
              }}
            >
              {results.total_votes}
            </strong>
          </div>

          <button
            id="copy-share-link-btn"
            onClick={handleCopyLink}
            className="btn btn-secondary btn-sm"
          >
            {copied ? <Check size={14} color="#10b981" /> : <Copy size={14} />}
            <span>{copied ? 'Copied!' : 'Copy Share Link'}</span>
          </button>
        </div>

        {/* Real-time Animated Bar Chart */}
        <div id="results-bar-container" style={{ marginBottom: '2.5rem' }}>
          {results.options?.map((opt, idx) => {
            const isLead = maxVotes > 0 && opt.votes === maxVotes;
            const pct = opt.percentage ? opt.percentage.toFixed(1) : '0.0';

            return (
              <div key={opt.id} className="result-item">
                <div className="result-header">
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                    <span style={{ color: 'var(--text-dim)', fontSize: '0.88rem', fontWeight: 700 }}>
                      {idx + 1}.
                    </span>
                    <span style={{ fontWeight: 600, color: 'var(--text-main)' }}>
                      {opt.text}
                    </span>
                    {isLead && results.total_votes > 0 && (
                      <span style={{ fontSize: '0.7rem', fontWeight: 800, background: 'rgba(6, 182, 212, 0.2)', color: '#38bdf8', padding: '0.1rem 0.45rem', borderRadius: '4px', border: '1px solid rgba(6, 182, 212, 0.3)' }}>
                        LEADING
                      </span>
                    )}
                  </div>
                  <div className="result-bar-stat">
                    <span>{opt.votes} {opt.votes === 1 ? 'vote' : 'votes'}</span>
                    <span style={{ color: 'var(--primary)', fontWeight: 800, minWidth: '48px', textAlign: 'right' }}>
                      {pct}%
                    </span>
                  </div>
                </div>

                <div className="result-bar-bg">
                  <div
                    className={`result-bar-fill ${isLead && results.total_votes > 0 ? 'lead' : ''}`}
                    style={{ width: `${Math.max(Number(pct), opt.votes > 0 ? 3 : 0)}%` }}
                  />
                  <div className="result-bar-content">
                    <span style={{ fontSize: '0.85rem', color: 'rgba(255,255,255,0.7)', fontWeight: 500 }}>
                      {opt.text}
                    </span>
                    <span style={{ fontSize: '0.85rem', color: 'rgba(255,255,255,0.9)', fontWeight: 700 }}>
                      {pct}%
                    </span>
                  </div>
                </div>
              </div>
            );
          })}
        </div>

        {/* Footer Actions */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem', paddingTop: '1.5rem', borderTop: '1px solid var(--border-subtle)' }}>
          <div style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap' }}>
            <Link
              to={`/poll/${results.share_code}`}
              className="btn btn-primary"
            >
              <Vote size={18} />
              <span>{results.has_voted ? 'View Voting Page' : 'Vote in this Poll'}</span>
            </Link>

            <button
              onClick={handleCopyLink}
              className="btn btn-secondary"
            >
              <Share2 size={16} />
              <span>Share Link</span>
            </button>
          </div>

          {/* Owner Controls */}
          {results.is_owner && (
            <button
              onClick={handleToggleStatus}
              className="btn btn-outline btn-sm"
              title={results.status === 'active' ? 'Close poll to new votes' : 'Reopen poll'}
            >
              {results.status === 'active' ? (
                <>
                  <ToggleRight size={16} color="#10b981" />
                  <span>Close Poll</span>
                </>
              ) : (
                <>
                  <ToggleLeft size={16} color="#ef4444" />
                  <span>Reopen Poll</span>
                </>
              )}
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
