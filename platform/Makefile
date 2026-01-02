.PHONY: fmt test test-cover test-short dev run build clean

# ============================================================================
# Development
# ============================================================================

# Run with live reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
	@air

# Run without live reload
run:
	@go run ./cmd/api

# ============================================================================
# Build
# ============================================================================

build:
	@go build -o ./tmp/api ./cmd/api

clean:
	@rm -rf ./tmp
	@rm -f coverage.out coverage.html

# ============================================================================
# Testing
# ============================================================================

fmt:
	@go fmt ./...

test:
	@go test -v ./...

test-cover:
	@go test -v -cover -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

test-short:
	@go test -v -short ./...

# ============================================================================
# Database
# ============================================================================

migrate-up:
	@echo "Running migrations..."
	@docker exec -i nexus-postgres psql -U nexus -d nexus < internal/credential/infrastructure/persistence/postgres/migrations/001_create_credentials.sql
	@echo "Migrations complete"

migrate-status:
	@docker exec -it nexus-postgres psql -U nexus -d nexus -c "\dt"

db-shell:
	@docker exec -it nexus-postgres psql -U nexus -d nexus

# ============================================================================
# Dependencies
# ============================================================================

deps:
	@go mod download
	@go mod tidy

tools:
	@go install github.com/cosmtrek/air@latest

# ============================================================================
# Help
# ============================================================================

help:
	@echo "Nexus Platform Commands:"
	@echo ""
	@echo "Development:"
	@echo "  make dev          - Run with live reload (requires air)"
	@echo "  make run          - Run without live reload"
	@echo "  make build        - Build binary"
	@echo "  make clean        - Clean build artifacts"
	@echo ""
	@echo "Testing:"
	@echo "  make fmt          - Format code"
	@echo "  make test         - Run all tests"
	@echo "  make test-cover   - Run tests with coverage"
	@echo "  make test-short   - Run short tests"
	@echo ""
	@echo "Database:"
	@echo "  make migrate-up   - Run migrations"
	@echo "  make migrate-status - Show tables"
	@echo "  make db-shell     - Open psql shell"
	@echo ""
	@echo "Dependencies:"
	@echo "  make deps         - Download dependencies"
	@echo "  make tools        - Install development tools"