#!/usr/bin/env bash
#
# scripts/verify-external-mvp.sh — DNP3 external-MVP verification gate (MEXT-021)
#
# Two-tier gate for the external interoperability claim (R4 / VEC-01):
#
#   Tier 1 — internal real-TCP: build + the real-TCP loopback tests that use a
#            real TCP transport and real DNP3 wire framing against the in-repo
#            outstation. This tier can pass today and is the foundation for the
#            external claim.
#
#   Tier 2 — external / third-party (VEC-01): fail-closed. Looks for genuine
#            external interop proof — a non-placeholder capture fixture under
#            active_work/testdata/external/ and/or external interop tests. Until
#            such proof lands (MEXT-022 real-TCP full MVP path, MEXT-033
#            third-party stack capture), this tier refuses to pass so the
#            external interop claim CANNOT be made prematurely.
#
# Exit 0 only when BOTH tiers pass. Tier 2 failing-closed yields a non-zero exit
# by design.
#
# Environment:
#   ALLOW_NO_EXTERNAL=1  skip the fail-closed Tier 2 check (run Tier 1 only).
#                        Intended for environments that only exercise the
#                        internal real-TCP tier; does NOT satisfy the external
#                        claim.
#
# This is a skeleton (MEXT-021). Real external tests/fixtures are wired in by
# MEXT-022 and MEXT-033.

set -u

# Locate the Go toolchain: prefer `go` on PATH, fall back to the per-user
# install used in this environment.
GO="$(command -v go || true)"
if [ -z "$GO" ]; then
    if [ -x "$HOME/go-install/go/bin/go" ]; then
        GO="$HOME/go-install/go/bin/go"
    else
        echo "verify-external-mvp: go toolchain not found" >&2
        exit 1
    fi
fi
export GO

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

EXT_FIXTURE_DIR="active_work/testdata/external"

fail() {
    echo "verify-external-mvp: FAIL — $*" >&2
    exit 1
}

step() {
    printf '\n\033[1m==> verify-external-mvp: %s\033[0m\n' "$*"
}

# -----------------------------------------------------------------------------
# Tier 1 — internal real-TCP
# -----------------------------------------------------------------------------

step "Tier 1: go build ./..."
"$GO" build ./... || fail "build failed"

# Real-TCP loopback tests: real TCP transport + real DNP3 wire framing against
# the in-repo outstation. These are the closest internal analogue to external
# interop and must stay green as the foundation for the external claim.
REAL_TCP_TARGETS=(
    "./test/integration=TestTCPMasterOutstationRead|TestTCPDirectCommunication|TestMasterOutstationEndToEndComprehensive"
    "./test/integration=TestOperateRealTCPSuccess|TestOperateRealTCPBlockedStatus|TestOperateStatusMatrixOnTCP|TestOperateStatusMatrixOnTCPDrop"
    "./test/integration=TestReconnectOnTCPNoStuckState|TestDeviceRestartNotRaisableOnV0Outstation"
    "./test/integration=TestPublicMVPLoopbackFullLifecycle|TestPublicMVPLoopbackOperateStatus|TestPublicMVPLoopbackErrorClassification"
    "./test/integration=TestRealTCPFullMVPPath"
)

step "Tier 1: real-TCP loopback tests"
for entry in "${REAL_TCP_TARGETS[@]}"; do
    pkg="${entry%%=*}"
    pats="${entry#*=}"
    # Convert the a|b|c list into a single -run regex (^a$|^b$|^c$).
    regex=""
    IFS='|' read -ra _parts <<< "$pats"
    first=1
    for p in "${_parts[@]}"; do
        if [ "$first" -eq 1 ]; then
            regex="^${p}$"
            first=0
        else
            regex="${regex}|^${p}$"
        fi
    done
    "$GO" test -count=1 -run "$regex" "$pkg" \
        || fail "Tier 1 real-TCP tests failed in $pkg"
done

# -----------------------------------------------------------------------------
# Tier 2 — external / third-party (VEC-01) — fail-closed
# -----------------------------------------------------------------------------

step "Tier 2: external / third-party (VEC-01) proof"

# A genuine external proof is a non-placeholder .vec capture fixture (one that
# does NOT contain "PLACEHOLDER" in its notes/title) and/or an external interop
# test. Until MEXT-022/MEXT-033 land such proof, this tier fails closed.
has_external_fixture=0
if [ -d "$EXT_FIXTURE_DIR" ]; then
    # Any .vec whose content does not mention PLACEHOLDER counts as real.
    while IFS= read -r -d '' f; do
        if ! grep -qi "PLACEHOLDER" "$f"; then
            has_external_fixture=1
            break
        fi
    done < <(find "$EXT_FIXTURE_DIR" -type f -name '*.vec' -print0)
fi

# TODO(MEXT-022): add an external interop test selector here once a real
# third-party / capture-replay test exists (e.g. a build-tagged
# test/integration/external_*_test.go). Until then has_external_test stays 0.
has_external_test=0

if [ "$has_external_fixture" -eq 1 ] || [ "$has_external_test" -eq 1 ]; then
    # Real external proof present — run the external interop tests.
    # TODO(MEXT-022/MEXT-033): wire the actual external test command here.
    step "Tier 2: external interop tests"
    "$GO" test -count=1 -run "External" ./test/integration/... \
        || fail "external interop tests failed"
    step "OK (external)"
    echo "verify-external-mvp: both tiers passed."
    exit 0
fi

# No external proof yet — fail closed unless explicitly skipped.
if [ "${ALLOW_NO_EXTERNAL:-0}" = "1" ]; then
    step "Tier 2: SKIPPED (ALLOW_NO_EXTERNAL=1) — internal real-TCP tier only"
    echo "verify-external-mvp: Tier 1 passed; Tier 2 skipped (external claim NOT satisfied)."
    exit 0
fi

fail "no external interop proof yet (MEXT-022/MEXT-033). Tier 2 is fail-closed; set ALLOW_NO_EXTERNAL=1 to run Tier 1 only."
