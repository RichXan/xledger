#!/bin/sh
set -eu

SOURCE_CONFIG="/app/config/source-config.yaml"
RUNTIME_CONFIG="/tmp/xledger-config.yaml"

# deploy 直接依赖 backend/config/config.yaml，
# 但容器内需要把本地 127.0.0.1 依赖切换为 compose 服务名。
sed \
  -e 's@127\.0\.0\.1:5432@postgres:5432@g' \
  -e 's@127\.0\.0\.1:6379@redis:6379@g' \
  "$SOURCE_CONFIG" > "$RUNTIME_CONFIG"

export CONFIG_FILE="$RUNTIME_CONFIG"

exec /app/xledger
