---
title: "080 - Application Layer"
owner: application-layer
---

# What is the DNP3 Application Layer?

## Purpose

The DNP3 Application Layer is where **user data and operations** live. It defines how data is encoded, what operations are available, and how requests and responses are structured.

## Problem Being Solved

The lower layers (Data Link, Transport) only move bytes. Someone must define:

1. **What bytes mean** - Data types, encoding, structure
2. **What operations exist** - READ, WRITE, OPERATE, etc.
3. **How to express data** - Object groups and variations
4. **How to handle errors** - Application-level responses

The Application Layer provides all of this.

## Application PDU Structure

The Application PDU (APDU) is the unit of application layer communication:

```
┌──────────────────────────────────────────────────────────────────┐
│                       APPLICATION PDU                             │
├──────────────────────────────────────────────────────────────────┤
│  Application Control (1 B) │ Function Code (1 B) │ Objects (...)   │
└──────────────────────────────────────────────────────────────────┘
```

### Fields

| Field | Size | Description |
|-------|------|-------------|
| Application Control | 1 byte | FIR, FIN, CON, UNS, Sequence |
| Function Code | 1 byte | Operation to perform |
| Objects | Variable | Data to read/write/operate on |

## Application Control Field

The Application Control byte provides session-level management:

```
┌────────┬────────┬────────┬────────┬────────┬────────┬────────┬────────┐
│   FIR  │  FIN  │  CON   │  UNS   │        Sequence Number              │
└────────┴────────┴────────┴────────┴────────┴────────┴────────┴────────┘
  Bit 7    Bit 6   Bit 5    Bit 4      Bits 3-0
```

### Field Descriptions

| Bit | Name | Description |
|-----|------|-------------|
| 7 | FIR | First Fragment: 1=first fragment |
| 6 | FIN | Final Fragment: 1=final fragment |
| 5 | CON | Confirmation: 1=confirmation required |
| 4 | UNS | Unsolicited: 1=unsolicited response |
| 3-0 | Sequence | Transaction sequence number (0-15) |

### FIR/FIN in Application Layer

Application layer FIR/FIN operates **independently** of transport layer:

| Transport | Application | Meaning |
|----------|-------------|---------|
| FIR=1, FIN=1 | FIR=1, FIN=1 | Single fragment, complete PDU |
| FIR=1, FIN=0 | (same) | First fragment |
| Multi-fragment | CON=1 | Multi-fragment with confirmation |

### CON (Confirmation Required)

When CON=1:
- Receiver must send application-level confirmation
- Used for critical operations (controls)

When CON=0:
- No confirmation required
- Used for reads, non-critical operations

### UNS (Unsolicited)

When UNS=1:
- This is an unsolicited response
- Outstation initiated (after being enabled)

When UNS=0:
- This is a solicited response
- Master initiated

## Function Codes

Function codes define the operation to perform:

### Master-to-Outstation Functions

