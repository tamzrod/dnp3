#!/bin/bash
# EVT-001 isolated OpenDNP3 3.1.2 interoperability run.
# Intended for a container with --network none. Refuses to continue when a
# default route exists or an external TCP connect succeeds.
set -u
OUT=${OUT:-/out}
BIN=${BIN:-/opt/evt/bin}
MASTER=${MASTER:-/opt/evt/evt_master}
OUTSTATION=${OUTSTATION:-/opt/evt/evt_outstation}
PY=${PY:-/opt/evt/derive_pcap.py}
mkdir -p "$OUT"
exec > >(tee "$OUT/run.log") 2>&1

fail=0
note() { printf 'RESULT %s\n' "$*"; }

echo "=== isolation ==="
ip -4 route > "$OUT/routes.txt" || true
ip -4 addr > "$OUT/addrs.txt" || true
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

echo "=== capture ==="
tcpdump -i lo -s 0 -U -w "$OUT/evt001.pcap" >/tmp/tcpdump.err 2>&1 &
TCPDUMP_PID=$!
sleep 0.5
if ! kill -0 "$TCPDUMP_PID" 2>/dev/null; then
  note "PCAP FAIL tcpdump did not stay up"
  cat /tmp/tcpdump.err || true
  fail=1
else
  note "PCAP capture started pid=$TCPDUMP_PID"
fi

cleanup() {
  if [[ -n "${TCPDUMP_PID:-}" ]]; then
    kill "$TCPDUMP_PID" 2>/dev/null || true
    wait "$TCPDUMP_PID" 2>/dev/null || true
  fi
  if [[ -n "${GO_PID:-}" ]]; then kill "$GO_PID" 2>/dev/null || true; fi
  if [[ -n "${OD_PID:-}" ]]; then kill "$OD_PID" 2>/dev/null || true; fi
}
trap cleanup EXIT

wait_file() {
  local file=$1 needle=$2 tries=$3
  local i
  for ((i = 0; i < tries; i++)); do
    if [[ -f "$file" ]] && grep -q "$needle" "$file"; then
      return 0
    fi
    sleep 0.2
  done
  return 1
}

start_go() {
  local port=$1
  rm -f /tmp/go.fifo
  mkfifo /tmp/go.fifo
  exec 3<>/tmp/go.fifo
  EVT_MAX_EVENTS=${EVT_MAX_EVENTS:-} "$BIN/go-outstation" "$port" < /tmp/go.fifo > "$OUT/go-outstation.stdout" 2>"$OUT/go-outstation.stderr" &
  GO_PID=$!
  if ! wait_file "$OUT/go-outstation.stdout" READY 50; then
    note "GO_OUTSTATION FAIL no READY"
    return 1
  fi
  note "GO_OUTSTATION READY port=$port"
}

stop_go() {
  if [[ -n "${GO_PID:-}" ]]; then
    echo quit >&3 || true
    wait "$GO_PID" 2>/dev/null || true
    GO_PID=
  fi
  exec 3>&- || true
  exec 3<&- || true
}

go_cmd() {
  echo "$*" >&3
  # The outstation writes one response line per command.
  sleep 0.2
}

run_master() {
  local mode=$1 port=$2 seconds=$3 log=$4
  "$MASTER" "$mode" "$port" "$seconds" "$log" > "${log}.stdout" 2>"${log}.stderr"
  return $?
}

require_line() {
  local log=$1 line=$2 name=$3
  if grep -F -q "$line" "$log"; then
    note "$name PASS"
  else
    note "$name FAIL missing [$line]"
    fail=1
  fi
}

require_absent() {
  local log=$1 line=$2 name=$3
  if grep -F -q "$line" "$log"; then
    note "$name FAIL unexpected [$line]"
    fail=1
  else
    note "$name PASS"
  fi
}

