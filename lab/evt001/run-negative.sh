#!/bin/bash
# EVT-001 negative interoperability run.
# Duplicate application confirm, then withheld confirm with bounded retry.
# Intended for a container with --network none.
set -u
OUT=${OUT:-/out}
BIN=${BIN:-/opt/evt/bin}
MASTER=${MASTER:-/opt/evt/evt_master}
PROXY=${PROXY:-/opt/evt/confirm_proxy.py}
CHECK=${CHECK:-/opt/evt/check_negative.py}
mkdir -p "$OUT"
exec > >(tee "$OUT/run.log") 2>&1

fail=0
note() { printf 'RESULT %s\n' "$*"; }

echo "=== isolation ==="
ip -4 route > "$OUT/routes.txt" || true
cat "$OUT/routes.txt"
if grep -q '^default' "$OUT/routes.txt"; then
  note "ISOLATION FAIL default route present"
  exit 2
fi
if python3 - << 'PY'
import socket, sys
s = socket.socket()
s.settimeout(2)
try:
    s.connect(("1.1.1.1", 443))
except Exception:
    sys.exit(0)
sys.exit(1)
PY
then
  note "ISOLATION PASS no default route and external connect failed"
else
  note "ISOLATION FAIL external connect succeeded"
  exit 2
fi

GO_PID=
PROXY_PID=
TCPDUMP_PID=
MASTER_PID=
exec 3>&- 2>/dev/null || true

cleanup() {
  if [[ -n "${TCPDUMP_PID:-}" ]]; then
    kill "$TCPDUMP_PID" 2>/dev/null || true
    wait "$TCPDUMP_PID" 2>/dev/null || true
  fi
  if [[ -n "${MASTER_PID:-}" ]]; then kill "$MASTER_PID" 2>/dev/null || true; fi
  if [[ -n "${PROXY_PID:-}" ]]; then kill "$PROXY_PID" 2>/dev/null || true; fi
  if [[ -n "${GO_PID:-}" ]]; then kill "$GO_PID" 2>/dev/null || true; fi
}
trap cleanup EXIT

wait_file() {
  local file=$1 needle=$2 tries=$3
  local i
  for ((i = 0; i < tries; i++)); do
    if [[ -f "$file" ]] && grep -q "$needle" "$file"; then
      return 0
    fi
    sleep 0.1
  done
  return 1
}

count_line() {
  local file=$1 line=$2
  grep -F -c "$line" "$file" 2>/dev/null || true
}

start_go() {
  rm -f /tmp/go.fifo
  mkfifo /tmp/go.fifo
  exec 3<>/tmp/go.fifo
  "$BIN/go-outstation" 20000 < /tmp/go.fifo > "$OUT/go-outstation.stdout" 2>"$OUT/go-outstation.stderr" &
  GO_PID=$!
  if ! wait_file "$OUT/go-outstation.stdout" READY 50; then
    note "GO_OUTSTATION FAIL no READY"
    return 1
  fi
  note "GO_OUTSTATION READY"
}

stop_go() {
  if [[ -n "${GO_PID:-}" ]]; then
    echo quit >&3 2>/dev/null || true
    wait "$GO_PID" 2>/dev/null || true
    GO_PID=
  fi
  exec 3>&- 2>/dev/null || true
  exec 3<&- 2>/dev/null || true
}

go_cmd() {
  local before now i
  before=$(wc -l < "$OUT/go-outstation.stdout")
  echo "$*" >&3
  for ((i = 0; i < 30; i++)); do
    now=$(wc -l < "$OUT/go-outstation.stdout")
    if [[ "$now" -gt "$before" ]]; then
      return 0
    fi
    sleep 0.05
  done
  note "GO_CMD FAIL $*"
  fail=1
  return 1
}

start_capture() {
  local file=$1
  tcpdump -i lo -s 0 -U -w "$file" >/tmp/tcpdump.err 2>&1 &
  TCPDUMP_PID=$!
  sleep 0.4
  if ! kill -0 "$TCPDUMP_PID" 2>/dev/null; then
    note "PCAP FAIL tcpdump did not stay up"
    cat /tmp/tcpdump.err || true
    fail=1
    return 1
  fi
  note "PCAP capture $file"
}

