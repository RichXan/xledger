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

bootstrap_test_commit="$(git -C "$WORKTREE_DIR" rev-list --reverse HEAD -- backend/internal/bootstrap/bootstrap_test.go | head -n 1)"
if [[ -z "$bootstrap_test_commit" ]]; then
  printf 'ERROR: unable to locate commit introducing backend/internal/bootstrap/bootstrap_test.go\n' >&2
  exit 1
fi

if ! baseline_commit="$(git -C "$WORKTREE_DIR" rev-parse "${bootstrap_test_commit}^" 2>/dev/null)"; then
  printf 'ERROR: commit introducing bootstrap tests has no parent to use as RED baseline\n' >&2
  exit 1
fi

printf '== Task 1 RED check on pre-test baseline %s ==\n' "$baseline_commit"
git -C "$WORKTREE_DIR" worktree add --detach "$tmp_worktree" "$baseline_commit" >/dev/null

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
printf 'RED verified: pre-test baseline bootstrap tests failed (exit=%d)\n' "$red_exit"

printf '\n== Task 1 GREEN check on current commit ==\n'
(
  cd "$BACKEND_DIR" && go test ./internal/bootstrap -v
)
printf 'GREEN verified: current commit bootstrap tests passed\n'
