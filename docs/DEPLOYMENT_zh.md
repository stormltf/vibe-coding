# 部署指南

本指南涵盖将 Vibe Coding 部署到生产环境的内容。

## 目录

- [前置要求](#前置要求)
- [环境配置](#环境配置)
- [Docker 部署](#docker-部署)
- [Kubernetes 部署](#kubernetes-部署)
- [手动部署](#手动部署)
- [监控](#监控)
- [备份与恢复](#备份与恢复)

## 前置要求

### 系统要求

| 组件 | 最低配置 | 推荐配置 |
|-----------|---------|-------------|
| CPU | 2 核 | 4+ 核 |
| 内存 | 4 GB | 8+ GB |
| 磁盘 | 20 GB | 50+ GB SSD |
| MySQL | 8.0+ | 8.0+ |
| Redis | 7.0+ | 7.0+ |

### 必需的环境变量

```bash
# 应用
APP_ENV=production
APP_SERVER_PORT=8888
APP_SERVER_HOST=0.0.0.0

# 数据库
APP_MYSQL_HOST=your-mysql-host
APP_MYSQL_PORT=3306
APP_MYSQL_USER=vibe_coding
APP_MYSQL_PASSWORD=your-secure-password
APP_MYSQL_DATABASE=vibe_coding

# Redis
APP_REDIS_HOST=your-redis-host
APP_REDIS_PORT=6379
APP_REDIS_PASSWORD=your-redis-password

# 安全
JWT_SECRET=your-jwt-secret-min-32-chars
ANTHROPIC_API_KEY=your-claude-api-key

# 日志
LOG_LEVEL=info
LOG_OUTPUT=/var/log/vibe-coding/app.log
```

## 环境配置

### 1. 生产环境配置文件

创建 `config/config.prod.yaml`：

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
  expire: 168h  # 7 天

rate_limit:
  enabled: true
  qps: 1000
  burst: 2000
```

## Docker 部署

### 使用 Docker Compose（推荐）

1. **设置环境变量**：

```bash
export ANTHROPIC_API_KEY="your-api-key"
export JWT_SECRET="$(openssl rand -base64 32)"
export MYSQL_PASSWORD="$(openssl rand -base64 16)"
```

2. **为生产环境更新 docker-compose.yaml**：

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

3. **启动服务**：

```bash
docker-compose up -d
```

4. **验证部署**：

```bash
docker-compose ps
docker-compose logs -f api
```

### 构建生产镜像

```bash
# 构建 API 服务器镜像
docker build -t vibe-coding:latest .

# 构建 Agent 服务器镜像
docker build -f agent-server/Dockerfile -t vibe-coding-agent:latest agent-server/
```

## Kubernetes 部署

### 1. 创建命名空间

```bash
kubectl create namespace vibe-coding
```

### 2. 创建 Secret

```bash
kubectl create secret generic vibe-coding-secrets \
  --from-literal=jwt-secret="$(openssl rand -base64 32)" \
  --from-literal=anthropic-api-key="your-api-key" \
  --from-literal=mysql-password="$(openssl rand -base64 16)" \
  --namespace=vibe-coding
```

### 3. 部署 MySQL

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

### 4. 部署 API 服务

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

### 5. 应用清单

```bash
kubectl apply -f k8s/
```

## 手动部署

### 1. 构建二进制文件

```bash
# 为 Linux 构建
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o build/api-server ./cmd/api

# 为 macOS (Apple Silicon) 构建
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o build/api-server-darwin-arm64 ./cmd/api
```

### 2. Systemd 服务

创建 `/etc/systemd/system/vibe-coding.service`：

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

启用并启动：

```bash
sudo systemctl daemon-reload
sudo systemctl enable vibe-coding
sudo systemctl start vibe-coding
sudo systemctl status vibe-coding
```

### 3. Nginx 反向代理

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

    # API 服务器
    location / {
        proxy_pass http://vibe_coding_api;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Agent 服务器 (SSE)
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

## 监控

### 健康检查

```bash
# API 健康检查
curl https://your-domain.com/ping

# 详细健康信息
curl https://your-domain.com/health
```

### 指标（Prometheus）

应用程序在 `/metrics` 端点暴露 Prometheus 指标：

- HTTP 请求持续时间
- 按状态码的请求计数
- 数据库查询持续时间
- 缓存命中/未命中率
- 活跃连接数

### 日志

日志写入到：

- **标准输出**：被 journald/Docker 捕获
- **文件**：可通过 `LOG_OUTPUT` 配置

查看日志：

```bash
# Systemd
journalctl -u vibe-coding -f

# Docker
docker-compose logs -f api

# Kubernetes
kubectl logs -f deployment/api -n vibe-coding
```

### 告警

推荐的告警：

| 告警 | 条件 | 严重级别 |
|-------|-----------|----------|
| 高错误率 | 错误率 > 5% | 警告 |
| 服务宕机 | 健康检查失败 | 严重 |
| 高延迟 | P99 > 1s | 警告 |
| 数据库连接池满 | 活跃连接 > 90% | 警告 |

## 备份与恢复

### 数据库备份

```bash
# 每日备份
0 2 * * * /usr/bin/mysqldump -u root -p${MYSQL_PASSWORD} vibe_coding | gzip > /backups/vibe_coding_$(date +\%Y\%m\%d).sql.gz

# 保留最近 30 天
0 3 * * * find /backups -name "vibe_coding_*.sql.gz" -mtime +30 -delete
```

### 恢复

```bash
# 从备份恢复
gunzip < /backups/vibe_coding_20241222.sql.gz | mysql -u root -p vibe_coding
```

## 安全清单

- [ ] 更改所有默认密码
- [ ] 使用强 JWT 密钥（32+ 字符）
- [ ] 启用 HTTPS 并使用有效证书
- [ ] 配置防火墙规则
- [ ] 启用速率限制
- [ ] 设置日志监控
- [ ] 定期安全更新
- [ ] 数据库备份自动化
- [ ] 生产环境禁用调试模式
- [ ] 使用环境变量管理密钥

## 故障排除

### 服务无法启动

```bash
# 检查日志
journalctl -u vibe-coding -n 50

# 检查端口占用
sudo lsof -i :8888

# 验证配置
./api-server -config config.prod.yaml -validate
```

### 数据库连接错误

```bash
# 测试 MySQL 连接
mysql -h your-host -u vibe_coding -p vibe_coding

# 检查 MySQL 日志
sudo tail -f /var/log/mysql/error.log
```

### 高内存使用

- 减少 MySQL 配置中的 `max_open_conns`
- 减少缓存池大小
- 启用响应压缩
- 增加更多内存或水平扩展

需要更多帮助，请参阅 [README](../README_zh.md) 或 [提交 issue](https://github.com/test-tt/vibe-coding/issues)。
