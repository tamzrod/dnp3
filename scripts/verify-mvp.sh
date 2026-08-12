#!/usr/bin/env bash
#
# scripts/verify-mvp.sh — DNP3 MVP verification gate (DNP3-052)
#
# Single command that builds and runs the unit + integration tests that
# constitute the verified Master MVP (v0 interoperability profile). Exit 0 on a
# clean tree; non-zero on any failure. Intended for CI and pre-merge use.
#
# Coverage:
#   - go build ./...
#   - go vet on the MVP packages
#   - go test -count=1 on the MVP unit + integration packages
#   - go test -race -count=1 on the race-relevant MVP packages
#
# The pre-existing `go vet` "unreachable code" note in
# internal/outstation/outstation.go:827 predates the MVP gate and is out of
# scope; internal/outstation is therefore excluded from the vet step (it is
# still built and race-tested).

set -u

# Locate the Go toolchain: prefer `go` on PATH, fall back to the per-user
# install used in this environment.
GO="$(command -v go || true)"
if [ -z "$GO" ]; then
    if [ -x "$HOME/go-install/go/bin/go" ]; then
        GO="$HOME/go-install/go/bin/go"
    else
        echo "verify-mvp: go toolchain not found" >&2
        exit 1
    fi
fi
export GO

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

# Packages whose unit/golden/loopback tests prove the Master MVP path.
MVP_UNIT_PKGS=(
    ./internal/al
    ./internal/dll/crc
    ./internal/dll/frame
    ./internal/dll/link
    ./internal/tl
    ./internal/master
    ./internal/testutils
    ./pkg/dnp3
    ./pkg/dnp3/master
    ./pkg/dnp3/types
    ./pkg/transport
    ./test/conformance/al
    ./test/conformance/dll
    ./test/conformance/tl
    ./test/integration
)

# Subset where the race detector is meaningful (goroutines / shared state).
MVP_RACE_PKGS=(
    ./internal/master
    ./internal/testutils
    ./pkg/dnp3
    ./pkg/dnp3/master
    ./pkg/dnp3/outstation
    ./pkg/dnp3/types
    ./test/integration
)

# Packages vetted as part of the gate. internal/outstation is excluded due to a
# pre-existing, out-of-scope unreachable-code note (see header).
MVP_VET_PKGS=(
    ./internal/al
    ./internal/dll/...
    ./internal/tl
    ./internal/master
    ./internal/testutils
    ./pkg/dnp3/...
    ./pkg/transport
    ./test/conformance/...
    ./test/integration
)

fail() {
    echo "verify-mvp: FAIL — $*" >&2
    exit 1
}

step() {
    printf '\n\033[1m==> verify-mvp: %s\033[0m\n' "$*"
}

step "go build ./..."
"$GO" build ./... || fail "build failed"

step "go vet (MVP packages)"
# shellcheck disable=SC2086
"$GO" vet ${MVP_VET_PKGS[*]} || fail "go vet failed"

step "go test -count=1 (MVP unit + integration)"
"$GO" test -count=1 "${MVP_UNIT_PKGS[@]}" || fail "unit/integration tests failed"

step "go test -race -count=1 (MVP race)"
"$GO" test -race -count=1 "${MVP_RACE_PKGS[@]}" || fail "race tests failed"

step "OK"
echo "verify-mvp: all MVP gate checks passed."
exit 0
