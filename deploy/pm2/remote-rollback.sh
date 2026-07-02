#!/bin/bash
# =============================================================================
# blog-server-go 远程回滚（解压到新 release -> 切 current -> pm2 reload）
# =============================================================================

set -euo pipefail

: "${DEPLOY_REMOTE_DIR:?}"

DEPLOY_ECOSYSTEM_FILE="${DEPLOY_ECOSYSTEM_FILE:-ecosystem.config.js}"
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

BACKUP_ARG="${1:-}"
BACKUP_NAME_PATTERN='^backup-[0-9]{8}-[0-9]{6}\.tar\.gz$'

assert_backup_basename() {
  local name="$1"
  if [[ ! "$name" =~ $BACKUP_NAME_PATTERN ]]; then
    echo "Invalid backup name (expected backup-YYYYMMDD-HHMMSS.tar.gz): ${name}" >&2
    return 1
  fi
}

assert_backup_in_dir() {
  local file="$1"
  local dir_real file_real
  dir_real="$(realpath "$BACKUP_DIR")"
  file_real="$(realpath "$file")"
  if [[ "$file_real" != "$dir_real"/* ]]; then
    echo "Backup path escapes backup dir: ${file}" >&2
    return 1
  fi
}

resolve_backup_file() {
  if [[ -n "$BACKUP_ARG" ]]; then
    local base="${BACKUP_ARG##*/}"
    assert_backup_basename "$base" || return 1
    local candidate="${BACKUP_DIR}/${base}"
    if [[ -f "$candidate" ]]; then
      echo "$candidate"
      return 0
    fi
    echo "Backup not found: ${base}" >&2
    return 1
  fi
  ls -1t "${BACKUP_DIR}"/backup-*.tar.gz 2>/dev/null | head -1
}

if [[ "${DEPLOY_ROLLBACK_LIST:-}" == "1" ]]; then
  echo "Available backups in ${BACKUP_DIR}:"
  ls -1t "${BACKUP_DIR}"/backup-*.tar.gz 2>/dev/null || echo "(none)"
  echo ""
  echo "Available releases in ${RELEASES_ROOT}:"
  ls -1dt "${RELEASES_ROOT}"/*/ 2>/dev/null | grep -E '/[0-9]{8}-[0-9]{6}/$' || echo "(none)"
  echo "Current -> $(readlink -f "$CURRENT_LINK" 2>/dev/null || echo '(not set)')"
  exit 0
fi

BACKUP_FILE="$(resolve_backup_file)"
if [[ -z "$BACKUP_FILE" || ! -f "$BACKUP_FILE" ]]; then
  echo "No backup found in ${BACKUP_DIR}" >&2
  exit 1
fi
assert_backup_in_dir "$BACKUP_FILE"

mkdir -p "${DEPLOY_REMOTE_DIR}/logs" "${DEPLOY_PUBLIC_DIR}/uploads"

echo "==> rollback from: ${BACKUP_FILE}"

local_ts="$(release_new_id)"
release_path="$(release_dir_for "$local_ts")"
mkdir -p "$release_path"

echo "==> extract backup -> ${release_path}"
tar -xzf "${BACKUP_FILE}" -C "$release_path"

chmod +x "${release_path}/bin/gateway" "${release_path}/bin/user" "${release_path}/bin/blog" "${release_path}/bin/rpg"

mkdir -p "${DEPLOY_PUBLIC_DIR}/uploads"
ln -sfn "${DEPLOY_PUBLIC_DIR}" "${release_path}/public"

echo "==> activate rollback release"
release_switch "$release_path"
release_prune_legacy_root
release_pm2_reload_ecosystem "${DEPLOY_ECOSYSTEM_FILE}" "${DEPLOY_PM2_APPS}"

echo "==> rollback done"
release_pm2_verify_all_retry "${DEPLOY_PM2_APPS}" 5 3
