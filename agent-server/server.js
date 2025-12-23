/**
 * Vibe Coding Agent Server
 * Uses Claude Agent SDK to generate web pages
 */

import express from 'express';
import cors from 'cors';
import { query } from '@anthropic-ai/claude-agent-sdk';
import { v4 as uuidv4 } from 'uuid';
import fs from 'fs';
import path from 'path';
import os from 'os';
import jwt from 'jsonwebtoken';

// JWT 配置 - 必须与 Go 后端保持一致
const JWT_SECRET = process.env.JWT_SECRET || 'your-secret-key-change-in-production';
const JWT_ISSUER = process.env.JWT_ISSUER || 'test-tt';

const app = express();
app.use(cors());
app.use(express.json());

// 请求日志中间件
app.use((req, res, next) => {
    console.log(`[${new Date().toISOString()}] ${req.method} ${req.url}`);
    next();
});

/**
 * JWT 认证中间件
 * 验证 Authorization header 中的 Bearer token
 */
function authMiddleware(req, res, next) {
    const authHeader = req.headers.authorization;

    if (!authHeader) {
        return res.status(401).json({
            success: false,
            code: 4001,
            error: 'Authorization header required'
        });
    }

    // 解析 Bearer token
    const parts = authHeader.split(' ');
    if (parts.length !== 2 || parts[0] !== 'Bearer') {
        return res.status(401).json({
            success: false,
            code: 4001,
            error: 'Invalid authorization format. Use: Bearer <token>'
        });
    }

    const token = parts[1];

    try {
        // 验证 token
        const decoded = jwt.verify(token, JWT_SECRET, {
            issuer: JWT_ISSUER,
            algorithms: ['HS256']
        });

        // 将用户信息附加到请求对象
        req.user = {
            userId: decoded.user_id,
            username: decoded.username
        };

        next();
    } catch (err) {
        if (err.name === 'TokenExpiredError') {
            return res.status(401).json({
                success: false,
                code: 4011,
                error: 'Token has expired'
            });
        }
        if (err.name === 'JsonWebTokenError') {
            return res.status(401).json({
                success: false,
                code: 4012,
                error: 'Invalid token'
            });
        }
        return res.status(401).json({
            success: false,
            code: 4010,
            error: 'Authentication failed'
        });
    }
}

// Store generation sessions
const sessions = new Map();

// System prompt for web page generation
const SYSTEM_PROMPT = `You are an expert web developer assistant. Your task is to generate beautiful, modern, and responsive web pages.

When generating code:
1. Create a complete HTML file with embedded CSS (in <style> tags)
2. Use modern CSS features like flexbox, grid, and CSS variables
3. Make designs responsive with media queries
4. Use a professional color scheme
5. Add smooth transitions and hover effects
6. Include proper semantic HTML structure
7. Make the design visually appealing and production-ready

Always output the complete HTML code that can be directly rendered in a browser.
Do NOT use external CSS files or JavaScript libraries unless specifically requested.
Respond with ONLY the HTML code, no explanations before or after.`;

/**
 * Generate web page using Claude Agent SDK with streaming
 */
async function generateWithAgent(prompt, sessionId) {
    const session = sessions.get(sessionId);
    if (!session) return;

    session.status = 'generating';
    session.messages.push({ role: 'user', content: prompt });
    session.streamContent = ''; // 用于存储流式内容

    try {
        let fullResponse = '';

        // Use Claude Agent SDK query
        for await (const message of query({
            prompt: `${SYSTEM_PROMPT}\n\nUser request: ${prompt}\n\nGenerate a complete HTML page with embedded CSS for this request. Output ONLY the HTML code.`,
            options: {
                allowedTools: ['Read', 'Write', 'Edit'],
                maxTurns: 5,
            }
        })) {
            if (message.type === 'assistant' && message.message?.content) {
                for (const block of message.message.content) {
                    if (block.type === 'text') {
                        fullResponse += block.text;
                        // 更新流式内容
                        session.streamContent = fullResponse;
                    }
                }
            }

            if (message.type === 'result') {
                session.status = message.subtype === 'success' ? 'completed' : 'error';
            }
        }

        // Extract HTML from response
        const htmlCode = extractHTML(fullResponse);
        const cssCode = extractCSS(htmlCode);

        // Store result
        session.result = {
            html: htmlCode,
            css: cssCode,
            message: generateSummary(prompt)
        };
        session.status = 'completed';
        session.messages.push({
            role: 'assistant',
            content: session.result.message
        });

    } catch (error) {
        console.error('Generation error:', error);
        session.status = 'error';
        session.error = error.message;

        // Fallback to a template if SDK fails
        const fallback = getFallbackTemplate(prompt);
        if (fallback) {
            session.result = fallback;
            session.status = 'completed';
            session.messages.push({
                role: 'assistant',
                content: fallback.message + '\n\n(Note: Using fallback template)'
            });
        }
    }
}

