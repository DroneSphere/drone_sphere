# 项目名称
PROJECT_NAME := app

# Go 命令
GO := go

# Swag 命令
SWAG := swag

# 构建目标目录
BUILD_DIR := ./bin

# 文档目录
DOCS_DIR := ./docs/http

# 当前目录
PWD := $(shell pwd)

# 默认目标
all: build

# 构建项目
.PHONY: build
build:
	@echo "Building $(PROJECT_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@$(GO) build -o $(BUILD_DIR)/$(PROJECT_NAME) ./cmd/$(PROJECT_NAME)
	@echo "Build complete. Binary saved to $(BUILD_DIR)/$(PROJECT_NAME)"

# 运行项目
.PHONY: run
run:
	@echo "Running $(PROJECT_NAME)..."
	@$(GO) run ./cmd/$(PROJECT_NAME)

# 清理构建文件和文档
.PHONY: clean
clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DOCS_DIR)
	@echo "Cleanup complete."

# 运行测试
.PHONY: test
test:
	@echo "Running tests..."
	@$(GO) test -v ./...
	@echo "Tests complete."

# 格式化代码
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@$(GO) fmt ./...
	@$(SWAG) fmt
	@echo "Code formatted."

# 生成 OpenAPI 文档
.PHONY: swag
swag:
	@echo "Generating OpenAPI docs..."
	@$(SWAG) init -g router.go -o $(DOCS_DIR)/v1 --dir internal/adapter/http/v1,api/http/v1
	@$(SWAG) init -g router.go -o $(DOCS_DIR)/dji --dir internal/adapter/http/dji,api/http/dji
	@echo "OpenAPI docs generated in $(DOCS_DIR)"

# 安装依赖
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	@$(GO) mod download
	@$(GO) mod tidy
	@$(GO) mod vendor
	@echo "Dependencies installed."

# 安装 Swag
.PHONY: install-swag
install-swag:
	@echo "Installing swag..."
	@$(GO) install github.com/swaggo/swag/cmd/swag@latest
	@echo "Swag installed."

# 显示帮助信息
help:
	@echo "Available targets:"
	@echo "  all         - Build the project (default target)"
	@echo "  build       - Build the project"
	@echo "  run         - Run the project"
	@echo "  clean       - Clean up build files and docs"
	@echo "  test        - Run tests"
	@echo "  fmt         - Format code"
	@echo "  swag        - Generate OpenAPI docs"
	@echo "  deps        - Install dependencies"
	@echo "  install-swag - Install swag tool"
	@echo "  help        - Show this help message"
