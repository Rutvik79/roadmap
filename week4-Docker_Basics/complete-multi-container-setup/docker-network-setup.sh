#!/bin/bash

echo "Setting up Docker network environment..."

# Create network
echo "Creating network..."
docker network create myapp-network

# Create volumes
echo "Creating volumes..."
docker volume create postgres-data
docker volume create redis-data

# Start PostgreSQL
echo "Starting PostgreSQL..."
docker run -d \
  --name postgres \
  --network myapp-network \
  -e POSTGRES_PASSWORD=mysecretpassword \
  -e POSTGRES_DB=myapp \
  -v postgres-data:/var/lib/postgresql/data \
  postgres:15-alpine

# Wait for Postgres to be ready
echo "Waiting for PostgreSQL to be ready..."
sleep 5

# Start Redis
echo "Starting Redis..."
docker run -d \
  --name redis \
  --network myapp-network \
  -v redis-data:/data \
  redis:7-alpine

# Build and start API
echo "Building API..."
cd api
docker build -t myapp-api:v1 .

echo "Starting API..."
docker run -d \
  --name api \
  --network myapp-network \
  -p 8080:8080 \
  -e DB_HOST=postgres \
  -e DB_USER=postgres \
  -e DB_PASSWORD=mysecretpassword \
  -e DB_NAME=myapp \
  -e REDIS_HOST=redis \
  myapp-api:v1

echo ""
echo "Setup complete!"
echo "API: http://localhost:8080"
echo ""
echo "Test with:"
echo "  curl http://localhost:8080/health"
echo "  curl http://localhost:8080/users"
echo ""
echo "Containers running:"
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"