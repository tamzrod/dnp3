# SCV Mission — DNP3 Analog Output Command (Group 41)

**SCV = Space Construction Vehicle** (StarCraft mission codename). **Owner priority:** make generic DNP3 Analog Output command support ready for future power-plant-controller setpoint integration. **Active branch:** `work`. **Current CWAL package:** `AO-003` (G41V2 int16 and G41V3 float32 implemented; G41V4 float64 deferred to a separate later invocation after independent CHECK; the owner approved the IEEE-754 float/NaN/Inf policy on 2026-11-11). **No field/production commands authorized.**

## Verified baseline (2026-11-11)
- PR #26 from `work` to `main` covers scoped v0 BI G1V1, AI G30V1, Counter G20V1 and CROB G12V1; R4 independent-lab evidence is owner-accepted. Do not revert or alter this evidence.
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
| **AO-003 CURRENT** (G41V2 and G41V3 done; V4 deferred) | Complete G41 integer and floating variations required by approved matrix, safety/lifecycle/regression tests | Signed/range/rounding/NaN/Inf as applicable, malformed/truncated/unsupported qualifiers, error statuses, timeout and no blind retry |
| **AO-004 future human promotion** | Independent isolated OpenDNP3 peer tests for supported G41 subset + original PCAP and packet-derived .vec | Controlled in-memory setpoint only; positive/negative evidence, independent decode, bounded CRC/readiness and reproducible verification; scoped evidence statement |
| **AO-005 future human promotion** | Update supported profile, README/API docs and release acceptance for the *actually verified* G41 subset | No generalized IEEE 1815 conformance, production readiness or presumed NGCP profile |

**SBO and Direct Operate No Response** remain out of scope until a separate human-approved profile. Do not silently enable these modes. Which G41 variations to implement first must be grounded in an authoritative external specification or actual peer-vectors rather than guesses. Support may be implemented generically ahead of NGCP mapping but must be advertised only after tests.

## Sequence and acceptance discipline
The only invocation to JR is **`operation cwal`**. Execute only the currently promoted task. `AO-002` is DONE (G41V1, bounded internal verification); the single CURRENT package is `AO-003`, where G41V2 (int16) and G41V3 (float32) are implemented and verified under the owner-approved IEEE-754 float policy, while G41V4 (float64) is deferred to a separate later invocation after independent CHECK on V3. After its verified checkpoint STOP for independent CHECK; later rows are roadmap candidates, not execution authority. Use non-force commits/pushes to `work` and GitHub readback. No live SCADA, NGCP/production network, physical I/O, merge, or release. The open PR #26 uses the permanent `work` head; scope of future commits must be reviewed before any PR merge.
