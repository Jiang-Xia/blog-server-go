#!/bin/bash
# =============================================================================
# 方案 B：releases/ 版本目录 + current 软链（blog-server-go PM2 部署共用）
# =============================================================================

: "${DEPLOY_REMOTE_DIR:?}"

RELEASES_ROOT="${DEPLOY_REMOTE_DIR}/releases"
BACKUP_DIR="${RELEASES_ROOT}/backups"
CURRENT_LINK="${DEPLOY_REMOTE_DIR}/current"
RELEASE_ID_PATTERN='^[0-9]{8}-[0-9]{6}$'

# 非交互 SSH（plink）下加载 nvm，确保 pm2 在 PATH 中。
release_ensure_pm2_env() {
  if command -v pm2 >/dev/null 2>&1; then
    return 0
  fi
  export NVM_DIR="${NVM_DIR:-$HOME/.nvm}"
  if [[ -s "$NVM_DIR/nvm.sh" ]]; then
    # shellcheck source=/dev/null
    source "$NVM_DIR/nvm.sh"
    nvm use default >/dev/null 2>&1 || nvm use node >/dev/null 2>&1 || true
  fi
  if command -v pm2 >/dev/null 2>&1; then
    return 0
  fi
  local pm2bin
  pm2bin="$(find "$NVM_DIR/versions/node" -maxdepth 3 -type f -name pm2 2>/dev/null | head -1)"
  if [[ -n "$pm2bin" ]]; then
    export PATH="$(dirname "$pm2bin"):$PATH"
  fi
  if command -v pm2 >/dev/null 2>&1; then
    return 0
  fi
  echo "ERROR: pm2 not found. SSH 登录服务器后执行: bash deploy/pm2/server-init.sh" >&2
  return 1
}

release_new_id() {
  date +%Y%m%d-%H%M%S
}

release_dir_for() {
  echo "${RELEASES_ROOT}/$1"
}

release_get_active() {
  if [[ -L "$CURRENT_LINK" ]]; then
    readlink -f "$CURRENT_LINK"
    return 0
  fi
  return 1
}

release_switch() {
  local release_path="$1"
  ln -sfn "$release_path" "$CURRENT_LINK"
  echo "==> switch current -> ${release_path}"
}

release_cleanup() {
  local keep="${DEPLOY_RELEASE_KEEP:-5}"
  local count=0 name
  for name in $(ls -1t "$RELEASES_ROOT" 2>/dev/null); do
    [[ "$name" == "backups" ]] && continue
    [[ "$name" =~ $RELEASE_ID_PATTERN ]] || continue
    count=$((count + 1))
    if (( count > keep )); then
      echo "==> remove old release: ${RELEASES_ROOT}/${name}"
      rm -rf "${RELEASES_ROOT}/${name}"
    fi
  done
}

release_prune_backups() {
  local keep="${DEPLOY_BACKUP_KEEP:-5}"
  ls -1t "$BACKUP_DIR"/backup-*.tar.gz 2>/dev/null | tail -n +$((keep + 1)) | while IFS= read -r f; do
    [[ -n "$f" ]] && rm -f "$f"
  done
}

release_pm2() {
  release_ensure_pm2_env || return 1
  PM2_SILENT=true pm2 "$@"
}

release_pm2_all_online() {
  local app="$1"
  local pid
  pid="$(release_pm2 pid "$app" 2>/dev/null || true)"
  [[ -n "$pid" && "$pid" != "0" && "$pid" != "[]" ]]
}

release_pm2_wait_online() {
  local app="$1"
  local max="${2:-60}"
  local waited=0
  while (( waited < max )); do
    if release_pm2_all_online "$app"; then
      echo "==> pm2 ${app} online (${waited}s)"
      return 0
    fi
    sleep 2
    waited=$((waited + 2))
  done
  echo "==> pm2 wait online timeout: ${app} (${max}s)" >&2
  return 1
}

release_pm2_reload_ecosystem() {
  local eco="$1"
  local apps_csv="${2:-gateway,user,blog,rpg}"

  release_ensure_pm2_env || return 1
  cd "$CURRENT_LINK"

  if release_pm2 describe "gateway" >/dev/null 2>&1; then
    echo "==> pm2 reload ecosystem: ${eco}"
    release_pm2 reload "$eco" --env production --update-env
    release_pm2 save
  else
    echo "==> pm2 first start (user -> blog,rpg -> gateway)"
    release_pm2 start "$eco" --env production --only user
    release_pm2 start "$eco" --env production --only blog,rpg
    sleep 2
    release_pm2 start "$eco" --env production --only gateway
    release_pm2 save
  fi

  local app
  IFS=',' read -ra app_list <<< "$apps_csv"
  for app in "${app_list[@]}"; do
    app="${app// /}"
    [[ -z "$app" ]] && continue
    release_pm2_wait_online "$app" 90
  done
}

release_pm2_verify() {
  local app="$1"
  release_ensure_pm2_env || return 1
  if ! release_pm2 describe "$app" >/dev/null 2>&1; then
    echo "ERROR: pm2 app not found: $app" >&2
    return 1
  fi
  echo "==> pm2 describe: ${app}"
  release_pm2 describe "$app" | grep -E 'status|script path|exec cwd' || true
  if ! release_pm2_all_online "$app"; then
    echo "ERROR: ${app} not online" >&2
    return 1
  fi
}

release_pm2_verify_all() {
  local apps_csv="$1"
  local app
  IFS=',' read -ra app_list <<< "$apps_csv"
  for app in "${app_list[@]}"; do
    app="${app// /}"
    [[ -z "$app" ]] && continue
    release_pm2_verify "$app"
  done
}

release_prune_legacy_root() {
  [[ -L "$CURRENT_LINK" ]] || return 0
  local item
  for item in bin configs ecosystem.config.js ecosystem.config.cjs dist package.json node_modules .env.production; do
    if [[ -e "${DEPLOY_REMOTE_DIR}/${item}" && ! -L "${DEPLOY_REMOTE_DIR}/${item}" ]]; then
      echo "==> remove legacy root item: ${item}"
      rm -rf "${DEPLOY_REMOTE_DIR}/${item}"
    fi
  done
}
