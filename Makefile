.PHONY: all docs dev build test clean air-install db-up db-down db-logs db-shell db-reset migrate-up migrate-down migrate-status

# Generate Swagger documentation
docs:
	@echo "📚 Generating Swagger documentation..."
	@swag init -g cmd/api/main.go --parseDependency --parseInternal

# Development mode with hot-reload (Air) or fallback to go run
dev:
	@if command -v air > /dev/null 2>&1; then \
		air; \
	else \
		echo "💡 Tip: 'air' binary not found. Run 'make air-install' to enable hot-reload."; \
		$(MAKE) docs; \
		go run cmd/api/main.go; \
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

# Database migrations using official golang-migrate container
migrate-up:
	@echo "🚀 Applying database migrations..."
	@docker compose --profile tools run --rm migrate up

migrate-down:
	@echo "⏪ Reverting last database migration..."
	@docker compose --profile tools run --rm migrate down 1

migrate-status:
	@echo "📊 Checking migration version..."
	@docker compose --profile tools run --rm migrate version

# Clean up build artifacts and temporary files
clean:
	@rm -rf bin tmp build-errors.log
