.PHONY: help install-node install-pnpm setup backend frontend migrate-up migrate-down migrate-status clean

NODE_VERSION := 20.19.0
NODE_DIR := /tmp/node-v$(NODE_VERSION)-linux-x64
export PATH := $(NODE_DIR)/bin:$(PATH)

help:
	@echo "Available commands:"
	@echo "  make install-node    - Install Node.js $(NODE_VERSION)"
	@echo "  make install-pnpm    - Install pnpm package manager"
	@echo "  make setup           - Setup the project (install dependencies)"
	@echo "  make backend         - Start backend services (database + server)"
	@echo "  make frontend        - Start frontend development server"
	@echo "  make migrate-up      - Run all pending migrations"
	@echo "  make migrate-down    - Rollback the last migration"
	@echo "  make migrate-status  - Show migration status"
	@echo "  make clean           - Stop all services"

install-node:
	@echo "Downloading Node.js $(NODE_VERSION)..."
	curl -fsSL https://nodejs.org/dist/v$(NODE_VERSION)/node-v$(NODE_VERSION)-linux-x64.tar.xz -o /tmp/node.tar.xz
	@echo "Extracting Node.js..."
	cd /tmp && tar -xf node.tar.xz
	@echo "Node.js $(NODE_VERSION) installed to $(NODE_DIR)"
	@echo "Node.js version: $$($(NODE_DIR)/bin/node --version)"
	@echo "npm version: $$($(NODE_DIR)/bin/npm --version)"
	@echo ""
	@echo "To use this Node.js version, run:"
	@echo "  export PATH=$(NODE_DIR)/bin:\$$PATH"

install-pnpm:
	@echo "Installing pnpm..."
	$(NODE_DIR)/bin/npm install -g pnpm
	@echo "pnpm version: $$(pnpm --version)"

setup:
	@echo "Setting up the project..."
	cd backend && go mod download
	cd frontend/app && rm -rf node_modules package-lock.json pnpm-lock.yaml && $(NODE_DIR)/bin/pnpm install
	@echo "Setup complete!"

backend:
	@echo "Starting backend services..."
	docker compose -f backend/docker-compose.backend.yml up -d
	cd backend && DATABASE_URL=postgres://xledger:xledger_secret@127.0.0.1:5432/xledger?sslmode=disable REDIS_URL=redis://127.0.0.1:6379/0 SMTP_HOST=smtp.example.com AUTH_CODE_PEPPER=local-pepper go run ./cmd/api serve

frontend:
	@echo "Starting frontend development server..."
	cd frontend/app && $(NODE_DIR)/bin/pnpm run dev

migrate-up:
	@echo "Running migrations..."
	cd backend && DATABASE_URL=postgres://xledger:xledger_secret@127.0.0.1:5432/xledger?sslmode=disable go run ./cmd/api migrate up

migrate-down:
	@echo "Rolling back migration..."
	cd backend && DATABASE_URL=postgres://xledger:xledger_secret@127.0.0.1:5432/xledger?sslmode=disable go run ./cmd/api migrate down

migrate-status:
	@echo "Checking migration status..."
	cd backend && DATABASE_URL=postgres://xledger:xledger_secret@127.0.0.1:5432/xledger?sslmode=disable go run ./cmd/api migrate status

clean:
	@echo "Stopping all services..."
	docker compose -f backend/docker-compose.backend.yml down -v
	pkill -f "vite" 2>/dev/null || true
	pkill -f "go run" 2>/dev/null || true
	@echo "All services stopped!"
