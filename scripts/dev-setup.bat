@echo off
setlocal

for %%I in ("%~dp0..") do set "ROOT=%%~fI"

if not exist "%ROOT%\src\backend\api\go.mod" (
    echo Backend API was not found under "%ROOT%\src\backend\api".
    pause
    exit /b 1
)

if not exist "%ROOT%\src\frontend\package.json" (
    echo Frontend was not found under "%ROOT%\src\frontend".
    pause
    exit /b 1
)

if not exist "%ROOT%\.env" if exist "%ROOT%\.env.example" (
    copy "%ROOT%\.env.example" "%ROOT%\.env" > nul
    echo Created .env from .env.example.
)

echo Starting Tenantly API, frontend, and operations-portal...
echo PostgreSQL must already be running for the API to connect.

start "Tenantly API" /D "%ROOT%\src\backend\api" cmd /k "go run ./cmd/server"
start "Tenantly Frontend" /D "%ROOT%\src\frontend" cmd /k "npm start"
start "Tenantly Operations Portal" /D "%ROOT%\src\frontend" cmd /k "npm run start:operations"
