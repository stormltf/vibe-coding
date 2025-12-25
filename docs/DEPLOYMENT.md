# Deployment Guide

This guide covers deploying Vibe Coding to production environments.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Environment Configuration](#environment-configuration)
- [Docker Deployment](#docker-deployment)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Manual Deployment](#manual-deployment)
- [Monitoring](#monitoring)
- [Backup and Recovery](#backup-and-recovery)

## Prerequisites

### System Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| CPU | 2 cores | 4+ cores |
| RAM | 4 GB | 8+ GB |
| Disk | 20 GB | 50+ GB SSD |
| MySQL | 8.0+ | 8.0+ |
| Redis | 7.0+ | 7.0+ |

### Required Environment Variables

```bash
# Application
APP_ENV=production
APP_SERVER_PORT=8888
APP_SERVER_HOST=0.0.0.0

# Database
APP_MYSQL_HOST=your-mysql-host
APP_MYSQL_PORT=3306
APP_MYSQL_USER=vibe_coding
APP_MYSQL_PASSWORD=your-secure-password
APP_MYSQL_DATABASE=vibe_coding

# Redis
APP_REDIS_HOST=your-redis-host
APP_REDIS_PORT=6379
APP_REDIS_PASSWORD=your-redis-password

# Security
JWT_SECRET=your-jwt-secret-min-32-chars
ANTHROPIC_API_KEY=your-claude-api-key

# Logging
LOG_LEVEL=info
LOG_OUTPUT=/var/log/vibe-coding/app.log
```

## Environment Configuration

### 1. Production Config File

Create `config/config.prod.yaml`:

```yaml
server:
  address: ":8888"
  timeout: 30s

mysql:
  host: "${APP_MYSQL_HOST}"
  port: ${APP_MYSQL_PORT}
  user: "${APP_MYSQL_USER}"
  password: "${APP_MYSQL_PASSWORD}"
  database: "${APP_MYSQL_DATABASE}"
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600

redis:
  host: "${APP_REDIS_HOST}"
  port: ${APP_REDIS_PORT}
  password: "${APP_REDIS_PASSWORD}"
  pool_size: 100

jwt:
  secret: "${JWT_SECRET}"
  expire: 168h  # 7 days

rate_limit:
  enabled: true
  qps: 1000
  burst: 2000
```

## Docker Deployment

### Using Docker Compose (Recommended)

1. **Set environment variables**:

```bash
export ANTHROPIC_API_KEY="your-api-key"
export JWT_SECRET="$(openssl rand -base64 32)"
export MYSQL_PASSWORD="$(openssl rand -base64 16)"
```

2. **Update docker-compose.yaml** for production:

```yaml
version: '3.8'

services:
  api:
    image: vibe-coding:latest
    restart: always
    ports:
      - "8888:8888"
    environment:
      - APP_ENV=production
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      - mysql
      - redis

  agent:
    image: vibe-coding-agent:latest
    restart: always
    ports:
      - "3001:3001"
    environment:
      - JWT_SECRET=${JWT_SECRET}
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}

  mysql:
    image: mysql:8.0
    restart: always
    environment:
      - MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD}
      - MYSQL_DATABASE=vibe_coding
      - MYSQL_USER=vibe_coding
      - MYSQL_PASSWORD=${MYSQL_PASSWORD}
    volumes:
      - mysql_data:/var/lib/mysql

  redis:
    image: redis:7-alpine
    restart: always
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data

volumes:
  mysql_data:
  redis_data:
```

3. **Start services**:

```bash
docker-compose up -d
```

4. **Verify deployment**:

```bash
docker-compose ps
docker-compose logs -f api
```

### Building Production Images

```bash
# Build API server image
docker build -t vibe-coding:latest .

# Build Agent server image
docker build -f agent-server/Dockerfile -t vibe-coding-agent:latest agent-server/
```

## Kubernetes Deployment

### 1. Create Namespace

```bash
kubectl create namespace vibe-coding
```

### 2. Create Secret

```bash
kubectl create secret generic vibe-coding-secrets \
  --from-literal=jwt-secret="$(openssl rand -base64 32)" \
  --from-literal=anthropic-api-key="your-api-key" \
  --from-literal=mysql-password="$(openssl rand -base64 16)" \
  --namespace=vibe-coding
```

### 3. Deploy MySQL

```yaml
# mysql-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mysql
  namespace: vibe-coding
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mysql
  template:
    metadata:
      labels:
        app: mysql
    spec:
      containers:
      - name: mysql
        image: mysql:8.0
        env:
        - name: MYSQL_ROOT_PASSWORD
          valueFrom:
            secretKeyRef:
              name: vibe-coding-secrets
              key: mysql-password
        - name: MYSQL_DATABASE
          value: vibe_coding
        ports:
        - containerPort: 3306
        volumeMounts:
        - name: mysql-storage
          mountPath: /var/lib/mysql
      volumes:
      - name: mysql-storage
        persistentVolumeClaim:
          claimName: mysql-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: mysql
  namespace: vibe-coding
spec:
  selector:
    app: mysql
  ports:
  - port: 3306
```

### 4. Deploy API Server

```yaml
# api-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
  namespace: vibe-coding
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api
  template:
    metadata:
      labels:
        app: api
    spec:
      containers:
      - name: api
        image: vibe-coding:latest
        ports:
        - containerPort: 8888
        env:
        - name: APP_ENV
          value: "production"
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: vibe-coding-secrets
              key: jwt-secret
        - name: ANTHROPIC_API_KEY
          valueFrom:
            secretKeyRef:
              name: vibe-coding-secrets
              key: anthropic-api-key
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /ping
            port: 8888
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /ping
            port: 8888
          initialDelaySeconds: 5
          periodSeconds: 10
---
apiVersion: v1
kind: Service
metadata:
  name: api
  namespace: vibe-coding
spec:
  selector:
    app: api
  ports:
  - port: 8888
  type: LoadBalancer
```

### 5. Apply Manifests

```bash
kubectl apply -f k8s/
```

## Manual Deployment

### 1. Build Binaries

```bash
# Build for Linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o build/api-server ./cmd/api

# Build for macOS (Apple Silicon)
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o build/api-server-darwin-arm64 ./cmd/api
```

### 2. Systemd Service

Create `/etc/systemd/system/vibe-coding.service`:

```ini
[Unit]
Description=Vibe Coding API Server
After=network.target mysql.service redis.service

[Service]
Type=simple
User=vibecoding
Group=vibecoding
WorkingDirectory=/opt/vibe-coding
ExecStart=/opt/vibe-coding/api-server -config /opt/vibe-coding/config.prod.yaml
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=vibe-coding

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable vibe-coding
sudo systemctl start vibe-coding
sudo systemctl status vibe-coding
```

### 3. Nginx Reverse Proxy

```nginx
upstream vibe_coding_api {
    least_conn;
    server 127.0.0.1:8888;
    server 127.0.0.1:8889;
    server 127.0.0.1:8890;
}

upstream vibe_coding_agent {
    server 127.0.0.1:3001;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /etc/ssl/certs/your-domain.crt;
    ssl_certificate_key /etc/ssl/private/your-domain.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;

    # API Server
    location / {
        proxy_pass http://vibe_coding_api;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Agent Server (SSE)
    location /agent/ {
        proxy_pass http://vibe_coding_agent;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_buffering off;
        proxy_cache off;
        proxy_set_header Connection '';
        proxy_http_version 1.1;
        chunked_transfer_encoding off;
    }
}

server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}
```

## Monitoring

### Health Checks

```bash
# API health
curl https://your-domain.com/ping

# Detailed health
curl https://your-domain.com/health
```

### Metrics (Prometheus)

The application exposes Prometheus metrics at `/metrics`:

- HTTP request duration
- Request count by status code
- Database query duration
- Cache hit/miss rates
- Active connections

### Logging

Logs are written to:

- **Standard output**: Captured by journald/Docker
- **File**: Configurable via `LOG_OUTPUT`

View logs:

```bash
# Systemd
journalctl -u vibe-coding -f

# Docker
docker-compose logs -f api

# Kubernetes
kubectl logs -f deployment/api -n vibe-coding
```

### Alerts

Recommended alerts:

| Alert | Condition | Severity |
|-------|-----------|----------|
| High Error Rate | error rate > 5% | Warning |
| Service Down | health check fails | Critical |
| High Latency | P99 > 1s | Warning |
| Database Connection Pool Full | active connections > 90% | Warning |

## Backup and Recovery

### Database Backup

```bash
# Daily backup
0 2 * * * /usr/bin/mysqldump -u root -p${MYSQL_PASSWORD} vibe_coding | gzip > /backups/vibe_coding_$(date +\%Y\%m\%d).sql.gz

# Keep last 30 days
0 3 * * * find /backups -name "vibe_coding_*.sql.gz" -mtime +30 -delete
```

### Recovery

```bash
# Restore from backup
gunzip < /backups/vibe_coding_20241222.sql.gz | mysql -u root -p vibe_coding
```

## Security Checklist

- [ ] Change all default passwords
- [ ] Use strong JWT secret (32+ characters)
- [ ] Enable HTTPS with valid certificates
- [ ] Configure firewall rules
- [ ] Enable rate limiting
- [ ] Set up log monitoring
- [ ] Regular security updates
- [ ] Database backup automation
- [ ] Disable debug mode in production
- [ ] Use environment variables for secrets

## Troubleshooting

### Service won't start

```bash
# Check logs
journalctl -u vibe-coding -n 50

# Check port availability
sudo lsof -i :8888

# Check configuration
./api-server -config config.prod.yaml -validate
```

### Database connection errors

```bash
# Test MySQL connection
mysql -h your-host -u vibe_coding -p vibe_coding

# Check MySQL logs
sudo tail -f /var/log/mysql/error.log
```

### High memory usage

- Reduce `max_open_conns` in MySQL config
- Reduce cache pool sizes
- Enable response compression
- Add more RAM or scale horizontally

For more help, see [README](../README.md) or [open an issue](https://github.com/test-tt/vibe-coding/issues).
