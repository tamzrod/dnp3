# EVT-001 post-merge reconciliation (2026-11-14)

Documentation-only. This file reconciles the merged PR #29 and the completed
EVT-001 PPR-001 F1–F4 / S1–S2 evidence into a truthful owner-reviewable
**closure-readiness decision**. It does **not** close or archive EVT-001, does
**not** promote AO-005, and does **not** change `CURRENT`/`QUEUED`/`Status`.
No tests, builds or protocol executions were performed in this invocation; every
verification result below is cited as previously recorded, or marked `NOT_RUN`
by this invocation.

## 1. Verified starting facts (live GitHub, this invocation)

| Fact | Value | How verified |
|---|---|---|
| PR #29 state | `closed`, `merged: true` | GitHub API `pulls/29` |
| PR #29 merged_at | `2026-11-14T04:21:22Z` | GitHub API `pulls/29` |
| PR #29 head ref / sha | `work` / `ffd03038c36989ec476f3960b6d0d6c2961c3067` | GitHub API `pulls/29` |
| PR #29 merge method | merge commit `81cd6c3d0896abe87b96e60e71e1df54b5b34d9d` | GitHub API `pulls/29` |
| `origin/main` tip | `81cd6c3d0896abe87b96e60e71e1df54b5b34d9d` | `git rev-parse origin/main` |
| Merge commit parents | `7a3f52a4…` (main) + `ffd03038…` (work) | `git log -1 --format='%P'` |
| `origin/main` tree | `8720c7081fa8baabe265bd03fc541da1ac308f14` | `git rev-parse 81cd6c3^{tree}` |
| PR-head tree `ffd0303` | `8720c7081fa8baabe265bd03fc541da1ac308f14` | `git rev-parse ffd0303^{tree}` |
| CI for PR head | run #252, event `pull_request`, conclusion `success` | GitHub API `actions/runs?head_sha=ffd0303` |
| supported-profile blob | `3e592e77ad6e31d5492c609c5fdd62382505904b` (work and `origin/main`) | `git rev-parse <ref>:active_work/supported-profile.md` |
| `origin/work` tip | `2603563c2b868e07304e58627b89ce15fd161ad5` | `git rev-parse origin/work` |

All values match the authorization's "verified starting facts". No mismatch.

**Merge shape (checked, not assumed).** The merge commit's tree is byte-identical
to the PR-head tree (`8720c708…`), i.e. the merge introduced **no** content
change on top of `work`; the `main` parent contributed nothing the `work` tree
did not already carry. The only path that differs between the pre-merge `main`
parent `7a3f52a4…` and `ffd0303` is a document move: `main` still had
`workflow/active_work/AO-004.md`, while `work` had archived it to
`archive/active-work/AO-004.md` (content preserved, +566 lines in the archive
copy). The merged tree does **not** contain `workflow/active_work/AO-004.md`, so
the archive move is the effective result. The S1/S2 code is present on
`origin/main` (`ExperimentalEventDelivery` in `pkg/dnp3/master/client.go` and
`pkg/dnp3/outstation/server.go`; no `FuncGetIdentifier` anywhere).

## 2. Completed EVT-001 work — implemented vs supported vs verified

`docs/verification/evt001/EVT-001-PPR-001.md` is the evidence record. Its
provenance distinction is preserved: **CI**, **agent-run tests**, and
**independent review** are different things.

| Area | Code state | Profile | Evidence class |
|---|---|---|---|
| Timestamped G2/G22/G32 event codec, Class 1/2/3 reads | IMPLEMENTED | not advertised | `UNIT_VERIFIED` + `INDEPENDENT_PEER_VERIFIED` (OpenDNP3 3.1.2, both directions) |
| Class-0 static reads (G1V1/G30V1/G20V1) | IMPLEMENTED | Target | `INTEGRATION_VERIFIED` + `INDEPENDENT_PEER_VERIFIED` |
| CROB G12V1 direct operate | IMPLEMENTED | Target | `UNIT_VERIFIED` + `INDEPENDENT_PEER_VERIFIED` |
| PPR-001-F1 (re-queue solicited events on send failure) | corrected | n/a | `UNIT_VERIFIED` (negative RED pre-fix / GREEN post-fix) |
| PPR-001-F2 (fail closed on un-splittable oversized block) | corrected | n/a | `UNIT_VERIFIED` |
| PPR-001-F3 (propagate application-confirm send failure) | corrected | n/a | `UNIT_VERIFIED` |
| PPR-001-F4 (reserve event slots; ignore-return fixed) | corrected | n/a | `UNIT_VERIFIED` + `INDEPENDENT_PEER_VERIFIED` |
| PPR-001-S1 (public unsolicited/event API Reject by default) | corrected | unchanged (Reject) | `UNIT_VERIFIED` (RED pre-fix / GREEN post-fix) |
| PPR-001-S2 (drop `FuncGetIdentifier` 0x15 alias) | corrected | unchanged | `UNIT_VERIFIED`; alias removal compile-enforced |
| Event/unsolicited delivery | IMPLEMENTED behind explicit opt-in | **Reject** | experimental lab path only; not advertised |

