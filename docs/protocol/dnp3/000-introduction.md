---
title: "000 - Introduction"
owner: knowledge-base
---

# What is DNP3?

## Purpose

DNP3 (Distributed Network Protocol 3) is a **communication protocol** designed specifically for interconnecting electrical grid equipment in SCADA (Supervisory Control and Data Acquisition) systems. It defines how data is structured, transmitted, and interpreted between master stations and remote field devices.

## Problem Being Solved

Before DNP3, utility companies faced significant challenges:

1. **Proprietary protocols** - Each vendor had unique communication methods
2. **Interoperability failure** - Equipment from different vendors couldn't communicate
3. **Limited standardization** - No common language for utility communications
4. **Evolution barriers** - Upgrading systems required complete replacement

DNP3 was designed to provide a **vendor-neutral, interoperable protocol** that could work across diverse equipment while supporting the unique requirements of electrical infrastructure.

## What DNP3 Provides

### Communication Architecture

DNP3 implements a **master-outstation model**:

```
┌─────────────────┐                    ┌─────────────────┐
│                 │                    │                 │
│   MASTER        │◄──────────────────►│   OUTSTATION    │
│   STATION       │    DNP3 Protocol   │   (RTU/IED)     │
│                 │                    │                 │
└─────────────────┘                    └─────────────────┘
```

- **Master Station**: Central control system that polls data and issues commands
- **Outstation**: Remote device that collects data and executes commands

### Key Protocol Characteristics

| Characteristic | Description |
|---------------|-------------|
| **Layered Architecture** | Data Link, Transport, and Application layers |
| **Event-Based** | Supports change-based reporting, not just polling |
| **Timestamped Data** | All events include precise timestamps |
| **Quality Flags** | Data quality is explicitly indicated |
| **Two-Way Communication** | Supports both solicited and unsolicited responses |
| **Secure Authentication** | Challenge-response security (IEEE 1815-2012) |

## Scope of DNP3

### What DNP3 Defines

- **Data Representation**: How binary states, analog values, counters, and other data types are encoded
- **Object Structure**: How data is organized into groups and variations
- **Function Codes**: Operations like READ, WRITE, OPERATE, DIRECT OPERATE
- **Communication Patterns**: Polling, unsolicited responses, confirmations
- **Timing Requirements**: Timeouts, retry behavior, response expectations
- **Security Mechanisms**: Challenge-response authentication

### What DNP3 Does NOT Define

- **Physical Transport**: DNP3 runs over TCP/IP, serial, or other transport layers
- **System Architecture**: How master stations are deployed or configured
- **Application Logic**: What data means or how it's used
- **Network Topology**: Network design and redundancy schemes
- **Device-Specific Features**: Vendor-unique extensions (while discouraged)

## Protocol Versions

| Version | Year | Key Features |
|---------|------|-------------|
| DNP3 | 1990s | Original protocol by Westronic/GE Harris |
| IEEE 1815-2010 | 2010 | First IEEE standard version |
| IEEE 1815-2012 | 2012 | Added Secure Authentication |
| IEEE 1815-2012+TC1 | 2016 | Technical Corrigendum 1 |

The current standard is **IEEE 1815-2012** with Technical Corrigendum 1.

## Typical Use Cases

### Electrical Utilities

- Substation automation
- Remote terminal unit (RTU) communication
- Protection relay data collection
- Feeder switching and control

### Water/Wastewater

- Pump station monitoring
- Tank level monitoring
- Flow measurement
- Remote valve control

### Oil and Gas

- Pipeline SCADA
- Wellhead monitoring
- Compressor station control
- Tank farm management

### Building/Industrial

- HVAC monitoring
- Energy management systems
- Fire alarm systems
- Lighting control

## How DNP3 Compares to Other Protocols

| Protocol | Domain | Comparison |
|----------|--------|------------|
| Modbus | Industrial | Simpler, less feature-rich; Modbus RTU/TCP is serial/network variant |
| IEC 60870-5-104 | SCADA | Similar domain; different object model and timing |
| IEC 61850 | Substation | More complex, object-oriented; complementary to DNP3 |
| OPC UA | Industrial | Enterprise integration; runs over DNP3 in some architectures |

## Key Takeaways

1. **DNP3 is a SCADA protocol** designed for critical infrastructure
2. **Master-outstation model** defines the communication pattern
3. **Layered architecture** provides clear separation of concerns
4. **Event-based reporting** reduces unnecessary traffic
5. **IEEE 1815 is the current standard** (2012 version with Secure Authentication)

## Relationships

- **Parent**: SCADA systems (see [030-core-concepts.md](030-core-concepts.md))
- **Related**: IEC 60870-5-104, Modbus, IEC 61850
- **Children**: Link Layer ([060-link-layer.md]), Transport Layer ([070-transport-layer.md]), Application Layer ([080-application-layer.md])

## References

- IEEE 1815-2012: Standard for Communication Performance and Timing Profiles Between Substation Computers and RTUs Using DNP3
- DNP3 Users Group: https://www.dnp.org/
