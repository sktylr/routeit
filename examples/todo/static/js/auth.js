const API_BASE = 'http://localhost:8080/auth';

/**
 * Logs a user in.
 * Handles fetch, token storage, and redirect.
 * Returns the HTTP status code to the caller.
 *
 * @param {string} email
 * @param {string} password
 * @returns {Promise<number>} HTTP status code
 */
export async function login(email, password) {
  try {
    const res = await fetch(`${API_BASE}/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password })
    });

    if (res.ok) {
      const data = await res.json();
      localStorage.setItem('access_token', data.access_token);
      localStorage.setItem('refresh_token', data.refresh_token);
      window.location.href = '/';
    }

    return res.status;
  } catch (err) {
    console.error("Network or unexpected error during login:", err);
    return -1; 
  }
}
