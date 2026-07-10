---
title: "090 - Object Model"
owner: object-model
---

# What is the DNP3 Object Model?

## Purpose

The DNP3 Object Model provides a **standardized way to address data** in outstation databases. Every piece of data has a unique address expressed through the object/group/variation/qualifier system.

## Problem Being Solved

SCADA systems need to:

1. **Address data consistently** - Binary input 5, analog input 10, etc.
2. **Read data efficiently** - Request all binary inputs in one message
3. **Handle different encodings** - Same data type can have multiple representations
4. **Specify ranges** - Read points 0-100, not individual points

The object model solves all of these through a hierarchical addressing scheme.

## Object Hierarchy

The DNP3 object hierarchy has four levels:

```
┌──────────────────────────────────────────────────────────────────┐
│                          OBJECT                                   │
├────────────────┬────────────────┬────────────────┬───────────────┤
│     Group      │   Variation    │   Qualifier    │    Index      │
│   (16 bits)    │   (16 bits)    │   (8 bits)     │  (variable)   │
│                │                │                │               │
│ What type of   │ How the data   │ How to find    │ Which specific│
│ data           │ is encoded     │ the data        │ data point    │
└────────────────┴────────────────┴────────────────┴───────────────┘
```

### Level 1: Group

The **Group** defines the general data type:

| Group | Data Type |
|-------|----------|
| 1 | Binary Input |
| 2 | Binary Input Event |
| 3 | Double-Bit Binary Input |
| 4 | Double-Bit Binary Input Event |
| 10 | Binary Output |
| 11 | Binary Output Event |
| 12 | Binary Output Command |
| 20 | Binary Counter |
| 21 | Frozen Counter |
| 22 | Counter Event |
| 23 | Frozen Counter Event |
| 30 | Analog Input |
| 31 | Analog Input Event |
| 32 | Frozen Analog Input |
| 33 | Frozen Analog Input Event |
| 40 | Analog Output |
| 41 | Analog Output Event |
| 43 | Analog Output Command |
| 50 | Time and Date |
| 51 | Time and Date Last Event |
| 52 | Time and Date Cohort |
| 60 | Analog Input Deadband |
| 61 | Analog Output Deadband |
| 70+ | System management, security, etc. |

### Level 2: Variation

The **Variation** defines the specific encoding format:

For Group 1 (Binary Input):

| Variation | Name | Encoding | Size |
|-----------|------|----------|------|
| 0 | Default | Device-specific | Variable |
| 1 | Single-bit | Bit | 1 bit |
| 2 | With Status | Bit + Flags | 1 byte |

For Group 30 (Analog Input):

| Variation | Name | Encoding | Size |
|-----------|------|----------|------|
| 0 | Default | Device-specific | Variable |
| 1 | 16-bit | Integer | 2 bytes |
| 2 | 32-bit | Integer | 4 bytes |
| 3 | 32-bit SP-FP | IEEE Float | 4 bytes |
| 4 | 64-bit SP-FP | IEEE Float | 8 bytes |
| 5 | 4-byte Float | IEEE Float | 4 bytes |
| 6 | 8-byte Float | IEEE Float | 8 bytes |

### Level 3: Qualifier

The **Qualifier** specifies how to address data:

#### Qualifier Codes

| Code | Name | Format | Description |
|------|------|--------|-------------|
| 0x00 | All | - | All points in group |
| 0x01 | 8-bit index | 1 byte | 0-255 points |
| 0x02 | 16-bit index | 2 bytes | 0-65535 points |
| 0x03 | 32-bit index | 4 bytes | Large ranges |
| 0x04 | 8-bit index + count | 1+1 bytes | Specific range |
| 0x05 | 16-bit index + count | 2+2 bytes | Large ranges |
| 0x06 | 16-bit index + count (pref) | 2+2 bytes | Preferred |
| 0x07 | 8-bit index + count (pref) | 1+1 bytes | Preferred |
| 0x08 | No range | - | No data |
| 0x17 | Object prefix 8-bit | 1 byte prefix | Prefixed |
| 0x27 | Object prefix 16-bit | 2 byte prefix | Prefixed |
| 0x37 | Object prefix + count 8-bit | 1+1 prefix | Prefix + count |
| 0x47 | Object prefix + count 16-bit | 2+2 prefix | Prefix + count |

### Level 4: Index

The **Index** specifies the specific point(s):

```
Range Qualifier: 0x05 (16-bit index + count)
                  │
                  ▼
┌─────────────────┬─────────────────┐
│ Start Index     │   Point Count   │
│   (2 bytes)     │    (2 bytes)    │
│   0x0005        │    0x000A       │
└─────────────────┴─────────────────┘

Reads points 5-14 (10 points total)
```

## Object Header Format

### Request Header

```
┌──────────────────────────────────────────────────────────────────┐
│                         OBJECT HEADER                             │
├──────────────────────────────────────────────────────────────────┤
│  Group (1 B) │ Variation (1 B) │ Qualifier (1 B) │ [Range Data]   │
└──────────────────────────────────────────────────────────────────┘
```

### Response Header with Prefix

```
┌──────────────────────────────────────────────────────────────────┐
│                         OBJECT HEADER                             │
├──────────────────────────────────────────────────────────────────┤
│  Group (1 B) │ Variation (1 B) │ Qualifier (1 B) │ [Prefix Data]  │
├──────────────────────────────────────────────────────────────────┤
│                       OBJECT DATA                                 │
└──────────────────────────────────────────────────────────────────┘
```

## Object Prefix

When qualifiers 0x17, 0x27, 0x37, 0x47 are used, each data point is prefixed with its index:

