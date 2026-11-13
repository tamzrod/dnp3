# EVT-001 independent OpenDNP3 3.1.2 evidence

Recorded 2026-11-13 on branch `work`. This file describes one isolated run.
Internal Go tests are not this evidence.

## Peer and isolation

The peer is OpenDNP3 3.1.2, commit `26b4c01e4839bbbda8866655e086471c4917ee53`,
tarball sha256 `183cc29222c3cb58099a9753c43defed5cb8677fda985c3f88e38b5c19a36ff2`.
The static library was built with FetchContent downloads disabled. The image
that supplied that library and the isolated root filesystem is
`evt001-opendnp3:3.1.2`, id
`sha256:5ec7c3898e92e7d1c10247a5db62942aa3a67e48197b5f8f01583b53fef2ff92`.

The passing `evt_master` and `evt_outstation` were compiled from
`lab/evt001` inside that image and mounted into the run. The Go programs were
`CGO_ENABLED=0` builds of `lab/evt001/go-master` and `lab/evt001/go-outstation`
from this tree, also mounted. The image layer itself was built before those
two harness edits, so its digest is the library image, not a byte-identical
copy of the mounted binaries.

The container used `--network none` with `NET_RAW` and `NET_ADMIN` for
loopback capture. No ports were published. Before any DNP3 exchange the
script required no IPv4 default route and a failed TCP connect to
`1.1.1.1:443`. That gate printed `RESULT ISOLATION PASS`.

Listeners bound `127.0.0.1` only. Link addresses were master 1 and outstation
1024. Points were memory values. No field device, physical I/O, NGCP, or
SCADA path was used.

## Capture

| File | SHA-256 |
| --- | --- |
| `docs/verification/evt001/evt001.pcap` | `4c56f54975c00b4183341a87ab63db010d0573c0f104b4634f89db7aafe6cb98` |
| `docs/verification/evt001/pcap-derived.txt` | `c35f6727906b26c45ae8fefa12b059ee89b4ffdc27a49685d6acd4c07cac1764` |
| `docs/verification/evt001/run.log` | `582764e8892ffea8bc7589872311d5d9843e0ec101f260754968dcd0a7496999` |

The capture is classic pcap, DLT 1, 201 packets, 52 DNP3 application
functions: CONFIRM 8, READ 10, ENABLE_UNSOLICITED 2, DISABLE_UNSOLICITED 7,
RESPONSE 19, UNSOLICITED 6. The transcript of script results is `run.log`.
It ends with `RESULT EVT001_INTEROP PASS`.

## Direction A — Go outstation, OpenDNP3 master

All of these printed PASS on the recorded run.

| Scenario | Observed |
| --- | --- |
| Class 1 binary, Group 2 Variation 2 | index 7, value 1, flags `81`, time count `676368550000` |
| Class 2 analog, Group 32 Variation 3 | index 3, value -2, flags `01`, same time count |
| Class 3 counter, Group 22 Variation 1 (no time supplied) | index 4, value 16909060, flags `21`, variation `1601` |
| Solicited read | no `unsol=1` |
| Second class read | no event lines |
| Later binary kept its timestamp | index 11, time count `699685567008` |
| Unsolicited disabled | no event while the master was silent; the next class read returned index 12 |
| Unsolicited enabled | index 9, value 0, flags `01`, time count `699685567008`, `unsol=1` |
| Confirm | the following class read did not return index 9 |
| Overflow | 3 accepts and 1 reject with `EVT_MAX_EVENTS=9` |
| Overflow IIN on a class-0 read | `IIN lsb=02 msb=08` (class 1 and event-buffer overflow) |

The time counts are milliseconds since 2000-01-01T00:00:00Z.
`676368550000` is 2021-06-07T08:09:10Z. `699685567008` is
2022-03-04T05:06:07.008Z. OpenDNP3 writes `DNPTime::value` as the raw 48-bit
field. The harness stored those IEEE counts. Its own log calendar treats the
same integer as a Unix millisecond count, so the calendar string in an
OpenDNP3 log is not the comparison used here.

