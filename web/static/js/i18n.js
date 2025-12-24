/**
 * i18n Module for Vibe Coding
 * Supports Chinese (zh-CN) and English (en-US)
 */

const I18n = {
    currentLang: 'en-US',

    translations: {
        'en-US': {
            // Navigation
            'nav.about': 'About',
            'nav.features': 'Features',
            'nav.demo': 'Demo',
            'nav.login': 'Login',
            'nav.register': 'Register',
            'nav.logout': 'Logout',

            // Hero
            'hero.typing': 'Vibe Coding',
            'hero.subtitle': 'Let the code flow with your creative energy.<br>Enter the zone where productivity meets passion.',
            'hero.start': 'Start Vibing',
            'hero.learn': 'Learn More',
            'hero.status.checking': 'Checking API status...',
            'hero.status.online': 'API Online',
            'hero.status.offline': 'API Offline',

            // About
            'about.title': 'What is <span class="highlight">Vibe Coding</span>?',
            'about.p1': '<strong>Vibe Coding</strong> is more than just writing code - it\'s about entering a state of creative flow where your ideas seamlessly transform into working software.',
            'about.p2': 'It\'s the feeling when you\'re completely immersed in your work, when solutions appear naturally, and when coding becomes an expression of your creativity rather than a chore.',
            'about.p3': 'Coined by <strong>Andrej Karpathy</strong>, vibe coding represents a paradigm shift in how we approach software development - embracing AI assistance, intuitive tools, and a mindset that values flow state over rigid processes.',
            'about.quote': '"There\'s a new kind of coding I call \'vibe coding\', where you fully give in to the vibes, embrace exponentials, and forget that the code even exists."',
            'about.cite': '- Andrej Karpathy',

            // How It Works
            'howItWorks.title': 'How It <span class="highlight">Works</span>',
            'howItWorks.desc': 'Experience the future of web development. Simply describe what you want, and watch it come to life.',
            'howItWorks.step1.title': 'Describe Your Vision',
            'howItWorks.step1.desc': 'Tell the AI what you want to build using natural language. No need to write code - just express your ideas.',
            'howItWorks.step2.title': 'AI Generates Code',
            'howItWorks.step2.desc': 'Watch as Claude AI transforms your description into clean, production-ready HTML, CSS, and JavaScript.',
            'howItWorks.step3.title': 'Preview & Iterate',
            'howItWorks.step3.desc': 'See your creation in real-time. Refine it through conversation until it\'s exactly what you envisioned.',
            'howItWorks.step4.title': 'Export & Deploy',
            'howItWorks.step4.desc': 'Download your code and deploy it anywhere. Full ownership, no vendor lock-in.',

            // Features
            'features.title': 'Core <span class="highlight">Principles</span>',
            'features.flow.title': 'Flow State',
            'features.flow.desc': 'Enter the zone where time disappears and code flows naturally from your fingertips.',
            'features.ai.title': 'AI-Assisted',
            'features.ai.desc': 'Leverage AI tools to amplify your creativity and handle the mundane details.',
            'features.rapid.title': 'Rapid Iteration',
            'features.rapid.desc': 'Move fast, experiment freely, and let your intuition guide the development process.',
            'features.passion.title': 'Passion-Driven',
            'features.passion.desc': 'Code because you love it, not because you have to. Joy leads to better outcomes.',
            'features.less.title': 'Less is More',
            'features.less.desc': 'Focus on what matters. Let go of unnecessary complexity and embrace simplicity.',
            'features.community.title': 'Community',
            'features.community.desc': 'Share the vibe with fellow developers. Collaboration amplifies creativity.',

            // Use Cases
            'useCases.title': 'What You Can <span class="highlight">Build</span>',
            'useCases.desc': 'From landing pages to complex dashboards, Vibe Coding empowers you to create anything you can imagine.',
            'useCases.landing.title': 'Landing Pages',
            'useCases.landing.desc': 'Beautiful, responsive landing pages that convert visitors into customers.',
            'useCases.dashboard.title': 'Dashboards',
            'useCases.dashboard.desc': 'Data-rich admin panels and analytics dashboards with charts and tables.',
            'useCases.forms.title': 'Forms & Surveys',
            'useCases.forms.desc': 'Interactive forms with validation, multi-step wizards, and surveys.',
            'useCases.components.title': 'Interactive Components',
            'useCases.components.desc': 'Carousels, modals, accordions, and other engaging UI components.',

            // Tech Stack
            'tech.title': 'Powered by <span class="highlight">Modern Tech</span>',

            // Demo / API Playground
            'demo.title': 'API <span class="highlight">Playground</span>',
            'demo.subtitle': 'Experience the backend API in action. Register, login, and explore the features.',
            'demo.login': 'Login',
            'demo.register': 'Register',
            'demo.profile': 'My Profile',
            'demo.directory': 'User Directory',
            'demo.refresh': 'Refresh',
            'demo.search': 'Search',
            'demo.search.placeholder': 'Search by name or email...',

            // Forms
            'form.email': 'Email',
            'form.email.placeholder': 'your@email.com',
            'form.password': 'Password',
            'form.password.placeholder': 'Enter password',
            'form.password.min': 'Min 6 characters',
            'form.name': 'Name',
            'form.name.placeholder': 'Your name',
            'form.age': 'Age',
            'form.age.placeholder': 'Your age',
            'form.submit.login': 'Login',
            'form.submit.register': 'Register',
            'form.submit.update': 'Update Profile',

            // Messages
            'msg.login.success': 'Login successful!',
            'msg.register.success': 'Registration successful!',
            'msg.profile.updated': 'Profile updated!',
            'msg.logout.success': 'Logged out successfully!',
            'msg.loading': 'Loading...',
            'msg.loading.users': 'Loading users...',
            'msg.no.users': 'No users found',
            'msg.no.users.search': 'No users found for "{keyword}"',
            'msg.search.results': 'Found {total} result(s) for "<strong>{keyword}</strong>"',
            'msg.error': 'Error: {message}',

            // Footer
            'footer.built': 'Built with Go + Hertz. Powered by Vibe.',

            // Language
            'lang.switch': 'Language',
            'lang.en': 'English',
            'lang.zh': '中文',

            // Workspace
            'workspace.project': 'New Project',
            'workspace.renameHint': 'Double-click to rename',
            'workspace.preview': 'Preview',
            'workspace.code': 'Code',
            'workspace.home': 'Home',
            'workspace.profile': 'Edit Profile',
            'workspace.logout': 'Logout',
            'workspace.welcome.title': "Let's build something amazing",
            'workspace.welcome.desc': "Describe what you want to create, and I'll help you build it.",
            'workspace.suggestion.landing': 'Landing page',
            'workspace.suggestion.login': 'Login form',
            'workspace.suggestion.dashboard': 'Dashboard',
            'workspace.suggestion.blog': 'Blog card',
            'workspace.input.placeholder': 'Describe what you want to create...',
            'workspace.copy': 'Copy code',
            'workspace.download': 'Download',
            'workspace.placeholder': 'Your creation will appear here',
            'workspace.modal.profile': 'Edit Profile',
            'workspace.modal.name': 'Name',
            'workspace.modal.email': 'Email',
            'workspace.modal.age': 'Age',
            'workspace.modal.save': 'Save Changes',
            'workspace.generate.success': 'Page generated successfully!',
            'workspace.generate.failed': 'Generation failed. Please try again or check if the AI service is available.',
            'workspace.generate.error': 'An error occurred during generation. Please try again later.',
            'workspace.generate.modifying': 'Modifying your page...',
            'workspace.generate.generating': 'Generating your page...',
            'workspace.modify.success': 'Page modified successfully!',
            'workspace.thinking.title': 'AI Thinking Process',
            'workspace.thinking.writing': 'Writing',
            'workspace.thinking.thinking': 'Thinking',
            'workspace.thinking.completed': 'Completed',
            'workspace.thinking.error': 'Error',
            'workspace.thinking.processing': 'Processing',

            // Project Management
            'workspace.projects': 'Projects',
            'workspace.projects.new': 'New Project',
            'workspace.projects.empty': 'No projects yet',
            'workspace.projects.empty.desc': 'Click "New Project" to get started',
            'workspace.projects.delete.confirm': 'Are you sure you want to delete this project?',
            'workspace.projects.save.success': 'Project saved',
            'workspace.projects.save.error': 'Failed to save project',
            'workspace.projects.load.error': 'Failed to load projects',
            'workspace.projects.create.error': 'Failed to create project',
            'workspace.projects.delete.error': 'Failed to delete project',
            'workspace.projects.unsaved': 'You have unsaved changes. Are you sure you want to leave?',
            // Sidebar UI (legacy keys for HTML compatibility)
            'workspace.newProject': 'New Project',
            'workspace.noProjects': 'No projects yet',
            'workspace.hide': 'Hide',
        },

        'zh-CN': {
            // 导航
            'nav.about': '关于',
            'nav.features': '特性',
            'nav.demo': '演示',
            'nav.login': '登录',
            'nav.register': '注册',
            'nav.logout': '退出',

            // Hero 区域
            'hero.typing': 'Vibe Coding',
            'hero.subtitle': '让代码随着你的创意能量流动。<br>进入生产力与激情交汇的心流状态。',
            'hero.start': '开始体验',
            'hero.learn': '了解更多',
            'hero.status.checking': '正在检查 API 状态...',
            'hero.status.online': 'API 在线',
            'hero.status.offline': 'API 离线',

            // 关于
            'about.title': '什么是 <span class="highlight">Vibe Coding</span>？',
            'about.p1': '<strong>Vibe Coding</strong> 不仅仅是写代码——它是进入一种创意流动状态，让你的想法无缝转化为可运行的软件。',
            'about.p2': '这是一种完全沉浸于工作中的感觉，解决方案自然而然地出现，编程成为你创造力的表达，而不是一项苦差事。',
            'about.p3': '由 <strong>Andrej Karpathy</strong> 提出，Vibe Coding 代表着我们对待软件开发方式的范式转变——拥抱 AI 辅助、直觉工具，以及重视心流状态而非刻板流程的心态。',
            'about.quote': '"有一种新的编程方式，我称之为 \'vibe coding\'，你完全沉浸在氛围中，拥抱指数级变化，忘记代码的存在。"',
            'about.cite': '- Andrej Karpathy',

            // 工作流程
            'howItWorks.title': '如何<span class="highlight">使用</span>',
            'howItWorks.desc': '体验 Web 开发的未来。只需描述你想要的，看着它变为现实。',
            'howItWorks.step1.title': '描述你的想法',
            'howItWorks.step1.desc': '用自然语言告诉 AI 你想构建什么。无需编写代码 - 只需表达你的想法。',
            'howItWorks.step2.title': 'AI 生成代码',
            'howItWorks.step2.desc': '看着 Claude AI 将你的描述转化为干净、可用于生产的 HTML、CSS 和 JavaScript 代码。',
            'howItWorks.step3.title': '预览与迭代',
            'howItWorks.step3.desc': '实时查看你的创作。通过对话不断优化，直到完全符合你的期望。',
            'howItWorks.step4.title': '导出与部署',
            'howItWorks.step4.desc': '下载你的代码并部署到任何地方。完全拥有所有权，没有供应商锁定。',

            // 特性
            'features.title': '核心<span class="highlight">原则</span>',
            'features.flow.title': '心流状态',
            'features.flow.desc': '进入时间消失、代码从指尖自然流淌的境界。',
            'features.ai.title': 'AI 辅助',
            'features.ai.desc': '利用 AI 工具放大你的创造力，处理繁琐的细节。',
            'features.rapid.title': '快速迭代',
            'features.rapid.desc': '快速行动，自由实验，让直觉引导开发过程。',
            'features.passion.title': '激情驱动',
            'features.passion.desc': '因为热爱而编程，而非被迫。快乐带来更好的结果。',
            'features.less.title': '少即是多',
            'features.less.desc': '专注于重要的事情。放下不必要的复杂性，拥抱简单。',
            'features.community.title': '社区',
            'features.community.desc': '与开发者同伴分享氛围。协作放大创造力。',

            // 使用场景
            'useCases.title': '你可以<span class="highlight">构建</span>',
            'useCases.desc': '从落地页到复杂的仪表盘，Vibe Coding 让你能够创建任何你能想象的东西。',
            'useCases.landing.title': '落地页',
            'useCases.landing.desc': '精美、响应式的落地页，将访客转化为客户。',
            'useCases.dashboard.title': '仪表盘',
            'useCases.dashboard.desc': '包含图表和表格的数据丰富的管理面板和分析仪表盘。',
            'useCases.forms.title': '表单与问卷',
            'useCases.forms.desc': '带验证功能的交互式表单、多步骤向导和问卷调查。',
            'useCases.components.title': '交互组件',
            'useCases.components.desc': '轮播图、弹窗、手风琴等吸引人的 UI 组件。',

            // 技术栈
            'tech.title': '由<span class="highlight">现代技术</span>驱动',

            // 演示 / API 操场
            'demo.title': 'API <span class="highlight">演练场</span>',
            'demo.subtitle': '体验后端 API 的实际运行。注册、登录并探索各项功能。',
            'demo.login': '登录',
            'demo.register': '注册',
            'demo.profile': '我的资料',
            'demo.directory': '用户目录',
            'demo.refresh': '刷新',
            'demo.search': '搜索',
            'demo.search.placeholder': '按名称或邮箱搜索...',

            // 表单
            'form.email': '邮箱',
            'form.email.placeholder': 'your@email.com',
            'form.password': '密码',
            'form.password.placeholder': '输入密码',
            'form.password.min': '至少 6 个字符',
            'form.name': '姓名',
            'form.name.placeholder': '你的姓名',
            'form.age': '年龄',
            'form.age.placeholder': '你的年龄',
            'form.submit.login': '登录',
            'form.submit.register': '注册',
            'form.submit.update': '更新资料',

            // 消息
            'msg.login.success': '登录成功！',
            'msg.register.success': '注册成功！',
            'msg.profile.updated': '资料已更新！',
            'msg.logout.success': '已成功退出！',
            'msg.loading': '加载中...',
            'msg.loading.users': '正在加载用户...',
            'msg.no.users': '未找到用户',
            'msg.no.users.search': '未找到 "{keyword}" 相关用户',
            'msg.search.results': '找到 {total} 个 "<strong>{keyword}</strong>" 的结果',
            'msg.error': '错误: {message}',

            // 页脚
            'footer.built': '使用 Go + Hertz 构建。由 Vibe 驱动。',

            // 语言
            'lang.switch': '语言',
            'lang.en': 'English',
            'lang.zh': '中文',

            // 工作区
            'workspace.project': '新项目',
            'workspace.renameHint': '双击重命名',
            'workspace.preview': '预览',
            'workspace.code': '代码',
            'workspace.home': '首页',
            'workspace.profile': '编辑资料',
            'workspace.logout': '退出登录',
            'workspace.welcome.title': '让我们构建一些精彩的东西',
            'workspace.welcome.desc': '描述你想要创建的内容，我会帮助你构建它。',
            'workspace.suggestion.landing': '落地页',
            'workspace.suggestion.login': '登录表单',
            'workspace.suggestion.dashboard': '仪表盘',
            'workspace.suggestion.blog': '博客卡片',
            'workspace.input.placeholder': '描述你想要创建的内容...',
            'workspace.copy': '复制代码',
            'workspace.download': '下载',
            'workspace.placeholder': '你的创作将在这里显示',
            'workspace.modal.profile': '编辑资料',
            'workspace.modal.name': '姓名',
            'workspace.modal.email': '邮箱',
            'workspace.modal.age': '年龄',
            'workspace.modal.save': '保存更改',
            'workspace.generate.success': '页面生成成功！',
            'workspace.generate.failed': '生成失败，请重试或检查 AI 服务是否可用。',
            'workspace.generate.error': '生成过程中发生错误，请稍后重试。',
            'workspace.generate.modifying': '正在修改页面...',
            'workspace.generate.generating': '正在生成页面...',
            'workspace.modify.success': '页面修改成功！',
            'workspace.thinking.title': 'AI 思考过程',
            'workspace.thinking.writing': '正在编写',
            'workspace.thinking.thinking': '正在思考',
            'workspace.thinking.completed': '已完成',
            'workspace.thinking.error': '出错',
            'workspace.thinking.processing': '处理中',

            // 项目管理
            'workspace.projects': '项目',
            'workspace.projects.new': '新建项目',
            'workspace.projects.empty': '暂无项目',
            'workspace.projects.empty.desc': '点击"新建项目"开始创建',
            'workspace.projects.delete.confirm': '确定要删除这个项目吗？',
            'workspace.projects.save.success': '项目已保存',
            'workspace.projects.save.error': '保存项目失败',
            'workspace.projects.load.error': '加载项目列表失败',
            'workspace.projects.create.error': '创建项目失败',
            'workspace.projects.delete.error': '删除项目失败',
            'workspace.projects.unsaved': '有未保存的更改，确定要离开吗？',
            // Sidebar UI (legacy keys for HTML compatibility)
            'workspace.newProject': '新建项目',
            'workspace.noProjects': '暂无项目',
            'workspace.hide': '收起',
        }
    },

    /**
     * Initialize i18n
     */
    init() {
        // Load saved language preference
        const saved = localStorage.getItem('lang');
        if (saved && this.translations[saved]) {
            this.currentLang = saved;
        } else {
            // Detect browser language
            const browserLang = navigator.language || navigator.userLanguage;
            if (browserLang.startsWith('zh')) {
                this.currentLang = 'zh-CN';
            }
        }
        this.updatePage();
    },

    /**
     * Get translation
     */
    t(key, params = {}) {
        const translation = this.translations[this.currentLang][key] ||
                           this.translations['en-US'][key] ||
                           key;

        // Replace parameters
        return translation.replace(/\{(\w+)\}/g, (match, param) => {
            return params[param] !== undefined ? params[param] : match;
        });
    },

    /**
     * Switch language
     */
    setLang(lang) {
        if (this.translations[lang]) {
            this.currentLang = lang;
            localStorage.setItem('lang', lang);
            this.updatePage();
        }
    },

    /**
     * Get current language
     */
    getLang() {
        return this.currentLang;
    },

    /**
     * Update all page elements with data-i18n attribute
     */
    updatePage() {
        // Update elements with data-i18n attribute
        document.querySelectorAll('[data-i18n]').forEach(el => {
            const key = el.getAttribute('data-i18n');
            const translation = this.t(key);

            if (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA') {
                if (el.getAttribute('placeholder')) {
                    el.placeholder = translation;
                } else {
                    el.value = translation;
                }
            } else {
                el.innerHTML = translation;
            }
        });

        // Update elements with data-i18n-placeholder attribute
        document.querySelectorAll('[data-i18n-placeholder]').forEach(el => {
            const key = el.getAttribute('data-i18n-placeholder');
            el.placeholder = this.t(key);
        });

        // Update language switcher display
        const langBtn = document.getElementById('lang-current');
        if (langBtn) {
            langBtn.textContent = this.currentLang === 'zh-CN' ? '中文' : 'EN';
        }

        // Dispatch event for dynamic content
        document.dispatchEvent(new CustomEvent('langChange', { detail: this.currentLang }));
    }
};

// Export for use in other files
window.I18n = I18n;
