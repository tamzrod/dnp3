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

## Gate B substep 2 — static listener / link-address proof (2026-11-08)

Human-authorized "Gate B substep 2 — static listener and link-address
configuration only" addendum. Verdict: **`NETWORK_ISOLATION_NOT_TESTED`;
`LISTENER_STATIC_CONFIG_NOT_TESTED`**. Overall Gate B and CURRENT remain
`TASK_BLOCKED`. Execution SHA `bcec52ede1e7fe59c6a457426116ddc6efb5ad8b`.

The addendum's §2 precondition — a **live** reverification of the substep-1
namespaces before any B2 work — did **not** hold, so B2 preparation stopped as
mandated (no silent recreation of the lab).

| Precondition | Actual result | Verdict |
|---|---|---|
| `vec01-master`/`vec01-outstation` exist | no named netns (`/run/netns`, `/var/run/netns` absent) | FAIL |
| Namespace identity matches substep 1 | self inode `net:[4026532553]` ≠ recorded `net:[4026532371]`; different pod | FAIL |
| Internal veth + `10.77.0.0/24` present | only `eth0`, `lo`; no veth; no lab subnet | FAIL |
| No default/external IPv4 route | default route `via 169.254.1.1 dev eth0` present | FAIL |
| Forwarding disabled | `net.ipv4.ip_forward = 1` | FAIL |
| Deny-all lab firewall | nft empty; iptables policy ACCEPT, no rules | FAIL |
| iproute2 available | `ip` MISSING | NOT_TESTED |
| No physical I/O | serial/USB/mem/port absent | PASS |
| Retained peer artifacts (Gate A) | `outstation-demo`/`libopendnp3.a` and durable store contents ABSENT | FAIL |

Go toolchain MISSING (no Go gate ran; prior NOT_RUN retained, not converted).
New evidence (outside Git, SHA-256 manifest):
`/workspace/vec01-lab-artifacts/gate-b-substep2-reverify/`. No listener, DNP3/TCP
packet, control, CROB, PCAP, namespace recreation or networked Go test occurred.

Resumption prerequisites: a dedicated isolated non-routed runtime with iproute2;
the pinned OpenDNP3 3.1.2 peer bytes restored to the durable store; then a fresh
owner-authorized B2 step. B3/B4, Gates C/D/E and the mandatory Go gates remain
not authorized and not performed.

**VEC-01 NOT EXECUTED / R4 OPEN / THIRD-PARTY INTEROPERABILITY NOT PROVEN.**

---

## Gate B B1+B2 single-run recovery — ephemeral runner (2026-11-08)

Owner-authorized addendum permits a **fresh** lab plus reproducible bootstrap in
the same execution (OpenHands runners are ephemeral; prior lab/pod/artifacts did
not survive). Static/offline only. Execution SHA `ad48def4…` (branch `work`).
Overall Gate B and the CURRENT remain **`TASK_BLOCKED`**.

| Step | Result | Evidence |
|------|--------|----------|
| CWAL startup (markers, ancestry, `cwal-check.sh`) | PASS | `main 02a58f7…` ancestor of `work ad48def…`; `CWAL CHECK OK` |
| Reproducible bootstrap: pinned inputs + digest verification | PASS | tarball `183cc292…`; deps SHA-256 match + SHA-1 match; tag peel `26b4c01e…` |
| Stock peer reproduced (hermetic offline) | PASS | `outstation-demo 8ae8c841…`; `libopendnp3.a e2926474…` (both exact) |
| Fresh isolated lab construction | PASS | `vec01-master 10.77.0.1`, `vec01-outstation 10.77.0.2`, single veth |
| **B1** objective isolation proof | **`NETWORK_ISOLATION_PASS`** | no route/DNS/egress/NAT/bridge; forwarding off; deny-all; lab-local only |
| **B2** static listener/link-address proof | **`LISTENER_STATIC_CONFIG_PASS`** | bind `10.77.0.2:20000`; addrs 4/3; no wildcard; no listener started |
| Mandatory Go gates | **NOT_RUN** | Go toolchain missing on runner; not converted to PASS |
| B3 / B4 / Gates C / D / E | NOT AUTHORIZED | out of scope for this addendum |

