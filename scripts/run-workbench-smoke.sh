#!/usr/bin/env bash
#
# scripts/run-workbench-smoke.sh — DNP3 workbench external-smoke (MEXT-023)
#
# Runs the programmatic end-to-end workbench smoke (scripts/workbench_e2e.go)
# against the PUBLIC API (pkg/dnp3/master + pkg/dnp3/outstation) over a REAL
# TCP loopback with no simulator transport:
#
#   Outstation Start → Master Connect → Class-0 Read (G1/G30/G20) →
#   per-group Reads → Operate (CROB) → clean shutdown.
#
# This is the CI-runnable analogue of the interactive workbench TUI smoke
# (cmd/workbench): it exercises the same external caller surface an operator
# drives by hand, but as a non-interactive `go run` so it can gate CI/docs. It
# does NOT require the workbench TUI binary to be built; if the workbench module
# is absent or `go run` cannot build the smoke program, the script reports a
# clear error and exits non-zero (per MEXT-023: "skip without failing unit
# packages if missing" — unit packages are unaffected; this script is the
# reproducible-from-README smoke entry point).
#
# Usage:
#   scripts/run-workbench-smoke.sh            # run the smoke (exit 0 = pass)
#
# Exit codes:
#   0  smoke passed (Connect → Read → Operate → Shutdown all green)
#   1  smoke failed (build error or a step asserted failure)
#   2  go toolchain not found
#
# See MEXT-023 in active_work/MEXT_MASTER_ROADMAP.md.

set -u

GO="$(command -v go || true)"
if [ -z "$GO" ]; then
    if [ -x "$HOME/go-install/go/bin/go" ]; then
        GO="$HOME/go-install/go/bin/go"
    else
        echo "run-workbench-smoke: go toolchain not found" >&2
        exit 2
    fi
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SMOKE="$SCRIPT_DIR/workbench_e2e.go"

if [ ! -f "$SMOKE" ]; then
    echo "run-workbench-smoke: smoke program not found at $SMOKE (workbench absent)" >&2
    exit 1
fi

echo "==> run-workbench-smoke: building + running $SMOKE"
# Build first so a build failure is reported distinctly from a runtime failure,
# then run the built binary. Both go through the PUBLIC API; no workbench TUI
# build is required.
TMPBIN="$(mktemp -t workbench-smoke.XXXXXX 2>/dev/null || mktemp)"
trap 'rm -f "$TMPBIN"' EXIT

if ! "$GO" build -o "$TMPBIN" "$SMOKE"; then
    echo "run-workbench-smoke: build failed" >&2
    exit 1
fi

if ! "$TMPBIN"; then
    echo "run-workbench-smoke: smoke run failed (a step asserted failure)" >&2
    exit 1
fi

echo "==> run-workbench-smoke: PASS"
exit 0
