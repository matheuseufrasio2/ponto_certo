.PHONY: all docs dev build test clean air-install

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

# Clean up build artifacts and temporary files
clean:
	@rm -rf bin tmp build-errors.log
