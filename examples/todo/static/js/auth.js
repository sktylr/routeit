const API_BASE = 'http://localhost:8080/auth';

/**
 * Logs a user in.
 * Handles fetch, token storage, and redirect.
 * @param {string} email
 * @param {string} password
 * @returns {Promise<number>} HTTP status code or -1 on network error
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
    console.error("Login error:", err);
    return -1;
  }
}

/**
 * Registers a new user.
 * Handles fetch, token storage, and redirect.
 * @param {string} email
 * @param {string} password
 * @returns {Promise<number>} HTTP status code or -1 on network error
 */
export async function register(name, email, password, confirmPassword) {
  try {
    const res = await fetch(`${API_BASE}/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, email, password, confirm_password: confirmPassword })
    });

    if (res.ok) {
      const data = await res.json();
      localStorage.setItem('access_token', data.access_token);
      localStorage.setItem('refresh_token', data.refresh_token);
      window.location.href = '/';
    }

    return res.status;
  } catch (err) {
    console.error("Register error:", err);
    return -1;
  }
}
