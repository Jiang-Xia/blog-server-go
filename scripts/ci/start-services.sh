#!/usr/bin/env bash
# 后台启动四微服务（CI / test-run.sh）
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"
PIDFILE="${ROOT}/.ci-services.pids"
LOGDIR="${ROOT}/.ci-logs"
mkdir -p "$LOGDIR"
: >"$PIDFILE"

start_one() {
  local name="$1" cfg="$2" main="$3" port="$4"
  echo "start $name :$port"
  CONFIG_PATH="$cfg" go run "$main" >"$LOGDIR/${name}.log" 2>"$LOGDIR/${name}.err.log" &
  echo "$name=$!" >>"$PIDFILE"
}

start_one user configs/user.yaml ./services/user/cmd/main.go 5002
sleep 2
start_one blog configs/blog.yaml ./services/blog/cmd/main.go 5001
start_one rpg  configs/rpg.yaml  ./services/rpg/cmd/main.go  5003
sleep 2
start_one gateway configs/gateway.yaml ./services/gateway/cmd/main.go 8000

wait_health() {
  local port="$1" name="$2"
  for i in $(seq 1 90); do
    if curl -sf "http://127.0.0.1:${port}/api/v1/health" >/dev/null; then
      echo "$name ready"
      return 0
    fi
    sleep 1
  done
  echo "$name health timeout, see $LOGDIR/${name}.err.log" >&2
  return 1
}

wait_health 5002 user
wait_health 5001 blog
wait_health 5003 rpg
wait_health 8000 gateway
echo "all services up"
