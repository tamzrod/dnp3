# Master Error Taxonomy (v0)

Consumer-facing guide to the errors returned by the v0 public Master API
(`pkg/dnp3/master`). Use this to classify `Read` / `IntegrityPoll` / `Operate`
failures without inspecting error-message strings or importing internal
packages.

> Applies to the **v0 interoperability profile**. See
> [`active_work/supported-profile.md`](../active_work/supported-profile.md) for
> the supported API surface.

## Classify, don't string-match

Every error returned by the public Master API can be reduced to a stable,
comparable [`ErrorCode`](../pkg/dnp3/dnp3.go) with
[`dnp3.ClassifyError`](../pkg/dnp3/dnp3.go). The classifier walks the error
chain (`errors.Is` / `errors.As`), so wrapped errors (e.g.
`"read failed: %w"`) are recognized.

```go
import "dnp3/pkg/dnp3"

resp, err := client.Read(ctx, req)
switch dnp3.ClassifyError(err) {
case dnp3.ErrorCodeTimeout:
    // no response in time — retryable against the same connection
case dnp3.ErrorCodeCRC:
    // corrupted frame on the wire — transient, retryable
case dnp3.ErrorCodeUnsupported:
    // caller asked for a group/function outside the v0 profile — fix the request
case dnp3.ErrorCodeDisconnect:
    // link is dead — reconnect (or build a fresh client) before retrying
case nil:
    // success
default:
    // ErrorCodeSequence / ErrorCodeConfiguration / ErrorCodeCanceled /
    // ErrorCodeBusy / ErrorCodeInvalid / ErrorCodeUnknown
}
```

`ClassifyError(nil) == ErrorCodeUnknown` (i.e. a nil error is "unclassified",
not "success"). Always check `err != nil` first.

## Categories

The category reflects the **underlying failure source**, not the operation that
surfaced it. For example, a `Read` that fails because the TCP link died reports
`ErrorCodeDisconnect`, not "read failed".

| Code | Meaning | Typical trigger | Retryable? |
|------|---------|-----------------|------------|
| `ErrorCodeTimeout` | No response/confirmation within the configured timeout (`WithTimeout`). | Outstation slow/down, network drop mid-request. | Yes — same connection (transient). |
| `ErrorCodeCRC` | A received link frame failed CRC-16 validation. | Line noise, corrupted segment. | Yes — transient; the master bounds retries. |
| `ErrorCodeSequence` | Application-layer sequence mismatch (confirm/response SEQ did not match the outstanding request). | Outstation sent an out-of-order confirm/response. | Yes — the master resets SEQ on retry. |
| `ErrorCodeUnsupported` | The caller requested an object group/variation, function, or option outside the v0 profile. | `Read` of a non-G1.1/G30.1/G20.1 group; `Operate` with `SelectThenOperate`/`DirectOperateNoResponse`; `EnableUnsolicited`; TLS. | No — fix the request. |
| `ErrorCodeDisconnect` | The transport/peer closed the connection (peer drop, idle-timeout, or `Close`). | Outstation restarted, network reset, `IdleTimeout` elapsed. | No — the link is dead. Build a fresh client (v0 `Close` is terminal) and reconnect. |
| `ErrorCodeConfiguration` | Invalid configuration supplied. | Bad addresses, missing transport, invalid timeout. | No — fix the `Config`. |
| `ErrorCodeCanceled` | The caller's context was cancelled while the operation was outstanding. | `ctx` cancelled / deadline exceeded. | No — caller-driven. |
| `ErrorCodeBusy` | A request to the same outstation is already in flight. | Concurrent `Read`/`Operate` on one client (v0 is single-flight). | Yes — after the in-flight call completes. |
| `ErrorCodeInvalid` | Malformed request/response not covered by a more specific category. | Unexpected APDU shape, bad object header. | Case-by-case. |
| `ErrorCodeUnknown` | Fallback for unrecognized errors (incl. caller-supplied). | Anything not matched above. | Case-by-case. |

### NACK

The v0 solicited path has **no separate NACK error code**. DNP3 link-layer NACKs
are not produced by the v0 outstation, and an application-layer confirmation
mismatch (the closest solicited-path analog to a "negative acknowledgment") is
classified as **`ErrorCodeSequence`**. If a future profile adds an explicit NACK
transport condition, it will get its own code; until then, treat
`ErrorCodeSequence` as the "the peer rejected / could not confirm this request"
category.

## Retry guidance

- **Transient & retryable on the same connection:** `ErrorCodeTimeout`,
  `ErrorCodeCRC`, `ErrorCodeSequence`, `ErrorCodeBusy`. The master already
  bounds retries per the configured `WithRetry` budget; surface these to the
  caller only after that budget is exhausted.
- **Not retryable on the same connection:** `ErrorCodeDisconnect`. The v0 TCP
  transport's `Close()` is terminal, so recovery means building a **fresh
  client** and reconnecting. The v0 guarantee is that the master never reports a
  dead link as `Connected` and never blocks (see MEXT-025).
- **Caller-fixable:** `ErrorCodeUnsupported`, `ErrorCodeConfiguration`,
  `ErrorCodeCanceled` — correct the request/config/context, then retry.

## Sentinel errors

Each `ErrorCode` maps to one or more exported sentinel errors in
[`pkg/dnp3`](../pkg/dnp3/dnp3.go) (`ErrTimeout`, `ErrCRC`, `ErrSequenceError`,
`ErrUnsupportedGroup`/`ErrUnsupportedFunction`/`ErrUnsupportedOption`,
`ErrNotConnected`/`ErrClosed`, `ErrConfiguration`/`*ConfigurationError`,
`ErrContextCanceled`, `ErrRequestOutstanding`, `ErrInvalidRequest`/
`ErrInvalidResponse`). Prefer `ClassifyError` over direct `errors.Is` checks in
application code; use the sentinels only when you need to distinguish two
sentinels that share a category.
