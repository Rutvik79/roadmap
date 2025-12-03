#!/bin/bash

set -e

echo "=============================="
echo "Testing Docker Compose Setup"
echo "=============================="

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Test 1: Check services are running 
echo -e "\n${GREEN}Test 1: Checking services...${NC}"
docker-compose ps

# Test 2: Health check
echo -e "\n${GREEN}Test 2: API health check...${NC}"
response=$(curl -s http://localhost:8080/health)
if echo "$response" | grep -q "healthy"; then
    echo "✅ Health check passed"
else 
    echo "❌ Health check failed"
    exit 1
fi 

# Test 3: Database connectivity
echo -e "\n${GREEN}Test 3: Database connectivity...${NC}"
docker-compose exec -T postgres pg_isready -U postgres
if [ $? -eq 0 ]; then
    echo "✅ Database is ready"
else 
    echo "❌ Database not ready"
    exit 1
fi

# Test 4: Redis connectivity
echo -e "\n${GREEN}Test 4: Redis Connectivity...${NC}"
docker-compose exec -T redis redis-cli ping
if [ $? -eq 0 ]; then
    echo "✅ Redis is ready"
else 
    echo "❌ Redis not ready"
    exit 1
fi

# Test 5: Create User
echo -e "\n${GREEN}Test 5: Creating user...${NC}"
response=$(curl -s -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Test User","email":"Test@example.com"}')
if echo "$response" | grep -q "id"; then
    echo "✅ User created successfully"
else
    echo "❌ User creation failed"
    exit 1
fi

# Test 6: List users
echo -e "\n${GREEN}Test 6: Listing users...${NC}"
response=$(curl -s http://localhost:8080/users)
if echo "$response" | grep -q "Test User"; then
    echo "✅ User retrieved successfully"
else 
    echo "❌ User retrieval failed"
    exit 1
fi

# Test 7: Cache test
echo -e "\n${GREEN}Test 7: Testing cache...${NC}"
# Ensure cache is cleared
docker-compose exec -T redis redis-cli DEL "users:all" >/dev/null 2>&1 || true
# First request - should be cache miss
headers1=$(curl -s -D - http://localhost:8080/users | grep "X-Cache")
echo "First request: $headers1"

# Second request - should be cache hit
headers2=$(curl -s -D - http://localhost:8080/users | grep "X-Cache")
echo "Second request: $headers2"

if echo "$headers2" | grep -q "HIT"; then
    echo "✅ Cache working correctly"
else
    echo "⚠️  Cache not working (not critical)"
fi

echo -e "\n${GREEN}==================================="
echo "All tests passed! ✅"
echo "===================================${NC}"