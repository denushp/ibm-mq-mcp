#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
SKILL_SOURCE="${REPO_ROOT}/skills/ibm-mq-mcp"

INSTALL_CODEX=true
INSTALL_CLAUDE=true
CLAUDE_SCOPE="user"
DRY_RUN=false
MQ_INSTALL_PATH_ARG=""
BINARY_ARG=""

usage() {
  cat <<'EOF'
Install IBM MQ MCP tooling for Codex and Claude.

Usage:
  scripts/install-ai-tooling.sh [options]

Options:
  --mq-install-path PATH   IBM MQ client install root. Defaults to $MQ_INSTALL_PATH.
  --binary PATH            Path to ibm-mq-mcp. Defaults to PATH lookup, then ./ibm-mq-mcp.
  --codex-only             Install only Codex skill and MCP config.
  --claude-only            Install only Claude skill and MCP config.
  --claude-scope SCOPE     Claude MCP scope: user or project. Default: user.
  --dry-run                Print actions without changing anything.
  --help                   Show this help.

Examples:
  scripts/install-ai-tooling.sh \
    --mq-install-path /opt/mqm \
    --binary "$(pwd)/ibm-mq-mcp"

  scripts/install-ai-tooling.sh \
    --mq-install-path "$MQ_INSTALL_PATH" \
    --dry-run
EOF
}

log() {
  printf '[ibm-mq-mcp] %s\n' "$*"
}

fail() {
  printf '[ibm-mq-mcp] ERROR: %s\n' "$*" >&2
  exit 1
}

run() {
  if [ "${DRY_RUN}" = true ]; then
    printf '+'
    for arg in "$@"; do
      printf ' %q' "${arg}"
    done
    printf '\n'
    return 0
  fi
  "$@"
}

resolve_binary() {
  if [ -n "${BINARY_ARG}" ]; then
    [ -x "${BINARY_ARG}" ] || fail "binary is not executable: ${BINARY_ARG}"
    printf '%s\n' "${BINARY_ARG}"
    return 0
  fi

  if command -v ibm-mq-mcp >/dev/null 2>&1; then
    command -v ibm-mq-mcp
    return 0
  fi

  if [ -x "${REPO_ROOT}/ibm-mq-mcp" ]; then
    printf '%s\n' "${REPO_ROOT}/ibm-mq-mcp"
    return 0
  fi

  fail "could not find ibm-mq-mcp. Pass --binary or put it on PATH."
}

install_skill_link() {
  local target_root="$1"
  local target_link="${target_root}/ibm-mq-mcp"
  run mkdir -p "${target_root}"
  run ln -snf "${SKILL_SOURCE}" "${target_link}"
}

command_exists_or_warn() {
  local command_name="$1"
  if command -v "${command_name}" >/dev/null 2>&1; then
    return 0
  fi

  if [ "${DRY_RUN}" = true ]; then
    log "warning: ${command_name} is not installed in this shell, but dry-run will continue"
    return 1
  fi

  fail "${command_name} is not installed or not on PATH"
}

remove_existing_codex_server() {
  if command -v codex >/dev/null 2>&1 && codex mcp get ibm-mq >/dev/null 2>&1; then
    run codex mcp remove ibm-mq
  fi
}

remove_existing_claude_server() {
  if command -v claude >/dev/null 2>&1; then
    if [ "${DRY_RUN}" = true ]; then
      printf '+ %q %q %q %q %q %q >/dev/null 2>&1 || true\n' claude mcp remove ibm-mq -s "${CLAUDE_SCOPE}"
    else
      claude mcp remove ibm-mq -s "${CLAUDE_SCOPE}" >/dev/null 2>&1 || true
    fi
  fi
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --mq-install-path)
      [ "$#" -ge 2 ] || fail "--mq-install-path requires a value"
      MQ_INSTALL_PATH_ARG="$2"
      shift 2
      ;;
    --binary)
      [ "$#" -ge 2 ] || fail "--binary requires a value"
      BINARY_ARG="$2"
      shift 2
      ;;
    --codex-only)
      INSTALL_CODEX=true
      INSTALL_CLAUDE=false
      shift
      ;;
    --claude-only)
      INSTALL_CODEX=false
      INSTALL_CLAUDE=true
      shift
      ;;
    --claude-scope)
      [ "$#" -ge 2 ] || fail "--claude-scope requires a value"
      CLAUDE_SCOPE="$2"
      shift 2
      ;;
    --dry-run)
      DRY_RUN=true
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      fail "unknown option: $1"
      ;;
  esac
done

case "${CLAUDE_SCOPE}" in
  user|project)
    ;;
  *)
    fail "--claude-scope must be user or project"
    ;;
esac

MQ_INSTALL_PATH_VALUE="${MQ_INSTALL_PATH_ARG:-${MQ_INSTALL_PATH:-}}"
[ -n "${MQ_INSTALL_PATH_VALUE}" ] || fail "set MQ_INSTALL_PATH or pass --mq-install-path"
[ -d "${MQ_INSTALL_PATH_VALUE}/lib64" ] || fail "expected MQ client libraries at ${MQ_INSTALL_PATH_VALUE}/lib64"
[ -d "${SKILL_SOURCE}" ] || fail "shared skill directory not found: ${SKILL_SOURCE}"

BINARY_PATH="$(resolve_binary)"
DYLD_LIBRARY_PATH_VALUE="${MQ_INSTALL_PATH_VALUE}/lib64"
LD_LIBRARY_PATH_VALUE="${MQ_INSTALL_PATH_VALUE}/lib64"

log "repo root: ${REPO_ROOT}"
log "skill source: ${SKILL_SOURCE}"
log "mq install path: ${MQ_INSTALL_PATH_VALUE}"
log "binary: ${BINARY_PATH}"

if [ "${INSTALL_CODEX}" = true ]; then
  command_exists_or_warn codex || true
  log "installing Codex skill"
  install_skill_link "${HOME}/.codex/skills"

  if command -v codex >/dev/null 2>&1 || [ "${DRY_RUN}" = true ]; then
    log "refreshing Codex MCP server entry"
    remove_existing_codex_server
    run codex mcp add ibm-mq \
      --env "MQ_INSTALL_PATH=${MQ_INSTALL_PATH_VALUE}" \
      --env "DYLD_LIBRARY_PATH=${DYLD_LIBRARY_PATH_VALUE}" \
      --env "LD_LIBRARY_PATH=${LD_LIBRARY_PATH_VALUE}" \
      -- "${BINARY_PATH}"
  fi
fi

if [ "${INSTALL_CLAUDE}" = true ]; then
  command_exists_or_warn claude || true
  log "installing Claude skill"
  install_skill_link "${HOME}/.claude/skills"

  if command -v claude >/dev/null 2>&1 || [ "${DRY_RUN}" = true ]; then
    log "refreshing Claude MCP server entry at scope=${CLAUDE_SCOPE}"
    remove_existing_claude_server
    run claude mcp add ibm-mq --scope "${CLAUDE_SCOPE}" \
      -e "MQ_INSTALL_PATH=${MQ_INSTALL_PATH_VALUE}" \
      -e "DYLD_LIBRARY_PATH=${DYLD_LIBRARY_PATH_VALUE}" \
      -e "LD_LIBRARY_PATH=${LD_LIBRARY_PATH_VALUE}" \
      -- "${BINARY_PATH}"
  fi
fi

log "done"
log "restart Codex to reload installed skills"
log "start a new Claude Code session if you want the personal Claude skill immediately available"
