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
│ Phase 4: Implementation                                      │
└─────────────────────────────────────────────────────────────┘
                              ↑
┌─────────────────────────────────────────────────────────────┐
│ Phase 3: Architecture Review                                 │
└─────────────────────────────────────────────────────────────┘
                              ↑
┌─────────────────────────────────────────────────────────────┐
│ Phase 2: Architecture                                        │
└─────────────────────────────────────────────────────────────┘
                              ↑
┌─────────────────────────────────────────────────────────────┐
│ Phase 1: Research 🔄 CURRENT                                 │
└─────────────────────────────────────────────────────────────┘
```

## Phase 1: Research

**Status**: 🔄 In Progress

### Goals

- Complete protocol analysis
- Identify all edge cases
- Document invariants
- Research implementations

### Deliverables

- [ ] IEEE 1815-2012 complete analysis
- [ ] Data link layer invariants documented
- [ ] Transport layer invariants documented
- [ ] Application layer invariants documented
- [ ] Secure authentication analysis
- [ ] Edge case inventory
- [ ] Existing implementation analysis

### Timeline

This phase is ongoing throughout the project.

### Exit Criteria

- All protocol features analyzed
- Edge cases documented
- Implementation approach defined
- Research reviewed by team

## Phase 2: Architecture

**Status**: 📋 Planned

### Goals

- Define package structure
- Design interfaces
- Document decisions
- Create ADRs

### Deliverables

- [ ] Package architecture document
- [ ] Data link layer interface design
- [ ] Transport layer interface design
- [ ] Application layer interface design
- [ ] Security architecture
- [ ] Error handling design
- [ ] Configuration design
- [ ] Key ADRs created

### Timeline

TBD based on research completion.

### Exit Criteria

- Architecture documents complete
- Key decisions documented in ADRs
- Interfaces reviewed
- Dependencies mapped

## Phase 3: Architecture Review

**Status**: 📋 Planned

### Goals

- Review architecture
- Address feedback
- Get approval

### Deliverables

- [ ] Team review
- [ ] External review (if available)
- [ ] Feedback incorporated
- [ ] Sign-off from maintainers

### Timeline

TBD based on architecture completion.

### Exit Criteria

- All reviews complete
- No blocking concerns
- Architecture approved
- Ready for implementation

## Phase 4: Implementation

**Status**: 📋 Planned

### Goals

- Implement data link layer
- Implement transport layer
- Implement application layer
- Implement secure authentication

### Milestones

#### 4.1 Data Link Layer

- [ ] Frame encoding/decoding
- [ ] CRC calculation
- [ ] Link state machine
- [ ] Address handling
- [ ] Unbalanced mode
- [ ] Balanced mode

#### 4.2 Transport Layer

- [ ] Segmentation
- [ ] Reassembly
- [ ] Header handling
- [ ] Flow control

#### 4.3 Application Layer

- [ ] PDU encoding/decoding
- [ ] All function codes
- [ ] All object groups
- [ ] All variations
- [ ] Response generation

#### 4.4 Secure Authentication

- [ ] Challenge handling
- [ ] Key management
- [ ] Session security
- [ ] Cryptographic operations

### Timeline

TBD based on architecture approval.

### Exit Criteria

- All layers implemented
- Unit tests > 80% coverage
- Integration tests passing
- No known bugs

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
