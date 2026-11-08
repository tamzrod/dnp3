# Independent Third-Party DNP3 Capture Procedure (VEC-01 / residual R4)

**Status: PLANNING ONLY — VEC-01 NOT EXECUTED / R4 OPEN / THIRD-PARTY INTEROP NOT PROVEN.**
This document is a procedure placeholder. It defines how an independent
third-party interoperability campaign *would* be prepared, executed and
reviewed. It is **not** an executed test, it contains **no** capture results, and
it does **not** close R4 or establish any interoperability claim.

Owner authorization for an actual campaign is required separately; this
document grants none.

---

## 1. Purpose and non-goals

The repository's v0 Master path is verified only against its own in-repo TCP
loopback and unit/golden tests (`./scripts/verify-external-mvp.sh`, Tiers 1–3).
Those tests use this repository's own outstation and are **not** independent
evidence. Residual **R4 / VEC-01** remains open precisely because no capture
from a separately implemented DNP3 stack exists.

This procedure defines the reproducible steps to obtain such evidence for the
**v0 profile only** (`active_work/supported-profile.md`): one TCP connection,
Class-0 static reads of Group 1 Variation 1 (Binary Input), Group 30 Variation 1
(Analog Input) and Group 20 Variation 1 (Counter), and one Direct-Operate
Group 12 Variation 1 CROB, over standard DNP3 link framing with LSB-first
multi-octet fields.

**Non-goals (explicitly out of scope):**

- No production, field or live-SCADA peer. No real control of physical outputs.
- No implementation or modification of a test harness, integration test, code,
  script, CI job or dependency under this task.
- No serial, TLS, secure authentication, time sync, file transfer, device
  attributes, restart/delay/freeze, select-before-operate, unsolicited delivery
  or event classes, and no object group/variation outside the v0 subset.
- No claim of IEEE 1815 certification, production readiness or universal DNP3
  compatibility. A single capture at most evidences the tested v0 path against
  one peer version.

---

## 2. Independent third-party peer selection and provenance

The peer must be an **independently implemented** DNP3 outstation/stack — i.e.
not authored by this repository's maintainers and not a second copy of this
library. A second instance of `pkg/dnp3/outstation`, its examples, or any
wrapper around this repository does **not** qualify and must be recorded as
ineligible if proposed.

Selection record (capture one row per candidate; use the first eligible one):

| Field | Value to record |
|-------|-----------------|
| Product / project name | |
| Implementer / vendor / organization | |
| Implementation lineage | Confirm it is not derived from `tamzrod/dnp3` |
| Version / build / commit | Exact release string and build identifier |
| License | SPDX id and source of the license text |
| Provenance | Where the binary/source was obtained (URL, package, artifact hash) |
| Outstation capability set | Confirm it can serve v0 static groups + G12V1 Direct-Operate |
| Eligibility verdict | ELIGIBLE / INELIGIBLE with one-line reason |

Also record the **master-side implementation** under test: this repository's
library at the exact executed commit SHA (see §10), built with the recorded Go
version, using only the v0 public surface (`pkg/dnp3/master`).

Independence must be reviewer-verifiable from the recorded provenance, not
asserted.

---

## 3. Isolated non-production test topology

- Two hosts (or one host plus one isolated network namespace / container) on a
  private, non-routed segment with no path to production SCADA networks.
- The peer's physical/simulated outputs are **simulated, non-operational
  endpoints** (for example loopback indicators or a simulator channel) so a
  Direct-Operate has no real-world effect. State this explicitly in the record.
- No shared credentials, no field-bus, no gateway into any operational network.
- Single logical link: one master process ↔ one outstation process.
- Prefer a capture point that sees the actual TCP stream (e.g. an on-host
  capture, or a transparent bridge) rather than application-level logging.

Topology record: host roles, OS/version, container or namespace isolation
mechanism, capture point location, and confirmation that no production network
was reachable during the run.

---

## 4. IP / port / address configuration

Record the exact values used; they must match the capture.

| Item | Value to record |
|------|-----------------|
| Master IP / interface | |
| Outstation IP / interface | |
| TCP port (DNP3 default 20000; record if changed) | |
| Master DNP3 link source address | |
| Outstation DNP3 link destination address | |
| Master DNP3 link destination address | |
| Outstation DNP3 link source address | |
| Addresses per direction | Both directions must be internally consistent |
| TCP keepalive / idle settings | |

