# Post-gate Backlog

Backlog stubs for capabilities explicitly deferred beyond the v0 interoperability
path. Each section is a planning note only. Nothing here is implemented, verified,
approved, or part of the current supported profile. A stub is not execution
authority: it records questions for a future, separately human-approved design
package.

---

## MEXT-040 — Select-Before-Operate (SBO) profile

Status: **DEFERRED / NOT SUPPORTED** (planning stub only)

### Current state

Select-Before-Operate is **not implemented, not verified, and not approved**.
The v0 supported profile rejects it: `active_work/supported-profile.md` lists
"Select-before-operate" under *Explicitly Unsupported* and dispositions
"Master select-before-operate and direct-operate-no-response" as **Reject**
("Not part of v0."). The only supported control mode is a single Direct Operate
of Group 12 Variation 1 CROB.

This stub does not change the v0 supported profile. The supported-profile
boundary and its rejected-option list remain authoritative and unchanged.

### What SBO would add (if ever approved)

An SBO profile would add a two-step control transaction: the master first issues
a Select for a control point, and only after a successful Select issues the
matching Operate. This is distinct from the v0 Direct Operate path and would
require new public API surface, new wire exchanges, and new negative-case
handling.

### Prerequisites before any design

The following must be resolved and separately approved before implementation:

1. **Explicit design approval.** A recorded KDE/architecture decision for the SBO
   profile, its public API shape, and its supported/unsupported boundary. No SBO
   design is approved by this stub.
2. **Protocol specification evidence.** IEEE 1815 citations for the Select and
   Operate function codes, the control object encoding used, the Select/Operate
   pairing rules, and the sequence/state requirements that bind a Select to its
   Operate.
3. **Supported-profile update.** A revised profile document that moves SBO from
   *Reject* to an approved, bounded capability with its own acceptance criteria.

### Design questions to be researched

- **Select→Operate transaction.** How the master pairs a Select with its Operate,
  how the outstation retains the selected point, and what bounds the lifetime of
  a selection (timeout, new select, disconnect).
- **Command status and timeout handling.** How Select status, Operate status, and
  timeouts are surfaced and classified. Select must not be treated as a completed
  control, and an Operate must not be assumed executed from transport acceptance
  alone.
- **Sequence and state.** How per-point selection state is tracked on both sides,
  how concurrent or interleaved controls are handled, and how sequence numbering
  interacts with the two-step exchange.
- **Rejection and ambiguity.** Behavior when an Operate arrives without a valid
  Select, when a Select is followed by a mismatched Operate, and when the outcome
  of an Operate is unknown (timeout, disconnect). Blind retry of an ambiguous
  control must be avoided.
- **Bounded resources.** Limits on outstanding selections, per-connection state,
  and timers, consistent with the project's bounded-memory and lifecycle rules.

### Required evidence for a future approval

Planning only — none of the following exists for SBO today:

- Golden-wire vector tests for the Select and Operate encodings and their
  LSB-first fields.
- Negative tests for timeout, disconnect, mismatched pairing, and
  ambiguous-command outcomes.
- Lifecycle and concurrency tests (selection state under reconnect and racing
  controls), including race testing where applicable.
- Non-production validation only. No live SCADA or field-device SBO commands may
  be issued.
- Independent, third-party interoperability evidence if any external
  compatibility claim is ever made. In-repo TCP loopback tests do not establish
  external interoperability or full IEEE 1815 conformance.

### Non-claims

This stub asserts no SBO support, no SBO conformance, and no external
interoperability. It does not authorize any Select-Before-Operate request, any
code change, or any change to the v0 supported profile.

---

## MEXT-041 — Event classes / unsolicited responses

Status: **DEFERRED / NOT SUPPORTED** (planning stub only)

### Current state

Event classes and unsolicited responses are **not implemented, not verified, and
not approved**. The v0 supported profile rejects them:
`active_work/supported-profile.md` lists "Unsolicited responses and event
delivery" under *Explicitly Unsupported*, dispositions Master
`EnableUnsolicited`, `DisableUnsolicited`, and `SetUnsolicitedHandler` as
**Reject** ("No unsolicited delivery path."), dispositions Outstation
`WithUnsolicitedMode` and `SetUnsolicitedHandler` as **Reject**, and
dispositions `types.ReadAllEvents` as **Reject** ("Event objects are out of
scope.").

This stub does not change the v0 supported profile. The supported-profile
boundary and its rejected-option list remain authoritative and unchanged. Only
the solicited Class 0 static read path is supported today; the application-layer
SEQ/CON behavior that is verified applies to the solicited request/response
path, not to unsolicited delivery.

### What events and unsolicited would add (if ever approved)

An events/unsolicited profile would add event object classes (Class 1/2/3 event
data), event buffering with retrieval, and unsolicited responses the outstation
sends without a matching master request. This is distinct from the v0 solicited
Class 0 static read path and would require new public API surface, new wire
exchanges, new confirmation/retry semantics, and new negative-case handling.

### Prerequisites before any design

The following must be resolved and separately approved before implementation:

1. **Explicit design approval.** A recorded KDE/architecture decision for the
   events/unsolicited profile, its public API shape, and its supported and
   unsupported boundary. No event or unsolicited design is approved by this
   stub.
2. **Protocol specification evidence.** IEEE 1815 citations for the event object
   groups and variations in scope, the unsolicited response function codes,
   enable/disable semantics, and the CON/SEQ confirmation rules that govern
   unsolicited delivery.
3. **Supported-profile update.** A revised profile document that moves events
   and unsolicited from *Reject* to an approved, bounded capability with its own
   acceptance criteria.

### Design questions to be researched

- **Event class group/variation coverage.** Which event object groups and
  variations (and which event classes) are in scope, how they map to points, and
  how they are bounded and encoded.
- **Unsolicited enable/disable semantics.** How enable/disable is negotiated per
  class, how the outstation tracks enablement state, and what happens on
  reconnect or configuration change.
- **CON/SEQ and application confirmation.** How the unsolicited SEQ numbering,
  CON flag, and application-layer confirmation interact, and how confirmation
  timeout and retry are bounded without duplicating or losing events.
- **Retry and timeout.** How unsolicited retransmission is bounded, how it
  coexists with solicited traffic, and how ambiguous outcomes are handled.
- **Duplicate handling and bounded queues.** How duplicate event delivery is
  detected and suppressed, and how the outstation's event buffer is bounded
  under backpressure, slow masters, and disconnects.
- **Bounded resources.** Limits on buffered events, outstanding confirmations,
  and per-connection state, consistent with the project's bounded-memory and
  lifecycle rules.

### Required evidence for a future approval

Planning only — none of the following exists for events/unsolicited today:

- Golden-wire vector tests for the event object encodings and unsolicited
  response frames, including LSB-first fields.
- Negative tests for confirmation timeout, disconnect mid-unsolicited, duplicate
  delivery, and event-buffer overflow.
- Lifecycle and concurrency tests (enable/disable across reconnect, racing
  solicited and unsolicited traffic), including race testing where applicable.
- Non-production validation only. No live SCADA or field-device unsolicited
  traffic may be issued.
- Independent, third-party interoperability evidence if any external
  compatibility claim is ever made. In-repo TCP loopback tests do not establish
  external interoperability or full IEEE 1815 conformance.

### Non-claims

This stub asserts no event or unsolicited support, no event/unsolicited
conformance, and no external interoperability. It does not authorize any
unsolicited delivery path, any code change, or any change to the v0 supported
profile.
