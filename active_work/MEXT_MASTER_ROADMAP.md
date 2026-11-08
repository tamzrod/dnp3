# DNP3 Master External-Use Roadmap (MEXT)

**Status:** MEXT-035 COMPLETE (historical series); post-gate work is unpromoted  
**Created:** 2026-08-12  
**Primary objective:** Master usable against real/external outstations with IEEE 1815-correct wire behavior for the v0 profile  
**Series:** MEXT (Master External)  
**Total tasks:** 40  
**External MVP complete at:** MEXT-035  
**Next task:** NONE — MEXT-040/041/042 remain unpromoted post-gate backlog

This is the historical task plan for the **completed external Master MVP** series.
It does not authorize execution. New implementation requires a separately approved
KDE design (where applicable) and a human-promoted CURRENT CWAL task in
`workflow/active_work/README.md` and `handoff.md`.

Do **not** infer task authorization from historical roadmap entries.  
Do **not** invent requirements beyond IEEE 1815 wire correctness for the v0
profile and external acceptance criteria below.  
Do **not** expand to Level 3/4 or post-gate features without new architecture
approval and a human-promoted CWAL package.  
The prior MEXT task-by-task history is recorded in `active_work/handoff.md`.

**Archived series:** `active_work/archive/DNP3_MASTER_ROADMAP_v1_internal_mvp.md`  
(Internal Master MVP complete at DNP3-056; external interop was never claimed.)

**Profile lock:** `active_work/supported-profile.md` remains authoritative for
the v0 object subset (G1V1, G30V1, G20V1, G12V1 Direct-Operate, TCP only).

---

## 1. CURRENT DNP3 LIBRARY STATE (post-MEXT-035 / PR #15)

Native Go library. The internal Master MVP and in-repo real-TCP v0 Master path are verified:

- **DLL / TL / AL:** Frame, CRC, fragment, APDU, object-header model present.
- **Master public API:** `Connect`, `Disconnect`, `Close`, `Read`, `IntegrityPoll`, `Operate`, `LastIIN`, error taxonomy, optional logger.
- **v0 subset:** Class-0 G1V1 / G30V1 / G20V1 + Direct-Operate G12V1.
- **Verification:** Simulator/unit/integration/race coverage plus the real-TCP MVP gate (`scripts/verify-external-mvp.sh`, required Tiers 1–3) recorded PASS at MEXT-035. PR #15 also removed the `go vet ./...` unreachable return and applied repository-wide `gofmt`; its merged `main` CI passed. These in-repo tests are **not** independent third-party interoperability proof.
- **Outstation:** Sufficient for in-repo TCP loopback; test peer for MEXT.

### Residual disposition after MEXT-035

| ID | Status | Evidence / limitation |
|----|--------|------------------------|
| R1 | Resolved for in-repo v0 path | MEXT-012/013 real-TCP direct-operate success and rejection cases |
| R2 | Resolved for v0 wire subset | MEXT-011 CROB IEEE 1815 bitfield goldens |
| R3 | Resolved for tested v0 subset | MEXT-014/015 multi-header parse and poll; MEXT-033 gate Tier 3 |
| R4 | **Open** | No independent VEC-01/PCAP or third-party-stack proof; third-party interoperability cannot be claimed |
| R5 | Mitigated, ongoing | MEXT-034/035 scoped README claims; continue forbidding production-readiness/conformance overclaims |

---

## 2. VERIFIED SCOPE AND REMAINING GAPS

| Area | Post-MEXT evidence | Remaining boundary |
|------|--------------------|--------------------|
| Connect + link handshake | In-repo real-TCP gate and golden frames | Independent peer not verified |
| Class-0 read G1/G20/G30 | In-repo real TCP and multi-header coverage | Other object groups unsupported |
| Direct-Operate G12V1 | In-repo real-TCP success/rejection matrix | No third-party control confirmation |
| CROB wire codes | IEEE 1815 bitfield goldens for v0 | Broader profile not claimed |
| Retry / timeout / FCB / CON | Focused tests, reconnect/rogue-peer negatives | Third-party stress not verified |
| Unsupported features | Explicit rejection per supported profile | Feature expansion needs separate authorization |
| Independent capture (VEC-01) | **Not available** | R4 open; third-party interoperability not claimed |