stop_capture() {
  if [[ -n "${TCPDUMP_PID:-}" ]]; then
    sleep 0.3
    kill "$TCPDUMP_PID" 2>/dev/null || true
    wait "$TCPDUMP_PID" 2>/dev/null || true
    TCPDUMP_PID=
  fi
}

echo "=== A duplicate application confirm ==="
rm -f "$OUT/go-outstation.stdout" "$OUT/a-unsol.log" "$OUT/proxy-duplicate.log"
if start_go; then
  start_capture "$OUT/evt001-duplicate.pcap" || true
  python3 "$PROXY" 20010 20000 duplicate "$OUT/proxy-duplicate.log" &
  PROXY_PID=$!
  if ! wait_file "$OUT/proxy-duplicate.log" "PROXY LISTENING" 50; then
    note "A_PROXY FAIL not listening"
    fail=1
  else
    "$MASTER" unsol 20010 8 "$OUT/a-unsol.log" > "$OUT/a-unsol.stdout" 2>"$OUT/a-unsol.stderr" &
    MASTER_PID=$!
    if wait_file "$OUT/a-unsol.log" "STATE OPEN" 40; then
      sleep 0.8
      go_cmd bin 9 0 1 2022-03-04T05:06:07.008Z || true
      if wait_file "$OUT/proxy-duplicate.log" HELD_CONFIRM 40; then
        go_cmd bin 13 1 1 2021-06-07T08:09:10Z || true
        touch /tmp/evt-proxy-release
        if ! wait_file "$OUT/proxy-duplicate.log" DUP_SENT 30; then
          note "A_DUP_SENT FAIL"
          fail=1
        fi
      else
        note "A_HELD_CONFIRM FAIL"
        fail=1
      fi
    else
      note "A_UNSOL open FAIL"
      fail=1
    fi
    wait "$MASTER_PID" || { note "A_UNSOL process FAIL"; fail=1; }
    MASTER_PID=
    wait "$PROXY_PID" 2>/dev/null || true
    PROXY_PID=
    sleep 0.4
    # The proxy held the only outstation connection. Follow-up reads dial
    # the outstation directly so they are not injected a second time.
    if [[ "$(count_line "$OUT/a-unsol.log" "BIN idx=9 value=0 flags=01 time=699685567008")" -eq 1 ]]; then
      note "A_FIRST_EVENT PASS"
    else
      note "A_FIRST_EVENT FAIL"
      fail=1
    fi
    if [[ "$(count_line "$OUT/a-unsol.log" "BIN idx=13 value=1 flags=81 time=676368550000")" -eq 1 ]]; then
      note "A_SECOND_EVENT PASS"
    else
      note "A_SECOND_EVENT FAIL"
      fail=1
    fi
    if grep -F -q "unsol=1" "$OUT/a-unsol.log"; then
      note "A_UNSOL_FLAG PASS"
    else
      note "A_UNSOL_FLAG FAIL"
      fail=1
    fi
    if "$MASTER" solicit 20000 4 "$OUT/a-gone.log" > "$OUT/a-gone.stdout" 2>"$OUT/a-gone.stderr"; then
      if grep -F -q "BIN idx=9" "$OUT/a-gone.log" || grep -F -q "BIN idx=13" "$OUT/a-gone.log"; then
        note "A_DUPLICATE_DOES_NOT_KEEP FAIL an acknowledged event remained"
        fail=1
      else
        note "A_BOTH_REMOVED_ONCE PASS"
      fi
    else
      note "A_BOTH_REMOVED_ONCE FAIL process"
      fail=1
    fi
    go_cmd bin 14 1 1 2021-06-07T08:09:10Z || true
    if "$MASTER" solicit 20000 4 "$OUT/a-operational.log" > "$OUT/a-operational.stdout" 2>"$OUT/a-operational.stderr"; then
      if grep -F -q "BIN idx=14 value=1 flags=81 time=676368550000" "$OUT/a-operational.log"; then
        note "A_OPERATIONAL PASS"
      else
        note "A_OPERATIONAL FAIL"
        fail=1
      fi
    else
      note "A_OPERATIONAL FAIL process"
      fail=1
    fi
  fi
  stop_capture
  stop_go
