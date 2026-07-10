---
title: "004 - Package Architecture"
status: draft
---

# Package Architecture

## Overview

This document defines the package structure for go-dnp3. The architecture 
is designed to provide clear separation of concerns while enabling 
flexibility and testability.

## Design Goals

### Separation of Concerns

Each package has a distinct, focused purpose:

- Clear boundaries
- Minimal coupling
- Maximum cohesion
- Independent evolution

### Dependency Direction

Dependencies flow downward:

```
User Code
    ↓
Application Layer
    ↓
Transport Layer
    ↓
Data Link Layer
    ↓
Network I/O
```

### Implementation Hiding

Internal details are hidden:

- `internal/` for private packages
- `pkg/` for public packages
- Clear public APIs
- Minimal exported types

## Proposed Package Structure

```
go-dnp3/
├── cmd/                  # Command-line tools
│   └── example/         # Example applications
│
├── internal/            # Private packages (not importable)
│   ├── dll/            # Data Link Layer implementation
│   │   ├── frame/      # Frame encoding/decoding
│   │   ├── link/       # Link state machine
│   │   └── crc/        # CRC calculations
│   │
│   ├── tl/             # Transport Layer implementation
│   │   ├── segment/    # Segment handling
│   │   └── reassemble/ # Message reassembly
│   │
│   ├── al/             # Application Layer implementation
│   │   ├── app/        # Application data
│   │   ├── objects/    # Object definitions
│   │   ├── decode/     # PDU decoding
│   │   └── encode/     # PDU encoding
│   │
│   ├── sa/             # Secure Authentication
│   │   ├── challenge/  # Challenge handling
│   │   ├── keys/       # Key management
│   │   └── session/    # Session security
│   │
│   └── master/         # Master station implementation
│   ├── client/         # Master client
│   └── commands/       # Command handling
│
├── pkg/                 # Public packages (importable)
│   ├── dnp3/           # Public API
│   │   ├── master/     # Master API
│   │   ├── outstation/ # Outstation API
│   │   └── types/      # Common types
│   │
│   └── frames/         # Frame types for external use
│
├── test/               # Test utilities and data
│   ├── conformance/   # Conformance test data
│   ├── interop/       # Interoperability test data
│   └── fuzz/          # Fuzzing tests
│
└── docs/              # Documentation
    ├── architecture/  # Architecture documents
    └── api/           # API documentation
```

## Package Responsibilities

### `pkg/dnp3` - Public API

The main public interface for users.

**Responsibilities:**
- Provide high-level API
- Handle connection management
- Expose configuration options
- Return user-friendly errors

**Public API Design:**
- Simple, intuitive interface
- Builder pattern for configuration
- Context-aware operations
- Callback-based event handling

### `internal/dll` - Data Link Layer

Implements IEEE 1815 data link layer.

**Responsibilities:**
- Frame encoding/decoding
- CRC-16 calculation and validation
- Link state machine (primary/secondary)
- Address handling
- Function codes

**Key Types:**
- Link Layer Configuration
- Link State Machine
- Frame Buffer

### `internal/tl` - Transport Layer

Implements IEEE 1815 transport layer.

**Responsibilities:**
- Segmentation
- Message reassembly
- Transport header handling
- Flow control

**Key Types:**
- Transport Configuration
- Segment Handler
- Reassembly Buffer

### `internal/al` - Application Layer

Implements IEEE 1815 application layer.

**Responsibilities:**
- Application data encoding/decoding
- Object group and variation handling
- Function code processing
- Response generation

**Key Types:**
- Application Configuration
- PDU Encoder/Decoder
- Object Database

### `internal/sa` - Secure Authentication

Implements IEEE 1815-2012 Secure Authentication.

**Responsibilities:**
- Challenge generation
- Challenge response validation
- Key management
- Session state tracking

**Key Types:**
- Authentication Configuration
- Session State
- Key Table

## Layer Interactions

### Data Flow

```
User Data
    ↓
Application Layer (encoding)
    ↓
Transport Layer (segmentation)
    ↓
Data Link Layer (framing)
    ↓
Network Socket
```

### Control Flow

```
Network Events
    ↓
Data Link Layer (validation)
    ↓
Transport Layer (reassembly)
    ↓
Application Layer (parsing)
    ↓
User Callbacks
```

## Dependency Rules

### Rule 1: No Upward Dependencies

Packages cannot depend on packages above them:

- `internal/dll` cannot depend on `internal/tl`
- `internal/tl` cannot depend on `internal/al`
- `internal/al` cannot depend on `pkg/dnp3`

### Rule 2: Internal Imports Only

External packages can only import `pkg/*`.

`internal/*` packages are private.

### Rule 3: Minimal Public API

Keep `pkg/` minimal:

- Only expose necessary types
- Use interfaces for flexibility
- Hide implementation details

## Testing Strategy

### Unit Tests

Each package has unit tests:

```
internal/dll/dll_test.go
internal/tl/tl_test.go
internal/al/al_test.go
```

### Integration Tests

Cross-package tests:

```
test/integration/
├── master_test.go
└── outstation_test.go
```

### Conformance Tests

Protocol compliance tests:

```
test/conformance/
├── dll_test.go
├── tl_test.go
└── al_test.go
```

## Future Considerations

### Serial Transport

Future serial support would add:

```
internal/serial/
├── port/      # Serial port abstraction
├── dll/       # Serial-specific data link
└── framer/    # Serial framing
```

### Additional Protocols

Supporting other protocols would add:

```
pkg/modbus/
pkg/iec104/
```

But these are explicitly out of scope for now.
