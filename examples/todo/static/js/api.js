import { refreshToken } from "./auth.js";

export const API_BASE = 'http://localhost:8080';

/**
 * Fetches paginated lists.
 * @param {number} page
 * @param {number} pageSize
 * @returns {Promise<{ status: number, data: any }>}
 */
export async function getLists(page = 1, pageSize = 10) {
  return authorizedRequest(`/lists?page=${page}&page_size=${pageSize}`, {
    method: 'GET'
  });
}

/**
 * Fetches a list by id.
 * @param {string} id the list's id
 * @returns {Promise<{ status: number, data: any }>}
 */
export async function getList(id) {
	return authorizedRequest(`/lists/${id}`, { method: 'GET' });
}

/**
 * Fetches paginated items for a list.
 * @param {string} id the list's id
 * @param {number} page the page of results to query
 * @param {number} pageSize the page size to query
 * @returns {Promise<{ status: number, data: any }>}
 */
export async function getItemsForList(id, page = 1, pageSize = 10) {
	return authorizedRequest(`/lists/${id}/items?page=${page}&page_size=${pageSize}`, {
		method: 'GET'
	});
}

/**
 * Create a new list
 * @param {string} name
 * @param {string} description
 * @returns {Promise<{ status: number, data: any }>}
 */
export async function createList(name, description = "") {
	const body = JSON.stringify({ name, description })
	return authorizedRequest("/lists", {
		method: 'POST',
		body,
	});
}

/**
 * Create a new item for a list
 * @param {string} listId
 * @param {string} name
 * @returns {Promise<{ status: number, data: any }>}
 */
export async function createItemForList(listId, name) {
	const body = JSON.stringify({ name })
	return authorizedRequest(`/lists/${listId}/items`, {
		method: 'POST',
		body,
	});
}

/**
 * Update a list
 * @param {string} id
 * @param {string} name
 * @param {string} description
 * @returns {Promise<{ status: number, data: any }>}
 */
export async function updateList(id, name, description) {
	const body = JSON.stringify({ name, description })
	return authorizedRequest(`/lists/${id}`, {
		method: 'PUT',
		body,
	});
}

/**
 * Update an item for a list
 * @param {string} listId
 * @param {string} itemId
 * @param {string} name
 * @returns {Promise<{ status: number, data: any }>}
 */
export async function updateItem(listId, itemId, name) {
	const body = JSON.stringify({ name })
	return authorizedRequest(`/lists/${listId}/items/${itemId}`, {
		method: 'PUT',
		body,
	});
}

/**
 * Mark an item as completed
 * @param {string} listId
 * @param {string} itemId
 * @returns {Promise<{ status: number, data: any }>}
 */
export async function markItemCompleted(listId, itemId) {
  const body = JSON.stringify({ status: 'COMPLETED' })
	return authorizedRequest(`/lists/${listId}/items/${itemId}`, {
		method: 'PATCH',
		body,
	});
}

/**
 * Mark an item as pending (incomplete)
 * @param {string} listId
 * @param {string} itemId
 * @returns {Promise<{ status: number, data: any }>}
 */
export async function markItemPending(listId, itemId) {
  const body = JSON.stringify({ status: 'PENDING' })
	return authorizedRequest(`/lists/${listId}/items/${itemId}`, {
		method: 'PATCH',
		body,
	});
}

/**
 * Deletes a list.
 * @param {string} id
 * @returns {Promise<{ status: number, data: any }>}
 */
export async function deleteList(id) {
  return authorizedRequest(`/lists/${id}`, { method: 'DELETE '})
}

/**
 * Deletes an item from a list.
 * @param {string} listId
 * @param {string} itemId
 * @returns {Promise<{ status: number, data: any }>}
 */
export async function deleteItem(listId, itemId) {
  return authorizedRequest(`/lists/${listId}/items/${itemId}`, { method: 'DELETE' })
}

/**
 * Shared helper for authenticated API requests.
 * Handles attaching JWT, retrying once on 401 with refresh, and redirecting to login if needed.
 *
 * @param {string} endpoint - API endpoint (relative to API_BASE)
 * @param {RequestInit} options - Fetch options
 * @returns {Promise<{ status: number, data: any }>}
 */
async function authorizedRequest(endpoint, options = {}) {
  const accessToken = localStorage.getItem('access_token');

  const headers = {
    ...options.headers,
    'Authorization': `Bearer ${accessToken}`,
    'Content-Type': 'application/json',
  };

  let res = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    headers
  });

  if (res.status === 401) {
    const newAccessToken = await refreshToken();

    if (!newAccessToken) {
      // No valid refresh → force logout
      window.location.href = '/login';
      return { status: 401, data: null };
    }

    // Retry with new access token
    const retryHeaders = {
      ...options.headers,
      'Authorization': `Bearer ${newAccessToken}`,
      'Content-Type': 'application/json',
    };

    res = await fetch(`${API_BASE}${endpoint}`, {
      ...options,
      headers: retryHeaders
    });
  }

  if (res.status === 401) {
    window.location.href = '/login';
    return { status: 401, data: null };
  }

  let data = null;
  try {
    data = await res.json();
  } catch {
    // ignore JSON parse errors (e.g. empty body)
  }

  return { status: res.status, data };
}
