.PHONY: build run stop clean test push

# Variables
IMAGE_NAME := my-go-api
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
REGISTRY := rutvik79

# Build multi-stage image
build:
	@echo "Building $(IMAGE_NAME):$(VERSION)..."
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ") \
		--build-arg GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown") \
		-t $(IMAGE_NAME):$(VERSION) \
		-t $(IMAGE_NAME):latest \
		-f Dockerfile.production \
		.
	@echo "Build complete!"
	@docker images | grep $(IMAGE_NAME) | head -1

# Run container
run:
	docker run -d \
		-p 8080:8080 \
		--name $(IMAGE_NAME) \
		$(IMAGE_NAME):latest

# Stop and remove container
stop:
	docker stop $(IMAGE_NAME) || true
	docker rm $(IMAGE_NAME) || true

# Clean up images
clean: stop
	docker rmi $(IMAGE_NAME):latest || true
	docker rmi $(IMAGE_NAME):$(VERSION) || true

# Test the API
test:
	@echo "Testing health endpoint..."
	curl -f http://localhost:8080/health || exit 1
	@echo "\nTesting version endpoint..."
	curl -f http://localhost:8080/version || exit 1
	@echo "\nAll tests passed!"

# Push to registry
push:
	docker tag $(IMAGE_NAME):$(VERSION) $(REGISTRY)/$(IMAGE_NAME):$(VERSION)
	docker tag $(IMAGE_NAME):latest $(REGISTRY)/$(IMAGE_NAME):latest
	docker push $(REGISTRY)/$(IMAGE_NAME):$(VERSION)
	docker push $(REGISTRY)/$(IMAGE_NAME):latest

# Build and run
all: build run

# Show help
help:
	@echo "Available targets:"
	@echo "  make build   - Build Docker image"
	@echo "  make run     - Run container"
	@echo "  make stop    - Stop and remove container"
	@echo "  make clean   - Remove images"
	@echo "  make test    - Test running container"
	@echo "  make push    - Push to Docker registry"
	@echo "  make all     - Build and run"