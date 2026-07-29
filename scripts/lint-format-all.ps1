# One-click lint & format for the whole project (backend + frontend).
# Mirrors what CI actually checks (.github/workflows/backend-api-ci.yml,
# frontend-ci.yml) so a clean run here means CI won't fail on formatting/lint.
#
# Run from Explorer: right-click this file -> "Run with PowerShell".
# Run from a terminal: powershell -ExecutionPolicy Bypass -File scripts\lint-format-all.ps1

$RepoRoot = Split-Path -Parent $PSScriptRoot
$BackendDir = Join-Path $RepoRoot "src\backend\api"
$FrontendDir = Join-Path $RepoRoot "src\frontend"
$HadFailure = $false

Write-Host "=== Backend: gofmt -s ===" -ForegroundColor Cyan
Push-Location $BackendDir
$unformatted = gofmt -s -l .
if ($LASTEXITCODE -ne 0) {
    Write-Host "gofmt failed to run - is Go installed and on PATH?" -ForegroundColor Red
    $HadFailure = $true
} elseif ($unformatted) {
    Write-Host $unformatted
    gofmt -s -w .
    Write-Host "Formatted the files listed above"
} else {
    Write-Host "All Go files already formatted"
}

Write-Host ""
Write-Host "=== Backend: golangci-lint ===" -ForegroundColor Cyan
$golangci = Get-Command golangci-lint -ErrorAction SilentlyContinue
if ($golangci) {
    golangci-lint run --timeout=5m
    if ($LASTEXITCODE -ne 0) {
        Write-Host "golangci-lint reported issues (see above) - this will fail CI" -ForegroundColor Red
        $HadFailure = $true
    }
} else {
    Write-Host "golangci-lint not installed locally - CI runs this check on every PR." -ForegroundColor Yellow
    Write-Host "Install it to catch issues before pushing: https://golangci-lint.run/welcome/install/" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "=== Backend: go vet ===" -ForegroundColor Cyan
go vet ./...
if ($LASTEXITCODE -ne 0) {
    Write-Host "go vet found issues (see above)" -ForegroundColor Red
    $HadFailure = $true
}
Pop-Location

Write-Host ""
Write-Host "=== Frontend: Prettier ===" -ForegroundColor Cyan
Push-Location $FrontendDir
npm run format
if ($LASTEXITCODE -ne 0) {
    Write-Host "Prettier formatting failed (see above)" -ForegroundColor Red
    $HadFailure = $true
}

Write-Host ""
Write-Host "=== Frontend: ESLint (auto-fix) ===" -ForegroundColor Cyan
npm run lint:fix
if ($LASTEXITCODE -ne 0) {
    Write-Host "ESLint failed (see above). Note: CI's lint step is currently" -ForegroundColor Yellow
    Write-Host "commented out in frontend-ci.yml, so this alone won't fail CI -" -ForegroundColor Yellow
    Write-Host "but it should still be fixed. If the error mentions a missing" -ForegroundColor Yellow
    Write-Host "'@eslint/js' module, that's a pre-existing missing devDependency," -ForegroundColor Yellow
    Write-Host "not something this run caused." -ForegroundColor Yellow
}
Pop-Location

Write-Host ""
if ($HadFailure) {
    Write-Host "Lint/format finished with issues that will fail CI - see above." -ForegroundColor Red
} else {
    Write-Host "All lint & format checks complete!" -ForegroundColor Green
}

if (-not $env:LINT_FORMAT_NOPAUSE) {
    Read-Host "Press Enter to close"
}

if ($HadFailure) { exit 1 } else { exit 0 }
