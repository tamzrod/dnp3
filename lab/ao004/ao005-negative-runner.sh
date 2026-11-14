#!/usr/bin/env bash
# AO-005 negative-interop evidence runner (LAB-ONLY; runs INSIDE the pinned
# lab/ao004 container). Owner-authorized 2026-11-14: bounded negative
# interoperability iteration (malformed G41 requests + after-send fault
# injection), isolated lab only.
#
#   MATRIX=A : hand-crafted raw DNP3 frames (independent injector) -> Go
#              outstation fixture, over the isolated two-namespace topology.
#              Malformed G41 DirectOperate objects, plus positive controls and a
#              link-layer CRC-corrupt frame.
#   MATRIX=B : opted-in Go master harness -> pinned OpenDNP3 3.1.2 C++ peer
#              through a fault proxy that withholds the response (drop-response)
#              or disconnects right after the request is on the wire
#              (close-after-request). Proves after-send UNKNOWN / fail-closed /
#              no-blind-retry.
#
# Raw frames are captured on the master-side veth inside the isolated topology
# (no default route, no forwarding, no external reachability).
set -u

MATRIX="${MATRIX:-A}"
OUT="${OUT:-/out}"
mkdir -p "$OUT"

exec > >(tee "$OUT/runner-$MATRIX.log") 2>&1

log() { printf '%s\n' "$*"; }
fail() { printf 'RESULT FAIL %s\n' "$*"; }

# --- 0. container isolation baseline (before the topology exists) --------------
ip -4 route > "$OUT/routes-before-$MATRIX.txt" 2>&1
if grep -q '^default' "$OUT/routes-before-$MATRIX.txt"; then
    fail "default route present in container baseline; refusing"; exit 2
fi
log "RESULT ISOLATION_BASELINE PASS (no default route before topology)"

# --- 1. private two-namespace veth topology ------------------------------------
/opt/vec01/lab-netns.sh up > "$OUT/netns-$MATRIX.log" 2>&1 || { fail "topology build"; exit 2; }
for ns in vec01-master vec01-outstation; do
    ip netns exec "$ns" ip route show > "$OUT/routes-$ns-$MATRIX.txt" 2>&1
    grep -q '^default' "$OUT/routes-$ns-$MATRIX.txt" && { fail "$ns has a default route"; exit 2; }
    fwd="$(ip netns exec "$ns" cat /proc/sys/net/ipv4/ip_forward)"
    [ "$fwd" = "0" ] || { fail "$ns ip_forward=$fwd"; exit 2; }
done
if ip netns exec vec01-master ping -c1 -W2 10.77.0.2 >/dev/null 2>&1; then
    log "RESULT REACH PASS (master reaches outstation over the internal veth)"
else
    fail "no internal veth reachability"; exit 2
fi

# --- 2. raw capture on the master-side veth ------------------------------------
ip netns exec vec01-master tcpdump -i veth-m -s 0 -U -w "$OUT/ao005-neg-$MATRIX.pcap" \
    > "$OUT/tcpdump-$MATRIX.log" 2>&1 &
TCPDUMP_PID=$!
sleep 1

run_matrix_a() {
    ip netns exec vec01-outstation /opt/ao005/go-outstation --listen \
        > "$OUT/outstation-$MATRIX.log" 2>&1 &
    SERVER_PID=$!
    for _ in $(seq 1 20); do
        grep -qi 'listening' "$OUT/outstation-$MATRIX.log" && break
        sleep 0.5
    done
    log "--- outstation readiness ($MATRIX) ---"
    cat "$OUT/outstation-$MATRIX.log"

    # positive controls, out-of-range index, then malformed objects; finally a
    # link-layer CRC-corrupt frame.
    cases="pos-v1 pos-v2 pos-v3 pos-v4 idx1-v1 idx1-v4 trunc-v1 trunc-v2 trunc-v3 trunc-v4 long-v1 badqual-17 var0 var5 count2 count0 idx2 fc-write"
    for c in $cases; do
        printf 'CASE matrix=A case=%s :: ' "$c"
        ip netns exec vec01-master /opt/ao005/injector --case "$c" --read-wait 1.5s 2>&1 \
            | grep -E 'request_hex|response_hex|response=NONE' | tr '\n' ' '
        printf '\n'
        sleep 0.3
    done
    printf 'CASE matrix=A case=crc-corrupt :: '
    ip netns exec vec01-master /opt/ao005/injector --case pos-v1 --corrupt-header-crc --read-wait 1.5s 2>&1 \
        | grep -E 'request_hex|response_hex|response=NONE' | tr '\n' ' '
    printf '\n'
    sleep 1
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
    log "--- outstation accepted-command log ($MATRIX) ---"
    cat "$OUT/outstation-$MATRIX.log"
}