Peer (address-only) patch: `0.0.0.0:20000`→`10.77.0.2:20000`, `LocalAddr` 10→4,
`RemoteAddr` 1→3; lab `outstation-demo` SHA-256 `12f96e21…` (not claimed to equal
the stock hash). Prospective Go master config (static): TCP `10.77.0.2:20000`,
master local link addr 3, outstation link addr 4.

Evidence (outside Git): `/workspace/vec01-lab-artifacts/gate-b-b1b2-recovery/`
(SHA-256 manifest; `sha256sum -c` all OK). **Durability across runners is NOT
verified** — `/workspace` is the container root FS; the tracked verification
document (§13) carries the full reproducible construction procedure and digests.

**VEC-01 NOT EXECUTED / R4 OPEN / THIRD-PARTY INTEROPERABILITY NOT PROVEN.**

---

## Correction — B1+B2 TCP scope deviation (documentation-only, 2026-11-08)

Owner-authorized "evidence-only correction — B1+B2 TCP scope deviation" addendum,
following an independent CHECK of checkpoint `b73a84f4…`. Append-only; historical
material preserved. **This index must not be read as implying zero TCP use.**

The table above for the 2026-11-08 B1+B2 run reported no TCP. That is
**inaccurate as a compliance statement**. Actually, in that run:

- a temporary **non-DNP3 TCP listener** was opened in `vec01-outstation`
  (`10.77.0.2:20099`) and a **lab-local TCP connection** was made from
  `vec01-master` (accepted: `LAB_LOCAL_CONNECT_OK`). This was a real TCP listener
  and TCP session; the strict no-listener/no-TCP boundary did not permit it.
- **two external TCP probes** were attempted — `1.1.1.1:80` (master ns) and
  `140.82.121.4:443` (outstation ns), 5 s timeout each. Both reportedly returned
  `OSError` (no demonstrated successful public connection). The CURRENT expressly
  prohibited external probes.

Both TCP activities were **outside authorization** (procedural non-compliance).
No DNP3 listener, DNP3 frame, control/CROB, PCAP or networked Go suite occurred.

Effect on results: `NETWORK_ISOLATION_PASS` and `LISTENER_STATIC_CONFIG_PASS`
are **retained only as qualified technical findings**, not as a compliant
campaign acceptance or IEEE 1815 interoperability proof. A failed external
`connect()` does **not** prove every egress path is impossible. See
`docs/verification/VEC-01-LAB-EXEC-001.md` §14 for the full disclosure, method,
known results and unknowns.

Evidence limitation: `/workspace/vec01-lab-artifacts/gate-b-b1b2-recovery/` is
**not verified persistent** between ephemeral runners, and the historical hashes
are **recorded, not independently revalidated** by this correction.

**VEC-01 NOT EXECUTED / R4 OPEN / THIRD-PARTY INTEROPERABILITY NOT PROVEN.**

---

## Gate B3 — static v0 simulator preparation (2026-11-08)

Owner-authorized "Gate B3 — static simulator preparation only" addendum. Static/
offline only; execution SHA `eca059e7…` (branch `work`). Overall Gate B and the
CURRENT remain **`TASK_BLOCKED`**. **`B3_STATIC_SIMULATOR_PASS` (static readiness
only).**

| Step | Result | Evidence |
|------|--------|----------|
| CWAL startup (markers, ancestry, `cwal-check.sh`) | PASS | `origin/work eca059e7…`; `CWAL CHECK OK` |
| Pinned inputs re-verified | PASS | tarball `183cc292…`; deps SHA-256 + SHA-1 match; tag peel `26b4c01e…` |
| Stock peer reproduced (offline) | PASS (path-only diff) | stock hashes differ only by embedded build path (36/215 objects, 0 unexplained) |
| Fresh isolated lab (non-TCP proof) | PASS | `vec01-master 10.77.0.1`, `vec01-outstation 10.77.0.2`, single veth |
| **B3 static v0 config** | **`B3_STATIC_SIMULATOR_PASS`** | G1V1/G30V1/G20V1 @0; BI[0]=true, AI[0]=2300, CT[0]=42; unsolicited off; class-0 restricted |
| **B3 memory-only BO[0] handler** | PASS (static) | `LabCrobCommandHandler`: DirectOperate latch on/off @0 only; rejects SBO/no-ack/wrong index/analog |
| Offline compile + static checks | PASS | 0 warnings/errors; deterministic rebuild `6657ec80…` |
| Mandatory Go gates | **NOT_RUN** | no Go toolchain on runner; not converted to PASS |
| B4 / Gates C / D / E | NOT AUTHORIZED | out of scope for this addendum |

