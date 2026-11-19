# AO-008 EG628 donor interoperability patch — verification note

Date: 2026-11-19
Branch: `work`
CURRENT: AO-008 (bounded corrective continuation; **not** an archive, promotion,
merge or deployment)
Base `work` HEAD at start: `0bbbc4e` (`origin/work` `0bbbc4e`)
Checkpoint commit: `181e2d5` (rebased onto `53c7b86`, the concurrent G12V1 CROB
pulse-semantics commit that landed on `origin/work` during this session)

## Scope actually worked

Independent donor-patch audit plus the two defects that independent CHECK left
open in the previous checkpoint:

1. **Static object-header count wrap (>255 / >65535).** Confirmed reproducible on
   `work`. Static builders emitted an 8-bit count qualifier for any point count,
   so a 300-point block announced count 44 and a 65536-point block announced
   count 0. A strict peer then mis-parses every following object header.
2. **Missing direct malformed READ negative gate.** Confirmed reproducible: a
   READ carrying an unknown qualifier was treated as a 4-octet header and fell
   through to "all static data", so a request the outstation did not understand
   was answered with a complete integrity response and IIN `00 00` — a false
   success.

### Donor commits re-checked, not blindly adopted

`a038637` (packed G1V1/G10V1) and `5842395` (G12V1 status echo, qualifier 0x17,
latch commands) were already integrated on `work`, as was the request-scoped IIN
clear from `f72bf05`. No history was rewritten and nothing was cherry-picked.

## Changes

`internal/outstation/outstation.go`

- `const maxStaticPoints = 0xFFFF` — the largest count any DNP3 count qualifier
  can describe.
- `appendCountQualifier` / `appendIndexedCountQualifier` choose 0x07 (8-bit) up
  to 255 points and 0x27 (16-bit) above it, and refuse a block above
  `maxStaticPoints` instead of wrapping the count.
- Indexed static builders (G1V2, G30V2, G20V5) fall back to a sequential 0x27
  block above 255 points, because an 8-bit index prefix cannot address point
  256. Indexing is preserved at and below 255 points.
- `blockSpan` now computes real static block spans (header + count × value
  octets, with packed G1V1/G10V1 sized per block) via `staticValueSize`, so a
  multi-block static response is split at object boundaries instead of treating
  the first sequential block as the whole remainder.
- `requestObjectHeaderSize` returns `(size, ok)` and reports unknown qualifiers,
  so a malformed READ is rejected explicitly rather than guessed at.
- A dropped static block is reported through IIN OBJECT_UNKNOWN instead of
  looking like a complete (if small) integrity response.

`internal/outstation/event_report.go` — `classMask` follows the new
`requestObjectHeaderSize` signature.

## Regression coverage added

- `static_count_boundary_test.go`: 255/256/300 qualifier selection and declared
  counts for G1V1, G10V1, G30V1, G20V1, G40V1; indexed→sequential fallback at
  256; the 65535/65536 limit; OBJECT_UNKNOWN above the limit and no leakage into
  the following read; multi-fragment Class-0 split with every declared count
  intact and every point delivered once; unsplittable block fails closed.
- `read_malformed_test.go`: five malformed READ headers rejected with
  OBJECT_UNKNOWN and no object data, plus a positive control proving all six
  v0-allow-listed qualifiers are still served.
- `test/integration/large_static_block_tcp_test.go`: 300 points per group over
  real TCP through IntegrityPoll, asserting exact index/value for every point
  and no false IIN event or error bits.
- `event_batch_test.go`: the sequential-split fixture was corrected from an
  invalid one-octet-per-point G1V1 block to a valid G20V1 block.
- `static_qualifier_test.go`: updated for the new signature, plus an unknown
  qualifier case.

### RED → GREEN evidence

Against donor `work` source (patch stashed), a probe on the builder output
printed:

```
n=300:   qualifier=0x07 declaredCount=44   blockLen=42
n=65536: qualifier=0x07 declaredCount=0    blockLen=8196
```

With the patch applied:

