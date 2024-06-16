#!/bin/bash
# 从 Git 拉取最新的代码
docker-compose down -v

# 从 Git 拉取最新的代码
#git pull

# 更新 镜像版本
docker-compose pull

# 构建启动容器
docker-compose up -d --build

# 查看日志
docker-compose logs -f