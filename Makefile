.PHONY: all docs dev build test clean air-install hooks-install db-up db-down db-logs db-shell db-reset seed db-seed migrate migrate-up migrate-down migrate-status

# Install native Git hooks for Conventional Commits and Atomic Commits validation
hooks-install:
	@echo "🪝 Configuring Git hooks from .githooks..."
	@git config core.hooksPath .githooks
	@chmod +x .githooks/*
	@echo "✅ Git hooks installed successfully"

# Generate Swagger documentation
docs:
	@echo "📚 Generating Swagger documentation..."
	@swag init -g main.go -d cmd/api,internal/handler,internal/domain,internal/service --packagePrefix github.com/matheuseufrasio2/ponto_certo --parseInternal

# Development mode with hot-reload (Air) or fallback to go run
dev: migrate
	@if ! command -v air > /dev/null 2>&1 && [ ! -f "$$(go env GOPATH)/bin/air" ]; then \
		echo "⬇️  Air not found. Installing Air..."; \
		go install github.com/air-verse/air@latest; \
	fi
	@if command -v air > /dev/null 2>&1; then \
		air; \
	else \
		$$(go env GOPATH)/bin/air; \
	fi

# Install Air for hot-reload
air-install:
	@echo "⬇️  Installing Air..."
	@go install github.com/air-verse/air@latest

# Compile production binary
build: docs
	@echo "🔨 Building API binary..."
	@mkdir -p bin
	@go build -o bin/api cmd/api/main.go

# Run test suite
test:
	@echo "🧪 Running test suite..."
	@go test -v ./...

# Database targets (Docker Compose)
db-up:
	@echo "🐘 Starting PostgreSQL container..."
	@docker compose up -d postgres
	@echo "⏳ Waiting for PostgreSQL to be healthy..."
	@until docker compose exec -T postgres pg_isready -U postgres -d ponto_certo > /dev/null 2>&1; do \
		sleep 1; \
	done
	@echo "✅ PostgreSQL is ready on localhost:5432"

db-down:
	@echo "🛑 Stopping PostgreSQL container..."
	@docker compose down

db-logs:
	@docker compose logs -f postgres

db-shell:
	@docker compose exec -it postgres psql -U postgres -d ponto_certo

db-reset:
	@echo "⚠️  Resetting database (dropping volumes)..."
	@docker compose down -v
	@$(MAKE) db-up

# Seed database with 20 realistic tenants & employees (Native Go CLI, agnostic for Neon Cloud & Docker)
seed:
	@go run cmd/seed/main.go

# Legacy alias
db-seed: seed

# Database migrations (Native Go CLI, agnostic for Neon Cloud and local Docker)
migrate:
	@go run cmd/migrate/main.go up

migrate-status:
	@go run cmd/migrate/main.go status

migrate-down:
	@go run cmd/migrate/main.go down

# Backwards compatibility aliases
migrate-up: migrate

# Clean up build artifacts and temporary files
clean:
	@rm -rf bin tmp build-errors.log
