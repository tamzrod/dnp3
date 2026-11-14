# AO-004 — PR #27 post-merge governance reconciliation (documentation-only)

Package: AO-004 (CURRENT) · Branch: `work` · Owner authorization commit: `fd831cf`
(`docs(cwal): authorize factual PR #27 post-merge governance reconciliation`, by `tamzrod`,
2026-11-12) · Route: `operation cwal` (single documentation-only reconciliation invocation).

Scope: record the **factual** result of the already-merged PR #27. This document separates the
**draft-only authorization** from the **later actual merge**, records independently verified
GitHub facts, and states explicitly that no new release or interoperability support is implied.
No Go/tests/CI/scripts/governance-rule change, new PR, further merge, release, force push, rebase,
live network, lab peer exchange or field command occurred. `active_work/supported-profile.md`
remains byte-identical.

## 1. Independently verified facts (read-only GitHub API + local git)

All values below were read live during this invocation from the GitHub REST API
(`api.github.com/repos/tamzrod/dnp3/...`) and from a fresh `git fetch`. Where the authorization
document asserted a premise that did not match the live readback, the discrepancy is called out
in §3 rather than silently repeated.

| Fact | Value | Source |
|------|-------|--------|
| Repository | `tamzrod/dnp3` | — |
| Permanent development branch | `work` | live `git/ref/heads/work` |
| PR #27 title | `draft(dnp3): G41V1–V4 analog direct-operate implementation and bounded safety regressions` | `pulls/27` |
| PR #27 state | `closed`, `merged = true`, `draft = false` | `pulls/27` |
| PR #27 `created_at` | `2026-11-11T23:58:26Z` (created as draft) | `pulls/27` |
| PR #27 `merged_at` | `2026-11-12T00:02:01Z` | `pulls/27` |
| PR #27 head SHA | `38c6df0a43e7093613ae4ab70dbd4750a04e1a35` | `pulls/27` |
| PR #27 pre-merge base SHA | `f55278cb1f91fb7e66e82a89458f6de9d777302a` | `pulls/27` |
| PR #27 commits / files | 32 commits, 41 changed files | `pulls/27` |
| GitHub merge commit | `2ff2ffca05c66ba73d8c08a561e1227d301348a1` (`Merge pull request #27 from tamzrod/work`) | `commits/2ff2ffc` |
| Merge commit parents | `f55278cb1f91fb7e66e82a89458f6de9d777302a`, `38c6df0a43e7093613ae4ab70dbd4750a04e1a35` | `commits/2ff2ffc` |
| Draft → ready transition | `ready_for_review` at `2026-11-12T00:01:57Z` | `issues/27/timeline` |
| Merge/close events | `merged` and `closed` at `2026-11-12T00:02:02Z` | `issues/27/timeline` |
| Live `main` ref | `2ff2ffca05c66ba73d8c08a561e1227d301348a1` | `git/ref/heads/main` |
| Live `work` ref | `fd831cfb0e28204bce20fe5dee2a276d7ef5beb2` | `git/ref/heads/work` |
| PR-head CI run #244 | run ID `43123823490`, workflow `CI`, event `pull_request`, `completed` / **`success`**, head `38c6df0a…` | `actions/runs/43123823490` |
| Merge-commit CI run #245 | run ID `43124269983`, workflow `CI`, event `push`, `completed` / **`success`**, head `2ff2ffca05c6…`, branch `main` | `actions/runs/43124269983` |
| `supported-profile.md` blob (`main`) | `8a96819917ffde2b6e6915b2cfc74371d1a02e06` | contents API `?ref=main` |
| `supported-profile.md` blob (`work`) | `8a96819917ffde2b6e6915b2cfc74371d1a02e06` | contents API `?ref=work` |
| PR #26 (historical) | `closed`, `merged = true`, `merged_at 2026-11-11T09:09:57Z`, merge commit `f55278cb1f91fb7e66e82a89458f6de9d777302a` | `pulls/26` |

CI run **#244** (`43123823490`) was a `pull_request`-event run whose recorded conclusion is
`success`; individual job logs were **not** independently inspected. CI run **#245**
(`43124269983`) is a separate `push`-event run on the merge commit — see §3.

## 2. Authorization vs. actual merge (no retroactive approval)

- The owner authorized **creation of a draft PR only**. The draft authorization is recorded in
  the earlier AO-004 task history and is **not** rewritten here.
- The **later recorded GitHub merge** (PR #27 → `main`, merge commit `2ff2ffc`, merged
  `2026-11-12T00:02:01Z`) is a **fact to reconcile**, not evidence that a merge was authorized.
- No intent, operator identity, or merge authority is attributed. The GitHub merge commit
  committer is recorded by the platform as `GitHub`; the PR author account is `tamzrod`. This
  document does not infer who performed the merge action.
- This reconciliation **does not undo, revert or alter** the already-merged PR #27, its commits,
  or `main`. It is append-only governance bookkeeping.

## 3. Discrepancy found during independent verification (honest reporting)

The owner authorization document (§"Independently verified facts to record") states:

> "GitHub connector returned **no separately verified merge-commit CI**; mark it `NOT_VERIFIED`."

