# 第一阶段：构建阶段
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 Go 模块文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制项目代码
COPY . .

# 安装 swag 工具
RUN go install github.com/swaggo/swag/cmd/swag@latest

# 生成 OpenAPI 文档
RUN swag init -g internal/adapter/http/v1/router.go -o ./docs/openapi

# 构建项目
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/app ./cmd/app

# 第二阶段：运行阶段
FROM alpine:latest AS runner

# 设置工作目录
WORKDIR /app

# 从构建阶段复制可执行文件
COPY --from=builder /app/bin/app ./bin/app

# 从构建阶段复制 OpenAPI 文档
COPY --from=builder /app/docs/openapi ./docs/openapi

# 暴露端口（根据项目需求调整）
EXPOSE 10086

# 设置启动命令
CMD ["./bin/app"]