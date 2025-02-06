# 项目名称
PROJECT_NAME := app

# Go 命令
GO := go

# Swag 命令
SWAG := swag

# 构建目标目录
BUILD_DIR := ./bin

# 文档目录
DOCS_DIR := ./docs/openapi

# 默认目标
all: build

# 构建项目
build:
	@echo "Building $(PROJECT_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@$(GO) build -o $(BUILD_DIR)/$(PROJECT_NAME) ./cmd/$(PROJECT_NAME)
	@echo "Build complete. Binary saved to $(BUILD_DIR)/$(PROJECT_NAME)"

# 运行项目
run:
	@echo "Running $(PROJECT_NAME)..."
	@$(GO) run ./cmd/$(PROJECT_NAME)

# 清理构建文件和文档
clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DOCS_DIR)
	@echo "Cleanup complete."

# 运行测试
test:
	@echo "Running tests..."
	@$(GO) test -v ./...
	@echo "Tests complete."

# 格式化代码
fmt:
	@echo "Formatting code..."
	@$(GO) fmt ./...
	@echo "Code formatted."

# 生成 OpenAPI 文档
swag:
	@echo "Generating OpenAPI docs..."
	@$(SWAG) init -g internal/adapter/http/v1/router.go -o $(DOCS_DIR)
	@echo "OpenAPI docs generated in $(DOCS_DIR)"

# 安装依赖
deps:
	@echo "Installing dependencies..."
	@$(GO) mod download
	@echo "Dependencies installed."

# 安装 Swag
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

.PHONY: all build run clean test fmt swag deps install-swag help