#!/usr/bin/env bash
#
# scripts/verify-external-mvp.sh - DNP3 external-MVP verification gate.
#
# Single external-claim command (MEXT-033). Runs the v0 external-MVP gate and
# exits 0 only when every required tier is green on a clean tree:
#
#   Tier 1 - real-TCP MVP path: build + the real-TCP loopback tests that use a
#            real TCP transport and real DNP3 wire framing against the in-repo
#            outstation (Connect -> IntegrityPoll -> Operate -> Disconnect),
#            plus the reconnect/no-stuck-state and rogue-peer negative tests.
#
#   Tier 2 - operate matrix: the Direct-Operate success/fail/drop matrix over
#            real TCP (no false success, no false timeout on a complete APDU).
#
#   Tier 3 - multi-header: the Class-0 single multi-object-header response
#            parse + per-group fallback (G1+G20+G30 in one APDU, no point loss).
#
# The gate models the external v0 claim: the v0 Master path is verified over
# real TCP with real DNP3 wire framing. It is NOT a full IEEE 1815 conformance
# claim, and it is NOT a third-party-interop claim (see the advisory tier
# below). MEXT-034 records the README wording; MEXT-035 records acceptance.
#
# Advisory (non-blocking) - VEC-01 third-party capture:
#   A genuine non-placeholder capture fixture under
#   active_work/testdata/external/*.vec and/or an external interop test would
#   strengthen the claim to true third-party interop. Until such proof lands,
#   this advisory tier reports "no third-party capture" but does NOT fail the
#   gate (the real-TCP tiers above are the MEXT-033 bar). Set
#   REQUIRE_EXTERNAL=1 to make the advisory tier blocking.
#
# Environment:
#   REQUIRE_EXTERNAL=1  make the VEC-01 advisory tier blocking (fail the gate
#                       unless a non-placeholder external capture/test exists).
#
# Exit 0 only when all required tiers (1-3) pass (and the advisory tier when
# REQUIRE_EXTERNAL=1).
#
# Evolution: MEXT-021 skeleton -> MEXT-022 real-TCP MVP -> MEXT-024 operate
# matrix -> MEXT-026 rogue-peer negatives -> MEXT-033 full gate (this).

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
    echo "verify-external-mvp: FAIL - $*" >&2
    exit 1
}

step() {
    printf '\n\033[1m==> verify-external-mvp: %s\033[0m\n' "$*"
}

# run_named_tests <pkg> <TestA|TestB|TestC>  - run the named tests (exact match)
# in pkg, fail the gate on any failure.
run_named_tests() {
    local pkg="$1" pats="$2"
    local regex="" first=1 p
    IFS='|' read -ra _parts <<< "$pats"
    for p in "${_parts[@]}"; do
        if [ "$first" -eq 1 ]; then
            regex="^${p}$"
            first=0
        else
            regex="${regex}|^${p}$"
        fi
    done
    "$GO" test -count=1 -run "$regex" "$pkg" \
        || fail "tests failed in $pkg ($pats)"
}

# -----------------------------------------------------------------------------
# Build
# -----------------------------------------------------------------------------

step "build: go build ./..."
"$GO" build ./... || fail "build failed"

# -----------------------------------------------------------------------------
# Tier 1 - real-TCP MVP path
# -----------------------------------------------------------------------------

step "Tier 1: real-TCP MVP path (Connect -> IntegrityPoll -> Operate -> Disconnect)"
run_named_tests "./test/integration" \
    "TestTCPMasterOutstationRead|TestTCPDirectCommunication|TestMasterOutstationEndToEndComprehensive|TestRealTCPFullMVPPath|TestPublicMVPLoopbackFullLifecycle|TestPublicMVPLoopbackOperateStatus|TestPublicMVPLoopbackErrorClassification"
# Reconnect/no-stuck-state + rogue-peer negatives are part of the real-TCP
# robustness foundation for the external claim.
run_named_tests "./test/integration" \
    "TestReconnectOnTCPNoStuckState|TestDeviceRestartNotRaisableOnV0Outstation|TestRoguePeerWrongAddressNoHang|TestRoguePeerBadCRCNoHang"

# -----------------------------------------------------------------------------
# Tier 2 - operate matrix
# -----------------------------------------------------------------------------

step "Tier 2: operate matrix (Direct-Operate success / fail / drop over TCP)"
run_named_tests "./test/integration" \
    "TestOperateRealTCPSuccess|TestOperateRealTCPBlockedStatus|TestOperateStatusMatrixOnTCP|TestOperateStatusMatrixOnTCPDrop"

# -----------------------------------------------------------------------------
# Tier 3 - multi-header Class-0 parse
# -----------------------------------------------------------------------------

step "Tier 3: multi-header Class-0 parse + per-group fallback"
run_named_tests "./pkg/dnp3/master" \
    "TestReadMultiHeaderReturnsAllGroups|TestIntegrityPollSingleMultiHeaderExchange|TestIntegrityPollFallbackPerGroup"

# -----------------------------------------------------------------------------
# Advisory (non-blocking) - VEC-01 third-party capture
# -----------------------------------------------------------------------------

step "Advisory: VEC-01 third-party capture proof"

has_external_fixture=0
if [ -d "$EXT_FIXTURE_DIR" ]; then
    while IFS= read -r -d '' f; do
        if ! grep -qi "PLACEHOLDER" "$f"; then
            has_external_fixture=1
            break
        fi
    done < <(find "$EXT_FIXTURE_DIR" -type f -name '*.vec' -print0)
fi

# An external interop test selector (build-tagged test/integration/external_*_test.go)
# would also count. None exists yet.
has_external_test=0

if [ "$has_external_fixture" -eq 1 ] || [ "$has_external_test" -eq 1 ]; then
    step "Advisory: external interop tests"
    "$GO" test -count=1 -run "External" ./test/integration/... \
        || fail "external interop tests failed"
    echo "verify-external-mvp: advisory - third-party capture proof present."
elif [ "${REQUIRE_EXTERNAL:-0}" = "1" ]; then
    fail "REQUIRE_EXTERNAL=1 but no third-party capture proof found (active_work/testdata/external/*.vec or external interop test)."
else
    echo "verify-external-mvp: advisory - no third-party capture proof (not required by MEXT-033; the v0 claim is real-TCP + wire-framing, not third-party interop)."
fi

# -----------------------------------------------------------------------------
# Result
# -----------------------------------------------------------------------------

step "OK"
echo "verify-external-mvp: all required tiers (1 real-TCP MVP, 2 operate matrix, 3 multi-header) passed."
exit 0
