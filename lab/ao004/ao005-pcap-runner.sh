#!/usr/bin/env bash
# AO-005 isolated G41 raw-PCAP evidence runner (LAB-ONLY; runs INSIDE the pinned
# lab/ao004 container). Owner-authorized 2026-11-14: collect original raw packet
# captures for both opt-in G41 directions.
#
#   DIR=A : pinned OpenDNP3 3.1.2 C++ master -> Go outstation fixture
#   DIR=B : Go master harness -> pinned OpenDNP3 3.1.2 C++ outstation peer
#
# Each direction runs eight single-shot Direct-Operate-with-response cases:
# G41V1..V4 at index 0 (SUCCESS, exactly one host write) and index 1
# (OUT_OF_RANGE, zero host writes). Raw frames are captured on the master-side
# veth inside the isolated two-namespace topology (no default route, no
# forwarding, no external reachability).
set -u

DIR="${DIR:-A}"
OUT="${OUT:-/out}"
mkdir -p "$OUT"

# Save the runner's own stdout/stderr (case results, isolation lines) alongside
# the captures so the evidence bundle is self-contained.
exec > >(tee "$OUT/runner-$DIR.log") 2>&1

log() { printf '%s\n' "$*"; }
fail() { printf 'RESULT FAIL %s\n' "$*"; }

# --- 0. container isolation baseline (before the topology exists) --------------
ip -4 route > "$OUT/routes-before-$DIR.txt" 2>&1
if grep -q '^default' "$OUT/routes-before-$DIR.txt"; then
    fail "default route present in container baseline; refusing"; exit 2
fi
log "RESULT ISOLATION_BASELINE PASS (no default route before topology)"

# --- 1. private two-namespace veth topology ------------------------------------
/opt/vec01/lab-netns.sh up > "$OUT/netns-$DIR.log" 2>&1 || { fail "topology build"; exit 2; }
for ns in vec01-master vec01-outstation; do
    ip netns exec "$ns" ip route show > "$OUT/routes-$ns-$DIR.txt" 2>&1
    grep -q '^default' "$OUT/routes-$ns-$DIR.txt" && { fail "$ns has a default route"; exit 2; }
    fwd="$(ip netns exec "$ns" cat /proc/sys/net/ipv4/ip_forward)"
    [ "$fwd" = "0" ] || { fail "$ns ip_forward=$fwd"; exit 2; }
done
if ip netns exec vec01-master ping -c1 -W2 10.77.0.2 >/dev/null 2>&1; then
    log "RESULT REACH PASS (master reaches outstation over the internal veth)"
else
    fail "no internal veth reachability"; exit 2
fi

# --- 2. raw capture on the master-side veth ------------------------------------
ip netns exec vec01-master tcpdump -i veth-m -s 0 -U -w "$OUT/ao005-dir$DIR.pcap" \
    > "$OUT/tcpdump-$DIR.log" 2>&1 &
TCPDUMP_PID=$!
sleep 1

# --- 3. start the outstation side ----------------------------------------------
if [ "$DIR" = "A" ]; then
    ip netns exec vec01-outstation /opt/ao005/go-outstation --listen \
        > "$OUT/outstation-$DIR.log" 2>&1 &
else
    # The C++ peer blocks on stdin; hold it open with a long sleep.
    ip netns exec vec01-outstation bash -c 'sleep 600 | /opt/ao004/g41_peer --listen' \
        > "$OUT/outstation-$DIR.log" 2>&1 &
fi
SERVER_PID=$!

# Wait (bounded) for the outstation to report its listener before traffic.
for _ in $(seq 1 20); do
    grep -qi 'listening' "$OUT/outstation-$DIR.log" && break
    sleep 0.5
done
log "--- outstation readiness ($DIR) ---"
cat "$OUT/outstation-$DIR.log"

# --- 4. run the eight single-shot cases ----------------------------------------
run_case() {
    local v="$1" i="$2"
    if [ "$DIR" = "A" ]; then
        ip netns exec vec01-master /opt/ao004/g41_master --run \
            --variation "$v" --index "$i" --timeout 6 2>&1 | grep -E "completed="
    else
        ip netns exec vec01-master /opt/ao005/go-master \
            --variation "$v" --index "$i" --timeout 6s 2>&1
    fi
}

for i in 0 1; do
    for v in 1 2 3 4; do
        printf 'CASE direction=%s variation=%s index=%s :: ' "$DIR" "$v" "$i"
        run_case "$v" "$i"
        sleep 0.5
    done
done

# --- 5. stop capture and outstation --------------------------------------------
sleep 1
kill "$TCPDUMP_PID" 2>/dev/null || true
wait "$TCPDUMP_PID" 2>/dev/null || true
kill "$SERVER_PID" 2>/dev/null || true
wait "$SERVER_PID" 2>/dev/null || true

log "--- outstation accepted-command log ($DIR) ---"
cat "$OUT/outstation-$DIR.log"

if [ -s "$OUT/ao005-dir$DIR.pcap" ]; then
    log "RESULT PCAP_PRESENT $OUT/ao005-dir$DIR.pcap ($(stat -c%s "$OUT/ao005-dir$DIR.pcap") bytes)"
else
    fail "no PCAP produced"
fi
log "RESULT DIRECTION_$DIR DONE"
