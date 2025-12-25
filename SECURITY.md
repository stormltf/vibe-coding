# Security Policy

## Supported Versions

| Version | Supported          |
|---------|--------------------|
| Latest  | :white_check_mark: |
| Older   | :x:                |

We provide security updates only for the latest version of Vibe Coding.

## Reporting a Vulnerability

If you discover a security vulnerability, please **DO NOT** open a public issue.

### How to Report

1. Send an email to: **security@vibe-coding.dev**
2. Include:
   - Description of the vulnerability
   - Steps to reproduce the issue
   - Potential impact
   - Suggested fix (if any)

### Response Timeline

- **Within 48 hours**: Initial response acknowledging receipt
- **Within 7 days**: Assessment and proposed remediation plan
- **Within 14 days**: Patch release for critical vulnerabilities
- **Within 30 days**: Patch release for non-critical vulnerabilities

### What Happens Next

1. We will verify the vulnerability
2. We will determine severity (Critical/High/Medium/Low)
3. We will develop a fix
4. We will coordinate release with you (if desired)
5. We will publicly disclose after the fix is deployed

## Security Best Practices for Users

### Environment Variables

- Never commit `ANTHROPIC_API_KEY` to version control
- Use strong, unique `JWT_SECRET` in production
- Rotate secrets regularly

### Database Security

- Change default MySQL passwords in production
- Restrict database access to localhost only
- Enable SSL for database connections in production
- Regularly update MySQL to the latest version

### Deployment

- Keep dependencies updated (`go get -u ./...`)
- Use HTTPS in production (TLS 1.3+)
- Enable firewall rules to restrict access
- Enable security headers via middleware
- Regularly audit logs for suspicious activity

### API Security

- Always validate and sanitize user input
- Use parameterized queries to prevent SQL injection
- Enable rate limiting to prevent abuse
- Implement CORS properly
- Keep API keys secret

## Security Features

Vibe Coding includes several built-in security features:

| Feature | Description |
|---------|-------------|
| JWT Authentication | Token-based auth with configurable expiration |
| Password Hashing | bcrypt with cost factor 12 |
| Rate Limiting | Three-tier rate limiting (IP, user, API) |
| CORS Protection | Configurable origin whitelist |
| Security Headers | HSTS, X-Frame-Options, CSP, etc. |
| SQL Injection Protection | GORM parameterized queries |
| XSS Protection | Input sanitization and output encoding |
| Request Validation | Struct validation for all inputs |

## Dependency Scanning

We use GitHub Dependabot to automatically monitor and update dependencies. Security advisories are tracked via:
- [GitHub Security Advisories](https://github.com/test-tt/vibe-coding/security/advisories)
- [Go Vulnerability Database](https://pkg.go.dev/vuln/)

## Reaching Out

For security-related questions that are not vulnerability reports:
- Open a discussion with the `security` label
- Email: security@vibe-coding.dev

## Security Audits

This project has not yet undergone a professional security audit. We welcome contributions from security researchers and encourage responsible disclosure.

### Key Areas for Review

If you're conducting a security review, please focus on:
- Authentication and authorization flows
- Input validation and sanitization
- Database query construction
- File upload handling (if any)
- Session management
- API rate limiting
- Secret management