```
n=300:   qualifier=0x27 declaredCount=300  blockLen=43
n=65536: builder emitted nothing (refused; IIN OBJECT_UNKNOWN)
```

The malformed-READ negative tests also fail on donor source with
`IIN = 00 00, want OBJECT_UNKNOWN` and pass with the patch. The probe file was
deleted after the comparison; it is not committed.

## Verification actually run

| Check | Result |
| --- | --- |
| `gofmt -l` | PASS (empty) |
| `git diff --check` | PASS |
| `go vet ./...` | PASS |
| `go build ./...` | PASS |
| `go test ./... -count=1` | PASS |
| `go test -race ./... -count=1` | PASS |
| `bash scripts/verify-external-mvp.sh` | PASS (tiers 1–3) |
| `bash scripts/cwal-check.sh` | NOT_RUN before commit (requires a clean tree); see checkpoint note |
| Pinned OpenDNP3 3.1.2 peer, `--network none` | NOT_RUN — Docker daemon unavailable in this environment |

### Known pre-existing flake (not introduced here)

`TestOperateG41StableOptInBothRolesRealTCP` intermittently fails with
`ErrControlInProgress` when the integration package runs under `-race` alongside
other packages. Reproduced on donor `work` **with this patch stashed** (1 of 6
runs under CPU load), and it passes 8/8 in isolation and 4/4 in a full
`-race ./...` run with the patch. It is a pre-existing timing sensitivity in that
G41 control test, unrelated to static object encoding.

## Findings NOT fixed here (outside this mission's wire scope)

The outstation encodes multi-octet values **big-endian** for the indexed and
non-v0 static variations, while `pkg/dnp3/master` decodes **little-endian** and
`AGENTS.md` mandates LSB-first. Verified with a probe: G30V2 for value 1000
emits `03 E8`, which the master reads as −6141.

Affected: G30V2/V3/V5, G20V5/V6, G40V2 (and the same pattern in the
corresponding indexed paths). The v0 profile path actually served and claimed
(G1V1, G30V1, G20V1) is little-endian and correct, which is why this does not
affect the supported profile.

This was deliberately **not** changed: it alters wire behavior outside the
authorized static-count/READ scope and would need its own design approval and
golden vectors. It is recorded here so it is not lost.

## Evidence boundaries (unchanged)

- EG628 Windows DNP3 Tester pcap/log, the object offset-38 attribution, full
  class 0+1+2+3 success, adapter revision/point count and the source build pin
  remain **NOT_VERIFIED / MISSING**. The donor screenshot is not a packet
  capture and was not treated as one.
- No third-party interop claim is made from the in-repo TCP loopback gate.
- The EG628 incident stays unresolved until an independent-master capture
  retest passes.
- G41 remains **UNSUPPORTED**; `active_work/supported-profile.md` is unchanged.

## Disposition

`AO008_EG628_PATCH_READY_FOR_INDEPENDENT_CHECK`.
CURRENT: AO-008, QUEUED: NONE, Status: TASK_IN_PROGRESS.
STOP for independent read-only CHECK. No archive, close, PR, merge, release or
deployment.

---

# SBO command-status timeout fix — supplementary verification note

# AO-008 EG628 interoperability patch check — 2026-11-19

Scope: bounded corrective continuation under CURRENT `AO-008` on branch `work`,
authorized by the latest `EG628 interoperability PATCH CWAL` addendum in
`workflow/active_work/AO-008.md`. No published profile, KDE, deployment, PR,
merge or upstream push was performed.

## Defect and root cause

Owner-reported symptom: Windows DNP3 Tester LatchOn/LatchOff at CROB index 1
failed with `the outstation returned no command statuses` in both Select-Before-Operate
and OperateOnly (Direct Operate). Live pcap decode showed:

- SELECT request APP `CC CD 03 0C 01 17 01 01 03 01 E8 03 00 00 00 00 00 00 00`
  returned a valid G12V1 echo with `IIN=0000`.
