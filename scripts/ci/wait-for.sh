#!/usr/bin/env bash
# 等待 MySQL / Redis 就绪（CI 与 docker-compose.test 通用）
set -euo pipefail

host="${CI_MYSQL_HOST:-127.0.0.1}"
port="${CI_MYSQL_PORT:-3306}"
user="${CI_MYSQL_USER:-root}"
pass="${CI_MYSQL_PASSWORD:-testpass}"
redis="${CI_REDIS_ADDR:-127.0.0.1:6379}"

echo "wait mysql ${host}:${port}..."
for i in $(seq 1 60); do
  if mysqladmin ping -h "$host" -P "$port" -u"$user" -p"$pass" --silent 2>/dev/null; then
    echo "mysql ready"
    break
  fi
  if [[ $i -eq 60 ]]; then
    echo "mysql timeout" >&2
    exit 1
  fi
  sleep 2
done

redis_host="${redis%%:*}"
redis_port="${redis##*:}"
echo "wait redis ${redis_host}:${redis_port}..."
for i in $(seq 1 30); do
  if redis-cli -h "$redis_host" -p "$redis_port" ping 2>/dev/null | grep -q PONG; then
    echo "redis ready"
    exit 0
  fi
  if [[ $i -eq 30 ]]; then
    echo "redis timeout" >&2
    exit 1
  fi
  sleep 1
done
