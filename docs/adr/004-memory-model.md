# ADR-004: Memory Model

## Status

Accepted

## Context

We need a memory model that:
1. Minimizes allocations in hot paths
2. Supports bounded buffers for flow control
3. Enables zero-copy parsing where possible
4. Provides predictable memory usage
5. Prevents memory leaks in long-running SCADA systems

## Decision

We will use the following memory management strategies:

### Buffer Pools

Use `sync.Pool` for frequently allocated buffers:

```go
// Buffer pool for frame buffers
var frameBufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, maxFrameSize)
    },
}

func GetFrameBuffer() []byte {
    return frameBufferPool.Get().([]byte)
}

func PutFrameBuffer(buf []byte) {
    // Reset slice length but keep capacity
    frameBufferPool.Put(buf[:maxFrameSize])
}
```

### Bounded Channels for Backpressure

Channels are used to apply backpressure:

```go
// Bounded queue for outgoing frames
writeCh := make(chan []byte, 100)

// Non-blocking send with backpressure
select {
case writeCh <- buf:
    // Queued successfully
default:
    return ErrBufferFull  // Apply backpressure
}
```

### Slice Reuse for Parsing

Reuse slices for parsing to avoid allocations:

```go
type Parser struct {
    // Reusable buffer for parsing
    buf []byte
}

func (p *Parser) parseHeader(data []byte) (*Header, error) {
    // Parse into existing buffer
    if cap(p.buf) < len(data) {
        p.buf = make([]byte, len(data))
    }
    p.buf = p.buf[:len(data)]
    copy(p.buf, data)
    
    // Parse from p.buf
    return p.parseFromBuffer()
}
```

### Fixed-Size Object Pools

For protocol objects that are frequently created/destroyed:

```go
type Event struct {
    mu       sync.Mutex
    freeList []*Event
    maxSize  int
}

func (e *EventPool) Get() *Event {
    e.mu.Lock()
    defer e.mu.Unlock()
    
    if len(e.freeList) > 0 {
        ev := e.freeList[len(e.freeList)-1]
        e.freeList = e.freeList[:len(e.freeList)-1]
        return ev
    }
    return &Event{}
}

func (e *EventPool) Put(ev *Event) {
    e.mu.Lock()
    defer e.mu.Unlock()
    
    if len(e.freeList) < e.maxSize {
        ev.reset()
        e.freeList = append(e.freeList, ev)
    }
}
```

### Escape Analysis Guidance

Be conscious of escape analysis:

```go
// AVOID: Returning slice from pool
func GetBuffer() []byte {
    return make([]byte, 1024)  // Escapes to heap
}

// PREFER: Passing buffer to function
func ProcessWithBuffer(buf []byte) {
    // buf stays on stack if caller owns it
}

// FOR PUBLIC API: Return owned buffer
func PublicAPIFunction() ([]byte, error) {
    buf := make([]byte, 1024)  // Explicit heap allocation
    // ... fill buffer
    return buf, nil  // Caller receives ownership
}
```

### Memory Limits

Configuration options for memory bounds:

```go
type Config struct {
    // Maximum frame size (affects buffer allocation)
    MaxFrameSize int
    
    // Maximum event buffer size
    MaxEventBufferSize int
    
    // Maximum command queue depth
    MaxCommandQueueDepth int
    
    // Worker pool sizes
    CRCWorkerCount int
}
```

## Consequences

### Positive

- Reduced allocations in hot paths
- Predictable memory usage
- Bounded buffers prevent unbounded growth
- Object pools reduce GC pressure
- Clear ownership semantics

### Negative

- More complex buffer management
- Risk of holding too many objects in pools
- Need to carefully track buffer ownership
- Pool tuning may be needed

### Trade-offs

We prioritize predictable memory usage and low GC pressure over simpler but potentially allocating code.

## Traceability

- Architecture: [docs/architecture/009-memory-model.md](docs/architecture/009-memory-model.md)
- Protocol: SCADA systems require bounded resource usage

## Related Decisions

- ADR-001: Package Structure (buffer pools are internal)
- ADR-003: Concurrency Model (pools are thread-safe)