---

## 3. TARGET EXTERNAL MASTER

```
External consumer (Ingestor / test / SCADA tool)
        ↓
pkg/dnp3/master  (public API only)
        ↓
internal master + AL + TL + DLL
        ↓
TCP
        ↓
Real or third-party outstation
```

Constraints: pure protocol library; v0 profile only until gate; IEEE 1815 wire for that path; clear reject outside profile.

---

## 4. EXTERNAL ACCEPTANCE GATE

The **in-repo v0 external-use MVP gate** was completed at MEXT-035, based on the six requirements below. This does **not** establish third-party interoperability, full IEEE 1815 compliance, or production SCADA readiness:

1. `./scripts/verify-mvp.sh` exit 0
2. `./scripts/verify-external-mvp.sh` exit 0 (real TCP master↔outstation)
3. CROB codes match IEEE 1815 bitfield goldens
4. Operate does not ControlTimeout on valid success APDU from in-repo outstation over TCP
5. Class-0 multi-header response returns all G1/G20/G30 points
6. README external claims match tests

**External MVP completed at MEXT-035 (in-repo real TCP; R4 remains open).**

See also `active_work/external-acceptance.md`.

---

## 5. MICRO-TASK ROADMAP

### Phase 0 — Archive & baseline

### MEXT-001 — Archive v1 series + open MEXT handoff
- **Purpose:** Freeze internal-MVP series; make MEXT recoverable.
- **Prerequisites:** none
- **Scope:** `active_work/` docs only
- **Work:** Archive copies under `active_work/archive/`; handoff points at MEXT.
- **Tests:** none (docs)
- **Acceptance:** Archive paths present; handoff NEXT progresses after done.
- **Commit message:** `docs: archive v1 micro-tasks; open MEXT series`
- **Handoff:** Archive locations + MEXT entry point.
- **Stop condition:** Docs on main.

### MEXT-002 — Record residuals R1–R5 in supported-profile
- **Purpose:** Single list of external blockers.
- **Prerequisites:** MEXT-001
- **Scope:** `active_work/supported-profile.md`
- **Work:** Add “External residuals” section with R1–R5.
- **Tests:** none
- **Acceptance:** Residuals visible without git history.
- **Commit message:** `docs(profile): record external residuals R1-R5`
- **Handoff:** Residual table.
- **Stop condition:** Profile updated.

### MEXT-003 — Baseline commit hash + verify-mvp lock
- **Purpose:** Pin internal baseline before external changes.
- **Prerequisites:** MEXT-001
- **Scope:** handoff / scripts README
- **Work:** Record HEAD sha where verify-mvp.sh is green.
- **Tests:** Re-run `./scripts/verify-mvp.sh`
- **Acceptance:** Exit 0; sha in handoff.
- **Commit message:** `docs: pin internal MVP baseline for MEXT`
- **Handoff:** Baseline sha.
- **Stop condition:** Gate green + documented.

### MEXT-004 — External acceptance criteria checklist file
- **Purpose:** Gate definition file.
- **Prerequisites:** MEXT-001
- **Scope:** `active_work/external-acceptance.md`
- **Work:** Ensure six gate items present (may already exist).
- **Tests:** none
- **Acceptance:** File matches roadmap §4.
- **Commit message:** `docs: external acceptance checklist`
- **Handoff:** Checklist path.
- **Stop condition:** File on main.

