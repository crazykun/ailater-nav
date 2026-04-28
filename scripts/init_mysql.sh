#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

log() {
  printf '[init-mysql] %s\n' "$1"
}

fail() {
  printf '[init-mysql] ERROR: %s\n' "$1" >&2
  exit 1
}

require_file() {
  local path="$1"
  [[ -f "$path" ]] || fail "required file not found: $path"
}

require_command() {
  local cmd="$1"
  command -v "$cmd" >/dev/null 2>&1 || fail "required command not found: $cmd"
}

cd "$REPO_ROOT"

log "repository root: $REPO_ROOT"
require_command go
require_file "$REPO_ROOT/config.yaml"
require_file "$REPO_ROOT/data/ai.json"

log "creating database if needed"
go run ./scripts/create_db/main.go

log "running schema migrations and importing legacy JSON data"
go run ./scripts/migrate/main.go

log "migration finished"
log "next step: start the service with 'go run ./main.go' or your process manager"
