---
title: "250 - Variations"
owner: variations
---

# What are DNP3 Variations?

## Purpose

DNP3 Variations define the **specific encoding format** for data within an object group. Each variation specifies how the data is represented in bytes.

## Problem Being Solved

The same data type (e.g., analog input) may need different encodings:

1. **Precision needs** - Some need more precision than others
2. **Bandwidth constraints** - Smaller encoding uses less bandwidth
3. **Device capability** - Simple devices may only support basic encoding
4. **Accuracy requirements** - Engineering units may need specific precision

Variations solve this by defining standard encoding formats.

## Variation Concept

### Group vs. Variation

```
Group: What type of data (e.g., Binary Input)
Variation: How the data is encoded (e.g., single-bit, with flags)
```

### Example: Binary Input (Group 1)

| Variation | Name | Encoding | Size |
|-----------|------|----------|------|
| 0 | Default | Device-specific | Variable |
| 1 | Single-bit | 1 bit | 1 bit |
| 2 | With flags | 1 bit + flags | 1 byte |

### Example: Analog Input (Group 30)

| Variation | Name | Encoding | Size |
|-----------|------|----------|------|
| 0 | Default | Device-specific | Variable |
| 1 | 16-bit | Signed integer | 2 bytes |
| 2 | 32-bit | Signed integer | 4 bytes |
| 3 | 32-bit SP-FP | IEEE float | 4 bytes |
| 4 | 64-bit DP-FP | IEEE double | 8 bytes |
| 5 | 4-byte float | IEEE float | 4 bytes |
| 6 | 8-byte float | IEEE double | 8 bytes |

## Variation 0 (Default)

### Meaning

Variation 0 means "use the device's default variation."

### Behavior

```
Master requests: READ Group 30, Variation 0
Outstation responds: Group 30, Variation [outstation's default]
```

### Considerations

- Each device defines its own defaults
- Masters should handle any variation in response
- Request specific variations when possible

## Common Variations by Group

### Binary Input (Group 1)

| Variation | Name | Format |
|-----------|------|--------|
| 0 | Default | Device default |
| 1 | Single-bit | 1 bit (0 or 1) |
| 2 | With flags | 1 bit + 1 byte flags |

### Binary Output (Group 10)

| Variation | Name | Format |
|-----------|------|--------|
| 0 | Default | Device default |
| 1 | Single-bit | 1 bit |
| 2 | With flags | 1 bit + 1 byte flags |

### Analog Input (Group 30)

| Variation | Name | Format | Range |
|-----------|------|--------|-------|
| 0 | Default | Device default | - |
| 1 | 16-bit | Int16 | ±32,767 |
| 2 | 32-bit | Int32 | ±2,147,483,647 |
| 3 | 32-bit float | IEEE 754 float | ±3.4e38 |
| 4 | 64-bit float | IEEE 754 double | ±1.8e308 |

### Counter (Group 20)

| Variation | Name | Format | Range |
|-----------|------|--------|-------|
| 0 | Default | Device default | - |
| 1 | 16-bit | UInt16 | 0-65,535 |
| 2 | 32-bit | UInt32 | 0-4,294,967,295 |
| 5 | 16-bit delta | Int16 delta | ±32,767 |
| 6 | 32-bit delta | Int32 delta | ±2,147,483,647 |

### Time and Date (Group 50)

| Variation | Name | Format |
|-----------|------|--------|
| 1 | Time and Date | 6 bytes |
| 2 | Time and Date with Quality | 7 bytes |
| 3 | Unsynchronized Time | 6 bytes |

## Encoding Formats

### Integer Encoding

#### Unsigned Integer

```
8-bit:   0 to 255
16-bit:  0 to 65,535
32-bit:  0 to 4,294,967,295
```

#### Signed Integer (Two's Complement)

```
8-bit:   -128 to 127
16-bit:  -32,768 to 32,767
32-bit:  -2,147,483,648 to 2,147,483,647
```

### Floating Point Encoding

#### IEEE 754 Single Precision (32-bit)

```
Sign: 1 bit
Exponent: 8 bits (biased by 127)
Mantissa: 23 bits

Range: ±1.18e-38 to ±3.4e38
Precision: ~7 decimal digits
```

#### IEEE 754 Double Precision (64-bit)

```
Sign: 1 bit
Exponent: 11 bits (biased by 1023)
Mantissa: 52 bits

Range: ±2.23e-308 to ±1.8e308
Precision: ~15 decimal digits
```

## Variation Selection

### Factors

| Factor | Recommendation |
|--------|----------------|
| Required precision | Higher precision → larger variation |
| Bandwidth constraints | Lower bandwidth → smaller variation |
| Device capability | Check device documentation |
| Engineering units | Match to process requirements |

### Recommendations

| Use Case | Recommended Variation |
|----------|----------------------|
| Binary status | G1V1 or G1V2 |
| Simple measurements | G30V1 (16-bit) |
| Precise measurements | G30V2 (32-bit) or G30V3 (float) |
| Very precise measurements | G30V4 (double) |
| Energy accumulation | G20V2 (32-bit counter) |

## Variation in Events

### Event Variations

Events typically include timestamp:

| Group | Variation | Content |
|-------|-----------|---------|
| 2 | 1 | Value + flags |
| 2 | 2 | Value + flags + timestamp |
| 2 | 3 | Value + flags + relative time |
| 32 | 1 | 16-bit + flags + timestamp |
| 32 | 2 | 32-bit + flags + timestamp |
| 32 | 3 | Float + flags + timestamp |

## Response Handling

### Masters Should

1. **Handle any variation** in responses
2. **Request specific variations** when possible
3. **Validate encoding** matches expected format
4. **Convert values** to engineering units if needed

### Outstations Should

1. **Implement common variations** (at minimum)
2. **Return default** when Variation 0 requested
3. **Document supported variations**
4. **Return NOT_SUPPORTED** for unimplemented variations

## Common Mistakes

### Mistake 1: Assuming Variation 0

**Problem**: Variation 0 behavior varies by device.

**Fix**: Always request specific variations when possible.

### Mistake 2: Wrong Integer Signedness

**Problem**: Treating unsigned as signed or vice versa.

**Fix**: Match variation to intended type.

### Mistake 3: Float Precision Mismatch

**Problem**: Assuming float has unlimited precision.

**Fix**: Float has ~7 digits precision. Use double for more.

## Engineering Notes

### Performance Considerations

| Variation | Bytes | Bandwidth | Precision |
|-----------|-------|-----------|-----------|
| 16-bit int | 2 | Low | ±1 |
| 32-bit int | 4 | Medium | ±1 |
| 32-bit float | 4 | Medium | ~7 digits |
| 64-bit float | 8 | High | ~15 digits |

### Implementation Notes

1. **Support minimum set**: G1V1/2, G30V1/2/3, G20V1/2
2. **Validate all inputs**: Check encoding matches variation
3. **Document defaults**: What does Variation 0 return?

## Relationships

- **Parent**: [090-object-model.md](090-object-model.md)
- **Related**: [100-measurements.md](100-measurements.md), [240-object-groups.md](240-object-groups.md)

## References

- IEEE 1815-2012 Annex A: Object Types
- IEEE 1815-2012 Tables A-1 through A-10
- DNP3 Users Group Technical Guidelines
