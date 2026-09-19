import React, { useState, useEffect, useCallback } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import Navbar from './components/Navbar';
import Toast from './components/Toast';
import Home from './pages/Home';
import Login from './pages/Login';
import Signup from './pages/Signup';
import Dashboard from './pages/Dashboard';
import CreatePoll from './pages/CreatePoll';
import PublicPoll from './pages/PublicPoll';
import PollResults from './pages/PollResults';
import { api } from './services/api';

function ProtectedRoute({ user, children }) {
  const token = localStorage.getItem('auth_token');
  if (!user && !token) {
    return <Navigate to="/login" replace />;
  }
  return children;
}

function PublicAuthRoute({ user, children }) {
  const token = localStorage.getItem('auth_token');
  if (user || token) {
    return <Navigate to="/dashboard" replace />;
  }
  return children;
}

export default function App() {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [toast, setToast] = useState(null);

  const showToast = useCallback((message, type = 'success') => {
    setToast({ message, type });
    setTimeout(() => {
      setToast(null);
    }, 4000);
  }, []);

  useEffect(() => {
    const token = localStorage.getItem('auth_token');
    if (token) {
      api.getMe()
        .then((res) => {
          setUser(res.user);
        })
        .catch(() => {
          localStorage.removeItem('auth_token');
          setUser(null);
        })
        .finally(() => {
          setLoading(false);
        });
    } else {
      setLoading(false);
    }
  }, []);

  const handleLoginSuccess = (userData) => {
    setUser(userData);
    showToast(`Welcome back, ${userData.name}!`, 'success');
  };

  const handleLogout = () => {
    localStorage.removeItem('auth_token');
    setUser(null);
    showToast('Logged out successfully', 'info');
  };

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <div className="spinner" style={{ width: '40px', height: '40px' }} />
      </div>
    );
  }

  return (
    <BrowserRouter>
      <div className="app-layout">
        <Navbar user={user} onLogout={handleLogout} />

        <main className="main-content">
          <Routes>
            <Route path="/" element={<Home user={user} />} />
            <Route 
              path="/login" 
              element={
                <PublicAuthRoute user={user}>
                  <Login onLoginSuccess={handleLoginSuccess} />
                </PublicAuthRoute>
              } 
            />
            <Route 
              path="/signup" 
              element={
                <PublicAuthRoute user={user}>
                  <Signup onLoginSuccess={handleLoginSuccess} />
                </PublicAuthRoute>
              } 
            />

            {/* Protected Creator Routes */}
            <Route 
              path="/dashboard" 
              element={
                <ProtectedRoute user={user}>
                  <Dashboard showToast={showToast} />
                </ProtectedRoute>
              } 
            />
            <Route 
              path="/create-poll" 
              element={
                <ProtectedRoute user={user}>
                  <CreatePoll showToast={showToast} />
                </ProtectedRoute>
              } 
            />

            {/* Public Audience & Results Routes (supports both /p/ and /poll/ paths) */}
            <Route 
              path="/p/:shareCode" 
              element={<PublicPoll showToast={showToast} />} 
            />
            <Route 
              path="/poll/:shareCode" 
              element={<PublicPoll showToast={showToast} />} 
            />
            <Route 
              path="/p/:id/results" 
              element={<PollResults showToast={showToast} />} 
            />
            <Route 
              path="/poll/:id/results" 
              element={<PollResults showToast={showToast} />} 
            />

            {/* Fallback */}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </main>

        {toast && (
          <Toast 
            message={toast.message} 
            type={toast.type} 
            onClose={() => setToast(null)} 
          />
        )}
      </div>
    </BrowserRouter>
  );
}
