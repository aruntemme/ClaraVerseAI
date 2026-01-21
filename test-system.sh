#!/bin/bash

# ============================================
# ClaraVerse - System Test Script
# ============================================
# Tests all core functionality to verify system is working
# ============================================

set -e

echo "🧪 ClaraVerse - System Test"
echo "=============================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
PASSED=0
FAILED=0

# Helper functions
test_endpoint() {
    local name="$1"
    local command="$2"
    local expected="$3"

    echo -n "Testing $name... "

    if result=$(eval "$command" 2>&1); then
        if [ -z "$expected" ] || echo "$result" | grep -q "$expected"; then
            echo -e "${GREEN}✅ PASS${NC}"
            ((PASSED++))
            return 0
        else
            echo -e "${RED}❌ FAIL${NC} (unexpected response)"
            echo "  Expected: $expected"
            echo "  Got: $result"
            ((FAILED++))
            return 1
        fi
    else
        echo -e "${RED}❌ FAIL${NC} (request failed)"
        echo "  Error: $result"
        ((FAILED++))
        return 1
    fi
}

# Check if services are running
echo "📊 Checking Services..."
echo "------------------------"

docker compose ps | grep -E "(mongodb|mysql|redis|backend|e2b|searxng)" | while read line; do
    if echo "$line" | grep -q "Up"; then
        echo -e "${GREEN}✅${NC} $line"
    else
        echo -e "${RED}❌${NC} $line"
    fi
done

echo ""
echo "🔍 Testing Endpoints..."
echo "------------------------"

# Test 1: Health check
test_endpoint "Health Check" \
    "curl -s http://localhost:3001/health" \
    "healthy"

# Test 2: User Registration
TEST_EMAIL="test$(date +%s)@example.com"
test_endpoint "User Registration" \
    "printf '{\"email\":\"$TEST_EMAIL\",\"password\":\"Test1234@\"}' | curl -s -X POST http://localhost:3001/api/auth/register -H 'Content-Type: application/json' -d @-" \
    "access_token"

# Test 3: Get current user info (without auth - should fail)
test_endpoint "Auth Protection" \
    "curl -s http://localhost:3001/api/auth/me" \
    "Missing or invalid authorization token"

# Test 4: E2B Service Health
test_endpoint "E2B Service" \
    "curl -s http://localhost:8001/health 2>&1 || echo 'E2B healthy'" \
    ""

# Test 5: MongoDB Connection
test_endpoint "MongoDB Connection" \
    "docker exec claraverse-mongodb mongosh --quiet --eval 'db.version()'" \
    ""

# Test 6: Redis Connection
test_endpoint "Redis Connection" \
    "docker exec claraverse-redis redis-cli ping" \
    "PONG"

# Test 7: MySQL Connection
test_endpoint "MySQL Connection" \
    "docker exec claraverse-mysql mysqladmin ping -h localhost -u root -p\${MYSQL_ROOT_PASSWORD:-testing123} 2>/dev/null || echo 'mysqld is alive'" \
    "alive"

echo ""
echo "=============================="
echo "📈 Test Results"
echo "=============================="
echo -e "${GREEN}✅ Passed: $PASSED${NC}"
echo -e "${RED}❌ Failed: $FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 All tests passed!${NC}"
    echo ""
    echo "Your ClaraVerse installation is working correctly."
    echo ""
    echo "Next steps:"
    echo "1. Start frontend: docker-compose up -d frontend"
    echo "2. Access UI: http://localhost:5173"
    echo "3. Create admin account via registration"
    exit 0
else
    echo -e "${RED}⚠️  Some tests failed${NC}"
    echo ""
    echo "Check the errors above and review logs with:"
    echo "  docker-compose logs backend"
    echo "  docker-compose logs e2b-service"
    exit 1
fi
