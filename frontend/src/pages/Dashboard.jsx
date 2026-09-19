import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../services/api';
import { 
  BarChart2, 
  PlusCircle, 
  Copy, 
  Check, 
  Eye, 
  ExternalLink, 
  ToggleLeft, 
  ToggleRight, 
  Trash2, 
  Clock, 
  Vote, 
  Activity, 
  AlertCircle 
} from 'lucide-react';

export default function Dashboard({ showToast }) {
  const [polls, setPolls] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [copiedId, setCopiedId] = useState(null);

  const fetchPolls = async () => {
    try {
      setLoading(true);
      const res = await api.getUserPolls();
      setPolls(res.polls || []);
    } catch (err) {
      setError(err.message || 'Failed to load polls');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPolls();
  }, []);

  const handleCopyLink = (shareCode, id) => {
    const shareUrl = `${window.location.origin}/poll/${shareCode}`;
    navigator.clipboard.writeText(shareUrl).then(() => {
      setCopiedId(id);
      showToast('Share link copied to clipboard!', 'success');
      setTimeout(() => setCopiedId(null), 2000);
    });
  };

  const handleToggleStatus = async (poll) => {
    const nextStatus = poll.status === 'active' ? 'closed' : 'active';
    try {
      await api.updatePollStatus(poll.id, nextStatus);
      showToast(`Poll ${nextStatus === 'active' ? 'reopened' : 'closed'} successfully`, 'info');
      setPolls(polls.map(p => p.id === poll.id ? { ...p, status: nextStatus } : p));
    } catch (err) {
      showToast(err.message || 'Failed to update status', 'error');
    }
  };

  const handleDelete = async (pollId) => {
    if (!window.confirm('Are you sure you want to delete this poll? All votes and data will be removed.')) {
      return;
    }

    try {
      await api.deletePoll(pollId);
      showToast('Poll deleted successfully', 'info');
      setPolls(polls.filter(p => p.id !== pollId));
    } catch (err) {
      showToast(err.message || 'Failed to delete poll', 'error');
    }
  };

  // Metrics
  const totalPolls = polls.length;
  const activePolls = polls.filter(p => p.status === 'active').length;
  const totalVotesReceived = polls.reduce((acc, p) => {
    return acc + (p.options ? p.options.reduce((sum, opt) => sum + (opt.votes || 0), 0) : 0);
  }, 0);

  return (
    <div>
      {/* Header & Metrics */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem', flexWrap: 'wrap', gap: '1rem' }}>
        <div>
          <h1>Poll Dashboard</h1>
          <p style={{ color: 'var(--text-muted)', marginTop: '0.25rem' }}>
            Monitor and manage your active live polls in real-time
          </p>
        </div>
        <Link to="/create-poll" className="btn btn-primary">
          <PlusCircle size={18} />
          <span>Create New Poll</span>
        </Link>
      </div>

      {/* Metric Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: '1.25rem', marginBottom: '2.5rem' }}>
        <div className="glass-card" style={{ padding: '1.5rem', display: 'flex', alignItems: 'center', gap: '1.25rem' }}>
          <div style={{ width: '48px', height: '48px', borderRadius: '12px', background: 'rgba(99, 102, 241, 0.15)', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--primary)' }}>
            <BarChart2 size={24} />
          </div>
          <div>
            <div style={{ fontSize: '0.82rem', color: 'var(--text-dim)', fontWeight: 600, textTransform: 'uppercase' }}>Total Polls</div>
            <div style={{ fontSize: '1.75rem', fontWeight: 800 }}>{totalPolls}</div>
          </div>
        </div>

        <div className="glass-card" style={{ padding: '1.5rem', display: 'flex', alignItems: 'center', gap: '1.25rem' }}>
          <div style={{ width: '48px', height: '48px', borderRadius: '12px', background: 'rgba(16, 185, 129, 0.15)', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#10b981' }}>
            <Activity size={24} />
          </div>
          <div>
            <div style={{ fontSize: '0.82rem', color: 'var(--text-dim)', fontWeight: 600, textTransform: 'uppercase' }}>Active Polls</div>
            <div style={{ fontSize: '1.75rem', fontWeight: 800 }}>{activePolls}</div>
          </div>
        </div>

        <div className="glass-card" style={{ padding: '1.5rem', display: 'flex', alignItems: 'center', gap: '1.25rem' }}>
          <div style={{ width: '48px', height: '48px', borderRadius: '12px', background: 'rgba(6, 182, 212, 0.15)', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#06b6d4' }}>
            <Vote size={24} />
          </div>
          <div>
            <div style={{ fontSize: '0.82rem', color: 'var(--text-dim)', fontWeight: 600, textTransform: 'uppercase' }}>Total Votes</div>
            <div style={{ fontSize: '1.75rem', fontWeight: 800 }}>{totalVotesReceived}</div>
          </div>
        </div>
      </div>

      {error && (
        <div className="alert alert-danger">
          <AlertCircle size={18} />
          <span>{error}</span>
        </div>
      )}

      {/* Polls List */}
      <div>
        <h3 style={{ marginBottom: '1.25rem' }}>Your Created Polls</h3>

        {loading ? (
          <div style={{ display: 'flex', justifyContent: 'center', padding: '4rem 0' }}>
            <div className="spinner" style={{ width: '36px', height: '36px' }} />
          </div>
        ) : polls.length === 0 ? (
          <div className="glass-card" style={{ padding: '4rem 2rem', textAlign: 'center' }}>
            <div style={{ width: '64px', height: '64px', borderRadius: '50%', background: 'rgba(99, 102, 241, 0.1)', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 1.5rem', color: 'var(--primary)' }}>
              <Vote size={32} />
            </div>
            <h3>No Polls Created Yet</h3>
            <p style={{ color: 'var(--text-muted)', maxWidth: '400px', margin: '0.5rem auto 1.75rem', fontSize: '0.95rem' }}>
              Create your first live poll, share the link with friends or audience, and watch votes stream in live!
            </p>
            <Link to="/create-poll" className="btn btn-primary">
              <PlusCircle size={18} />
              <span>Create Your First Poll</span>
            </Link>
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
            {polls.map((poll) => {
              const pollVotes = poll.options ? poll.options.reduce((sum, o) => sum + (o.votes || 0), 0) : 0;
              return (
                <div key={poll.id} className="glass-card" style={{ padding: '1.5rem' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: '0.75rem', marginBottom: '1rem' }}>
                    <div>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '0.5rem', flexWrap: 'wrap' }}>
                        <h4 style={{ fontSize: '1.15rem' }}>{poll.question}</h4>
                        {poll.status === 'active' ? (
                          <span className="badge-active">Active</span>
                        ) : (
                          <span className="badge-closed">Closed</span>
                        )}
                        <span style={{ fontSize: '0.75rem', color: 'var(--text-dim)', background: 'rgba(255,255,255,0.05)', padding: '0.15rem 0.5rem', borderRadius: '4px' }}>
                          Code: <strong>{poll.share_code}</strong>
                        </span>
                      </div>
                      
                      <div style={{ display: 'flex', alignItems: 'center', gap: '1.25rem', color: 'var(--text-muted)', fontSize: '0.85rem' }}>
                        <span>Options: <strong>{poll.options ? poll.options.length : 0}</strong></span>
                        <span>•</span>
                        <span>Total Votes: <strong style={{ color: 'var(--text-main)' }}>{pollVotes}</strong></span>
                        {poll.expires_at && (
                          <>
                            <span>•</span>
                            <span style={{ display: 'flex', alignItems: 'center', gap: '0.3rem' }}>
                              <Clock size={14} />
                              Expires: {new Date(poll.expires_at).toLocaleString()}
                            </span>
                          </>
                        )}
                      </div>
                    </div>

                    {/* Action buttons */}
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', flexWrap: 'wrap' }}>
                      <button
                        onClick={() => handleCopyLink(poll.share_code, poll.id)}
                        className="btn btn-secondary btn-sm"
                        title="Copy audience voting link"
                      >
                        {copiedId === poll.id ? <Check size={14} color="#10b981" /> : <Copy size={14} />}
                        <span>{copiedId === poll.id ? 'Copied!' : 'Copy Link'}</span>
                      </button>

                      <Link
                        to={`/poll/${poll.share_code}`}
                        className="btn btn-outline btn-sm"
                        title="Open voting page"
                        target="_blank"
                      >
                        <ExternalLink size={14} />
                        <span>Vote Page</span>
                      </Link>

                      <Link
                        to={`/poll/${poll.id}/results`}
                        className="btn btn-primary btn-sm"
                        title="View live results with animated charts"
                      >
                        <Eye size={14} />
                        <span>Live Results</span>
                      </Link>

                      <button
                        onClick={() => handleToggleStatus(poll)}
                        className="btn btn-outline btn-sm"
                        title={poll.status === 'active' ? 'Close poll to new votes' : 'Reopen poll'}
                      >
                        {poll.status === 'active' ? (
                          <>
                            <ToggleRight size={14} color="#10b981" />
                            <span>Close</span>
                          </>
                        ) : (
                          <>
                            <ToggleLeft size={14} color="#ef4444" />
                            <span>Reopen</span>
                          </>
                        )}
                      </button>

                      <button
                        onClick={() => handleDelete(poll.id)}
                        className="btn btn-danger btn-sm btn-icon"
                        title="Delete Poll"
                      >
                        <Trash2 size={14} />
                      </button>
                    </div>
                  </div>

                  {/* Options bar preview */}
                  <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap', marginTop: '0.75rem', paddingTop: '0.75rem', borderTop: '1px solid var(--border-subtle)' }}>
                    {poll.options && poll.options.map((opt) => (
                      <span key={opt.id} style={{ fontSize: '0.8rem', background: 'rgba(255,255,255,0.03)', border: '1px solid var(--border-subtle)', padding: '0.2rem 0.6rem', borderRadius: '6px' }}>
                        {opt.text} <span style={{ color: 'var(--primary)', fontWeight: 700 }}>({opt.votes || 0})</span>
                      </span>
                    ))}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
