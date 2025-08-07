#!/bin/bash

DOCKER_USER="r1na1"

services=("user" "lecture" "chat" "push" "quiz" "agent")

for service in "${services[@]}"; do
  echo "🏷️  打标签: $service → $DOCKER_USER/$service:latest"
  docker tag $service $DOCKER_USER/$service:latest

  echo "📤 推送到 Docker Hub: $DOCKER_USER/$service:latest"
  docker push $DOCKER_USER/$service:latest

  echo "✅ $service 完成"
  echo "-----------------------------"
done

echo "🎉 所有服务构建并推送完成！"
