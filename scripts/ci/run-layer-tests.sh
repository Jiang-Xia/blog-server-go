#!/usr/bin/env bash
# 按行业分层顺序跑测试；集成层需服务已启动。
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

export DEV_LOGIN_BASE="${DEV_LOGIN_BASE:-http://127.0.0.1:8000}"
export TEST_BASE="${TEST_BASE:-$DEV_LOGIN_BASE}"
export CI_JWT_SECRET="${CI_JWT_SECRET:-ci-integration-test-secret}"

run_unit="${RUN_UNIT:-1}"
run_smoke="${RUN_SMOKE:-1}"
run_integration="${RUN_INTEGRATION:-1}"
run_e2e="${RUN_E2E:-1}"

if [[ "$run_unit" == "1" ]]; then
  echo "=== unit ==="
  bash scripts/ci/check-coverage.sh
fi

if [[ "$run_smoke" == "1" ]]; then
  echo "=== smoke ==="
  go test -tags=smoke ./test/smoke/... -count=1 -v
fi

if [[ "$run_integration" == "1" ]]; then
  echo "=== integration ==="
  go test -tags=integration ./test/integration/... -count=1 -v
fi

if [[ "$run_e2e" == "1" ]]; then
  echo "=== e2e ==="
  go test -tags=e2e ./test/e2e/... -count=1 -v
fi

echo "=== all requested test layers passed ==="
