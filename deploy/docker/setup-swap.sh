#!/bin/sh
# 2G 机器 swap 兜底（Plan 10 部署前可选执行）
set -e
if [ ! -f /swapfile ]; then
  fallocate -l 1G /swapfile
  chmod 600 /swapfile
  mkswap /swapfile
  swapon /swapfile
  grep -q '/swapfile' /etc/fstab || echo '/swapfile none swap sw 0 0' >> /etc/fstab
  echo "swap enabled"
else
  echo "swapfile already exists"
fi
