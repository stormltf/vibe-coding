# Architecture

This document provides an in-depth look at the Vibe Coding system architecture, design decisions, and component interactions.

## Table of Contents

- [Overview](#overview)
- [System Architecture](#system-architecture)
- [Component Details](#component-details)
- [Data Flow](#data-flow)
- [Design Patterns](#design-patterns)
- [Technology Choices](#technology-choices)

## Overview

Vibe Coding is a microservices application that generates web pages using AI. The system consists of two main services:

1. **API Server** (Go) - Handles authentication, project management, and serving the frontend
2. **Agent Server** (Node.js) - Handles AI code generation using Claude Agent SDK

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                            Browser                              │
└───────────────────────────────┬─────────────────────────────────┘
                                │ HTTP/WebSocket
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Load Balancer                           │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                ┌───────────────┴───────────────┐
                ▼                               ▼
┌──────────────────────────┐      ┌──────────────────────────┐
│      API Server          │      │     Agent Server         │
│      (Go / Hertz)        │◄────►│    (Node.js / Express)   │
│                          │      │                          │
│  - Router                │      │  - Claude Agent SDK      │
│  - Middleware            │      │  - SSE Streaming         │
│  - Handlers              │      │  - Session Management    │
│  - Services              │      │                          │
│  - DAO Layer             │      │                          │
└────────┬─────────────────┘      └──────────────────────────┘
         │
    ┌────┴────────────────────────────┐
    ▼                                 ▼
┌─────────────────┐         ┌─────────────────┐
│  Redis Cache    │         │  MySQL Database │
│                 │         │                 │
│  - Session Data │         │  - Users        │
│  - Hot Data     │         │  - Projects     │
│  - Rate Limit   │         │  - Generations  │
└─────────────────┘         └─────────────────┘
```

## Component Details

### API Server (Go)

**Framework**: [CloudWeGo Hertz](https://github.com/cloudwego/hertz)

**Key Components**:

| Package | Responsibility |
|---------|----------------|
| `cmd/api` | Application entry point |
| `internal/router` | Route definitions and middleware setup |
| `internal/handler` | HTTP request handlers |
| `internal/service` | Business logic layer |
| `internal/dao` | Data access layer |
| `internal/model` | Data models and structs |
| `internal/middleware` | Custom middleware implementations |
| `pkg/` | Reusable packages (cache, database, logger, etc.) |

### Agent Server (Node.js)

**Framework**: Express.js with SSE support

**Key Components**:

| Module | Responsibility |
|--------|----------------|
| `server.js` | Express server setup |
| `/api/generate` | Create new generation task |
| `/api/stream/:id` | SSE endpoint for real-time streaming |
| `/api/sessions` | List all generation sessions |

### Database Layer

**MySQL Schema**:

```sql
users          -- User accounts and authentication data
projects       -- Generated web page projects
generations    -- AI generation task records
sessions       -- Active session storage
```

**Cache Layer (Redis)**:

- User sessions (JWT token blacklist)
- Rate limiting counters
- Hot data caching (three-tier cache)

## Data Flow

### Authentication Flow

```
User ──POST /api/v1/auth/login──► Router
                                    │
                                    ▼
                              Middleware
                                    │
                          ┌─────────┴─────────┐
                          ▼                   ▼
                    JWT Validation       Rate Limit
                          │                   │
                          └─────────┬─────────┘
                                    ▼
                              Auth Handler
                                    │
                          ┌─────────┴─────────┐
                          ▼                   ▼
                    Verify Password     Generate JWT
                          │                   │
                          └─────────┬─────────┘
                                    ▼
                              Return Token
                                    │
                                    ▼
                                User
```

### Code Generation Flow

```
User ──POST /api/generate──► Agent Server
                                │
                                ▼
                        Create Session ID
                                │
                                ▼
                        Return Session ID
                                │
                    ┌───────────┴───────────┐
                    ▼                       ▼
              User connects to            Claude
              SSE /api/stream/:id      Agent SDK
                    │                       │
                    └───────────┬───────────┘
                                ▼
                        Stream Generated
                        Code to User
                                │
                                ▼
                            Complete
```

## Design Patterns

### Repository Pattern

The DAO layer implements the Repository pattern, abstracting database operations:

```go
type UserDAO interface {
    Create(ctx context.Context, user *model.User) error
    FindByID(ctx context.Context, id int64) (*model.User, error)
    FindByEmail(ctx context.Context, email string) (*model.User, error)
    Update(ctx context.Context, user *model.User) error
    Delete(ctx context.Context, id int64) error
}
```

### Middleware Chain Pattern

Hertz uses middleware chains for cross-cutting concerns:

```
Request → Recovery → RequestID → AccessLog → CORS → RateLimit → JWT → Handler
```

### Three-Tier Caching Pattern

```
Request → L1 (LocalCache) → L2 (Redis) → L3 (MySQL) → Backfill
           ↓ miss              ↓ miss
```

**Features**:
- Cache penetration protection (Bloom filter + null caching)
- Cache breakdown protection (Singleflight)
- Cache avalanche protection (TTL randomization)

### Dependency Injection

Services receive dependencies through constructors:

```go
type UserService struct {
    dao  dao.UserDAO
    cache cache.Cache
}

func NewUserService(dao dao.UserDAO, cache cache.Cache) *UserService {
    return &UserService{dao: dao, cache: cache}
}
```

## Technology Choices

### Why Go (Hertz)?

| Reason | Description |
|--------|-------------|
| Performance | Compiled language, excellent concurrency |
| Type Safety | Compile-time error detection |
| Ecosystem | Rich standard library and third-party packages |
| Deployment | Single binary, easy cross-compilation |
| Hertz | High-performance HTTP framework by ByteDance |

### Why Node.js (Agent Server)?

| Reason | Description |
|--------|-------------|
| Claude SDK | Official Node.js SDK for Claude Agent |
| Async I/O | Natural fit for streaming responses |
| SSE Support | Excellent support for Server-Sent Events |
| Development | Rapid iteration for AI integration |

### Why MySQL?

| Reason | Description |
|--------|-------------|
| Maturity | Proven, battle-tested database |
| ACID | Reliable transaction support |
| Ecosystem | Wide tooling and expertise availability |
| GORM | Excellent Go ORM with MySQL support |

### Why Redis?

| Reason | Description |
|--------|-------------|
| Speed | In-memory operations |
| Data Structures | Rich set of data types |
| Persistence | Optional disk persistence |
| Ecosystem | Widely used and well-documented |

## Security Architecture

### Authentication

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ POST /api/v1/auth/login
       ▼
┌─────────────────┐
│   Verify        │
│   Credentials   │
└────────┬────────┘
         │
    ┌────┴────┐
    ▼         ▼
┌──────┐  ┌──────┐
│  DB  │  │bcrypt│
└───┬──┘  └───┬──┘
    │        │
    └────┬───┘
         ▼
    ┌─────────┐
    │  Issue  │
    │   JWT   │
    └────┬────┘
         │
         ▼
    ┌─────────┐
    │ Return  │
    │  Token  │
    └─────────┘
```

### Authorization

- JWT-based stateless authentication
- Role-based access control (future)
- Route-level middleware protection

## Performance Considerations

### Benchmarks

Tested on Apple M4 Pro / 16GB / macOS:

| Endpoint | QPS | P99 Latency |
|----------|-----|-------------|
| GET /ping | ~65,000 | 1.2ms |
| GET /api/v1/users | ~44,000 | 1.8ms |
| POST /api/v1/users | ~12,000 | 3.5ms |

### Optimization Strategies

1. **Connection Pooling**: Database connection reuse
2. **Caching**: Three-tier cache for hot data
3. **Rate Limiting**: Protect against abuse
4. **Singleflight**: Merge concurrent requests
5. **Response Compression**: Gzip middleware

## Future Architecture Considerations

### Potential Improvements

- **Message Queue**: For asynchronous task processing
- **Distributed Tracing**: OpenTelemetry implementation
- **Service Mesh**: For microservice communication
- **Read Replicas**: For database scaling
- **CDN**: For static asset delivery
- **WebSocket**: For real-time collaboration
- **Kubernetes**: For orchestration and scaling

For more information, see the [README](../README.md) and [CONTRIBUTING](../CONTRIBUTING.md).
