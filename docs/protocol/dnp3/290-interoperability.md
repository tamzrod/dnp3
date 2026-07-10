---
title: "290 - Interoperability"
owner: interoperability
---

# What is DNP3 Interoperability?

## Purpose

DNP3 Interoperability ensures that **devices from different vendors** can communicate correctly. This document covers requirements and testing for multi-vendor environments.

## Problem Being Solved

Multi-vendor SCADA systems need:

1. **Consistent behavior** - Same operation produces same result everywhere
2. **Complete specification** - No ambiguity in protocol
3. **Conformance testing** - Verify implementations meet standard
4. **Common subset** - Features that all devices support

## Interoperability Requirements

### Must Support

All DNP3 devices must support:

| Feature | Description |
|---------|-------------|
| TCP transport | Standard TCP/IP communication |
| Unbalanced mode | Master-outstation pattern |
| Function 2 (READ) | Data reading |
| Function 3 (WRITE) | Data writing |
| Function 41/42 | Unsolicited enable/disable |
| Common variations | G1V1, G1V2, G30V1, etc. |

### Should Support

Most implementations should support:

| Feature | Description |
|---------|-------------|
| Function 4/5 | Select-Operate controls |
| Fragmentation | Large messages |
| Events | Change-based reporting |
| Class polling | Priority-based data collection |
| Time sync | Clock synchronization |

## Common Variations

### Minimum Set for Interoperability

| Group | Variation | Data Type |
|-------|-----------|-----------|
| 1 | 1 or 2 | Binary Input |
| 10 | 1 or 2 | Binary Output |
| 30 | 1, 2, or 3 | Analog Input |
| 40 | 1, 2, or 3 | Analog Output |
| 20 | 1 or 2 | Counter |

### Recommended Set

| Group | Variation | Data Type |
|-------|-----------|-----------|
| 1 | 1, 2 | Binary Input |
| 2 | 1, 2 | Binary Event |
| 10 | 1, 2 | Binary Output |
| 30 | 1, 2, 3 | Analog Input |
| 32 | 1, 2, 3 | Frozen Analog |
| 20 | 1, 2 | Counter |
| 21 | 1, 2 | Frozen Counter |

## Interoperability Testing

### Test Categories

| Category | Description |
|----------|-------------|
| Data encoding | Verify encoding formats |
| Function codes | Test all supported functions |
| Object handling | Verify group/variation responses |
| Event generation | Test change detection |
| Control operations | Test select-operate |
| Time sync | Verify timestamp accuracy |
| Fragmentation | Test large messages |

### Test Scenarios

#### Scenario 1: Basic Data Read

```
Test: Read binary input 0
Expected: Response with Group 1 data
Verification: Encoding matches standard
```

#### Scenario 2: Control Operation

```
Test: Select, then Operate binary output 0
Expected: SELECT success, then OPERATE success
Verification: Control executes correctly
```

#### Scenario 3: Event Generation

```
Test: Change input, verify event
Expected: Event with timestamp and flags
Verification: Event in correct class
```

## Multi-Vendor Considerations

### Address Assignment

| Range | Usage |
|-------|-------|
| 0x0001-0xFFFA | Normal addresses |
| 0xFFFF | Broadcast |

### Variation Handling

Masters should:

1. **Request specific variations** when possible
2. **Handle Variation 0** responses
3. **Validate data encoding** in responses

### Capability Discovery

Masters should:

1. **Test for optional features** before use
2. **Handle NOT_SUPPORTED** responses gracefully
3. **Adapt polling** based on capabilities

## Protocol Deviations

### Common Deviations

| Deviation | Impact | Mitigation |
|-----------|--------|------------|
| Non-standard variations | Parse errors | Request standard variations |
| Wrong flags | Data quality issues | Validate flag values |
| Missing function support | Operation failures | Test before use |
| Incorrect encoding | Data corruption | Validate encoding |

### Vendor-Specific Extensions

DNP3 allows vendor-specific:
- Function codes 64-127
- Object groups 128-255

**Warning**: These may not interoperate.

## Testing Tools

### Protocol Analyzers

- Wireshark DNP3 dissector
- Commercial protocol analyzers

### Test Masters

- Open source DNP3 test tools
- Commercial test masters

### Reference Implementations

- OpenDNP3 (C++)
- dnp3 (Rust)
- Various commercial products

## Certification

### DNP3 Users Group Testing

The DNP3 Users Group provides:
- Conformance test procedures
- Interoperability testing events
- Vendor certification program

### Test Levels

| Level | Description |
|-------|-------------|
| Level 1 | Basic functionality |
| Level 2 | Events and controls |
| Level 3 | Advanced features |
| Level 4 | Full protocol |

## Best Practices

### For Master Implementations

1. **Test specific variations** when possible
2. **Handle all standard variations** in responses
3. **Use conservative timeouts**
4. **Implement retry logic**
5. **Log all communication** for debugging

### For Outstation Implementations

1. **Implement all common variations**
2. **Return NOT_SUPPORTED** for unimplemented features
3. **Validate all inputs**
4. **Set appropriate IIN flags**
5. **Document capabilities**

## Common Interoperability Issues

### Issue 1: Variation Mismatch

**Problem**: Master expects one variation, outstation returns another.

**Solution**: Master should handle any variation. Request specific variations when needed.

### Issue 2: Missing Function Support

**Problem**: Outstation doesn't support required function.

**Solution**: Test capabilities before deployment. Use alternatives.

### Issue 3: Non-Standard Behavior

**Problem**: Implementation doesn't follow standard.

**Solution**: Test with multiple implementations. Report deviations.

## Engineering Notes

### Testing Recommendations

1. **Test with multiple vendors** before deployment
2. **Use conformance test suites** from Users Group
3. **Log all anomalies** for troubleshooting
4. **Keep firmware updated** for bug fixes

### Deployment Recommendations

1. **Verify capabilities** during commissioning
2. **Document configuration** for each device
3. **Test failover scenarios** for redundancy
4. **Monitor interoperability** in production

## Relationships

- **Related**: [300-conformance.md](300-conformance.md), [320-common-misconceptions.md](320-common-misconceptions.md)

## References

- IEEE 1815-2012 Annex B: Conformance
- DNP3 Users Group Technical Guidelines
- DNP3 Users Group Test Procedures
