#!/bin/bash
# =============================================================================
# blog-server-go 服务器首次初始化（已有 Nest + MySQL/Redis/Nginx 时执行一次）
# 用法：bash deploy/pm2/server-init.sh
# =============================================================================

set -euo pipefail

APP_ROOT="/opt/jxapp/server/blog-server-go"
PUBLIC_ROOT="/opt/jxapp/server/blog-server/public"

echo "==> 创建部署目录"
sudo mkdir -p "${APP_ROOT}/logs"
sudo mkdir -p "${PUBLIC_ROOT}/uploads"
sudo chown -R "$(whoami):$(whoami)" /opt/jxapp

if ! command -v pm2 >/dev/null 2>&1; then
  echo "==> 安装 NVM + Node（仅用于 PM2 进程管理）"
  if [ ! -d "$HOME/.nvm" ]; then
    curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.4/install.sh | bash
  fi
  # shellcheck source=/dev/null
  source "$HOME/.nvm/nvm.sh"
  nvm install 22
  nvm use 22
  nvm alias default 22
  npm install -g pm2
else
  echo "==> PM2 已安装: $(pm2 -v)"
fi

echo "==> PM2 开机自启（若尚未配置）"
pm2 startup systemd -u "$(whoami)" --hp "$HOME" || true

echo "==> 完成。配置 deploy/pm2/env.production 后执行 deploy/pm2/deploy.ps1"
