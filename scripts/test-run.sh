#!/usr/bin/env bash
# 本地/CI 全量测试：默认本机 MySQL/Redis（3306/6379），不启 docker-compose。
# 可选隔离库：USE_DOCKER=1 bash scripts/test-run.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

USE_DOCKER="${USE_DOCKER:-0}"
export CI_MYSQL_HOST="${CI_MYSQL_HOST:-127.0.0.1}"
export CI_JWT_SECRET="${CI_JWT_SECRET:-ci-integration-test-secret}"
export DEV_LOGIN_BASE="${DEV_LOGIN_BASE:-http://127.0.0.1:8000}"
export TEST_BASE="${TEST_BASE:-$DEV_LOGIN_BASE}"

if [[ "$USE_DOCKER" == "1" ]]; then
  export CI_MYSQL_PORT="${CI_MYSQL_PORT:-3307}"
  export CI_MYSQL_USER="${CI_MYSQL_USER:-root}"
  export CI_MYSQL_PASSWORD="${CI_MYSQL_PASSWORD:-testpass}"
  export CI_MYSQL_DATABASE="${CI_MYSQL_DATABASE:-x_my_blog}"
  export CI_REDIS_ADDR="${CI_REDIS_ADDR:-127.0.0.1:6380}"
  export CI_REDIS_DB="${CI_REDIS_DB:-2}"
else
  export CI_MYSQL_PORT="${CI_MYSQL_PORT:-3306}"
  export CI_MYSQL_USER="${CI_MYSQL_USER:-root}"
  export CI_MYSQL_PASSWORD="${CI_MYSQL_PASSWORD:-testpass}"
  export CI_MYSQL_DATABASE="${CI_MYSQL_DATABASE:-x_my_blog}"
  export CI_REDIS_ADDR="${CI_REDIS_ADDR:-127.0.0.1:6379}"
  export CI_REDIS_DB="${CI_REDIS_DB:-1}"
  if [[ -z "${CI_USE_LOCAL_CONFIG:-}" && "${GITHUB_ACTIONS:-}" != "true" ]]; then
    export CI_USE_LOCAL_CONFIG=1
  fi
fi

cleanup() {
  bash scripts/ci/stop-services.sh || true
  if [[ "$USE_DOCKER" == "1" ]]; then
    docker compose -f deploy/docker/docker-compose.test.yml down -v 2>/dev/null || true
  fi
}
trap cleanup EXIT

if [[ "$USE_DOCKER" == "1" ]]; then
  echo ">>> docker compose test infra (3307/6380)"
  docker compose -f deploy/docker/docker-compose.test.yml up -d --wait
  bash scripts/ci/wait-for.sh
else
  echo ">>> local mysql/redis (3306/6379)"
  sleep 3
fi

if [[ "${CI_USE_LOCAL_CONFIG:-0}" == "1" ]]; then
  echo ">>> keep local configs/*.yaml"
else
  go run ./scripts/ci/env.go ./scripts/ci/prepare_config.go
fi
go run ./scripts/ci/env.go ./scripts/ci/migrate_schemas.go
go run ./scripts/ci/env.go ./scripts/ci/seed_test_data.go
bash scripts/ci/start-services.sh
bash scripts/ci/run-layer-tests.sh

echo ">>> test-run.sh OK"
