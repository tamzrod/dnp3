---
title: "300 - Conformance"
owner: conformance
---

# What is DNP3 Conformance?

## Purpose

DNP3 Conformance ensures implementations meet the IEEE 1815 standard. Conformance testing verifies that devices behave as specified.

## Problem Being Solved

Implementations may:

1. **Deviate from standard** - Behavior differs from specification
2. **Implement incorrectly** - Known bugs in implementations
3. **Miss required features** - Incomplete implementations
4. **Use non-standard extensions** - Vendor-specific behaviors

Conformance testing addresses these issues.

## Conformance Levels

### Level 1: Basic Functionality

| Requirement | Description |
|------------|-------------|
| Link layer | Basic framing and CRC |
| Transport | Fragmentation/reassembly |
| READ | Static data reading |
| Variation 0 | Default variations |

### Level 2: Events and Controls

| Requirement | Description |
|------------|-------------|
| Events | Event generation |
| Controls | Select-Operate |
| Unsolicited | Enable/Disable |
| Confirm | Application confirmation |

### Level 3: Advanced Features

| Requirement | Description |
|------------|-------------|
| Time sync | Clock synchronization |
| Class polling | Priority-based polling |
| File transfer | File operations |
| Security | Secure Authentication |

### Level 4: Full Protocol

All features of IEEE 1815-2012.

## Mandatory Requirements

### All Implementations Must

1. **Parse all standard formats** - Handle correct encoding
2. **Return NOT_SUPPORTED** - For unimplemented features
3. **Set appropriate IIN flags** - Indicate device status
4. **Validate inputs** - Check range and format
5. **Follow state machines** - Implement correct behavior

## Conformance Testing

### Test Categories

| Category | Coverage |
|----------|----------|
| Data link tests | Frame format, CRC, addressing |
| Transport tests | Fragmentation, reassembly |
| Application tests | Function codes, objects |
| Event tests | Generation, reporting |
| Control tests | Select-Operate |
| Time tests | Synchronization accuracy |

### Test Methods

| Method | Description |
|--------|-------------|
| Black box | External test stimuli and responses |
| White box | Internal state verification |
| Interop testing | Multi-vendor verification |

## Test Vectors

### Data Encoding Tests

Verify correct encoding for all variations:

```
Input: Binary Input ON
Expected: G1V1: 0x01, G1V2: 0x01 + flags

Input: Analog value 1234.5
Expected: G30V1: 0x04D2, G30V2: 0x000004D2, G30V3: IEEE float
```

### State Machine Tests

Verify correct state transitions:

```
Initial: IDLE
Input: RESET_LINK_STATIONS
Expected: ACK sent, state remains IDLE

Initial: IDLE
Input: CONFIRMED_USER_DATA
Expected: RESPONSE sent
```

### Event Tests

Verify event generation:

```
Setup: Deadband = 10.0
Input: Change from 100.0 to 106.0
Expected: Event generated (change > deadband)

Input: Change from 106.0 to 108.0
Expected: No event (change < deadband)
```

## Conformance Documents

### IEEE 1815 Annex B

The standard specifies conformance requirements in Annex B.

### DNP3 Users Group Testing

The Users Group maintains:
- Test procedures
- Test vectors
- Reference implementations
- Certification program

## Certification Process

### Self-Certification

Vendors may:
1. Perform own testing
2. Document results
3. Claim conformance level
4. No external verification

### Users Group Certification

1. Submit implementation for testing
2. Users Group performs tests
3. Results reviewed
4. Certification granted (if passing)

## Common Conformance Issues

### Issue 1: Wrong Encoding

**Problem**: Data encoded incorrectly.

**Fix**: Verify encoding against standard.

### Issue 2: Missing State Transitions

**Problem**: State machine incomplete.

**Fix**: Implement all valid transitions.

### Issue 3: Incomplete Validation

**Problem**: Invalid inputs not rejected.

**Fix**: Validate all inputs per standard.

### Issue 4: Wrong IIN Flags

**Problem**: Status flags incorrect.

**Fix**: Set flags per specification.

## Testing Tools

### Reference Implementations

- OpenDNP3 reference stack
- Test masters
- Protocol analyzers

### Test Suites

| Suite | Coverage |
|-------|----------|
| Data link tests | Link layer |
| Transport tests | Fragmentation |
| Application tests | Functions, objects |
| Integration tests | Full protocol |

### Wireshark

Wireshark DNP3 dissector helps:
- Verify packet format
- Debug protocol issues
- Analyze conformance

## Implementation Checklist

### Required for All

- [ ] Data link framing
- [ ] CRC calculation
- [ ] Fragmentation/reassembly
- [ ] READ function
- [ ] WRITE function
- [ ] RESPONSE function
- [ ] Standard variations

### Recommended

- [ ] Events
- [ ] Select-Operate
- [ ] Unsolicited
- [ ] Class polling
- [ ] Time sync

### Optional

- [ ] File transfer
- [ ] Secure Authentication
- [ ] All variations

## Engineering Notes

### Quality Assurance

1. **Test against standard** - Not against other implementations
2. **Use test vectors** - Reference inputs/outputs
3. **Test edge cases** - Boundary conditions
4. **Verify error handling** - Invalid inputs

### Documentation

Document:
- Conformance level claimed
- Supported variations
- Known limitations
- Deviations from standard

## Relationships

- **Related**: [290-interoperability.md](290-interoperability.md), [310-performance-considerations.md](310-performance-considerations.md)

## References

- IEEE 1815-2012 Annex B: Conformance
- DNP3 Users Group Testing Procedures
- DNP3 Users Group Technical Guidelines
