/**
 * Workspace Application for Vibe Coding
 */

const Workspace = {
    chatMessages: [],
    currentView: 'preview',
    panelWidth: 420,

    /**
     * Initialize workspace
     */
    async init() {
        // Check authentication
        await Auth.init();
        if (!Auth.isLoggedIn()) {
            window.location.href = '/';
            return;
        }

        // Update user info
        this.updateUserInfo();

        // Bind events
        this.bindEvents();
        this.bindPanelResizer();

        // Subscribe to auth changes
        Auth.subscribe(user => {
            if (!user) {
                window.location.href = '/';
            } else {
                this.updateUserInfo();
            }
        });
    },

    /**
     * Update user info in header
     */
    updateUserInfo() {
        const user = Auth.getUser();
        if (user) {
            const avatar = document.getElementById('ws-user-avatar');
            const name = document.getElementById('ws-user-name');
            const email = document.getElementById('ws-user-email');

            if (avatar) avatar.textContent = this.getInitials(user.name);
            if (name) name.textContent = user.name;
            if (email) email.textContent = user.email;

            // Update profile form
            const profileName = document.getElementById('ws-profile-name');
            const profileEmail = document.getElementById('ws-profile-email');
            const profileAge = document.getElementById('ws-profile-age');

            if (profileName) profileName.value = user.name || '';
            if (profileEmail) profileEmail.value = user.email || '';
            if (profileAge) profileAge.value = user.age || '';
        }
    },

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
     * Bind event handlers
     */
    bindEvents() {
        // User menu toggle
        const userMenuBtn = document.getElementById('btn-user-menu');
        const userDropdown = document.getElementById('ws-user-dropdown');

        if (userMenuBtn && userDropdown) {
            userMenuBtn.addEventListener('click', (e) => {
                e.stopPropagation();
                userDropdown.classList.toggle('hidden');
            });

            document.addEventListener('click', () => {
                userDropdown.classList.add('hidden');
            });
        }

        // Profile button
        const profileBtn = document.getElementById('ws-btn-profile');
        if (profileBtn) {
            profileBtn.addEventListener('click', () => {
                userDropdown.classList.add('hidden');
                UI.openModal('ws-profile-modal');
            });
        }

        // Logout button
        const logoutBtn = document.getElementById('ws-btn-logout');
        if (logoutBtn) {
            logoutBtn.addEventListener('click', async () => {
                await Auth.logout();
                UI.success('Logged out successfully');
            });
        }

        // Profile form
        const profileForm = document.getElementById('ws-profile-form');
        if (profileForm) {
            profileForm.addEventListener('submit', async (e) => {
                e.preventDefault();
                const name = document.getElementById('ws-profile-name').value;
                const email = document.getElementById('ws-profile-email').value;
                const age = parseInt(document.getElementById('ws-profile-age').value) || 0;

                try {
                    await Auth.updateProfile({ name, email, age });
                    UI.success('Profile updated successfully');
                    UI.closeModal('ws-profile-modal');
                } catch (error) {
                    UI.error(error.message);
                }
            });
        }

        // View tabs
        document.querySelectorAll('.view-tab').forEach(tab => {
            tab.addEventListener('click', () => {
                const view = tab.dataset.view;
                this.switchView(view);
            });
        });

        // Chat input
        const chatInput = document.getElementById('chat-input');
        const sendBtn = document.getElementById('btn-send');

        if (chatInput) {
            // Auto-resize textarea
            chatInput.addEventListener('input', () => {
                chatInput.style.height = 'auto';
                chatInput.style.height = Math.min(chatInput.scrollHeight, 150) + 'px';
            });

            // Handle Enter key
            chatInput.addEventListener('keydown', (e) => {
                if (e.key === 'Enter' && !e.shiftKey) {
                    e.preventDefault();
                    this.sendMessage();
                }
            });
        }

        if (sendBtn) {
            sendBtn.addEventListener('click', () => this.sendMessage());
        }

        // Suggestion chips
        document.querySelectorAll('.suggestion-chip').forEach(chip => {
            chip.addEventListener('click', () => {
                const prompt = chip.dataset.prompt;
                if (chatInput) {
                    chatInput.value = prompt;
                    chatInput.focus();
                }
            });
        });

        // Code tabs
        document.querySelectorAll('.code-tab').forEach(tab => {
            tab.addEventListener('click', () => {
                document.querySelectorAll('.code-tab').forEach(t => t.classList.remove('active'));
                tab.classList.add('active');
            });
        });

        // Modal close on overlay click
        document.querySelectorAll('.modal-overlay').forEach(overlay => {
            overlay.addEventListener('click', (e) => {
                if (e.target === overlay) {
                    overlay.classList.add('hidden');
                }
            });
        });

        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                document.querySelectorAll('.modal-overlay').forEach(m => m.classList.add('hidden'));
            }
        });
    },

    /**
     * Switch between preview and code view
     */
    switchView(view) {
        this.currentView = view;

        // Update tabs
        document.querySelectorAll('.view-tab').forEach(tab => {
            tab.classList.toggle('active', tab.dataset.view === view);
        });

        // Update views
        const previewView = document.getElementById('preview-view');
        const codeView = document.getElementById('code-view');

        if (view === 'preview') {
            previewView.classList.add('active');
            codeView.classList.remove('active');
        } else {
            previewView.classList.remove('active');
            codeView.classList.add('active');
        }
    },

    /**
     * Send message
     */
    async sendMessage() {
        const input = document.getElementById('chat-input');
        const message = input.value.trim();

        if (!message) return;

        // Clear input
        input.value = '';
        input.style.height = 'auto';

        // Hide welcome and suggestions
        const welcome = document.querySelector('.welcome-message');
        const suggestions = document.getElementById('suggestions');
        if (welcome) welcome.style.display = 'none';
        if (suggestions) suggestions.style.display = 'none';

        // Add user message
        this.addMessage(message, 'user');

        // Simulate AI response
        setTimeout(() => {
            this.addMessage("I understand you want to create something. This is a demo workspace - the AI integration is not yet implemented. But you can see how the interface works!", 'ai');
        }, 1000);
    },

    /**
     * Add message to chat
     */
    addMessage(text, type) {
        const messagesContainer = document.getElementById('chat-messages');
        const user = Auth.getUser();

        const messageDiv = document.createElement('div');
        messageDiv.className = 'chat-message';

        const avatar = type === 'user' ? this.getInitials(user?.name || 'You') : 'AI';
        const avatarClass = type === 'ai' ? 'ai' : '';

        messageDiv.innerHTML = `
            <div class="message-avatar ${avatarClass}">${avatar}</div>
            <div class="message-content ${type}">
                <p>${this.escapeHtml(text)}</p>
            </div>
        `;

        messagesContainer.appendChild(messageDiv);
        messagesContainer.scrollTop = messagesContainer.scrollHeight;

        this.chatMessages.push({ text, type });
    },

    /**
     * Escape HTML
     */
    escapeHtml(str) {
        const div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    },

    /**
     * Bind panel resizer
     */
    bindPanelResizer() {
        const resizer = document.getElementById('panel-resizer');
        const chatPanel = document.getElementById('chat-panel');

        if (!resizer || !chatPanel) return;

        let isResizing = false;
        let startX = 0;
        let startWidth = 0;

        resizer.addEventListener('mousedown', (e) => {
            isResizing = true;
            startX = e.clientX;
            startWidth = chatPanel.offsetWidth;
            resizer.classList.add('active');
            document.body.style.cursor = 'col-resize';
            document.body.style.userSelect = 'none';
        });

        document.addEventListener('mousemove', (e) => {
            if (!isResizing) return;

            const diff = e.clientX - startX;
            const newWidth = Math.max(320, Math.min(600, startWidth + diff));
            chatPanel.style.width = newWidth + 'px';
            this.panelWidth = newWidth;
        });

        document.addEventListener('mouseup', () => {
            if (isResizing) {
                isResizing = false;
                resizer.classList.remove('active');
                document.body.style.cursor = '';
                document.body.style.userSelect = '';
            }
        });
    },
};

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    Workspace.init();
});

// Export for debugging
window.Workspace = Workspace;
