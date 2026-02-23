#!/usr/bin/env bash
set -euo pipefail

# ============================================================================
# Nexus Platform — Smoke Test
#
# Starts Postgres, runs migrations, boots the API server, and verifies
# health endpoints + a basic auth flow respond correctly.
# ============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLATFORM_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
COMPOSE_FILE="$PLATFORM_DIR/../deployments/docker/docker-compose.yml"

CONTAINER_NAME="nexus-postgres"
DB_USER="nexus"
DB_NAME="nexus"
DB_PORT=5439
SERVER_PORT=8080

SERVER_PID=""
PASS=0
FAIL=0

# ============================================================================
# Helpers
# ============================================================================

cleanup() {
    echo ""
    echo "==> Cleaning up..."
    if [[ -n "$SERVER_PID" ]] && kill -0 "$SERVER_PID" 2>/dev/null; then
        kill "$SERVER_PID" 2>/dev/null || true
        wait "$SERVER_PID" 2>/dev/null || true
        echo "    Server stopped (PID $SERVER_PID)"
    fi
    echo "    (Postgres container left running — use 'make docker-down' to stop)"
}
trap cleanup EXIT

log()  { echo "==> $*"; }
pass() { echo "    PASS: $1"; PASS=$((PASS + 1)); }
fail() { echo "    FAIL: $1"; FAIL=$((FAIL + 1)); }

wait_for() {
    local desc="$1" cmd="$2" retries="${3:-30}" delay="${4:-1}"
    log "Waiting for $desc..."
    for i in $(seq 1 "$retries"); do
        if eval "$cmd" >/dev/null 2>&1; then
            echo "    $desc ready (attempt $i/$retries)"
            return 0
        fi
        sleep "$delay"
    done
    echo "    $desc did not become ready after $retries attempts"
    return 1
}

assert_status() {
    local desc="$1" method="$2" url="$3" expected_status="$4"
    shift 4
    local response
    response=$(curl -s -o /dev/null -w "%{http_code}" -X "$method" "$url" "$@" 2>/dev/null)
    if [[ "$response" == "$expected_status" ]]; then
        pass "$desc (HTTP $response)"
    else
        fail "$desc — expected $expected_status, got $response"
    fi
}

assert_body_contains() {
    local desc="$1" method="$2" url="$3" needle="$4"
    shift 4
    local body
    body=$(curl -s -X "$method" "$url" "$@" 2>/dev/null)
    if echo "$body" | grep -q "$needle"; then
        pass "$desc (body contains '$needle')"
    else
        fail "$desc — body missing '$needle': $body"
    fi
}

# ============================================================================
# Pre-flight
# ============================================================================

log "Pre-flight checks"

if ! command -v docker &>/dev/null; then
    echo "    ERROR: docker is not installed or not in PATH"
    exit 1
fi

if ! docker info &>/dev/null; then
    echo "    ERROR: Docker daemon is not running"
    exit 1
fi
echo "    Docker is running"

if ! command -v curl &>/dev/null; then
    echo "    ERROR: curl is not installed"
    exit 1
fi

# ============================================================================
# Step 1: Start Postgres
# ============================================================================

log "Starting Postgres container..."
docker compose -f "$COMPOSE_FILE" up -d postgres

wait_for "Postgres" "docker exec $CONTAINER_NAME pg_isready -U $DB_USER -d $DB_NAME" 30 1

# ============================================================================
# Step 2: Run Migrations
# ============================================================================

log "Running migrations..."
cd "$PLATFORM_DIR"
make migrate-up

# ============================================================================
# Step 3: Build and Start Server
# ============================================================================

log "Building server..."
go build -o "$PLATFORM_DIR/tmp/smoke-api" ./cmd/server

log "Starting server on port $SERVER_PORT..."
DATABASE_PORT=$DB_PORT \
    APP_ENV=development \
    PORT=$SERVER_PORT \
    "$PLATFORM_DIR/tmp/smoke-api" &
SERVER_PID=$!

wait_for "API server" "curl -sf http://localhost:$SERVER_PORT/healthz" 15 1

# ============================================================================
# Step 4: Health Check Assertions
# ============================================================================

log "Running health check assertions..."

# Liveness probe
assert_status "GET /healthz returns 200" GET "http://localhost:$SERVER_PORT/healthz" 200
assert_body_contains "GET /healthz body has status up" GET "http://localhost:$SERVER_PORT/healthz" '"status":"up"'

# Readiness probe
assert_status "GET /readyz returns 200" GET "http://localhost:$SERVER_PORT/readyz" 200
assert_body_contains "GET /readyz body has status up" GET "http://localhost:$SERVER_PORT/readyz" '"status":"up"'

# Full health (includes database component)
assert_status "GET /health returns 200" GET "http://localhost:$SERVER_PORT/health" 200
assert_body_contains "GET /health includes database check" GET "http://localhost:$SERVER_PORT/health" '"database"'

# Individual component check
assert_status "GET /health/database returns 200" GET "http://localhost:$SERVER_PORT/health/database" 200

# ============================================================================
# Step 5: Basic API Assertions
# ============================================================================

log "Running API assertions..."

# Magic link — public endpoint, no auth needed
assert_status "POST /api/v1/auth/magic-link returns 200" \
    POST "http://localhost:$SERVER_PORT/api/v1/auth/magic-link" 200 \
    -H "Content-Type: application/json" \
    -d '{"email":"smoke-test@example.com"}'

# Schema list — public endpoint
assert_status "GET /api/v1/schemas returns 200" \
    GET "http://localhost:$SERVER_PORT/api/v1/schemas" 200

# ============================================================================
# Results
# ============================================================================

echo ""
echo "============================================"
echo "  Smoke Test Results"
echo "============================================"
echo "  Passed: $PASS"
echo "  Failed: $FAIL"
echo "============================================"

if [[ $FAIL -gt 0 ]]; then
    exit 1
fi

echo ""
echo "All smoke tests passed."
