# Contributing to Vibe Coding

Thank you for your interest in contributing to Vibe Coding! This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for everyone.

## How to Contribute

### Reporting Bugs

Before submitting a bug report:

1. Check the [existing issues](https://github.com/test-tt/vibe-coding/issues) to avoid duplicates
2. Ensure you're using the latest version
3. Collect relevant information (logs, screenshots, environment details)

When submitting a bug report, include:

- A clear, descriptive title
- Steps to reproduce the issue
- Expected vs actual behavior
- Environment details (OS, Go version, Node.js version)
- Relevant logs or error messages

### Suggesting Features

Feature requests are welcome! Please:

1. Check if the feature has already been requested
2. Clearly describe the use case and benefits
3. Consider if the feature aligns with the project's goals

### Pull Requests

#### Before You Start

1. Fork the repository
2. Create a feature branch from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. Set up your development environment (see README.md)

#### Development Workflow

1. **Make your changes**
   - Follow the existing code style
   - Write clear, concise commit messages
   - Keep changes focused and atomic

2. **Test your changes**
   ```bash
   # Run tests
   make test

   # Run linter
   make lint

   # Format code
   make fmt
   ```

3. **Commit your changes**
   - Use conventional commit messages:
     - `feat:` for new features
     - `fix:` for bug fixes
     - `docs:` for documentation changes
     - `refactor:` for code refactoring
     - `test:` for adding tests
     - `chore:` for maintenance tasks

   Example:
   ```bash
   git commit -m "feat: add user avatar upload support"
   ```

4. **Push and create PR**
   ```bash
   git push origin feature/your-feature-name
   ```
   Then open a Pull Request on GitHub.

#### PR Guidelines

- Provide a clear description of the changes
- Reference any related issues
- Ensure all CI checks pass
- Respond to review feedback promptly
- Keep the PR focused on a single concern

## Development Setup

### Prerequisites

- Go >= 1.22
- Node.js >= 18
- MySQL >= 8.0
- Redis >= 7.0
- Make

### Local Development

```bash
# Clone your fork
git clone https://github.com/YOUR-USERNAME/vibe-coding.git
cd vibe-coding

# Add upstream remote
git remote add upstream https://github.com/test-tt/vibe-coding.git

# Install dependencies
make tidy
cd agent-server && npm install && cd ..

# Initialize database
mysql -u root -p < scripts/init.sql

# Start development servers
make dev          # Terminal 1: Go backend
make agent-dev    # Terminal 2: Agent server
```

### Running Tests

```bash
# Unit tests
make test

# With coverage
make test-cover

# Specific package
go test ./internal/service/...
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Generate Swagger docs (if API changed)
make swagger
```

## Project Structure

```
.
├── cmd/api/           # Application entrypoint
├── config/            # Configuration files
├── internal/          # Private application code
│   ├── handler/       # HTTP handlers
│   ├── service/       # Business logic
│   ├── dao/           # Data access
│   ├── model/         # Data models
│   ├── middleware/    # HTTP middleware
│   └── router/        # Route definitions
├── pkg/               # Public libraries
├── agent-server/      # Node.js Agent Server
├── web/               # Frontend files
├── scripts/           # Database scripts
└── docs/              # Documentation
```

## Coding Standards

### Go Code

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Keep functions small and focused
- Add comments for exported functions
- Handle errors explicitly
- Use meaningful variable names

Example:
```go
// CreateUser creates a new user with the given information.
// Returns the created user and any error encountered.
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    if err := validate.Struct(req); err != nil {
        return nil, errcode.ErrInvalidParams.WithMessage(err.Error())
    }

    user := &model.User{
        Name:  req.Name,
        Email: req.Email,
    }

    if err := s.dao.Create(ctx, user); err != nil {
        return nil, err
    }

    return user, nil
}
```

### JavaScript/Node.js Code

- Use ES6+ features
- Follow existing code patterns
- Add JSDoc comments for functions
- Handle async errors properly

### Database Changes

- Add migration scripts to `scripts/`
- Use descriptive names: `migrate_add_feature.sql`
- Document breaking changes
- Test migrations locally first

## Getting Help

- Open an issue for bugs or feature requests
- Check existing documentation in the repository
- Review closed issues for similar problems

## Recognition

Contributors will be acknowledged in:
- The project README
- Release notes for significant contributions

Thank you for contributing to Vibe Coding!
