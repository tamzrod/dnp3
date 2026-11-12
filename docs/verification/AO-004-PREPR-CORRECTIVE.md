# AO-004 Pre-PR Corrective — command status 6 semantics and G41 `FuncWrite` boundary

Package: AO-004 (CURRENT) · Branch: `work` · Baseline: `45d699a` (owner authorization)
Route: `operation cwal` (single bounded corrective microtask)

Scope: two read-only pre-PR audit findings, F1 (status octet 6 naming) and F2 (G41
`FuncWrite` boundary). No independent-peer/PCAP execution, no lab network, no live
control, no PR/merge/release, no supported-profile expansion. `active_work/supported-profile.md`
remains byte-identical (blob `8a96819917ffde2b6e6915b2cfc74371d1a02e06`, G41 UNSUPPORTED).

## 1. Independent source provenance (pinned peer)

Fetched the pinned OpenDNP3 3.1.2 generated enum directly from the approved commit
`26b4c01e4839bbbda8866655e086471c4917ee53`:

- URL: `https://raw.githubusercontent.com/dnp3/opendnp3/26b4c01e4839bbbda8866655e086471c4917ee53/cpp/lib/include/opendnp3/gen/CommandStatus.h`
- SHA-256: `69d02051a0e17756a967dff7617af75e4eb4def772042ec1c10e2fca59829885`
- Authoritative lines:
  - `/// something is stopping the command, often a local/remote interlock` → `HARDWARE_ERROR = 6`
  - `/// command not accepted because the requested function is specifically blocked at the outstation` → `BLOCKED = 15`

Conclusion: wire command status **6 is `HARDWARE_ERROR`**, not `BLOCKED`. The previous
in-repo name (`Blocked = 6`) was a genuine naming defect. This is consistent with
`docs/verification/AO-001.md` §3.3. The numeric wire value is unchanged.

## 2. F1 — command status octet 6 naming, compatibility and tests

### Old vs new

| Layer | Old | New | Wire value |
|-------|-----|-----|-----------|
| `internal/master/master.go` | `CommandStatusBlocked CommandStatus = 6` | `CommandStatusHardwareError CommandStatus = 6`; `CommandStatusBlocked = CommandStatusHardwareError` (deprecated alias) | 6 (unchanged) |
| `pkg/dnp3/types/commands.go` | `ControlBlocked ControlStatus = 6` | `ControlHardwareError ControlStatus = 6`; `ControlBlocked = ControlHardwareError` (deprecated alias); `String()` → `"hardware_error"` | 6 (unchanged) |
| `pkg/dnp3/master/client.go` | `case master.CommandStatusBlocked: return types.ControlBlocked` | `case master.CommandStatusHardwareError: return types.ControlHardwareError` | 6 (unchanged) |

No other numeric codes changed; no distinct wire `BLOCKED` status was invented; no
renumbering. Exported `CommandStatusBlocked` / `ControlBlocked` remain valid at the same
numeric value, so existing callers and tests compile unchanged (non-breaking).

### Tests (F1)

- Updated affected status matrix/label tests to the preferred `HardwareError` name:
  `internal/master/{command_status_test.go,g41v1_control_test.go,g41v1_response_structure_test.go,g41v2_control_test.go,g41v3_control_test.go,g41v4_control_test.go}`,
  `pkg/dnp3/master/{operate_status_test.go,g41v1_public_test.go,g41v2_public_test.go,g41v3_public_test.go,g41v4_public_test.go}`,
  `pkg/dnp3/outstation/{bridge_status_matrix_test.go,analog_status_failclosed_test.go}`,
  `pkg/dnp3/types/types_test.go`.
- New `internal/master/command_status_hardware_error_test.go`:
  - `TestCommandStatusHardwareErrorWireValue` — pins `HardwareError == 6`, alias equality, non-success.
  - `TestParseCommandStatusHardwareError` — response octet `0x06` parses to `HardwareError`, never success.
  - `TestOperateWithStatusHardwareError` — end-to-end master surfaces wire `6` as `HardwareError`.
- New `pkg/dnp3/types/types_test.go::TestControlHardwareErrorAlias` — public `6` naming, alias equality, `String()`.

The status remains a **non-success** explicit failure; it is not conflated with the
ambiguous UNKNOWN/quarantine path.

## 3. F2 — fail-closed G41 `FuncWrite` boundary

### Reachability traced

