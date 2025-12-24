#!/bin/bash

# 开发环境启动脚本 - 同时启动 Go 后端和 Agent Server
# 使用方法: ./scripts/dev.sh [start|stop|restart|status]

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PID_DIR="$PROJECT_ROOT/.pids"
LOG_DIR="$PROJECT_ROOT/logs"

# 确保目录存在
mkdir -p "$PID_DIR" "$LOG_DIR"

# PID 文件
GO_PID_FILE="$PID_DIR/go-server.pid"
AGENT_PID_FILE="$PID_DIR/agent-server.pid"

# 日志文件
GO_LOG_FILE="$LOG_DIR/go-server.log"
AGENT_LOG_FILE="$LOG_DIR/agent-server.log"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查进程是否运行
is_running() {
    local pid_file=$1
    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        if ps -p "$pid" > /dev/null 2>&1; then
            return 0
        fi
    fi
    return 1
}

# 启动 Go 后端
start_go_server() {
    if is_running "$GO_PID_FILE"; then
        log_warn "Go server is already running (PID: $(cat $GO_PID_FILE))"
        return 0
    fi

    log_info "Starting Go server on port 8888..."
    cd "$PROJECT_ROOT"
    nohup go run ./cmd/api/main.go -config=./config/config.dev.yaml > "$GO_LOG_FILE" 2>&1 &
    echo $! > "$GO_PID_FILE"
    sleep 2

    if is_running "$GO_PID_FILE"; then
        log_info "Go server started (PID: $(cat $GO_PID_FILE))"
    else
        log_error "Failed to start Go server"
        return 1
    fi
}

# 启动 Agent Server
start_agent_server() {
    if is_running "$AGENT_PID_FILE"; then
        log_warn "Agent server is already running (PID: $(cat $AGENT_PID_FILE))"
        return 0
    fi

    log_info "Starting Agent server on port 3001..."
    cd "$PROJECT_ROOT/agent-server"

    # 设置 JWT 环境变量
    export JWT_SECRET="${JWT_SECRET:-dev-secret-key-at-least-32-chars!}"
    export JWT_ISSUER="${JWT_ISSUER:-test-tt}"

    nohup npm start > "$AGENT_LOG_FILE" 2>&1 &
    echo $! > "$AGENT_PID_FILE"
    sleep 2

    if is_running "$AGENT_PID_FILE"; then
        log_info "Agent server started (PID: $(cat $AGENT_PID_FILE))"
    else
        log_error "Failed to start Agent server"
        return 1
    fi
}

# 停止服务 (通过端口)
stop_by_port() {
    local name=$1
    local port=$2
    local pid_file=$3

    local pids=$(lsof -ti:$port 2>/dev/null)
    if [ -n "$pids" ]; then
        log_info "Stopping $name on port $port..."
        echo "$pids" | xargs kill 2>/dev/null || true

        # 等待进程结束
        sleep 1

        # 强制终止残留进程
        pids=$(lsof -ti:$port 2>/dev/null)
        if [ -n "$pids" ]; then
            log_warn "Force killing $name..."
            echo "$pids" | xargs kill -9 2>/dev/null || true
        fi

        rm -f "$pid_file"
        log_info "$name stopped"
    else
        log_warn "$name is not running on port $port"
        rm -f "$pid_file"
    fi
}

# 启动所有服务
start_all() {
    log_info "Starting all services..."
    start_go_server
    start_agent_server
    echo ""
    log_info "All services started!"
    echo ""
    echo "  Go Server:    http://localhost:8888"
    echo "  Agent Server: http://localhost:3001"
    echo ""
    echo "  Logs:"
    echo "    Go Server:    $GO_LOG_FILE"
    echo "    Agent Server: $AGENT_LOG_FILE"
    echo ""
    echo "  Use 'make stop' or './scripts/dev.sh stop' to stop all services"
}

# 停止所有服务
stop_all() {
    log_info "Stopping all services..."
    stop_by_port "Agent server" 3001 "$AGENT_PID_FILE"
    stop_by_port "Go server" 8888 "$GO_PID_FILE"
    log_info "All services stopped"
    return 0
}

# 检查端口是否被占用
is_port_in_use() {
    lsof -ti:$1 > /dev/null 2>&1
}

# 显示状态
show_status() {
    echo ""
    echo "Service Status:"
    echo "==============="

    if is_port_in_use 8888; then
        local go_pid=$(lsof -ti:8888 | head -1)
        echo -e "  Go Server (8888):    ${GREEN}Running${NC} (PID: $go_pid)"
    else
        echo -e "  Go Server (8888):    ${RED}Stopped${NC}"
    fi

    if is_port_in_use 3001; then
        local agent_pid=$(lsof -ti:3001 | head -1)
        echo -e "  Agent Server (3001): ${GREEN}Running${NC} (PID: $agent_pid)"
    else
        echo -e "  Agent Server (3001): ${RED}Stopped${NC}"
    fi
    echo ""
}

# 查看日志
show_logs() {
    local service=$1
    case $service in
        go)
            tail -f "$GO_LOG_FILE"
            ;;
        agent)
            tail -f "$AGENT_LOG_FILE"
            ;;
        *)
            tail -f "$GO_LOG_FILE" "$AGENT_LOG_FILE"
            ;;
    esac
}

# 主函数
case "${1:-start}" in
    start)
        start_all
        ;;
    stop)
        stop_all
        ;;
    restart)
        stop_all
        sleep 1
        start_all
        ;;
    status)
        show_status
        ;;
    logs)
        show_logs "${2:-all}"
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status|logs [go|agent]}"
        exit 1
        ;;
esac
