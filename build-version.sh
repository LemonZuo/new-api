#!/bin/bash

# 获取最新的标签名
TAG=$(git describe --tags --abbrev=0)

# 检查是否获取到标签名
if [ -z "$TAG" ]; then
  echo "未找到任何 Git 标签。"
  exit 1
fi

# 转义标签名中的特殊字符
ESCAPED_TAG=$(printf '%s\n' "$TAG" | sed 's/[\/&]/\\&/g')

# 定义一个函数，基于操作系统名称选择合适的 sed 命令
update_version() {
  OS_NAME=$(uname)

  if [ "$OS_NAME" = "Darwin" ]; then
    # macOS 使用 BSD sed
    sed -i '' "s/^VERSION=.*/VERSION=$ESCAPED_TAG/" .env
  elif [ "$OS_NAME" = "Linux" ]; then
    # Linux 使用 GNU sed
    sed -i "s/^VERSION=.*/VERSION=$ESCAPED_TAG/" .env
  else
    # 其他操作系统，提示不支持
    echo "不支持的操作系统：$OS_NAME"
    exit 1
  fi
}

# 调用函数更新 VERSION 值
update_version

# 验证替换是否成功
if grep -q "^VERSION=$TAG" .env; then
  echo ".env 文件中的 VERSION 已更新为 $TAG"
else
  echo "更新 .env 文件中的 VERSION 时出现错误。"
  exit 1
fi
