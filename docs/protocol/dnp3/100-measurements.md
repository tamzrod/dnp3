---
title: "100 - Measurements"
owner: measurements
---

# What are DNP3 Measurements?

## Purpose

DNP3 measurements are the **data values** exchanged between masters and outstations. This document covers binary inputs, analog inputs, counters, and their representations.

## Problem Being Solved

SCADA systems need to:

1. **Represent physical states** - Breaker on/off, voltage level
2. **Encode data efficiently** - Binary needs 1 bit, not 1 byte
3. **Indicate data quality** - Is the value good, bad, or uncertain?
4. **Track changes** - Only send what changed

Measurements solve these requirements with standardized data types and encoding.

## Measurement Categories

### Binary (Digital) Measurements

Represent two-state information:

| Group | Name | Use Cases |
|-------|------|-----------|
| 1 | Binary Input | Breaker status, switch position, alarm state |
| 2 | Binary Input Event | Change of state events |
| 3 | Double-Bit Binary Input | Positions with intermediate state (Open/Close/Transitioning) |
| 4 | Double-Bit Binary Input Event | Double-bit change events |
| 10 | Binary Output | Controllable outputs (with status) |
| 11 | Binary Output Event | Output change events |

### Analog Measurements

Represent continuous values:

| Group | Name | Use Cases |
|-------|------|-----------|
| 30 | Analog Input | Voltage, current, power, temperature |
| 31 | Analog Input Event | Analog change events |
| 32 | Frozen Analog Input | Snapshot of analog at freeze time |
| 33 | Frozen Analog Input Event | Frozen analog change events |

### Counter Measurements

Represent accumulating values:

| Group | Name | Use Cases |
|-------|------|-----------|
| 20 | Binary Counter | Pulse counts, energy accumulators |
| 21 | Frozen Counter | Counter value at freeze time |
| 22 | Counter Event | Counter change events |
| 23 | Frozen Counter Event | Frozen counter change events |

## Binary Input Encoding

### Variation 1: Single-Bit

```
┌────────────────────────┐
│      Value (1 bit)     │
└────────────────────────┘
0 = OFF / False / Open
1 = ON / True / Closed
```

### Variation 2: With Flags

```
┌────────────────────────┐
│  Value (bit 0)  │ Flags│
│     1 bit       │7 bits│
└────────────────────────┘

Flags (see 260-quality-flags.md):
  0x01 = ONLINE
  0x04 = RESTART
  0x10 = COMM_LOST
  0x20 = ROLLOVER
  0x40 = DISCONTINUITY
```

### Double-Bit Binary (Group 3)

```
┌────────────────────────┐
│   Value (bits 0-1)     │
└────────────────────────┘

00 = Intermediate (transitioning)
01 = Determinate Open
10 = Determinate Closed
11 = Indefinite (bad state)
```

## Analog Input Encoding

### Variation 1: 16-bit Signed Integer

```
┌──────────────────────────────────────────────────────────────────┐
│                        16-bit Signed Integer                      │
├──────────────────────────────────────────────────────────────────┤
│  Sign│ 14 magnitude bits                                          │
│   0  │ 32767 to -32768                                            │
└──────────────────────────────────────────────────────────────────┘

Range: -32,768 to +32,767
Size: 2 bytes
```

### Variation 2: 32-bit Signed Integer

```
┌──────────────────────────────────────────────────────────────────┐
│                        32-bit Signed Integer                      │
├──────────────────────────────────────────────────────────────────┤
│  Range: -2,147,483,648 to +2,147,483,647                         │
│  Size: 4 bytes                                                    │
└──────────────────────────────────────────────────────────────────┘
```

### Variation 3: 32-bit Single-Precision Floating Point

```
┌──────────────────────────────────────────────────────────────────┐
│                    IEEE 754 Single-Precision Float                │
├──────────────────────────────────────────────────────────────────┤
│  Sign │ Exponent │ Mantissa                                      │
│   1   │    8     │    23                                          │
│  ±1.18e-38 to ±3.4e38                                             │
│  Size: 4 bytes                                                    │
└──────────────────────────────────────────────────────────────────┘
```

### Variation 4: 64-bit Double-Precision Floating Point

```
┌──────────────────────────────────────────────────────────────────┐
│                    IEEE 754 Double-Precision Float                │
├──────────────────────────────────────────────────────────────────┤
│  Sign │ Exponent │ Mantissa                                      │
│   1   │   11     │    52                                          │
│  ±2.23e-308 to ±1.8e308                                           │
│  Size: 8 bytes                                                    │
└──────────────────────────────────────────────────────────────────┘
```

### Variation 5: 4-byte Float (Alternative)

Same as Variation 3 (IEEE 754 single-precision).

### Variation 6: 8-byte Float (Alternative)

Same as Variation 4 (IEEE 754 double-precision).

## Counter Encoding

### Variation 1: 16-bit Counter

```
┌──────────────────────────────────────────────────────────────────┐
│                        16-bit Unsigned Counter                   │
├──────────────────────────────────────────────────────────────────┤
│  Range: 0 to 65,535                                               │
│  Size: 2 bytes                                                    │
└──────────────────────────────────────────────────────────────────┘
```

### Variation 2: 32-bit Counter

```
┌──────────────────────────────────────────────────────────────────┐
│                        32-bit Unsigned Counter                    │
├──────────────────────────────────────────────────────────────────┤
│  Range: 0 to 4,294,967,295                                        │
│  Size: 4 bytes                                                    │
└──────────────────────────────────────────────────────────────────┘
```

