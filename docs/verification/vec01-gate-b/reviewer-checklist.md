# Gate B independent preflight checklist (blank — for the named reviewer)

This checklist is **blank on purpose**. A **named independent reviewer distinct from the
executor** completes it after inspecting the **actual live environment and exact evidence**.
This packet cannot grant, simulate or self-award sign-off. Historical isolation results are
**not** fresh approval.

## Reviewer identity and scope

| Field | Value (to be completed by the reviewer) |
|-------|------------------------------------------|
| Reviewer name | |
| Reviewer role / qualification | |
| Relationship to executor (must be distinct) | |
| Date / time of inspection (with timezone) | |
| Review channel / authenticated method | |
| Executor identity (for independence check) | |
| Exact reviewed scope | |
| Branch / commit SHA reviewed | |

## Inspected artifacts and hashes (reviewer records actual values)

| Item | Reviewer-observed value | Matches packet? |
|------|-------------------------|-----------------|
| OpenDNP3 source tarball SHA-256 | | |
| asio zip SHA-256 | | |
| exe4cpp zip SHA-256 | | |
| ser4cpp zip SHA-256 | | |
| Rebuilt lab `outstation-demo` SHA-256 | | |
| Rebuilt lab `libopendnp3.a` SHA-256 | | |
| Lab demo source diff SHA-256 | | |

## Live isolation checks (re-run on the live lab; non-TCP only)

| # | Check | Reviewer result (PASS/FAIL/NOT_TESTED) | Evidence |
|---|-------|----------------------------------------|----------|
| 1 | Only `10.77.0.0/24` route; no external IPv4 route | | |
| 2 | No IPv6 global route | | |
| 3 | Forwarding disabled (`0`) | | |
| 4 | Deny-all lab firewall; only lo/veth allowed | | |
| 5 | No NAT / bridge / published host port | | |
| 6 | No lab listening sockets before activation | | |
| 7 | No physical I/O (ttyS*/ttyUSB*/mem/port/GPIO absent) | | |
| 8 | Lab-local reachability only | | |
| 9 | Host/pod boundary unchanged | | |

## Simulator and control-boundary confirmation

| Item | Reviewer result | Notes |
|------|-----------------|-------|
| Bind is `10.77.0.2:20000`, **not** `0.0.0.0` | | |
| Link addresses master `3` / outstation `4` | | |
| Variations G1V1 / G30V1 / G20V1 at index 0 | | |
| Fixed points BI[0]=true, AI[0]=2300, CT[0]=42 | | |
| **Sole** memory-backed BO[0] Direct-Operate target | | |
| No physical/field output path | | |
| Single-shot control, no blind retry | | |
| Capture placement defined (future Gate E) | | |
| Bounded deadlines / watchdog / stop rules defined | | |

## Decision

| Field | Value (reviewer completes) |
|-------|----------------------------|
| Decision (`APPROVE` / `REJECT` / `INCOMPLETE`) | |
| Exact scope of approval | |
| Conditions / limitations | |
| Dated signature | |

**Reminder:** Gate B approval is a prerequisite for any listener activation, TCP/DNP3
packet, or CROB. This checklist being present does **not** constitute approval.
