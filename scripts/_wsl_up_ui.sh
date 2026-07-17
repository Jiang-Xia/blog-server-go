#!/usr/bin/env bash
set -euo pipefail
cd /mnt/d/study/myGithub/blog-server-go

# ensure dockerd
if ! docker info >/dev/null 2>&1; then
  sudo dockerd >/tmp/dockerd.log 2>&1 &
  for i in $(seq 1 30); do
    docker info >/dev/null 2>&1 && break
    sleep 1
  done
fi

COMPOSE=(docker compose
  -f deploy/docker/docker-compose.yml
  -f deploy/docker/docker-compose.scale.yml
  -f deploy/docker/docker-compose.frontends.yml)

echo "==> down old stacks (microservices + monolith if any)"
"${COMPOSE[@]}" down --remove-orphans 2>/dev/null || true
docker compose -f deploy/docker/docker-compose.monolith.yml down --remove-orphans 2>/dev/null || true

echo "==> up scale=3 + uniapp + admin"
"${COMPOSE[@]}" up -d --build \
  --scale user=3 --scale blog=3 --scale rpg=3 --scale gateway=3

echo "==> wait nacos healthy / console"
for i in $(seq 1 90); do
  if curl -sf http://127.0.0.1:8848/nacos/ >/dev/null 2>&1; then echo "nacos ok"; break; fi
  [[ $i -eq 90 ]] && exit 1
  sleep 2
done

echo "==> wait edge API"
for i in $(seq 1 90); do
  if curl -sf http://127.0.0.1:8000/health >/dev/null 2>&1; then echo "api ok"; break; fi
  [[ $i -eq 90 ]] && { "${COMPOSE[@]}" ps; exit 1; }
  sleep 2
done

echo "==> wait uniapp/admin"
for i in $(seq 1 60); do
  u=$(curl -sf -o /dev/null -w '%{http_code}' http://127.0.0.1:8008/ || true)
  a=$(curl -sf -o /dev/null -w '%{http_code}' http://127.0.0.1:9856/ || true)
  echo "  [$i] uniapp=$u admin=$a"
  if [[ "$u" == "200" && "$a" == "200" ]]; then break; fi
  [[ $i -eq 60 ]] && exit 1
  sleep 3
done

echo ""
echo "==> replicas"
for svc in user blog rpg gateway; do
  n="$("${COMPOSE[@]}" ps -q "$svc" | wc -l | tr -d ' ')"
  echo "  $svc: $n"
done

echo ""
echo "==> nacos hosts"
for name in blog.user blog.blog blog.rpg; do
  json="$(curl -sf "http://127.0.0.1:8848/nacos/v1/ns/instance/list?serviceName=${name}&groupName=DEFAULT_GROUP" || true)"
  count="$(printf '%s' "$json" | python3 -c 'import sys,json; d=json.load(sys.stdin); print(len(d.get("hosts") or []))' 2>/dev/null || echo '?')"
  echo "  $name: $count"
done

echo ""
echo "==> CORS preflight sample"
curl -si -X OPTIONS "http://127.0.0.1:8000/api/v1/pub/stats" \
  -H "Origin: http://localhost:8008" \
  -H "Access-Control-Request-Method: GET" | head -20

echo ""
echo "==> pub/stats"
curl -sf http://127.0.0.1:8000/api/v1/pub/stats; echo

echo ""
"${COMPOSE[@]}" ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"
echo ""
echo "API     http://localhost:8000"
echo "uniapp  http://localhost:8008/"
echo "admin   http://localhost:9856/"
echo "Nacos   http://localhost:8848/nacos"
echo "DONE"
