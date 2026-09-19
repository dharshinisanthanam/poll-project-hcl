import React from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { BarChart3, PlusCircle, LayoutDashboard, LogIn, UserPlus, LogOut } from 'lucide-react';

export default function Navbar({ user, onLogout }) {
  const navigate = useNavigate();
  const location = useLocation();

  const handleLogout = () => {
    onLogout();
    navigate('/login');
  };

  const isAuthed = !!(user || localStorage.getItem('auth_token'));

  return (
    <header className="navbar">
      <div className="nav-container">
        <Link to="/" className="nav-brand">
          <div className="brand-icon">
            <BarChart3 size={22} />
          </div>
          <span>LivePoll</span>
          <span className="brand-badge">Real-Time</span>
        </Link>

        <nav className="nav-links">
          {isAuthed ? (
            <>
              <Link 
                to="/dashboard" 
                className={`nav-link ${location.pathname === '/dashboard' ? 'active' : ''}`}
              >
                <LayoutDashboard size={18} />
                <span>Dashboard</span>
              </Link>
              <Link 
                to="/create-poll" 
                className={`btn btn-primary btn-sm`}
              >
                <PlusCircle size={16} />
                <span>Create Poll</span>
              </Link>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginLeft: '0.5rem' }}>
                <span style={{ fontSize: '0.88rem', color: 'var(--text-muted)' }}>
                  Hi, <strong style={{ color: 'var(--text-main)' }}>{user?.name || 'Creator'}</strong>
                </span>
                <button 
                  onClick={handleLogout} 
                  className="btn btn-outline btn-sm"
                  title="Logout"
                >
                  <LogOut size={16} />
                  <span>Logout</span>
                </button>
              </div>
            </>
          ) : (
            <>
              <Link 
                to="/login" 
                className={`nav-link ${location.pathname === '/login' ? 'active' : ''}`}
              >
                <LogIn size={18} />
                <span>Login</span>
              </Link>
              <Link 
                to="/signup" 
                className="btn btn-primary btn-sm"
              >
                <UserPlus size={16} />
                <span>Sign Up</span>
              </Link>
            </>
          )}
        </nav>
      </div>
    </header>
  );
}
