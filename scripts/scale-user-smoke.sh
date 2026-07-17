#!/usr/bin/env bash
# 四服务各多实例冒烟：Nacos 注册实例数 + pub/stats 负载命中分布。
# 用法（WSL）: bash scripts/scale-user-smoke.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

COMPOSE=(docker compose
  -f deploy/docker/docker-compose.yml
  -f deploy/docker/docker-compose.scale.yml)

echo "==> 副本数量"
for svc in user blog rpg gateway; do
  n="$("${COMPOSE[@]}" ps -q "$svc" | wc -l | tr -d ' ')"
  echo "  $svc: $n"
done

echo ""
echo "==> 等待 edge/gateway :8000"
for i in $(seq 1 60); do
  if curl -sf "http://127.0.0.1:8000/health" >/dev/null 2>&1; then
    echo "ready"
    break
  fi
  if [[ "$i" -eq 60 ]]; then
    echo "未就绪" >&2
    exit 1
  fi
  sleep 2
done

echo ""
echo "==> Nacos Kitex 注册实例"
for name in blog.user blog.blog blog.rpg; do
  json="$(curl -sf "http://127.0.0.1:8848/nacos/v1/ns/instance/list?serviceName=${name}&groupName=DEFAULT_GROUP" || true)"
  count="$(printf '%s' "$json" | grep -o '"ip"' | wc -l | tr -d ' ')"
  echo "  $name: ${count:-0}"
done
echo "  控制台: http://localhost:8848/nacos → 服务管理"

echo ""
echo "==> GET /api/v1/pub/stats × ${ROUNDS:-30}"
ROUNDS="${ROUNDS:-30}"
ok=0
for i in $(seq 1 "$ROUNDS"); do
  if curl -sf "http://127.0.0.1:8000/api/v1/pub/stats" >/dev/null; then
    ok=$((ok + 1))
  fi
done
echo "成功 $ok / $ROUNDS"

hit_svc() {
  local svc="$1" pattern="$2"
  echo ""
  echo "==> $svc 命中 ($pattern)"
  mapfile -t CIDS < <("${COMPOSE[@]}" ps -q "$svc")
  for cid in "${CIDS[@]}"; do
    name="$(docker inspect -f '{{.Name}}' "$cid" | sed 's#^/##')"
    host="$(docker inspect -f '{{.Config.Hostname}}' "$cid")"
    n="$(docker logs "$cid" 2>&1 | grep -c "$pattern" || true)"
    echo "  $name hostname=$host hits=$n"
  done
}

hit_svc gateway 'pub/stats gateway_instance='
hit_svc user 'CountUsers instance='
hit_svc blog 'GetPubStats instance='

echo ""
echo "完成。Nacos 各服务约 3 实例、且多 hostname 有 hits>0 即多实例生效。"
echo "说明：blog/rpg cron 也会×3，仅适合学习，勿当生产。"
