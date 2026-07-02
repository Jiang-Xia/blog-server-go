#!/usr/bin/env bash
# 行业规范本地/CI 全量测试流水线：基础设施 → 配置 → migrate → seed → 启服务 → 四层测试 → 清理
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

USE_DOCKER="${USE_DOCKER:-1}"
export CI_MYSQL_HOST="${CI_MYSQL_HOST:-127.0.0.1}"
export CI_MYSQL_PORT="${CI_MYSQL_PORT:-3307}"
export CI_MYSQL_USER="${CI_MYSQL_USER:-root}"
export CI_MYSQL_PASSWORD="${CI_MYSQL_PASSWORD:-testpass}"
export CI_MYSQL_DATABASE="${CI_MYSQL_DATABASE:-x_my_blog}"
export CI_REDIS_ADDR="${CI_REDIS_ADDR:-127.0.0.1:6380}"
export CI_REDIS_DB="${CI_REDIS_DB:-2}"
export CI_JWT_SECRET="${CI_JWT_SECRET:-ci-integration-test-secret}"
export DEV_LOGIN_BASE="${DEV_LOGIN_BASE:-http://127.0.0.1:8000}"
export TEST_BASE="${TEST_BASE:-$DEV_LOGIN_BASE}"

cleanup() {
  bash scripts/ci/stop-services.sh || true
  if [[ "$USE_DOCKER" == "1" ]]; then
    docker compose -f deploy/docker/docker-compose.test.yml down -v 2>/dev/null || true
  fi
}
trap cleanup EXIT

if [[ "$USE_DOCKER" == "1" ]]; then
  echo ">>> docker compose test infra"
  docker compose -f deploy/docker/docker-compose.test.yml up -d --wait
fi

bash scripts/ci/wait-for.sh
go run ./scripts/ci/env.go ./scripts/ci/prepare_config.go
go run ./scripts/ci/env.go ./scripts/ci/migrate_schemas.go
go run ./scripts/ci/env.go ./scripts/ci/seed_test_data.go
bash scripts/ci/start-services.sh
bash scripts/ci/run-layer-tests.sh

echo ">>> test-run.sh OK"
