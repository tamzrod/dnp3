# AO-008 EG628 donor interoperability patch — verification note

Date: 2026-11-19
Branch: `work`
CURRENT: AO-008 (bounded corrective continuation; **not** an archive, promotion,
merge or deployment)
Base `work` HEAD at start: `0bbbc4e` (`origin/work` `0bbbc4e`)

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
