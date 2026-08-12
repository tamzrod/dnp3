#!/usr/bin/env bash
#
# scripts/run-conformance.sh — DNP3 layer conformance gate (DNP3-098)
#
# Runs the existing DLL / TL / AL conformance suites against the current code
# base as a single, CI-runnable entry point. Exit 0 only when every suite is
# green (plain + race); non-zero on any failure.
#
# These suites are also covered by scripts/verify-mvp.sh; this script provides
# a focused, independently-runnable gate for the layered conformance vectors.
#
# Layers exercised:
#   - DLL (test/conformance/dll): CRC-16, frame layout, link control frames
#   - TL  (test/conformance/tl):  transport header encode/decode, reassembly
#   - AL  (test/conformance/al):  application control byte, sequence range
#
# Usage:
#   scripts/run-conformance.sh            # plain + race
#   scripts/run-conformance.sh -race      # race only
#   scripts/run-conformance.sh -plain     # plain only

set -u

GO="$(command -v go || true)"
if [ -z "$GO" ]; then
    if [ -x "$HOME/go-install/go/bin/go" ]; then
        GO="$HOME/go-install/go/bin/go"
    else
        echo "run-conformance: go toolchain not found" >&2
        exit 1
    fi
fi
export GO

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

CONF_PKGS=(
    ./test/conformance/dll
    ./test/conformance/tl
    ./test/conformance/al
)

RUN_PLAIN=1
RUN_RACE=1
for arg in "$@"; do
    case "$arg" in
        -plain) RUN_RACE=0 ;;
        -race)  RUN_PLAIN=0 ;;
        *) echo "run-conformance: unknown option '$arg'" >&2; exit 2 ;;
    esac
done

fail() {
    echo "run-conformance: FAIL — $*" >&2
    exit 1
}

step() {
    printf '\n\033[1m==> run-conformance: %s\033[0m\n' "$*"
}

step "go build ./... (conformance prerequisites)"
"$GO" build ./... || fail "build failed"

if [ "$RUN_PLAIN" -eq 1 ]; then
    step "conformance suites (plain)"
    for pkg in "${CONF_PKGS[@]}"; do
        "$GO" test -count=1 "$pkg" || fail "$pkg plain tests failed"
        printf '  %-30s OK\n' "$pkg"
    done
fi

if [ "$RUN_RACE" -eq 1 ]; then
    step "conformance suites (-race)"
    for pkg in "${CONF_PKGS[@]}"; do
        "$GO" test -race -count=1 "$pkg" || fail "$pkg race tests failed"
        printf '  %-30s OK (race)\n' "$pkg"
    done
fi

echo
echo "run-conformance: all conformance suites passed."
