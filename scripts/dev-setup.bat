@echo off
echo Setting up Tenantly development environment...

REM Create .env file from example if it doesn't exist
if not exist .env (
    copy .env.example .env
    echo Created .env file from .env.example
    echo Please update the .env file with your actual configuration values
)

REM Start PostgreSQL with Docker
echo Starting PostgreSQL database...
docker run -d --name tenantly-postgres -e POSTGRES_DB=tenantly -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=password -p 5432:5432 postgres:15-alpine

REM Wait for PostgreSQL to be ready
echo Waiting for PostgreSQL to be ready...
timeout /t 10 /nobreak > nul

REM Run database migrations
echo Running database migrations...
cd backend\api && go run cmd\server\main.go migrate || echo Migration will be implemented in later tasks
cd ..\..

echo Development environment setup complete!
echo.
echo To start the services:
echo 1. Go API: cd backend\api ^&^& go run cmd\server\main.go
echo 2. Angular Frontend: cd frontend ^&^& npm install ^&^& npm start
echo 3. C# Notification Service: cd backend\notification-service ^&^& dotnet run
echo.
echo Or use Docker Compose: docker-compose up

pause