### MEXT-005 — README external-claim lock
- **Purpose:** Prevent over-claiming while MEXT in progress.
- **Prerequisites:** MEXT-002
- **Scope:** `README.md` status block only
- **Work:** Explicit: internal MVP only; external claim blocked until MEXT-035.
- **Tests:** none
- **Acceptance:** No false third-party interop claim.
- **Commit message:** `docs(readme): lock external interop claim until MEXT-035`
- **Handoff:** README wording.
- **Stop condition:** Docs only.

### Phase 1 — Spec-correct wire fixes

### MEXT-010 — CROB control-code IEEE 1815 bitfield audit
- **Purpose:** Resolve R2 (fact-finding).
- **Prerequisites:** MEXT-003
- **Scope:** Read-only audit of encode path + constants
- **Work:** Table current vs IEEE 1815 bitfield; write `active_work/crob-code-audit.md`.
- **Tests:** none required
- **Acceptance:** Audit file committed.
- **Commit message:** `docs(crob): audit control-code values vs IEEE 1815`
- **Handoff:** Audit file.
- **Stop condition:** Audit committed.

### MEXT-011 — CROB control-code correction + golden vectors
- **Purpose:** Fix R2.
- **Prerequisites:** MEXT-010
- **Scope:** CROB encode/constants + tests
- **Work:** Align codes to IEEE 1815; golden hex; keep public API stable if possible.
- **Tests:** Golden encode; Operate simulator tests updated.
- **Acceptance:** Goldens match 1815; verify-mvp.sh green.
- **Commit message:** `fix(control): CROB codes to IEEE 1815 bitfield`
- **Handoff:** Code map + test names.
- **Stop condition:** MVP gate green.

### MEXT-012 — Direct-Operate response: status object optional path
- **Purpose:** Fix R1 (ControlTimeout on valid success).
- **Prerequisites:** MEXT-003
- **Scope:** Master Operate response parse only
- **Work:** Success if (a) G12V1 status=0 or (b) matching SEQ + no error IIN when IIN-only success allowed. No false success on error IIN. No hang when complete APDU arrived.
- **Tests:** Canned: status present; absent+clear IIN; error IIN; truncated.
- **Acceptance:** No false ControlTimeout on complete success APDU.
- **Commit message:** `fix(master): operate response without mandatory status echo`
- **Handoff:** Success rules.
- **Stop condition:** Unit + verify-mvp green.

### MEXT-013 — Operate real-TCP vs in-repo outstation
- **Purpose:** Prove R1 fix on real TCP.
- **Prerequisites:** MEXT-012
- **Scope:** `test/integration` TCP Operate
- **Work:** Connect → Operate → assert success policy; document response shape.
- **Tests:** TCP operate integration.
- **Acceptance:** Green on real TCP loopback.
- **Commit message:** `test: operate success on real TCP outstation`
- **Handoff:** Observed response shape.
- **Stop condition:** Integration green.

### MEXT-014 — Multi-object-header Class-0 parse harden
- **Purpose:** Fix R3.
- **Prerequisites:** MEXT-003
- **Scope:** Response object parse for Read/IntegrityPoll
- **Work:** Advance through multiple headers; no point loss for G1+G20+G30 one APDU.
- **Tests:** Multi-header fixture; single-header regression.
- **Acceptance:** All points from one multi-header response.
- **Commit message:** `fix(al/master): multi-object-header Class-0 parse`
- **Handoff:** Parser rules.
- **Stop condition:** Tests + verify-mvp green.

### MEXT-015 — IntegrityPoll prefers single multi-group request
- **Purpose:** Remove per-group workaround as primary path.
- **Prerequisites:** MEXT-014
- **Scope:** IntegrityPoll
- **Work:** One Class-0/multi-group read primary; document or remove fallback.
- **Tests:** Full set from one exchange when peer allows.
- **Acceptance:** Documented; tests green.
- **Commit message:** `fix(master): IntegrityPoll single multi-header path`
- **Handoff:** Poll strategy.
- **Stop condition:** verify-mvp green.

