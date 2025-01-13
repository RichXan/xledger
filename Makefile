SHELL := /bin/bash
.PHONY: all build clean run test proto docker-up docker-down docker-restart db-drop db-reset

# 默认目标
all: build

# 构建所有服务
build:
	@echo "Building all services..."
	go build -o bin/api-service cmd/api-service/main.go

# 清理构建产物
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f */*/tmp/

# 运行所有服务
run: build run-api
	@echo "Running all services..."


# 运行api服务
run-api: 
	@echo "Running API service..."
	go run cmd/api-service/main.go -c config/config.yml start

# 运行测试
test:
	@echo "Running tests..."
	go test -v ./...

# 生成 protobuf 代码
proto:
	@echo "Generating protobuf code..."
	protoc --proto_path=proto --go_out=. --micro_out=. proto/user/user.proto

# Docker 相关命令
docker-up:
	@echo "Starting all containers..."
	docker-compose up -d

docker-down:
	@echo "Stopping all containers..."
	docker-compose down

docker-restart: docker-down docker-up

# 查看服务日志
logs:
	@echo "Showing service logs..."
	docker-compose logs -f

# 清理 Docker 资源
docker-clean: docker-down
	@echo "Cleaning Docker resources..."
	docker-compose down -v
	docker system prune -f

# 重新构建并启动服务
rebuild: clean build docker-restart

# 检查服务状态
status:
	@echo "Checking service status..."
	docker-compose ps

# 连接到postgres
postgres-cli:
	docker exec -it xledger-postgres psql -U admin -d xledger

# 连接到mysql
mysql-cli:
	docker exec -it xledger-mysql mysql -uroot -proot123 xledger

redis-cli:
	docker exec -it xledger-redis redis-cli

# 初始化数据库
init-db:
	@echo "Initializing database..."
	chmod +x scripts/migration/postgres/db_init.sh
	./scripts/migration/postgres/db_init.sh

# 删除数据库
db-drop:
	@echo "Dropping database..."
	chmod +x scripts/migration/postgres/db_drop.sh
	./scripts/migration/postgres/db_drop.sh

# 重置数据库（删除并重新初始化）
db-reset: db-drop init-db

# 启动 Swagger UI
swagger-ui:
	@echo "Starting Swagger UI..."
	docker-compose up -d swagger-ui

# 停止 Swagger UI
swagger-ui-down:
	@echo "Stopping Swagger UI..."
	docker-compose stop swagger-ui

# 重启 Swagger UI
swagger-ui-restart: swagger-ui-down swagger-ui

# 帮助信息
help:
	@echo "Available commands:"
	@echo "  make build          - Build all services"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make run           - Run all services"
	@echo "  make run-api       - Run API service only"
	@echo "  make run-user      - Run User service only"
	@echo "  make run-post      - Run Post service only"
	@echo "  make test           - Run tests"
	@echo "  make proto          - Generate protobuf code"
	@echo "  make docker-up      - Start all containers"
	@echo "  make docker-down    - Stop all containers"
	@echo "  make docker-restart - Restart all containers"
	@echo "  make logs           - View service logs"
	@echo "  make docker-clean   - Clean Docker resources"
	@echo "  make rebuild        - Rebuild and restart services"
	@echo "  make status         - Check service status"
	@echo "  make mysql-cli      - Connect to MySQL CLI"
	@echo "  make redis-cli      - Connect to Redis CLI"
	@echo "  make init-db        - Initialize database"
	@echo "  make db-drop        - Drop database and remove volume"
	@echo "  make db-reset       - Reset database (drop and reinitialize)"
	@echo "  make swagger-ui     - Start Swagger UI"
	@echo "  make swagger-ui-down - Stop Swagger UI"
	@echo "  make swagger-ui-restart - Restart Swagger UI"
	@echo "  make docker-pull    - Pull Docker images from mirror" 