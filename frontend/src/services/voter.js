// Manages a unique client voter token in localStorage for duplicate vote prevention

export function getVoterToken() {
  let token = localStorage.getItem('poll_voter_token');
  if (!token) {
    token = 'voter_' + Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15) + '_' + Date.now();
    localStorage.setItem('poll_voter_token', token);
  }
  return token;
}