| Code | Name | Description |
|------|------|-------------|
| 1 | CONFIRM | Application layer confirmation |
| 2 | READ | Read data from outstation |
| 3 | WRITE | Write data to outstation |
| 4 | SELECT | Select for operate |
| 5 | OPERATE | Execute previously selected operation |
| 6 | DIRECT_OPERATE | Direct operate (select + operate) |
| 7 | DIRECT_OPERATE_NR | Direct operate, no response |
| 8-9 | Reserved | - |
| 10 | FREEZE | Freeze counter values |
| 11-12 | Reserved | - |
| 13 | FILE_OPEN | Open file for transfer |
| 14 | FILE_CLOSE | Close file |
| 15 | FILE_READ | Read file data |
| 16 | FILE_WRITE | Write file data |
| 17-20 | Reserved | - |
| 21 | GET_IDENTIFIER | Get device identification |
| 22 | GET_LABEL | Get device label |
| 23 | GET_DESCRIPTION | Get device description |
| 24 | CHANGE_FILENAME | Change file name |
| 25 | START_UPLOAD | Start file upload |
| 26 | START_DOWNLOAD | Start file download |
| 27 | AUTHENTICATE | Authenticate request |
| 28 | AUTHENTICATE_CONF | Authentication confirmation |
| 29 | ABORT | Abort file transfer |
| 30-31 | Reserved | - |
| 32 | TIME_SYNC | Time synchronization |
| 33 | RECORD_CURRENT_TIME | Record current time |
| 34-36 | Reserved | - |
| 37 | FREEZE_CLEAR | Freeze and clear counters |
| 38 | FREEZE_AT_TIME | Freeze at specified time |
| 39-40 | Reserved | - |
| 41 | ENABLE_UNSOLICITED | Enable unsolicited responses |
| 42 | DISABLE_UNSOLICITED | Disable unsolicited responses |
| 43-47 | Reserved | - |
| 48 | ASSIGN_CLASS | Assign data to classes |
| 49-50 | Reserved | - |
| 51 | DELAY_MEASUREMENT | Measure communication delay |
| 52 | RECORD_BATTERY_VOLTAGE | Record battery voltage |
| 53 | START_RESTART | Initiate device restart |
| 54 | INITIALIZE_APPLICATION | Initialize application |
| 55-56 | Reserved | - |
| 57 | START_SYNCHRONIZATION | Start time synchronization |
| 58 | STOP_SYNCHRONIZATION | Stop time synchronization |
| 59 | CLOCK_SYNC_BROADCAST | Clock sync via broadcast |
| 60-63 | Reserved | - |
| 64-127 | Manufacturer-specific | Vendor-defined functions |
| 128-255 | Reserved | - |

### Outstation-to-Master Responses

| Code | Name | Description |
|------|------|-------------|
| 0 | RESPONSE | Response with data |
| 1 | UNSOLICITED_RESPONSE | Unsolicited response |
| 127 | NO_ACK | No acknowledgment required |

## Response Structure

### Internal Indication (IIN) Field

Every response includes a 2-byte Internal Indication field:

```
┌──────────────────────────────────────────────────────────────────┐
│                     INTERNAL INDICATION (2 B)                     │
├──────────────────────────────────────────────────────────────────┤
│                      IIN.1 (1st byte)                            │
│  ┌──────┬──────┬──────┬──────┬──────┬──────┬──────┬──────┐       │
│  │ ALL  │ BYTE │ 64K  │ 16K  │  MEM │ CHECK│ BUSY │ BRSV │       │
│  │ STOP │ OVER │ LIM  │ LIM  │ AVAIL│ FAIL │      │      │       │
│  └──────┴──────┴──────┴──────┴──────┴──────┴──────┴──────┘       │
│                                                                   │
│                      IIN.2 (2nd byte)                            │
│  ┌──────┬──────┬──────┬──────┬──────┬──────┬──────┬──────┐       │
│  │ ABT  │  ATB │  DLT │  CFG │ MEM │  SYN │  ENA │  EIB │       │
│  │ TRAN │      │ AVAIL│ ERR  │ UNAVA│      │      │      │       │
│  └──────┴──────┴──────┴──────┴──────┴──────┴──────┴──────┘       │
└──────────────────────────────────────────────────────────────────┘
```

### Key IIN Flags

| Flag | Meaning |
|------|---------|
| ALL_STOP | Device is stopped (not operating) |
| BYTE_OVER | Buffer overflow |
| 64K_LIMIT | At 64K data limit |
| 16K_LIMIT | At 16K data limit |
| MEM_UNAVAIL | Requested memory unavailable |
| CHECK_FAIL | Checksum/test failure |
| BUSY | Device is busy |
| BRSV | Parameter/block unavailable |
| ABT_TRAN | File transfer aborted |
| ATB | Analog output block transfer in progress |
| DLT_AVAIL | Data log available |
| CFG_ERR | Configuration error |
| MEM_UNAVAILABLE | Internal memory unavailable |
| SYN | Clock needs synchronization |
| ENA | General enable off |
| EIB | Internal indicator block missing |

## Object Format

