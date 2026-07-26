# Knowledge Article: Testing Strategy

**Article ID**: KDE-KNOW-003  
**Domain**: Best Practices  
**Version**: 1.0.0  
**Date**: 2026-07-26  
**Status**: Active  

---

## Summary

Comprehensive testing strategy for the go-dnp3 library ensuring protocol correctness and interoperability.

## Testing Pyramid

```
        ┌─────────┐
        │  E2E    │  End-to-end integration tests
        │ Tests   │
       ┌┴─────────┴┐
       │  Protocol  │  Protocol conformance tests
       │  Tests     │
      ┌┴───────────┴┐
      │   Unit      │  Individual component tests
      │   Tests     │
      └─────────────┘
```

## Test Categories

### Unit Tests

| Layer | Coverage Target | Framework |
|-------|----------------|-----------|
| DLL | 90%+ | Go testing |
| TL | 90%+ | Go testing |
| AL | 90%+ | Go testing |
| SA | 80%+ | Go testing |

### Protocol Tests

- Conformance test vectors
- Edge case validation
- Error handling verification

### Integration Tests

- TCP transport tests
- Mock device tests
- Full APDU round-trip

## Test Infrastructure

| Component | Purpose |
|-----------|---------|
| testutils/mock_transport | In-memory transport |
| testutils/mock_device | Device simulators |
| testutils/testdata | Test vectors |

## Testing Best Practices

### Do

1. Test protocol invariants
2. Use conformance test vectors
3. Test error conditions
4. Validate round-trip encoding

### Don't

1. Mock protocol layers unnecessarily
2. Skip edge cases
3. Test implementation details
4. Ignore protocol specifications

## Continuous Integration

| Check | Frequency |
|-------|-----------|
| Unit tests | Every PR |
| Integration tests | Every PR |
| Conformance tests | Weekly |
| Benchmarks | On release |

## Related Knowledge

- KDE-KNOW-001: Architecture
- KDE-KNOW-002: Conformance

---

*Generated: 2026-07-26*
