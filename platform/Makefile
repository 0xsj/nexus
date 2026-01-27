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
# Database Migrations
# ============================================================================

# Run all migrations (up)
migrate-up: migrate-identity migrate-wallet migrate-credential migrate-verification
	@echo "All migrations complete"

# Run all migrations (down) - DESTRUCTIVE
migrate-down: migrate-verification-down migrate-credential-down migrate-wallet-down migrate-identity-down
	@echo "All migrations rolled back"

# Individual module migrations (up)
migrate-identity:
	@echo "Running identity migrations..."
	@docker exec -i nexus-postgres psql -U nexus -d nexus < internal/identity/infrastructure/persistence/postgres/migrations/001_create_identity_tables.sql
	@docker exec -i nexus-postgres psql -U nexus -d nexus < internal/identity/infrastructure/persistence/postgres/migrations/002_alter_identity_tables.sql
	@echo "Identity migrations complete"

migrate-wallet:
	@echo "Running wallet migrations..."
	@docker exec -i nexus-postgres psql -U nexus -d nexus < internal/wallet/infrastructure/persistence/postgres/migrations/001_create_wallet_tables.sql
	@echo "Wallet migrations complete"

migrate-credential:
	@echo "Running credential migrations..."
	@docker exec -i nexus-postgres psql -U nexus -d nexus < internal/credential/infrastructure/persistence/postgres/migrations/001_create_credentials.sql
	@echo "Credential migrations complete"

migrate-verification:
	@echo "Running verification migrations..."
	@docker exec -i nexus-postgres psql -U nexus -d nexus < internal/verification/infrastructure/persistence/postgres/migrations/001_create_verification_tables.sql
	@echo "Verification migrations complete"

# Individual module migrations (down) - DESTRUCTIVE
migrate-identity-down:
	@echo "Rolling back identity migrations..."
	@echo "WARNING: This will drop identity tables!"
	@docker exec -i nexus-postgres psql -U nexus -d nexus -c "DROP TABLE IF EXISTS user_emails, user_wallets, api_keys, sessions, connections, users CASCADE;"

migrate-wallet-down:
	@echo "Rolling back wallet migrations..."
	@echo "WARNING: This will drop wallet tables!"
	@docker exec -i nexus-postgres psql -U nexus -d nexus -c "DROP TABLE IF EXISTS wallet_challenges, wallets CASCADE;"

migrate-credential-down:
	@echo "Rolling back credential migrations..."
	@docker exec -i nexus-postgres psql -U nexus -d nexus < internal/credential/infrastructure/persistence/postgres/migrations/001_create_credentials_down.sql

migrate-verification-down:
	@echo "Rolling back verification migrations..."
	@docker exec -i nexus-postgres psql -U nexus -d nexus < internal/verification/infrastructure/persistence/postgres/migrations/001_create_verification_tables_down.sql

# Database utilities
migrate-status:
	@echo "Database tables:"
	@docker exec -it nexus-postgres psql -U nexus -d nexus -c "\dt"

migrate-status-full:
	@echo "Database tables with details:"
	@docker exec -it nexus-postgres psql -U nexus -d nexus -c "\dt+"

db-shell:
	@docker exec -it nexus-postgres psql -U nexus -d nexus

db-reset: migrate-down migrate-up
	@echo "Database reset complete"

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
	@echo "  make dev              - Run with live reload (requires air)"
	@echo "  make run              - Run without live reload"
	@echo "  make build            - Build binary"
	@echo "  make clean            - Clean build artifacts"
	@echo ""
	@echo "Testing:"
	@echo "  make fmt              - Format code"
	@echo "  make test             - Run all tests"
	@echo "  make test-cover       - Run tests with coverage"
	@echo "  make test-short       - Run short tests"
	@echo ""
	@echo "Database Migrations:"
	@echo "  make migrate-up       - Run all migrations"
	@echo "  make migrate-down     - Roll back all migrations (DESTRUCTIVE)"
	@echo "  make migrate-status   - Show database tables"
	@echo "  make db-shell         - Open psql shell"
	@echo "  make db-reset         - Reset database (down + up)"
	@echo ""
	@echo "Individual Migrations:"
	@echo "  make migrate-identity      - Run identity migrations"
	@echo "  make migrate-wallet        - Run wallet migrations"
	@echo "  make migrate-credential    - Run credential migrations"
	@echo "  make migrate-verification  - Run verification migrations"
	@echo ""
	@echo "Dependencies:"
	@echo "  make deps             - Download dependencies"
	@echo "  make tools            - Install development tools"