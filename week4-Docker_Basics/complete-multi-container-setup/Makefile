.PHONY: build up down restart logs test clean help

# Variables
NETWORK_NAME := myapp-network 
API_IMAGE := myapp-api:v1
API_CONTAINER := api
DB_CONTAINER := postgres
REDIS_CONTAINER := redis
POSTGRES_VOLUME := postgres-data
REDIS_VOLUME := redis-data

# Build API IMAGE
build: 
	@echo "Building API image..."
	cd api && docker build -t $(API_IMAGE) .

# create network volumes
setup:
	@echo "Creating network..."
	docker network create $(NETWORK_NAME) 2>/dev/null || true
	@echo "Creating volumes..."
	docker volume create $(POSTGRES_VOLUME) 2>/dev/null || true
	docker volume create $(REDIS_VOLUME) 2>/dev/null || true

# Start all containers
up: setup
	@echo "Starting PostgreSQl..."
	docker run -d \
	--name $(DB_CONTAINER) \
	--network $(NETWORK_NAME) \
	-e POSTGRES_PASSWORD=mysecretpassword \
	-e POSTGRES_DB=myapp \
	-v $(POSTGRES_VOLUME):/var/lib/postgresql/data \
	postgres:15-alpine
	@echo "Waiting for PostgreSQL..."
	@sleep 5
	@echo "Starting Redis..."
	docker run -d \
	--name $(REDIS_CONTAINER) \
	--network $(NETWORK_NAME) \
	-v $(REDIS_VOLUME):/data \
	redis:7-alpine
	@echo "Starting API..."
	docker run -d \
	--name $(API_CONTAINER) \
	--network $(NETWORK_NAME) \
	-p 8080:8080 \
	-e DB_HOST=$(DB_CONTAINER) \
	-e DB_USER=postgres \
	-e DB_PASSWORD=mysecretpassword \
	-e DB_NAME=myapp \
	-e REDIS_HOST=$(REDIS_CONTAINER) \
	$(API_IMAGE)
	@echo "\nAll services started!"
	@echo "API: http://localhost:8080"

# Stop all Containers
down:
	@echo "Stopping containers..."
	docker stop $(API_CONTAINER) $(DB_CONTAINER) $(REDIS_CONTAINER) 2>/dev/null || true
	docker rm $(API_CONTAINER) $(DB_CONTAINER) $(REDIS_CONTAINER) 2>/dev/null || true


# Restart all containers
restart: down up

# View logs
logs: 
	@echo "=== API logs ==="
	docker logs --tail 50 $(API_CONTAINER)
	@echo "\n===PostgreSQL Logs ==="
	docker logs --tail 20 $(DB_CONTAINER)
	@echo "\n=== Redis Logs ==="
	docker logs --tail 20 $(REDIS_CONTAINER)

# Follow logs
logs-follow:
	docker logs -f $(API_CONTAINER)

# Run tests
test:
	@echo "Testing API..."
	curl -f http://localhost:8080/health || exit 1
	@echo "\nCreating test user..."
	curl -X POST http://localhost:8080/users \
	-H "Content-Type: application/json" \
	-d '{"name":"Test","email":"test@example.com"}'
	@echo "\nListing users..."
	curl http://localhost:8080/users
	@echo "\nTest passed!"

# Clean everything (including volumes)
clean: down
	@echo "Removing network..."
	docker network rm $(NETWORK_NAME) 2>/dev/null || true
	@echo "Removing volumes..."
	docker volume rm $(POSTGRES_VOLUME) 2>/dev/null || true
	@echo "Cleanup complete!"

# Help
help:
	@echo "Available Targets:"
	@echo "	make build 			- Build API image"
	@echo "	make setup 			- Create network and volumes"
	@echo "	make up 			- Start all services"
	@echo " make down 			- Stop all services"
	@echo " make restart 		- Restart all services"
	@echo " make logs			- View logs"
	@echo " make logs-follow 	- Follow API logs"
	@echo " make test 			- Run tests"
	@echo " make clean 			- Remove everything"