Addresses are configured symmetrically: the master's source address equals the
outstation's configured master address, and the outstation's source address
equals the master's configured outstation address. A mismatch is a negative case
(§7), not a normal run.

---

## 5. v0-only operation matrix

Execute these steps in order on one connection. Each step names the expected
wire behavior and the artifact to keep. The matrix is deliberately limited to
the v0 supported profile.

| # | Step | Expected v0 behavior | Keep |
|---|------|----------------------|------|
| 1 | TCP connect | Connection established; no TLS handshake | Capture segment |
| 2 | Link reset | Master emits Reset Link Stations (control function code as in the golden `active_work/testdata/link-reset-link-stations.hex`); peer ACKs | Frames + decoded fields |
| 3 | Link status | Master requests link status; peer returns valid link status | Frames |
| 4 | Class-0 integrity read | Master sends one all-objects Class-0 request (G60V1, qualifier 0x06); peer returns static G1V1 / G30V1 / G20V1 | Frames + decoded points |
| 5 | Multi-header check | If the peer packs multiple object headers in one APDU, all G1/G20/G30 points must be decoded (no point loss) | Frames + decoded point set |
| 6 | Direct-Operate G12V1 | Master sends one Direct-Operate CROB (qualifier 0x00 index8, count 1) with a v0 IEEE 1815 control code (0x01/0x02/0x04/0x08/0x10/0x80); peer returns the command result | Frames + application status |
| 7 | Clean close | Master disconnects; TCP closes without a stuck state | Capture segment |

Only step 6 issues a control, and only against the non-production simulated
output from §3 with prior separate permission. No other control, select, freeze,
time-sync or event operation is permitted.

---

## 6. Capture collection, clock correlation and artifact format

- Capture the raw TCP stream with a standard packet capture tool (e.g. a
  `pcap`/`pcapng` export) and, in parallel, record the master's application-level
  diagnostics with a monotonic clock.
- Correlate capture timestamps to the application log by a shared monotonic
  reference (e.g. capture start epoch and a matching application start epoch),
  and record the offset and its uncertainty.
- Convert the capture into the repository's reviewable `.vec` fixture format
  (see `active_work/testdata/external/FORMAT.md`): a metadata header (`id`,
  `title`, `source`, `capture`, `role`, `master-addr`, `outstation-addr`,
  optional `pcap`, `notes`) followed by ordered `@ <direction> <layer>` frame
  records in hex.
- Produce **two** artifacts:
  1. the **original capture** (raw `.pcap`/`.pcapng` plus the derived `.vec`),
     retained unmodified with a recorded hash;
  2. a **redacted reproducible sidecar** (`.vec`) with any sensitive values
     removed (host names, IPs if required, operator identifiers), but with all
     DNP3 protocol bytes preserved exactly. State every redaction applied.
- The `.vec` `id` must equal its filename stem; a PCAP sidecar uses the same
  basename.

---

## 7. Wire inspection checks

An independent reviewer inspects the artifacts against the following; each is a
checkable assertion, not an interpretation:

1. **Framing** — each link frame begins with the DNP3 start bytes `05 64`;
   length, control, and both addresses decode consistently.
2. **CRC** — one CRC covers the complete link header, and one CRC covers each
   16-octet payload block *including the final partial block*; all CRCs validate.
3. **LSB-first fields** — multi-octet fields (point indices, analog/counter
   values, CROB times, addresses) are little-endian; verify byte order on at
   least one known value.
4. **Payload bound** — no link payload exceeds 250 octets.
5. **Object headers** — the Class-0 response decodes to the v0 groups/variations
   (G1V1, G30V1, G20V1) with a supported qualifier; the Direct-Operate request
   decodes to G12V1, qualifier 0x00, count 1, with a valid 1815 control code.
6. **SEQ / CON** — application sequence numbers advance 0–15 with wraparound on
   the solicited path; CON (confirm) is set as required and the peer's response
   carries the matching sequence number.
7. **Application status** — the Direct-Operate result is read from the protocol
   command status (or, for an IIN-only response, from a clear/error IIN per the
   R1 policy), **never** from transport acceptance alone. Ambiguous outcomes are
   recorded as ambiguous.

---

## 8. Deterministic negative cases

Each negative case is run deliberately and its bounded outcome recorded:

