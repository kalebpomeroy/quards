.PHONY: migrate migrate-help dev build clean test

# Database migrations
migrate:
	@echo "Running database migrations..."
	@go run cmd/migrate/main.go

migrate-help:
	@go run cmd/migrate/main.go -h

# Development
dev:
	@echo "Starting development server with live reload..."
	@air

# Build
build:
	@echo "Building application..."
	@go build -o bin/quards main.go
	@go build -o bin/migrate cmd/migrate/main.go

# Clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf tmp/

# Test
test:
	@echo "Running tests..."
	@go test ./...

# Help
help:
	@echo "Available commands:"
	@echo "  migrate       - Run database migrations"
	@echo "  migrate-help  - Show migration help"
	@echo "  dev          - Start development server with live reload"
	@echo "  build        - Build application binaries"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  help         - Show this help message"