/**
 * Extract HTML from response
 */
function extractHTML(response) {
    console.log('extractHTML input length:', response.length);

    // Try to find HTML code block (markdown format)
    const htmlMatch = response.match(/```html\s*([\s\S]*?)```/);
    if (htmlMatch) {
        console.log('Found markdown HTML block');
        return htmlMatch[1].trim();
    }

    // Try to find raw HTML
    const docMatch = response.match(/<!DOCTYPE[\s\S]*<\/html>/i);
    if (docMatch) {
        console.log('Found raw HTML');
        return docMatch[0].trim();
    }

    // Return full response if it looks like HTML
    if (response.includes('<!DOCTYPE') || response.includes('<html')) {
        console.log('Response looks like HTML');
        return response.trim();
    }

    console.log('No HTML found, returning as-is');
    return response;
}

/**
 * Extract CSS from HTML
 */
function extractCSS(html) {
    const styleMatch = html.match(/<style[^>]*>([\s\S]*?)<\/style>/i);
    return styleMatch ? styleMatch[1].trim() : '';
}

/**
 * Generate summary message
 */
function generateSummary(prompt) {
    const lower = prompt.toLowerCase();

    if (lower.includes('landing') || lower.includes('saas')) {
        return "I've created a modern SaaS landing page with:\n\n" +
            "- **Navigation bar** with logo and links\n" +
            "- **Hero section** with headline and CTAs\n" +
            "- **Features section** with cards\n" +
            "- **Responsive design** for all devices\n\n" +
            "Feel free to customize the content and colors!";
    }

    if (lower.includes('login') || lower.includes('form')) {
        return "I've created a login form with:\n\n" +
            "- **Email and password fields**\n" +
            "- **Form validation styling**\n" +
            "- **Remember me** option\n" +
            "- **Responsive design**\n\n" +
            "The form is ready for integration with your backend!";
    }

    if (lower.includes('dashboard')) {
        return "I've created a dashboard with:\n\n" +
            "- **Sidebar navigation**\n" +
            "- **Stats cards** showing key metrics\n" +
            "- **Clean layout** with proper spacing\n" +
            "- **Professional color scheme**\n\n" +
            "You can add more widgets as needed!";
    }

    if (lower.includes('blog') || lower.includes('card')) {
        return "I've created a blog card component with:\n\n" +
            "- **Image with hover effect**\n" +
            "- **Category and tags**\n" +
            "- **Author section**\n" +
            "- **Clean typography**\n\n" +
            "Perfect for a blog listing page!";
    }

    return "I've generated a web page based on your request.\n\n" +
        "The design includes:\n" +
        "- Modern, clean layout\n" +
        "- Responsive design\n" +
        "- Professional styling\n\n" +
        "Feel free to customize it further!";
}

/**
 * Fallback templates when SDK is unavailable
 */
