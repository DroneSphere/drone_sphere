#!/bin/bash

# 镜像名称
IMAGE_NAME="ghcr.io/dronesphere/drone-sphere-neo:latest"

# 容器名称
CONTAINER_NAME="dronesphere"

# 拉取最新镜像
echo "Pulling latest image..."
docker pull $IMAGE_NAME

# 停止并删除旧容器
echo "Stopping and removing old container..."
docker stop $CONTAINER_NAME || true
docker rm $CONTAINER_NAME || true

# 启动新容器
echo "Starting new container..."
docker run -d \
  --name $CONTAINER_NAME \
  -p 10086:10086 \
  $IMAGE_NAME

echo "Container updated successfully!"