```
┌──────────────────────────────────────────────────────────────────┐
│                    RESPONSE WITH PREFIX                           │
├──────────────────────────────────────────────────────────────────┤
│  Group=1 │ Var=1 │ Qualifier=0x17 │                              │
├──────────────────────────────────────────────────────────────────┤
│  Index=0 │ Data=1 │ Flags=0x01 │                                  │
│  Index=1 │ Data=0 │ Flags=0x01 │                                  │
│  Index=2 │ Data=1 │ Flags=0x01 │                                  │
│  ...                        │                                     │
└──────────────────────────────────────────────────────────────────┘
```

## Common Read Patterns

### Read All Points in Group

```
Group=1, Variation=0, Qualifier=0x00 (All)

Response: All binary inputs with their default variation
```

### Read Specific Range

```
Group=30, Variation=1, Qualifier=0x06
Index=0x0000, Count=0x000A (Points 0-9)

Response: Analog inputs 0-9, 16-bit encoding
```

### Read Single Point

```
Group=12, Variation=1, Qualifier=0x05
Index=0x0005, Count=0x0001 (Point 5 only)

Response: Binary output 5
```

## Object Descriptor

The combination of Group + Variation forms an **Object Descriptor**:

| Descriptor | Meaning |
|------------|---------|
| G1V1 | Binary Input (single-bit) |
| G1V2 | Binary Input (with flags) |
| G30V1 | Analog Input (16-bit) |
| G30V2 | Analog Input (32-bit) |
| G30V3 | Analog Input (32-bit float) |
| G12V1 | Binary Output (single-bit command) |
| G43V1 | Analog Output (16-bit command) |

## Variation 0 (Default)

Variation 0 means "use the device's default variation":

- Each device can define default encodings
- Allows device-specific optimizations
- Masters should handle any variation in response

**Best Practice**: Always specify variation explicitly. Variation 0 may return unexpected encoding.

## Variation Masks

Some variations use bit masks to extract specific bits:

| Variation | Mask | Bits Used |
|-----------|------|-----------|
| G1V1 | 0x01 | Bit 0 |
| G2V1 | 0x03 | Bits 0-1 |
| G3V1 | 0x03 | Bits 0-1 |

## Request Examples

### Read Binary Inputs 0-15

```
Function: READ
Objects:
  - Group=1, Variation=2, Qualifier=0x06
    Range: Start=0, Count=16
```

### Read All Frozen Counters

```
Function: READ
Objects:
  - Group=21, Variation=0, Qualifier=0x00
```

### Read Analog Input Events

```
Function: READ
Objects:
  - Group=32, Variation=1, Qualifier=0x07
    Range: Start=0, Count=32
```

## Response Examples

### Binary Input Response (G1V2)

```
Group=1, Variation=2, Qualifier=0x17
  Index=0, Value=1, Flags=0x01 (ONLINE)
  Index=1, Value=0, Flags=0x01 (ONLINE)
  Index=2, Value=1, Flags=0x01 (ONLINE)
```

### Analog Input Response (G30V1)

```
Group=30, Variation=1, Qualifier=0x17
  Index=0, Value=1234, Flags=0x01 (ONLINE)
  Index=1, Value=5678, Flags=0x01 (ONLINE)
  Index=2, Value=9012, Flags=0x81 (ONLINE, RESTART)
```

## Qualifier Selection Guidelines

| Situation | Recommended Qualifier |
|-----------|----------------------|
| Read all points | 0x00 |
| Read < 256 points | 0x07 (prefixed) |
| Read 256-65535 points | 0x06 (prefixed) |
| Need explicit indices | 0x06 or 0x07 |
| Write with indices | 0x05 or 0x07 |

## Common Mistakes

### Mistake 1: Wrong Qualifier for Range

**Problem**: Request doesn't specify range correctly.

**Fix**: Match qualifier to range. 0x06 needs 2-byte start + 2-byte count.

### Mistake 2: Ignoring Variation 0

**Problem**: Variation 0 returns unexpected encoding.

**Fix**: Always specify variation explicitly. Handle any variation in responses.

### Mistake 3: Wrong Index Endianness

**Problem**: Index interpreted incorrectly.

**Fix**: Index is big-endian (network order).

### Mistake 4: Not Handling Prefix

**Problem**: Data doesn't include expected index prefix.

**Fix**: Match request qualifier to expected response format.

## Engineering Notes

### Performance Implications

| Aspect | Impact |
|--------|--------|
| Small ranges | Multiple requests may be needed |
| Prefix qualifiers | More bytes per point |
| Large ranges | Can approach transport limits |
| Variation choice | Affects encoding efficiency |

### Memory Implications

| Data Type | Typical Size per Point |
|-----------|------------------------|
| Binary | 1-2 bytes |
| Counter | 4-8 bytes |
| Analog | 2-8 bytes |
| With flags | +1 byte |
| With timestamp | +6-10 bytes |

### Implementation Notes

1. **Validate all fields**: Group, variation, qualifier, range
2. **Support prefix qualifiers**: Often used in responses
3. **Handle variation 0**: May appear in responses
4. **Check index bounds**: Don't access out-of-range points

## Relationships

- **Parent**: [080-application-layer.md](080-application-layer.md)
- **Related**: [100-measurements.md](100-measurements.md), [230-function-codes.md](230-function-codes.md), [250-variations.md](250-variations.md)

## References

- IEEE 1815-2012 Section 7.5: Object Definition
- IEEE 1815-2012 Annex A: Object Types
- DNP3 Users Group Technical Guidelines - Object Library
