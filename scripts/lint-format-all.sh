#!/bin/bash

# One-click lint & format for the whole project (backend + frontend).
# Mirrors what CI actually checks (.github/workflows/backend-api-ci.yml,
# frontend-ci.yml) so a clean run here means CI won't fail on formatting/lint.

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
HAD_FAILURE=0

echo "=== Backend: gofmt -s ==="
cd "$REPO_ROOT/src/backend/api"
UNFORMATTED=$(gofmt -s -l .)
if [ -n "$UNFORMATTED" ]; then
    echo "$UNFORMATTED"
    gofmt -s -w .
    echo "Formatted the files listed above"
else
    echo "All Go files already formatted"
fi

echo ""
echo "=== Backend: golangci-lint ==="
if command -v golangci-lint >/dev/null 2>&1; then
    golangci-lint run --timeout=5m
    if [ $? -ne 0 ]; then
        echo "golangci-lint reported issues (see above) - this will fail CI"
        HAD_FAILURE=1
    fi
else
    echo "golangci-lint not installed locally - CI runs this check on every PR."
    echo "Install it to catch issues before pushing: https://golangci-lint.run/welcome/install/"
fi

echo ""
echo "=== Backend: go vet ==="
go vet ./...
if [ $? -ne 0 ]; then
    echo "go vet found issues (see above)"
    HAD_FAILURE=1
fi

echo ""
echo "=== Frontend: Prettier ==="
cd "$REPO_ROOT/src/frontend"
npm run format
if [ $? -ne 0 ]; then
    echo "Prettier formatting failed (see above)"
    HAD_FAILURE=1
fi

echo ""
echo "=== Frontend: ESLint (auto-fix) ==="
npm run lint:fix
if [ $? -ne 0 ]; then
    echo "ESLint failed (see above). Note: CI's lint step is currently commented"
    echo "out in frontend-ci.yml, so this alone won't fail CI - but it should"
    echo "still be fixed. If the error mentions a missing '@eslint/js' module,"
    echo "that's a pre-existing missing devDependency, not something this run caused."
fi

echo ""
if [ $HAD_FAILURE -ne 0 ]; then
    echo "Lint/format finished with issues that will fail CI - see above."
    exit 1
else
    echo "All lint & format checks complete!"
    exit 0
fi
