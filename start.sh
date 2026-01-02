#!/bin/bash

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project root
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOCKER_DIR="${PROJECT_ROOT}/deployments/docker"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Commands
cmd_up() {
    log_info "Starting Nexus services..."
    docker-compose -f "${DOCKER_DIR}/docker-compose.yml" up -d "$@"
    log_success "Nexus services started"
    echo ""
    cmd_status
}

cmd_down() {
    log_info "Stopping Nexus services..."
    docker-compose -f "${DOCKER_DIR}/docker-compose.yml" down "$@"
    log_success "Nexus services stopped"
}

cmd_restart() {
    log_info "Restarting Nexus services..."
    docker-compose -f "${DOCKER_DIR}/docker-compose.yml" restart "$@"
    log_success "Nexus services restarted"
}

cmd_logs() {
    docker-compose -f "${DOCKER_DIR}/docker-compose.yml" logs "$@"
}

cmd_status() {
    echo -e "${BLUE}Nexus Services:${NC}"
    echo "────────────────────────────────────────────────────"
    printf "%-15s %-10s %s\n" "SERVICE" "PORT" "URL"
    echo "────────────────────────────────────────────────────"
    printf "%-15s %-10s %s\n" "API" "8090" "http://localhost:8090"
    printf "%-15s %-10s %s\n" "Frontend" "3010" "http://localhost:3010"
    printf "%-15s %-10s %s\n" "PostgreSQL" "5439" "postgres://nexus:nexus@localhost:5439/nexus"
    printf "%-15s %-10s %s\n" "Redis" "6390" "redis://localhost:6390"
    printf "%-15s %-10s %s\n" "Vault" "8200" "http://localhost:8200"
    printf "%-15s %-10s %s\n" "NATS" "4222" "nats://localhost:4222"
    printf "%-15s %-10s %s\n" "NATS Monitor" "8222" "http://localhost:8222"
    printf "%-15s %-10s %s\n" "MinIO API" "9010" "http://localhost:9010"
    printf "%-15s %-10s %s\n" "MinIO Console" "9011" "http://localhost:9011"
    printf "%-15s %-10s %s\n" "Mailpit SMTP" "1030" "smtp://localhost:1030"
    printf "%-15s %-10s %s\n" "Mailpit UI" "8030" "http://localhost:8030"
    echo "────────────────────────────────────────────────────"
    echo ""
    docker-compose -f "${DOCKER_DIR}/docker-compose.yml" ps
}

cmd_ps() {
    docker-compose -f "${DOCKER_DIR}/docker-compose.yml" ps "$@"
}

cmd_db() {
    log_info "Connecting to PostgreSQL..."
    docker exec -it nexus-postgres psql -U nexus -d nexus
}

cmd_redis() {
    log_info "Connecting to Redis..."
    docker exec -it nexus-redis redis-cli
}

cmd_clean() {
    log_warn "This will remove all Nexus containers and volumes!"
    read -p "Are you sure? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log_info "Cleaning up Nexus..."
        docker-compose -f "${DOCKER_DIR}/docker-compose.yml" down -v --remove-orphans
        log_success "Nexus cleaned up"
    fi
}

cmd_build() {
    log_info "Building Nexus images..."
    docker-compose -f "${DOCKER_DIR}/docker-compose.yml" build "$@"
    log_success "Build complete"
}

cmd_migrate() {
    log_info "Running migrations..."
    # For now, just execute SQL files directly
    # Later can integrate with golang-migrate or similar
    if [ -n "$1" ]; then
        docker exec -i nexus-postgres psql -U nexus -d nexus < "$1"
        log_success "Migration applied: $1"
    else
        log_warn "Usage: $0 migrate <migration_file.sql>"
    fi
}

cmd_help() {
    echo "Nexus Development CLI"
    echo ""
    echo "Usage: $0 <command> [options]"
    echo ""
    echo "Commands:"
    echo "  up [services...]      Start services (default: all)"
    echo "  down                  Stop services"
    echo "  restart [services...] Restart services"
    echo "  logs [services...]    View logs (-f for follow)"
    echo "  status                Show service status and URLs"
    echo "  ps                    List running containers"
    echo "  db                    Connect to PostgreSQL"
    echo "  redis                 Connect to Redis"
    echo "  build [services...]   Build images"
    echo "  migrate <file>        Run a migration file"
    echo "  clean                 Remove all containers and volumes"
    echo "  help                  Show this help"
    echo ""
    echo "Examples:"
    echo "  $0 up                 # Start all services"
    echo "  $0 up postgres redis  # Start only postgres and redis"
    echo "  $0 logs -f api        # Follow API logs"
    echo "  $0 migrate platform/internal/credential/infrastructure/persistence/postgres/migrations/001_create_credentials.sql"
}

# Main
case "${1:-help}" in
    up)         shift; cmd_up "$@" ;;
    down)       shift; cmd_down "$@" ;;
    restart)    shift; cmd_restart "$@" ;;
    logs)       shift; cmd_logs "$@" ;;
    status)     cmd_status ;;
    ps)         shift; cmd_ps "$@" ;;
    db)         cmd_db ;;
    redis)      cmd_redis ;;
    build)      shift; cmd_build "$@" ;;
    migrate)    shift; cmd_migrate "$@" ;;
    clean)      cmd_clean ;;
    help|--help|-h) cmd_help ;;
    *)          log_error "Unknown command: $1"; cmd_help; exit 1 ;;
esac