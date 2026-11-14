# AO-005 PRECHECK — Group 41 analog output promotion contract (DRAFT)

> **Status: DRAFT / UNPROMOTED / NON-AUTHORITATIVE.**
> This document is a **precheck for owner and KDE review only**. It is **not** an
> active task, is **not** queued, creates **no** promotion authority, sets **no**
> `SUPPORTED` state, and does **not** modify `active_work/supported-profile.md`.
> Nothing here authorizes code, tests, protocol execution, a PR, a merge, a
> release, or any field/NGCP/SCADA command. Group 41 remains **UNSUPPORTED** in
> the published v0 profile until a separate human owner decision (with any
> applicable KDE approval) promotes it.
>
> Prepared 2026-11-14 on branch `work`. Every row below is grounded in the
> current `main` source, `docs/verification/AO-004.md` appendices and the
> retained `active_work/testdata/external/ao004-*` manifests. No new tests or
> protocol executions were run to produce this draft.

## 1. Candidate capability (proposed only)

| Item | Proposed value |
|---|---|
| Group / variations | G41 V1 signed int32, V2 int16, V3 float32, V4 float64 |
| Object count | one point per request |
| Operation | **Direct Operate WITH response** only |
| Excluded | Select-Before-Operate; Direct Operate No Response; G41 via WRITE; multi-point |
| Status contract | explicit per-point success / non-success; never a false SUCCESS |
| Field target | none; NGCP point mapping, scale and units are **unknown** and out of scope |

This candidate mirrors what the source already implements as a rejected-default
path; it is a *proposal* to promote, not a statement that it is supported.

## 2. Recorded independent-peer evidence (from AO-004)

Peer: pinned **OpenDNP3 3.1.2**, commit `26b4c01e4839bbbda8866655e086471c4917ee53`,
tarball sha256 `183cc29222c3cb58099a9753c43defed5cb8677fda985c3f88e38b5c19a36ff2`.

| Direction | Scenario | Result | Source |
|---|---|---|---|
| OpenDNP3 3.1.2 master → Go outstation | G41V1–V4 idx0 | **PASS** 4/4 SUCCESS, exact float bits (`0x3fc00000`, `0x3ff8000000000000`) | Appendix T; §U.2/U.3 |
| OpenDNP3 3.1.2 master → Go outstation | G41V1–V4 idx1 | **PASS** OUT_OF_RANGE, `status=0x0C`, zero writes, IIN `0x0000` | Appendix T |
| Go master → OpenDNP3 3.1.2 C++ outstation | G41V1–V4 round-trip | recorded positive; **FAIL/DIVERGENCE** only on signed-zero/subnormal (see §4) | Appendix H |
| Unit (internal + bridge echo/mapping) | G41V1–V4 idx0/idx1 | **PASS** (`UNIT_ONLY`) | Appendix T tests |

Retained artifacts and hashes (verified this invocation with `sha256sum -c`):

- `active_work/testdata/external/ao004-20261112/` — sanitized pcap
  `ao004-20261112-isolated-g41v1-v4.pcap`
  sha256 `d3cd3eaa6aeb6b584a670d632d68eb9100aee6975ab4ef485e30bf853a819d37`,
  plus 8 packet-derived `.vec` and `SHA256SUMS.txt` (9/9 OK).
- `active_work/testdata/external/ao004-negative-20261113/` — 9 pcaps
  (8 positive/non-success controls + 1 peer-unavailable), 9 `.vec`,
  `SHA256SUMS.txt` (all OK).
- Historical AO-004 originals remain `ORIGINAL_PCAP_UNAVAILABLE`; the committed
  captures are **new** captures, not recovered originals.

Classification caution: the idx0/idx1 rows are **scoped interoperability
evidence** on the pinned peer only. They do not establish general IEEE 1815
conformance, third-party certification, production readiness, or NGCP mapping.

## 3. Remaining independent-wire gaps (must be resolved or explicitly limited)

| Gap | Independent-wire status | Unit status |
|---|---|---|
| Malformed G41 (wrong qualifier/count/variation/length/extra object, payload-CRC corruption) | **NOT_RUN** — pinned public master API cannot emit malformed objects; injection disallowed | `UNIT_ONLY` (internal/outstation + internal/master G41 tests) |
| Mid-command lost response / disconnect / timeout | **NOT_RUN** — no deterministic single-shot fault injector available | covered by Go timeout/disconnect unit suites |
| Concurrent commands | **NOT_RUN** | asserted by unit tests |
| No blind retry / quarantine after an unknown outcome | **NOT_RUN** on the wire | behavior asserted in unit tests |
| Reverse direction full matrix | partly recorded (Appendix H), not re-run under AO-004 | — |

