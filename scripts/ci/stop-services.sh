#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
PIDFILE="${ROOT}/.ci-services.pids"

if [[ -f "$PIDFILE" ]]; then
  while IFS='=' read -r name pid; do
    [[ -z "${pid:-}" ]] && continue
    if kill -0 "$pid" 2>/dev/null; then
      kill "$pid" 2>/dev/null || true
      echo "stopped $name ($pid)"
    fi
  done <"$PIDFILE"
  rm -f "$PIDFILE"
fi

# 兜底：按端口清理（勿用于生产）
for port in 8000 5001 5002 5003; do
  if command -v fuser >/dev/null 2>&1; then
    fuser -k "${port}/tcp" 2>/dev/null || true
  fi
done