echo "=== A Go outstation -> OpenDNP3 master ==="
if start_go 20000; then
  go_cmd bin 7 1 1 2021-06-07T08:09:10Z
  go_cmd ana 3 -2 2 2021-06-07T08:09:10Z
  go_cmd ctr 4 16909060 3 none
  if run_master solicit 20000 4 "$OUT/a-solicit.log"; then
    note "A_SOLICIT process PASS"
  else
    note "A_SOLICIT process FAIL"
    fail=1
  fi
  require_line "$OUT/a-solicit.log" "BIN idx=7 value=1 flags=81 time=676368550000" "A_BINARY"
  require_line "$OUT/a-solicit.log" "ANA idx=3 value=-2 flags=01 time=676368550000" "A_ANALOG"
  require_line "$OUT/a-solicit.log" "CTR idx=4 value=16909060 flags=21" "A_COUNTER"
  require_line "$OUT/a-solicit.log" "gv=1601" "A_COUNTER_VARIATION"
  if grep -F "unsol=1" "$OUT/a-solicit.log" >/dev/null; then
    note "A_SOLICIT_NOT_UNSOL FAIL"
    fail=1
  else
    note "A_SOLICIT_NOT_UNSOL PASS"
  fi

  if run_master solicit 20000 3 "$OUT/a-empty.log"; then
    note "A_EMPTY process PASS"
  else
    note "A_EMPTY process FAIL"
    fail=1
  fi
  require_absent "$OUT/a-empty.log" "BIN idx=" "A_EMPTY_EVENTS"
  require_absent "$OUT/a-empty.log" "ANA idx=" "A_EMPTY_ANALOG"
  require_absent "$OUT/a-empty.log" "CTR idx=" "A_EMPTY_COUNTER"

  go_cmd bin 11 1 1 2022-03-04T05:06:07.008Z
  if run_master solicit 20000 4 "$OUT/a-retain.log"; then
    note "A_RETAIN process PASS"
  else
    note "A_RETAIN process FAIL"
    fail=1
  fi
  require_line "$OUT/a-retain.log" "BIN idx=11 value=1 flags=81 time=699685567008" "A_RETAIN_EVENT"

  "$MASTER" silent 20000 4 "$OUT/a-silent.log" > "$OUT/a-silent.log.stdout" 2>"$OUT/a-silent.log.stderr" &
  SILENT_PID=$!
  if wait_file "$OUT/a-silent.log" "STATE OPEN" 30; then
    go_cmd bin 12 1 1 2021-06-07T08:09:10Z
  else
    note "A_SILENT open FAIL"
    fail=1
  fi
  wait "$SILENT_PID" || note "A_SILENT process FAIL"
  require_absent "$OUT/a-silent.log" "BIN idx=" "A_DISABLED_NO_UNSOL"
  if run_master solicit 20000 4 "$OUT/a-after-disable.log"; then
    require_line "$OUT/a-after-disable.log" "BIN idx=12 value=1 flags=81 time=676368550000" "A_DISABLED_CLASS_READ"
  else
    note "A_DISABLED_CLASS_READ FAIL process"
    fail=1
  fi

  "$MASTER" unsol 20000 6 "$OUT/a-unsol.log" > "$OUT/a-unsol.log.stdout" 2>"$OUT/a-unsol.log.stderr" &
  UNSOL_PID=$!
  if wait_file "$OUT/a-unsol.log" "STATE OPEN" 30; then
    sleep 0.8
    go_cmd bin 9 0 1 2022-03-04T05:06:07.008Z
  else
    note "A_UNSOL open FAIL"
    fail=1
  fi
  wait "$UNSOL_PID" || { note "A_UNSOL process FAIL"; fail=1; }
  require_line "$OUT/a-unsol.log" "BIN idx=9 value=0 flags=01 time=699685567008" "A_UNSOL_EVENT"
  require_line "$OUT/a-unsol.log" "unsol=1" "A_UNSOL_FLAG"
  if run_master solicit 20000 3 "$OUT/a-confirmed-gone.log"; then
    require_absent "$OUT/a-confirmed-gone.log" "BIN idx=9" "A_CONFIRM_REMOVES"
  else
    note "A_CONFIRM_REMOVES FAIL process"
    fail=1
  fi

  stop_go
  EVT_MAX_EVENTS=9 start_go 20000 || { note "A_OVERFLOW outstation FAIL"; fail=1; }
  go_cmd fill 4
  if grep -q 'FILL accepted=3 rejected=1' "$OUT/go-outstation.stdout"; then
    note "A_OVERFLOW_REJECT PASS"
  else
    note "A_OVERFLOW_REJECT FAIL"
    fail=1
  fi
  if run_master class0 20000 4 "$OUT/a-overflow-iin.log"; then
    if grep -E 'IIN lsb=[0-9a-f][0-9a-f] msb=[0-9a-f][89a-f]' "$OUT/a-overflow-iin.log" >/dev/null \
      && grep -E 'IIN lsb=[0-9a-f][2367abef] msb=' "$OUT/a-overflow-iin.log" >/dev/null; then
      note "A_OVERFLOW_IIN PASS"
    else
      note "A_OVERFLOW_IIN FAIL"
      fail=1
    fi
  else
    note "A_OVERFLOW_IIN FAIL process"
    fail=1
  fi
  stop_go
