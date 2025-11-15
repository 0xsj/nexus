# ============================================================================
# Configuration
# ============================================================================

DB_USER ?= nexus_user
DB_PASS ?= nexus_pass
DB_HOST ?= localhost
DB_PORT ?= 5433
DB_NAME ?= nexus
DB_SSLMODE ?= disable

# Construct DATABASE_URL
DATABASE_URL = postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

# gRPC Server
GRPC_PORT ?= 9090

# HTTP Gateway
HTTP_PORT ?= 8080

# Project
PROJECT_NAME = nexus

# ============================================================================
# Help
# ============================================================================

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ============================================================================
# Docker Commands
# ============================================================================

.PHONY: docker-up
docker-up: ## Start core infrastructure (postgres, redis, nats)
	docker-compose -f deployments/docker/docker-compose.yml --env-file deployments/docker/.env.docker up -d
	@echo "✅ Core infrastructure started"
	@echo "PostgreSQL: localhost:5433"
	@echo "Redis: localhost:6380"
	@echo "NATS: localhost:4223 (HTTP: localhost:8223)"

.PHONY: docker-up-full
docker-up-full: docker-up ## Start core + observability stack
	docker-compose -f deployments/docker/docker-compose.observability.yml --env-file deployments/docker/.env.docker up -d
	@echo "✅ Full stack started (core + observability)"
	@echo "Prometheus: http://localhost:9091"
	@echo "Grafana: http://localhost:3001 (admin/admin)"
	@echo "Jaeger UI: http://localhost:16687"

.PHONY: docker-down
docker-down: ## Stop all services
	docker-compose -f deployments/docker/docker-compose.yml down
	docker-compose -f deployments/docker/docker-compose.observability.yml down 2>/dev/null || true

.PHONY: docker-down-volumes
docker-down-volumes: ## Stop all services and remove volumes (⚠️  DELETES DATA)
	docker-compose -f deployments/docker/docker-compose.yml down -v
	docker-compose -f deployments/docker/docker-compose.observability.yml down -v 2>/dev/null || true
	@echo "⚠️  All data volumes removed"

.PHONY: docker-logs
docker-logs: ## Show logs for core services
	docker-compose -f deployments/docker/docker-compose.yml logs -f

.PHONY: docker-logs-obs
docker-logs-obs: ## Show logs for observability services
	docker-compose -f deployments/docker/docker-compose.observability.yml logs -f

.PHONY: docker-ps
docker-ps: ## Show running services
	@docker-compose -f deployments/docker/docker-compose.yml ps
	@echo ""
	@docker-compose -f deployments/docker/docker-compose.observability.yml ps 2>/dev/null || true

.PHONY: docker-restart
docker-restart: docker-down docker-up ## Restart core services

# ============================================================================
# Database Commands
# ============================================================================

.PHONY: db-shell
db-shell: ## Connect to PostgreSQL shell
	docker-compose -f deployments/docker/docker-compose.yml exec postgres psql -U nexus_user -d nexus

# ============================================================================
# Migration Commands
# ============================================================================

.PHONY: migrate-install
migrate-install: ## Install golang-migrate tool
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "✅ golang-migrate installed"

.PHONY: migrate-create
migrate-create: ## Create a new migration (usage: make migrate-create name=create_users_table)
	@if [ -z "$(name)" ]; then \
		echo "Error: name is required. Usage: make migrate-create name=create_users_table"; \
		exit 1; \
	fi
	migrate create -ext sql -dir migrations -seq $(name)
	@echo "✅ Migration files created in migrations/"

.PHONY: migrate-up
migrate-up: ## Apply all pending migrations
	@echo "Applying migrations to: $(DB_HOST):$(DB_PORT)/$(DB_NAME)"
	migrate -path migrations -database "$(DATABASE_URL)" up
	@echo "✅ Migrations applied"

.PHONY: migrate-up-one
migrate-up-one: ## Apply next pending migration
	migrate -path migrations -database "$(DATABASE_URL)" up 1
	@echo "✅ One migration applied"

