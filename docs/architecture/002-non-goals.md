---
title: "002 - Non-Goals"
status: draft
---

# Non-Goals

## Overview

This document defines what go-dnp3 explicitly does **not** aim to do.
Clear non-goals help maintain focus and prevent scope creep.

## Out of Scope

### Protocol Wrappers

We are not:

- A wrapper around opendnp3
- A wrapper around lib60870
- A wrapper around any other DNP3 implementation
- An FFI binding to C libraries

**Rationale**: Wrappers inherit limitations and don't follow Go idioms.

### Legacy Protocol Support

We do not implement:

- DNP3 versions prior to IEEE 1815-2012
- Vendor-specific extensions not in the standard
- Deprecated features still in the spec

**Rationale**: Focus on the current standard ensures clarity and correctness.

### Serial Transport

Initially, we do not support:

- DNP3 over serial (RS-232)
- DNP3 over serial (RS-485)
- EIA-232/EIA-485 specific handling

**Rationale**: Modern deployments use TCP/IP. Serial support may be added later.

### Other SCADA Protocols

We do not implement:

- Modbus
- IEC 61850
- IEC 60870-5-104
- OPC UA
- Any other protocol

**Rationale**: Scope must be bounded. Focus on DNP3 excellence.

## Explicitly Not Features

### Compatibility Layers

We do not provide:

- Drop-in replacement APIs for other libraries
- Legacy API support from day one
- Migration tooling from other implementations

**Rationale**: Idiomatic Go design takes priority over compatibility.

### GUI Components

We do not provide:

- Graphical user interfaces
- Web dashboards
- Configuration GUIs
- Visualization tools

**Rationale**: Protocol library only. GUIs are separate projects.

### Enterprise Features

We do not provide:

- User management
- Role-based access control beyond protocol
- Audit logging (beyond protocol requirements)
- Configuration management systems

**Rationale**: Focus on protocol correctness, not enterprise features.

### Cloud Services

We do not provide:

- Cloud hosting
- Protocol gateways
- SaaS offerings
- Managed services

**Rationale**: Library only. Services are separate projects.

## What We Won't Optimize For

### Short-Term Development Speed

We won't:

- Rush implementation for early deadlines
- Skip architecture review
- Skip security review
- Skip testing

**Rationale**: Correctness is paramount for critical infrastructure.

### Maximum Features

We won't:

- Add features not in the spec
- Support vendor extensions by default
- Include convenience methods that obscure behavior

**Rationale**: Minimal, correct implementation.

### Maximum Performance Over Correctness

We won't:

- Skip validation for speed
- Use unsafe operations without justification
- Optimize before correctness is verified

**Rationale**: Correctness > Performance. Always.

## Why These Non-Goals Matter

Clear non-goals:

- Prevent scope creep
- Enable focused development
- Maintain project identity
- Ensure timely delivery

## Future Considerations

Some non-goals may become goals in the future:

- Serial transport support
- Windows compatibility
- Embedded systems support
- Hardware acceleration

These require additional research and architecture.
