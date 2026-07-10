---
title: "230 - Function Codes"
owner: function-codes
---

# What are DNP3 Function Codes?

## Purpose

DNP3 Function Codes define the **operations** that can be performed. Each function code specifies a specific action (READ, WRITE, OPERATE, etc.) and its expected behavior.

## Problem Being Solved

Communication needs standardized operations:

1. **Common vocabulary** - All devices understand same commands
2. **Consistent behavior** - Same function does same thing everywhere
3. **Extensibility** - Room for vendor-specific functions

Function codes provide this vocabulary.

## Function Code Overview

### Master-to-Outstation Functions

| Code | Name | Description |
|------|------|-------------|
| 0 | - | (Reserved) |
| 1 | CONFIRM | Application layer confirmation |
| 2 | READ | Read data from outstation |
| 3 | WRITE | Write data to outstation |
| 4 | SELECT | Select for operate |
| 5 | OPERATE | Execute selected operation |
| 6 | DIRECT_OPERATE | Direct operate |
| 7 | DIRECT_OPERATE_NR | Direct operate, no response |
| 8-9 | - | (Reserved) |
| 10 | FREEZE | Freeze counters |
| 11-12 | - | (Reserved) |
| 13 | FILE_OPEN | Open file |
| 14 | FILE_CLOSE | Close file |
| 15 | FILE_READ | Read file data |
| 16 | FILE_WRITE | Write file data |
| 17-20 | - | (Reserved) |
| 21 | GET_IDENTIFIER | Get device ID |
| 22 | GET_LABEL | Get device label |
| 23 | GET_DESCRIPTION | Get device description |
| 24 | CHANGE_FILENAME | Change file name |
| 25 | START_UPLOAD | Start file upload |
| 26 | START_DOWNLOAD | Start file download |
| 27 | AUTHENTICATE | Authenticate request |
| 28 | AUTHENTICATE_CONF | Authenticate confirm |
| 29 | ABORT | Abort file transfer |
| 30-31 | - | (Reserved) |
| 32 | TIME_SYNC | Time synchronization |
| 33 | RECORD_CURRENT_TIME | Record current time |
| 34-36 | - | (Reserved) |
| 37 | FREEZE_CLEAR | Freeze and clear |
| 38 | FREEZE_AT_TIME | Freeze at time |
| 39-40 | - | (Reserved) |
| 41 | ENABLE_UNSOLICITED | Enable unsolicited |
| 42 | DISABLE_UNSOLICITED | Disable unsolicited |
| 43-47 | - | (Reserved) |
| 48 | ASSIGN_CLASS | Assign data to classes |
| 49-50 | - | (Reserved) |
| 51 | DELAY_MEASUREMENT | Measure delay |
| 52 | RECORD_BATTERY_VOLTAGE | Record battery |
| 53 | START_RESTART | Initiate restart |
| 54 | INITIALIZE_APPLICATION | Initialize application |
| 55-56 | - | (Reserved) |
| 57 | START_SYNCHRONIZATION | Start time sync |
| 58 | STOP_SYNCHRONIZATION | Stop time sync |
| 59 | CLOCK_SYNC_BROADCAST | Broadcast time sync |
| 60-63 | - | (Reserved) |
| 64-127 | - | (Vendor-specific) |

### Outstation-to-Master Responses

| Code | Name | Description |
|------|------|-------------|
| 0 | RESPONSE | Response with data |
| 1 | UNSOLICITED_RESPONSE | Unsolicited response |
| 127 | NO_ACK | No acknowledgment |

## Detailed Function Descriptions

### Function 1: CONFIRM

**Purpose**: Acknowledge application layer request.

**Used by**: Both master and outstation

**Objects**: None

**Behavior**: Indicates receipt and processing of previous message.

### Function 2: READ

**Purpose**: Request data from outstation.

**Used by**: Master only

**Objects**: Object headers specifying data to read

**Response**: RESPONSE with requested data

### Function 3: WRITE

**Purpose**: Write data to outstation.

**Used by**: Master only

**Objects**: Object headers with data to write

**Response**: RESPONSE with status

### Function 4: SELECT

**Purpose**: Reserve a control point for subsequent operate.

**Used by**: Master only

**Objects**: Control command parameters

**Response**: RESPONSE with selection status

### Function 5: OPERATE

**Purpose**: Execute previously selected control.

**Used by**: Master only

