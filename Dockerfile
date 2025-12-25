# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# 安装必要工具
RUN apk add --no-cache git ca-certificates tzdata

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/api/main.go

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# 安装必要工具
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 从 builder 复制文件
COPY --from=builder /app/server .
COPY --from=builder /app/config/config.yaml ./config/

# 创建日志目录
RUN mkdir -p logs

# 暴露端口
EXPOSE 8888

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8888/ping || exit 1

# 运行
CMD ["./server", "-config=./config/config.yaml"]
