.PHONY: fmt test test-cover test-short test-integration dev run build clean

# ============================================================================
# Configuration
# ============================================================================

DOCKER_CONTAINER := nexus-postgres
DB_USER := nexus
DB_NAME := nexus

# Auto-discover all bounded contexts with schema.sql files
SCHEMA_FILES := $(shell find internal -name "schema.sql" -path "*/persistence/postgres/*" 2>/dev/null)
CONTEXTS := $(shell find internal -name "schema.sql" -path "*/persistence/postgres/*" -exec dirname {} \; 2>/dev/null | xargs -I{} dirname {} | xargs -I{} dirname {} | xargs -I{} dirname {} | xargs -I{} basename {} | sort -u)

# ============================================================================
# Development
# ============================================================================

# Run with live reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
	@air

# Run without live reload
run:
	@go run ./cmd/server

# ============================================================================
# Build
# ============================================================================

build:
	@go build -o ./tmp/api ./cmd/server

clean:
	@rm -rf ./tmp
	@rm -f coverage.out coverage.html

# ============================================================================
# Testing
# ============================================================================

fmt:
	@gofmt -w -s .
	@gofmt -l .

test:
	@go test -v ./...

test-cover:
	@go test -v -cover -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

test-short:
	@go test -v -short ./...

test-integration:
	@echo "Running integration tests..."
	@go test -tags=integration -v ./...

# ============================================================================
# Database Migrations
# ============================================================================

# List discovered contexts
migrate-list:
	@echo "Discovered bounded contexts with schemas:"
	@for ctx in $(CONTEXTS); do echo "  - $$ctx"; done

# Run all discovered migrations (up)
migrate-up:
	@echo "Running all migrations..."
	@for schema in $(SCHEMA_FILES); do \
		ctx=$$(echo $$schema | sed 's|internal/||' | cut -d'/' -f1); \
		echo "  Migrating: $$ctx"; \
		docker exec -i $(DOCKER_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) < $$schema || exit 1; \
	done
	@echo "All migrations complete"

# Run migration for a specific context: make migrate-context CTX=identity
migrate-context:
	@if [ -z "$(CTX)" ]; then \
		echo "Usage: make migrate-context CTX=<context_name>"; \
		echo "Available contexts:"; \
		for ctx in $(CONTEXTS); do echo "  - $$ctx"; done; \
		exit 1; \
	fi
	@schema="internal/$(CTX)/infrastructure/persistence/postgres/schema.sql"; \
	if [ -f "$$schema" ]; then \
		echo "Running $(CTX) migrations..."; \
		docker exec -i $(DOCKER_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) < $$schema; \
		echo "$(CTX) migrations complete"; \
	else \
		echo "Schema not found: $$schema"; \
		exit 1; \
	fi

# Drop all tables (DESTRUCTIVE) - drops everything in public schema
migrate-down:
	@echo "WARNING: This will drop all tables in the database!"
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ] || exit 1
	@echo "Dropping all tables..."
	@docker exec -i $(DOCKER_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "\
		DO \$$\$$ \
		DECLARE r RECORD; \
		BEGIN \
			FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP \
				EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE'; \
			END LOOP; \
		END \$$\$$;"
	@echo "All tables dropped"

# Reset database (down + up)
db-reset:
	@echo "WARNING: This will reset the entire database!"
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ] || exit 1
	@$(MAKE) migrate-down-force
	@$(MAKE) migrate-up
	@echo "Database reset complete"

