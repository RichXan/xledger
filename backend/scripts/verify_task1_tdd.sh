#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
WORKTREE_DIR="$(cd "$BACKEND_DIR/.." && pwd)"

tmp_worktree="$(mktemp -d "${TMPDIR:-/tmp}/xledger-task1-red-XXXXXX")"
cleanup() {
  if git -C "$WORKTREE_DIR" worktree list | awk '{print $1}' | grep -Fxq "$tmp_worktree"; then
    git -C "$WORKTREE_DIR" worktree remove --force "$tmp_worktree" >/dev/null 2>&1 || true
  fi
  rm -rf "$tmp_worktree"
}
trap cleanup EXIT

printf '== Task 1 RED check on parent commit ==\n'
git -C "$WORKTREE_DIR" worktree add --detach "$tmp_worktree" HEAD~1 >/dev/null

set +e
(
  cd "$tmp_worktree/backend" && go test ./internal/bootstrap -v
)
red_exit=$?
set -e

if [[ $red_exit -eq 0 ]]; then
  printf 'ERROR: expected parent commit bootstrap tests to fail, but they passed\n' >&2
  exit 1
fi
printf 'RED verified: parent commit bootstrap tests failed (exit=%d)\n' "$red_exit"

printf '\n== Task 1 GREEN check on current commit ==\n'
(
  cd "$BACKEND_DIR" && go test ./internal/bootstrap -v
)
printf 'GREEN verified: current commit bootstrap tests passed\n'
