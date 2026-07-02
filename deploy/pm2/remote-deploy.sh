#!/bin/bash
# =============================================================================
# blog-server-go 远程部署（方案 B：releases + current，PM2 reload 四服务）
# =============================================================================

set -euo pipefail

: "${DEPLOY_REMOTE_DIR:?}"
: "${DEPLOY_ECOSYSTEM_FILE:?}"
: "${DEPLOY_TAR_PATH:?}"

DEPLOY_PM2_APPS="${DEPLOY_PM2_APPS:-BlogGo_User,BlogGo_Blog,BlogGo_Rpg,BlogGo_Gateway}"
DEPLOY_PUBLIC_DIR="${DEPLOY_PUBLIC_DIR:-${DEPLOY_REMOTE_DIR}/public}"

source /tmp/release-lib.sh
export NVM_DIR="${NVM_DIR:-$HOME/.nvm}"
if [[ -s "$NVM_DIR/nvm.sh" ]]; then
  # shellcheck source=/dev/null
  source "$NVM_DIR/nvm.sh"
  nvm use default >/dev/null 2>&1 || nvm use node >/dev/null 2>&1 || true
fi
release_ensure_pm2_env || exit 1

BACKUP_KEEP="${DEPLOY_BACKUP_KEEP:-5}"

migrate_legacy_layout_if_needed() {
  if [[ -L "$CURRENT_LINK" || -e "$CURRENT_LINK" ]]; then
    return 0
  fi
  if [[ ! -d "${DEPLOY_REMOTE_DIR}/bin" && ! -f "${DEPLOY_REMOTE_DIR}/${DEPLOY_ECOSYSTEM_FILE}" ]]; then
    return 0
  fi

  local ts rid item
  ts="$(release_new_id)"
  rid="$(release_dir_for "$ts")"
  mkdir -p "$rid"

  for item in bin configs "${DEPLOY_ECOSYSTEM_FILE}"; do
    [[ -e "${DEPLOY_REMOTE_DIR}/${item}" ]] && mv "${DEPLOY_REMOTE_DIR}/${item}" "$rid/"
  done

  release_switch "$rid"
  link_shared_public "$rid"
  echo "==> migrated legacy layout -> ${rid}"
  if release_pm2 describe "BlogGo_Gateway" >/dev/null 2>&1; then
    echo "==> pm2 reload after legacy migration"
    release_pm2_reload_ecosystem "${DEPLOY_ECOSYSTEM_FILE}" "${DEPLOY_PM2_APPS}"
  fi
}

backup_before_deploy() {
  local active items=()
  active="$(release_get_active 2>/dev/null || true)"

  if [[ -z "$active" || ! -d "$active" ]]; then
    echo "==> skip backup (no active release)"
    return 0
  fi

  [[ -d "${active}/bin" ]] && items+=(bin)
  [[ -d "${active}/configs" ]] && items+=(configs)
  [[ -f "${active}/${DEPLOY_ECOSYSTEM_FILE}" ]] && items+=("${DEPLOY_ECOSYSTEM_FILE}")

  if ((${#items[@]} == 0)); then
    echo "==> skip backup (empty active release)"
    return 0
  fi

  mkdir -p "$BACKUP_DIR"
  local ts backup_file
  ts="$(release_new_id)"
  backup_file="${BACKUP_DIR}/backup-${ts}.tar.gz"
  echo "==> backup: ${backup_file}"
  tar -czf "$backup_file" -C "$active" "${items[@]}"
  release_prune_backups
  echo "==> backup kept: latest ${BACKUP_KEEP}"
}

link_shared_public() {
  local release_path="$1"
  mkdir -p "${DEPLOY_PUBLIC_DIR}/uploads"
  ln -sfn "${DEPLOY_PUBLIC_DIR}" "${release_path}/public"
}

mkdir -p "$RELEASES_ROOT" "$BACKUP_DIR" "${DEPLOY_REMOTE_DIR}/logs"

migrate_legacy_layout_if_needed
backup_before_deploy

local_ts="$(release_new_id)"
release_path="$(release_dir_for "$local_ts")"
mkdir -p "$release_path"

echo "==> extract: ${DEPLOY_TAR_PATH} -> ${release_path}"
tar -xzf "${DEPLOY_TAR_PATH}" -C "$release_path"

chmod +x "${release_path}/bin/gateway" "${release_path}/bin/user" "${release_path}/bin/blog" "${release_path}/bin/rpg"

link_shared_public "$release_path"

echo "==> activate release (pm2 reload ecosystem)"
release_switch "$release_path"
release_prune_legacy_root
release_pm2_reload_ecosystem "${DEPLOY_ECOSYSTEM_FILE}" "${DEPLOY_PM2_APPS}"
release_cleanup

rm -f "${DEPLOY_TAR_PATH}"

echo "==> done"
release_pm2_verify_all_retry "${DEPLOY_PM2_APPS}" 5 3
