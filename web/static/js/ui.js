/**
 * UI Module for Vibe Coding
 */

const UI = {
    // ============================================
    // Toast Notifications
    // ============================================

    /**
     * Show toast notification
     */
    toast(message, type = 'info', duration = 3000) {
        const container = document.getElementById('toast-container');
        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        toast.textContent = message;
        container.appendChild(toast);

        setTimeout(() => {
            toast.style.animation = 'slideIn 0.3s ease reverse';
            setTimeout(() => toast.remove(), 300);
        }, duration);
    },

    success(message) {
        this.toast(message, 'success');
    },

    error(message) {
        this.toast(message, 'error');
    },

    info(message) {
        this.toast(message, 'info');
    },

    // ============================================
    // Modal
    // ============================================

    /**
     * Open modal by ID
     */
    openModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.classList.remove('hidden');
        }
    },

    /**
     * Close modal by ID
     */
    closeModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.classList.add('hidden');
        }
    },

    /**
     * Show modal (legacy)
     */
    showModal(title, content) {
        const modalTitle = document.getElementById('modal-title');
        const modalBody = document.getElementById('modal-body');
        const modalOverlay = document.getElementById('modal-overlay');

        if (modalTitle) modalTitle.textContent = title;
        if (modalBody) modalBody.innerHTML = content;
        if (modalOverlay) modalOverlay.classList.remove('hidden');
    },

    /**
     * Hide modal (legacy)
     */
    hideModal() {
        const modalOverlay = document.getElementById('modal-overlay');
        if (modalOverlay) modalOverlay.classList.add('hidden');
    },

    // ============================================
    // Loading State
    // ============================================

    /**
     * Show loading in element
     */
    showLoading(elementId) {
        const el = document.getElementById(elementId);
        if (el) {
            el.innerHTML = `<div class="loading">${I18n.t('msg.loading')}</div>`;
        }
    },

    // ============================================
    // Auth UI
    // ============================================

    /**
     * Update auth UI based on user state
     */
    updateAuthUI(user) {
        const navAuth = document.getElementById('nav-auth');
        const navUser = document.getElementById('nav-user');
        const userName = document.getElementById('user-name');
        const userAvatar = document.getElementById('user-avatar');

        if (user) {
            // Logged in
            if (navAuth) navAuth.classList.add('hidden');
            if (navUser) navUser.classList.remove('hidden');
            if (userName) userName.textContent = user.name;
            if (userAvatar) userAvatar.textContent = this.getInitials(user.name);

            // Fill profile form if exists
            const profileName = document.getElementById('profile-name');
            const profileEmail = document.getElementById('profile-email');
            const profileAge = document.getElementById('profile-age');

            if (profileName) profileName.value = user.name || '';
            if (profileEmail) profileEmail.value = user.email || '';
            if (profileAge) profileAge.value = user.age || '';
        } else {
            // Logged out
            if (navAuth) navAuth.classList.remove('hidden');
            if (navUser) navUser.classList.add('hidden');
        }
    },

    // ============================================
    // User List
    // ============================================

    /**
     * Render user list
     */
    renderUserList(data, keyword = '') {
        const container = document.getElementById('user-list');
        const list = data.list || [];

        if (list.length === 0) {
            const msg = keyword
                ? I18n.t('msg.no.users.search', { keyword: this.escapeHtml(keyword) })
                : I18n.t('msg.no.users');
            container.innerHTML = `<div class="loading">${msg}</div>`;
            return;
        }

        let headerHtml = '';
        if (keyword) {
            headerHtml = `<div class="search-result-header">${I18n.t('msg.search.results', { total: data.total, keyword: this.escapeHtml(keyword) })}</div>`;
        }

        container.innerHTML = headerHtml + list.map(user => `
            <div class="user-item" data-id="${user.id}">
                <div class="user-avatar">${this.getInitials(user.name)}</div>
                <div class="user-info">
                    <div class="name">${this.highlightText(user.name, keyword)}</div>
                    <div class="email">${this.highlightText(user.email || 'No email', keyword)}</div>
                </div>
                <div class="user-meta">
                    <div>Age: ${user.age || '-'}</div>
                    <div>ID: ${user.id}</div>
                </div>
            </div>
        `).join('');
    },

    /**
     * Highlight matching text
     */
    highlightText(text, keyword) {
        if (!keyword || !text) return this.escapeHtml(text);
        const escaped = this.escapeHtml(text);
        const regex = new RegExp(`(${this.escapeRegex(keyword)})`, 'gi');
        return escaped.replace(regex, '<mark>$1</mark>');
    },

    /**
     * Escape regex special characters
     */
    escapeRegex(str) {
        return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    },

    /**
     * Render pagination
     */
    renderPagination(data, currentPage, onPageChange) {
        const container = document.getElementById('pagination');
        const totalPages = data.pages || 1;

        if (totalPages <= 1) {
            container.innerHTML = '';
            return;
        }

        let html = '';

        // Previous button
        html += `<button ${currentPage <= 1 ? 'disabled' : ''} data-page="${currentPage - 1}">&laquo;</button>`;

        // Page numbers
        for (let i = 1; i <= totalPages; i++) {
            if (i === 1 || i === totalPages || (i >= currentPage - 2 && i <= currentPage + 2)) {
                html += `<button class="${i === currentPage ? 'active' : ''}" data-page="${i}">${i}</button>`;
            } else if (i === currentPage - 3 || i === currentPage + 3) {
                html += '<button disabled>...</button>';
            }
        }

        // Next button
        html += `<button ${currentPage >= totalPages ? 'disabled' : ''} data-page="${currentPage + 1}">&raquo;</button>`;

        container.innerHTML = html;

        // Add click handlers
        container.querySelectorAll('button[data-page]').forEach(btn => {
            btn.addEventListener('click', () => {
                const page = parseInt(btn.dataset.page);
                if (page && !btn.disabled) {
                    onPageChange(page);
                }
            });
        });
    },

    // ============================================
    // Typing Effect
    // ============================================

    /**
     * Typing animation
     */
    typeText(elementId, text, speed = 100) {
        const element = document.getElementById(elementId);
        let index = 0;

        const type = () => {
            if (index < text.length) {
                element.textContent += text.charAt(index);
                index++;
                setTimeout(type, speed);
            }
        };

        element.textContent = '';
        type();
    },

    // ============================================
    // Tab Switching
    // ============================================

    /**
     * Initialize tabs
     */
    initTabs() {
        const tabs = document.querySelectorAll('.tab');
        const loginForm = document.getElementById('login-form');
        const registerForm = document.getElementById('register-form');

        tabs.forEach(tab => {
            tab.addEventListener('click', () => {
                tabs.forEach(t => t.classList.remove('active'));
                tab.classList.add('active');

                const tabName = tab.dataset.tab;
                if (tabName === 'login') {
                    loginForm.classList.remove('hidden');
                    registerForm.classList.add('hidden');
                } else {
                    loginForm.classList.add('hidden');
                    registerForm.classList.remove('hidden');
                }
            });
        });
    },

    // ============================================
    // Utilities
    // ============================================

    /**
     * Get initials from name
     */
    getInitials(name) {
        if (!name) return '?';
        return name
            .split(' ')
            .map(word => word.charAt(0))
            .join('')
            .toUpperCase()
            .substring(0, 2);
    },

    /**
     * Escape HTML to prevent XSS
     */
    escapeHtml(str) {
        if (!str) return '';
        const div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    },

    /**
     * Format date
     */
    formatDate(dateStr) {
        if (!dateStr) return '-';
        const date = new Date(dateStr);
        return date.toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
        });
    },
};

// Export for use in other files
window.UI = UI;
