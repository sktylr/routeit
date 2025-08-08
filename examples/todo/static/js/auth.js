const API_BASE = 'http://localhost:8080/auth';

/**
 * Logs a user in.
 * Handles fetch, token storage, and redirect.
 * @param {string} email
 * @param {string} password
 * @returns {Promise<number>} HTTP status code or -1 on network error
 */
export async function login(email, password) {
  return makeRequest("/login", { email, password })
}

/**
 * Registers a new user.
 * Handles fetch, token storage, and redirect.
 * @param {string} email
 * @param {string} password
 * @returns {Promise<number>} HTTP status code or -1 on network error
 */
export async function register(name, email, password, confirmPassword) {
  return makeRequest("/register", { name, email, password, confirm_password: confirmPassword })
}

/**
 * Private helper for auth-related API requests.
 * @param {string} endpoint - e.g. "/login" or "/register"
 * @param {object} body - Request payload
 * @returns {Promise<number>} HTTP status code or -1 on network error
 */
async function makeRequest(endpoint, body) {
  try {
    const res = await fetch(`${API_BASE}${endpoint}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body)
    });

    if (res.ok) {
      const data = await res.json();
      localStorage.setItem('access_token', data.access_token);
      localStorage.setItem('refresh_token', data.refresh_token);
      window.location.href = '/';
    }

    return res.status;
  } catch (err) {
    console.error(`Auth request error [${endpoint}]:`, err);
    return -1;
  }
}
