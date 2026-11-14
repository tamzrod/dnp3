# SCV Mission — DNP3 Analog Output Command (Group 41)

**SCV = Space Construction Vehicle** (StarCraft mission codename). **Owner priority:** make generic DNP3 Analog Output command support ready for future power-plant-controller setpoint integration. **Active branch:** `work`. **Current CWAL package:** `AO-004` (independent isolated OpenDNP3 peer validation for the G41 subset — **NOT RUN**; AO-003 is DONE with G41V2/V3/V4 source/evidence checked and G41V1 done in AO-002; the owner approved the IEEE-754 float/NaN/Inf policy on 2026-11-11). **No field/production commands authorized.**

## Verified baseline (2026-11-11)
- PR #26 from `work` to `main` (**merged 2026-11-11**; merge commit `f55278cb1f91fb7e66e82a89458f6de9d777302a`) covers scoped v0 BI G1V1, AI G30V1, Counter G20V1 and CROB G12V1; R4 independent-lab evidence is owner-accepted. Do not revert or alter this evidence.
- `active_work/supported-profile.md` still rejects Group 41 analog output **commands** and `types.NewAnalogControl`. Outstation's public `CommandHandler.HandleAnalogCommand` exists but cannot be treated as tested/supported G41 functionality.
- **G1V2 remains unsupported and has a known mis-decode.** Not in this mission.
- NGCP's precise command variation, index mapping, scale/units, Direct Operate vs SBO policy, select/operate timing, and response contract are **not provided**. Do not invent utility-specific requirements or hardwire MW/Mvar/voltage/PF points in the generic library.

## Mission outcome
Provide tested, bounded, standards-grounded DNP3 Group 41 Analog Output **commands** for the standalone Go Master and Outstation, with explicit per-variation and operate-mode support; expose a generic callback for host-controlled setpoints. Separate library capability from any PPC mapping or permission to command a live grid device.

## Ordered CWAL work packages (not auto-promoted)
| Package | Deliverable | Acceptance |
|---|---|---|
| **AO-001 DONE** | Audit G41 wire path, API, unsupported gates, standard/peer reference; write bounded design & exact tests | Only verified supported variation/qualifier/operation matrix; precise source and golden references; approved implementation plan; no Go behavior change |
| **AO-002 DONE** | Implement G41 direct-operate encode/decode and strict command-status handling for verified variation(s), Master + Outstation bridge | Unit and golden-wire tests, invalid request rejection, no false SUCCESS, no side effects on rejected command |
| **AO-003 DONE** (G41V2/V3/V4 source/evidence checked; V1 done in AO-002) | Complete G41 integer and floating variations required by approved matrix, safety/lifecycle/regression tests | Signed/range/rounding/NaN/Inf as applicable, malformed/truncated/unsupported qualifiers, error statuses, timeout and no blind retry |
| **AO-004 CURRENT** (independent peer validation **NOT RUN**) | Independent isolated OpenDNP3 peer tests for supported G41 subset + original PCAP and packet-derived .vec | Controlled in-memory setpoint only; positive/negative evidence, independent decode, bounded CRC/readiness and reproducible verification; scoped evidence statement |
| **AO-005 future human promotion** | Update supported profile, README/API docs and release acceptance for the *actually verified* G41 subset | No generalized IEEE 1815 conformance, production readiness or presumed NGCP profile |

**SBO and Direct Operate No Response** remain out of scope until a separate human-approved profile. Do not silently enable these modes. Which G41 variations to implement first must be grounded in an authoritative external specification or actual peer-vectors rather than guesses. Support may be implemented generically ahead of NGCP mapping but must be advertised only after tests.

## Sequence and acceptance discipline
The only invocation to JR is **`operation cwal`**. Execute only the currently promoted task. `AO-001` and `AO-002` are DONE. `AO-003` is DONE (G41V1 in AO-002; G41V2 int16, G41V3 float32 and G41V4 float64 implemented under the owner-approved IEEE-754 float policy and source/evidence checked). The single CURRENT package is now `AO-004` (independent isolated OpenDNP3 peer validation — **NOT RUN**; its bounded lab-isolation/readiness-first mandate requires a fresh later `operation cwal` after the transition independent CHECK). After the verified checkpoint STOP for independent CHECK; later rows are roadmap candidates, not execution authority. Use non-force commits/pushes to `work` and GitHub readback. No live SCADA, NGCP/production network, physical I/O, merge, or release. PR #26 (merged 2026-11-11) used the permanent `work` head; PR #27 (merged into `main` 2026-11-12, merge commit `2ff2ffca05c66ba73d8c08a561e1227d301348a1`) merged the G41V1–V4 implementation while AO-004 independent-peer G41 interoperability remains **NOT RUN** and the supported profile is unchanged; scope of future commits must be reviewed before any PR merge.
