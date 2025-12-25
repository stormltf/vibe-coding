# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- Websocket support for real-time collaboration
- Multi-language support for code generation
- Project export as downloadable HTML/CSS/JS
- Custom theme selection

## [0.1.0] - 2024-12-22

### Added
- Initial release of Vibe Coding
- Natural language to web page generation using Claude Agent SDK
- User authentication (register, login, JWT-based sessions)
- Real-time code streaming via Server-Sent Events (SSE)
- Three-tier caching system (LocalCache, Redis, MySQL)
- 15+ middleware components (CORS, RateLimit, JWT, Metrics, etc.)
- Swagger API documentation
- Docker support with docker-compose
- MySQL database with initialization scripts
- Comprehensive Makefile for development tasks
- CI/CD pipeline with GitHub Actions
- Support for Go 1.22+ and Node.js 18+

### Security
- Password hashing with bcrypt (cost factor 12)
- JWT token authentication
- SQL injection protection via GORM
- XSS protection middleware
- CORS configuration
- Security headers (HSTS, X-Frame-Options, CSP)

### Performance
- Benchmarks:
  - GET /ping: ~65,000 QPS, P99 1.2ms
  - GET /api/v1/users: ~44,000 QPS, P99 1.8ms
  - POST /api/v1/users: ~12,000 QPS, P99 3.5ms

### Documentation
- Comprehensive README with quick start guide
- Contributing guidelines
- API documentation via Swagger
- Project architecture documentation

## [0.0.1] - 2024-12-20

### Added
- Project initialization
- Basic Go + Hertz server setup
- Database connection layer
- Initial middleware stack

---

## Version Format

The version format is `MAJOR.MINOR.PATCH`:

- **MAJOR**: Incompatible API changes
- **MINOR**: Backward-compatible functionality additions
- **PATCH**: Backward-compatible bug fixes

## Types of Changes

- `Added` - New features
- `Changed` - Changes in existing functionality
- `Deprecated` - Soon-to-be removed features
- `Removed` - Removed features
- `Fixed` - Bug fixes
- `Security` - Security vulnerability fixes
