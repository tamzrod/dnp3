# ADR-001: Package Structure

## Status

Accepted

## Context

We need to organize the go-dnp3 codebase into packages that:
1. Clearly separate concerns by protocol layer
2. Hide implementation details from users
3. Enable independent testing and evolution
4. Follow Go conventions for public vs private APIs

## Decision

We will use a layered architecture with three main directories:

```
go-dnp3/
├── internal/    # Private packages (not importable by external code)
├── pkg/        # Public packages (importable by users)
└── cmd/        # Command-line tools
```

### internal/ Structure

Internal packages follow the protocol layer hierarchy:

```
internal/
├── dll/            # Data Link Layer
│   ├── frame/     # Frame encoding/decoding
│   ├── link/      # Link state machine
│   └── crc/       # CRC-16-DNP calculations
│
├── tl/             # Transport Layer
│   ├── segment/   # Segment handling
│   └── reassemble/ # Message reassembly
│
├── al/             # Application Layer
│   ├── app/       # Application data structures
│   ├── objects/   # Object group/variation definitions
│   ├── decode/    # PDU decoding
│   └── encode/    # PDU encoding
│
├── sa/             # Secure Authentication
│   ├── challenge/ # Challenge handling
│   ├── keys/      # Key management
│   └── session/   # Session state
│
└── master/        # Master station implementation
    ├── client/    # Master client
    └── commands/  # Command handling
```

### pkg/ Structure

Public packages provide the user-facing API:

```
pkg/
├── dnp3/          # Main public API
│   ├── master/    # Master station API
│   ├── outstation/ # Outstation API
│   └── types/     # Common types
│
└── frames/        # Frame types for external use
```

### Dependency Rules

1. **No upward dependencies**: Packages cannot depend on packages "above" them
   - `internal/dll` cannot depend on `internal/tl`
   - `internal/tl` cannot depend on `internal/al`
   - `internal/*` cannot depend on `pkg/dnp3`

2. **Internal imports only**: External packages can only import `pkg/*`
   - `internal/*` packages are private by Go convention

3. **Minimal public API**: Keep `pkg/` minimal
   - Only expose necessary types
   - Use interfaces for flexibility
   - Hide implementation details

## Consequences

### Positive

- Clear separation of concerns by protocol layer
- Implementation details hidden from users
- Independent testing of each layer
- Follows Go conventions for internal packages
- Enables future changes without breaking public API

### Negative

- More directory structure to navigate
- Requires discipline to maintain dependency rules
- Internal packages cannot be imported by external test packages

### Trade-offs

We prioritize long-term maintainability and clear boundaries over short-term convenience of direct access to internal components.

## Traceability

- Architecture: [docs/architecture/004-package-architecture.md](docs/architecture/004-package-architecture.md)
- Protocol: Applies to all protocol layers (docs/protocol/dnp3/*)

## Related Decisions

- ADR-003: Concurrency Model (goroutines are internal to packages)
- ADR-002: Error Handling Strategy (errors are defined at pkg/ level)
