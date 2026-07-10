---
title: "260 - Quality Flags"
owner: quality-flags
---

# What are DNP3 Quality Flags?

## Purpose

DNP3 Quality Flags indicate the **reliability and validity** of measurement data. They tell the master whether the value can be trusted.

## Problem Being Solved

Data quality varies:

1. **Sensor failure** - Value may be invalid
2. **Communication loss** - Data may be stale
3. **Device restart** - Data may have changed
4. **Overflow** - Value may exceed range

Quality flags communicate this.

## Flags Byte Format

Every measurement includes a flags byte:

```
┌────────┬────────┬────────┬────────┬────────┬────────┬────────┬────────┐
│  0x80  │  0x40  │  0x20  │  0x10  │  0x08  │  0x04  │  0x02  │  0x01  │
│   DB   │  RST   │  ROL   │  CLF   │   SB   │  CHT   │  PFO   │ ONLINE │
└────────┴────────┴────────┴────────┴────────┴────────┴────────┴────────┘
```

## Flag Definitions

### ONLINE (0x01)

**Meaning**: Data is fresh and valid.

| Value | Meaning |
|-------|---------|
| Set | Value is online, fresh |
| Clear | Value is offline, stale |

### PFO (0x02) - Past File Overwritten

**Meaning**: Event was overwritten before being sent.

| Value | Meaning |
|-------|---------|
| Set | Events were lost (buffer overflow) |
| Clear | No events lost |

### CHT (0x04) - Change

**Meaning**: Value has changed since last read.

| Value | Meaning |
|-------|---------|
| Set | Value changed |
| Clear | Value unchanged |

### SB (0x08) - Software Configuration Error

**Meaning**: Value affected by software configuration error.

| Value | Meaning |
|-------|---------|
| Set | Configuration error affects value |
| Clear | Configuration OK |

### CLF (0x10) - Communications Lost

**Meaning**: Communication to data source was lost.

| Value | Meaning |
|-------|---------|
| Set | Source communication failed |
| Clear | Communication OK |

### ROL (0x20) - Rollover

**Meaning**: Counter has rolled over.

| Value | Meaning |
|-------|---------|
| Set | Counter rolled over |
| Clear | No rollover |

### RST (0x40) - Restart

**Meaning**: Device was restarted.

| Value | Meaning |
|-------|---------|
| Set | Restart occurred, data may be unreliable |
| Clear | No restart |

### DB (0x80) - Debug/Reserved

**Meaning**: Reserved for future use or debug.

## Common Flag Combinations

### Good Data

```
0x01 (ONLINE only)
```

Value is current and valid.

### Stale Data

```
0x00 (No flags)
```

Value is offline or stale.

### After Restart

```
0x05 (ONLINE | RST)
```

Device restarted but value is current.

### Communication Lost

```
0x11 (ONLINE | CLF)
```

Value may be old, communication failed.

### Counter Overflow

```
0x21 (ONLINE | ROL)
```

Counter value with rollover indication.

### Debug Data

```
0x81 (ONLINE | DB)
```

Debug or reserved flag set.

## Flag Handling

### Master Behavior

Masters should:

1. **Always check flags**: Don't assume data is good
2. **Handle offline data**: Show stale indication to operators
3. **Log flag changes**: Track data quality over time
4. **Clear restart flag**: After verification

### Outstation Behavior

Outstations should:

1. **Set ONLINE for good data**: Normal operation
2. **Clear ONLINE for stale data**: Sensor offline
3. **Set CLF on communication failure**: Source unavailable
4. **Set RST on restart**: Flag for master attention
5. **Set ROL on counter overflow**: Track rollover

## Quality in Events

Events always include quality flags at time of event:

```
Event Record:
├─ Value: 1234.5
├─ Flags: ONLINE (0x01)
└─ Timestamp: 2024-01-15 10:30:45.123
```

## Quality in Static Data

Static data includes current quality:

```
Static Record:
├─ Value: 5678
├─ Flags: ONLINE (0x01)
```

## IIN Flags (Internal Indication)

Separate from measurement flags, IIN indicates device status:

### IIN.1 Flags

| Flag | Bit | Meaning |
|------|-----|---------|
| ALL_STOP | 0 | Device stopped |
| BYTE_OVER | 1 | Buffer overflow |
| 64K_LIMIT | 2 | At 64K limit |
| 16K_LIMIT | 3 | At 16K limit |
| MEM_UNAVAIL | 4 | Memory unavailable |
| CHECK_FAIL | 5 | Self-test failure |
| BUSY | 6 | Device busy |
| BRSV | 7 | Block unavailable |

### IIN.2 Flags

| Flag | Bit | Meaning |
|------|-----|---------|
| ABT_TRAN | 0 | Transfer aborted |
| ATB | 1 | Block transfer active |
| DLT_AVAIL | 2 | Data log available |
| CFG_ERR | 3 | Configuration error |
| MEM_UNAVAILABLE | 4 | Memory issue |
| SYN | 5 | Clock needs sync |
| ENA | 6 | Enable off |
| EIB | 7 | IIN block missing |

## Common Mistakes

### Mistake 1: Ignoring Flags

**Problem**: Assuming all data is good.

**Fix**: Always check ONLINE flag before using value.

### Mistake 2: Not Handling Offline

**Problem**: Using stale values.

**Fix**: Display stale indication, use fallback values.

### Mistake 3: Not Clearing RST

**Problem**: Restart flag never cleared.

**Fix**: After verification, clear or acknowledge restart.

## Engineering Notes

### Performance Considerations

| Aspect | Impact |
|--------|--------|
| Flag checking | Minimal overhead |
| Flag storage | 1 byte per point |
| Flag transmission | 1 byte per value in message |

### Implementation Notes

1. **Validate flags**: Check at reception
2. **Update flags**: Update when conditions change
3. **Store flags**: Maintain with value
4. **Display flags**: Show quality to operators

## Relationships

- **Parent**: [100-measurements.md](100-measurements.md)
- **Related**: [150-events.md](150-events.md), [140-database.md](140-database.md)

## References

- IEEE 1815-2012 Section 7.5.1: Flags
- IEEE 1815-2012 Annex A: Object Types
