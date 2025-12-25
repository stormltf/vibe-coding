# Vibe Coding

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Node.js](https://img.shields.io/badge/Node.js-18+-339933?style=flat&logo=node.js)](https://nodejs.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/test-tt/test-tt/actions/workflows/ci.yml/badge.svg)](https://github.com/test-tt/test-tt/actions)

AI-powered web page generation platform built with **Claude Agent SDK**. Describe what you want in natural language, and let AI generate modern web pages for you.

English | [简体中文](README_zh.md)

## Features

- **Natural Language to Code** - Describe your idea, get production-ready HTML/CSS
- **Real-time Streaming** - Watch code generation in real-time via SSE
- **Modern Tech Stack** - Go + Hertz backend, Node.js Agent Server, Claude AI
- **User Authentication** - JWT-based auth with secure password handling
- **Project Persistence** - Save and manage your generated projects

## Quick Start

### Prerequisites

| Requirement | Version | Installation |
|------------|---------|--------------|
| Go | >= 1.22 | [Download](https://go.dev/dl/) |
| Node.js | >= 18 | [Download](https://nodejs.org/) |
| MySQL | >= 8.0 | `brew install mysql` or [Download](https://dev.mysql.com/downloads/) |
| Redis | >= 7.0 | `brew install redis` or [Download](https://redis.io/download/) |
| Claude API Key | - | [Get API Key](https://console.anthropic.com/) |

### Step 1: Clone & Install Dependencies

```bash
# Clone the repository
git clone https://github.com/test-tt/vibe-coding.git
cd vibe-coding

# Install Go dependencies
make tidy

# Install Node.js dependencies for Agent Server
cd agent-server && npm install && cd ..
```

### Step 2: Start MySQL & Redis

```bash
# macOS (using Homebrew)
brew services start mysql
brew services start redis

# Or using Docker (skip to Step 4 for full Docker setup)
```

### Step 3: Initialize Database

```bash
# Run the initialization script
mysql -u root -p < scripts/init.sql
```

This creates the `vibe_coding` database with all required tables and test data.

**Test Account:**
- Email: `test@example.com`
- Password: `password123`

### Step 4: Configure API Key

```bash
# Set your Claude API key
export ANTHROPIC_API_KEY="your-api-key-here"
```

### Step 5: Start Services

Open **two terminal windows**:

**Terminal 1 - Go Backend:**
```bash
make dev
# Server runs at http://localhost:8888
```

**Terminal 2 - Agent Server:**
```bash
make agent
# Agent runs at http://localhost:3001
```

### Step 6: Access the Application

- **Homepage:** http://localhost:8888
- **AI Workspace:** http://localhost:8888/workspace.html
- **API Docs:** http://localhost:8888/swagger/index.html

---

## Docker Deployment (Alternative)

Skip Steps 2-5 and use Docker instead:

```bash
# Set API key
export ANTHROPIC_API_KEY="your-api-key-here"

# Start all services (MySQL + Redis + API)
make docker-up

# View logs
make docker-logs

# Stop services
make docker-down
```

---

## Project Structure

```
.
├── agent-server/          # Claude Agent SDK service (Node.js)
│   ├── server.js          # Express server with SSE streaming
│   └── package.json
├── cmd/api/               # Go application entrypoint
│   └── main.go
├── config/                # Configuration files
│   ├── config.yaml        # Default config
│   ├── config.dev.yaml    # Development config
│   └── config.prod.yaml   # Production config
├── internal/              # Internal packages
│   ├── handler/           # HTTP request handlers
│   ├── service/           # Business logic
│   ├── dao/               # Data access layer
│   ├── model/             # Data models
│   ├── middleware/        # HTTP middlewares (15+)
│   └── router/            # Route definitions
├── pkg/                   # Public packages
│   ├── cache/             # Three-tier caching system
│   ├── database/          # MySQL wrapper
│   ├── logger/            # Zap-based logging
│   ├── jwt/               # JWT utilities
│   └── ...
├── web/                   # Frontend (HTML/CSS/JS)
│   ├── index.html         # Homepage
│   └── workspace.html     # AI Workspace
├── scripts/               # Database scripts
│   └── init.sql           # Complete DB initialization
├── docker-compose.yaml    # Docker orchestration
├── Makefile               # Build commands
└── README.md
```

---

## API Reference

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | Register new user |
| POST | `/api/v1/auth/login` | Login and get JWT token |
| POST | `/api/v1/auth/logout` | Logout (invalidate token) |
| GET | `/api/v1/auth/profile` | Get current user profile |
| PUT | `/api/v1/auth/password` | Change password |

### AI Generation (Agent Server)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/generate` | Create generation task |
| GET | `/api/stream/:id` | SSE stream for results |
| GET | `/api/sessions` | List all sessions |

### Example: Login

```bash
curl -X POST http://localhost:8888/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "password123"}'
```

Response:
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

### Example: Generate Web Page

```bash
curl -X POST http://localhost:3001/api/generate \
  -H "Authorization: Bearer <your-token>" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Create a modern SaaS landing page with hero section"}'
```

---

## Make Commands

```bash
# Development
make dev              # Start with hot reload (Air)
make run              # Run directly
make run-dev          # Run with dev config
make run-prod         # Run with prod config

# Agent Server
make agent            # Start Agent Server
make agent-dev        # Start with auto-reload
make agent-install    # Install npm dependencies

# Build
make build            # Build binary
make swagger          # Generate Swagger docs

# Testing
make test             # Run tests
make test-cover       # Run with coverage

# Code Quality
make lint             # Run golangci-lint
make fmt              # Format code

# Docker
make docker-up        # Start all services
make docker-down      # Stop all services
make docker-logs      # View logs

# Help
make help             # Show all commands
```

---

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `ANTHROPIC_API_KEY` | Claude API key | (required) |
| `APP_SERVER_PORT` | Server port | 8888 |
| `APP_MYSQL_HOST` | MySQL host | 127.0.0.1 |
| `APP_MYSQL_DATABASE` | Database name | vibe_coding |
| `APP_REDIS_HOST` | Redis host | 127.0.0.1 |
| `JWT_SECRET` | JWT signing secret | (see config) |

### Config Files

- `config/config.yaml` - Default configuration
- `config/config.dev.yaml` - Development overrides
- `config/config.prod.yaml` - Production overrides

---

## Architecture Highlights

### Three-tier Caching

```
Request → L1 LocalCache (64MB, ~50μs)
              ↓ miss
          L2 Redis (~1ms)
              ↓ miss
          L3 MySQL (~5ms)
              ↓
          Backfill L2 + L1
```

**Protection mechanisms:**
- Cache penetration: Bloom filter + null value caching
- Cache breakdown: Singleflight request merging
- Cache avalanche: TTL randomization

### Middleware Stack (15+)

- Recovery, RequestID, AccessLog, CORS
- RateLimit (3-tier), JWT Auth
- Prometheus Metrics, Gzip, Timeout
- OpenTelemetry Tracing, Circuit Breaker
- I18n, Security Headers

### Performance Benchmarks

| Endpoint | QPS | P99 Latency |
|----------|-----|-------------|
| GET /ping | ~65,000 | 1.2ms |
| GET /api/v1/users | ~44,000 | 1.8ms |
| POST /api/v1/users | ~12,000 | 3.5ms |

*Tested on Apple M4 Pro / 16GB / macOS*

---

## Development Guide

### Adding a New API Endpoint

1. Define model in `internal/model/`
2. Implement DAO in `internal/dao/`
3. Add business logic in `internal/service/`
4. Create handler in `internal/handler/` (with Swagger annotations)
5. Register route in `internal/router/router.go`
6. Run `make swagger` to update docs

### Code Style

```bash
# Before committing
make fmt              # Format code
make lint             # Check issues
make test             # Run tests
```

---

## Troubleshooting

### Port Already in Use

```bash
# Kill process on port 8888
kill -9 $(lsof -t -i:8888)

# Kill process on port 3001
kill -9 $(lsof -t -i:3001)
```

### Database Connection Failed

1. Ensure MySQL is running: `mysql.server status`
2. Check credentials in `config/config.dev.yaml`
3. Verify database exists: `mysql -u root -p -e "SHOW DATABASES;"`

### Redis Connection Failed

1. Ensure Redis is running: `redis-cli ping`
2. Check connection in config file

### Agent Server JWT Error

Ensure JWT secret matches between Go backend and Agent Server:

```bash
# Start Agent Server with matching JWT config
make agent
```

---

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- [Claude Agent SDK](https://www.npmjs.com/package/@anthropic-ai/claude-agent-sdk) by Anthropic
- [Hertz](https://github.com/cloudwego/hertz) by ByteDance
- [GORM](https://gorm.io/) for database operations
- All other open-source libraries that make this project possible