## Direction B — OpenDNP3 outstation, Go master

All of these printed PASS on the recorded run.

| Scenario | Observed |
| --- | --- |
| Class read | binary index 7 value true at 2021-06-07T08:09:10Z; analog index 3 value -2; counter index 4 value 16909060 |
| Second class read | `SECOND count=0` |
| Class 0 | a `CLASS0` line was produced. The event response IIN bytes were `9e00`, then `9000` after the events were read |
| Unsolicited disabled | watch produced no `UNSOL` line; the class read returned index 9 value false at 2022-03-04T05:06:07.008Z |
| Unsolicited enabled | `UNSOL` index 9 value true at that same timestamp |
| Confirm | the following class read was `EVENT count=0` |

The Go master reports a binary point's state in `value` and clears flag
`0x80` from `quality`. `quality=01` with `value=true` is the same flag octet
`0x81` that direction A printed as `flags=81`.

OpenDNP3 event responses use qualifier `0x28` (16-bit count, 16-bit index).
The Go decoder accepts `0x17` and `0x28`. The Go outstation still sends
`0x17`. The general object-header helper still treats `0x28` as a 16-bit
start/stop range for static and control requests. Event decode does not use
that helper.

## Link handshake correction

The first direction-B attempt failed in the Go master. OpenDNP3 sends a null
unsolicited response (primary unconfirmed user data, control `0x44`) when
the TCP session opens, and the secondary ACK follows that frame. The master
treated the primary frame as the reset ACK and closed. The recorded capture
is the rerun after the correction.

While Reset Link or Request Link Status is outstanding, the master now skips
at most four correctly addressed outstation-primary user-data frames and
still requires the secondary response. A mis-addressed primary frame, a
secondary NACK, and a secondary ACK with the wrong destination still fail
the handshake. After both handshake steps succeed, a buffered unsolicited
fragment is confirmed with the UNS bit set. OpenDNP3 ignores a null
unsolicited confirmation that leaves UNS clear, and it defers class reads
until that null is confirmed. The outstation harness starts with an empty
unsolicited class mask, so confirming the null does not push the loaded
events before the class read. `ENABLE_UNSOLICITED` sets the mask.

## Corrections covered by this run

`DecodeDNPTime` adds the 48-bit millisecond count with `time.UnixMilli`.
The previous `time.Duration` conversion overflowed. A year-2300 value and
the 48-bit maximum are regression-tested. This run's timestamps match the
counts above.

IIN bit 0 in each octet is the least significant bit, matching pinned
OpenDNP3 `IINField.h`. Named flags still mean the same conditions. Class 1
is octet 1 mask `0x02`. Event-buffer overflow is octet 2 mask `0x08`. The
direction-A class-0 line `lsb=02 msb=08` is the wire observation of those
two bits together.

## Not executed against this peer

OpenDNP3 confirms a fragment that requests confirmation, and it does so
once. The run did not make that master emit a duplicate confirm, and it did
not make that master withhold confirm. The Go tests cover a duplicate
confirm, a wrong sequence, and three identical unsolicited sends. Those
results are `UNIT_VERIFIED` only.

Relative time, Group 52, float event variations, and Secure Authentication
were not part of this run.

## Go gates

On this tree, before the checkpoint commit: `gofmt -l` printed nothing, `go vet ./...` passed, `go build ./...` passed, `go test ./... -count=1` passed, `go test -race ./... -count=1` passed, `bash scripts/verify-external-mvp.sh` passed, and `git diff --check` passed.

## Profile

`active_work/supported-profile.md` blob after the MEXT-019 row correction is
`3e592e77ad6e31d5492c609c5fdd62382505904b`. That row now names the OpenDNP3
LSB map. The profile still lists unsolicited responses and event delivery as
unsupported. This run does not promote them.