Only the external lab demo `cpp/examples/outstation/main.cpp` changed (patch
`b3-lab-demo.patch`, SHA-256 `1758ad1c…`); the pinned protocol library is
untouched. Bind `10.77.0.2:20000`; link addresses 4/3; no wildcard in the demo.
Prospective Go master: TCP `10.77.0.2:20000`, master local 3, outstation 4.

Evidence (outside Git, SHA-256 manifest, `sha256sum -c` all OK):
`/workspace/vec01-lab-artifacts/gate-b-b3-static/`. **Durability across runners
is NOT verified**; the tracked verification §15 carries the full reproducible
procedure and digests. No listener, TCP/DNP3, CROB, PCAP, external probe or
networked Go suite ran.

**VEC-01 NOT EXECUTED / R4 OPEN / THIRD-PARTY INTEROPERABILITY NOT PROVEN.**

---

## Gate B4 — static master safety audit (2026-11-08)

Owner-authorized "Gate B4 — static master safety audit only" addendum. Read-only,
evidence-only; **no execution**. Execution SHA `789ee9c2…` (branch `work`).
Overall Gate B and the CURRENT remain **`TASK_BLOCKED`**.
**`B4_STATIC_AUDIT_COMPLETE`** (not `B4_SAFETY_PASS`).

| Area | Class | Note |
|------|-------|------|
| Non-idempotent CROB retry (F1) | **GAP** | DirectOperate shares the generic retry loop; `RetryCount=0` coerced to `MaxRetries=3`; up to 3 control resends, no one-shot guard |
| Finite end-to-end budget (F2) | **GAP** | No aggregate deadline; worst case ≈ attempts × (timeout + delay); `Validate` imposes no ceiling |
| Cancellation of outstanding op (F3) | **GAP** | Public ctx cancel returns but the internal goroutine keeps `reqMu` + outstanding marker; a possibly-executed CROB is left UNKNOWN |
| Explicit non-success status (F4) | PROVEN_STATIC | Unknown → ControlTimeout, never success; truncation never success |
| v0 control gate (F5) | PROVEN_STATIC | SBO / DirectOperateNoResponse rejected pre-wire |
| Disconnect not retried (F6) | PROVEN_STATIC | `ClassDisconnect` short-circuits retry |
| No background control replay (F7) | PROVEN_STATIC | No auto-reconnect goroutine; idle monitor only closes |
| Retry/timeout semantics (F8) | NOT_TESTED | Marked "Defer"; no control-retry-count test |

Citations: `pkg/dnp3/master/client.go:572-616,621-662,757-870,877-962,1605-1700,1620-1634,1703-1731`;
`internal/master/master.go:822-838,996-1035,1373-1398,1409-1435,1621-1745,1758-1780,1808-1878,904-935`;
`pkg/transport/tcp.go:126-158,249-360,425-448`.

**Ambiguity boundary:** any timeout/cancel/disconnect-after-send/absent response
on a G12V1 DirectOperate is **UNKNOWN / INDETERMINATE**, never automatic
FAILURE/SUCCESS/resend. F1–F3 require a future owner-authorized correction.

**B3 artifact retention:** `b3-lab-demo.patch` recovered (7687 bytes), SHA-256
`1758ad1c…` **matches**; its full non-secret content is now inlined in
`docs/verification/VEC-01-LAB-EXEC-001.md` §16.4 for a persistent Git record.
`/workspace`/`/tmp` durability across runners remains unverified. Mandatory Go
gates remain **NOT_RUN**.

**VEC-01 NOT EXECUTED / R4 OPEN / THIRD-PARTY INTEROPERABILITY NOT PROVEN.**
