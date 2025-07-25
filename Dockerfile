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

# 构建项目
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/app ./cmd/app

# 第二阶段：运行阶段
FROM alpine:latest AS runner

# 设置工作目录
WORKDIR /app

# 从构建阶段复制可执行文件
COPY --from=builder /app/bin/app ./bin/app

# 设置 TZ 环境变量为东八区
ENV TZ=Asia/Shanghai

# 安装 tzdata 包，它包含了 /usr/share/zoneinfo 下的所有时区数据
# 并创建 /etc/localtime 软链接，写入 /etc/timezone 文件
RUN apk update && \
    apk add --no-cache tzdata && \
    ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && \
    echo $TZ > /etc/timezone && \
    rm -rf /var/cache/apk/*

# 暴露端口（根据项目需求调整）
EXPOSE 10086

# 设置启动命令
CMD ["./bin/app"]