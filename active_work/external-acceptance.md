# External Master Acceptance Checklist

**Series:** MEXT  
**Gate task:** MEXT-035  
**In-repo v0 external-MVP gate completed at MEXT-035.** This is not proof of independent third-party interoperability, full IEEE 1815 conformance, or production readiness. Historical MEXT-035 in-repo TCP acceptance and the later dated independent VEC-01 lab acceptance are recorded separately below.

## Commands

```bash
./scripts/verify-mvp.sh
./scripts/verify-external-mvp.sh   # required from MEXT-033 onward
```

## Checklist

- [x] `verify-mvp.sh` exit 0 (no internal regression) — MEXT-035 handoff; re-executed PASS under DNP3-058
- [x] `verify-external-mvp.sh` exit 0 (in-repo real TCP master↔outstation path, Tiers 1–3) — MEXT-035 and DNP3-058 PASS. At those dates the script's VEC-01 advisory tier was non-blocking and unbacked; independent VEC-01 v0 evidence was accepted later (2026-11-11, see below).
- [x] CROB control codes match IEEE 1815 bitfield golden vectors (MEXT-011) ✅
- [x] Operate does not ControlTimeout on valid success APDU from in-repo outstation over TCP (MEXT-012/013) — parse-side fix done in MEXT-012; proven on real TCP in MEXT-013 (`TestOperateRealTCPSuccess`)
- [x] Class-0 multi-object-header response returns all G1/G20/G30 points (MEXT-014/015; MEXT-033 Tier 3 in-repo gate)
- [x] README external claims match the tested in-repo v0 profile, excluding independent third-party interoperability and full IEEE 1815 conformance (MEXT-034/035)

## Evidence reconciliation — 2026-11-08

| Field | Verified historical record |
|-------|----------------------------|
| MEXT-035 gate | COMPLETE in `active_work/handoff.md` (original gate SHA not independently established in this reconciliation) |
| `verify-mvp.sh` | PASS at MEXT-035; re-executed PASS in DNP3-058 on `2afbeb5acf0fce9126edb122290b29ba3f532151` |
| `verify-external-mvp.sh` | PASS Tiers 1–3 at MEXT-035 and DNP3-058, **not** third-party VEC-01 proof |
| Additional Go checks | DNP3-058 executed Go 1.22.12 build, vet, unit, race, CI-equivalent coverage gates with PASS evidence in `docs/verification/DNP3-058.md` |
| MEXT-040 | Documentation-only SBO deferral completed; feature **NOT SUPPORTED** |
| MEXT-041 | Documentation-only event/unsolicited deferral completed; feature **NOT SUPPORTED** |
| PR #17 | Merged into `main` at `95ac10289ead46c09264a9da2bb877f0e0507410`; latest main CI success |
| R4 / VEC-01 | **OPEN** — no independent PCAP/third-party-stack interoperability proof |

This synchronization is a documentation reconciliation, **not** an assertion that the old MEXT-035 checks were newly rerun on this documentation commit. See historical handoff and DNP3-058 evidence for precise execution contexts.

The `R4 / VEC-01` row above is the **2026-11-08 historical record** (R4 was then open for absence of independent evidence). It is superseded in current status by the dated independent VEC-01 acceptance record below; the historical wording is preserved here as history and is not rewritten.

## Independent VEC-01 lab acceptance — 2026-11-11 (scoped, v0 path only)

This is a **separate, dated, scope-specific** acceptance record. It is distinct
from the historical MEXT-035 in-repo TCP acceptance above, and it does **not**
assert IEEE 1815 conformance, production/SCADA readiness, release readiness, or
interoperability with any peer other than the one tested.

| Field | Record |
|-------|--------|
| Accepted checkpoint | `work` `6c297df0fb7c702f66818e506008e00e9d716ea4` (owner-accepted independent read-only CHECK) |
| Peer | OpenDNP3 **3.1.2** (peel `26b4c01e4839bbbda8866655e086471c4917ee53`; tarball `183cc29222c3cb58099a9753c43defed5cb8677fda985c3f88e38b5c19a36ff2`), independently implemented C++ stack |
| Setup | Isolated `network_mode: none` container, two private namespaces joined by one lab-only veth; no external route/DNS/NAT/ports; memory-only simulated CROB handler |
| Scope | Class-0 G1V1/G30V1/G20V1 read; one G12V1 Direct-Operate CROB; one negative payload-block CRC case |
| Observed | `BI[0]=true`; `AI[0]=2300` ONLINE; `CT[0]=42` ONLINE; CROB `BO[0]` LATCH_ON protocol status SUCCESS, peer readback `state=true`; negative: last payload-block CRC bad (header valid), master failed closed with no control |
| Independent decode | tshark 4.4.19 (distinct dissector) plus pinned peer log |
| Durable evidence | `active_work/testdata/external/vec01-20261111/` — two original PCAPs, packet-derived `.vec` vectors, `SHA256SUMS.txt`, `README.md` |
| Verification | `docs/verification/VEC-01-LAB-EXEC-001.md` §§50–55 (fresh independent clone at `3442518` matched all SHA-256 and re-decoded identically) |
| Disposition | `R4 RESOLVED — VEC-01 v0 independent-lab evidence accepted` (documentation-only, scoped; see `active_work/supported-profile.md`) |

**Limitations (must not be suppressed).**

- One peer and one version in one isolated lab. This does not establish
  interoperability with other vendors, versions, or profiles, and does not
  demonstrate production performance or reliability.
- Unresolved defect: the master mis-decodes **G1V2** (Binary Input With Flags)
  by consuming the flags octet as the point value. Outside the locked v0 G1V1
  profile; **not** fixed and **not** claimed supported. It remains a visible
  peer-compatibility caveat.
- No new object groups/variations, SBO, unsolicited/events, TLS, authentication,
  serial, time sync, freeze, or file transfer are supported by this acceptance.
- This record is **not** a release gate, branch integration, or general
  certification. `main` integration, CI on `work`, release/version decisions,
  and any broader support-matrix promotion remain separate, unauthorised items
  (see `docs/verification/VEC-01-LAB-EXEC-001.md` §56.3).
