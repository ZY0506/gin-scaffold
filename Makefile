.PHONY: help run build test lint tidy docker-up docker-down

APP_NAME = gin-scaffold
BUILD_DIR = ./tmp

help: ## 显示帮助信息
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

run: ## 运行服务（热重载，需安装 air）
	air

build: ## 编译二进制
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

test: ## 运行测试
	go test -v -cover ./...

lint: ## 代码检查（需安装 golangci-lint）
	golangci-lint run ./...

tidy: ## 整理依赖
	go mod tidy

docker-up: ## 启动 Docker 容器
	docker-compose -f deployments/docker-compose.yml up -d

docker-down: ## 停止 Docker 容器
	docker-compose -f deployments/docker-compose.yml down

docker-build: ## Docker 构建
	docker-compose -f deployments/docker-compose.yml build

migrate: ## 数据库迁移
	go run ./cmd/server -migrate

fmt: ## 格式化代码
	go fmt ./...