### MEXT-016 — Qualifier allow-list vs 1815 for v0 path
- **Purpose:** Request qualifiers standard-correct.
- **Prerequisites:** MEXT-003
- **Scope:** Read/Operate builders
- **Work:** Lock 0x06/range/index goldens; reject others.
- **Tests:** Golden request headers.
- **Acceptance:** Goldens committed.
- **Commit message:** `test: lock v0 request qualifier goldens`
- **Handoff:** Golden paths.
- **Stop condition:** Tests green.

### MEXT-017 — Link handshake external frame vectors
- **Purpose:** Reset + link status byte vectors.
- **Prerequisites:** MEXT-003
- **Scope:** DLL/master handshake tests
- **Work:** Golden frames; Connect requires both exchanges.
- **Tests:** Vector-driven handshake.
- **Acceptance:** Mismatch fails Connect.
- **Commit message:** `test(dll): external-style link handshake vectors`
- **Handoff:** Fixture names.
- **Stop condition:** Tests green.

### MEXT-018 — Application SEQ + CON on solicited path audit
- **Purpose:** Spec continuity under external-like responses.
- **Prerequisites:** MEXT-003
- **Scope:** Master seq/confirm
- **Work:** CON=1 confirm; SEQ match; no advance on failed send.
- **Tests:** Fill gaps.
- **Acceptance:** Matches 1815 solicited rules for v0 path.
- **Commit message:** `test(master): SEQ and CON solicited path lock`
- **Handoff:** Gaps fixed or ticketed.
- **Stop condition:** Tests green.

### MEXT-019 — IIN table freeze vs 1815
- **Purpose:** No semantic drift before external claim.
- **Prerequisites:** MEXT-003
- **Scope:** `internal/al` IIN + docs
- **Work:** Freeze bit map; tests for critical bits.
- **Tests:** Encode/decode known masks.
- **Acceptance:** Table frozen; tests green.
- **Commit message:** `docs(al): freeze IIN bit map for external v0`
- **Handoff:** Table location.
- **Stop condition:** Docs + tests.

### Phase 2 — External interop harness

### MEXT-020 — VEC-01 capture fixture format
- **Purpose:** Structure for R4.
- **Prerequisites:** MEXT-004
- **Scope:** `active_work/testdata/external/`
- **Work:** Hex/PCAP-sidecar format + sample placeholder.
- **Tests:** Loader smoke if code added.
- **Acceptance:** Format documented; directory exists.
- **Commit message:** `test: VEC-01 external capture fixture format`
- **Handoff:** Format spec.
- **Stop condition:** Docs/fixtures in tree.

### MEXT-021 — verify-external-mvp.sh skeleton
- **Purpose:** Gate script exists.
- **Prerequisites:** MEXT-004
- **Scope:** `scripts/verify-external-mvp.sh`
- **Work:** Build + external TCP tests; fail-closed once tests exist.
- **Tests:** Script executable.
- **Acceptance:** Documented in scripts/README.
- **Commit message:** `test: add verify-external-mvp.sh skeleton`
- **Handoff:** Script path.
- **Stop condition:** Script committed.

### MEXT-022 — Real-TCP full MVP path test
- **Purpose:** Connect, IntegrityPoll, Operate, Close over TCP.
- **Prerequisites:** MEXT-013, MEXT-015
- **Scope:** `test/integration`
- **Work:** End-to-end; no simulator transport.
- **Tests:** Points + operate policy.
- **Acceptance:** Green locally and in script.
- **Commit message:** `test: real-TCP full MVP master path`
- **Handoff:** Test name.
- **Stop condition:** Integration green.

