.PHONY: build run test clean docker-build docker-up docker-down lint fmt help dev

# 变量
APP_NAME := test-tt
BUILD_DIR := ./build
MAIN_FILE := ./cmd/api/main.go
DOCKER_IMAGE := $(APP_NAME):latest
ENV ?= dev

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
