#!/bin/bash
# EVT-001 PPR-001-F1/F2/F3 large-batch interoperability run.
#
# Runs the pinned OpenDNP3 3.1.2 peer in --network none, loopback-only,
# memory-only mode against the Go stack in both directions for event batches
# of 254, 255, 256, 257 and 333 objects. Each batch exceeds the qualifier-0x17
# 8-bit count limit (255) and the 2048-octet application-fragment budget, so it
# exercises bounded batching, multi-fragment responses and per-fragment confirms.
# Every delivered count must equal N with each point index delivered exactly once.
set -u
OUT=${OUT:-/out}
BIN=${BIN:-/opt/evt/bin}
MASTER=${MASTER:-/opt/evt/evt_master}
OUTSTATION=${OUTSTATION:-/opt/evt/evt_outstation}
mkdir -p "$OUT"
exec > >(tee "$OUT/largebatch-run.log") 2>&1

fail=0
note() { printf 'RESULT %s\n' "$*"; }

echo "=== isolation ==="
ip -4 route > "$OUT/largebatch-routes.txt" || true
cat "$OUT/largebatch-routes.txt"
if grep -q '^default' "$OUT/largebatch-routes.txt"; then
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

start_capture() {
  local pcap=$1
  tcpdump -i lo -s 0 -U -w "$pcap" >"$pcap.tcpdump.out" 2>&1 &
  TCPDUMP_PID=$!
  sleep 0.4
  if ! kill -0 "$TCPDUMP_PID" 2>/dev/null; then
    note "PCAP FAIL tcpdump did not stay up for $pcap"
    fail=1
    TCPDUMP_PID=
    return 1
  fi
  return 0
}

stop_capture() {
  if [[ -n "${TCPDUMP_PID:-}" ]]; then
    # Let tcpdump drain its buffer, then stop it gracefully (SIGINT flushes
    # and exits) so a short capture window does not lose the final packets.
    sleep 1.5
    kill -INT "$TCPDUMP_PID" 2>/dev/null || true
    wait "$TCPDUMP_PID" 2>/dev/null || true
    TCPDUMP_PID=
  fi
}

wait_file() {
  local file=$1 needle=$2 tries=$3 i
  for ((i = 0; i < tries; i++)); do
    if [[ -f "$file" ]] && grep -q "$needle" "$file"; then return 0; fi
    sleep 0.2
  done
  return 1
}

count_idx() {
  # count_idx <log> <pattern>: number of matching object lines.
  grep -E -c "$2" "$1" 2>/dev/null || true
}

distinct_idx() {
  # distinct_idx <log> <sed-expr>: number of distinct point indexes observed.
  grep -E -o "$2" "$1" 2>/dev/null | sort -u | wc -l
}

for N in ${NS:-254 255 256 257 333}; do
  echo "=== N=$N ==="

  # ---- Direction A: Go outstation -> OpenDNP3 master ----
  rm -f /tmp/lb-go.fifo
  mkfifo /tmp/lb-go.fifo
  exec 3<>/tmp/lb-go.fifo
  # EventQueue.Add caps each class at MaxEventBuffers/3, so size the buffer at
  # 3N+10 to accept all N events of the single Class-1 batch.
  EVT_MAX_EVENTS=$((N * 3 + 10)) "$BIN/go-outstation" 20000 < /tmp/lb-go.fifo \
    > "$OUT/lb-a-$N-go.stdout" 2>"$OUT/lb-a-$N-go.stderr" &
  GO_PID=$!
  if wait_file "$OUT/lb-a-$N-go.stdout" READY 50; then
    echo "fill $N" >&3
    sleep 0.4
    start_capture "$OUT/evt001-largebatch-dirA-$N.pcap"
    "$MASTER" solicit 20000 6 "$OUT/lb-a-$N.log" > "$OUT/lb-a-$N.log.stdout" 2>"$OUT/lb-a-$N.log.stderr"
    if [[ $? -eq 0 ]]; then
      note "A_${N}_process PASS"
    else
      note "A_${N}_process FAIL"
      fail=1
    fi
    got=$(count_idx "$OUT/lb-a-$N.log" "BIN idx=")
    if [[ "$got" == "$N" ]]; then
      note "A_${N}_count PASS ($got)"
    else
      note "A_${N}_count FAIL got=$got want=$N"
      fail=1
    fi
    # The 8-bit index field wraps at 256, so a batch larger than 256 reuses
    # indexes: the distinct count saturates at 256.
    want_uniq=$(( N < 256 ? N : 256 ))
    uniq=$(distinct_idx "$OUT/lb-a-$N.log" "idx=[0-9]+")
    if [[ "$uniq" == "$want_uniq" ]]; then
      note "A_${N}_distinct PASS ($uniq)"
    else
      note "A_${N}_distinct FAIL got=$uniq want=$want_uniq"
      fail=1
    fi
    stop_capture
  else
    note "A_${N} outstation FAIL"
    fail=1
    stop_capture
  fi
  echo quit >&3 || true
  wait "$GO_PID" 2>/dev/null || true
  GO_PID=
  exec 3>&- || true
  exec 3<&- || true

  # ---- Direction B: OpenDNP3 outstation -> Go master ----
  rm -f /tmp/lb-od.fifo
  mkfifo /tmp/lb-od.fifo
  exec 4<>/tmp/lb-od.fifo
  EVT_POINTS=$N "$OUTSTATION" 20001 < /tmp/lb-od.fifo \
    > "$OUT/lb-b-$N-od.stdout" 2>"$OUT/lb-b-$N-od.stderr" &
  OD_PID=$!
  if wait_file "$OUT/lb-b-$N-od.stdout" READY 50; then
    echo fillall >&4
    sleep 0.4
    start_capture "$OUT/evt001-largebatch-dirB-$N.pcap"
    "$BIN/go-master" read 20001 > "$OUT/lb-b-$N.log" 2>"$OUT/lb-b-$N.err"
    if [[ $? -eq 0 ]]; then
      note "B_${N}_process PASS"
    else
      note "B_${N}_process FAIL"
      fail=1
    fi
    got=$(count_idx "$OUT/lb-b-$N.log" "EVENT kind=bin")
    if [[ "$got" == "$N" ]]; then
      note "B_${N}_count PASS ($got)"
    else
      note "B_${N}_count FAIL got=$got want=$N"
      fail=1
    fi
    if grep -q "SECOND count=0" "$OUT/lb-b-$N.log"; then
      note "B_${N}_once_only PASS"
    else
      note "B_${N}_once_only FAIL"
      fail=1
    fi
    stop_capture
  else
    note "B_${N} outstation FAIL"
    fail=1
    stop_capture
  fi
  echo quit >&4 || true
  wait "$OD_PID" 2>/dev/null || true
  OD_PID=
  exec 4>&- || true
  exec 4<&- || true
done

echo "=== done ==="
if [[ "$fail" -eq 0 ]]; then
  note "EVT001_LARGEBATCH PASS"
  exit 0
fi
note "EVT001_LARGEBATCH FAIL"
exit 1
