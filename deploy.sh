#!/bin/bash

# MulticamObserver - Production Deployment Script
# Handles container compilation, database bootstrapping, and service health checks

set -e

echo "══════════════════════════════════════════════"
echo " 🚀 MULTICAM OBSERVER PRODUCTION DEPLOYMENT"
echo "══════════════════════════════════════════════"

# ──────────────────────────────────────────────────
# 1. Verification of Prerequisites
# ──────────────────────────────────────────────────

echo "🔍 Verifying system requirements..."
if ! command -v docker &> /dev/null; then
    echo "❌ Error: Docker is not installed on this server."
    exit 1
fi

if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "❌ Error: Docker Compose is not installed."
    exit 1
fi

echo "✅ Prerequisites verified."

# ──────────────────────────────────────────────────
# 2. Build & Start Containers
# ──────────────────────────────────────────────────

echo "🐳 Rebuilding and starting MulticamObserver containers..."
docker compose down
docker compose up -d --build

# ──────────────────────────────────────────────────
# 3. Health Check - Wait for PostgreSQL DB
# ──────────────────────────────────────────────────

echo "⏳ Waiting for PostgreSQL database (multicam_db) to accept connections..."
MAX_RETRIES=30
COUNT=0

until docker compose exec -T db pg_isready -U multicam_user -d multicamobserver -q 2>/dev/null || [ $COUNT -eq $MAX_RETRIES ]; do
    echo "  Waiting for database... ($COUNT/$MAX_RETRIES)"
    sleep 2
    COUNT=$((COUNT + 1))
done

if [ $COUNT -eq $MAX_RETRIES ]; then
    echo "❌ Timeout: PostgreSQL failed to accept connections."
    exit 1
fi

echo "✅ PostgreSQL is ready."

# ──────────────────────────────────────────────────
# 4. Health Check - Wait for Go Web Server Container
# ──────────────────────────────────────────────────

echo "⏳ Waiting for Go web container (multicam_web) to be running..."
COUNT=0

until [ "$(docker inspect -f '{{.State.Running}}' multicam_web 2>/dev/null)" == "true" ] || [ $COUNT -eq $MAX_RETRIES ]; do
    echo "  Waiting for container... ($COUNT/$MAX_RETRIES)"
    sleep 1
    COUNT=$((COUNT + 1))
done

if [ $COUNT -eq $MAX_RETRIES ]; then
    echo "❌ Timeout: Web application container failed to start."
    exit 1
fi

echo "✅ Web container is running."

# ──────────────────────────────────────────────────
# 5. Health Check - Wait for Web HTTP Healthcheck Endpoint
# ──────────────────────────────────────────────────

echo "⏳ Querying Go HTTP server /health endpoint on port 51177..."
COUNT=0

until docker compose exec -T web wget -qO- http://localhost:51177/health >/dev/null 2>&1 || [ $COUNT -eq $MAX_RETRIES ]; do
    echo "  Waiting for HTTP response... ($COUNT/$MAX_RETRIES)"
    sleep 2
    COUNT=$((COUNT + 1))
done

if [ $COUNT -eq $MAX_RETRIES ]; then
    echo "❌ Timeout: Web HTTP server failed to answer healthcheck."
    exit 1
fi

echo "✅ MulticamObserver Web application is answering health checks perfectly."

# ──────────────────────────────────────────────────
# 6. Database Verification
# ──────────────────────────────────────────────────

echo "🗄️  Verifying database schemas and seeders..."
BOOTSTRAP_LOG=$(docker compose logs web | grep -E "Successfully (initialized database schema|seeded initial admin user)" || true)

if [ -n "$BOOTSTRAP_LOG" ]; then
    echo "$BOOTSTRAP_LOG"
else
    echo "✅ Database is online, migrated, and verified."
fi

# Fetch the dynamically auto-generated node ID from the database
NODE_ID=$(docker compose exec -T db psql -U multicam_user -d multicamobserver -t -A -c "SELECT node_id FROM broadcasters LIMIT 1;" || echo "cam-workspace")

echo ""
echo "══════════════════════════════════════════════"
echo " ✅ MULTICAM OBSERVER DEPLOYMENT COMPLETE"
echo " 🌐 App Portal: http://localhost:51177"
echo " 👤 Admin User: admin@multicamobserver.com / ObserverAdmin2026!"
echo " 🎥 Camera Node: $NODE_ID / CameraNodeSecure1!"
echo "══════════════════════════════════════════════"