- OPERATE request APP `CD CE 04 0C 01 17 01 01 03 01 E8 03 00 00 00 00 00 00 00`
  returned only `C0 CE 81 00 01` — `IIN=0001` and **no G12V1 command-status object**.
- Direct Operate at the same index echoed status `0x00` correctly.

`DNP3_DEBUG=1` on the failing regression test produced:

```text
handleSelect: group=12, variation=1, index=1, valueLen=11
handleOperate: received
handleOperate: group=12, variation=1, index=1
handleOperate: select timeout for index=1
```

Root cause: `pkg/dnp3/outstation/server.go` built the internal
`outstation.Config` literal without setting `SBOTimeout`; it remained `0`.
`internal/outstation/outstation.go:handleOperate` then evaluated
`time.Since(pending.Timestamp) > 0`, so every OPERATE after an otherwise valid
SELECT was rejected as expired before the command handler could run. The IIN-only
response is what the tester reported as “no command statuses”.

## Fix

- `internal/outstation/outstation.go`: added `defaultSBOTimeoutMS = 10000`, used
  it in `DefaultConfig`, and made `handleOperate` treat a non-positive configured
  timeout as the 10-second default. This fails safe for every caller that builds
  a `Config` literal without the field.
- `pkg/dnp3/outstation/server.go`: added `Config.SBOTimeout time.Duration`,
  default `10 * time.Second`, `WithSBOTimeout`, and propagation of the public
  timeout plus `MaxFragmentSize` into the internal outstation config. This is an
  additive configuration API; no published profile claim changed.

No tester-specific branching, no false SUCCESS, no handler-by-pass change, and no
change to G12V1 echo/qualifier/status encoding was made.

## RED → GREEN regression evidence

EG628 adapter end-to-end raw TCP test (captured bytes against the real adapter
handler with a fake MMA2 backend, no physical I/O):

- Before fix: `evidence/protocols/dnp3-crob-interop/sbo-regression-failing-before-fix-20261119T091708Z.txt`
- After fix: `evidence/protocols/dnp3-crob-interop/sbo-regression-passing-after-fix-20261119T092211Z.txt`
- Failing pcap frame extract: `evidence/protocols/dnp3-crob-interop/sbo-failing-frames-from-live-pcap-20261119T092223Z.txt`

Library-level tests added:

- `internal/outstation/sbo_timeout_test.go` — zero `SBOTimeout` still permits
  SELECT→OPERATE, echoes status `0x00`, invokes the handler once, and does not
  write on SELECT.
- `pkg/dnp3/outstation/server_test.go` — default timeout is 10 s and
  `WithSBOTimeout` is applied.

EG628 adapter tests (`src/eg628/dnp3adapter/sbo_regression_test.go`) cover all
four CROB operations (LatchOn, LatchOff, PulseOn, PulseOff) through Direct
Operate, Select-Before-Operate (SELECT only and OPERATE after successful SELECT),
byte-exact G12V1 status echoes, IIN `0000`, and zero physical writes on SELECT.

## Gates actually run

All commands ran on branch `work` with `GOWORK=off` where applicable.

| Gate | Result |
| --- | --- |
| `gofmt -w <changed files>` | PASS |
| `git diff --check` | PASS (clean) |
| `gofmt -l .` | PASS (empty) |
| `go vet ./...` | PASS |
| `go build ./...` | PASS |
| `go test ./... -count=1` | PASS (all packages) |
| `bash scripts/verify-external-mvp.sh` | PASS (Tiers 1–3 OK) |
| `go test -race ./... -count=1` | **FAIL — environment limitation**: `FATAL: ThreadSanitizer: unsupported VMA range / Found 39 - Supported 48`. Not a code result. |
| `bash scripts/cwal-check.sh` | **FAIL at checkpoint**: working tree intentionally dirty with unpushed fix; owner directed no upstream push yet. No force-push or branch reconciliation performed. |

Published supported-profile blob `3e592e77ad6e31d5492c609c5fdd62382505904b`
was not modified. G41 remains published UNSUPPORTED. No PR, merge, release,
KDE or governance change was performed.

