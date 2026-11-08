# VEC-01 OpenDNP3 lab execution — evidence index (TASK_BLOCKED)

Concise index for the isolated master ↔ OpenDNP3 3.1.2 interoperability
campaign (CURRENT `VEC-01-LAB-EXEC-001`). The campaign is **BLOCKED at Gate B**;
nothing was executed on the wire. Full detail and hashes:
`docs/verification/VEC-01-LAB-EXEC-001.md`.

**VEC-01 NOT EXECUTED / R4 OPEN / THIRD-PARTY INTEROPERABILITY NOT PROVEN.**

Durable lab artifact store (outside Git, outside `/tmp`):
`/workspace/vec01-lab-artifacts`.

## Campaign matrix

| Gate | Step | Status | Evidence |
|------|------|--------|----------|
| Startup | markers, `cwal-check.sh`, remote ancestry | PASS | `origin/main=02a58f7` ancestor of `origin/work=f44d28b`; `CWAL CHECK OK` |
| A | retrieve/reproduce pinned peer + deps, durable retention | PASS | tarball `183cc292…`; tag peeled `26b4c01e…`; all dep SHA-256/SHA-1 match; hermetic rebuild byte-identical `8ae8c841…` / `e2926474…` |
| B | named independent reviewer isolation/output sign-off | **BLOCKED** | no named independent reviewer available; no autonomous self-approval permitted |
| C | positive matrix (TCP connect, link reset/status, Class-0 G60V1, G1/G30/G20 decode, G12V1 direct-operate, close) | NOT REACHED | Gate B is a hard precondition |
| D | negatives (bad CRC, wrong address, drop, command reject, incomplete capture) | NOT REACHED | Gate B is a hard precondition |
| E | raw PCAP, faithful `.vec`, SHA-256 manifest, reviewer verdict | NOT REACHED | no capture started; no `.vec` fabricated |
| Go gates | `gofmt`, `git diff --check`, `go vet`, `go build`, `go test`, `-race`, `verify-mvp.sh`, `verify-external-mvp.sh` | `git diff --check` PASS; others NOT_RUN | Go toolchain absent; real-TCP gates also prohibited pre-approval |

## Peer provenance (re-verified from retrieved bytes)

| Item | Value |
|------|-------|
| Upstream | `https://github.com/dnp3/opendnp3` (archived; Apache-2.0) |
| Tag `3.1.2` peeled commit | `26b4c01e4839bbbda8866655e086471c4917ee53` |
| Source tarball SHA-256 | `183cc29222c3cb58099a9753c43defed5cb8677fda985c3f88e38b5c19a36ff2` |
| `outstation-demo` SHA-256 | `8ae8c84125c4920352b6daef4145e2a91de9e02cd6e2c1b94bf6995ebeb093bb` |
| `libopendnp3.a` SHA-256 | `e2926474288fb8d2c6c7962192bad2ff0f724f63c07d2209fe5b30c26fbe0677` |
| Dependencies | asio `asio-1-16-0`, exe4cpp `fb878a4…`, ser4cpp `3c44973…`, catch `v2.11.3` (unused) — all SHA-256/SHA-1 match |

## Honest scoped / non-claims

- The pinned peer was **reproduced**, not executed; this is build evidence only.
- No listener, no TCP session, no DNP3 packet, no capture, no control occurred.
- No interoperability, conformance, certification, or production-readiness claim.
- The blocker is the mandatory **independent reviewer** authority requirement,
  not a technical protocol failure.
- Closing R4, changing `supported-profile.md`, production use, and feature
  expansion each require separate owner approval — none was sought or given.

## Gate B evidence-preparation substep (human-authorized addendum, 2026-11-08)

Executed under the owner-authorized **Gate B pushback addendum (EVIDENCE
PREPARATION ONLY)**; task remains `TASK_BLOCKED`. **Passive/static only: no
listener, no TCP/DNP3 packet, no capture, no control, no probing.** Substep
execution SHA: `9d6833c914608c517748f9c26a8a5fc06701170f`.

| Item | Result | Reference |
|------|--------|-----------|
| Retained peer re-hash | PASS (matches pinned) | `gate-b-reviewer-packet/inspection/peer-artifact-manifest.txt` |
| Network/isolation (passive) | NOT ISOLATED (routed k8s pod, internet egress, `ip_forward=1`) | `…/inspection/network-inventory.txt` |
| Firewall / tooling (list-only) | no policy; `ip` absent → cannot build `10.77.0.0/24` | `…/inspection/firewall-inventory.txt` |
| Peer configuration | binds `0.0.0.0:20000`, addrs 10/1; default BI = **G1V2** (not G1V1); AI[0]=G30V5 | `…/inspection/peer-configuration.txt` |
| Simulator / output mapping | points memory-backed; **no BO[0] target**; CROB = status echo only; no physical I/O | `…/inspection/simulator-and-output-mapping.txt` |
| UNAPPROVED reviewer packet | assembled with blank reviewer checklist | `/workspace/vec01-lab-artifacts/gate-b-reviewer-packet/` |

