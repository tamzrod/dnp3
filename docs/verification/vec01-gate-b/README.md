# VEC-01 Gate B reviewer packet (repository-resident)

**Package:** `VEC-01-LAB-EXEC-001`  ·  **Branch:** `work`  ·  **Status:** `TASK_BLOCKED`
**Packet status:** `GATE_B_REVIEW_PACKET_READY` (provisional)  ·  **Sign-off:** `GATE_B_SIGNOFF_PENDING`
**Repository:** `tamzrod/dnp3`  ·  **VEC-01 NOT EXECUTED / R4 OPEN / THIRD-PARTY INTEROPERABILITY NOT PROVEN.**

This directory is the durable, reviewable **Gate B** evidence package required by the
owner-approved addendum (2026-11-10) that designates *this Git repository* as the Gate B
evidence store. It contains the smallest reviewer-accessible packet: pinned upstream
provenance, a reproducible offline rebuild recipe, the exact lab-only C++ patch that
produces the v0 simulator, the static network/bind/link topology, the simulator point and
control boundary, the bounded deadline/watchdog plan, and a reviewer checklist with blank
identity/date/decision fields.

The packet is **documentation and rebuildable inputs only**. It contains **no** private
keys, credentials, identifying private-network data, third-party archives, compiled
binaries, packet captures or `.vec` traces.

## Packet contents

| File | Purpose |
|------|---------|
| `README.md` | This index, scope and status. |
| `provenance.md` | Pinned OpenDNP3 3.1.2 source/dependency URLs, expected SHA-256/SHA-1, tag/peel refs. |
| `rebuild-and-lab.md` | Exact offline rebuild recipe, isolation construction, and reproduction steps. |
| `lab-config.md` | Exact lab-only C++ demo patch (address + v0 variations + fixed points + memory-only BO[0]), and the prospective Go master configuration. |
| `reviewer-checklist.md` | Reviewer-readable checklist with **blank** identity/date/decision fields. |
| `SHA256SUMS.txt` | SHA-256 manifest of the committed packet files (repository persistence proof). |

## Authoritative source documents (relative links, not copied)

- Execution/evidence master record: [`../VEC-01-LAB-EXEC-001.md`](../VEC-01-LAB-EXEC-001.md)
- Concise campaign index: [`../../active_work/vec-01-opendnp3-lab-execution-evidence.md`](../../active_work/vec-01-opendnp3-lab-execution-evidence.md)
- Deployment/readiness provenance: [`../../active_work/vec-01-opendnp3-lab-deployment-readiness.md`](../../active_work/vec-01-opendnp3-lab-deployment-readiness.md)
- Supported v0 profile boundary: [`../../active_work/supported-profile.md`](../../active_work/supported-profile.md)
- External `.vec` format (future Gate E only): [`../../active_work/testdata/external/FORMAT.md`](../../active_work/testdata/external/FORMAT.md)
- Promoted CURRENT task (read-only here): [`../../workflow/active_work/VEC-01-LAB-EXEC-001.md`](../../workflow/active_work/VEC-01-LAB-EXEC-001.md)

## Scope of this packet (read carefully)

**In scope — freshly verified this pass:**

- Pinned upstream provenance URLs and expected digests (bytes retrieved and hash-checked
  against the recorded values; see `provenance.md`).
- The exact lab-only C++ demo patch text (address bind, v0 static variations, fixed
  memory points, memory-only BO[0] Direct-Operate handler) — verbatim, non-secret.
- The prospective Go master configuration and the static bind/link address mapping.
- The bounded deadline/watchdog/stop-rule plan.
- A reviewer checklist awaiting a real named reviewer.

**Out of scope / NOT claimed:**

- **No** fresh network-isolation PASS. The isolation results recorded in the master
  evidence (§11, §15.3) are **historically documented**, not re-verified in this packet
  build. A reviewer must re-run the isolation checks on a live lab at sign-off.
- **No** compiled peer binary is committed. The lab `outstation-demo` / `libopendnp3.a`
  SHA-256 values (§13.4, §15.5) are **rebuild expectations**, not retained bytes; a
  path-dependent rebuild is expected to differ only by embedded source-path strings.
- **No** listener activation, TCP/DNP3 traffic, CROB execution, capture or `.vec`.
- **No** Gate B PASS, B4 independent PASS, interoperability or certification claim.

## Reviewer handoff

A **named independent reviewer distinct from the executor** must, at sign-off:

1. Inspect a **live, freshly built** isolated topology and re-run the objective isolation
   checks (`rebuild-and-lab.md` §4).
2. Confirm non-routed / no-OT-reachability and **absence of physical output**.
3. Hash-check the pinned OpenDNP3 3.1.2 peer bytes against `provenance.md`.
4. Confirm the v0 variations G1V1/G30V1/G20V1 and the **sole memory-backed BO[0]** CROB target.
5. Confirm the single-shot control policy and capture placement.
6. Record a **dated** approval of exact topology, absence of physical output, and the sole
   simulator CROB target, with the reviewed config/binary hashes and exact scope.

Record the decision in `reviewer-checklist.md`. This packet **cannot grant or simulate**
that approval. Historical isolation results are not fresh approval.
