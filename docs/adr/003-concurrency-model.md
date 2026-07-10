# ADR-003: Concurrency Model

## Status

Accepted

## Context

We need a concurrency model that:
1. Follows Go idioms naturally
2. Encapsulates goroutines within packages
3. Provides thread-safe public APIs
4. Supports graceful shutdown
5. Enables proper resource cleanup

## Decision

We will use the following concurrency patterns:

### Encapsulated Goroutines

Goroutines are internal to packages. The public API is safe for concurrent use.

```go
// internal/dll/link/link.go
type LinkStateMachine struct {
    conn   net.Conn
    done   chan struct{}  // Internal shutdown signal
    
    // Channels for communication
    sendCh chan []byte
    recvCh chan []byte
    
    mu     sync.Mutex    // Protects internal state
    state  State
}
```

### Channel-Based Communication

Channels are used for:
1. Frame communication between layers
2. Command injection from public API
3. Event notification to callbacks

```go
// Channels are unbuffered for critical paths
frameCh := make(chan *Frame)

// Buffered channels for queues with backpressure
commandCh := make(chan Command, 100)

// Use select with context for cancellation
select {
case frameCh <- frame:
    // Delivered
case <-ctx.Done():
    return ctx.Err()
}
```

### Worker Pools for Parallel Processing

For CPU-bound work (e.g., CRC calculation):

```go
type CRCWorkerPool struct {
    workers int
    jobs    chan []byte
    results chan uint16
    done    chan struct{}
}

func NewCRCWorkerPool(workers int) *CRCWorkerPool {
    p := &CRCWorkerPool{
        workers: workers,
        jobs:    make(chan []byte, workers*2),
        results: make(chan uint16, workers*2),
        done:    make(chan struct{}),
    }
    for i := 0; i < workers; i++ {
        go p.worker()
    }
    return p
}
```

### Context for Cancellation

Context flows through all operations:

```go
func (c *Connection) ProcessFrames(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case frame := <-c.recvCh:
            if err := c.processFrame(frame); err != nil {
                return err
            }
        }
    }
}
```

### Graceful Shutdown

Each component implements graceful shutdown:

```go
func (c *Connection) Close() error {
    // Signal goroutines to stop
    close(c.done)
    
    // Wait for goroutines to finish
    c.wg.Wait()
    
    // Close underlying connection
    return c.conn.Close()
}
```

### Mutex Usage Guidelines

| Situation | Mutex Type | Rationale |
|-----------|------------|-----------|
| Simple counter | atomic operations | Avoid mutex for simple cases |
| Read-heavy data | sync.RWMutex | Allow concurrent readers |
| Complex state | sync.Mutex | Protect compound operations |
| One-shot operations | sync.Once | Ensure single initialization |

```go
// Example: RWMutex for registry
type ObjectRegistry struct {
    mu    sync.RWMutex
    items map[uint16]*Object
}

func (r *ObjectRegistry) Get(id uint16) *Object {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.items[id]
}

func (r *ObjectRegistry) Set(id uint16, obj *Object) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.items[id] = obj
}
```

### Channel Ownership

Clear ownership model:

```go
type Component struct {
    // Owned by Component, closed on shutdown
    internalCh chan struct{}
    
    // Received from external callers (read-only)
    externalCh <-chan Request
    
    // Sent to external callers (write-only)
    resultCh chan<- Result
}
```

## Consequences

### Positive

- Idiomatic Go concurrency patterns
- Thread-safe public API
- Clear goroutine lifecycle management
- Graceful shutdown support
- Context-based cancellation throughout

### Negative

- More complex code structure
- Need to manage goroutine lifecycle
- Channel leaks if not careful
- Harder to debug than sequential code

### Trade-offs

We prioritize correctness and idiomatic Go over simpler but less concurrent code.

## Traceability

- Architecture: [docs/architecture/008-concurrency-model.md](docs/architecture/008-concurrency-model.md)
- Protocol: Concurrency model for SCADA real-time requirements

## Related Decisions

- ADR-001: Package Structure (goroutines are internal)
- ADR-002: Error Handling Strategy (errors passed through channels)
- ADR-004: Memory Model (buffer management with concurrency)
