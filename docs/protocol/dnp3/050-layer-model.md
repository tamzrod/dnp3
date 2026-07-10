---
title: "050 - Layer Model"
owner: knowledge-base
---

# What is the DNP3 protocol layer model?

## Purpose

DNP3 implements a **layered protocol architecture** that separates concerns and enables modular implementation. Understanding the layer model is essential for understanding how DNP3 data flows from application to network.

## Problem Being Solved

Complex protocols need structure to be:

1. **Implementable** - Clear boundaries between functions
2. **Testable** - Each layer can be verified independently
3. **Maintainable** - Changes to one layer don't require changes to others
4. **Interoperable** - Layers are specified, not implementations

## Layer Architecture Overview

DNP3 uses a **three-layer model** (often shown as part of a four-layer TCP/IP model):

```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│                   APPLICATION LAYER                         │
│         (User Application Data & Commands)                  │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│                  TRANSPORT LAYER                            │
│           (Segmentation & Reassembly)                        │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│                  DATA LINK LAYER                            │
│             (Framing & Error Detection)                      │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│                    PHYSICAL LAYER                           │
│              (TCP/IP, Serial, etc.)                          │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Layer Responsibilities

### 1. Physical Layer (Not DNP3)

The physical layer is **not defined by DNP3**. The protocol assumes a reliable byte stream:

| Transport | Physical Layer |
|-----------|----------------|
| TCP/IP | Ethernet, fiber, etc. |
| UDP | IP networks |
| Serial | RS-232, RS-485 |

**Implication**: Network reliability is assumed. DNP3 layers handle remaining reliability.

### 2. Data Link Layer

**Primary function**: Framing and error detection

**Responsibilities**:
- Frame structure and boundaries
- CRC-16 error detection
- Address handling (source and destination)
- Link state management
- Confirmation mechanisms (FCB)

**Key concept**: The data link layer is where **DNP3 proper** begins.

See [060-link-layer.md](060-link-layer.md) for detailed link layer specification.

### 3. Transport Layer

**Primary function**: Message segmentation and reassembly

**Responsibilities**:
- Break large messages into fragments
- Reassemble received fragments
- Track first/last fragment (FIR/FIN)
- Manage transport sequence

**Key concept**: The transport layer handles messages larger than one frame.

See [070-transport-layer.md](070-transport-layer.md) for detailed transport layer specification.

### 4. Application Layer

**Primary function**: Application data representation and operations

**Responsibilities**:
- Object encoding and decoding
- Function codes (READ, WRITE, OPERATE, etc.)
- Data types and quality flags
- Events and time synchronization
- Secure authentication

**Key concept**: The application layer is where **user data** lives.

See [080-application-layer.md](080-application-layer.md) for detailed application layer specification.

## Data Flow

### Sending (Encoding)

```
┌──────────────────────────────────────────────────────────────────┐
│ APPLICATION LAYER                                                 │
│ User Data: Read Request for Group 1 Variation 1                  │
│ ┌────────────────────────────────────────────────────────────┐   │
│ │ Application PDU (Function Code + Objects)                   │   │
│ └────────────────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────────┐
│ TRANSPORT LAYER                                                   │
│ Break into fragments (if needed)                                  │
│ ┌─────────┐ ┌─────────┐ ┌─────────┐                             │
│ │ FIR/FIN │ │ FIR/FIN │ │ FIR/FIN │                             │
│ │   +     │ │   +     │ │   +     │                             │
│ │  Seq#   │ │  Seq#   │ │  Seq#   │                             │
│ └────┬────┘ └────┬────┘ └────┬────┘                             │
└──────┼───────────┼───────────┼──────────────────────────────────┘
       ▼           ▼           ▼
┌──────────────────────────────────────────────────────────────────┐
│ DATA LINK LAYER                                                   │
│ Encapsulate in frames with CRC                                    │
│ ┌─────────┐ ┌─────────┐ ┌─────────┐                             │
│ │ Header  │ │ Header  │ │ Header  │                             │
│ │   +     │ │   +     │ │   +     │                             │
│ │  Data   │ │  Data   │ │  Data   │                             │
│ │   +     │ │   +     │ │   +     │                             │
│ │  CRC    │ │  CRC    │ │  CRC    │                             │
│ └─────────┘ └─────────┘ └─────────┘                             │
└──────────────────────────────────────────────────────────────────┘
                              │
                              ▼
                         TCP/IP Stack
```

### Receiving (Decoding)

```
TCP/IP Stack
                              │
                              ▼