Outstanding UNVERIFIED: reviewer identity/decision; environment isolation and
no-OT-reachability; peer v0 variation compliance; verifiable BO[0] CROB target;
Class-0 qualifier / multi-header / sequence behaviour; Go toolchain.

Resumption prerequisites (all required): authenticated independent reviewer
sign-off on the actual environment and hashes; demonstrably isolated non-routable
topology; peer configured to v0 variations with a verifiable simulator-only
BO[0]; and owner authorization to proceed. Until then: `TASK_BLOCKED`.

**VEC-01 NOT EXECUTED / R4 OPEN / THIRD-PARTY INTEROPERABILITY NOT PROVEN.**

---

## Owner correction — objective Gate B replaces reviewer sign-off (2026-11-08)

This note supersedes the **prospective reviewer/signature requirement** in the older Gate B notes above. The older statements accurately describe the prior TASK_BLOCKED reason at that point in history; they are not current authority. Under the owner-amended CURRENT, technical **PASS / FAIL / NOT_TESTED** evidence, not additional-person approval, determines readiness.

| Objective Gate B requirement | Current evidence | Disposition |
|---|---|---|
| No internet or production/OT egress; non-routed isolated lab | Existing Kubernetes pod routed, internet egress, forwarding enabled; no safe isolated segment demonstrated | **FAIL** |
| Outstation listener only on lab interface and correct 3/4 link addresses | Stock demo binds `0.0.0.0:20000`, uses 10/1; proposed safe config not yet proven | **NOT_TESTED / BLOCKED** |
| Exact v0 G1V1/G30V1/G20V1 static variations | Stock demo observed G1V2/G30V5, exact v0 response not demonstrated | **FAIL / NOT_TESTED** |
| Memory-only BO[0] with observable CROB result | No verified target; stock demo status echo only | **FAIL** |
| Finite deadlines and no blind control retry | Required but not proven on an executable isolated setup | **NOT_TESTED** |
| Pinned peer artifacts and hash integrity | Previously recorded reproducible matching binaries | **Recorded PASS; rehash at use** |

Gate C/D/E and all previously unrun Go checks remain **NOT_TESTED**, not PASS. JR must not operate the peer until objective Gate B checks PASS in a demonstrably isolated environment. Preserve original raw PCAP and trace-derived `.vec` only when later authorized execution actually occurs. No mandatory external reviewer/signature. No master-code fixes, new scope, self-certification of full IEEE 1815, production access or R4 closure.

**VEC-01 NOT EXECUTED / R4 OPEN / THIRD-PARTY INTEROPERABILITY NOT PROVEN.**

---

## Gate B substep 1 — isolation construction and proof (2026-11-08)

Owner-authorized "isolation construction and proof only" substep. **Result:
`NETWORK_ISOLATION_PASS`** (isolation only). Gate B overall remains
`TASK_BLOCKED`. Execution SHA `675183a64179c084ece77728bb57f779c679af4e`.

A dedicated isolated lab was constructed with two private network namespaces
(`vec01-master`, `vec01-outstation`) joined by an internal veth pair only
(subnet `10.77.0.0/24`, addrs `10.77.0.1`/`10.77.0.2`), forwarding disabled and
a deny-all policy applied **inside the lab only**. The host/pod routes, firewall
and cluster network policy were **not** modified (host netns inode unchanged).

| Objective boundary check | Result |
|---|---|
| No external IPv4/IPv6 route (passive lookups) | PASS (unreachable) |
| No route to cluster API / pod gateway | PASS (unreachable) |
| No DNS resolution | PASS (fails) |
| No NAT / bridge / published listener | PASS (none) |
| Lab-local reachability only | PASS (veth only) |
| Forwarding disabled in lab | PASS (`0`) |
| No IPv6 globals | PASS (0) |
| Physical I/O absent (serial/GPIO/mem) | PASS (absent) |
| Host/pod boundary unchanged | PASS (inode unchanged) |

Evidence (outside Git, SHA-256 manifest included):
`/workspace/vec01-lab-artifacts/gate-b-isolation-proof/`.

Still BLOCKED and separately authorized-only: **B2** listener safety (proposed
`10.77.0.2:20000`, link addrs 3/4), **B3** peer v0 capability (G1V1/G30V1/G20V1)
with a verifiable memory-backed **BO[0]** CROB target, **B4** bounds/retry
policy, then Gates C/D/E and the mandatory Go gates. No listener, DNP3/TCP
packet, control, CROB, PCAP or networked Go test ran in this substep.

**VEC-01 NOT EXECUTED / R4 OPEN / THIRD-PARTY INTEROPERABILITY NOT PROVEN.**
