#!/bin/bash

echo "Testing container connectivity..."

# Test 1: API health
echo -e "\n1. Testing API health..."
curl -f http://localhost:8080/health || echo "FAILED"

# Test 2: Database connection (via API)
echo -e "\n2. Testing database connection..."
curl -f http://localhost:8080/users || echo "FAILED"

# Test 3: create user
echo -e "\n3. Creating test user..."
curl -X POST http://localhost:8080/users \
    -H "Content-Type: application/json" \
    -d '{"name": "Test User", "email": "test@example.com"}' || echo "FAILED"

# Test 4: List users
echo -e "\n4. Listing users..."
curl -f http://localhost:8080/users || echo "FAILED"

# Test 5: Check Redis connection from API container
echo -e "\n5. Testing Redis connectivity..."
docker exec api ping -c 1 redis > /dev/null 2>&1 && echo "Redis reachable" || echo "Redis NOT reachable"

# Test 6: Check Postgres connection from API container
echo -e "\n6. Testing Postgres connectivity..."
docker exec api ping -c 1 postgres > /dev/null 2>&1 && echo "Postgres reachable" || echo "Postgres NOT reachable"

echo -e "\nAll tests complete!"