run_matrix_b() {
    # C++ peer blocks on stdin; hold it open with a long sleep.
    ip netns exec vec01-outstation bash -c 'sleep 600 | /opt/ao004/g41_peer --listen' \
        > "$OUT/peer-$MATRIX.log" 2>&1 &
    SERVER_PID=$!
    for _ in $(seq 1 20); do
        grep -qi 'listening' "$OUT/peer-$MATRIX.log" && break
        sleep 0.5
    done
    log "--- peer readiness ($MATRIX) ---"
    cat "$OUT/peer-$MATRIX.log"

    # presend scenario without a proxy: Operate before Connect.
    printf 'CASE matrix=B case=presend-notconn :: '
    ip netns exec vec01-master /opt/ao005/neg-master --mode presend-notconn 2>&1 | grep -E 'neg-master:' | tr '\n' ' '
    printf '\n'
    sleep 0.5

    # positive control: normal pass-through of the fault proxy (control applied
    # once, success relayed) — proves the proxy itself is transparent.
    ip netns exec vec01-master /opt/ao005/faultproxy --mode pass > "$OUT/proxy-pass.log" 2>&1 &
    PROXY_PID=$!
    sleep 1
    printf 'CASE matrix=B case=pass :: '
    ip netns exec vec01-master /opt/ao005/neg-master --mode pass 2>&1 | grep -E 'neg-master:' | tr '\n' ' '
    printf '\n'
    sleep 1
    kill "$PROXY_PID" 2>/dev/null || true
    wait "$PROXY_PID" 2>/dev/null || true
    sleep 0.5

    # presend-cancel through the proxy: a definite pre-send failure must NOT
    # quarantine; the follow-up Operate must be admitted and applied once.
    ip netns exec vec01-master /opt/ao005/faultproxy --mode pass > "$OUT/proxy-presend-cancel.log" 2>&1 &
    PROXY_PID=$!
    sleep 1
    printf 'CASE matrix=B case=presend-cancel :: '
    ip netns exec vec01-master /opt/ao005/neg-master --mode presend-cancel 2>&1 | grep -E 'neg-master:' | tr '\n' ' '
    printf '\n'
    sleep 1
    kill "$PROXY_PID" 2>/dev/null || true
    wait "$PROXY_PID" 2>/dev/null || true
    sleep 0.5

    # after-send scenarios via the fault proxy.
    for fm in drop-response close-after-request; do
        ip netns exec vec01-master /opt/ao005/faultproxy --mode "$fm" > "$OUT/proxy-$fm.log" 2>&1 &
        PROXY_PID=$!
        sleep 1
        printf 'CASE matrix=B case=after-send-%s :: ' "$fm"
        ip netns exec vec01-master /opt/ao005/neg-master --mode after-send 2>&1 | grep -E 'neg-master:' | tr '\n' ' '
        printf '\n'
        sleep 1
        kill "$PROXY_PID" 2>/dev/null || true
        wait "$PROXY_PID" 2>/dev/null || true
        sleep 0.5
    done
    sleep 1
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
    log "--- peer accepted-command log ($MATRIX) ---"
    cat "$OUT/peer-$MATRIX.log"
}

if [ "$MATRIX" = "A" ]; then
    run_matrix_a
else
    run_matrix_b
fi

sleep 1
kill "$TCPDUMP_PID" 2>/dev/null || true
wait "$TCPDUMP_PID" 2>/dev/null || true

if [ -s "$OUT/ao005-neg-$MATRIX.pcap" ]; then
    log "RESULT PCAP_PRESENT $OUT/ao005-neg-$MATRIX.pcap ($(stat -c%s "$OUT/ao005-neg-$MATRIX.pcap") bytes)"
else
    fail "no PCAP produced"
fi
log "RESULT MATRIX_$MATRIX DONE"
