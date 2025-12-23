/**
 * Main Application Entry for Vibe Coding
 */

const App = {
    /**
     * Initialize application
     */
    async init() {
        // Initialize i18n first
        I18n.init();

        // Initialize UI components
        this.bindEvents();
        this.bindLangSwitcher();
        this.bindModalEvents();

        // Check API status
        this.checkAPIStatus();

        // Initialize auth
        await Auth.init();
        UI.updateAuthUI(Auth.getUser());

        // Subscribe to auth changes
        Auth.subscribe(user => {
            UI.updateAuthUI(user);
        });

        // Start typing animation
        UI.typeText('typing-text', I18n.t('hero.typing'), 150);

        // Periodic API status check
        setInterval(() => this.checkAPIStatus(), 30000);
    },

    /**
     * Bind language switcher events
     */
    bindLangSwitcher() {
        const langBtn = document.getElementById('lang-btn');
        const langDropdown = document.getElementById('lang-dropdown');

        if (!langBtn || !langDropdown) return;

        // Toggle dropdown
        langBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            langDropdown.classList.toggle('hidden');
        });

        // Select language
        document.querySelectorAll('.lang-option').forEach(option => {
            option.addEventListener('click', () => {
                const lang = option.getAttribute('data-lang');
                I18n.setLang(lang);
                langDropdown.classList.add('hidden');
                // Restart typing animation with new language
                const typingText = document.getElementById('typing-text');
                if (typingText) {
                    typingText.textContent = '';
                    UI.typeText('typing-text', I18n.t('hero.typing'), 150);
                }
            });
        });

        // Close dropdown when clicking outside
        document.addEventListener('click', () => {
            langDropdown.classList.add('hidden');
        });
    },

    /**
     * Bind modal events
     */
    bindModalEvents() {
        // Login form
        const loginForm = document.getElementById('login-form');
        if (loginForm) {
            loginForm.addEventListener('submit', async (e) => {
                e.preventDefault();
                const email = document.getElementById('login-email').value;
                const password = document.getElementById('login-password').value;

                try {
                    await Auth.login(email, password);
                    UI.success(I18n.t('msg.login.success'));
                    UI.closeModal('login-modal');
                    e.target.reset();
                } catch (error) {
                    UI.error(error.message);
                }
            });
        }

        // Register form
        const registerForm = document.getElementById('register-form');
        if (registerForm) {
            registerForm.addEventListener('submit', async (e) => {
                e.preventDefault();
                const name = document.getElementById('register-name').value;
                const email = document.getElementById('register-email').value;
                const password = document.getElementById('register-password').value;

                try {
                    await Auth.register(name, email, password);
                    UI.success(I18n.t('msg.register.success'));
                    UI.closeModal('register-modal');
                    e.target.reset();
                } catch (error) {
                    UI.error(error.message);
                }
            });
        }

        // Profile form
        const profileForm = document.getElementById('profile-form');
        if (profileForm) {
            profileForm.addEventListener('submit', async (e) => {
                e.preventDefault();
                const name = document.getElementById('profile-name').value;
                const email = document.getElementById('profile-email').value;
                const age = parseInt(document.getElementById('profile-age').value) || 0;

                try {
                    await Auth.updateProfile({ name, email, age });
                    UI.success(I18n.t('msg.profile.updated'));
                    UI.closeModal('profile-modal');
                } catch (error) {
                    UI.error(error.message);
                }
            });
        }

        // Switch between login and register modals
        const switchToRegister = document.getElementById('switch-to-register');
        if (switchToRegister) {
            switchToRegister.addEventListener('click', (e) => {
                e.preventDefault();
                UI.closeModal('login-modal');
                UI.openModal('register-modal');
            });
        }

        const switchToLogin = document.getElementById('switch-to-login');
        if (switchToLogin) {
            switchToLogin.addEventListener('click', (e) => {
                e.preventDefault();
                UI.closeModal('register-modal');
                UI.openModal('login-modal');
            });
        }

        // Close modals on overlay click
        document.querySelectorAll('.modal-overlay').forEach(overlay => {
            overlay.addEventListener('click', (e) => {
                if (e.target === overlay) {
                    overlay.classList.add('hidden');
                }
            });
        });

        // Close modals on Escape key
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                document.querySelectorAll('.modal-overlay').forEach(m => m.classList.add('hidden'));
            }
        });
    },

    /**
     * Bind event handlers
     */
    bindEvents() {
        // Nav login button - open login modal
        const btnLogin = document.getElementById('btn-login');
        if (btnLogin) {
            btnLogin.addEventListener('click', () => {
                UI.openModal('login-modal');
            });
        }

        // Nav register button - open register modal
        const btnRegister = document.getElementById('btn-register');
        if (btnRegister) {
            btnRegister.addEventListener('click', () => {
                UI.openModal('register-modal');
            });
        }

        // User profile dropdown toggle
        const btnProfile = document.getElementById('btn-profile');
        const userDropdown = document.getElementById('user-dropdown');
        if (btnProfile && userDropdown) {
            btnProfile.addEventListener('click', (e) => {
                e.stopPropagation();
                userDropdown.classList.toggle('hidden');
            });

            document.addEventListener('click', () => {
                userDropdown.classList.add('hidden');
            });
        }

        // Edit profile button
        const btnEditProfile = document.getElementById('btn-edit-profile');
        if (btnEditProfile) {
            btnEditProfile.addEventListener('click', () => {
                if (userDropdown) userDropdown.classList.add('hidden');
                UI.openModal('profile-modal');
            });
        }

        // Logout button
        const btnLogout = document.getElementById('btn-logout');
        if (btnLogout) {
            btnLogout.addEventListener('click', async () => {
                if (userDropdown) userDropdown.classList.add('hidden');
                await Auth.logout();
                UI.success(I18n.t('msg.logout.success'));
            });
        }

        // Start button - go to workspace or open register modal
        const btnStart = document.getElementById('btn-start');
        if (btnStart) {
            btnStart.addEventListener('click', () => {
                if (Auth.isLoggedIn()) {
                    // Go to workspace
                    window.location.href = '/workspace.html';
                } else {
                    // Open register modal
                    UI.openModal('register-modal');
                }
            });
        }
    },

    /**
     * Check API status
     */
    async checkAPIStatus() {
        const dot = document.getElementById('status-dot');
        const text = document.getElementById('status-text');

        if (!dot || !text) return;

        const online = await API.ping();

        if (online) {
            dot.className = 'status-dot online';
            text.textContent = I18n.t('hero.status.online');
        } else {
            dot.className = 'status-dot offline';
            text.textContent = I18n.t('hero.status.offline');
        }
    },
};

// Initialize app when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    App.init();
});

// Export for debugging
window.App = App;
