# Knowledge Article: Protocol Conformance

**Article ID**: KDE-KNOW-002  
**Domain**: Standards  
**Version**: 1.0.0  
**Date**: 2026-07-26  
**Status**: Active  

---

## Summary

The go-dnp3 library aims for IEEE 1815-2012 conformance with validated interoperability.

## Standards Compliance

| Standard | Status | Notes |
|----------|--------|-------|
| IEEE 1815-2012 | Target | Core protocol |
| IEC 62351-6 | Target | Secure authentication |
| IEEE 1815.1 | Future | IEC 61850 mapping |

## Conformance Requirements

### Must Have (Required)

- [ ] Data Link Layer: All frame types
- [ ] Transport Layer: Segmentation/reassembly
- [ ] Application Layer: All function codes
- [ ] IIN handling: All flags

### Should Have (Recommended)

- [ ] Secure Authentication (SA)
- [ ] File transfer
- [ ] Time synchronization

### Could Have (Optional)

- [ ] Serial transport
- [ ] DNP3 over TLS

## Validation Strategy

1. **Unit Tests**: Each layer has unit tests
2. **Conformance Tests**: IEEE test vectors
3. **Interop Tests**: Test with other implementations
4. **Wireshark**: Validate protocol dissection

## Test Vectors

Reference test vectors from:
- IEEE 1815-2012 Annexes
- TMW AMBDT test suite
- Wireshark DNP3 dissector

## Known Deviations

| Deviation | Reason | Workaround |
|-----------|--------|------------|
| None documented | - | - |

## Related Knowledge

- KDE-KNOW-001: Architecture
- KDE-KNOW-003: Testing Strategy

---

*Generated: 2026-07-26*
