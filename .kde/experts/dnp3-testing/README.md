# DNP3 Testing Expert

**Expert ID**: DNP3-TEST-EXPERT-001  
**Domain**: DNP3 Protocol Testing  
**Version**: 1.0.0  
**Status**: Active  

---

## Overview

This expert contains testing domain knowledge for DNP3 protocol implementation verification and conformance validation.

## Domain Knowledge

### Testing Categories

| Category | Focus | Tools |
|----------|-------|-------|
| Unit Tests | Individual functions | Go testing |
| Integration Tests | Component interaction | Mock devices |
| Protocol Tests | DNP3 conformance | Test vectors |
| Interop Tests | Multi-vendor | Live devices |
| Fuzz Tests | Edge cases | go-fuzz |

### Test Vectors

| Source | Coverage | Format |
|--------|----------|--------|
| IEEE 1815 Annexes | Data Link, Transport, Application | Binary |
| Wireshark | Protocol dissection | PCAP |
| TMW AMBDT | Master/outstation | Binary |
| OpenDNP3 | Reference implementation | Go |

### Coverage Targets

| Layer | Target | Notes |
|-------|--------|-------|
| Data Link | 95%+ | All frame types |
| Transport | 95%+ | Segmentation |
| Application | 90%+ | Function codes |
| Secure Auth | 80%+ | SA operations |

## Rules and Constraints

### Testing Rules

1. **Conformance First**: Test against IEEE 1815 test vectors
2. **Round-Trip Validation**: Encode then decode, verify equality
3. **Edge Cases**: Test boundary conditions
4. **Error Handling**: Verify error response correctness
5. **Interop**: Test with at least one other implementation

### Test Design Constraints

| Constraint | Requirement |
|------------|-------------|
| Determinism | Tests must be deterministic |
| Isolation | No test dependencies |
| Cleanup | Resources must be released |
| Naming | Descriptive test names |

## Best Practices

### Unit Testing

1. Table-driven tests for multiple inputs
2. Test both valid and invalid inputs
3. Mock external dependencies
4. Measure code coverage

### Integration Testing

1. Use mock transports
2. Test full APDU round-trip
3. Verify state machines
4. Test timeout handling

### Protocol Testing

1. Use official test vectors
2. Test all function codes
3. Verify IIN handling
4. Test fragmentation/reassembly

### Fuzz Testing

1. Random malformed input
2. Coverage-guided fuzzing
3. Regression test for found bugs
4. Continuous fuzzing in CI

## Reference Standards

- IEEE 1815-2012: Test Vectors
- IEC 62351-6: SA Testing
- Go Testing: `testing` package

## Test Infrastructure

| Component | Purpose |
|-----------|---------|
| mock_transport | In-memory transport |
| mock_device | Device simulator |
| testdata | Test vectors |
| conformance | Vector validation |

## Related Artifacts

| Artifact | Purpose |
|----------|---------|
| KDE-KNOW-003 | Testing Strategy |
| DNP3-EXPERT-001 | General DNP3 protocol |

---

**Expert Status**: ACTIVE  
**Last Updated**: 2026-07-26
