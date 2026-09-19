import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../services/api';
import { Plus, Trash2, Clock, Sparkles, AlertCircle, ArrowRight, CheckCircle2, Copy } from 'lucide-react';

export default function CreatePoll({ showToast }) {
  const [question, setQuestion] = useState('');
  const [options, setOptions] = useState(['Python', 'JavaScript', 'Go', 'Java']);
  const [expiresIn, setExpiresIn] = useState('never');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [createdPoll, setCreatedPoll] = useState(null);
  const navigate = useNavigate();

  const handleAddOption = () => {
    if (options.length >= 10) {
      setError('A poll can have at most 10 options');
      return;
    }
    setOptions([...options, '']);
  };

  const handleRemoveOption = (index) => {
    if (options.length <= 2) {
      setError('A poll must have at least 2 options');
      return;
    }
    const next = options.filter((_, i) => i !== index);
    setOptions(next);
  };

  const handleOptionChange = (index, value) => {
    const next = [...options];
    next[index] = value;
    setOptions(next);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    if (question.trim().length < 5) {
      setError('Question must be at least 5 characters long');
      return;
    }

    const trimmedOptions = options.map(o => o.trim()).filter(Boolean);
    if (trimmedOptions.length < 2) {
      setError('Please provide at least 2 non-empty options');
      return;
    }

    // Check duplicate options
    const unique = new Set(trimmedOptions.map(o => o.toLowerCase()));
    if (unique.size !== trimmedOptions.length) {
      setError('Options must all be unique');
      return;
    }

    setLoading(true);

    try {
      const res = await api.createPoll(question.trim(), trimmedOptions, expiresIn);
      setCreatedPoll(res.poll);
      showToast('Poll created successfully!', 'success');
    } catch (err) {
      setError(err.message || 'Failed to create poll');
    } finally {
      setLoading(false);
    }
  };

  const copyShareLink = () => {
    if (!createdPoll) return;
    const shareUrl = `${window.location.origin}/poll/${createdPoll.share_code}`;
    navigator.clipboard.writeText(shareUrl).then(() => {
      showToast('Share link copied to clipboard!', 'success');
    });
  };

  if (createdPoll) {
    const shareUrl = `${window.location.origin}/poll/${createdPoll.share_code}`;
    return (
      <div style={{ maxWidth: '600px', margin: '3rem auto 0' }}>
        <div className="glass-card" style={{ padding: '3rem 2.5rem', textAlign: 'center' }}>
          <div style={{ width: '64px', height: '64px', borderRadius: '50%', background: 'rgba(16, 185, 129, 0.15)', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 1.5rem', color: '#10b981' }}>
            <CheckCircle2 size={36} />
          </div>

          <h2>Poll Created Successfully!</h2>
          <p style={{ color: 'var(--text-muted)', fontSize: '0.95rem', margin: '0.5rem 0 2rem' }}>
            Share this link with your audience so they can cast their live votes:
          </p>

          <div style={{ display: 'flex', gap: '0.5rem', background: 'rgba(10, 15, 26, 0.8)', padding: '0.6rem 0.85rem', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-subtle)', alignItems: 'center', marginBottom: '1.5rem' }}>
            <input
              type="text"
              readOnly
              value={shareUrl}
              style={{ background: 'transparent', border: 'none', color: 'var(--text-main)', width: '100%', outline: 'none', fontSize: '0.95rem', fontWeight: 600 }}
            />
            <button onClick={copyShareLink} className="btn btn-primary btn-sm" style={{ flexShrink: 0 }}>
              <Copy size={16} />
              <span>Copy</span>
            </button>
          </div>

          <div style={{ display: 'flex', gap: '1rem', justifyContent: 'center', flexWrap: 'wrap' }}>
            <button
              onClick={() => navigate(`/poll/${createdPoll.id}/results`)}
              className="btn btn-primary btn-lg"
            >
              <span>Go to Live Results</span>
              <ArrowRight size={18} />
            </button>

            <button
              onClick={() => navigate(`/poll/${createdPoll.share_code}`)}
              className="btn btn-secondary btn-lg"
              target="_blank"
            >
              <span>Open Vote Page</span>
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div style={{ maxWidth: '640px', margin: '1rem auto 0' }}>
      <div className="glass-card" style={{ padding: '2.5rem 2rem' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '1.75rem' }}>
          <div className="brand-icon" style={{ width: '42px', height: '42px' }}>
            <Sparkles size={22} />
          </div>
          <div>
            <h2>Create a Live Poll</h2>
            <p style={{ color: 'var(--text-muted)', fontSize: '0.9rem' }}>
              Ask your question and collect votes with instant real-time visualization
            </p>
          </div>
        </div>

        {error && (
          <div className="alert alert-danger">
            <AlertCircle size={18} style={{ flexShrink: 0 }} />
            <span>{error}</span>
          </div>
        )}

        <form onSubmit={handleSubmit}>
          {/* Question */}
          <div className="form-group">
            <label className="form-label">Poll Question *</label>
            <textarea
              id="poll-question-input"
              required
              rows={3}
              className="form-textarea"
              placeholder="e.g., What is your favorite programming language?"
              value={question}
              onChange={(e) => setQuestion(e.target.value)}
            />
            <div style={{ display: 'flex', justifyContent: 'flex-end', fontSize: '0.78rem', color: 'var(--text-dim)', marginTop: '0.35rem' }}>
              {question.length}/500
            </div>
          </div>

          {/* Options */}
          <div className="form-group">
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem' }}>
              <label className="form-label" style={{ margin: 0 }}>Options ({options.length}/10) *</label>
              {options.length < 10 && (
                <button
                  type="button"
                  onClick={handleAddOption}
                  className="btn btn-outline btn-sm"
                  style={{ color: 'var(--primary)' }}
                >
                  <Plus size={14} />
                  <span>Add Option</span>
                </button>
              )}
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
              {options.map((opt, idx) => (
                <div key={idx} style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <span style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-dim)', width: '22px' }}>
                    {idx + 1}.
                  </span>
                  <input
                    type="text"
                    required
                    className="form-input"
                    placeholder={`Option ${idx + 1}`}
                    value={opt}
                    onChange={(e) => handleOptionChange(idx, e.target.value)}
                  />
                  {options.length > 2 && (
                    <button
                      type="button"
                      onClick={() => handleRemoveOption(idx)}
                      className="btn btn-outline btn-sm btn-icon"
                      style={{ color: '#ef4444' }}
                      title="Remove option"
                    >
                      <Trash2 size={16} />
                    </button>
                  )}
                </div>
              ))}
            </div>
          </div>

          {/* Expiration */}
          <div className="form-group">
            <label className="form-label" style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
              <Clock size={16} />
              <span>Poll Expiration</span>
            </label>
            <select
              id="poll-expiration-select"
              className="form-select"
              value={expiresIn}
              onChange={(e) => setExpiresIn(e.target.value)}
            >
              <option value="never">Never expires (Keep active indefinitely)</option>
              <option value="1h">1 Hour</option>
              <option value="1d">1 Day (24 Hours)</option>
              <option value="7d">7 Days</option>
            </select>
          </div>

          {/* Submit */}
          <button
            id="create-poll-submit-btn"
            type="submit"
            disabled={loading}
            className="btn btn-primary btn-lg"
            style={{ width: '100%', marginTop: '1rem' }}
          >
            {loading ? <div className="spinner" /> : 'Publish Live Poll'}
          </button>
        </form>
      </div>
    </div>
  );
}
