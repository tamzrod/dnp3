# Agent and Contributor Rules — tamzrod/dnp3

## Operation routing (DNP3)
- **Operation CWAL:** Read [operation cwal.md](operation%20cwal.md), [BLACK_SHEEP_WALL.md](BLACK_SHEEP_WALL.md), and the matching top marker blocks in `workflow/active_work/README.md` and `handoff.md`. Run `bash scripts/cwal-check.sh`. Execute **only** a human-promoted CURRENT task; `CURRENT: NONE` forbids implementation.
- **BLACK SHEEP WALL:** For scoped investigations, read the minimum relevant code/tests/specification; standalone investigation and independent audit are read-only.
- **KDE authority:** The `.kde/` governance hierarchy and approved design decisions retain authority. CWAL coordinates execution but cannot self-approve architectural or governance changes.
- **Safety and checkpoint:** No automatic PR, merge, push to `main`, live field command, or unverifiable checkpoint. Promote queued tasks only under the exact human-approved package and after remote verification.

## Identity and authority
- Repository: `tamzrod/dnp3`; native Go implementation of IEEE 1815 (DNP3).
- This document is an operational checklist, **not** a replacement for `.kde/` governance, approved architecture, or documented protocol requirements.
- Follow, in order: external governance and applicable safety requirements; repository KDE governance and approved decisions; verified supported profile and tests; this checklist. Report conflicts instead of silently selecting a convenient interpretation.
- Final technical and merge approval rests with repository maintainer `tamzrod`. AI agents may propose, implement on a feature branch, and test; they must not self-approve or bypass required review.

## Mandatory workflow
1. Read the relevant `.kde/` guidance, `docs/architecture/`, `active_work/supported-profile.md`, and affected package tests.
2. Define problem, scope, protocol invariants, failure modes, and acceptance criteria before modifying behavior.
3. For an architectural or protocol change, record evidence and obtain required design approval before implementation.
4. Change only what the task requires. Keep public APIs stable unless an approved change explicitly requires a break.
5. Add or update unit, golden-wire, integration, and negative tests for each affected behavior.
6. Run the relevant checks, report commands and results, and distinguish untested assumptions from verified outcomes.
7. Submit changes through a PR; do not merge or push directly to `main` as part of agent-driven work.

## Protocol and safety invariants
- IEEE 1815 wire encoding, object group/variation, qualifiers, CRC blocks, addressing, sequence numbers, fragmentation, and error semantics must be supported by standards references or tested vectors.
- Preserve LSB-first encoding for DNP3 multi-octet fields and the verified link framing limits; never change wire behavior merely to make an internal test pass.
- Reject unsupported operations explicitly. The v0 interoperability profile is TCP-only with Class-0 G1V1/G30V1/G20V1 reads and G12V1 direct operate. Consult the supported-profile document for the detailed boundary.
- Do not enable TLS, serial, unsolicited delivery, authentication, select-before-operate, or other deferred features without an approved profile, tests, and documented verification.
- Controls must not be treated as successful on transport acceptance alone. Check protocol command status; handle timeout, refusal, disconnect, retries, and ambiguous outcomes without assuming execution.
- Never retry a control blindly when its previous outcome may be unknown. Use explicit policy and idempotency/operational safeguards where applicable.
- Bound memory use, frame sizes, queues, concurrency, deadlines, and retry counts; prevent goroutine leaks and stale request state.
- A malformed, hostile, or disconnected peer must not crash or deadlock the process. Fail closed on unsupported security/transport options; never silently downgrade.
- Do not describe the library as IEEE 1815 conformant, third-party interoperable, or production SCADA-ready without independent evidence and an explicit corresponding acceptance record.

## Go implementation quality
- Keep the implementation native Go with minimal dependencies and no unnecessary CGO or third-party protocol wrapper.
- Use idiomatic Go and `context.Context` for cancellable operations; ensure lifecycle methods and reconnection paths are well-defined.
- Avoid speculative abstractions, unbounded buffering, and behavior changes hidden inside refactors.
- Do not log credentials, sensitive operational payloads, or unnecessary control data. Diagnostic logging must remain optional and silent by default where the current API provides that behavior.
- Document exported API changes and preserve clear, classified error handling.

## Required verification
At minimum, run when applicable:
```sh
gofmt -w <changed-go-files>
go test ./... -count=1
go vet ./...
go build ./...
./scripts/verify-external-mvp.sh
```
- Run race tests for concurrency-sensitive changes (`go test -race ./...`) where supported.
- Add independent capture/third-party interoperability evidence before claiming external compatibility; the in-repo TCP loopback gate alone does **not** establish it.
- Report every failing or skipped check and why. Never claim a check passed unless it was actually run and observed passing.
- If verification cannot be completed, label the PR blocked or unverified; do not invent results.

## Branching, reviews, and changes
- Use narrow feature or fix branches and clear, conventional commit messages.
- PRs must describe the problem, scope, design approval/evidence, tests, risk and rollback, and supported-profile/documentation impact.
- No force-push to shared branches, no bypass of branch protection, no agent self-merge, and no unauthorized changes to governance, CI, release, or security policy.
- Do not modify production devices or issue real SCADA controls as part of automated tests. Use deterministic simulators, fixtures, and isolated test endpoints.

## Source-of-truth hygiene
- `active_work/supported-profile.md` and its evidence references define verified versus unsupported behavior.
- Update relevant docs with implementation, but do not rewrite history or portray planned features as implemented.
- Existing documents can be stale; if documents conflict, surface the discrepancy for maintainer resolution rather than silently overwriting authoritative governance.