function getFallbackTemplate(prompt) {
    const lower = prompt.toLowerCase();

    // Landing page template
    if (lower.includes('landing') || lower.includes('saas')) {
        return {
            html: `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SaaS Landing Page</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; }
        .navbar { display: flex; justify-content: space-between; align-items: center; padding: 20px 40px; }
        .logo { font-size: 24px; font-weight: 700; color: #6366f1; }
        .nav-links { display: flex; gap: 30px; align-items: center; }
        .nav-links a { text-decoration: none; color: #475569; }
        .btn-primary { background: linear-gradient(135deg, #6366f1, #8b5cf6); color: white; padding: 12px 24px; border: none; border-radius: 8px; cursor: pointer; font-weight: 600; }
        .btn-outline { background: transparent; border: 2px solid #6366f1; color: #6366f1; padding: 12px 24px; border-radius: 8px; cursor: pointer; font-weight: 600; }
        .hero { text-align: center; padding: 100px 20px; background: linear-gradient(135deg, #f8fafc 0%, #e2e8f0 100%); }
        .hero h1 { font-size: 48px; color: #1e293b; margin-bottom: 20px; }
        .hero p { font-size: 20px; color: #64748b; margin-bottom: 40px; max-width: 600px; margin-left: auto; margin-right: auto; }
        .hero-buttons { display: flex; gap: 16px; justify-content: center; }
        .features { padding: 80px 40px; text-align: center; }
        .features h2 { font-size: 36px; margin-bottom: 50px; color: #1e293b; }
        .feature-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 30px; max-width: 1000px; margin: 0 auto; }
        .feature-card { padding: 40px 30px; background: white; border-radius: 16px; box-shadow: 0 4px 20px rgba(0,0,0,0.08); }
        .feature-icon { font-size: 40px; margin-bottom: 20px; }
        .feature-card h3 { font-size: 20px; margin-bottom: 10px; color: #1e293b; }
        .feature-card p { color: #64748b; }
        @media (max-width: 768px) {
            .hero h1 { font-size: 32px; }
            .hero-buttons { flex-direction: column; }
            .navbar { padding: 15px 20px; }
        }
    </style>
</head>
<body>
    <nav class="navbar">
        <div class="logo">SaaSify</div>
        <div class="nav-links">
            <a href="#features">Features</a>
            <a href="#pricing">Pricing</a>
            <button class="btn-primary">Get Started</button>
        </div>
    </nav>
    <section class="hero">
        <h1>Build Better Products, Faster</h1>
        <p>The all-in-one platform for modern teams to collaborate, build, and ship amazing products.</p>
        <div class="hero-buttons">
            <button class="btn-primary">Start Free Trial</button>
            <button class="btn-outline">Watch Demo</button>
        </div>
    </section>
    <section class="features" id="features">
        <h2>Everything you need</h2>
        <div class="feature-grid">
            <div class="feature-card">
                <div class="feature-icon">⚡</div>
                <h3>Lightning Fast</h3>
                <p>Optimized for speed and performance across all devices.</p>
            </div>
            <div class="feature-card">
                <div class="feature-icon">🔒</div>
                <h3>Secure by Default</h3>
                <p>Enterprise-grade security built into every feature.</p>
            </div>
            <div class="feature-card">
                <div class="feature-icon">📊</div>
                <h3>Analytics</h3>
                <p>Deep insights into your data and user behavior.</p>
            </div>
        </div>
    </section>
</body>
</html>`,
            css: '',
            message: "I've created a modern SaaS landing page with:\n\n" +
                "- **Navigation bar** with logo and links\n" +
                "- **Hero section** with headline and CTAs\n" +
                "- **Features section** with 3 cards\n" +
                "- **Responsive design**\n\n" +
                "Feel free to customize the content and colors!"
        };
    }

    // Login form template
    if (lower.includes('login') || lower.includes('form')) {
        return {
            html: `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Login</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, sans-serif; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); min-height: 100vh; display: flex; align-items: center; justify-content: center; }
        .login-container { width: 100%; max-width: 420px; padding: 20px; }
        .login-card { background: white; border-radius: 16px; padding: 40px; box-shadow: 0 20px 60px rgba(0,0,0,0.3); }
        .login-header { text-align: center; margin-bottom: 30px; }
        .login-header h1 { font-size: 28px; color: #1e293b; margin-bottom: 8px; }
        .login-header p { color: #64748b; }
        .form-group { margin-bottom: 20px; }
        .form-group label { display: block; font-size: 14px; font-weight: 500; color: #475569; margin-bottom: 8px; }
        .form-group input { width: 100%; padding: 14px 16px; border: 2px solid #e2e8f0; border-radius: 10px; font-size: 16px; transition: all 0.2s; }
        .form-group input:focus { outline: none; border-color: #6366f1; box-shadow: 0 0 0 4px rgba(99,102,241,0.1); }
        .form-options { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
        .checkbox { display: flex; align-items: center; gap: 8px; color: #64748b; font-size: 14px; }
        .forgot-link { color: #6366f1; text-decoration: none; font-size: 14px; }
        .btn-submit { width: 100%; padding: 14px; background: linear-gradient(135deg, #6366f1, #8b5cf6); color: white; border: none; border-radius: 10px; font-size: 16px; font-weight: 600; cursor: pointer; transition: transform 0.2s; }
        .btn-submit:hover { transform: translateY(-2px); }
        .login-footer { text-align: center; margin-top: 24px; color: #64748b; }
        .login-footer a { color: #6366f1; text-decoration: none; font-weight: 500; }
    </style>
</head>
<body>
    <div class="login-container">
        <div class="login-card">
            <div class="login-header">
                <h1>Welcome back</h1>
                <p>Sign in to your account</p>
            </div>
            <form class="login-form">
                <div class="form-group">
                    <label>Email</label>
                    <input type="email" placeholder="you@example.com" required>
                </div>
                <div class="form-group">
                    <label>Password</label>
                    <input type="password" placeholder="Enter your password" required>
                </div>
                <div class="form-options">
                    <label class="checkbox">
                        <input type="checkbox"> Remember me
                    </label>
                    <a href="#" class="forgot-link">Forgot password?</a>
                </div>
                <button type="submit" class="btn-submit">Sign In</button>
            </form>
            <div class="login-footer">
                <p>Don't have an account? <a href="#">Sign up</a></p>
            </div>
        </div>
    </div>
</body>
</html>`,
            css: '',
            message: "I've created a login form with:\n\n" +
                "- **Email and password fields**\n" +
                "- **Remember me checkbox**\n" +
                "- **Forgot password link**\n" +
                "- **Beautiful gradient background**\n\n" +
                "The form is ready for integration!"
        };
    }

    // Dashboard template
    if (lower.includes('dashboard')) {
        return {
            html: `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Dashboard</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, sans-serif; background: #f1f5f9; }
        .dashboard { display: flex; min-height: 100vh; }
        .sidebar { width: 260px; background: #1e293b; color: white; padding: 20px; }
        .sidebar-header { padding: 20px 0; border-bottom: 1px solid #334155; margin-bottom: 20px; }
        .logo { font-size: 24px; font-weight: 700; }
        .nav-item { display: flex; align-items: center; gap: 12px; padding: 14px 16px; color: #94a3b8; text-decoration: none; border-radius: 10px; margin-bottom: 8px; transition: all 0.2s; }
        .nav-item:hover, .nav-item.active { background: #334155; color: white; }
        .icon { font-size: 20px; }
        .main-content { flex: 1; padding: 30px; }
        .topbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 30px; }
        .topbar h1 { font-size: 28px; color: #1e293b; }
        .stats-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin-bottom: 30px; }
        .stat-card { background: white; padding: 24px; border-radius: 16px; display: flex; align-items: center; gap: 16px; box-shadow: 0 2px 10px rgba(0,0,0,0.05); }
        .stat-icon { width: 56px; height: 56px; border-radius: 14px; display: flex; align-items: center; justify-content: center; font-size: 24px; }
        .stat-icon.blue { background: #dbeafe; }
        .stat-icon.green { background: #dcfce7; }
        .stat-icon.purple { background: #f3e8ff; }
        .stat-value { font-size: 28px; font-weight: 700; color: #1e293b; display: block; }
        .stat-label { color: #64748b; font-size: 14px; }
        @media (max-width: 768px) {
            .sidebar { display: none; }
            .stats-grid { grid-template-columns: 1fr; }
        }
    </style>
</head>
<body>
    <div class="dashboard">
        <aside class="sidebar">
            <div class="sidebar-header">
                <span class="logo">Dashboard</span>
            </div>
            <nav>
                <a href="#" class="nav-item active"><span class="icon">📊</span> Overview</a>
                <a href="#" class="nav-item"><span class="icon">📈</span> Analytics</a>
                <a href="#" class="nav-item"><span class="icon">👥</span> Users</a>
                <a href="#" class="nav-item"><span class="icon">⚙️</span> Settings</a>
            </nav>
        </aside>
        <main class="main-content">
            <header class="topbar">
                <h1>Overview</h1>
            </header>
            <div class="stats-grid">
                <div class="stat-card">
                    <div class="stat-icon blue">👥</div>
                    <div>
                        <span class="stat-value">2,543</span>
                        <span class="stat-label">Total Users</span>
                    </div>
                </div>
                <div class="stat-card">
                    <div class="stat-icon green">💰</div>
                    <div>
                        <span class="stat-value">$45,234</span>
                        <span class="stat-label">Revenue</span>
                    </div>
                </div>
                <div class="stat-card">
                    <div class="stat-icon purple">📦</div>
                    <div>
                        <span class="stat-value">1,234</span>
                        <span class="stat-label">Orders</span>
                    </div>
                </div>
            </div>
        </main>
    </div>
</body>
</html>`,
            css: '',
            message: "I've created a dashboard with:\n\n" +
                "- **Sidebar navigation**\n" +
                "- **Stats cards** with metrics\n" +
                "- **Clean professional design**\n" +
                "- **Responsive layout**\n\n" +
                "You can add more widgets as needed!"
        };
    }

    // Blog card template
    if (lower.includes('blog') || lower.includes('card')) {
        return {
            html: `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Blog Card</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, sans-serif; background: #f8fafc; min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 40px 20px; }
        .container { width: 100%; max-width: 420px; }
        .blog-card { background: white; border-radius: 20px; overflow: hidden; box-shadow: 0 10px 40px rgba(0,0,0,0.1); transition: transform 0.3s; }
        .blog-card:hover { transform: translateY(-8px); }
        .card-image { position: relative; height: 220px; overflow: hidden; background: linear-gradient(135deg, #6366f1, #8b5cf6); }
        .card-image img { width: 100%; height: 100%; object-fit: cover; }
        .category { position: absolute; top: 16px; left: 16px; background: rgba(99,102,241,0.9); color: white; padding: 6px 14px; border-radius: 20px; font-size: 13px; font-weight: 500; }
        .card-content { padding: 24px; }
        .tags { display: flex; gap: 8px; margin-bottom: 16px; flex-wrap: wrap; }
        .tag { background: #f1f5f9; color: #64748b; padding: 6px 12px; border-radius: 6px; font-size: 12px; font-weight: 500; }
        .card-content h2 { font-size: 20px; color: #1e293b; margin-bottom: 12px; line-height: 1.4; }
        .card-content p { color: #64748b; font-size: 15px; line-height: 1.6; margin-bottom: 20px; }
        .card-footer { display: flex; justify-content: space-between; align-items: center; padding-top: 20px; border-top: 1px solid #f1f5f9; }
        .author { display: flex; align-items: center; gap: 12px; }
        .author-avatar { width: 40px; height: 40px; background: linear-gradient(135deg, #6366f1, #a855f7); border-radius: 50%; display: flex; align-items: center; justify-content: center; color: white; font-weight: 600; font-size: 14px; }
        .author-name { font-weight: 600; color: #1e293b; display: block; font-size: 14px; }
        .post-date { color: #94a3b8; font-size: 13px; }
        .read-more { color: #6366f1; text-decoration: none; font-weight: 600; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <article class="blog-card">
            <div class="card-image">
                <span class="category">Technology</span>
            </div>
            <div class="card-content">
                <div class="tags">
                    <span class="tag">React</span>
                    <span class="tag">JavaScript</span>
                    <span class="tag">Web Dev</span>
                </div>
                <h2>Building Modern Web Applications</h2>
                <p>Learn how to create scalable and maintainable web applications using modern tools and best practices.</p>
                <div class="card-footer">
                    <div class="author">
                        <div class="author-avatar">JD</div>
                        <div>
                            <span class="author-name">John Doe</span>
                            <span class="post-date">Dec 23, 2025</span>
                        </div>
                    </div>
                    <a href="#" class="read-more">Read more</a>
                </div>
            </div>
        </article>
    </div>
</body>
</html>`,
            css: '',
            message: "I've created a blog card with:\n\n" +
                "- **Image placeholder** with category badge\n" +
                "- **Tags** for topics\n" +
                "- **Author section** with avatar\n" +
                "- **Hover animation**\n\n" +
                "Perfect for a blog listing page!"
        };
    }

    return null;
}