.PHONY: migrate-down
migrate-down: ## Rollback last migration
	@echo "⚠️  Rolling back last migration from: $(DB_HOST):$(DB_PORT)/$(DB_NAME)"
	migrate -path migrations -database "$(DATABASE_URL)" down 1
	@echo "⚠️  Last migration rolled back"

.PHONY: migrate-down-all
migrate-down-all: ## Rollback all migrations (⚠️  DANGEROUS)
	@echo "⚠️  WARNING: This will rollback ALL migrations from $(DB_HOST):$(DB_PORT)/$(DB_NAME)"
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ]
	migrate -path migrations -database "$(DATABASE_URL)" down -all
	@echo "⚠️  All migrations rolled back"

.PHONY: migrate-version
migrate-version: ## Show current migration version
	@migrate -path migrations -database "$(DATABASE_URL)" version

.PHONY: migrate-force
migrate-force: ## Force migration version (usage: make migrate-force version=1)
	@if [ -z "$(version)" ]; then \
		echo "Error: version is required. Usage: make migrate-force version=1"; \
		exit 1; \
	fi
	@echo "⚠️  WARNING: Forcing migration version to $(version)"
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ]
	migrate -path migrations -database "$(DATABASE_URL)" force $(version)
	@echo "⚠️  Migration version forced to $(version)"