### MEXT-023 — Workbench external-smoke steps
- **Purpose:** Manual/CI smoke via workbench if present.
- **Prerequisites:** MEXT-022
- **Scope:** scripts or examples
- **Work:** Document or script; skip without failing unit packages if missing.
- **Tests:** Script or manual steps.
- **Acceptance:** Reproducible from README.
- **Commit message:** `test: workbench external smoke steps`
- **Handoff:** How to run.
- **Stop condition:** Docs/script present.

### MEXT-024 — Operate success/fail matrix on TCP
- **Purpose:** Status policy coverage.
- **Prerequisites:** MEXT-013
- **Scope:** Integration
- **Work:** Success, not-supported, timeout on drop.
- **Tests:** Table-driven.
- **Acceptance:** No false success; no false timeout on complete APDU.
- **Commit message:** `test: operate status matrix on TCP`
- **Handoff:** Matrix.
- **Stop condition:** Tests green.

### MEXT-025 — Reconnect + DeviceRestart IIN on TCP
- **Purpose:** Lifecycle under restart bit.
- **Prerequisites:** MEXT-022
- **Scope:** Integration
- **Work:** Drop/reconnect; DeviceRestart if raisable.
- **Tests:** No stuck state.
- **Acceptance:** verify-mvp still green.
- **Commit message:** `test: reconnect and DeviceRestart on TCP`
- **Handoff:** Behavior notes.
- **Stop condition:** Tests green.

### MEXT-026 — Negative: bad CRC / wrong address must not hang
- **Purpose:** External robustness.
- **Prerequisites:** MEXT-022
- **Scope:** Master receive path
- **Work:** Inject bad CRC/addr; bounded error.
- **Tests:** CRC + TCP cases.
- **Acceptance:** No deadlock.
- **Commit message:** `test: master rejects bad CRC/addr without hang`
- **Handoff:** Cases.
- **Stop condition:** Tests green.

### Phase 3 — Public consumer contract

### MEXT-030 — Freeze v0 public Master surface
- **Purpose:** Consumer-stable API.
- **Prerequisites:** MEXT-022
- **Scope:** `pkg/dnp3/master` docs only
- **Work:** Supported methods + reject list.
- **Tests:** none
- **Acceptance:** Godoc matches profile.
- **Commit message:** `docs(master): freeze v0 public API for external use`
- **Handoff:** API list.
- **Stop condition:** Docs committed.

### MEXT-031 — examples/master external-use example
- **Purpose:** Copy-paste consumer path.
- **Prerequisites:** MEXT-022
- **Scope:** `examples/master`
- **Work:** Connect → IntegrityPoll → Operate → Close; README.
- **Tests:** Compile/build.
- **Acceptance:** Builds; run steps documented.
- **Commit message:** `feat(examples): master external-use example`
- **Handoff:** Example path.
- **Stop condition:** Builds.

### MEXT-032 — Error taxonomy consumer notes
- **Purpose:** External debuggability.
- **Prerequisites:** MEXT-030
- **Scope:** docs
- **Work:** timeout vs NACK vs unsupported vs disconnect.
- **Tests:** none
- **Acceptance:** Doc linked from README.
- **Commit message:** `docs: master error taxonomy for external consumers`
- **Handoff:** Doc path.
- **Stop condition:** Docs committed.

### MEXT-033 — verify-external-mvp.sh full gate
- **Purpose:** Single external claim command.
- **Prerequisites:** MEXT-021, MEXT-022, MEXT-024
- **Scope:** `scripts/verify-external-mvp.sh`
- **Work:** Real-TCP MVP + operate matrix + multi-header; exit 0 only if green.
- **Tests:** Script.
- **Acceptance:** Exit 0 on clean tree.
- **Commit message:** `test: verify-external-mvp.sh full gate`
- **Handoff:** Script usage.
- **Stop condition:** Exit 0.

### MEXT-034 — README external claim (conditional)
- **Purpose:** Claim only after gate.
- **Prerequisites:** MEXT-033, MEXT-011, MEXT-015
- **Scope:** `README.md`
- **Work:** External v0 verified via script; not full IEEE 1815.
- **Tests:** none
- **Acceptance:** Wording matches reality.
- **Commit message:** `docs(readme): claim external v0 master path`
- **Handoff:** Claim text.
- **Stop condition:** Docs only.

