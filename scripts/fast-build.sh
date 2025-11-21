#!/bin/bash

# Fast Docker Build Script for CI/CD
# This script optimizes Docker builds for speed and caching

set -e

echo "🚀 Starting Optimized Docker Build..."

# Enable Docker BuildKit for better performance
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1

# Build arguments for optimization
BUILD_ARGS="--build-arg BUILDKIT_INLINE_CACHE=1"

# Check if we're in CI environment
if [ "$CI" = "true" ]; then
    echo "📦 CI Environment detected - enabling additional optimizations"
    
    # Pull latest images for cache
    echo "⬇️ Pulling latest images for cache..."
    docker pull golang:1.25-alpine || true
    docker pull node:25-alpine || true
    docker pull nginx:alpine || true
    docker pull mcr.microsoft.com/dotnet/sdk:8.0 || true
    docker pull mcr.microsoft.com/dotnet/aspnet:8.0 || true
    docker pull postgres:18.0-alpine || true
    
    # Try to pull previous builds for cache
    docker pull tenantly-api:latest || true
    docker pull tenantly-frontend:latest || true
    docker pull tenantly-notification-service:latest || true
fi

# Build services in parallel with caching
echo "🔨 Building services with optimized caching..."

# Determine which compose file to use
COMPOSE_FILE="docker-compose.yml"
if [ "$ENVIRONMENT" = "production" ]; then
    COMPOSE_FILE="docker-compose.yml -f docker-compose.prod.yml"
elif [ "$ENVIRONMENT" = "development" ]; then
    COMPOSE_FILE="docker-compose.yml -f docker-compose.dev.yml"
fi

# Build all services in parallel
docker-compose -f $COMPOSE_FILE build \
    --parallel \
    --compress \
    --force-rm \
    --pull \
    $BUILD_ARGS

echo "✅ Build completed successfully!"

# Optional: Run quick health checks
if [ "$SKIP_HEALTH_CHECK" != "true" ]; then
    echo "🔍 Running health checks..."
    
    # Start services for health check
    docker-compose -f $COMPOSE_FILE up -d
    
    # Wait for services to be ready
    sleep 10
    
    # Check if services are responding
    if docker-compose -f $COMPOSE_FILE ps | grep -q "Up"; then
        echo "✅ Health check passed"
    else
        echo "❌ Health check failed"
        docker-compose -f $COMPOSE_FILE logs
        exit 1
    fi
    
    # Clean up
    docker-compose -f $COMPOSE_FILE down
fi

# Tag images for caching in next build
if [ "$CI" = "true" ]; then
    echo "🏷️ Tagging images for cache..."
    docker tag tenantly-api:latest tenantly-api:cache || true
    docker tag tenantly-frontend:latest tenantly-frontend:cache || true
    docker tag tenantly-notification-service:latest tenantly-notification-service:cache || true
fi

echo "🎉 Fast build completed in $(date)"

# Display build summary
echo ""
echo "📊 Build Summary:"
docker images | grep tenantly | head -10