// API Routes

/**
 * Create new generation session
 * 需要认证
 */
app.post('/api/generate', authMiddleware, async (req, res) => {
    const { prompt } = req.body;

    if (!prompt) {
        return res.status(400).json({
            success: false,
            error: 'Prompt is required'
        });
    }

    const sessionId = uuidv4();

    // Create session
    sessions.set(sessionId, {
        id: sessionId,
        prompt,
        status: 'pending',
        messages: [],
        result: null,
        error: null,
        createdAt: new Date()
    });

    // Start generation in background
    generateWithAgent(prompt, sessionId);

    res.json({
        success: true,
        sessionId
    });
});

/**
 * Get session status and result
 * 需要认证
 */
app.get('/api/session/:sessionId', authMiddleware, (req, res) => {
    const { sessionId } = req.params;
    const session = sessions.get(sessionId);

    if (!session) {
        return res.status(404).json({
            success: false,
            error: 'Session not found'
        });
    }

    res.json({
        success: true,
        session: {
            id: session.id,
            status: session.status,
            messages: session.messages,
            result: session.result,
            error: session.error,
            streamContent: session.streamContent || '' // 返回流式内容供前端显示
        }
    });
});

/**
 * Stream generation (SSE) with real-time content
 * 需要认证
 */
app.get('/api/stream/:sessionId', authMiddleware, (req, res) => {
    const { sessionId } = req.params;
    const session = sessions.get(sessionId);

    if (!session) {
        return res.status(404).json({
            success: false,
            error: 'Session not found'
        });
    }

    res.setHeader('Content-Type', 'text/event-stream');
    res.setHeader('Cache-Control', 'no-cache');
    res.setHeader('Connection', 'keep-alive');
    res.setHeader('Access-Control-Allow-Origin', '*');

    let lastContentLength = 0;

    const sendUpdate = () => {
        const currentSession = sessions.get(sessionId);
        if (!currentSession) {
            res.end();
            return true;
        }

        // 发送流式内容增量
        if (currentSession.streamContent && currentSession.streamContent.length > lastContentLength) {
            const newContent = currentSession.streamContent.slice(lastContentLength);
            lastContentLength = currentSession.streamContent.length;

            res.write(`data: ${JSON.stringify({
                type: 'content',
                content: newContent,
                fullContent: currentSession.streamContent
            })}\n\n`);
        }

        // 检查是否完成
        if (currentSession.status === 'completed') {
            res.write(`data: ${JSON.stringify({
                type: 'complete',
                status: 'completed',
                result: currentSession.result
            })}\n\n`);
            res.end();
            return true;
        }

        if (currentSession.status === 'error') {
            res.write(`data: ${JSON.stringify({
                type: 'error',
                status: 'error',
                error: currentSession.error,
                result: currentSession.result // fallback result if available
            })}\n\n`);
            res.end();
            return true;
        }

        // 发送心跳保持连接
        res.write(`data: ${JSON.stringify({ type: 'heartbeat', status: currentSession.status })}\n\n`);
        return false;
    };

    // Send initial status
    res.write(`data: ${JSON.stringify({ type: 'start', status: 'generating' })}\n\n`);

    // Poll for updates every 200ms for smoother streaming
    const interval = setInterval(() => {
        if (sendUpdate()) {
            clearInterval(interval);
        }
    }, 200);

    req.on('close', () => {
        clearInterval(interval);
    });
});