.PHONY: migrate-sync
migrate-sync: ## Copy migrations from root to embedded directory
	@echo "Syncing migrations to embedded directory..."
	@mkdir -p pkg/database/postgres/migrations
	@rm -f pkg/database/postgres/migrations/*.sql
	@cp migrations/*.sql pkg/database/postgres/migrations/ 2>/dev/null || echo "No migrations to sync yet"
	@echo "✅ Migrations synced"

# ============================================================================
# Redis Commands
# ============================================================================

.PHONY: redis-cli
redis-cli: ## Connect to Redis CLI
	docker-compose -f deployments/docker/docker-compose.yml exec redis redis-cli

.PHONY: redis-flush
redis-flush: ## Flush all Redis data (⚠️  DELETES DATA)
	docker-compose -f deployments/docker/docker-compose.yml exec redis redis-cli FLUSHALL
	@echo "⚠️  Redis data flushed"

# ============================================================================
# NATS Commands
# ============================================================================

.PHONY: nats-stream-info
nats-stream-info: ## Show NATS JetStream stream info
	@echo "NATS Stream Information:"
	@docker exec nexus-nats nats stream info events 2>/dev/null || \
		echo "⚠️  Stream 'events' not found or NATS CLI not available in container"

.PHONY: nats-stream-list
nats-stream-list: ## List all NATS streams
	@docker exec nexus-nats nats stream list 2>/dev/null || \
		echo "⚠️  NATS CLI not available in container"

.PHONY: nats-purge-stream
nats-purge-stream: ## Purge all messages from events stream (⚠️ DELETES EVENTS)
	@echo "⚠️  WARNING: This will purge ALL events from the stream"
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ]
	@docker exec nexus-nats nats stream purge events --force 2>/dev/null || \
		echo "⚠️  Failed to purge stream"
	@echo "⚠️  Stream purged"

.PHONY: nats-consumers
nats-consumers: ## Show NATS consumers for events stream
	@docker exec nexus-nats nats consumer list events 2>/dev/null || \
		echo "⚠️  NATS CLI not available in container"

# ============================================================================
# Proto & Code Generation
# ============================================================================

.PHONY: proto-install
proto-install: ## Install buf CLI
	@echo "Installing buf..."
	@go install github.com/bufbuild/buf/cmd/buf@latest
	@echo "✅ buf installed"

.PHONY: proto-lint
proto-lint: ## Lint proto files
	@echo "Linting proto files..."
	cd proto && buf lint
	@echo "✅ Proto files linted"

.PHONY: proto-breaking
proto-breaking: ## Check for breaking changes in proto files
	@echo "Checking for breaking changes..."
	cd proto && buf breaking --against '.git#branch=main'
	@echo "✅ No breaking changes detected"

.PHONY: proto-generate
proto-generate: ## Generate code from proto files
	@echo "Generating code from proto files..."
	cd proto && buf generate
	@echo "✅ Proto code generated"

.PHONY: proto-clean
proto-clean: ## Clean generated proto code
	@echo "Cleaning generated proto code..."
	rm -rf api/
	@echo "✅ Proto code cleaned"

.PHONY: swagger
swagger: ## Start Swagger UI server
	@echo "Starting Swagger UI at http://localhost:8081"
	@docker run -p 8081:8080 --rm \
		-e SWAGGER_JSON=/openapi/openapi.swagger.json \
		-v $(PWD)/openapi:/openapi \
		swaggerapi/swagger-ui

# ============================================================================
# Wire Dependency Injection
# ============================================================================

.PHONY: wire
wire: ## Generate Wire dependency injection code
	@echo "Generating Wire dependencies..."
	@if [ -d "pkg/di" ]; then cd pkg/di && go run github.com/google/wire/cmd/wire@latest; fi
	@if [ -d "cmd/server" ]; then cd cmd/server && go run github.com/google/wire/cmd/wire@latest; fi
	@echo "✅ Wire dependencies generated"

# ============================================================================
# SQLC Code Generation
# ============================================================================

.PHONY: sqlc
sqlc: ## Generate sqlc code
	@echo "Generating sqlc code..."
	@if [ -d "internal/users/infrastructure/postgres" ]; then \
		cd internal/users/infrastructure/postgres && sqlc generate; \
	fi
	@echo "✅ sqlc code generated"

# ============================================================================
# Generate All Code
# ============================================================================

.PHONY: generate
generate: proto-generate wire sqlc ## Generate all code (proto + wire + sqlc)
	@echo "✅ All code generated"

# ============================================================================
# Application Commands
# ============================================================================

.PHONY: build
build: ## Build the application
	@echo "Building $(PROJECT_NAME)..."
	go build -o bin/server ./cmd/server
	@echo "✅ Built bin/server"

.PHONY: run
run: ## Run the application
	@echo "Running $(PROJECT_NAME)..."
	go run ./cmd/server

.PHONY: dev
dev: docker-up ## Start infrastructure and run application
	@echo "Starting development server..."
	@sleep 2
	@make run

# ============================================================================
# Testing
# ============================================================================

.PHONY: test
test: ## Run tests
	go test -v -race -cover ./...

.PHONY: test-unit
test-unit: ## Run unit tests only
	go test -v -race -short ./...

.PHONY: test-integration
test-integration: docker-up ## Run integration tests
	go test -v -race -tags=integration ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# ============================================================================
# Code Quality
# ============================================================================

.PHONY: lint
lint: ## Run linter
	golangci-lint run --timeout 5m

.PHONY: fmt
fmt: ## Format code
	go fmt ./...
	goimports -w .

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: tidy
tidy: ## Tidy go modules
	go mod tidy

# ============================================================================
# Dependency Management
# ============================================================================

.PHONY: deps
deps: ## Download dependencies
	go mod download

.PHONY: deps-upgrade
deps-upgrade: ## Upgrade all dependencies
	go get -u ./...
	go mod tidy

# ============================================================================
# Clean
# ============================================================================

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -rf tmp/

.PHONY: clean-all
clean-all: clean proto-clean docker-down-volumes ## Clean everything including Docker volumes
	@echo "✅ Everything cleaned"

# ============================================================================
# Development Tools
# ============================================================================

.PHONY: dev-tools
dev-tools: proto-install migrate-install ## Install all development tools
	@echo "Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "✅ Development tools installed"

# ============================================================================
# Git Hooks
# ============================================================================

.PHONY: setup-hooks
setup-hooks: ## Set up git hooks
	@echo "Setting up git hooks..."
	@mkdir -p .git/hooks
	@echo '#!/bin/sh' > .git/hooks/pre-commit
	@echo 'make fmt' >> .git/hooks/pre-commit
	@echo 'make vet' >> .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "✅ Git hooks installed"

# ============================================================================
# Hot Reload Development
# ============================================================================

.PHONY: dev-watch
dev-watch: ## Run with hot-reload using air
	@echo "Starting development server with hot-reload..."
	@if ! command -v air > /dev/null; then \
		echo "❌ Air not installed. Run: make dev-tools"; \
		exit 1; \
	fi
	air

.PHONY: dev-race
dev-race: ## Run with race detector
	@echo "Running with race detector..."
	air -- -race

.PHONY: dev-clean
dev-clean: ## Clean development artifacts
	@echo "Cleaning development artifacts..."
	rm -rf tmp/
	rm -f build-errors.log
	@echo "✅ Development artifacts cleaned"

# ============================================================================
# Full Setup
# ============================================================================

.PHONY: setup
setup: dev-tools docker-up migrate-sync migrate-up ## Full project setup
	@echo "✅ Project setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Copy .env.example to .env and configure"
	@echo "  2. Run 'make generate' to generate code"
	@echo "  3. Run 'make dev' to start the server"

.PHONY: reset
reset: clean-all setup ## Reset and setup project from scratch
	@echo "✅ Project reset complete!"

# ============================================================================
# gRPC Server Commands
# ============================================================================

.PHONY: grpc-start
grpc-start: ## Start gRPC server
	@echo "Starting gRPC server on port $(GRPC_PORT)..."
	go run ./cmd/server

.PHONY: grpc-health
grpc-health: ## Check gRPC server health
	@echo "Checking gRPC server health on localhost:$(GRPC_PORT)..."
	@grpcurl -plaintext localhost:$(GRPC_PORT) grpc.health.v1.Health/Check || \
		echo "❌ Server not responding or grpcurl not installed"

.PHONY: grpc-list
grpc-list: ## List available gRPC services
	@echo "Available gRPC services on localhost:$(GRPC_PORT):"
	@grpcurl -plaintext localhost:$(GRPC_PORT) list || \
		echo "❌ Server not running or grpcurl not installed. Run: brew install grpcurl"

.PHONY: grpc-describe
grpc-describe: ## Describe a gRPC service (usage: make grpc-describe service=users.v1.UserService)
	@if [ -z "$(service)" ]; then \
		echo "Error: service is required. Usage: make grpc-describe service=users.v1.UserService"; \
		exit 1; \
	fi
	@grpcurl -plaintext localhost:$(GRPC_PORT) describe $(service)

.PHONY: grpc-call
grpc-call: ## Call a gRPC method (usage: make grpc-call method=users.v1.UserService/GetUser data='{"user_id":"123"}')
	@if [ -z "$(method)" ]; then \
		echo "Error: method is required. Usage: make grpc-call method=users.v1.UserService/GetUser data='{\"user_id\":\"123\"}'"; \
		exit 1; \
	fi
	@if [ -z "$(data)" ]; then \
		grpcurl -plaintext localhost:$(GRPC_PORT) $(method); \
	else \
		grpcurl -plaintext -d '$(data)' localhost:$(GRPC_PORT) $(method); \
	fi

.PHONY: grpc-install-tools
grpc-install-tools: ## Install grpcurl for testing gRPC endpoints
	@echo "Installing grpcurl..."
	@if command -v brew > /dev/null; then \
		brew install grpcurl; \
	else \
		go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest; \
	fi
	@echo "✅ grpcurl installed"

.PHONY: grpc-gateway-start
grpc-gateway-start: ## Start gRPC-Gateway (REST proxy)
	@echo "Starting gRPC-Gateway on port $(HTTP_PORT)..."
	@echo "⚠️  Not yet implemented - will proxy REST → gRPC"
	# TODO: Implement when we add HTTP gateway server

.PHONY: grpc-test-create-user
grpc-test-create-user: ## Test CreateUser endpoint
	@echo "Testing CreateUser..."
	@grpcurl -plaintext \
		-d '{"email":"test@example.com","username":"testuser","password":"password123"}' \
		localhost:$(GRPC_PORT) users.v1.UserService/CreateUser

.PHONY: grpc-test-get-user
grpc-test-get-user: ## Test GetUser endpoint (usage: make grpc-test-get-user id=USER_ID)
	@if [ -z "$(id)" ]; then \
		echo "Error: id is required. Usage: make grpc-test-get-user id=USER_ID"; \
		exit 1; \
	fi
	@echo "Testing GetUser..."
	@grpcurl -plaintext \
		-d '{"user_id":"$(id)"}' \
		localhost:$(GRPC_PORT) users.v1.UserService/GetUser