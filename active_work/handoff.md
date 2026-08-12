# DNP3 Implementation Handoff

**Last updated:** 2026-08-12  
**Roadmap:** `active_work/DNP3_MASTER_ROADMAP.md`  
**Profile:** `active_work/supported-profile.md`

## Status

- Planning complete.
- No micro-tasks implemented yet.
- Source tree is at the state after the interoperability remediation work.

## Completed Tasks

(none)

## Next READY Tasks

Any of the following may be taken immediately (no prerequisites):

- **DNP3-001** — Residual endian audit  *(recommended first)*
- **DNP3-002** — Formal object-header model (encode)
- **DNP3-006** — Link handshake ACK validation
- **DNP3-008** — Application sequence continuity
- **DNP3-011** — IIN bit semantics verification & correction
- **DNP3-022** — Context cancellation on Connect
- **DNP3-030** — Complete supported-profile rejection matrix
- **DNP3-031** — Transport disconnect detection
- **DNP3-036** — Deterministic outstation simulator (MVP profile)
- **DNP3-041** — Timeout configuration validation
- **DNP3-043** — Error type taxonomy
- **DNP3-044** — Logging hooks
- **DNP3-049** — Master address configuration validation
- **DNP3-059** — Transport fragment size boundary tests
- **DNP3-065** — Double-check DLL EncodedSize usage
- **DNP3-072** — Master handoff.md template (this file exists; task can mark complete once format is confirmed useful)
- **DNP3-080, 084, 085, 087, 088, 098** (Outstation-side READY tasks)

## Recommended First Task

**DNP3-001 — Residual endian audit**

After completing a task:

1. Run the relevant unit / golden / loopback tests.
2. Update this handoff (move task to Completed, note any new READY tasks, record test commands used).
3. Commit with the exact commit message from the roadmap.
4. Every 3 completed tasks: verify → update handoff → commit → push to main.

## Test Commands (baseline)

```bash
go test ./...
go test -race ./internal/master/... ./pkg/dnp3/...
go test ./test/integration/...
```

(Exact packages will be refined as tasks progress.)

## Notes for Agents

- Do **not** implement outside the defined micro-tasks.
- Do **not** invent requirements or add unsupported objects/features.
- Prefer deterministic fixtures and golden vectors.
- Keep changes commit-sized and independently testable.
- Master has priority; Outstation tasks exist only to support Master correctness and testing.

## MVP Gate

MVP is declared complete only when **DNP3-056** passes.

```
TOTAL TASKS: 100
MASTER TASKS: 72
OUTSTATION TASKS: 28
MVP COMPLETE AT: DNP3-056
NEXT TASK: DNP3-001 — Residual endian audit
```
