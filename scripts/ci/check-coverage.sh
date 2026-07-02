#!/usr/bin/env bash
# 单元测试覆盖率门禁（默认 pkg 总覆盖率 ≥ 40%）
set -euo pipefail

MIN="${MIN_PKG_COVERAGE:-40}"
PROFILE="${COVERAGE_PROFILE:-coverage.out}"
# 仅统计已有 _test.go 的 pkg 子包（随单元测试扩充而扩大）
PKGS="./pkg/crypto ./pkg/jwtauth ./pkg/pagination ./pkg/errcode ./pkg/timeutil"

go test $PKGS -count=1 -covermode=atomic -coverprofile="$PROFILE"

total="$(go tool cover -func="$PROFILE" | awk '/^total:/ {gsub(/%/,"",$3); print $3}')"
echo "pkg coverage: ${total}% (min ${MIN}%)"

awk -v t="$total" -v m="$MIN" 'BEGIN { exit (t+0 >= m+0 ? 0 : 1) }' || {
  echo "coverage below threshold ${MIN}%" >&2
  exit 1
}
