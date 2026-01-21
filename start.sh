#!/bin/bash

# ============================================
# ClaraVerse - Quick Start Script
# ============================================
# Starts all ClaraVerse services in the correct order
# ============================================

set -e

echo "🚀 ClaraVerse - Starting All Services"
echo "======================================"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Step 1: Clean up old containers
echo -e "${BLUE}Step 1:${NC} Cleaning up old containers..."
docker stop claraverse-backend-prod-debug claraverse-frontend-dev 2>/dev/null || true
docker rm claraverse-backend-prod-debug claraverse-frontend-dev 2>/dev/null || true
echo -e "${GREEN}✓${NC} Cleanup complete"
echo ""

# Step 2: Start base services
echo -e "${BLUE}Step 2:${NC} Starting base services (MongoDB, MySQL, Redis, SearXNG, E2B)..."
docker compose up -d mongodb mysql redis searxng e2b-service
echo -e "${GREEN}✓${NC} Base services starting"
echo ""

# Step 3: Wait for health checks
echo -e "${BLUE}Step 3:${NC} Waiting for services to become healthy..."
echo -e "${YELLOW}This may take 30-45 seconds...${NC}"

TIMEOUT=60
ELAPSED=0
while [ $ELAPSED -lt $TIMEOUT ]; do
    # Check if all base services are healthy
    HEALTHY=$(docker compose ps --format json | jq -r 'select(.Service == "mongodb" or .Service == "mysql" or .Service == "redis" or .Service == "searxng" or .Service == "e2b-service") | select(.Health == "healthy")' | wc -l)

    if [ "$HEALTHY" -eq 5 ]; then
        echo -e "${GREEN}✓${NC} All base services are healthy!"
        break
    fi

    echo -n "."
    sleep 2
    ELAPSED=$((ELAPSED + 2))
done

if [ $ELAPSED -ge $TIMEOUT ]; then
    echo -e "\n${YELLOW}⚠️  Timeout waiting for services. Proceeding anyway...${NC}"
fi
echo ""

# Step 4: Start backend
echo -e "${BLUE}Step 4:${NC} Starting backend..."
docker compose start backend
sleep 3
echo -e "${GREEN}✓${NC} Backend started"
echo ""

# Step 5: Start frontend
echo -e "${BLUE}Step 5:${NC} Starting frontend..."
docker compose start frontend
sleep 2
echo -e "${GREEN}✓${NC} Frontend started"
echo ""

# Step 6: Verify everything is running
echo -e "${BLUE}Step 6:${NC} Verifying services..."
echo ""

docker compose ps

echo ""
echo "======================================"
echo -e "${GREEN}🎉 ClaraVerse is now running!${NC}"
echo "======================================"
echo ""
echo "Access points:"
echo -e "  ${BLUE}Backend API:${NC}  http://localhost:3001"
echo -e "  ${BLUE}Health Check:${NC} http://localhost:3001/health"
echo -e "  ${BLUE}Frontend UI:${NC}  http://localhost:80"
echo ""
echo "Next steps:"
echo "  1. Create an account: See README.md for registration"
echo "  2. View logs: docker compose logs -f"
echo "  3. Stop all: docker compose down"
echo ""
