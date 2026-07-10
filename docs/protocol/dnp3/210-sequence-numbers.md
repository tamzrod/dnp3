---
title: "210 - Sequence Numbers"
owner: sequence-numbers
---

# What are DNP3 Sequence Numbers?

## Purpose

DNP3 Sequence Numbers provide **message ordering** and **duplicate detection** across all protocol layers. They ensure messages are processed in order and duplicates are identified.

## Problem Being Solved

Network communication has issues:

1. **Reordering** - Packets may arrive out of order
2. **Duplicates** - Retries may create duplicates
3. **Missing messages** - Gaps indicate lost messages

Sequence numbers solve all of these.

## Types of Sequence Numbers

DNP3 uses three types of sequence numbers:

| Type | Layer | Bits | Range | Purpose |
|------|-------|------|-------|---------|
| FCB | Data Link | 1 | 0-1 | Data link confirmation |
| Transport Seq | Transport | 6 | 0-63 | Fragment ordering |
| App Seq | Application | 4 | 0-15 | Transaction ordering |

## Data Link FCB

See [220-fcb.md](220-fcb.md) for detailed FCB specification.

### FCB Overview

```
┌───┬───┬───────────┬───────┬──────────────────────────────┐
│DIR│PRM│    FCB    │  FCV  │            RES               │
└───┴───┴───────────┴───────┴──────────────────────────────┘
            ↑
            │
     Toggles on each
     confirmed frame
```

- Primary toggles FCB on confirmed frames
- Secondary echoes FCB in acknowledgment
- Mismatch indicates retransmission or gap

## Transport Sequence Number

### Format

```
┌────────┬────────┬──────────────────────────────┐
│  FIR   │  FIN   │       Sequence Number        │
│  1 bit │  1 bit │          6 bits              │
└────────┴────────┴──────────────────────────────┘
```

### Range

- 0-63 (6 bits)
- Increments by 1 per fragment
- Wraps from 63 to 0

### Behavior

| Event | Action |
|-------|--------|
| New message (FIR=1) | Start with expected sequence |
| Next fragment | Increment sequence |
| Wrap at 63 | Continue at 0 |

### Duplicate Detection

```
Expected Seq = 5
Received Seq = 5 → Valid, process
Received Seq = 6 → Valid, process
Received Seq = 5 → Duplicate, discard
Received Seq = 7 → Gap detected, error
```

## Application Sequence Number

### Format

Located in Application Control byte:

```
┌────────┬────────┬────────┬────────┬──────────────────────┐
│  FIR   │  FIN   │  CON   │  UNS   │    Sequence (4 bits)  │
└────────┴────────┴────────┴────────┴──────────────────────┘
```

### Range

- 0-15 (4 bits)
- Used for request-response matching

### Matching

```
Master sends: App Seq = 5
Outstation responds: App Seq = 5
Master confirms: App Seq = 5

Matching confirms:
- Request was received
- Response corresponds to request
```

## Sequence Number Relationships

```mermaid
flowchart LR
    subgraph "Data Link"
        DL_FCB[FCB<br/>1 bit]
    end
    
    subgraph "Transport"
        T_SEQ[Transport Seq<br/>6 bits]
    end
    
    subgraph "Application"
        A_SEQ[App Seq<br/>4 bits]
    end
    
    DL_FCB --> T_SEQ --> A_SEQ
    
    Note: Each layer has independent sequence tracking
```

## State Tracking

### Master State

```
┌─────────────────────────────────────────────────────┐
│                    MASTER STATE                       │
├─────────────────────────────────────────────────────┤
│  DL Transmit FCB: 0 or 1                             │
│  DL Receive FCB: 0 or 1                              │
│  Transport Transmit Seq: 0-63                        │
│  Transport Receive Seq: 0-63                         │
│  App Transmit Seq: 0-15                              │
│  App Receive Seq: 0-15                              │
└─────────────────────────────────────────────────────┘
```

### Outstation State

```
┌─────────────────────────────────────────────────────┐
│                  OUTSTATION STATE                     │
├─────────────────────────────────────────────────────┤
│  DL Transmit FCB: 0 or 1                             │
│  DL Receive FCB: 0 or 1                              │
│  Transport Transmit Seq: 0-63                         │
│  Transport Receive Seq: 0-63                         │
│  App Transmit Seq: 0-15                              │
│  App Receive Seq: 0-15                              │
└─────────────────────────────────────────────────────┘
```

## Sequence Reset Conditions

### When Sequences Reset

| Sequence | Reset Condition |
|----------|-----------------|
| DL FCB | RESET_LINK_STATIONS command |
| Transport Seq | Session start |
| App Seq | Application reset (rare) |

### Session Start

```
Connection established
    │
    ▼
Master sends RESET_LINK_STATIONS
    │
    ▼
FCB state cleared on both sides
    │
    ▼
Transport sequences reset to 0
```

## Sequence Number Wrap

### Transport Wrap (0-63)

```
Seq = 63
Next fragment: Seq = 0

Not an error - just continues
```

### Application Wrap (0-15)

```
Seq = 15
Next request: Seq = 0

Not an error - just continues
```

## Error Handling

### Sequence Errors

| Error | Detection | Response |
|-------|-----------|----------|
| Duplicate fragment | Same Seq received | Discard |
| Gap in sequence | Expected ≠ Received | Discard, request retry |
| FCB mismatch | Echo ≠ Expected | Reject, request retransmit |
| App Seq mismatch | Response Seq ≠ Request Seq | Error handling |

### Recovery

```
Error detected
    │
    ▼
Discard invalid message
    │
    ▼
Request retransmission
    │
    ▼
Reset sequence state if needed
```

## Common Mistakes

### Mistake 1: Not Tracking FCB State

**Problem**: FCB confirmation failures.

**Fix**: Explicitly track and toggle FCB.

### Mistake 2: Wrong Sequence on Retry

**Problem**: Retry uses new sequence.

**Fix**: Retry must use same sequence as original.

### Mistake 3: Sequence Drift

**Problem**: Sequences get out of sync.

**Fix**: Reset link on session start. Track state explicitly.

### Mistake 4: Confusing Sequence Types

**Problem**: Using wrong sequence for wrong purpose.

**Fix**: Each sequence has specific purpose. Don't confuse.

## Engineering Notes

### Implementation Notes

1. **Track all sequences**: DL, Transport, App independently
2. **Reset on session**: Clear state on connection
3. **Validate received**: Check expected before processing
4. **Wrap correctly**: Handle modulo arithmetic
5. **Timeout handling**: Reset stale sequences

### Debugging Tips

| Symptom | Likely Cause |
|---------|-------------|
| Duplicate processing | FCB not toggling |
| Messages discarded | Sequence mismatch |
| Resets frequently | Network instability |
| Sequence jumps | Implementation bug |

## Relationships

- **Related**: [060-link-layer.md](060-link-layer.md), [070-transport-layer.md](070-transport-layer.md), [080-application-layer.md](080-application-layer.md), [220-fcb.md](220-fcb.md)

## References

- IEEE 1815-2012 Section 5: Data Link Layer - FCB
- IEEE 1815-2012 Section 6: Transport Layer - Sequence
- IEEE 1815-2012 Section 7: Application Layer - Sequence
