# Knowledge Article: DNP3 Library Architecture

**Article ID**: KDE-KNOW-001  
**Domain**: Technical  
**Version**: 1.0.0  
**Date**: 2026-07-26  
**Status**: Active  

---

## Summary

The go-dnp3 library implements IEEE 1815 (DNP3) protocol in pure Go with no C dependencies.

## Architecture Overview

```
pkg/dnp3/          # Public API
pkg/transport/     # Transport layer
internal/          # Internal implementations
  ├── dll/         # Data Link Layer
  ├── tl/          # Transport Layer
  ├── al/          # Application Layer
  ├── master/      # Master role
  ├── outstation/  # Outstation role
  └── sa/          # Secure Authentication
```

## Layer Responsibilities

### Data Link Layer (DLL)

| Responsibility | Details |
|----------------|---------|
| Frame encoding | CRC16, sync bytes |
| Link state | Master/outstation states |
| Address handling | 0-65534 valid addresses |

### Transport Layer (TL)

| Responsibility | Details |
|----------------|---------|
| Segmentation | Split large APDUs |
| Reassembly | Reconstruct received data |
| Flow control | Prevent buffer overflow |

### Application Layer (AL)

| Responsibility | Details |
|----------------|---------|
| APDU encoding | Binary protocol format |
| Function codes | 40+ function codes |
| IIN handling | Internal Indications |

## Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| Pure Go | No CGO, cross-compile friendly |
| Channel-based | Go concurrency patterns |
| Context support | Cancellation and timeouts |
| Interface-based | Pluggable transport |

## Dependencies

- Standard library only
- No external dependencies
- Go 1.21+ required

## Related Knowledge

- KDE-KNOW-002: Protocol Conformance
- KDE-KNOW-003: Testing Strategy

---

*Generated: 2026-07-26*
