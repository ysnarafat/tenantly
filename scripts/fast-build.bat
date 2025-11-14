@echo off
REM Fast Docker Build Script for CI/CD (Windows)
REM This script optimizes Docker builds for speed and caching

echo 🚀 Starting Optimized Docker Build...

REM Enable Docker BuildKit for better performance
set DOCKER_BUILDKIT=1
set COMPOSE_DOCKER_CLI_BUILD=1

REM Check if we're in CI environment
if "%CI%"=="true" (
    echo 📦 CI Environment detected - enabling additional optimizations
    
    REM Pull latest images for cache
    echo ⬇️ Pulling latest images for cache...
    docker pull golang:1.25-alpine 2>nul || echo "Could not pull golang image"
    docker pull node:25-alpine 2>nul || echo "Could not pull node image"
    docker pull nginx:alpine 2>nul || echo "Could not pull nginx image"
    docker pull mcr.microsoft.com/dotnet/sdk:8.0 2>nul || echo "Could not pull dotnet sdk image"
    docker pull mcr.microsoft.com/dotnet/aspnet:8.0 2>nul || echo "Could not pull dotnet aspnet image"
    docker pull postgres:18.0-alpine 2>nul || echo "Could not pull postgres image"
    
    REM Try to pull previous builds for cache
    docker pull tenantly-api:latest 2>nul || echo "No previous api image found"
    docker pull tenantly-frontend:latest 2>nul || echo "No previous frontend image found"
    docker pull tenantly-notification-service:latest 2>nul || echo "No previous notification service image found"
)

REM Build services with caching
echo 🔨 Building services with optimized caching...

REM Determine which compose file to use
set COMPOSE_FILE=docker-compose.yml
if "%ENVIRONMENT%"=="production" (
    set COMPOSE_FILE=docker-compose.yml -f docker-compose.prod.yml
) else if "%ENVIRONMENT%"=="development" (
    set COMPOSE_FILE=docker-compose.yml -f docker-compose.dev.yml
)

docker-compose -f %COMPOSE_FILE% build --parallel --compress --force-rm --pull --build-arg BUILDKIT_INLINE_CACHE=1

if %errorlevel% neq 0 (
    echo ❌ Build failed
    pause
    exit /b 1
)

echo ✅ Build completed successfully!

REM Optional: Run quick health checks
if not "%SKIP_HEALTH_CHECK%"=="true" (
    echo 🔍 Running health checks...
    
    REM Start services for health check
    docker-compose -f %COMPOSE_FILE% up -d
    
    REM Wait for services to be ready
    timeout /t 10 /nobreak > nul
    
    REM Check if services are responding
    docker-compose -f %COMPOSE_FILE% ps | findstr "Up" > nul
    if %errorlevel% equ 0 (
        echo ✅ Health check passed
    ) else (
        echo ❌ Health check failed
        docker-compose -f %COMPOSE_FILE% logs
        docker-compose -f %COMPOSE_FILE% down
        pause
        exit /b 1
    )
    
    REM Clean up
    docker-compose -f %COMPOSE_FILE% down
)

REM Tag images for caching in next build
if "%CI%"=="true" (
    echo 🏷️ Tagging images for cache...
    docker tag tenantly-api:latest tenantly-api:cache 2>nul || echo "Could not tag api image"
    docker tag tenantly-frontend:latest tenantly-frontend:cache 2>nul || echo "Could not tag frontend image"
    docker tag tenantly-notification-service:latest tenantly-notification-service:cache 2>nul || echo "Could not tag notification service image"
)

echo 🎉 Fast build completed!

REM Display build summary
echo.
echo 📊 Build Summary:
docker images | findstr tenantly

pause