else
  note "A_DIRECTION FAIL outstation did not start"
  fail=1
fi

echo "=== B OpenDNP3 outstation -> Go master ==="
rm -f /tmp/od.fifo
mkfifo /tmp/od.fifo
exec 4<>/tmp/od.fifo
"$OUTSTATION" 20001 < /tmp/od.fifo > "$OUT/od-outstation.stdout" 2>"$OUT/od-outstation.stderr" &
OD_PID=$!
if wait_file "$OUT/od-outstation.stdout" READY 50; then
  note "OD_OUTSTATION READY"
  echo load >&4
  sleep 0.3
  "$BIN/go-master" read 20001 > "$OUT/b-read.log" 2>"$OUT/b-read.err" || { note "B_READ process FAIL"; fail=1; }
  require_line "$OUT/b-read.log" "EVENT kind=bin idx=7 value=true" "B_BINARY"
  require_line "$OUT/b-read.log" "time=2021-06-07T08:09:10Z" "B_BINARY_TIME"
  require_line "$OUT/b-read.log" "EVENT kind=ana idx=3 value=-2" "B_ANALOG"
  require_line "$OUT/b-read.log" "EVENT kind=ctr idx=4 value=16909060" "B_COUNTER"
  require_line "$OUT/b-read.log" "SECOND count=0" "B_SECOND_EMPTY"
  if grep -q '^CLASS0 error' "$OUT/b-read.log"; then
    note "B_CLASS0 FAIL"
    fail=1
  elif grep -q '^CLASS0 ' "$OUT/b-read.log"; then
    note "B_CLASS0 PASS"
  else
    note "B_CLASS0 FAIL missing"
    fail=1
  fi

  "$BIN/go-master" negative 20001 3 > "$OUT/b-negative.log" 2>"$OUT/b-negative.err" &
  NEG_PID=$!
  if wait_file "$OUT/b-negative.log" WATCHING 40; then
    echo live >&4
    sleep 0.2
  else
    note "B_NEGATIVE watch FAIL"
    fail=1
  fi
  wait "$NEG_PID" || { note "B_NEGATIVE process FAIL"; fail=1; }
  if grep -q '^UNSOL ' "$OUT/b-negative.log"; then
    note "B_NEGATIVE_NO_UNSOL FAIL"
    fail=1
  else
    note "B_NEGATIVE_NO_UNSOL PASS"
  fi
  require_line "$OUT/b-negative.log" "EVENT kind=bin idx=9 value=false" "B_NEGATIVE_CLASS_READ"
  require_line "$OUT/b-negative.log" "time=2022-03-04T05:06:07.008Z" "B_NEGATIVE_TIME"

  "$BIN/go-master" unsol 20001 6 > "$OUT/b-unsol.log" 2>"$OUT/b-unsol.err" &
  BUNSOL_PID=$!
  if wait_file "$OUT/b-unsol.log" ENABLED 40; then
    echo live-on >&4
  else
    note "B_UNSOL enable FAIL"
    fail=1
  fi
  wait "$BUNSOL_PID" || { note "B_UNSOL process FAIL"; fail=1; }
  require_line "$OUT/b-unsol.log" "UNSOL kind=bin idx=9 value=true" "B_UNSOL_EVENT"
  require_line "$OUT/b-unsol.log" "time=2022-03-04T05:06:07.008Z" "B_UNSOL_TIME"
  "$BIN/go-master" empty 20001 > "$OUT/b-empty.log" 2>"$OUT/b-empty.err" || { note "B_EMPTY process FAIL"; fail=1; }
  require_line "$OUT/b-empty.log" "EVENT count=0" "B_CONFIRM_REMOVES"
else
  note "B_DIRECTION FAIL outstation did not start"
  fail=1
fi
echo quit >&4 || true
wait "$OD_PID" 2>/dev/null || true
OD_PID=

echo "=== pcap derive ==="
sleep 0.5
kill "$TCPDUMP_PID" 2>/dev/null || true
wait "$TCPDUMP_PID" 2>/dev/null || true
TCPDUMP_PID=
if python3 "$PY" "$OUT/evt001.pcap" > "$OUT/pcap-derived.txt"; then
  cat "$OUT/pcap-derived.txt"
  note "PCAP_DERIVE PASS"
else
  cat "$OUT/pcap-derived.txt" 2>/dev/null || true
  note "PCAP_DERIVE FAIL"
  fail=1
fi

if [[ "$fail" -eq 0 ]]; then
  note "EVT001_INTEROP PASS"
  exit 0
fi
note "EVT001_INTEROP FAIL"
exit 1
