import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { 
  BarChart3, 
  Zap, 
  Database, 
  Radio, 
  ArrowRight, 
  Vote, 
  ShieldCheck, 
  Sparkles,
  Layers
} from 'lucide-react';

export default function Home({ user }) {
  const [code, setCode] = useState('');
  const navigate = useNavigate();

  const handleJoinPoll = (e) => {
    e.preventDefault();
    if (code.trim()) {
      navigate(`/poll/${code.trim().toUpperCase()}`);
    }
  };

  return (
    <div>
      {/* Hero Section */}
      <section style={{ textAlign: 'center', padding: '3.5rem 1rem 4rem' }}>
        <div className="badge-live" style={{ marginBottom: '1.5rem' }}>
          <div className="pulse-dot" />
          <span>REAL-TIME LIVE POLLING PLATFORM</span>
        </div>

        <h1 style={{ fontSize: 'clamp(2.4rem, 5vw, 3.8rem)', fontWeight: 800, lineHeight: 1.15, maxWidth: '850px', margin: '0 auto 1.5rem', letterSpacing: '-0.03em' }}>
          Instant Live Polls with{' '}
          <span style={{ background: 'linear-gradient(135deg, #6366f1, #06b6d4)', WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent' }}>
            Zero-Refresh
          </span>{' '}
          Real-Time Updates
        </h1>

        <p style={{ fontSize: '1.15rem', color: 'var(--text-muted)', maxWidth: '650px', margin: '0 auto 2.5rem', lineHeight: 1.6 }}>
          Create interactive polls, share unique instant links with your audience, and watch vote counters and animated charts stream in real-time.
        </p>

        {/* Quick Action Buttons */}
        <div style={{ display: 'flex', justifyContent: 'center', gap: '1rem', flexWrap: 'wrap', marginBottom: '3rem' }}>
          <Link
            to={user ? '/create-poll' : '/signup'}
            className="btn btn-primary btn-lg"
          >
            <Sparkles size={18} />
            <span>{user ? 'Create a Poll Now' : 'Get Started Free'}</span>
            <ArrowRight size={18} />
          </Link>

          {user && (
            <Link to="/dashboard" className="btn btn-secondary btn-lg">
              <Layers size={18} />
              <span>Go to Dashboard</span>
            </Link>
          )}
        </div>

        {/* Audience Quick Join Box */}
        <div className="glass-card" style={{ maxWidth: '460px', margin: '0 auto', padding: '1.75rem' }}>
          <div style={{ fontSize: '0.9rem', fontWeight: 600, color: 'var(--text-muted)', marginBottom: '0.85rem', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.4rem' }}>
            <Vote size={18} color="var(--secondary)" />
            <span>Have a poll code? Enter it below:</span>
          </div>

          <form onSubmit={handleJoinPoll} style={{ display: 'flex', gap: '0.5rem' }}>
            <input
              type="text"
              required
              placeholder="e.g. ABC123"
              value={code}
              onChange={(e) => setCode(e.target.value.toUpperCase())}
              className="form-input"
              style={{ textTransform: 'uppercase', letterSpacing: '0.1em', fontWeight: 700, textAlign: 'center', fontSize: '1.05rem' }}
              maxLength={8}
            />
            <button type="submit" className="btn btn-primary" style={{ flexShrink: 0 }}>
              <span>Join Poll</span>
            </button>
          </form>
        </div>
      </section>

      {/* Tech Architecture Features */}
      <section style={{ marginTop: '2rem' }}>
        <div style={{ textAlign: 'center', marginBottom: '2.5rem' }}>
          <h2 style={{ fontSize: '1.85rem' }}>Powered by 4 Core Technologies</h2>
          <p style={{ color: 'var(--text-muted)', fontSize: '0.95rem', marginTop: '0.35rem' }}>
            Built strictly for the GUVI Developer Internship assignment specifications
          </p>
        </div>

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', gap: '1.5rem' }}>
          {/* React */}
          <div className="glass-card" style={{ padding: '2rem 1.5rem' }}>
            <div style={{ width: '48px', height: '48px', borderRadius: '12px', background: 'rgba(6, 182, 212, 0.15)', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#06b6d4', marginBottom: '1.25rem' }}>
              <BarChart3 size={24} />
            </div>
            <h3 style={{ fontSize: '1.2rem', marginBottom: '0.5rem' }}>React 18 UI</h3>
            <p style={{ color: 'var(--text-muted)', fontSize: '0.9rem', lineHeight: 1.6 }}>
              Dynamic, responsive user interface featuring animated progress bars, live indicators, and seamless SPA navigation without page refreshes.
            </p>
          </div>

          {/* Go + Gin */}
          <div className="glass-card" style={{ padding: '2rem 1.5rem' }}>
            <div style={{ width: '48px', height: '48px', borderRadius: '12px', background: 'rgba(99, 102, 241, 0.15)', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#6366f1', marginBottom: '1.25rem' }}>
              <Zap size={24} />
            </div>
            <h3 style={{ fontSize: '1.2rem', marginBottom: '0.5rem' }}>Go + Gin API</h3>
            <p style={{ color: 'var(--text-muted)', fontSize: '0.9rem', lineHeight: 1.6 }}>
              High-throughput backend performing strict input validation, bcrypt password hashing, JWT authentication, and Gorilla WebSocket fanout.
            </p>
          </div>

          {/* MongoDB */}
          <div className="glass-card" style={{ padding: '2rem 1.5rem' }}>
            <div style={{ width: '48px', height: '48px', borderRadius: '12px', background: 'rgba(16, 185, 129, 0.15)', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#10b981', marginBottom: '1.25rem' }}>
              <Database size={24} />
            </div>
            <h3 style={{ fontSize: '1.2rem', marginBottom: '0.5rem' }}>MongoDB Persistence</h3>
            <p style={{ color: 'var(--text-muted)', fontSize: '0.9rem', lineHeight: 1.6 }}>
              Reliable persistent storage for user profiles, poll metadata, and immutable vote audit records with compound uniqueness for deduplication.
            </p>
          </div>

          {/* Redis */}
          <div className="glass-card" style={{ padding: '2rem 1.5rem' }}>
            <div style={{ width: '48px', height: '48px', borderRadius: '12px', background: 'rgba(239, 68, 68, 0.15)', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#ef4444', marginBottom: '1.25rem' }}>
              <Radio size={24} />
            </div>
            <h3 style={{ fontSize: '1.2rem', marginBottom: '0.5rem' }}>Redis Pub/Sub & Counters</h3>
            <p style={{ color: 'var(--text-muted)', fontSize: '0.9rem', lineHeight: 1.6 }}>
              Drives real-time counting via atomic <code style={{ color: '#f87171' }}>HINCRBY</code> and dispatches <code style={{ color: '#f87171' }}>VOTE_UPDATED</code> events across distributed nodes.
            </p>
          </div>
        </div>
      </section>

      {/* Real-Time Architecture Diagram Card */}
      <section className="glass-card" style={{ marginTop: '3.5rem', padding: '2.5rem 2rem' }}>
        <h3 style={{ fontSize: '1.35rem', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.6rem' }}>
          <ShieldCheck size={22} color="var(--secondary)" />
          <span>Real-Time Flow Architecture</span>
        </h3>

        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.85rem', color: 'var(--text-muted)', fontSize: '0.92rem', lineHeight: 1.7 }}>
          <p>
            1. <strong>Audience Votes:</strong> Voter submits an option via <code style={{ color: 'var(--secondary)' }}>POST /api/polls/:id/vote</code> with client voter token.
          </p>
          <p>
            2. <strong>Validation & Mongo Storage:</strong> Go backend checks poll status and deduplication, saving the vote audit log in MongoDB.
          </p>
          <p>
            3. <strong>Redis Atomic Increment:</strong> Redis executes <code style={{ color: '#f87171' }}>HINCRBY poll:&lt;id&gt;:votes &lt;option_id&gt; 1</code> in microseconds.
          </p>
          <p>
            4. <strong>Redis Pub/Sub Event:</strong> Redis publishes the <code style={{ color: '#f87171' }}>VOTE_UPDATED</code> payload to channel <code style={{ color: '#f87171' }}>poll:&lt;id&gt;:updates</code>.
          </p>
          <p>
            5. <strong>WebSocket Broadcast:</strong> Go WebSocket Hub broadcasts the update to all connected browser windows watching that poll — instantly animating the results without page refresh!
          </p>
        </div>
      </section>
    </div>
  );
}
