#!/bin/bash

# Get version info
VERSION=${1:-dev}
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo "Building Docker image..."
echo "Version: $VERSION"
echo "Build Date: $BUILD_DATE"
echo "Git Commit: $GIT_COMMIT"

# Build Image
docker build \
    -f Dockerfile.production \
    --build-arg VERSION=$VERSION \
    --build-arg BUILD_DATE=$BUILD_DATE \
    --build-arg GIT_COMMIT=$GIT_COMMIT \
    -t my-go-api:$VERSION \
    -t my-go-api:latest \
    .

echo "Build Complete"
echo "Image: my-go-api:$VERSION"

# Show image size
docker images | grep my-go-api | head -1