Application layer objects follow a specific format:

```
┌──────────────────────────────────────────────────────────────────┐
│                        OBJECT PREFIX                              │
├──────────────────────────────────────────────────────────────────┤
│  Group (1 B) │ Variation (1 B) │ Qualifier │ Object Count        │
└──────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────────┐
│                        OBJECT DATA                                │
├──────────────────────────────────────────────────────────────────┤
│  Variation-specific data encoding                                  │
└──────────────────────────────────────────────────────────────────┘
```

See [090-object-model.md](090-object-model.md) for detailed object format.

## Request-Response Pattern

### Solicited Request

```
Master ──── READ (Group 1, Variation 1) ───────────► Outstation
     │
     │ ◄─── RESPONSE ────────────────────────────────
     │       Objects with values
     │
     │ ──── CONFIRM ─────────────────────────────────► (if CON=1)
```

### Unsolicited Response

```
Master ──── ENABLE_UNSOLICITED ───────────────────► Outstation
     │
Outstation ◄─── EVENT_DETECTED ──────────────────────
     │        (UNS=1, CON=1)
     │
     │ ──── CONFIRM ─────────────────────────────────►
```

### Control Operation

```
Master ──── SELECT (Group 12, Variation 1) ──────────► Outstation
     │
     │ ◄─── RESPONSE ────────────────────────────────
     │       (Select successful)
     │
     │ ──── OPERATE ─────────────────────────────────►
     │
     │ ◄─── RESPONSE ────────────────────────────────
     │       (Operation complete)
```

## Confirmation Behavior

### When CON=1

The receiver must send a **CONFIRM** function code:

```
Master ──── REQUEST (CON=1) ────────────────────────► Outstation
     │
     │ ◄─── RESPONSE ────────────────────────────────
     │
     │ ◄─── CONFIRM ────────────────────────────────
     │
     │ ──── ACK ────────────────────────────────────►
```

### Timeout Behavior

If confirmation not received:

1. Sender retries original request
2. Retry count configurable
3. After max retries, report error

## Common Mistakes

### Mistake 1: Confusing Application and Transport FIR/FIN

**Problem**: Application layer CON confused with transport layer confirmation.

**Fix**: They're independent. Application CON is for application-level confirmation.

### Mistake 2: Wrong Sequence Number Scope

**Problem**: Sequence number resets too often.

**Fix**: Sequence number is per-session. Only reset on session start or explicit reset.

### Mistake 3: Ignoring IIN Flags

**Problem**: Critical status information in IIN is ignored.

**Fix**: Always check IIN flags in responses.

### Mistake 4: Not Handling Unsolicited

**Problem**: Master can't handle unsolicited responses.

**Fix**: Masters must handle UNS=1 responses, including confirming them.

## Engineering Notes

### Performance Implications

| Aspect | Impact |
|--------|--------|
| Small requests | Efficient, low latency |
| Large object reads | Can approach transport limit |
| Many objects | Multiple requests may be needed |
| Confirmation | Adds round-trip latency |

### Reliability Implications

- Application confirmation provides end-to-end reliability
- IIN flags provide device status
- Function codes have specific semantics - implement correctly

### Implementation Notes

1. **Parse completely**: Don't make assumptions about object format
2. **Validate objects**: Check group, variation, qualifiers
3. **Handle all functions**: Even unsupported functions should return proper error
4. **Track sequence**: CON requires sequence tracking

## Relationships

- **Parent**: [050-layer-model.md](050-layer-model.md)
- **Precedes**: [090-object-model.md](090-object-model.md), [230-function-codes.md](230-function-codes.md)
- **Related**: [070-transport-layer.md](070-transport-layer.md), [170-unsolicited-responses.md](170-unsolicited-responses.md)

## References

- IEEE 1815-2012 Section 7: Application Layer
- IEEE 1815-2012 Section 7.2: Application PDU Structure
- IEEE 1815-2012 Section 7.3: Function Codes
- IEEE 1815-2012 Section 7.4: Data Encoding