PPR-001 evidence classes (as recorded): F1/F2/F3 `VERIFIED`; F4 `UNIT_VERIFIED` +
`INDEPENDENT_PEER_VERIFIED`; S1/S2 `VERIFIED` (default-reject + canonical-mapping
tests) plus `INDEPENDENT_PEER_VERIFIED` on the pinned peer.

### Evidence provenance and limits (unchanged)

- **CI:** GitHub PR-head run #252 (event `pull_request`) concluded `success` on
  `ffd0303`. This is CI status, not independent review.
- **Agent-run tests:** Go gates and the pinned OpenDNP3 3.1.2 isolated
  `--network none` suites (positive, large-batch N=254/255/256/257/333, negative
  confirms) were run by the implementing agent on the corrected tree; hashes are
  in `EVT-001-PPR-001.md`.
- **Independent review:** the reviewer audited the S1/S2 source and the
  documented evidence but did **not** rerun all peer tests or decode every
  capture. That limitation is preserved here.
- No general IEEE 1815 conformance, third-party interoperability certification,
  production-readiness or field/NGCP claim is made.

## 3. Unresolved acceptance gates and known limitations

- **Published profile not promoted:** events/unsolicited and G41 analog controls
  remain `Reject` in `active_work/supported-profile.md`
  (`3e592e77ad6e31d5492c609c5fdd62382505904b`, byte-identical to `main`).
- **Event delivery is experimental:** reachable only via the explicit,
  non-default `WithExperimentalEventDelivery()` switch; it is not a v0 capability.
- **Indexes above 255** are rejected; relative-time, Group 52 and float event
  variations are out of scope.
- **A partly transmitted solicited response** is delivery-indeterminate and can
  be re-delivered at least once on the next poll.
- **Independent peer evidence is loopback-only** and does not establish external
  interoperability.

## 4. Disposition

**`EVT-001_CLOSURE_READY`** (decision for the owner).

All EVT-001 implementation work named in the CURRENT authorization — PPR-001
F1–F4 and S1–S2 — is corrected, unit-verified and (for F4/S1/S2) independently
peer-verified on the pinned OpenDNP3 3.1.2 loopback lab; the scoped gates pass;
and PR #29 is merged to `main` with the merge tree identical to the verified PR
head. The remaining items are **profile-promotion and scope decisions**, not
open defects:

1. whether to leave event/unsolicited delivery permanently `Reject` (with the
   experimental switch retained) or promote a tested subset in a later,
   separately approved profile; and
2. whether the experimental opt-in remains a permanent API surface or is
   re-scoped.

No unresolved code defect or failing gate blocks closure. This document does not
perform the closure: archiving EVT-001 and changing `CURRENT`/`QUEUED`/`Status`
requires an explicit subsequent human owner authorization with exact markers and
permitted paths. `CURRENT` stays `EVT-001`, `QUEUED: NONE`, `Status:
TASK_IN_PROGRESS`.

## 5. Owner-authorized formal closure — 2026-11-14

The owner subsequently authorized formal EVT-001 closure based on this reconciliation. EVT-001 is archived under `archive/active-work/EVT-001.md`; both handoff marker blocks are set to `CURRENT: NONE`, `QUEUED: NONE`, `Status: WORK_PACKAGE_READY`. This supersedes the *pending closure* status recorded above, without rewriting historical findings. See `docs/verification/EVT-001.md` for closure disposition and verification limits. The published v0 supported profile is unchanged; no AO-005 promotion or next task was authorized.
