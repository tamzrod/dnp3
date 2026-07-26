# DNP3 Protocol Expert

**Expert ID**: DNP3-EXPERT-001  
**Domain**: DNP3 (Distributed Network Protocol 3)  
**Version**: 1.0.0  
**Status**: Active  

---

## Overview

This expert contains domain knowledge for IEEE 1815 (DNP3) protocol implementation and analysis.

## Domain Knowledge

### Protocol Characteristics

| Characteristic | Value |
|---------------|-------|
| Standard | IEEE 1815-2012 |
| Transport | TCP/IP, Serial |
| Layers | Data Link, Transport, Application |
| Primary Use | SCADA systems |

### Key Concepts

- **Data Link Layer**: Frame encoding, CRC16, link state machine
- **Transport Layer**: Segmentation, reassembly, flow control
- **Application Layer**: APDU encoding/decoding, function codes, IIN
- **Secure Authentication**: IEC 62351-6 compliant challenge handling

## Rules and Constraints

### Implementation Rules

1. **CRC16 Required**: All data link frames must use CRC16
2. **IIN Bytes**: Every response must include Internal Indications
3. **Function Codes**: Must match IEEE 1815 specification
4. **No Wrapper**: Pure Go implementation, no C dependencies

### Protocol Constraints

| Layer | Constraint |
|-------|------------|
| Data Link | Max 292 bytes per frame |
| Transport | Max 249 bytes per transport segment |
| Application | Max 2048 bytes per APDU |

## Best Practices

### Go Implementation

1. Use `encoding/binary` for protocol parsing
2. Implement `io.Reader`/`io.Writer` interfaces
3. Use context for cancellation
4. Channel-based concurrency for master/outstation

### Testing

1. Use conformance test vectors
2. Test with multiple peer implementations
3. Validate against Wireshark dissector

## Reference Standards

- IEEE 1815-2012: DNP3 Protocol
- IEC 62351-6: Security for SCADA
- IEEE 1815.1: DNP3 Mapping to IEC 61850

## Related Artifacts

| Artifact | Purpose |
|----------|---------|
| internal/dll/ | Data Link Layer implementation |
| internal/tl/ | Transport Layer implementation |
| internal/al/ | Application Layer implementation |
| internal/master/ | Master role implementation |
| internal/outstation/ | Outstation role implementation |

---

**Expert Status**: ACTIVE  
**Last Updated**: 2026-07-26
