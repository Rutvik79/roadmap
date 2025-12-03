#!/bin/bash

echo "Cleaning up Docker environment..."

# Stop containers
echo "Stopping containers..."
docker stop api postgres redis 2>/dev/null

# Remove containers
echo "Remove containers..."
docker rm api postgres redis 2>/dev/null

# Remove network
echo "Removing network..."
docker network rm myapp-network 2>/dev/null

# Optional: Remove volumes (data will be lost!)
read -p "Remove volumes? (data will be deleted) [y/N]: " -n 1 -r
echo 
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Removing volumes..."
    docker volume rm postgres-data redis-data 2>/dev/null
else
    echo "Keeping volumes (data preserved)"
fi

# Remove Image
echo "Removing Image..."
docker rmi myapp-api:v1

echo "Cleanup complete!"