**Bounded acceptance decision required.** Promotion must either (a) require
specific isolated negative-wire proof for the malformed and mid-command cases, or
(b) obtain owner/KDE approval of an explicitly limited profile claim. Unit
evidence must **never** be silently equated to independent-wire proof.

## 4. Float divergence — fail-closed, scope proposal

On the Go-master → OpenDNP3 3.1.2 path, the peer normalizes the echoed G41V3/V4
setpoint: `-0.0` → `+0.0`, and the smallest subnormal → `0` (Appendix H; cause is
peer-side ser4cpp float (de)serialization). The Go master correctly fails closed
on the raw-bit mismatch (`ErrControlOutcomeUnknown`); no production float/parser/
exact-bit/status code was weakened.

Proposed handling for any promotion: keep Go exact-bit validation and
fail-closed behavior unchanged, and **explicitly exclude** signed-zero and
subnormal setpoints from the claimed interoperability scope (a documented
limitation), rather than weakening the parser to match the peer.

## 5. Public API contract mismatch (blocking for promotion)

- `pkg/dnp3/master/client.go` `Operate` currently admits **G12V1 and G41 V1–V4**
  Direct-Operate-with-response (fail-closed object gate), while
  `active_work/supported-profile.md` mandates **`Reject`** for
  "Command handler analog controls … Group 41 and related variations".
- `pkg/dnp3/outstation/server.go` exposes
  `CommandHandler.HandleAnalogCommand` (v0 minimal handler returns
  not-supported).
- `pkg/dnp3/types/commands.go` `NewAnalogControl(index, value float64, …)` still
  hardcodes `Group 41, Variation 1` with a legacy/incorrect label; it is not
  automatically promotable and must not be treated as the verified V1 path
  (`NewAnalogInt32Control` is the verified V1 constructor).

A promotion therefore needs a **separate, later implementation gate** to
reconcile the default v0 rejection with any explicitly supported G41 capability,
including which constructor(s) are public and what the default `Operate` gate
admits.

## 6. Proposed future approval / verification gates

Before any promotion (all still requiring separate human owner authorization):

1. Owner/KDE approval of the exact G41 subset and its explicit exclusions.
2. Public outstation analog-handler behavior defined and tested (status mapping,
   echo of index/variation/value, no false SUCCESS).
3. Single-shot / quarantine semantics for a control with an unknown outcome,
   wire-proven where the gap table requires it.
4. Negative tests: malformed, wrong qualifier/count/variation, truncation,
   timeout, disconnect, concurrency — with the independent-wire vs `UNIT_ONLY`
   distinction stated.
5. Full Go gates: `gofmt -l` clean, `go vet ./...`, `go build ./...`,
   `go test ./... -count=1`, `go test -race ./... -count=1`,
   `bash scripts/verify-external-mvp.sh`, `bash scripts/cwal-check.sh`,
   `git diff --check`.
6. Reproduced pinned OpenDNP3 3.1.2 isolated `--network none` runs in both
   directions with durable PCAP/`.vec` and SHA-256 provenance.

Explicit exclusions for any promoted claim: no generalized IEEE 1815 conformance;
no production or field/NGCP/SCADA readiness; no NGCP point/scale/units mapping;
no SBO or Direct-Operate-No-Response; no G41 via WRITE; no multi-point; no
weakening of the exact-bit float fail-closed behavior.

## 7. Decision requested from the owner

1. Whether to promote a G41 subset at all, and if so the exact variations and
   operate mode.
2. Whether promotion requires new isolated negative-wire proof (gap table §3) or
   accepts an explicitly limited claim.
3. How to reconcile the `Operate` object gate, `CommandHandler.HandleAnalogCommand`
   and `types.NewAnalogControl` with the profile (separate implementation gate).
4. Whether the signed-zero/subnormal limitation (§4) is acceptable as a documented
   exclusion.

Until those decisions and any applicable KDE approval are recorded, Group 41
stays **UNSUPPORTED**, AO-005 stays **unpromoted**, and
`active_work/supported-profile.md` stays byte-identical.
