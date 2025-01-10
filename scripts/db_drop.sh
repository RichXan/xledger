#!/bin/bash

echo "Warning: This will delete all data in the database!"
echo "Waiting for 5 seconds before proceeding..."
sleep 5

# 停止 MySQL 容器
echo "Stopping MySQL container..."
docker stop xledger-mysql || true

# 删除 Docker 数据卷
echo "Removing Docker volume..."
docker volume rm xledger_mysql_data || true

# 重新启动 MySQL 容器
echo "Starting MySQL container..."
docker start xledger-mysql

# 等待 MySQL 启动
echo "Waiting for MySQL to be ready..."
for i in {1..30}; do
    if docker exec xledger-mysql mysqladmin ping -h"localhost" -p"root123" --silent > /dev/null 2>&1; then
        break
    fi
    echo -n "."
    sleep 1
done
echo ""

# 创建数据库
echo "Creating database..."
docker exec xledger-mysql mysql -uroot -proot123 -e "CREATE DATABASE IF NOT EXISTS x_ledger CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

echo "Database has been reset!" 