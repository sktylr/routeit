import { API_BASE } from "./api.js";

/**
 * Logs a user in.
 * Handles fetch, token storage, and redirect.
 * @param {string} email
 * @param {string} password
 * @returns {Promise<number>} HTTP status code or -1 on network error
 */
export async function login(email, password) {
  return makeRequest("/login", { email, password }).then(res => res.status)
}

/**
 * Registers a new user.
 * Handles fetch, token storage, and redirect.
 * @param {string} email
 * @param {string} password
 * @returns {Promise<number>} HTTP status code or -1 on network error
 */
export async function register(name, email, password, confirmPassword) {
  return makeRequest("/register", {
    name,
    email,
    password,
    confirm_password: confirmPassword
  }).then(res => res.status)
}

/**
 * Refreshes the access token using the stored refresh token.
 * @returns {Promise<string>} New access token or "" if refresh fails
 */
export async function refreshToken() {
  const storedRefresh = localStorage.getItem('refresh_token');
  if (!storedRefresh) {
    console.warn('No refresh token found in storage.');
    return '';
  }

  return makeRequest(
    "/refresh",
    { refresh_token: storedRefresh },
    { redirect: false }
  ).then(res => res.accessToken)
}

/**
 * Private helper for auth-related API requests.
 * Handles fetch, token storage, and optional redirect.
 *
 * @param {string} endpoint - e.g. "/login", "/register", "/refresh"
 * @param {object} body - Request payload
 * @param {object} [options] - Optional config
 * @param {boolean} [options.redirect=true] - Whether to redirect to "/" on success
 * @returns {Promise<{ status: number, accessToken: string, refreshToken: string }>}
 */
async function makeRequest(endpoint, body, { redirect = true } = {}) {
  try {
    const res = await fetch(`${API_BASE}/auth${endpoint}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body)
    });

    let accessToken = '';
    let refreshToken = '';

    if (res.ok) {
      const data = await res.json();
      accessToken = data.access_token || '';
      refreshToken = data.refresh_token || '';

      if (accessToken) localStorage.setItem('access_token', accessToken);
      if (refreshToken) localStorage.setItem('refresh_token', refreshToken);

      if (redirect) {
        window.location.href = '/';
      }
    }

    return { status: res.status, accessToken, refreshToken };
  } catch (err) {
    console.error(`Auth request error [${endpoint}]:`, err);
    return { status: -1, accessToken: '', refreshToken: '' };
  }
}
