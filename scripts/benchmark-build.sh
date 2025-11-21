#!/bin/bash

# Docker Build Benchmark Script
# Measures build performance before and after optimizations

set -e

echo "🏁 Docker Build Performance Benchmark"
echo "======================================"

# Function to measure build time
measure_build() {
    local build_type=$1
    local compose_file=$2
    
    echo "📊 Testing: $build_type"
    echo "Using: $compose_file"
    
    # Clean up previous builds
    docker-compose -f $compose_file down --rmi all --volumes --remove-orphans 2>/dev/null || true
    docker system prune -f >/dev/null 2>&1
    
    # Measure build time
    start_time=$(date +%s)
    
    if [ "$build_type" = "Optimized" ]; then
        export DOCKER_BUILDKIT=1
        export COMPOSE_DOCKER_CLI_BUILD=1
        docker-compose -f $compose_file build --parallel --compress --force-rm
    else
        unset DOCKER_BUILDKIT
        unset COMPOSE_DOCKER_CLI_BUILD
        docker-compose -f $compose_file build
    fi
    
    end_time=$(date +%s)
    build_time=$((end_time - start_time))
    
    echo "⏱️  Build completed in: ${build_time} seconds"
    echo ""
    
    return $build_time
}

# Test standard build
echo "🐌 Testing Standard Build (No Optimizations)..."
measure_build "Standard" "docker-compose.yml"
standard_time=$?

# Test optimized build
echo "🚀 Testing Optimized Build..."
measure_build "Optimized" "docker-compose.yml"
optimized_time=$?

# Calculate improvement
if [ $standard_time -gt 0 ]; then
    improvement=$(( (standard_time - optimized_time) * 100 / standard_time ))
    time_saved=$((standard_time - optimized_time))
    
    echo "📈 Performance Results:"
    echo "======================"
    echo "Standard Build:  ${standard_time} seconds"
    echo "Optimized Build: ${optimized_time} seconds"
    echo "Time Saved:      ${time_saved} seconds"
    echo "Improvement:     ${improvement}%"
    
    if [ $improvement -gt 50 ]; then
        echo "🎉 Excellent optimization! Over 50% improvement!"
    elif [ $improvement -gt 25 ]; then
        echo "✅ Good optimization! 25-50% improvement!"
    elif [ $improvement -gt 0 ]; then
        echo "👍 Some improvement, but could be better."
    else
        echo "⚠️  No improvement detected. Check your optimizations."
    fi
else
    echo "❌ Could not measure standard build time"
fi

echo ""
echo "💡 Tips for further optimization:"
echo "- Use Docker registry caching in CI/CD"
echo "- Implement multi-stage builds"
echo "- Optimize .dockerignore files"
echo "- Use BuildKit features"
echo "- Enable parallel builds"