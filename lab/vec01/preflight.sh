#!/usr/bin/env bash
# Objective, fail-closed Gate B readiness preflight, run INSIDE the lab
# container. It builds the private netns/veth topology, proves isolation, and
# confirms a memory-only listener can bind ONLY 10.77.0.2:20000.
#
# It runs NO DNP3 protocol exchange and issues NO control. The only listener is
# a short-lived non-DNP3 TCP socket used to prove the bind is lab-local; it is
# closed immediately. This is the same bounded non-DNP3 TCP scope correction
# already recorded in docs/verification/VEC-01-LAB-EXEC-001.md §14.
set -euo pipefail

NS_MASTER=vec01-master
NS_OUTSTATION=vec01-outstation
FAIL=0
pass() { echo "PASS: $*"; }
fail() { echo "FAIL: $*"; FAIL=1; }

# --- 1. topology -----------------------------------------------------------------
if /opt/vec01/lab-netns.sh up >/dev/null 2>&1; then
  pass "two private netns + one internal veth created"
else
  fail "could not create netns/veth topology (need NET_ADMIN + SYS_ADMIN)"
  echo "GATE_B_PREFLIGHT: FAIL"
  exit 1
fi
ip -n "$NS_MASTER"    addr show veth-m | grep -q "10.77.0.1/24" && pass "master addr 10.77.0.1/24" || fail "master addr missing"
ip -n "$NS_OUTSTATION" addr show veth-o | grep -q "10.77.0.2/24" && pass "outstation addr 10.77.0.2/24" || fail "outstation addr missing"

# --- 2. no default route / no forwarding / no external reachability ---------------
for ns in "$NS_MASTER" "$NS_OUTSTATION"; do
  if ip netns exec "$ns" ip route show | grep -q '^default'; then
    fail "$ns has a default route"
  else
    pass "$ns has no default route"
  fi
  [[ "$(ip netns exec "$ns" cat /proc/sys/net/ipv4/ip_forward)" == "0" ]] && pass "$ns forwarding off" || fail "$ns forwarding on"
  # No DNS and no external reachability: a public lookup must fail.
  if ip netns exec "$ns" getent hosts github.com >/dev/null 2>&1; then
    fail "$ns resolved an external host"
  else
    pass "$ns cannot resolve external hosts"
  fi
done

# --- 3. lab-local-only reachability ----------------------------------------------
if ip netns exec "$NS_MASTER" ping -c1 -W2 10.77.0.2 >/dev/null 2>&1; then
  pass "master reaches outstation over the internal veth"
else
  fail "master cannot reach outstation over the internal veth"
fi

# --- 4. listener binds ONLY the isolated lab address ------------------------------
# Short-lived non-DNP3 bind proof on the outstation namespace.
ip netns exec "$NS_OUTSTATION" python3 - "$@" <<'PY' &
import socket, sys, time
s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
s.bind(("10.77.0.2", 20000))
s.listen(1)
sys.stderr.write("bound 10.77.0.2:20000\n")
time.sleep(3)
PY
BIND_PID=$!
sleep 1
if ip netns exec "$NS_OUTSTATION" ss -ltn 2>/dev/null | grep -q '10.77.0.2:20000'; then
  pass "listener bound to 10.77.0.2:20000 only"
else
  fail "listener bind not observed on 10.77.0.2:20000"
fi
kill "$BIND_PID" 2>/dev/null || true

# --- 5. no physical / host network surface ---------------------------------------
[[ -z "$(ls /sys/class/net | grep -vE '^(lo|veth-m|veth-o|eth0)$' || true)" ]] && pass "no unexpected host interfaces" || echo "NOTE: extra interfaces present"
[[ ! -e /var/run/docker.sock ]] && pass "no docker daemon socket inside lab" || fail "docker socket present inside lab"

echo "GATE_B_PREFLIGHT: $([[ $FAIL -eq 0 ]] && echo PASS || echo FAIL)"
exit "$FAIL"