# Force drop without confirmation (for scripting)
migrate-down-force:
	@echo "Dropping all tables..."
	@docker exec -i $(DOCKER_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "\
		DO \$$\$$ \
		DECLARE r RECORD; \
		BEGIN \
			FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP \
				EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE'; \
			END LOOP; \
		END \$$\$$;"
	@echo "All tables dropped"

# Database utilities
db-status:
	@echo "Database tables:"
	@docker exec -it $(DOCKER_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "\dt"

db-status-full:
	@echo "Database tables with details:"
	@docker exec -it $(DOCKER_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "\dt+"

db-shell:
	@docker exec -it $(DOCKER_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME)

db-schema:
	@echo "Database schema for all tables:"
	@docker exec -it $(DOCKER_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "\d+"

# ============================================================================
# Code Generation
# ============================================================================

# Auto-discover and run sqlc for all contexts
sqlc:
	@echo "Generating SQLC code..."
	@for dir in $$(find internal -name "sqlc.yaml" -exec dirname {} \;); do \
		ctx=$$(echo $$dir | sed 's|internal/||' | cut -d'/' -f1); \
		echo "  Generating: $$ctx"; \
		cd $$dir && sqlc generate && cd - > /dev/null; \
	done
	@echo "SQLC generation complete"

# Generate sqlc for a specific context: make sqlc-context CTX=identity
sqlc-context:
	@if [ -z "$(CTX)" ]; then \
		echo "Usage: make sqlc-context CTX=<context_name>"; \
		exit 1; \
	fi
	@dir="internal/$(CTX)/infrastructure/persistence/postgres"; \
	if [ -f "$$dir/sqlc.yaml" ]; then \
		echo "Generating SQLC for $(CTX)..."; \
		cd $$dir && sqlc generate; \
		echo "$(CTX) SQLC generation complete"; \
	else \
		echo "sqlc.yaml not found: $$dir/sqlc.yaml"; \
		exit 1; \
	fi

# ============================================================================
# Dependencies
# ============================================================================

deps:
	@go mod download
	@go mod tidy

tools:
	@go install github.com/cosmtrek/air@latest
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# ============================================================================
# Docker
# ============================================================================

docker-up:
	@docker compose -f ../deployments/docker/docker-compose.yml up -d

docker-down:
	@docker compose -f ../deployments/docker/docker-compose.yml down

docker-logs:
	@docker compose -f ../deployments/docker/docker-compose.yml logs -f

docker-ps:
	@docker compose -f ../deployments/docker/docker-compose.yml ps

# ============================================================================
# CI Pipeline
# ============================================================================

ci: fmt test
	@echo "CI checks passed"

ci-full: fmt test test-integration
	@echo "Full CI checks passed"

# ============================================================================
# Help
# ============================================================================

help:
	@echo "Nexus Platform Commands:"
	@echo ""
	@echo "Development:"
	@echo "  make dev                  - Run with live reload (requires air)"
	@echo "  make run                  - Run without live reload"
	@echo "  make build                - Build binary"
	@echo "  make clean                - Clean build artifacts"
	@echo ""
	@echo "Testing:"
	@echo "  make fmt                  - Format code"
	@echo "  make test                 - Run unit tests"
	@echo "  make test-cover           - Run tests with coverage"
	@echo "  make test-short           - Run short tests"
	@echo "  make test-integration     - Run all integration tests"
	@echo ""
	@echo "Database:"
	@echo "  make migrate-list         - List discovered bounded contexts"
	@echo "  make migrate-up           - Run all discovered migrations"
	@echo "  make migrate-context CTX=x- Run migration for specific context"
	@echo "  make migrate-down         - Drop all tables (DESTRUCTIVE, with confirmation)"
	@echo "  make db-status            - Show database tables"
	@echo "  make db-status-full       - Show database tables with details"
	@echo "  make db-schema            - Show full schema for all tables"
	@echo "  make db-shell             - Open psql shell"
	@echo "  make db-reset             - Reset database (down + up)"
	@echo ""
	@echo "Code Generation:"
	@echo "  make sqlc                 - Generate SQLC code for all contexts"
	@echo "  make sqlc-context CTX=x   - Generate SQLC for specific context"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-up            - Start all services"
	@echo "  make docker-down          - Stop all services"
	@echo "  make docker-logs          - View service logs"
	@echo "  make docker-ps            - Show running services"
	@echo ""
	@echo "CI:"
	@echo "  make ci                   - Run CI checks (fmt + test)"
	@echo "  make ci-full              - Run full CI (fmt + test + integration)"
	@echo ""
	@echo "Dependencies:"
	@echo "  make deps                 - Download dependencies"
	@echo "  make tools                - Install development tools"