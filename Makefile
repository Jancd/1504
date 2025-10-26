.PHONY: build run test clean help setup

# 变量定义
BINARY_NAME=video-generator
MAIN_PATH=./cmd/server
BUILD_DIR=./bin
CONFIG_PATH=./configs/config.yaml

# 默认目标
.DEFAULT_GOAL := help

## help: 显示帮助信息
help:
	@echo "可用命令:"
	@echo "  make setup        - 初始化项目环境"
	@echo "  make build        - 编译项目"
	@echo "  make run          - 运行服务"
	@echo "  make dev          - 开发模式运行(自动重载)"
	@echo "  make dev-full     - 启动前后端开发环境"
	@echo "  make frontend     - 启动前端开发服务"
	@echo "  make test         - 运行测试"
	@echo "  make clean        - 清理编译文件"
	@echo "  make check        - 检查依赖"
	@echo "  make lint         - 代码检查"

## setup: 初始化项目环境
setup:
	@echo "==> 初始化项目环境..."
	@chmod +x scripts/*.sh
	@./scripts/setup.sh

## build: 编译项目
build:
	@echo "==> 编译项目..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "==> 编译完成: $(BUILD_DIR)/$(BINARY_NAME)"

## run: 运行服务
run: build
	@echo "==> 启动服务..."
	@$(BUILD_DIR)/$(BINARY_NAME)

## dev: 开发模式运行
dev:
	@echo "==> 开发模式启动..."
	@go run $(MAIN_PATH)/main.go

## test: 运行测试
test:
	@echo "==> 运行测试..."
	@go test -v ./...

## clean: 清理编译文件
clean:
	@echo "==> 清理编译文件..."
	@rm -rf $(BUILD_DIR)
	@rm -rf data/projects/*
	@echo "==> 清理完成"

## check: 检查依赖
check:
	@echo "==> 检查依赖..."
	@echo "检查Go版本..."
	@go version
	@echo "检查FFmpeg..."
	@ffmpeg -version | head -n 1
	@echo "检查Git..."
	@git --version
	@echo "==> 依赖检查完成"

## lint: 代码检查
lint:
	@echo "==> 代码检查..."
	@go fmt ./...
	@go vet ./...

## tidy: 整理依赖
tidy:
	@echo "==> 整理依赖..."
	@go mod tidy
	@echo "==> 依赖整理完成"

## install: 安装到系统
install: build
	@echo "==> 安装到系统..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "==> 安装完成"

## dev-full: 启动完整开发环境(前后端)
dev-full:
	@echo "==> 启动完整开发环境..."
ifeq ($(OS),Windows_NT)
	@scripts\start-dev.bat
else
	@chmod +x scripts/start-dev.sh
	@./scripts/start-dev.sh
endif

## frontend: 启动前端开发服务
frontend:
	@echo "==> 启动前端开发服务..."
	@cd frontend && npm install && npm run dev

## frontend-build: 构建前端生产版本
frontend-build:
	@echo "==> 构建前端生产版本..."
	@cd frontend && npm install && npm run build

## frontend-clean: 清理前端依赖和构建文件
frontend-clean:
	@echo "==> 清理前端文件..."
	@rm -rf frontend/node_modules
	@rm -rf frontend/dist