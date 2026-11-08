#!/usr/bin/env bash
# Non-mutating Operation CWAL current-state gate.
set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"
fail() { echo "CWAL CHECK FAIL: $*" >&2; exit 1; }
for file in "operation cwal.md" "BLACK_SHEEP_WALL.md" "workflow/active_work/README.md" "handoff.md" "active_work/supported-profile.md"; do
  [[ -f "$file" ]] || fail "missing $file"
done
read_markers() {
  local file="$1" key value
  for key in Repository Package Branch CURRENT QUEUED Status; do
    value="$(sed -n "s/^$key: //p" "$file" | head -n 1)"
    [[ -n "$value" ]] || fail "missing $key in $file"
    printf '%s\n' "$value"
  done
}
mapfile -t a < <(read_markers workflow/active_work/README.md)
mapfile -t b < <(read_markers handoff.md)
for i in "${!a[@]}"; do
  [[ "${a[$i]}" == "${b[$i]}" ]] || fail "state markers disagree (index $i)"
done
[[ "${a[0]}" == "tamzrod/dnp3" ]] || fail "wrong repository"
if [[ "${a[3]}" == "NONE" ]]; then
  [[ "${a[2]}" == "NONE" && "${a[1]}" == "NONE" && "${a[4]}" == "NONE" ]] || fail "NONE CURRENT with inconsistent package/branch/queue"
  [[ "${a[5]}" == "NO_CURRENT" || "${a[5]}" == "WORK_PACKAGE_READY" ]] || fail "unexpected empty-state status"
  echo "CWAL CHECK OK: NO_CURRENT; no implementation authority"
  exit 0
fi
[[ "${a[1]}" != "NONE" && "${a[2]}" != "NONE" ]] || fail "CURRENT lacks package or branch"
[[ "${a[2]}" != "main" ]] || fail "main is not a work branch"
[[ "${a[3]}" =~ ^[A-Za-z0-9._-]+$ ]] || fail "invalid CURRENT identifier"
[[ -f "workflow/active_work/${a[3]}.md" ]] || fail "CURRENT task file missing"
[[ "${a[5]}" == "TASK_IN_PROGRESS" || "${a[5]}" == "TASK_BLOCKED" || "${a[5]}" == "TASK_READY" ]] || fail "unexpected active status"
[[ "$(git branch --show-current)" == "${a[2]}" ]] || fail "checkout is not designated branch"
[[ -z "$(git status --porcelain)" ]] || fail "working tree dirty; preserve and reconcile safely"
git rev-parse --verify "refs/remotes/origin/${a[2]}" >/dev/null 2>&1 || fail "missing local remote-tracking branch; fetch safely"
[[ "$(git rev-parse HEAD)" == "$(git rev-parse "refs/remotes/origin/${a[2]}")" ]] || fail "local/remote-tracking SHA mismatch; fetch and reconcile"
echo "CWAL CHECK OK: CURRENT=${a[3]} branch=${a[2]} (remote freshness requires explicit fetch)"