┌──────────────────────────────────────────────────────────────────┐
│ DATA LINK LAYER                                                   │
│ Validate CRC, extract data                                        │
│ ┌────────────────────────────────────────────────────────────┐   │
│ │ Frame Data (after stripping header/CRC)                     │   │
│ └────────────────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────────┐
│ TRANSPORT LAYER                                                   │
│ Reassemble fragments into complete message                       │
│ ┌────────────────────────────────────────────────────────────┐   │
│ │ Complete Application PDU                                    │   │
│ └────────────────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────────┐
│ APPLICATION LAYER                                                 │
│ Parse objects and execute function                               │
│ User Application receives parsed data                            │
└──────────────────────────────────────────────────────────────────┘
```

## Fragment vs Frame

### Frame (Data Link Layer)

- Unit of transmission on the data link
- Maximum size: 292 bytes (including header and CRC)
- Has its own header, data, and CRC
- Used for error detection and addressing

### Fragment (Transport Layer)

- Unit of application data
- One fragment fits within one frame
- Multiple fragments make up a complete message
- Identified by FIR (First) and FIN (Final) bits

### PDU (Application Layer)

- Protocol Data Unit
- Complete application message
- May span multiple fragments
- Contains function code and objects

## Layer Encapsulation

```
┌─────────────────────────────────────────────────────────┐
│                    APPLICATION PDU                       │
│  ┌─────────────────────────────────────────────────┐   │
│  │  Function Code │ Objects...                      │   │
│  └─────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────┐
│                 TRANSPORT FRAGMENT                       │
│  ┌──────────┬────────────────────────────────────────┐  │
│  │ FIR/FIN  │        Application PDU                 │  │
│  │ + Seq#   │                                        │  │
│  └──────────┴────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────┐
│                    DATA LINK FRAME                       │
│  ┌────┬───────┬─────┬─────┬───────────┬──────┬───────┐ │
│  │0x05│Length │Ctrl │Dst  │  Source   │ Data │ CRC   │ │
│  │ 64 │       │     │Addr │   Addr    │      │       │ │
│  └────┴───────┴─────┴─────┴───────────┴──────┴───────┘ │
└─────────────────────────────────────────────────────────┘
```

## Layer Interactions

### State Machines

Each layer maintains its own state machine:

| Layer | State Tracking |
|-------|---------------|
| Data Link | Link status, FCB state, confirmation state |
| Transport | Fragment sequence, reassembly buffer |
| Application | Transaction state, pending operations |

### Independence

Layers are designed to be **somewhat independent**:

- Data link doesn't understand application content
- Transport doesn't know what objects mean
- Application doesn't know about framing

This enables:
- Layer replacement (e.g., TCP vs serial)
- Testing in isolation
- Protocol analysis at each layer

## Unbalanced vs Balanced Mode

DNP3 supports two data link layer modes:

### Unbalanced Mode

- Master controls all communication
- Master sends function codes to primary station
- Outstations are secondary (respond only)
- Most common for TCP/IP implementations

### Balanced Mode

- Either device can initiate
- Peer-to-peer capability
- Used for some serial implementations
- More complex state management

See [060-link-layer.md](060-link-layer.md) for detailed mode specification.

## Common Misconceptions

### Misconception 1: "DNP3 has 7 layers like OSI"

**Reality**: DNP3 has 3-4 layers (physical is not DNP3). It's simpler than OSI.

### Misconception 2: "Frames and fragments are the same thing"

**Reality**: Frames are data link units with CRC. Fragments are transport units within frames.

### Misconception 3: "I only need to understand one layer"

**Reality**: All layers interact. Understanding the whole model is essential.

### Misconception 4: "Transport layer adds overhead"

**Reality**: Transport layer enables large messages. Without it, messages would be limited to frame size (292 bytes).

## Engineering Notes

### Implementation Implications

1. **Layer isolation**: Implement each layer separately
2. **Error propagation**: Errors at lower layers propagate up
3. **Testing**: Test each layer independently
4. **Performance**: Bottlenecks can occur at any layer

### Debugging Implications

When debugging:

| Problem | Likely Layer |
|---------|-------------|
| CRC errors | Data link |
| Missing fragments | Transport |
| Wrong data type | Application |
| Timeout errors | Multiple layers |

## Relationships

- **Parent**: [030-core-concepts.md](030-core-concepts.md)
- **Children**: [060-link-layer.md](060-link-layer.md), [070-transport-layer.md](070-transport-layer.md), [080-application-layer.md](080-application-layer.md)
- **Related**: Fragmentation ([200-fragmentation.md])

## References

- IEEE 1815-2012 Section 4: Protocol Structure
- IEEE 1815-2012 Section 5: Data Link Layer
- IEEE 1815-2012 Section 6: Transport Functions