Live readback during this invocation **found a merge-commit CI run**: workflow `CI`, run ID
`43124269983` (run number `245`), event `push`, head SHA `2ff2ffca05c66ba73d8c08a561e1227d301348a1`,
branch `main`, status `completed`, conclusion **`success`**, created `2026-11-12T00:02:04Z`. It appears in both `actions/runs?head_sha=2ff2ffc…` (total_count 1) and
`actions/runs?branch=main`.

Per BLACK_SHEEP_WALL independent verification, this document records the **observed** run rather
than the authorization's `NOT_VERIFIED` premise, and flags the divergence for maintainer
resolution. Limits on this observation are stated explicitly:

- Conclusion `success` verifies the **workflow-level** reported result only.
- Individual **job logs were not independently inspected**.
- This does not constitute AO-004 independent-peer G41 interoperability evidence.

## 4. Ancestry and divergence (current, not historical)

- `work` is **1 commit ahead** of the merge base and `main` is **402 commits ahead** of the
  `work` line for the counted refs: `git rev-list --left-right --count origin/main...origin/work`
  = `402  1`.
- The only `work` commit not in `main` is `fd831cf` (this authorization, documentation-only).
- `origin/main` contains the merge commit `2ff2ffc`; `f55278cb…` (PR #26 merge) is an ancestor of
  `main` but is **not** on the local `work` first-parent line.
- Note: the clone is **shallow** (`git rev-parse --is-shallow-repository` = `true`), so historical
  pre-merge divergence numbers from older reports are **historical review observations**, not
  current claims. The counts above are the live values read in this invocation.
- `work` (GitHub) at start of this invocation: `fd831cf`; this reconciliation adds one further
  documentation-only commit and pushes non-force to `work`. No `main` update.

## 5. Current authority markers and profile (unchanged)

- Both top six-line authority marker blocks on `work` and on `main` remain:
  `Repository: tamzrod/dnp3`, `Package: AO-004`, `Branch: work`, `CURRENT: AO-004`,
  `QUEUED: NONE`, `Status: TASK_IN_PROGRESS`.
- Merging PR #27 does **not** constitute AO-004 closure/promotion or a release.
- `active_work/supported-profile.md` is byte-identical (`8a96819917ffde2b6e6915b2cfc74371d1a02e06`)
  on both `main` and `work`: **Group 41 remains UNSUPPORTED** in the published v0 profile.
- The already-merged implementation is **not** AO-004 independent OpenDNP3 peer evidence. Original
  PCAP and packet-derived `.vec` G41 verification remain **NOT RUN**. AO-005 is **not** authorized.

## 6. No new release / interoperability / support claim

This reconciliation grants **no** new release, no interoperability or conformance support, no
production/SCADA readiness, and no NGCP profile. It records governance state only.

## 7. JR's observed checks (PASS / FAIL / NOT_RUN)

| Check | Command / method | Result |
|-------|------------------|--------|
| Preflight remote freshness | `git fetch origin work`; `merge --ff-only` | PASS (`d5542eb..fd831cf`) |
| Authority markers (work) | read both top blocks | PASS (six lines identical) |
| Authority markers (main) | read both top blocks | PASS (six lines identical) |
| Owner authorization ancestry | `git merge-base --is-ancestor fd831cf HEAD` | PASS |
| CWAL gate | `bash scripts/cwal-check.sh` | PASS (`CURRENT=AO-004`) |
| PR #27 facts | GitHub API `pulls/27` | PASS (matches authorization) |
| Merge commit + parents | GitHub API `commits/2ff2ffc` | PASS |
| PR #27 timeline | GitHub API `issues/27/timeline` | PASS (`ready_for_review` → `merged`/`closed`) |
| PR-head CI run #244 | GitHub API `actions/runs/43123823490` | PASS (`success`; job logs NOT_RUN) |
| Merge-commit CI | GitHub API `actions/runs/43124269983` | FOUND (`success`) — see §3 |
| Live refs `work`/`main` | GitHub API `git/ref` | PASS |
| Profile blob `main`/`work` | contents API + `git rev-parse` | PASS (unchanged) |
| PR #26 historical state | GitHub API `pulls/26` | PASS (`merged`) |
| Go build/test/vet/race | — | NOT_RUN (documentation-only; no code changed) |
| `git diff --check` | local | PASS |
| Whitespace / exact paths | `git status --porcelain` | PASS (allowlist only) |

## 8. Changed paths (this reconciliation, allowlisted only)

- New: `docs/verification/AO-004-PR27-GOVERNANCE-RECON.md` (this file).
- Append-only, below unchanged top marker blocks: `workflow/active_work/README.md`, `handoff.md`.
- Narrow PR-state-only phrasing: `active_work/SCV_ANALOG_OUTPUT_ROADMAP.md`.
- No edits under `internal/`, `pkg/`, `test/`, `scripts/`, `AGENTS.md`, `operation cwal.md`,
  `BLACK_SHEEP_WALL.md`, `.kde/`, or any historical archive/verification document.
- `active_work/supported-profile.md` **untouched** (blob `8a96819917ffde2b6e6915b2cfc74371d1a02e06`).

## 9. Exit condition

Reconciliation committed with accurate factual provenance and current state, pushed non-force to
`work`, remote SHA verified. **STOP for a separate independent read-only CHECK.** AO-004 remains
CURRENT and G41 UNSUPPORTED; a subsequent distinct `operation cwal` may resume AO-004 isolated lab
readiness only after CHECK acceptance. This instruction does not authorize external-peer protocol
execution, release, code changes or AO-005 promotion.
