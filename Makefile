.PHONY: build run test clean docker-build docker-up docker-down lint fmt help dev agent agent-dev agent-install

# 变量
APP_NAME := test-tt
BUILD_DIR := ./build
MAIN_FILE := ./cmd/api/main.go
DOCKER_IMAGE := $(APP_NAME):latest
ENV ?= dev
AGENT_DIR := ./agent-server

# JWT 配置 - Agent Server 与 Go 后端共享
export JWT_SECRET ?= dev-secret-key-at-least-32-chars!
export JWT_ISSUER ?= test-tt

# Go 相关
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod
GOFMT := gofmt

# 默认目标
.DEFAULT_GOAL := help

## build: 编译项目
build:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

## run: 运行项目
run:
	$(GOCMD) run $(MAIN_FILE)

## run-dev: 使用开发环境配置运行
run-dev:
	$(GOCMD) run $(MAIN_FILE) -config=./config/config.dev.yaml

## run-prod: 使用生产环境配置运行
run-prod:
	$(GOCMD) run $(MAIN_FILE) -config=./config/config.prod.yaml

## run-config: 使用指定配置文件运行
run-config:
	$(GOCMD) run $(MAIN_FILE) -config=./config/config.yaml

## dev: 热重载开发模式 (需要安装 air)
dev:
	@which air > /dev/null || (echo "Installing air..." && go install github.com/air-verse/air@latest)
	air

## test: 运行测试
test:
	$(GOTEST) -v -race ./...

## test-cover: 运行测试并生成覆盖率报告
test-cover:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## bench: 运行基准测试
bench:
	$(GOTEST) -bench=. -benchmem ./...

## lint: 代码检查
lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

## fmt: 格式化代码
fmt:
	$(GOFMT) -w .

## tidy: 整理依赖
tidy:
	$(GOMOD) tidy

## clean: 清理构建产物
clean:
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Cleaned"

## docker-build: 构建 Docker 镜像
docker-build:
	docker build -t $(DOCKER_IMAGE) .

## docker-up: 启动 Docker Compose
docker-up:
	docker-compose up -d

## docker-down: 停止 Docker Compose
docker-down:
	docker-compose down

## docker-logs: 查看 Docker 日志
docker-logs:
	docker-compose logs -f

## help: 显示帮助信息
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

# ========== Agent Server ==========

## agent-install: 安装 Agent Server 依赖
agent-install:
	@echo "Installing Agent Server dependencies..."
	cd $(AGENT_DIR) && npm install
	@echo "Agent Server dependencies installed"

## agent: 启动 Agent Server (生产模式)
agent:
	@echo "Starting Agent Server..."
	@echo "JWT_SECRET is set: $(if $(JWT_SECRET),yes,no)"
	cd $(AGENT_DIR) && npm start

## agent-dev: 启动 Agent Server (开发模式，自动重载)
agent-dev:
	@echo "Starting Agent Server in dev mode..."
	@echo "JWT_SECRET is set: $(if $(JWT_SECRET),yes,no)"
	cd $(AGENT_DIR) && npm run dev

## all: 同时启动 Go 后端和 Agent Server (需要两个终端)
all:
	@echo "Please run in separate terminals:"
	@echo "  Terminal 1: make dev"
	@echo "  Terminal 2: make agent"
	@echo ""
	@echo "Or use: make start-all (background mode)"

## start-all: 后台启动所有服务
start-all:
	@echo "Starting all services in background..."
	@make dev &
	@sleep 2
	@make agent &
	@echo "Services started. Use 'make stop-all' to stop."

## stop-all: 停止所有服务
stop-all:
	@echo "Stopping all services..."
	@pkill -f "air" || true
	@pkill -f "node.*agent-server" || true
	@echo "Services stopped"
