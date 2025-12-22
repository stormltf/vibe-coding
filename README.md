# Go Web API 标杆项目

基于 Hertz 框架的生产级 Go Web API 项目模板。

## 技术栈

| 组件 | 技术选型 | 说明 |
|------|---------|------|
| Web 框架 | [Hertz](https://github.com/cloudwego/hertz) | 字节跳动高性能 HTTP 框架 |
| 日志 | [Zap](https://github.com/uber-go/zap) | Uber 高性能日志库 |
| ORM | [GORM](https://gorm.io/) | Go 语言 ORM 框架 |
| 缓存 | [go-redis](https://github.com/redis/go-redis) | Redis 官方客户端 |
| 配置 | [Viper](https://github.com/spf13/viper) | 配置管理（YAML + 环境变量）|
| JSON | [Sonic](https://github.com/bytedance/sonic) | 字节跳动高性能 JSON 库 |
| 认证 | [JWT](https://github.com/golang-jwt/jwt) | JSON Web Token |
| 监控 | [Prometheus](https://prometheus.io/) | 指标监控 |
| 校验 | [Validator](https://github.com/go-playground/validator) | 参数校验 |

## 项目特性

### 核心功能

- **参数校验**: 基于 validator 的请求参数校验
- **分页查询**: 统一的分页查询支持
- **限流保护**: 基于令牌桶的 IP 限流
- **JWT 认证**: Bearer Token 认证中间件
- **链路追踪**: 基于 logid 的请求链路追踪
- **指标监控**: Prometheus 指标采集

### 中间件

| 中间件 | 功能 |
|--------|------|
| Recovery | 捕获 panic，防止服务崩溃 |
| RequestID | 为每个请求生成唯一 ID |
| AccessLog | 记录请求日志（自动携带 logid）|
| CORS | 跨域支持 |
| RateLimit | 请求限流 |
| JWTAuth | JWT 认证 |
| Metrics | Prometheus 指标采集 |

### 日志系统

- 控制台彩色输出 + JSON 文件日志
- 自动轮转（按大小/时间）
- logid 链路追踪
- 简化 API：`logger.InfoCtxf(ctx, "msg", "key", value)`

## 项目结构

```
.
├── cmd/                        # 应用入口
│   └── api/
│       └── main.go
├── config/                     # 配置
│   ├── config.go               # 配置定义
│   ├── config.yaml             # 默认配置
│   ├── config.dev.yaml         # 开发环境
│   └── config.prod.yaml        # 生产环境
├── internal/                   # 内部代码
│   ├── dao/                    # 数据访问层
│   ├── handler/                # 请求处理器
│   ├── middleware/             # 中间件
│   │   ├── access_log.go       # 访问日志
│   │   ├── cors.go             # 跨域
│   │   ├── jwt.go              # JWT 认证
│   │   ├── metrics.go          # Prometheus 指标
│   │   ├── ratelimit.go        # 限流
│   │   ├── recovery.go         # Panic 恢复
│   │   └── request_id.go       # 请求 ID
│   ├── model/                  # 数据模型
│   ├── router/                 # 路由
│   └── service/                # 业务逻辑
├── pkg/                        # 公共包
│   ├── cache/                  # Redis 封装
│   ├── database/               # MySQL 封装
│   ├── errcode/                # 错误码定义
│   ├── jwt/                    # JWT 工具
│   ├── logger/                 # 日志封装
│   ├── pagination/             # 分页工具
│   ├── response/               # 统一响应
│   └── validate/               # 参数校验
├── scripts/                    # 脚本
│   └── init.sql                # 数据库初始化
├── .air.toml                   # Air 热重载配置
├── Dockerfile
├── docker-compose.yaml
├── Makefile
└── README.md
```

## 快速开始

### 本地开发

```bash
# 安装依赖
make tidy

# 热重载开发（推荐）
make dev

# 或使用配置文件运行
make run-dev     # 开发环境
make run-prod    # 生产环境
```

### Docker 部署

```bash
# 构建并启动
make docker-up

# 查看日志
make docker-logs

# 停止
make docker-down
```

### 常用命令

```bash
make build        # 编译
make test         # 运行测试
make test-cover   # 测试覆盖率
make lint         # 代码检查
make fmt          # 格式化代码
make dev          # 热重载开发
make help         # 查看帮助
```

## API 接口

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | `/ping` | 健康检查 | 否 |
| GET | `/metrics` | Prometheus 指标 | 否 |
| GET | `/api/v1/users` | 获取用户列表（分页）| 否 |
| GET | `/api/v1/users/:id` | 获取用户详情 | 否 |
| POST | `/api/v1/users` | 创建用户 | JWT |
| PUT | `/api/v1/users/:id` | 更新用户 | JWT |
| DELETE | `/api/v1/users/:id` | 删除用户 | JWT |

### 分页参数

```
GET /api/v1/users?page=1&page_size=10
```

响应格式：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [...],
    "total": 100,
    "page": 1,
    "page_size": 10,
    "pages": 10
  }
}
```

### JWT 认证

需要在请求头添加：
```
Authorization: Bearer <token>
```

## 配置说明

### 多环境配置

```bash
config/
├── config.yaml       # 默认配置
├── config.dev.yaml   # 开发环境
└── config.prod.yaml  # 生产环境
```

### 环境变量

支持通过环境变量覆盖配置，前缀为 `APP_`：

```bash
APP_SERVER_PORT=9000
APP_MYSQL_HOST=mysql.example.com
APP_REDIS_HOST=redis.example.com
APP_JWT_SECRET=your-secret-key
```

### 配置示例

```yaml
env: dev

server:
  host: 0.0.0.0
  port: 8888

mysql:
  host: 127.0.0.1
  port: 3306
  username: root
  password: ""
  database: test

redis:
  host: 127.0.0.1
  port: 6379

log:
  level: info
  filename: logs/app.log

jwt:
  secret: your-secret-key
  expire_time: 24h

ratelimit:
  rate: 100
  burst: 200
```

## 开发规范

### 分层架构

```
Handler -> Service -> DAO -> Model
    ↓         ↓        ↓
 请求处理   业务逻辑   数据访问
```

### 添加新接口

1. `internal/model/` - 定义数据模型
2. `internal/dao/` - 实现数据访问
3. `internal/service/` - 实现业务逻辑
4. `internal/handler/` - 实现请求处理
5. `internal/router/` - 注册路由

### 日志使用

```go
import "github.com/test-tt/pkg/logger"

// 带 context（推荐，自动携带 logid）
logger.InfoCtxf(ctx, "user created", "id", user.ID, "name", user.Name)

// 不带 context
logger.Infof("server started", "port", 8888)

// 使用字段函数
logger.Info("user created", logger.Uint64("id", user.ID))
```

### 参数校验

```go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=2,max=50"`
    Age   int    `json:"age" validate:"gte=0,lte=150"`
    Email string `json:"email" validate:"omitempty,email"`
}

// 在 handler 中校验
if err := validate.Struct(&req); err != nil {
    response.Fail(c, errcode.ErrInvalidParams.WithMessage(validate.FirstError(err)))
    return
}
```

### 错误码

```go
response.Fail(c, errcode.ErrUserNotFound)
response.Fail(c, errcode.ErrInvalidParams.WithMessage("name is required"))
```

## 监控

### Prometheus 指标

访问 `/metrics` 获取指标数据：

- `http_requests_total` - HTTP 请求总数
- `http_request_duration_seconds` - HTTP 请求延迟
- `http_requests_in_flight` - 当前处理中的请求数

### Grafana 配置

可导入 Prometheus 数据源，创建 Dashboard 监控：
- QPS
- 请求延迟 (P50/P90/P99)
- 错误率
- 在途请求数

## License

MIT
