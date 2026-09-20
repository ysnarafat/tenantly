#!/usr/bin/env bash
# PostToolUse hook: format + lint a single Go file after Write/Edit.
#
# Reads the tool payload as JSON on stdin, extracts tool_input.file_path, and if
# it is a .go file:
#   1. gofmt -s -w   (matches CI's `gofmt -s -l` check)
#   2. golangci-lint on the edited file's *package* (matches CI's golangci-lint
#      step: errcheck / staticcheck / unused / ineffassign, etc.)
#
# golangci-lint is scoped to the single package (not the whole module) so it
# stays fast enough to run on every edit (~1-4s). It is NOT installed on PATH
# here and Windows Application Control blocks it unless it is the freshly-built
# copy in src/backend/api/tmp/. If that binary is missing we fall back to
# `go vet` and print how to build it — `go vet` does NOT catch errcheck, so the
# fallback is a degraded check, not an equivalent one.
#
# Non-blocking: always exits 0. Findings are printed so they surface in the
# tool result; they do not fail the edit.

set -o pipefail

# --- extract file_path from the hook JSON on stdin -------------------------
file_path=$(node -e "let d='';process.stdin.on('data',c=>d+=c);process.stdin.on('end',()=>{try{process.stdout.write(JSON.parse(d).tool_input.file_path||'')}catch(e){}})")

# only act on Go files
case "$file_path" in
  *.go) ;;
  *) exit 0 ;;
esac

# normalize Windows backslashes -> forward slashes for bash path handling
fp="${file_path//\\//}"

# gofmt is fine with either slash style; run it on the original path
gofmt -s -w "$file_path"

# derive the package dir, backend module root, and package path relative to it
pkg_dir="${fp%/*}"
api_dir="${fp%%/src/backend/api/*}/src/backend/api"
pkg_rel="${pkg_dir#*/src/backend/api/}"   # e.g. internal/services ; == pkg_dir if file is at module root
lint_bin="$api_dir/tmp/golangci-lint.exe"

if [[ -x "$lint_bin" ]]; then
  # Run from the module root with the package path (as CI does). NB: running
  # `run .` from inside the package dir exits 0 even when issues are found, so
  # the failure branch would never fire — always run from api_dir.
  if ! (cd "$api_dir" && "$lint_bin" run "./$pkg_rel/" --max-issues-per-linter=0 --max-same-issues=0); then
    echo "⚠️  golangci-lint reported issues in $pkg_rel — fix before committing (CI runs the same linter)."
  fi
else
  (cd "$api_dir" && go vet ./...)
  echo "ℹ️  golangci-lint binary not found at src/backend/api/tmp/golangci-lint.exe."
  echo "    Ran 'go vet' only — errcheck/staticcheck are NOT checked, so CI may still fail."
  echo "    Build it once (survives in gitignored tmp/): from a scratch dir,"
  echo "    go mod init t && go get github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 &&"
  echo "    go build -o \"$api_dir/tmp/golangci-lint.exe\" github.com/golangci/golangci-lint/v2/cmd/golangci-lint"
fi

exit 0