### Variation 3: 32-bit Delta

Delta value (change since last read) in 32-bit signed integer.

### Variation 5: 16-bit Delta

Delta value in 16-bit signed integer.

### Variation 6: 32-bit Counter without Rollover Flag

Like Variation 2 but without rollover flag.

### Variation 7: 16-bit Counter without Rollover Flag

Like Variation 1 but without rollover flag.

### Variation 8: 32-bit Counter with Flag

Variation 2 plus rollover/overflow flag.

### Variation 9: 16-bit Counter with Flag

Variation 1 plus rollover/overflow flag.

## Flags Byte

Every measurement includes a flags byte indicating quality:

```
┌────────┬────────┬────────┬────────┬────────┬────────┬────────┬────────┐
│  0x80  │  0x40  │  0x20  │  0x10  │  0x08  │  0x04  │  0x02  │  0x01  │
│   DB   │  RST   │  Rol   │  CLF   │   SB   │  CHT   │  PFO   │  ONLINE│
└────────┴────────┴────────┴────────┴────────┴────────┴────────┴────────┘
```

| Flag | Name | Meaning When Set |
|------|------|------------------|
| 0x01 | ONLINE | Data is fresh (always set for good data) |
| 0x02 | PFO | Past file overwritten |
| 0x04 | RST | Device restart occurred |
| 0x08 | SB | Software configuration error |
| 0x10 | CLF | Communications lost |
| 0x20 | ROLLOVER | Counter rolled over |
| 0x40 | DISCONTINUITY | Data discontinuous |
| 0x80 | OVERRANGE | Value beyond sensor range |

See [260-quality-flags.md](260-quality-flags.md) for detailed flag descriptions.

## Timestamps

Events include timestamps. Static data does not.

### Time Representation

DNP3 uses multiple time formats:

| Format | Group | Precision | Range |
|--------|-------|-----------|-------|
| Absolute Time | 50 | Milliseconds since 1970 | Limited |
| Time and Date | 50 | Second + DNP3 date | Good |
| Long Absolute Time | 51 | Milliseconds | Better |
| Unsynchronized Time | 52 | With gap indicator | Best |

### Timestamp Encoding (Group 50, Variation 1)

```
┌──────────────────────────────────────────────────────────────────┐
│                    TIME AND DATE (6 bytes)                         │
├──────────────────────────────────────────────────────────────────┤
│  Milliseconds (2 B) │ Days since Jan 1, 1980 (2 B) │             │
│  0-86399            │ 0-65535 (179 years)          │             │
├──────────────────────────────────────────────────────────────────┤
│  Hours │  Minutes │ Seconds │ Day of Week │  Month  │ Year │      │
│  5 bits│  6 bits  │ 5 bits  │  3 bits    │  4 bits │ 7 bits│    │
└──────────────────────────────────────────────────────────────────┘
```

## Scaled Values

Some analog variations use scaled representations:

### Variation 10: 32-bit Signed Integer (Alternate)

### Variation 11: 32-bit Signed Integer with Flag

### Variation 12: 64-bit Signed Integer

### Variation 13: 64-bit Signed Integer with Flag

## Database Representation

### Point Addressing

Each measurement has an index:

```
Binary Input 0, Binary Input 1, ... Binary Input N
Analog Input 0, Analog Input 1, ... Analog Input N
Counter 0, Counter 1, ... Counter N
```

Indices are **per-group**, not global.

### Default Variations

Each outstation defines default variations for:
- Static data (current values)
- Event data

Masters should handle any variation in responses.

## Common Misconceptions

### Misconception 1: "Flags are optional"

**Reality**: Flags are part of the data. Online/offline status must be checked.

### Misconception 2: "All analog values are floats"

**Reality**: Many implementations use integers with a scaling factor. Float is optional.

### Misconception 3: "Counters are just integers"

**Reality**: Counters have specific behavior (rollover, freeze, etc.) defined by the protocol.

### Misconception 4: "Timestamps are always UTC"

**Reality**: DNP3 doesn't specify timezone. Timestamp is raw milliseconds. Time sync establishes reference.

## Engineering Notes

### Performance Considerations

| Encoding | Bytes | Precision | Use Case |
|----------|-------|-----------|----------|
| 16-bit int | 2 | Good | Low bandwidth |
| 32-bit int | 4 | Better | Most applications |
| 32-bit float | 4 | Very good | Non-integer values |
| 64-bit float | 8 | Excellent | High precision |

### Range vs. Precision Trade-off

- **Integers**: Fixed range, exact values, efficient
- **Floats**: Variable range, approximate values, more bytes

### Implementation Notes

1. **Support all common variations**: 1, 2 for binary; 1, 2, 3 for analog
2. **Check flags**: Don't assume data is good
3. **Handle units**: DNP3 doesn't encode units - document them
4. **Support frozen**: Frozen values are important for billing

## Relationships

- **Parent**: [090-object-model.md](090-object-model.md)
- **Related**: [150-events.md](150-events.md), [260-quality-flags.md](260-quality-flags.md), [270-deadbands.md](270-deadbands.md)

## References

- IEEE 1815-2012 Annex A: Object Types
- IEEE 1815-2012 Table A-1: Binary Input Objects
- IEEE 1815-2012 Table A-2: Analog Input Objects
- IEEE 1815-2012 Table A-3: Counter Objects
