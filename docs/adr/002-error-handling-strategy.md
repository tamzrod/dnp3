# ADR-002: Error Handling Strategy

## Status

Accepted

## Context

We need a consistent error handling strategy that:
1. Provides clear, actionable error messages
2. Enables programmatic error handling by users
3. Supports error wrapping and context preservation
4. Distinguishes between transient and permanent errors
5. Enables proper error logging and observability

## Decision

We will use Go's error handling idioms with the following conventions:

### Error Definition

Errors are defined in `pkg/dnp3/` with exported error variables:

```go
package dnp3

// Sentinel errors for the public API
var (
    ErrConnectionClosed = errors.New("dnp3: connection closed")
    ErrTimeout         = errors.New("dnp3: operation timed out")
    ErrBufferFull     = errors.New("dnp3: buffer full")
    // ... more sentinel errors
)

// Wrapped errors use fmt.Errorf with %w
return fmt.Errorf("dnp3: read frame: %w", err)
```

### Error Categories

Errors are categorized by their nature:

| Category | Suffix | Example | Behavior |
|----------|--------|---------|----------|
| Transient | (contextual) | ErrTimeout | Retry may succeed |
| Permanent | (contextual) | ErrNotSupported | Retry will fail |
| Protocol | (contextual) | ErrInvalidCRC | Indicates protocol violation |
| Config | (contextual) | ErrInvalidAddress | Indicates misconfiguration |

### Error Type Hierarchy

```go
// Base error interface
type Error interface {
    Error() string
    Unwrap() error
    Code() ErrorCode  // For programmatic handling
}

// ErrorCode for programmatic error handling
type ErrorCode int

const (
    CodeConnection ErrorCode = iota
    CodeTimeout
    CodeProtocol
    CodeValidation
    // ... more codes
)

// ValidationError for input validation failures
type ValidationError struct {
    Field   string
    Value   interface{}
    Reason  string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("dnp3: validation failed for %s: %s", e.Field, e.Reason)
}

// ProtocolError for protocol-level errors
type ProtocolError struct {
    Layer   string  // "dll", "tl", "al"
    Code    uint8   // Protocol-specific error code
    Details string
}
```

### Layer-Specific Errors

Each layer (`internal/dll`, `internal/tl`, `internal/al`) defines its own errors using the same pattern:

```go
// internal/dll/errors.go
package dll

var (
    ErrInvalidCRC       = errors.New("dll: invalid CRC")
    ErrInvalidFrame     = errors.New("dll: invalid frame")
    ErrUnexpectedFunc   = errors.New("dll: unexpected function code")
)

// These are wrapped when returning to upper layers
return nil, fmt.Errorf("dll: decode: %w", ErrInvalidCRC)
```

### Context Preservation

Use `fmt.Errorf` with `%w` to preserve error chain:

```go
// Good: preserves full context
return fmt.Errorf("tl: reassemble: %w", err)

// Avoid: loses context
return fmt.Errorf("tl: %s", err.Error())
```

### No Panics in Production Code

- Panics are prohibited in production code paths
- Use `recover()` only at goroutine boundaries for safety
- Convert panics to errors in goroutine entry points:

```go
func (p *Processor) run() {
    defer func() {
        if r := recover(); r != nil {
            p.handlePanic(r)
        }
    }()
    // ... normal processing
}
```

## Consequences

### Positive

- Clear error hierarchy for users
- Programmatic error handling via ErrorCode
- Consistent error messages across layers
- Error chain preservation for debugging
- No surprising panics

### Negative

- More error types to maintain
- Error wrapping can make debugging harder if not done carefully
- Need to document all error codes

### Trade-offs

We prioritize user experience and debuggability over minimal error types.

## Traceability

- Architecture: [docs/architecture/004-package-architecture.md](docs/architecture/004-package-architecture.md)
- Protocol: Error handling is per IEEE 1815 specification

## Related Decisions

- ADR-001: Package Structure (errors defined at pkg/ level)
- ADR-003: Concurrency Model (errors passed through channels)