### MEXT-035 — External MVP acceptance record
- **Purpose:** Close external MVP.
- **Prerequisites:** MEXT-033, MEXT-034
- **Scope:** supported-profile.md, handoff.md
- **Work:** Record gate run; EXTERNAL MVP COMPLETE.
- **Tests:** Re-run both verify scripts.
- **Acceptance:** Both exit 0; handoff records MEXT-035 complete.
- **Commit message:** `docs: record external Master MVP complete (MEXT-035)`
- **Handoff:** Gate record.
- **Stop condition:** Both scripts green.

### Phase 4 — Post-gate only (do not start early)

### MEXT-040 — Backlog stub: SBO profile
- **Prerequisites:** MEXT-035 — docs only, defer SBO.
- **Commit message:** `docs: defer SBO to post-external series`

### MEXT-041 — Backlog stub: event classes / unsolicited
- **Prerequisites:** MEXT-035 — docs only.
- **Commit message:** `docs: defer events/unsolicited post-external`

### MEXT-042 — Backlog stub: third-party stack capture
- **Prerequisites:** MEXT-035 — procedure placeholder.
- **Commit message:** `docs: third-party capture procedure placeholder`

---

## 6. DEPENDENCY / READY MAP

```
MEXT-001 → 002, 003, 004
002 → 005
003 → 010, 012, 014, 016, 017, 018, 019
010 → 011
012 → 013 → 024
014 → 015
004 → 020, 021
013 + 015 → 022 → 023, 025, 026, 030, 031
021 + 022 + 024 → 033
011 + 015 + 033 → 034 → 035
035 → 040, 041, 042
```

Parallel-safe after MEXT-003: 010, 012, 014, 016.

---

## 7. EXTERNAL MVP DEFINITION

**In scope:** TCP Master→one outstation; Class-0 G1/G30/G20; Direct-Operate G12V1 with 1815 CROB codes; Operate success on real TCP; lifecycle; verify-external-mvp.sh green.

**Out of scope until post-gate:** SBO, unsolicited, events, time sync, file transfer, SA, TLS, serial, full Level 2/3/4 cert, Ingestor/Kafka.

---

## 8. RISKS

| Risk | Mitigation |
|------|------------|
| Outstation response shape ≠ sim | MEXT-012/013 dual success path |
| CROB change breaks sim tests | Update goldens + sim together |
| Scope creep to full 1815 | Phase 4 docs-only |
| False external claim | MEXT-005 until MEXT-035 |
| Multi-header fix regresses single-header | Regression tests in MEXT-014 |

---

## 9. CURRENT EXECUTION AUTHORITY / NEXT STEP

The MEXT-001–035 sequence is historical and completed per the MEXT handoff.
MEXT-040/041/042 are post-gate **backlog stubs**, not automatically promoted
or approved for implementation. The two CWAL correction tasks DNP3-056
(unreachable return) and DNP3-057 (formatting) were merged through PR #15.

GitHub evidence: PR #15 merged on 2026-11-08 as `b38d7691aeb1adccf52fe665e368d4d77766977c`;
`main` CI for that merge completed successfully. See
`docs/verification/DNP3-056.md`, `docs/verification/DNP3-057.md`, and
`active_work/handoff.md` for detailed evidence and limitations.

**CWAL execution state (authoritative pointers are outside this historical roadmap):**

```
CURRENT: NONE
QUEUED: NONE
Status: WORK_PACKAGE_READY (completed correction package; no active implementation task)
NEXT TASK: NONE — human promotion and applicable KDE design approval required
```

No direct `main` push from an execution agent, independent PR creation, merge,
deployment, or live field control is authorized by this roadmap.

