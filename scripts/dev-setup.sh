#!/bin/bash
set -e

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
API_DIR="$ROOT/src/backend/api"
FRONTEND_DIR="$ROOT/src/frontend"

if [ ! -f "$API_DIR/go.mod" ]; then
    echo "Backend API was not found under \"$API_DIR\"." >&2
    exit 1
fi

if [ ! -f "$FRONTEND_DIR/package.json" ]; then
    echo "Frontend was not found under \"$FRONTEND_DIR\"." >&2
    exit 1
fi

if [ ! -f "$ROOT/.env" ] && [ -f "$ROOT/.env.example" ]; then
    cp "$ROOT/.env.example" "$ROOT/.env"
    echo "Created .env from .env.example."
fi

LOG_DIR="$(mktemp -d "${TMPDIR:-/tmp}/tenantly-dev.XXXXXX")"

nohup sh -c 'cd "$1" && exec go run ./cmd/server' sh "$API_DIR" \
    >"$LOG_DIR/api.log" 2>&1 < /dev/null &
API_PID=$!

nohup sh -c 'cd "$1" && exec npm start' sh "$FRONTEND_DIR" \
    >"$LOG_DIR/frontend.log" 2>&1 < /dev/null &
FRONTEND_PID=$!

nohup sh -c 'cd "$1" && exec npm run start:operations' sh "$FRONTEND_DIR" \
    >"$LOG_DIR/operations.log" 2>&1 < /dev/null &
OPERATIONS_PID=$!

echo "Started Tenantly API (PID $API_PID), frontend (PID $FRONTEND_PID), and operations-portal (PID $OPERATIONS_PID)."
echo "PostgreSQL must already be running for the API to connect."
echo "API:              http://localhost:8080"
echo "Frontend:         http://localhost:4200"
echo "Operations Portal: http://localhost:4300"
echo "Logs:             $LOG_DIR"
