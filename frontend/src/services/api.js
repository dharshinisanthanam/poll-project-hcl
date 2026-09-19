import { getVoterToken } from './voter';

const rawApiUrl = (import.meta.env.VITE_API_URL || '').trim();
let API_BASE = '/api';

if (rawApiUrl) {
  if (typeof window !== 'undefined' && window.location.hostname !== 'localhost' && window.location.hostname !== '127.0.0.1' && rawApiUrl.includes('localhost')) {
    API_BASE = '/api';
  } else {
    let cleaned = rawApiUrl.replace(/\/+$/, '');
    // Ensure /api suffix exists if given an absolute URL pointing to root domain
    if (cleaned.startsWith('http') && !cleaned.endsWith('/api')) {
      cleaned += '/api';
    }
    API_BASE = cleaned;
  }
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

  const response = await fetch(url, {
    ...options,
    headers,
  });

  const data = await response.json().catch(() => ({}));

  if (!response.ok) {
    const errorMsg = data.error || data.message || `HTTP Error ${response.status}`;
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

  async login(email, password) {
    return request('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
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
