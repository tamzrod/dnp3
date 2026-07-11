---
title: "010 - Roadmap"
status: draft
---

# Roadmap

## Overview

This document outlines the development roadmap for go-dnp3. The roadmap 
is organized into phases, with each phase building on the previous one.

## Roadmap Phases

```
┌─────────────────────────────────────────────────────────────┐
│ Phase 7: Production Release                                  │
└─────────────────────────────────────────────────────────────┘
                              ↑
┌─────────────────────────────────────────────────────────────┐
│ Phase 6: Optimization                                       │
└─────────────────────────────────────────────────────────────┘
                              ↑
┌─────────────────────────────────────────────────────────────┐
│ Phase 5: Conformance Testing                                 │
└─────────────────────────────────────────────────────────────┘
                              ↑
┌─────────────────────────────────────────────────────────────┐
│ Phase 4: Implementation 🔄 IN PROGRESS                      │
│ ├── 4.1: Data Link Layer ✅ COMPLETE                        │
│ ├── 4.2: Transport Layer ✅ COMPLETE                        │
│ ├── 4.3: Application Layer ✅ COMPLETE                       │
│ └── 4.4: Secure Authentication ⏳ NEXT                      │
└─────────────────────────────────────────────────────────────┘
                              ↑
┌─────────────────────────────────────────────────────────────┐
│ Phase 3: Architecture Review ✅ COMPLETE                     │
└─────────────────────────────────────────────────────────────┘
                              ↑
┌─────────────────────────────────────────────────────────────┐
│ Phase 2: Architecture ✅ COMPLETE                            │
└─────────────────────────────────────────────────────────────┘
                              ↑
┌─────────────────────────────────────────────────────────────┐
│ Phase 1: Research ✅ COMPLETE                               │
└─────────────────────────────────────────────────────────────┘
```

## Phase 1: Research

**Status**: ✅ Complete

**Completion Date**: 2026-07-10

### Goals

- Complete protocol analysis
- Identify all edge cases
- Document invariants
- Research implementations

### Deliverables

- [x] IEEE 1815-2012 complete analysis
- [x] Data link layer invariants documented
- [x] Transport layer invariants documented
- [x] Application layer invariants documented
- [x] Secure authentication analysis
- [x] Edge case inventory
- [x] Existing implementation analysis

### Timeline

Completed during initial repository development.

### Exit Criteria

- [x] All protocol features analyzed
- [x] Edge cases documented
- [x] Implementation approach defined
- [x] Research documented in protocol knowledge base

## Phase 2: Architecture

**Status**: ✅ Complete

**Completion Date**: 2026-07-10

### Goals

- Define package structure
- Design interfaces
- Document decisions
- Create ADRs

### Deliverables

- [x] Package architecture document
- [x] Data link layer interface design
- [x] Transport layer interface design
- [x] Application layer interface design
- [x] Security architecture
- [x] Error handling design
- [x] Configuration design
- [x] Key ADRs created (ADR-001 through ADR-005)

### Timeline

Completed as part of KDSE Runtime Session.

### Exit Criteria

- [x] Architecture documents complete
- [x] Key decisions documented in ADRs (5 ADRs)
- [x] Interfaces designed
- [x] Dependencies mapped
- [x] Architecture approved

## Phase 3: Architecture Review

**Status**: ✅ Complete

**Completion Date**: 2026-07-10

### Goals

- Review architecture
- Address feedback
- Get approval

### Deliverables

- [x] Team review (KDSE Runtime Session)
- [x] Architecture approved by assessor
- [x] ADRs created and accepted
- [x] Sign-off from architecture authority

### Timeline

Completed as part of KDSE Runtime Session.

### Exit Criteria

- [x] All reviews complete
- [x] No blocking concerns
- [x] Architecture approved
- [x] Ready for implementation

## Phase 4: Implementation

**Status**: 🔄 In Progress

### Goals

- Implement data link layer
- Implement transport layer
- Implement application layer
- Implement secure authentication

### Milestones

#### 4.1 Data Link Layer

- [x] Frame encoding/decoding
- [x] CRC calculation
- [x] Link state machine
- [x] Address handling
- [x] Unbalanced mode
- [x] Balanced mode

**Status**: ✅ Complete (2026-07-10)

#### 4.2 Transport Layer

- [x] Segmentation
- [x] Reassembly
- [x] Header handling
- [x] Flow control

**Status**: ✅ Complete (2026-07-10)

#### 4.3 Application Layer

- [x] PDU encoding/decoding
- [x] All function codes
- [x] Application control field
- [x] IIN (Internal Indication)
- [ ] All object groups (planned)
- [ ] All variations (planned)
- [x] Response generation

**Status**: ✅ Verified Complete (2026-07-11)

#### 4.4 Secure Authentication

- [ ] Challenge handling
- [ ] Key management
- [ ] Session security
- [ ] Cryptographic operations

**Status**: ⏳ Next

### Exit Criteria

- [x] Data Link Layer implemented
- [x] Transport Layer implemented
- [x] Application Layer core implemented
- [ ] Secure Authentication implemented
- [x] Unit tests > 80% coverage
- [ ] Integration tests passing
- [x] No known critical bugs

## Phase 5: Conformance Testing

**Status**: 📋 Planned

### Goals

- Verify protocol compliance
- Test interoperability
- Validate edge cases

### Deliverables

- [ ] Conformance test suite
- [ ] Interop test results
- [ ] Edge case validation
- [ ] Bug fixes

### Timeline

TBD based on implementation.

### Exit Criteria

- All critical conformance tests pass
- Interoperability verified
- Edge cases handled

## Phase 6: Optimization

**Status**: 📋 Planned

### Goals

- Improve performance
- Reduce memory usage
- Optimize hot paths

### Deliverables

- [ ] Performance benchmarks
- [ ] Optimization applied
- [ ] Performance targets met

### Timeline

TBD based on testing.

### Exit Criteria

- Performance targets met
- No regressions
- Benchmarks documented

## Phase 7: Production Release

**Status**: 📋 Planned

### Goals

- First stable release
- Complete documentation
- Community ready

### Deliverables

- [ ] v1.0.0 release
- [ ] Documentation complete
- [ ] Examples provided
- [ ] CHANGELOG updated
- [ ] Release announcement

### Timeline

TBD based on previous phases.

### Exit Criteria

- Stable release available
- Documentation complete
- Examples working
- Community engaged

## Feature Priorities

### P0 (Must Have)

- Data link layer (unbalanced)
- Transport layer
- Application layer (core)
- Master station support
- Outstation support
- Basic configuration

### P1 (Should Have)

- Secure authentication
- Time synchronization
- File transfer
- Balanced mode
- Counter events
- Analog events

### P2 (Nice to Have)

- Serial transport
- Performance optimization
- Hardware acceleration
- Embedded support

## Unknowns

The following are TBD and will be refined as we progress:

- Exact timelines
- Resource allocation
- Feature scope
- Performance targets

## Contributing to Roadmap

The roadmap is a living document:

- Updated based on feedback
- Refined based on progress
- Adjusted based on priorities

To propose changes:

1. Open an issue
2. Discuss in team
3. Update roadmap
4. Document rationale

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 0.1.0 | TBD | Initial roadmap |
| 0.2.0 | 2026-07-10 | KDSE Runtime Session: Phase 1-3 completed, ADRs added |
| 0.3.0 | 2026-07-11 | Phase 4.1-4.3 completed, roadmap milestones updated |
