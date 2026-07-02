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
  export PM2_VERIFY_APP="$app"
  release_pm2 jlist 2>/dev/null | node -e "
    const chunks = [];
    process.stdin.on('data', (d) => chunks.push(d));
    process.stdin.on('end', () => {
      const app = process.env.PM2_VERIFY_APP;
      let list = [];
      try { list = JSON.parse(Buffer.concat(chunks).toString('utf8') || '[]'); }
      catch (e) { process.exit(1); }
      const hit = list.find((p) => p.name === app);
      if (!hit) process.exit(1);
      const st = (hit.pm2_env && hit.pm2_env.status) || '';
      process.exit(st === 'online' ? 0 : 1);
    });
  " >/dev/null 2>&1
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

# cwd 是否为 $DEPLOY_REMOTE_DIR/current（软链路径），reload 零停机前提。
release_pm2_cwd_uses_current_link() {
  local app="$1"
  local expected="${DEPLOY_REMOTE_DIR}/current"
  export PM2_VERIFY_APP="$app"
  export PM2_EXPECTED_CWD="$expected"
  release_pm2 jlist | node -e "
    const chunks = [];
    process.stdin.on('data', (d) => chunks.push(d));
    process.stdin.on('end', () => {
      const app = process.env.PM2_VERIFY_APP;
      const expected = process.env.PM2_EXPECTED_CWD;
      let list = [];
      try { list = JSON.parse(Buffer.concat(chunks).toString('utf8') || '[]'); }
      catch (e) { process.exit(1); }
      const hit = list.find((p) => p.name === app);
      if (!hit) process.exit(1);
      const cwd = (hit.pm2_env && hit.pm2_env.pm_cwd) || '';
      process.exit(cwd === expected ? 0 : 1);
    });
  " >/dev/null 2>&1
}

release_pm2_start_ordered() {
  local eco="$1"
  echo "==> pm2 first start (BlogGo_User -> BlogGo_Blog,BlogGo_Rpg -> BlogGo_Gateway): ${eco}"
  release_pm2 start "$eco" --env production --only BlogGo_User
  release_pm2_wait_online BlogGo_User 120
  release_pm2 start "$eco" --env production --only BlogGo_Blog,BlogGo_Rpg
  release_pm2_wait_online BlogGo_Blog 120
  release_pm2_wait_online BlogGo_Rpg 120
  release_pm2 start "$eco" --env production --only BlogGo_Gateway
  release_pm2_wait_online BlogGo_Gateway 120
  release_pm2 save
}

release_pm2_reload_ordered() {
  local eco="$1"
  echo "==> pm2 reload zero-downtime: ${eco}"
  release_pm2 reload "$eco" --env production --only BlogGo_User --update-env
  release_pm2_wait_online BlogGo_User 90
  release_pm2 reload "$eco" --env production --only BlogGo_Blog,BlogGo_Rpg --update-env
  release_pm2_wait_online BlogGo_Blog 90
  release_pm2_wait_online BlogGo_Rpg 90
  release_pm2 reload "$eco" --env production --only BlogGo_Gateway --update-env
  release_pm2_wait_online BlogGo_Gateway 90
  release_pm2 save
}

release_pm2_delete_bloggo_apps() {
  local apps_csv="$1"
  local app
  IFS=',' read -ra stop_list <<< "$apps_csv"
  for app in "${stop_list[@]}"; do
    app="${app// /}"
    [[ -z "$app" ]] && continue
    release_pm2 delete "$app" 2>/dev/null || true
  done
}

release_pm2_reload_ecosystem() {
  local eco="$1"
  local apps_csv="${2:-BlogGo_User,BlogGo_Blog,BlogGo_Rpg,BlogGo_Gateway}"

  release_ensure_pm2_env || return 1
  export DEPLOY_REMOTE_DIR
  cd "${DEPLOY_REMOTE_DIR}/current"

  local legacy
  for legacy in gateway user blog rpg; do
    release_pm2 delete "$legacy" 2>/dev/null || true
  done

  if release_pm2 describe "BlogGo_Gateway" >/dev/null 2>&1; then
    if release_pm2_cwd_uses_current_link "BlogGo_User"; then
      release_pm2_reload_ordered "$eco"
    else
      echo "==> pm2 cwd 非 current 软链，一次性 recreate 后后续可 reload"
      release_pm2_delete_bloggo_apps "$apps_csv"
      release_pm2_start_ordered "$eco"
    fi
  else
    release_pm2_start_ordered "$eco"
  fi

  local app
  IFS=',' read -ra app_list <<< "$apps_csv"
  for app in "${app_list[@]}"; do
    app="${app// /}"
    [[ -z "$app" ]] && continue
    release_pm2_wait_online "$app" 120
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

# 部署末尾验证：Go 服务冷启动略慢，失败时重试避免误报。
release_pm2_verify_all_retry() {
  local apps_csv="$1"
  local max="${2:-5}"
  local interval="${3:-3}"
  local attempt=1
  while (( attempt <= max )); do
    if release_pm2_verify_all "$apps_csv"; then
      return 0
    fi
    if (( attempt < max )); then
      echo "==> pm2 verify retry ${attempt}/${max} in ${interval}s..." >&2
      sleep "$interval"
    fi
    attempt=$((attempt + 1))
  done
  return 1
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
