.PHONY: help build up down restart logs ps test clean dev prod

# Default Target
.DEFAULT_GOAL := help

# Load environment variables
include .env
export

# Help target
help:
	@echo "Avaiable targets:"
	@echo "	make build		- Build all services"
	@echo "	make up			- Start all services"
	@echo "	make down		- Stop all services"
	@echo "	make restart	- Restart all services"
	@echo "	make logs		- View logs"
	@echo "	make ps			- List running services"
	@echo "	make test		- Run API tests"
	@echo "	make clean		- Remove all conatiners and volumes"
	@echo "	make dev		- Start in developement mode"
	@echo "	make prod		- Start in production mode"

# Build all services
build:
	docker-compose build

# Start all services
up: 
	docker-compose up -d
	@echo "Waiting for services to be healthy..."
	@sleep 5
	@docker-compose ps

# Ṣtop all services
down:
	docker-compose down

# Restart all services
restart: down up

# View logs
logs:
	docker-compose logs -f

# Show service status
ps:
	docker-compose ps

# Run tests
test:
	./test-compose.sh
# 	@echo "Testing health endpoint..."
# 	@curl -f http://localhost:8080/health || exit 1
# 	@echo "\nCreating test user..."
# 	@curl -X POST http://localhost:8080/users \
# 		-H "Content-Type: application/json"
# 		-d '{"name": "Test", "email": "test@example.com"}'
# 	@echo "\nListing users..."
# 	@curl http://localhost:8080/users
# 	@echo "\n✅ All test passed!"

# Clean everything
clean:
	docker-compose down -v
	docker system prune -f

# Develpoment mode
dev:
	docker-compose -f docker-compose.yml -f docker-compose.dev.yml up

# Production mode
prod: 
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# Database shell
db-shell:
	docker-compose exec postgres psql -U postgres -d myapp

# Redis CLI
redis-cli:
	docker-compose exec redis redis-cli

# API shell
api-shell:
	docker-compose exec api sh 

# View API logs only
api-logs:
	docker-compose logs -f api 

# Rebuild and restart specific service
rebuild-api:
	docker-compose up -d --build api