Incoming application `FuncWrite` → `Outstation.ProcessRequest` (`internal/outstation/outstation.go`
`case al.FuncWrite`) → `handleWrite` → `parseControlObjectHeader` → per-group `switch` →
previously `case 41:` decoded the G41 value and called `DataWriter.WriteAnalogOutput` →
public bridge `internalDataHandler.WriteAnalogOutput` → `CommandHandler.HandleAnalogCommand`.

The approved G41 command profile permits **only `FuncDirectOperate` with response**. The
prior reachable `FuncWrite` path could therefore execute a host analog command for an
application WRITE.

### Change

`internal/outstation/outstation.go` `handleWrite` `case 41:` now rejects before any decode,
handler call or state mutation:

```go
case 41: // Analog Output (Group 41)
    // Fail closed (AO-004 pre-PR corrective F2): the approved G41 command
    // profile permits only FuncDirectOperate WITH RESPONSE. ...
    o.iin.ParameterError = true
    return nil, fmt.Errorf("G41 analog output is not supported via WRITE; use DirectOperate with response")
```

Valid G41 `FuncDirectOperate` one-point `0x28` success/negative handling (`executeControl`) is
untouched. `FuncSelect`, `FuncOperate` and `FuncDirectOperateNoResp` were not newly enabled.
Non-G41 `FuncWrite` (e.g. G12/CROB) is unchanged.

### Tests (F2)

- Replaced the stale success-expecting `internal/outstation/g41v4_control_test.go::TestOutstationG41V4WritePathDecode`
  with `TestOutstationG41V4WritePathRejectsWithNoHandlerCall` (honest rejection: zero handler
  calls, no mutation, no success response, `ParameterError` IIN).
- New `internal/outstation/g41_write_rejection_test.go` exercising the real `ProcessRequest` path:
  - `TestOutstationG41WriteRejectsAllVariationsWithNoHandlerCall` — V1/V2/V3/V4 well-formed WRITE: zero handler calls, no mutation, `ParameterError`.
  - `TestOutstationG41WriteRejectsMalformedAndMultiPoint` — truncated, non-finite, multi-point and unsupported-variation WRITE: zero side effects.
  - `TestOutstationG41WriteDoesNotBreakCROBWrite` — non-G41 G12V1 WRITE still dispatches exactly once.
- Existing DirectOperate regression for each variation remains green (`g41v{1,2,3,4}_control_test.go`).

## 4. Verification results (observed)

| Check | Command | Result |
|-------|---------|--------|
| Format | `gofmt -l <touched go>` | PASS (no output) |
| Build | `go build ./...` | PASS |
| Vet | `go vet ./...` | PASS |
| Full suite | `go test ./... -count=1` | PASS (all packages ok) |
| Race | `go test -race ./internal/master ./internal/outstation ./pkg/dnp3/... ./test/integration -count=1` | PASS |
| External gate | `./scripts/verify-external-mvp.sh` | PASS (tiers 1–3 + advisory) |
| Whitespace | `git diff --check` | PASS |
| CWAL gate | `bash scripts/cwal-check.sh` | OK post-commit (`CURRENT=AO-004`) |
| Profile blob | `git hash-object active_work/supported-profile.md` | `8a96819917ffde2b6e6915b2cfc74371d1a02e06` (unchanged) |

Go toolchain: `go1.22.0 linux/amd64` (container-scoped user install; not on PATH by default).

### Revert-proof mutation probes (observed)

- F2: temporarily restored permissive G41 `FuncWrite` decoding → `TestOutstationG41WriteRejectsAllVariationsWithNoHandlerCall`
  and `.../RejectsMalformedAndMultiPoint` **FAIL** (success response / handler call observed). Reverted.
- F1: temporarily renumbered `CommandStatusHardwareError` to 15 → `TestCommandStatusHardwareErrorWireValue`,
  `TestParseCommandStatusHardwareError`, `TestOperateWithStatusHardwareError` **FAIL**. Reverted.

## 5. Scope and boundaries

Changed production files (allowlisted): `internal/master/master.go`, `internal/outstation/outstation.go`,
`pkg/dnp3/master/client.go`, `pkg/dnp3/types/commands.go`. `pkg/dnp3/outstation/server.go` required no
change. Tests changed only under the allowlisted test roots. No integration file was required.
No `main` push, no PR/merge, no release, no field/NGCP/SCADA/physical I/O, no external-peer execution.

Next: **STOP for independent CHECK** of this corrective. A distinct later `operation cwal` may resume
the original AO-004 isolation/readiness-first plan only after that independent CHECK.
