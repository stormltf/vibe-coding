/**
 * API Client for Vibe Coding Backend
 */

const API = {
    baseURL: '/api/v1',

    /**
     * Get stored token
     */
    getToken() {
        return localStorage.getItem('token');
    },

    /**
     * Set token
     */
    setToken(token) {
        localStorage.setItem('token', token);
    },

    /**
     * Remove token
     */
    removeToken() {
        localStorage.removeItem('token');
    },

    /**
     * Make HTTP request
     */
    async request(endpoint, options = {}) {
        const url = `${this.baseURL}${endpoint}`;
        const token = this.getToken();

        const config = {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers,
            },
            ...options,
        };

        if (token) {
            config.headers['Authorization'] = `Bearer ${token}`;
        }

        try {
            const response = await fetch(url, config);
            const data = await response.json();

            if (data.code !== 0) {
                throw new APIError(data.code, data.message);
            }

            return data.data;
        } catch (error) {
            if (error instanceof APIError) {
                throw error;
            }
            throw new APIError(-1, error.message || 'Network error');
        }
    },

    /**
     * GET request
     */
    get(endpoint, params = {}) {
        const query = new URLSearchParams(params).toString();
        const url = query ? `${endpoint}?${query}` : endpoint;
        return this.request(url, { method: 'GET' });
    },

    /**
     * POST request
     */
    post(endpoint, body = {}) {
        return this.request(endpoint, {
            method: 'POST',
            body: JSON.stringify(body),
        });
    },

    /**
     * PUT request
     */
    put(endpoint, body = {}) {
        return this.request(endpoint, {
            method: 'PUT',
            body: JSON.stringify(body),
        });
    },

    /**
     * DELETE request
     */
    delete(endpoint, body = {}) {
        return this.request(endpoint, {
            method: 'DELETE',
            body: JSON.stringify(body),
        });
    },

    // ============================================
    // Auth Endpoints
    // ============================================

    /**
     * Register new user
     */
    register(name, email, password) {
        return this.post('/auth/register', { name, email, password });
    },

    /**
     * Login user
     */
    login(email, password) {
        return this.post('/auth/login', { email, password });
    },

    /**
     * Logout user
     */
    logout() {
        return this.post('/auth/logout');
    },

    /**
     * Get current user profile
     */
    getProfile() {
        return this.get('/auth/profile');
    },

    /**
     * Update user profile
     */
    updateProfile(data) {
        return this.put('/auth/profile', data);
    },

    /**
     * Change password
     */
    changePassword(oldPassword, newPassword) {
        return this.put('/auth/password', {
            old_password: oldPassword,
            new_password: newPassword,
        });
    },

    /**
     * Delete account
     */
    deleteAccount(password) {
        return this.delete('/auth/account', { password });
    },

    // ============================================
    // User Endpoints
    // ============================================

    /**
     * Get users list
     */
    getUsers(page = 1, pageSize = 10, keyword = '') {
        const params = { page, page_size: pageSize };
        if (keyword) {
            params.keyword = keyword;
        }
        return this.get('/users', params);
    },

    /**
     * Get user by ID
     */
    getUser(id) {
        return this.get(`/users/${id}`);
    },

    /**
     * Create user
     */
    createUser(data) {
        return this.post('/users', data);
    },

    /**
     * Update user
     */
    updateUser(id, data) {
        return this.put(`/users/${id}`, data);
    },

    /**
     * Delete user
     */
    deleteUser(id) {
        return this.delete(`/users/${id}`);
    },

    // ============================================
    // Health Check
    // ============================================

    /**
     * Check API health
     */
    async ping() {
        try {
            const response = await fetch('/ping');
            const data = await response.json();
            return data.code === 0;
        } catch {
            return false;
        }
    },
};

/**
 * Custom API Error
 */
class APIError extends Error {
    constructor(code, message) {
        super(message);
        this.name = 'APIError';
        this.code = code;
    }
}

// Export for use in other files
window.API = API;
window.APIError = APIError;