else
  note "A_DIRECTION FAIL outstation did not start"
  fail=1
fi

echo "=== B withheld application confirm ==="
rm -f "$OUT/go-outstation.stdout" "$OUT/b-unsol.log" "$OUT/proxy-withhold.log" /tmp/evt-proxy-pass /tmp/evt-proxy-release
if start_go; then
  start_capture "$OUT/evt001-withhold.pcap" || true
  python3 "$PROXY" 20010 20000 withhold "$OUT/proxy-withhold.log" &
  PROXY_PID=$!
  if ! wait_file "$OUT/proxy-withhold.log" "PROXY LISTENING" 50; then
    note "B_PROXY FAIL not listening"
    fail=1
  else
    "$MASTER" unsol 20010 12 "$OUT/b-unsol.log" > "$OUT/b-unsol.stdout" 2>"$OUT/b-unsol.stderr" &
    MASTER_PID=$!
    if wait_file "$OUT/b-unsol.log" "STATE OPEN" 40; then
      sleep 0.8
      go_cmd bin 15 0 1 2022-03-04T05:06:07.008Z || true
    else
      note "B_UNSOL open FAIL"
      fail=1
    fi
    wait "$MASTER_PID" || { note "B_UNSOL process FAIL"; fail=1; }
    MASTER_PID=
    # The master has exited. Give the proxy a moment to record PROXY_END,
    # then close it so the outstation accepts the recovery master.
    sleep 0.5
    if ! grep -q '^PROXY_END ' "$OUT/proxy-withhold.log"; then
      kill "$PROXY_PID" 2>/dev/null || true
    fi
    wait "$PROXY_PID" 2>/dev/null || true
    PROXY_PID=
    sleep 0.4
    unsol_n=$(grep -c '^UNSOL ' "$OUT/proxy-withhold.log" || true)
    if [[ "$unsol_n" -eq 3 ]]; then
      note "B_RETRY_COUNT PASS"
    else
      note "B_RETRY_COUNT FAIL unsol=$unsol_n"
      fail=1
    fi
    shas=$(awk '/^UNSOL / { print $5 }' "$OUT/proxy-withhold.log" | sort -u | wc -l)
    if [[ "$shas" -eq 1 && "$unsol_n" -eq 3 ]]; then
      note "B_RETRY_IDENTICAL PASS"
    else
      note "B_RETRY_IDENTICAL FAIL"
      fail=1
    fi
    dropped=$(grep -c '^DROPPED_CONFIRM ' "$OUT/proxy-withhold.log" || true)
    if [[ "$dropped" -ge 1 ]]; then
      note "B_CONFIRM_WITHHELD PASS dropped=$dropped"
    else
      note "B_CONFIRM_WITHHELD FAIL"
      fail=1
    fi
    if "$MASTER" solicit 20000 4 "$OUT/b-recover.log" > "$OUT/b-recover.stdout" 2>"$OUT/b-recover.stderr"; then
      if grep -F -q "BIN idx=15 value=0 flags=01 time=699685567008" "$OUT/b-recover.log"; then
        note "B_RECOVERY PASS"
      else
        note "B_RECOVERY FAIL event missing"
        fail=1
      fi
    else
      note "B_RECOVERY FAIL process"
      fail=1
    fi
  fi
  stop_capture
  stop_go
else
  note "B_DIRECTION FAIL outstation did not start"
  fail=1
fi

echo "=== pcap check ==="
if python3 "$CHECK" duplicate "$OUT/evt001-duplicate.pcap"; then
  note "A_PCAP PASS"
else
  note "A_PCAP FAIL"
  fail=1
fi
if python3 "$CHECK" withhold "$OUT/evt001-withhold.pcap"; then
  note "B_PCAP PASS"
else
  note "B_PCAP FAIL"
  fail=1
fi

if [[ "$fail" -eq 0 ]]; then
  note "EVT001_NEGATIVE PASS"
  exit 0
fi
note "EVT001_NEGATIVE FAIL"
exit 1
