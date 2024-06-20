#!/bin/bash
# 更新镜像
docker-compose pull
# 停止并删除容器
docker-compose down -v
# 启动容器
#docker-compose up -d --build
docker-compose up -d
# 查看日志
docker-compose logs -f