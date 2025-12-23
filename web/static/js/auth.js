/**
 * Authentication Module for Vibe Coding
 */

const Auth = {
    user: null,
    listeners: [],

    /**
     * Initialize auth state
     */
    async init() {
        const token = API.getToken();
        if (token) {
            try {
                this.user = await API.getProfile();
                this.notifyListeners();
            } catch (error) {
                // Token invalid, clear it
                API.removeToken();
                this.user = null;
            }
        }
        return this.user;
    },

    /**
     * Check if user is logged in
     */
    isLoggedIn() {
        return !!this.user && !!API.getToken();
    },

    /**
     * Get current user
     */
    getUser() {
        return this.user;
    },

    /**
     * Register new user
     */
    async register(name, email, password) {
        const result = await API.register(name, email, password);
        API.setToken(result.token);
        this.user = result.user;
        this.notifyListeners();
        return result;
    },

    /**
     * Login user
     */
    async login(email, password) {
        const result = await API.login(email, password);
        API.setToken(result.token);
        this.user = result.user;
        this.notifyListeners();
        return result;
    },

    /**
     * Logout user
     */
    async logout() {
        try {
            await API.logout();
        } catch (error) {
            // Ignore logout errors
            console.warn('Logout error:', error.message);
        }
        API.removeToken();
        this.user = null;
        this.notifyListeners();
    },

    /**
     * Update profile
     */
    async updateProfile(data) {
        this.user = await API.updateProfile(data);
        this.notifyListeners();
        return this.user;
    },

    /**
     * Change password
     */
    async changePassword(oldPassword, newPassword) {
        await API.changePassword(oldPassword, newPassword);
    },

    /**
     * Delete account
     */
    async deleteAccount(password) {
        await API.deleteAccount(password);
        API.removeToken();
        this.user = null;
        this.notifyListeners();
    },

    /**
     * Subscribe to auth state changes
     */
    subscribe(callback) {
        this.listeners.push(callback);
        return () => {
            this.listeners = this.listeners.filter(l => l !== callback);
        };
    },

    /**
     * Notify all listeners of state change
     */
    notifyListeners() {
        this.listeners.forEach(callback => {
            try {
                callback(this.user);
            } catch (error) {
                console.error('Auth listener error:', error);
            }
        });
    },
};

// Export for use in other files
window.Auth = Auth;
