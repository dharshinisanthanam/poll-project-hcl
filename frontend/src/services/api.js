import { getVoterToken } from './voter';

const LIVE_BACKEND_ORIGIN = 'https://poll-project-hcl-1.onrender.com';
const rawApiUrl = (import.meta.env.VITE_API_URL || '').trim();
let API_BASE = '/api';

if (rawApiUrl && rawApiUrl !== '/api') {
  if (typeof window !== 'undefined' && window.location.hostname !== 'localhost' && window.location.hostname !== '127.0.0.1' && rawApiUrl.includes('localhost')) {
    API_BASE = `${LIVE_BACKEND_ORIGIN}/api`;
  } else {
    let cleaned = rawApiUrl.replace(/\/+$/, '');
    if (cleaned.startsWith('http') && !cleaned.endsWith('/api')) {
      cleaned += '/api';
    }
    API_BASE = cleaned;
  }
} else if (typeof window !== 'undefined' && window.location.hostname !== 'localhost' && window.location.hostname !== '127.0.0.1') {
  API_BASE = `${LIVE_BACKEND_ORIGIN}/api`;
}


function getAuthHeader() {
  const token = localStorage.getItem('auth_token');
  return token ? { Authorization: `Bearer ${token}` } : {};
}

async function request(endpoint, options = {}) {
  const url = `${API_BASE}${endpoint}`;
  const headers = {
    'Content-Type': 'application/json',
    ...getAuthHeader(),
    ...options.headers,
  };

  let response;
  try {
    response = await fetch(url, {
      ...options,
      headers,
    });
  } catch (networkErr) {
    throw new Error('Unable to reach backend service. Please ensure the backend server is active and accessible.');
  }

  const data = await response.json().catch(() => ({}));

  if (!response.ok) {
    let errorMsg = data.error || data.message;
    if (!errorMsg) {
      if (response.status === 405) {
        errorMsg = 'Backend endpoint method not allowed (HTTP 405). If running in the cloud, verify that the backend is active and VITE_API_URL is configured.';
      } else if (response.status === 404) {
        errorMsg = 'Backend API endpoint not found (HTTP 404). Please ensure the backend server is running.';
      } else if (response.status >= 500) {
        errorMsg = `Backend server error (HTTP ${response.status}). Please check backend service logs.`;
      } else {
        errorMsg = `HTTP Error ${response.status}`;
      }
    }
    const err = new Error(errorMsg);
    err.status = response.status;
    err.data = data;
    throw err;
  }

  return data;
}

export const api = {
  // Auth
  async signup(name, email, password) {
    return request('/auth/signup', {
      method: 'POST',
      body: JSON.stringify({ name, email, password }),
    });
  },

  async login(identifier, password) {
    return request('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email: identifier, identifier, password }),
    });
  },

  async getMe() {
    return request('/auth/me');
  },

  // Polls
  async createPoll(question, options, expiresIn = 'never') {
    return request('/polls', {
      method: 'POST',
      body: JSON.stringify({
        question,
        options,
        expires_in: expiresIn,
      }),
    });
  },

  async getUserPolls() {
    return request('/polls');
  },

  async getPollByShareCode(shareCode) {
    return request(`/polls/share/${shareCode}`);
  },

  async getPollById(id) {
    return request(`/polls/${id}`);
  },

  async updatePollStatus(id, status) {
    return request(`/polls/${id}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status }),
    });
  },

  async deletePoll(id) {
    return request(`/polls/${id}`, {
      method: 'DELETE',
    });
  },

  // Voting
  async castVote(pollIdOrShareCode, optionId) {
    const voterToken = getVoterToken();
    return request(`/polls/${pollIdOrShareCode}/vote`, {
      method: 'POST',
      body: JSON.stringify({
        option_id: optionId,
        voter_token: voterToken,
      }),
    });
  },

  async getResults(pollIdOrShareCode) {
    const voterToken = getVoterToken();
    return request(`/polls/${pollIdOrShareCode}/results?voter_token=${encodeURIComponent(voterToken)}`);
  },
};
