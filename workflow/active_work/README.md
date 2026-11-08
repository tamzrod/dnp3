# DNP3 — CURRENT Active Work

Repository: tamzrod/dnp3
Package: DNP3-POSTGATE-VERIFICATION-CORRECTION-001
Branch: work/mext-postgate-roadmap
CURRENT: DNP3-058
QUEUED: NONE
Status: TASK_READY

## Current authorization — mandatory Go verification
The repository owner promoted DNP3-058 on 2026-11-08 because Go checks for MEXT-040/041 were NOT_RUN and the maintainer explicitly requires actual execution. Read `workflow/active_work/DNP3-058.md` and perform only the verification correction. Install/provision a compatible Go toolchain in the executor environment if necessary; a missing toolchain is not a waiver. Required gates must all be executed and PASS, otherwise retain CURRENT as TASK_BLOCKED.

## Prior historical package (completed, not reopened)
MEXT-040 (SBO deferral) and MEXT-041 (events/unsolicited deferral) were correctly archived with documentation-only evidence and a final remote checkpoint at `16ef1f7278b44f6a005edcfe076b376c0551bb38`. Their original evidence truthfully records Go checks NOT_RUN at the time. Do NOT alter those documents; write supplemental actual verification results in `docs/verification/DNP3-058.md`.

## Branch / governance
`work/mext-postgate-roadmap` is the persistent work branch. It does not grant standing implementation authority. No code, wire, protocol, supported-profile, CI, dependency or field-device change is allowed. PR creation, merging, releases and field tests each require separate human authorization. The only current task is DNP3-058. MEXT-042 remains unpromoted.

## Verification and checkpoint boundary
No DNP3-058 Go commands have been run yet; this commit is **TASK_READY only**. Run CWAL startup gate on an exact remote-synced clean checkout; provision Go if absent; execute all mandatory checks in task body, record exact commands/version/exit codes and remote SHA. Archive, close markers, push, read back, and report TASK_CHECKPOINTED only if all gates actually pass. Otherwise TASK_BLOCKED.
