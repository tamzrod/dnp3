# DNP3 — CURRENT Active Work

Repository: tamzrod/dnp3
Package: DNP3-POSTGATE-VERIFICATION-CORRECTION-001
Branch: work/mext-postgate-roadmap
CURRENT: NONE
QUEUED: NONE
Status: WORK_PACKAGE_READY

## Package closure
The Go verification correction is complete. DNP3-058 executed the mandatory Go
gates that MEXT-040/041 had truthfully recorded as NOT_RUN, on execution SHA
`2afbeb5acf0fce9126edb122290b29ba3f532151` with a provisioned official Go
1.22.12 toolchain. Every required check PASSED: `gofmt -l .` clean,
`git diff --check`, `go vet ./...`, `go build ./...`, `go test ./... -count=1`,
`go test -race ./internal/outstation/...`, `./scripts/verify-mvp.sh`,
`./scripts/verify-external-mvp.sh` (Tiers 1-3), and the CI-equivalent
`go test -race -coverprofile ./...`. Evidence is in `docs/verification/DNP3-058.md`;
the task is archived at `archive/active-work/DNP3-058.md`. The historical
MEXT-040/041 evidence remains unchanged and truthful for its time.
`CURRENT: NONE` forbids further implementation until a human promotes a new
package.

## Prior historical package (completed, not reopened)
MEXT-040 (SBO deferral) and MEXT-041 (events/unsolicited deferral) were archived
with documentation-only evidence and a final remote checkpoint at
`16ef1f7278b44f6a005edcfe076b376c0551bb38`. Their original evidence truthfully
records Go checks NOT_RUN at the time; DNP3-058 supersedes those missing checks
with new proof rather than rewriting history. Do not alter those documents.
MEXT-042 remains unpromoted.

## Work branch policy
`work/mext-postgate-roadmap` is the persistent non-main execution branch for this
post-gate roadmap series. It does not grant standing implementation authority.
New packages require fresh human promotion and scope. No automatic PR, merge,
main push, production deployment or field command.

## Support boundary
MEXT-035 recorded in-repo real-TCP v0 Master MVP completion. Unsupported SBO and
unsolicited/event functionality remain deferred. No independent VEC-01/third-party
interoperability proof or complete IEEE 1815 conformance claim. DNP3-056/057/058
correction tasks are completed and archived; do not reopen them.

## Verification state
PASS (actual execution, Go 1.22.12): `gofmt -l .` clean, `git diff --check`,
`go vet ./...`, `go build ./...`, `go test ./... -count=1`,
`go test -race ./internal/outstation/...`, `./scripts/verify-mvp.sh`,
`./scripts/verify-external-mvp.sh`, and the CI-equivalent race+coverage run. See
`docs/verification/DNP3-058.md` for exact commands and results.

## Remaining limitations
No third-party interoperability/capture evidence exists; no IEEE 1815 conformance
claim is made. PR creation, merge, release and field-device tests each require
separate human authorization.
