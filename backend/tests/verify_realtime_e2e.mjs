// Real-Time Multi-Client Verification Script for Live Polling Tool
// Simulates multiple concurrent audience browsers and verifies zero-refresh real-time updates via WebSockets & Redis Pub/Sub

import http from 'http';

async function request(url, options = {}, body = null) {
  return new Promise((resolve, reject) => {
    const u = new URL(url);
    const reqOptions = {
      hostname: u.hostname,
      port: u.port,
      path: u.pathname + u.search,
      method: options.method || 'GET',
      headers: options.headers || {},
    };

    const req = http.request(reqOptions, (res) => {
      let data = '';
      res.on('data', (chunk) => data += chunk);
      res.on('end', () => {
        try {
          const parsed = JSON.parse(data);
          resolve({ status: res.statusCode, body: parsed });
        } catch {
          resolve({ status: res.statusCode, body: data });
        }
      });
    });

    req.on('error', reject);
    if (body) {
      req.write(typeof body === 'string' ? body : JSON.stringify(body));
    }
    req.end();
  });
}

// Minimal WebSocket client using standard Node.js
async function connectWebSocket(pollId) {
  const WebSocket = (await import('ws')).default;
  return new Promise((resolve, reject) => {
    const ws = new WebSocket(`ws://localhost:8080/ws/polls/${pollId}`);
    ws.on('open', () => resolve(ws));
    ws.on('error', reject);
  });
}

async function runVerification() {
  console.log('====================================================');
  console.log('  Live Polling Tool — Real-Time Multi-Browser E2E   ');
  console.log('====================================================\n');

  // Step 1: User Signup
  const email = `creator_${Date.now()}@guvi.com`;
  console.log(`[1] Registering Poll Creator (${email})...`);
  const signupRes = await request('http://localhost:8080/api/auth/signup', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
  }, {
    name: 'Dharshini Intern',
    email,
    password: 'securePassword123!',
  });

  if (signupRes.status !== 201) {
    throw new Error(`Signup failed with status ${signupRes.status}: ${JSON.stringify(signupRes.body)}`);
  }
  const token = signupRes.body.user.token;
  console.log('    ✓ Signup successful, received JWT token.\n');

  // Step 2: Create Poll
  console.log('[2] Creating Live Poll: "What is your favorite programming language?"...');
  const pollRes = await request('http://localhost:8080/api/polls', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
  }, {
    question: 'What is your favorite programming language?',
    options: ['Python', 'JavaScript', 'Go', 'Java'],
    expires_in: 'never',
  });

  if (pollRes.status !== 201) {
    throw new Error(`Create poll failed: ${JSON.stringify(pollRes.body)}`);
  }
  const poll = pollRes.body.poll;
  const pollId = poll.id;
  const shareCode = poll.share_code;
  console.log(`    ✓ Poll created! ID: ${pollId} | Share Code: ${shareCode}\n`);

  // Step 3: Connect Browser 1 & Browser 2 via WebSockets
  console.log('[3] Connecting Browser 1 (Audience Live Results View) via WebSocket...');
  const browser1Messages = [];
  const ws1 = await connectWebSocket(pollId);
  ws1.on('message', (msg) => {
    const parsed = JSON.parse(msg.toString());
    browser1Messages.push(parsed);
    console.log(`    [Browser 1 Received Live Update] Event: ${parsed.event} | Total Votes: ${parsed.payload.total_votes}`);
  });
  console.log('    ✓ Browser 1 WebSocket connected.\n');

  console.log('[4] Connecting Browser 2 (Audience 2 Live Results View) via WebSocket...');
  const browser2Messages = [];
  const ws2 = await connectWebSocket(pollId);
  ws2.on('message', (msg) => {
    const parsed = JSON.parse(msg.toString());
    browser2Messages.push(parsed);
    console.log(`    [Browser 2 Received Live Update] Event: ${parsed.event} | Total Votes: ${parsed.payload.total_votes}`);
  });
  console.log('    ✓ Browser 2 WebSocket connected.\n');

  // Step 4: Audience Browser 3 Votes for "Go" (opt_3)
  const browser3Token = 'browser_3_session_token_' + Date.now();
  console.log('[5] Audience Browser 3 casting vote for "Go" (opt_3)...');
  const vote1Res = await request(`http://localhost:8080/api/polls/${shareCode}/vote`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
  }, {
    option_id: 'opt_3',
    voter_token: browser3Token,
  });

  if (vote1Res.status !== 200) {
    throw new Error(`Vote 1 failed: ${JSON.stringify(vote1Res.body)}`);
  }
  console.log('    ✓ Vote 1 accepted by backend.\n');

  // Wait 600ms for WebSocket push
  await new Promise(r => setTimeout(r, 600));

  // Verify Browser 1 and Browser 2 received the update
  if (browser1Messages.length < 1 || browser2Messages.length < 1) {
    throw new Error(`Failed to receive real-time update on both browsers! Browser1=${browser1Messages.length}, Browser2=${browser2Messages.length}`);
  }
  const update1 = browser1Messages[0];
  if (update1.payload.total_votes !== 1 || update1.payload.options.find(o => o.id === 'opt_3').votes !== 1) {
    throw new Error('Vote count did not match expected count of 1 for Go!');
  }
  console.log('    ✓ VERIFIED: Browser 1 and Browser 2 updated live without refreshing!\n');

  // Step 5: Audience Browser 4 Votes for "Python" (opt_1)
  console.log('[6] Audience Browser 4 casting vote for "Python" (opt_1)...');
  const vote2Res = await request(`http://localhost:8080/api/polls/${shareCode}/vote`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
  }, {
    option_id: 'opt_1',
    voter_token: 'browser_4_session_token_' + Date.now(),
  });

  if (vote2Res.status !== 200) {
    throw new Error(`Vote 2 failed: ${JSON.stringify(vote2Res.body)}`);
  }
  console.log('    ✓ Vote 2 accepted by backend.\n');

  await new Promise(r => setTimeout(r, 600));

  if (browser1Messages.length < 2 || browser2Messages.length < 2) {
    throw new Error('Failed to receive second real-time update on both browsers!');
  }
  const update2 = browser1Messages[1];
  console.log(`    ✓ VERIFIED: Browser 1 and Browser 2 both updated to Total Votes = ${update2.payload.total_votes} in real-time!\n`);

  // Step 6: Duplicate Vote Prevention Test
  console.log('[7] Testing duplicate voting rejection from Browser 3...');
  const dupRes = await request(`http://localhost:8080/api/polls/${shareCode}/vote`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
  }, {
    option_id: 'opt_3',
    voter_token: browser3Token, // exact same token
  });

  if (dupRes.status === 409) {
    console.log('    ✓ VERIFIED: Duplicate vote correctly rejected with 409 Conflict!\n');
  } else {
    throw new Error(`Expected 409 Conflict on duplicate vote, got ${dupRes.status}`);
  }

  // Cleanup
  ws1.close();
  ws2.close();

  console.log('====================================================');
  console.log('  ALL REAL-TIME END-TO-END VERIFICATIONS PASSED!    ');
  console.log('====================================================');
}

runVerification().catch((err) => {
  console.error('\n❌ Verification failed:', err);
  process.exit(1);
});