/**
 * List all sessions (for debugging)
 * 需要认证
 */
app.get('/api/sessions', authMiddleware, (req, res) => {
    const list = [];
    for (const [id, session] of sessions) {
        list.push({
            id: session.id,
            prompt: session.prompt?.substring(0, 50),
            status: session.status,
            hasResult: !!session.result,
            createdAt: session.createdAt
        });
    }
    res.json({ sessions: list });
});

/**
 * Health check
 */
app.get('/health', (req, res) => {
    res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

// Clean up old sessions periodically
setInterval(() => {
    const now = Date.now();
    for (const [id, session] of sessions) {
        if (now - session.createdAt.getTime() > 30 * 60 * 1000) { // 30 minutes
            sessions.delete(id);
        }
    }
}, 5 * 60 * 1000); // Every 5 minutes

const PORT = process.env.PORT || 3001;
app.listen(PORT, () => {
    console.log(`\n🚀 Vibe Agent Server running on port ${PORT}`);
    console.log(`\nEndpoints:`);
    console.log(`  POST /api/generate     - Start page generation`);
    console.log(`  GET  /api/session/:id  - Get session status`);
    console.log(`  GET  /api/stream/:id   - Stream updates (SSE)`);
    console.log(`  GET  /health           - Health check\n`);
});
