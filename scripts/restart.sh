#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
BINARY_NAME="ai-later-nav"
PID_FILE="${REPO_ROOT}/${BINARY_NAME}.pid"

log() {
  printf '[restart] %s\n' "$1"
}

fail() {
  printf '[restart] ERROR: %s\n' "$1" >&2
  exit 1
}

cd "$REPO_ROOT"

# Stop existing process
stop_existing() {
  local pid=""
  
  # Try reading PID file first
  if [[ -f "$PID_FILE" ]]; then
    pid=$(cat "$PID_FILE" 2>/dev/null || true)
  fi
  
  # Fallback to finding process by name
  if [[ -z "$pid" ]] || ! kill -0 "$pid" 2>/dev/null; then
    pid=$(pgrep -x "$BINARY_NAME" 2>/dev/null | head -1 || true)
  fi
  
  if [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null; then
    log "stopping existing process (PID: $pid)"
    kill "$pid" 2>/dev/null || true
    
    # Wait up to 5 seconds for graceful shutdown
    local waited=0
    while kill -0 "$pid" 2>/dev/null && [[ $waited -lt 5 ]]; do
      sleep 1
      ((waited++))
    done
    
    # Force kill if still running
    if kill -0 "$pid" 2>/dev/null; then
      log "force killing process $pid"
      kill -9 "$pid" 2>/dev/null || true
    fi
    
    log "process stopped"
  else
    log "no existing process found"
  fi
  
  # Clean up PID file
  rm -f "$PID_FILE"
}

# Build new binary
build() {
  log "building ${BINARY_NAME}..."
  go build -o "$BINARY_NAME" .
  log "build complete"
}

# Start new process
start() {
  log "starting ${BINARY_NAME}..."
  nohup "./${BINARY_NAME}" > "${BINARY_NAME}.log" 2>&1 &
  local pid=$!
  echo "$pid" > "$PID_FILE"
  log "started with PID: $pid"
  log "log file: ${REPO_ROOT}/${BINARY_NAME}.log"
}

# Main
log "repository root: $REPO_ROOT"

stop_existing
build
start

log "restart complete"
