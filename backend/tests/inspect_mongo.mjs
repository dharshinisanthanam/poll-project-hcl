// Inspect MongoDB collections and documents
import http from 'http';

async function check() {
  const res = await fetch('http://localhost:8080/health');
  const health = await res.json();
  console.log('Backend Health:', health);

  // Fetch from backend
  const rootRes = await fetch('http://localhost:8080/');
  const root = await rootRes.json();
  console.log('Backend Root:', root);
}

check().catch(console.error);
