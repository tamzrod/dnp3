---
title: "010 - History"
owner: knowledge-base
---

# What is the history of DNP3?

## Purpose

Understanding DNP3's history explains **why** the protocol has its specific design decisions. The evolution from early utility protocols to IEEE 1815 reveals the engineering trade-offs that shaped modern DNP3.

## Problem Being Solved

The original DNP3 development (circa 1990) addressed critical industry needs:

1. **Fragmented vendor ecosystem** - No common protocol existed
2. **Proprietary lock-in** - Utilities were tied to single vendors
3. **Limited functionality** - Earlier protocols couldn't handle event data well
4. **No standard data model** - Each vendor defined data differently

## Historical Timeline

### 1990-1995: Origins

**Original DNP3** was developed by **Westronic** (later acquired by GE Harris) in Winnipeg, Canada.

Key drivers:
- Electric utility requirements for substation automation
- Need for standardized inter-vendor communication
- Growing complexity of SCADA systems

### 1995-2000: Industry Adoption

DNP3 gained widespread adoption in the electric utility industry:

- Major vendors adopted DNP3 (ABB, SEL, Schweitzer, etc.)
- Device interoperability improved significantly
- Protocol Users Group established for maintenance

### 2000-2010: Maturation

DNP3 evolved with industry needs:

- Enhanced event handling
- Improved security considerations
- Expanded analog data types
- Better time synchronization

### 2010: IEEE Standardization

**IEEE 1815-2010** was published, making DNP3 an official standard:

- First standardized version of DNP3
- Defined timing profiles
- Established conformance requirements

### 2012: Secure Authentication Added

**IEEE 1815-2012** addressed critical security requirements:

- Added challenge-response authentication
- Defined cryptographic key management
- Specified session security procedures

### 2016: Technical Corrigendum

**IEEE 1815-2012 + TC1** corrected ambiguities:

- Clarified timing requirements
- Fixed contradictory specifications
- Enhanced security definitions

## Key Milestones

| Year | Event | Significance |
|------|-------|-------------|
| 1990 | DNP3 invented | Protocol created by Westronic |
| 1995 | Users Group formed | Industry collaboration begins |
| 2000 | Wide vendor adoption | De facto standard established |
| 2010 | IEEE 1815-2010 | First official standard |
| 2012 | IEEE 1815-2012 | Secure Authentication added |
| 2016 | Technical Corrigendum | Ambiguities resolved |

## Design Philosophy Evolution

### Original Design Principles

The original DNP3 incorporated principles from:

1. **SCADA requirements** - Robustness, determinism, low overhead
2. **Utility operational practices** - Polling patterns, event handling
3. **Technical constraints** - Limited bandwidth, serial communication era
4. **Failure handling** - Clear recovery procedures

### Standards-Based Refinements

IEEE standardization refined the protocol:

1. **Formal specifications** - Removed ambiguity
2. **Conformance testing** - Ensured interoperability
3. **Timing profiles** - Defined performance expectations
4. **Security integration** - Added modern security

## Influences on Protocol Design

### Positive Influences

| Source | Influence on DNP3 |
|--------|------------------|
| Early SCADA protocols | Master-outstation model |
| Electric utility requirements | Event-based reporting |
| Reliability engineering | Confirmation mechanisms |
| IEC standards | Object group/variation concept |

### Design Constraints That Shaped DNP3

1. **Serial communication origins** - Explains conservative frame sizes
2. **Bandwidth limitations** - Influenced binary encoding choices
3. **Device resource constraints** - Affected feature complexity
4. **Critical infrastructure** - Drove reliability requirements

## Common Misconceptions About History

### Misconception 1: "DNP3 is an old, outdated protocol"

**Reality**: DNP3 has evolved continuously. The IEEE 1815-2012 version includes modern security features. The protocol is actively maintained and widely deployed.

### Misconception 2: "DNP3 and IEEE 1815 are different"

**Reality**: IEEE 1815 is the standardized version of DNP3. They are the same protocol; IEEE 1815 is simply the standardized name.

### Misconception 3: "The Users Group controls DNP3"

**Reality**: The Users Group provides guidance and conformance testing, but IEEE holds the official standard. Implementations should reference IEEE 1815.

## Engineering Notes

### Why History Matters for Implementation

Understanding historical context helps explain:

- **Conservative design** - Protocol prioritizes robustness over features
- **Event-based architecture** - Reflects utility operational needs
- **Confirmation mechanisms** - Designed for unreliable networks
- **Binary encoding** - Optimized for serial communication origins

### Lessons Learned

1. **Stability over features** - Protocol changes slowly, deliberately
2. **Interoperability first** - Vendor neutrality was paramount
3. **Operational focus** - Real-world usage shaped every decision
4. **Security as evolution** - Security added later, not afterthought

## Relationships

- **Precedes**: IEEE 1815-2010 standardization
- **Related**: Other SCADA protocols (Modbus, IEC 104)
- **Foundation for**: Modern SCADA architecture

## References

- IEEE 1815-2012: Standard for Communication Performance and Timing Profiles
- DNP3 Users Group: History and Technical Guidelines
- Westronic/GE Harris historical documentation
