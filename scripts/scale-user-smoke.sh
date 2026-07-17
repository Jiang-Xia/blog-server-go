#!/usr/bin/env bash
# 四服务各多实例冒烟：etcd 注册条数 + pub/stats 负载命中分布。
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
echo "==> etcd Kitex 注册"
for name in blog.user blog.blog blog.rpg; do
  keys="$("${COMPOSE[@]}" exec -T etcd etcdctl get --prefix "kitex/registry-etcd/${name}/" --keys-only 2>/dev/null | grep -c "$name" || true)"
  echo "  $name: ${keys:-0}"
  "${COMPOSE[@]}" exec -T etcd etcdctl get --prefix "kitex/registry-etcd/${name}/" --keys-only 2>/dev/null || true
done

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
echo "完成。etcd 各服务约 3 条、且多 hostname 有 hits>0 即多实例生效。"
echo "说明：blog/rpg cron 也会×3，仅适合学习，勿当生产。"