**Objects**: Control command parameters (same as SELECT)

**Response**: RESPONSE with operation status

### Function 6: DIRECT_OPERATE

**Purpose**: Select and operate in one request.

**Used by**: Master only

**Objects**: Control command parameters

**Response**: RESPONSE with operation status

### Function 7: DIRECT_OPERATE_NR

**Purpose**: Direct operate, no response expected.

**Used by**: Master only

**Objects**: Control command parameters

**Response**: None (timeout indicates failure)

### Function 10: FREEZE

**Purpose**: Freeze counter values at current value.

**Used by**: Master only

**Objects**: Counter points to freeze

**Response**: RESPONSE with frozen values

### Function 32: TIME_SYNC

**Purpose**: Synchronize outstation clock to master.

**Used by**: Master only

**Objects**: Time value

**Response**: RESPONSE with status

### Function 41: ENABLE_UNSOLICITED

**Purpose**: Enable unsolicited responses for classes.

**Used by**: Master only

**Objects**: Class qualifiers

**Response**: RESPONSE with status

### Function 42: DISABLE_UNSOLICITED

**Purpose**: Disable unsolicited responses for classes.

**Used by**: Master only

**Objects**: Class qualifiers

**Response**: RESPONSE with status

### Function 48: ASSIGN_CLASS

**Purpose**: Assign data points to classes.

**Used by**: Master only

**Objects**: Point references with class assignments

**Response**: RESPONSE with status

### Function 51: DELAY_MEASUREMENT

**Purpose**: Measure round-trip communication delay.

**Used by**: Master only

**Objects**: None (uses timing)

**Response**: RESPONSE with delay measurement

## Response Codes

### Application Response

Not function codes, but status codes in response:

| Code | Name | Description |
|------|------|-------------|
| 0 | SUCCESS | Operation completed |
| 1 | TIMEOUT | Device did not respond |
| 2 | NO_SELECT | SELECT required before OPERATE |
| 3 | BAD_VALUE | Value out of range |
| 4 | NOT_SUPPORTED | Function not supported |
| 5 | ALREADY_ACTIVE | Control already in progress |
| 6+ | - | Reserved |

## Internal Indication (IIN)

Responses include IIN flags indicating device status:

```
IIN.1:
  - ALL_STOP: Device stopped
  - BYTE_OVER: Buffer overflow
  - 64K_LIMIT: At 64K limit
  - 16K_LIMIT: At 16K limit
  - MEM_UNAVAIL: Memory unavailable
  - CHECK_FAIL: Self-test failure
  - BUSY: Device busy
  - BRSV: Parameter unavailable

IIN.2:
  - ABT_TRAN: Transfer aborted
  - ATB: Block transfer in progress
  - DLT_AVAIL: Data log available
  - CFG_ERR: Configuration error
  - MEM_UNAVAILABLE: Memory issue
  - SYN: Clock needs sync
  - ENA: Enable off
  - EIB: IIN block missing
```

## Common Function Sequences

### Read Data

```
READ → RESPONSE
```

### Control Operation

```
SELECT → RESPONSE → OPERATE → RESPONSE → CONFIRM
```

### Time Synchronization

```
TIME_SYNC → RESPONSE
```

### Enable Unsolicited

```
ENABLE_UNSOLICITED → RESPONSE
```

## Common Mistakes

### Mistake 1: OPERATE Without SELECT

**Problem**: Some devices require SELECT first.

**Fix**: Always use SELECT-OPERATE sequence unless device documentation says otherwise.

### Mistake 2: Ignoring IIN Flags

**Problem**: Missing important status information.

**Fix**: Check IIN on every response.

### Mistake 3: Wrong Response Handling

**Problem**: Treating all responses as success.

**Fix**: Check response code, not just presence of response.

## Engineering Notes

### Implementation Notes

1. **Validate function codes**: Don't process unknown functions
2. **Return NOT_SUPPORTED**: For unimplemented functions
3. **Check permissions**: Some functions require security
4. **Log all functions**: For audit and debugging

## Relationships

- **Parent**: [080-application-layer.md](080-application-layer.md)
- **Related**: [110-controls.md](110-controls.md), [180-time-synchronization.md](180-time-synchronization.md)

## References

- IEEE 1815-2012 Section 7.3: Function Codes
- IEEE 1815-2012 Table 7-1: Application Function Codes
- IEEE 1815-2012 Table 7-2: Response Status Codes
