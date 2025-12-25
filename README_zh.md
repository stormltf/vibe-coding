# Vibe Coding

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Node.js](https://img.shields.io/badge/Node.js-18+-339933?style=flat&logo=node.js)](https://nodejs.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/test-tt/test-tt/actions/workflows/ci.yml/badge.svg)](https://github.com/test-tt/test-tt/actions)

基于 **Claude Agent SDK** 构建的 AI 网页生成平台。用自然语言描述你想要的内容，让 AI 为你生成现代化的网页。

[English](README.md) | 简体中文

## 功能特性

- **自然语言生成代码** - 描述需求，获得生产级 HTML/CSS
- **实时流式输出** - 通过 SSE 实时观看代码生成过程
- **现代化技术栈** - Go + Hertz 后端，Node.js Agent 服务，Claude AI
- **用户认证系统** - JWT 认证 + 安全密码处理
- **项目管理** - 保存和管理生成的项目

## 快速开始

### 前置要求

| 要求 | 版本 | 安装 |
|------------|---------|--------------|
| Go | >= 1.22 | [下载](https://go.dev/dl/) |
| Node.js | >= 18 | [下载](https://nodejs.org/) |
| MySQL | >= 8.0 | `brew install mysql` 或 [下载](https://dev.mysql.com/downloads/) |
| Redis | >= 7.0 | `brew install redis` 或 [下载](https://redis.io/download/) |
| Claude API Key | - | [获取 API Key](https://console.anthropic.com/) |

### 步骤 1：克隆并安装依赖

```bash
# 克隆仓库
git clone https://github.com/test-tt/vibe-coding.git
cd vibe-coding

# 安装 Go 依赖
make tidy

# 安装 Agent Server 的 Node.js 依赖
cd agent-server && npm install && cd ..
```

### 步骤 2：启动 MySQL 和 Redis

```bash
# macOS (使用 Homebrew)
brew services start mysql
brew services start redis

# 或使用 Docker (跳转到步骤 4 进行完整 Docker 设置)
```

### 步骤 3：初始化数据库

```bash
# 运行初始化脚本
mysql -u root -p < scripts/init.sql
```

这将创建 `vibe_coding` 数据库及所有必需的表和测试数据。

**测试账号：**
- 邮箱：`test@example.com`
- 密码：`password123`

### 步骤 4：配置 API Key

```bash
# 设置 Claude API Key
export ANTHROPIC_API_KEY="your-api-key-here"
```

### 步骤 5：启动服务

打开 **两个终端窗口**：

**终端 1 - Go 后端：**
```bash
make dev
# 服务运行在 http://localhost:8888
```

**终端 2 - Agent 服务：**
```bash
make agent
# Agent 运行在 http://localhost:3001
```

### 步骤 6：访问应用

- **首页：** http://localhost:8888
- **AI 工作区：** http://localhost:8888/workspace.html
- **API 文档：** http://localhost:8888/swagger/index.html

---

## Docker 部署（替代方案）

跳过步骤 2-5，直接使用 Docker：

```bash
# 设置 API Key
export ANTHROPIC_API_KEY="your-api-key-here"

# 启动所有服务（MySQL + Redis + API）
make docker-up

# 查看日志
make docker-logs

# 停止服务
make docker-down
```

---

## 项目结构

```
.
├── agent-server/          # Claude Agent SDK 服务 (Node.js)
│   ├── server.js          # Express 服务器，支持 SSE 流式输出
│   └── package.json
├── cmd/api/               # Go 应用入口
│   └── main.go
├── config/                # 配置文件
│   ├── config.yaml        # 默认配置
│   ├── config.dev.yaml    # 开发环境配置
│   └── config.prod.yaml   # 生产环境配置
├── internal/              # 内部包
│   ├── handler/           # HTTP 请求处理器
│   ├── service/           # 业务逻辑
│   ├── dao/               # 数据访问层
│   ├── model/             # 数据模型
│   ├── middleware/        # HTTP 中间件 (15+)
│   └── router/            # 路由定义
├── pkg/                   # 公共包
│   ├── cache/             # 三层缓存系统
│   ├── database/          # MySQL 封装
│   ├── logger/            # Zap 日志
│   ├── jwt/               # JWT 工具
│   └── ...
├── web/                   # 前端 (HTML/CSS/JS)
│   ├── index.html         # 首页
│   └── workspace.html     # AI 工作区
├── scripts/               # 数据库脚本
│   └── init.sql           # 完整数据库初始化
├── docker-compose.yaml    # Docker 编排
├── Makefile               # 构建命令
└── README.md
```

---

## API 参考

### 认证接口

| 方法 | 端点 | 描述 |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | 注册新用户 |
| POST | `/api/v1/auth/login` | 登录并获取 JWT token |
| POST | `/api/v1/auth/logout` | 登出（使 token 失效） |
| GET | `/api/v1/auth/profile` | 获取当前用户信息 |
| PUT | `/api/v1/auth/password` | 修改密码 |

### AI 生成接口（Agent 服务）

| 方法 | 端点 | 描述 |
|--------|----------|-------------|
| POST | `/api/generate` | 创建生成任务 |
| GET | `/api/stream/:id` | SSE 结果流 |
| GET | `/api/sessions` | 列出所有会话 |

### 示例：登录

```bash
curl -X POST http://localhost:8888/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "password123"}'
```

响应：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "name": "Test User",
      "email": "test@example.com"
    }
  }
}
```

### 示例：生成网页

```bash
curl -X POST http://localhost:3001/api/generate \
  -H "Authorization: Bearer <your-token>" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "创建一个现代风格的 SaaS 落地页，包含英雄区块"}'
```

---

## Make 命令

```bash
# 开发
make dev              # 使用热重载启动 (Air)
make run              # 直接运行
make run-dev          # 使用开发配置运行
make run-prod         # 使用生产配置运行

# Agent 服务
make agent            # 启动 Agent 服务
make agent-dev        # 使用自动重载启动
make agent-install    # 安装 npm 依赖

# 构建
make build            # 构建二进制文件
make swagger          # 生成 Swagger 文档

# 测试
make test             # 运行测试
make test-cover       # 运行测试并生成覆盖率报告

# 代码质量
make lint             # 运行 golangci-lint
make fmt              # 格式化代码

# Docker
make docker-up        # 启动所有服务
make docker-down      # 停止所有服务
make docker-logs      # 查看日志

# 帮助
make help             # 显示所有命令
```

---

## 配置

### 环境变量

| 变量 | 描述 | 默认值 |
|----------|-------------|---------|
| `ANTHROPIC_API_KEY` | Claude API 密钥 | (必需) |
| `APP_SERVER_PORT` | 服务端口 | 8888 |
| `APP_MYSQL_HOST` | MySQL 主机 | 127.0.0.1 |
| `APP_MYSQL_DATABASE` | 数据库名 | vibe_coding |
| `APP_REDIS_HOST` | Redis 主机 | 127.0.0.1 |
| `JWT_SECRET` | JWT 签名密钥 | (见配置文件) |

### 配置文件

- `config/config.yaml` - 默认配置
- `config/config.dev.yaml` - 开发环境覆盖配置
- `config/config.prod.yaml` - 生产环境覆盖配置

---

## 架构亮点

### 三层缓存

```
请求 → L1 本地缓存 (64MB, ~50μs)
         ↓ 未命中
     L2 Redis (~1ms)
         ↓ 未命中
     L3 MySQL (~5ms)
         ↓
     回填 L2 + L1
```

**保护机制：**
- 缓存穿透：布隆过滤器 + 空值缓存
- 缓存击穿：Singleflight 请求合并
- 缓存雪崩：TTL 随机化

### 中间件栈 (15+)

- 恢复、请求ID、访问日志、CORS
- 三层限流、JWT 认证
- Prometheus 指标、Gzip、超时
- OpenTelemetry 链路追踪、熔断器
- 国际化、安全头

### 性能基准

| 端点 | QPS | P99 延迟 |
|----------|-----|-------------|
| GET /ping | ~65,000 | 1.2ms |
| GET /api/v1/users | ~44,000 | 1.8ms |
| POST /api/v1/users | ~12,000 | 3.5ms |

*测试环境：Apple M4 Pro / 16GB / macOS*

---

## 开发指南

### 添加新的 API 端点

1. 在 `internal/model/` 中定义模型
2. 在 `internal/dao/` 中实现 DAO
3. 在 `internal/service/` 中添加业务逻辑
4. 在 `internal/handler/` 中创建处理器（带 Swagger 注解）
5. 在 `internal/router/router.go` 中注册路由
6. 运行 `make swagger` 更新文档

### 代码规范

```bash
# 提交前检查
make fmt              # 格式化代码
make lint             # 检查问题
make test             # 运行测试
```

---

## 故障排除

### 端口已被占用

```bash
# 终止占用 8888 端口的进程
kill -9 $(lsof -t -i:8888)

# 终止占用 3001 端口的进程
kill -9 $(lsof -t -i:3001)
```

### 数据库连接失败

1. 确保 MySQL 正在运行：`mysql.server status`
2. 检查 `config/config.dev.yaml` 中的凭据
3. 验证数据库是否存在：`mysql -u root -p -e "SHOW DATABASES;"`

### Redis 连接失败

1. 确保 Redis 正在运行：`redis-cli ping`
2. 检查配置文件中的连接信息

### Agent 服务 JWT 错误

确保 Go 后端和 Agent 服务之间的 JWT 密钥匹配：

```bash
# 使用匹配的 JWT 配置启动 Agent 服务
make agent
```

---

## 贡献

我们欢迎贡献！请参阅 [CONTRIBUTING.md](CONTRIBUTING.md) 了解指南。

1. Fork 仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开启 Pull Request

[贡献指南](CONTRIBUTING_zh.md) | [安全政策](SECURITY_zh.md) | [行为准则](CODE_OF_CONDUCT_zh.md)

---

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

---

## 致谢

- [Claude Agent SDK](https://www.npmjs.com/package/@anthropic-ai/claude-agent-sdk) by Anthropic
- [Hertz](https://github.com/cloudwego/hertz) by ByteDance
- [GORM](https://gorm.io/) 数据库操作
- 所有其他使本项目成为可能的开源库
