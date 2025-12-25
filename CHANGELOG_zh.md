# 更新日志

本项目的所有重要变更都将记录在此文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
本项目遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [未发布]

### 计划中
- WebSocket 支持实时协作
- 多语言代码生成
- 项目导出为可下载的 HTML/CSS/JS
- 自定义主题选择

## [0.1.0] - 2024-12-22

### 新增
- Vibe Coding 首次发布
- 使用 Claude Agent SDK 的自然语言生成网页功能
- 用户认证（注册、登录、基于 JWT 的会话）
- 通过服务器发送事件 (SSE) 的实时代码流式传输
- 三层缓存系统（本地缓存、Redis、MySQL）
- 15+ 中间件组件（CORS、速率限制、JWT、指标等）
- Swagger API 文档
- Docker 支持和 docker-compose
- 带初始化脚本的 MySQL 数据库
- 完善的开发任务 Makefile
- 基于 GitHub Actions 的 CI/CD 流水线
- 支持 Go 1.22+ 和 Node.js 18+

### 安全
- 使用 bcrypt 的密码哈希（成本因子 12）
- JWT Token 认证
- 通过 GORM 防止 SQL 注入
- XSS 保护中间件
- CORS 配置
- 安全头（HSTS、X-Frame-Options、CSP）

### 性能
- 基准测试：
  - GET /ping: ~65,000 QPS, P99 1.2ms
  - GET /api/v1/users: ~44,000 QPS, P99 1.8ms
  - POST /api/v1/users: ~12,000 QPS, P99 3.5ms

### 文档
- 包含快速入门指南的综合 README
- 贡献指南
- Swagger API 文档
- 项目架构文档

## [0.0.1] - 2024-12-20

### 新增
- 项目初始化
- 基本 Go + Hertz 服务器设置
- 数据库连接层
- 初始中间件栈

---

## 版本格式

版本格式为 `MAJOR.MINOR.PATCH`：

- **MAJOR**：不兼容的 API 变更
- **MINOR**：向后兼容的功能新增
- **PATCH**：向后兼容的问题修复

## 变更类型

- `新增` - 新功能
- `变更` - 现有功能的变更
- `弃用` - 即将移除的功能
- `移除` - 已移除的功能
- `修复` - 问题修复
- `安全` - 安全漏洞修复
