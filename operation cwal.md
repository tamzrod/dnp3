# OPERATION CWAL — DNP3

Repository: `tamzrod/dnp3`. Inspired by the current `tamzrod/Atlas-PPC` Operation CWAL procedure; adapted to the DNP3 library and subordinate to existing KDE governance.

## Mission and authority
Execute only a **human-promoted, bounded CURRENT microtask**, with evidence-based verification and confirmed remote Git checkpoints. This document is a workflow overlay, not a change to KDE governance authority. Refer to `.kde/governance/GOVERNANCE-HIERARCHY.md`, approved architecture, and `active_work/supported-profile.md`. Governance-changing proposals require their normal human approval.

## Startup / scope gate
1. Read `AGENTS.md`, this file, `BLACK_SHEEP_WALL.md` (routing rules), and the **top current-state blocks only** of `workflow/active_work/README.md` and `handoff.md`.
2. Compare exact `Repository:`, `Package:`, `Branch:`, `CURRENT:`, `QUEUED:`, `Status:` markers in both files. Validate the designated branch and remote ref, clean/safely preservable local state, and the actual CURRENT task file.
3. Only a human may promote, reorder, replace or expand tasks. A task named in `CURRENT` must have its matching `workflow/active_work/<ID>.md` file, explicitly approved architecture, edit paths, acceptance, negative cases and verification commands. A roadmap, historical task, handoff narrative or a queued task is not execution authority.
4. `CURRENT: NONE` prohibits implementation, even if planning files exist. This installation intentionally starts with **no promoted package**.
5. Never force-checkout, reset, overwrite local work, guess a branch, force-push, push to `main`, create a PR, merge or deploy as part of CWAL execution. Stop on mismatched markers, divergent refs, unknown authority or unresolved safety risk.

## Context budget / BLACK SHEEP WALL
- Start at CURRENT's named functions/files and relevant standard, protocol vector or test. Read only the relevant excerpt and the next necessary caller/definition/dependency.
- Ask a concrete question before opening more context; stop when the CURRENT acceptance can be implemented and verified.
- Avoid full tree scans, unrelated histories, archived tasks, future task bodies, and full logs. Use focused commands and summarize file/symbol/evidence instead.
- Only a separate authorization may bootstrap or refresh an index/cache. An index does not authorize task execution.
- A discovery requiring scope expansion becomes a recorded blocker, never a silent extra task.

## Execution loop
1. Run `bash scripts/cwal-check.sh`. If no CURRENT, report `NO_CURRENT` and stop.
2. Read **CURRENT body only**. Check its human authorization, KDE architecture decision, file allowlist, safety constraints and acceptance tests. Stop if any are unresolved.
3. Implement one coherent deliverable, including matching tests and documentation. Do not conflate a change to an unsupported feature with completed support.
4. Run task-defined checks and appropriate `go test ./... -count=1`, `go vet ./...`, `go build ./...`, and `./scripts/verify-external-mvp.sh` where applicable. Run race/negative/golden-wire/integration tests when relevant. Record actual PASS, FAIL, NOT_RUN. A test not executed is not evidence.
5. Partial safe progress may be committed and pushed to the designated branch as `TASK_IN_PROGRESS`, leaving CURRENT untouched. Record a concise handoff of exact completed work and blockers.
6. On full acceptance, archive CURRENT under `archive/active-work/<ID>.md` preserving original task requirements; add `docs/verification/<ID>.md` with commands/results, evidence, limitations, SHA; update both current-state marker blocks consistently; commit and push to the designated branch.
7. Fetch/read back the actual remote SHA, archived task and both current-state markers. Only then report `TASK_CHECKPOINTED`. Promote the next **already human-promoted QUEUED** task in the same verified checkpoint if explicitly authorized; read the next task body only after its prior checkpoint is confirmed.
8. On failure keep CURRENT unchanged (apart from truthful `TASK_BLOCKED` status/evidence) and stop. Never skip to a later queued task.

## DNP3-specific acceptance
- Wire format, link CRC (including payload blocks), little-endian fields, function codes, qualifiers, sequence/confirm behavior and fragmentation must have protocol-specification evidence plus golden-vector tests when changed.
- Preserve explicit rejection of unsupported capabilities in `active_work/supported-profile.md`. Do not claim generic IEEE 1815 conformance or third-party interoperability from in-repo TCP tests.
- Exercise malformed-frame, disconnection, timeout, retry/ambiguous-control, bounded-memory and lifecycle cases when affected. No live SCADA writes or uncontrolled field-device commands.
- External compatibility claims require separately recorded independent stack/capture evidence, not merely `verify-external-mvp.sh` PASS.

## Package closure / independent audit
- After final verified CURRENT, set `CURRENT: NONE`, `QUEUED: NONE`, `Status: WORK_PACKAGE_READY` and record remaining limitations. This does not grant PR/merge/deploy authorization.
- An explicitly commissioned independent pre-PR audit is **read-only**: inspect diff from current `main`, merge base, code/protocol correctness, test evidence, scope, secrets, and governance/verification impact. It must not change code, markers or task authority.
- Corrective work needs a **new human-promoted CURRENT** with matching markers. A reported audit finding alone cannot authorize a fix.
- PR creation, merge, release and field-device tests each require separate applicable human authorization.
- Recovery reads only current remote branch/SHA, markers, CURRENT, last concise checkpoint and required unresolved tests. Do not replay archived work.

**Loop:** current markers → current task → smallest necessary context → tests → verified remote checkpoint → next pre-promoted task or stop.