| Case | Method | Required bounded outcome |
|------|--------|--------------------------|
| Bad CRC | Corrupt a payload CRC in a captured frame and replay (or use a peer that emits a bad CRC) | Master rejects with a bounded CRC error; no hang, no false success |
| Mismatched address | Peer configured with a wrong source/destination address | Handshake fails with a bounded address error; no hang |
| Dropped connection | Abruptly close the TCP socket mid-exchange | Master surfaces a classified disconnect/timeout within the configured deadline; no deadlock |
| Command reject | Peer rejects the Direct-Operate | Master classifies failure from protocol status/IIN; **never** ControlSuccess, and never a blind retry when the outcome is unknown |
| Capture incomplete | Capture stops mid-run | Artifact marked incomplete; the run is **rejected** as evidence (see §10) |

A negative case that hangs, deadlocks, or is retried blindly is a **finding**,
not a pass.

---

## 9. Bounded timing and safe stop

- Define a wall-clock budget per step and for the whole run; record it up front.
- On any timeout, hang, unexpected control outcome, or loss of isolation from
  production networks: stop immediately, close the connection, do not retry
  blindly, and mark the run ABORTED with the reason.
- The master retry/timeout policy must be the recorded v0 configuration; no
  manual retry of a control whose previous outcome is unknown.
- No step may be repeated to "make it pass" without recording the earlier
  attempt and its outcome.

---

## 10. Evidence provenance and integrity

Record for every artifact and run:

| Field | Value |
|-------|-------|
| Executed master SHA | Exact `tamzrod/dnp3` commit under test |
| Go toolchain | `go version` string used to build the master |
| Peer stack / version | From §2 |
| Peer configuration | Full config used (addresses, port, output mapping) |
| Capture tool + version | e.g. packet capture tool and version |
| Analysis tool + version | Any decoder used |
| Timestamps | Run start/end, capture start, clock offset |
| Artifact hashes | SHA-256 of the original `.pcap`/`.pcapng`, the derived `.vec`, and the redacted sidecar |
| Redactions | Exact list of what was removed and why |

Integrity rules:

- Hashes must be computed from the retained artifacts and be reproducible by a
  reviewer from the retained files.
- **No artifact may be fabricated, hand-edited, or synthesized from the
  specification.** A `.vec` whose frames were not observed on a real wire is
  `synthetic-from-spec` and is **not** VEC-01 evidence (the existing
  `sample-vec01-placeholder.vec` is such a placeholder).
- An incomplete or interrupted capture is recorded as incomplete and is **not**
  accepted as evidence.

---

## 11. Reviewer-independent acceptance / rejection mapping

A reviewer who did not perform the run maps the evidence as follows:

| Condition | Verdict |
|-----------|---------|
| Eligible independent peer + complete capture + all §7 checks pass + all §8 negatives bounded + provenance complete | **ACCEPT** (v0 path evidenced against this peer/version only) |
| Peer ineligible (derived from this repo) | **REJECT** — not independent |
| Any §7 check fails, or an §8 negative hangs/deadlocks | **REJECT** — record the finding |
| Capture incomplete, artifact synthetic, or provenance/hashes missing | **REJECT** — insufficient evidence |
| Ambiguous control outcome treated as success | **REJECT** — false success |

Acceptance is scoped: it may support "the v0 path interoperated with peer X
version Y under this configuration." It never supports a claim of IEEE 1815
certification, production readiness, or universal compatibility.

---

## 12. Authorization gates for a future actual campaign

The following require **separate, explicit owner approval** before any execution
(this document is not that approval):

1. Promoting a new bounded task to actually run the campaign.
2. Any peer that could drive a real output; simulated non-production outputs
   with written permission are the only permitted control target.
3. Any capture that touches a network shared with operational systems.
4. Publishing any capture or claim derived from it.

---

## 13. Limitations and non-claims

- This is a **planning document only**. VEC-01 is **NOT EXECUTED**; residual
  **R4 remains OPEN**; third-party interoperability is **NOT PROVEN**.
- The repository's own `./scripts/verify-external-mvp.sh` PASS covers only
  in-repo Tiers 1–3 (real-TCP MVP, operate matrix, multi-header) and is never
  VEC-01 evidence.
- No IEEE 1815 conformance, production-readiness, or universal-compatibility
  claim is made or implied.
- A single accepted capture evidences the tested v0 path against one peer
  version only; it does not generalize to other versions or stacks.