**Separate EG628-side action (outside this repository's CWAL scope):** the owner's EG628 mission
authorization subsequently covered installing the built binary on the EG628 DEV unit with backup
and rollback, and a DO2→DI2 loopback retest passed. That action and its evidence are recorded only
in the EG628 workspace (`docs/DISCOVERY_LOG.md` entries 077–078,
`evidence/software/sbo-fix-deploy-verify-20261119T092757Z.txt`); no deployment was performed by, or
checkpointed into, this DNP3 repository.

## Final external tester outcome — 2026-11-19T09:42Z (supersedes earlier NOT_VERIFIED boundary)

Owner confirmed the Windows DNP3 Tester retry was OK. This is independently VERIFIED from the retained passive pcap, with the exact object bytes preserved in the EG628 workspace at
`evidence/protocols/dnp3-crob-interop/tester-pulseoff-ontime-success-after-fix-20261119T094232Z.txt`:

- `1795081262` (09:40:22Z): Latch On SELECT and OPERATE both echoed G12V1 trailing status `0x00`, `IIN=0000`.
- `1795081271` (09:41:11Z): the previously failing Pulse Off SBO request (`0C 01 17 01 01 02 01 E8 03 00 00 00 00 00 00 00`) now returned `C0 CF 81 00 00 0C 01 17 01 01 02 01 E8 03 00 00 00 00 00 00 00` — `IIN=0000`, status `0x00`.

Pulse Off acceptance additionally required the EG628-side pulse-width-only CROB fix (`OnTime=1000`, `OffTime=0`), which is outside this repository. This repository's scoped SBO-timeout patch is unchanged by that later EG628 fix. The original Windows tester PASS boundary is therefore CLOSED for the captured Latch On / Pulse Off retry; no IEEE 1815 conformance or general interoperability claim is made. The repository checkpoint is being pushed to `work` without force, PR, merge or profile change.

---

# Second corrective continuation — short READ + EG628 Pulse Off wire — 2026-11-19

CURRENT: AO-008 (bounded corrective continuation; **not** an archive, promotion,
merge or deployment). Base `work` HEAD at start: `50d4c8d` (`origin/work` `50d4c8d`,
after the concurrent SBO-timeout and short-READ-authorization commits).

## Scope actually worked

1. **Short READ independent-CHECK pushback.** `buildReadResponse` answered any
   request shorter than 3 octets with a full integrity response, and ignored
   surplus bytes after the last complete object header. Both are false successes:
   the request named no parseable object, yet the master saw IIN `00 00` plus
   complete static data.
2. **EG628 Pulse Off wire regression.** Added the captured G12V1 Pulse Off CROB
   as an exact-byte regression for Direct Operate and SBO SELECT→OPERATE.

## Defect 1 — short/truncated READ answered as success (RED→GREEN)

**Defect.** A READ of 0, 1 or 2 octets fell through the `len(requestData) < 3`
branch and returned `buildAllStaticData()`. Separately, the header walk stopped
when fewer than 3 octets remained, so a valid header plus 1–2 stray trailing
octets was answered as a complete read. A strict master cannot distinguish either
response from a genuine integrity scan.

**Fix.** New `validateReadHeaders` pre-pass walks every object header before
anything is built and rejects: a request shorter than 3 octets; a qualifier
outside the supported table; a header that runs past the end of the data; and
surplus trailing octets after the last complete header. A rejected READ sets IIN
OBJECT_UNKNOWN, returns no object data, and has no side effect because nothing is
built or staged.

**RED proven** against the pre-fix source with the new tests in place: all four
short forms (zero-length, nil, 1 octet, 2 octets) and all three trailing-partial
forms returned `IIN = 00 00` with full static data instead of OBJECT_UNKNOWN.
**GREEN proven** with the fix: all pass, and the valid-request positive control
still serves every allow-listed qualifier.

**Three pre-existing tests corrected, not weakened.** `event_class_stub_test.go`
and `static_variation_test.go` sent `60 01 06 00` — a valid 3-octet all-objects
header plus a spurious trailing `0x00`. They passed only because the old code
silently ignored surplus bytes. The repository's own vector
`active_work/testdata/class0-integrity-request.hex` states the all-objects header
"is exactly three octets" (`3C 01 06`), and `buildPollRequest`/`classEnableHeader`
emit exactly three octets. The tests were updated to the canonical three-octet
form; their assertions are unchanged.

New coverage: `TestReadRejectsShortRequest` (4 subcases),
`TestReadRejectsTrailingPartialHeader` (3 subcases),
`TestRefusedReadDoesNotLeakIINIntoNextRequest` (refusal then valid READ).

## Defect 2 — EG628 Pulse Off: library needs no fix, only a pinned regression

The observed `BAD_FORMAT` (0x03) on the tester's Pulse Off CROB
(`op=0x02, count=1, OnTime=1000ms, OffTime=0ms`, qualifier `0x17`, index 1) was
an **EG628 adapter** `validatePulse` decision, resolved in `tamzrod/eg628` at
`a56234f` by `effectivePulseTimings()`. It was not a `tamzrod/dnp3` framing or
duration bug, and no library pulse-timing logic was added: `executeControl`
decodes the CROB fields LSB-first and passes them to the handler unchanged, and
`buildG12V1EchoObject` re-emits the request header and control fields verbatim
with the handler's real status.

`internal/outstation/g12v1_pulse_wire_test.go` pins that contract using the
captured application-layer bytes
`0C 01 17 01 01 02 01 E8 03 00 00 00 00 00 00 00`:

- Direct Operate with an accepting handler: handler sees
  `Code=0x02, Count=1, OnTime=1000, OffTime=0, Qualifier=0x17`; echo is
  byte-identical with trailing status `0x00`.
- SBO SELECT→OPERATE: SELECT echoes `0x00` and does not call the handler;
  OPERATE executes once and echoes identically.
- Refusing handler (`0x03 BAD_FORMAT`): the master reads `0x03`, never a
  manufactured SUCCESS, on both Direct Operate and SELECT→OPERATE.
- Unclassifiable handler error: fails closed with no response object.

**Mutation-checked.** Swapping the OnTime/OffTime decode offsets makes
`TestG12V1CapturedPulseOffDirectOperateEchoesUnchanged` fail with
`OnTime:0 OffTime:1000`, so the regression has teeth against the exact defect
class the incident was originally mis-attributed to. Source restored afterwards.

## Gates actually run — all PASS

`git diff --check`; `gofmt -l` (empty); `go vet ./...`; `go build ./...`;
`go test ./... -count=1`; `go test -race ./... -count=1`; `bash
scripts/verify-external-mvp.sh` (tiers 1–3 OK). `cwal-check.sh` run at the clean
checkpoint.

**NOT_RUN:** pinned OpenDNP3 3.1.2 `--network none` peer — Docker daemon
unavailable on this host (`DAEMON_UNAVAILABLE`). No claim is made from the
in-repo loopback gate about third-party interoperability.

**Note on toolchain:** the container had no Go toolchain, so Go 1.22.12 was
installed from the official `go.dev` distribution to run the required gates.
No repository file depends on that install.

## Evidence boundaries unchanged

EG628 Windows Tester pcap/log, offset-38 attribution, full class 0+1+2+3 success,
adapter revision and build pin remain **NOT_VERIFIED/MISSING**. The original
Class 0+1+2+3 READ offset-38 incident remains **OPEN** — only the captured CROB
retry passed. G41 stays **UNSUPPORTED**; `active_work/supported-profile.md` blob
`3e592e77ad6e31d5492c609c5fdd62382505904b` unchanged. `main` `82414d6`
untouched.

## Disposition

`AO008_SHORT_READ_AND_PULSE_WIRE_READY_FOR_INDEPENDENT_CHECK`. `CURRENT: AO-008`,
`QUEUED: NONE`, `Status: TASK_IN_PROGRESS`. Non-force checkpoint to `work`;
**STOP for independent read-only CHECK**. No PR, merge, release or deployment.
