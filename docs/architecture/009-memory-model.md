---
title: "009 - Memory Model"
status: draft
---

# Memory Model

## Overview

This document defines the memory model for go-dnp3. We design for 
efficient memory usage while ensuring correctness and safety.

## Design Principles

### 1. Predictable Allocation

Memory allocation patterns should be:

- Predictable
- Bounded
- Efficient
- No unbounded growth

### 2. Zero-Copy When Possible

Avoid unnecessary copying:

- Parse in place
- Use slices as views
- Return references when safe

### 3. Pool Everything

Reuse allocations:

- Buffer pools
- Object pools
- Connection pools

### 4. Bounded Resources

All resources have limits:

- Connection limits
- Queue depths
- Buffer sizes
- Session timeouts

## Memory Layout

### Frame Structure

```
┌─────────────────────────────────────────────────────────────────┐
│ Data Link Frame (292 bytes max)                                  │
├────────┬────────┬────────┬────────┬──────────┬────────┬────────┤
│ 0x0564 │ Length │ Ctrl   │ Dest   │  Source  │  Data  │  CRC   │
│ (2 B)  │ (1 B)  │ (1 B)  │ (2 B)  │  (2 B)   │ (0-292)│ (2 B)  │
└────────┴────────┴────────┴────────┴──────────┴────────┴────────┘
         ↑
         └── Variable length
```

### PDU Structure

```
┌─────────────────────────────────────────────────────────────────┐
│ Application PDU                                                  │
├────────┬────────┬────────┬─────────────────────────────────────┤
│ Ctrl   │  FC    │  IIN   │           Objects                   │
│ (1 B)  │ (1 B)  │ (2 B)  │         (variable)                  │
└────────┴────────┴────────┴─────────────────────────────────────┘
```

## Buffer Management

### Buffer Pool

Reuse buffers to reduce GC pressure:

```go
type BufferPool struct {
    small sync.Pool // < 1KB
    medium sync.Pool // 1KB - 10KB
    large sync.Pool // > 10KB
}

func (p *BufferPool) Get(size int) []byte {
    switch {
    case size < 1024:
        b := p.small.Get()
        return b.([]byte)[:size]
    case size < 10240:
        b := p.medium.Get()
        return b.([]byte)[:size]
    default:
        return make([]byte, size)
    }
}
```

### Pre-allocated Buffers

For fixed-size operations:

```go
// Frame buffers are always the same size
var frameBuffer = make([]byte, 292)

// Encode into pre-allocated buffer
func EncodeFrame(f *Frame, buf []byte) (int, error) {
    // Use buf directly
    // ...
}
```

### Bounded Queues

Prevent memory exhaustion:

```go
// Bounded command queue
type CommandQueue struct {
    ch chan Command
}

func NewCommandQueue(capacity int) *CommandQueue {
    return &CommandQueue{
        ch: make(chan Command, capacity), // Bounded!
    }
}
```

## Slice Usage

### Views vs Copies

Prefer views over copies:

```go
// BAD: Unnecessary copy
func (f *Frame) GetData() []byte {
    result := make([]byte, len(f.data))
    copy(result, f.data)
    return result
}

// GOOD: Return view
func (f *Frame) GetData() []byte {
    return f.data // Return view
}
```

### Slice Tricks

```go
// Pre-allocate with capacity
buf := make([]byte, 0, expectedSize)

// Trim to actual size
buf = buf[:actualSize]

// Reuse slice header
buf = buf[:0] // Reset length, keep capacity
```

## String Handling

### Avoid Unnecessary Strings

Protocol data is often binary:

```go
// BAD: Convert to string
data := string(bytes)

// GOOD: Keep as bytes
data := bytes
```

### When to Use Strings

Only for user-facing data:

```go
// OK: Error messages
return fmt.Errorf("invalid frame: %v", err)

// OK: Logging
log.Printf("Processing frame: %x", data)

// Avoid: Internal protocol data
```

## Object Pools

### Object Reuse

For frequently allocated objects:

```go
type EventPool struct {
    pool sync.Pool
}

func NewEventPool() *EventPool {
    return &EventPool{
        pool: sync.Pool{
            New: func() interface{} {
                return &Event{}
            },
        },
    }
}

func (p *EventPool) Get() *Event {
    return p.pool.Get().(*Event)
}

func (p *EventPool) Put(e *Event) {
    e.Reset() // Clear for reuse
    p.pool.Put(e)
}
```

### Reset Methods

All pooled objects have reset:

```go
type Event struct {
    Group    uint8
    Variation uint8
    Value    interface{}
    Timestamp time.Time
}

func (e *Event) Reset() {
    e.Group = 0
    e.Variation = 0
    e.Value = nil
    e.Timestamp = time.Time{}
}
```

## Memory Limits

### Connection Limits

Each connection uses bounded memory:

```
Memory per connection ≈
    Frame buffer (292 bytes)
  + PDU buffer (up to 2048 bytes)
  + State overhead (~1 KB)
  + Event queue (bounded)
  = ~5-10 KB per connection
```

### Global Limits

System-wide limits:

```go
type Limits struct {
    MaxConnections    int
    MaxBufferPoolSize int
    MaxEventQueueSize int
    MaxSessionTimeout time.Duration
}
```

### Resource Limits

Built-in limits:

```go
const (
    MaxFrameSize = 292
    MaxPDUSize = 2048
    MaxQueueDepth = 1000
    MaxConnections = 10000
)
```

## Garbage Collection

### GC-Friendly Design

Minimize GC pressure:

- Object pooling
- Pre-allocation
- Stack allocation where possible
- Avoid finalizers

### GC Metrics

Monitor GC behavior:

```go
import "runtime/debug"

func PrintGCStats() {
    stats := &debug.GCStats{}
    debug.ReadGCStats(stats)
    
    fmt.Printf("GC count: %d\n", stats.NumGC)
    fmt.Printf("Pause total: %v\n", stats.PauseTotal)
    fmt.Printf("Pause avg: %v\n", stats.PauseTotal/time.Duration(stats.NumGC))
}
```

## Profiling Memory

### Memory Profiles

```bash
# Generate memory profile
go test -memprofile=mem.prof -memprofilerate=1 ./...

# View profile
go tool pprof mem.prof
```

### Allocations

```go
// Track allocations in tests
func TestMemory(t *testing.T) {
    testing.AllocsPerRun(1000, func() {
        // Code to test
    })
}
```

### Benchmarking

```go
func BenchmarkEncode(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        Encode(...)
    }
}
```

## Common Patterns

### Message Framing

```go
type Framer struct {
    buf     []byte
    pos     int
    maxSize int
}

func (f *Framer) Reset() {
    f.buf = f.buf[:0]
    f.pos = 0
}

func (f *Framer) AddByte(b byte) error {
    if len(f.buf) >= f.maxSize {
        return ErrFrameTooLarge
    }
    f.buf = append(f.buf, b)
    return nil
}
```

### Streaming Parser

```go
type Parser struct {
    buf   []byte
    ready chan []byte // Completed frames
}

func (p *Parser) Parse(data []byte) {
    p.buf = append(p.buf, data...)
    
    for {
        frame, err := p.findFrame(p.buf)
        if err != nil {
            break
        }
        p.ready <- frame
        p.buf = p.buf[len(frame):]
    }
}
```

### Ring Buffer

For fixed-size buffers:

```go
type RingBuffer struct {
    data []byte
    head int
    tail int
    size int
}

func (rb *RingBuffer) Write(b []byte) (int, error) {
    for i, v := range b {
        rb.data[rb.tail] = v
        rb.tail = (rb.tail + 1) % rb.size
        if rb.tail == rb.head {
            return i, ErrBufferFull
        }
    }
    return len(b), nil
}
```

## Memory Safety

### Bounds Checking

Always check bounds:

```go
func (f *Frame) GetByte(pos int) (byte, error) {
    if pos < 0 || pos >= len(f.data) {
        return 0, ErrOutOfBounds
    }
    return f.data[pos], nil
}
```

### Slice Safety

Never return escaped slices:

```go
// SAFE: Copy on return
func (f *Frame) GetDataCopy() []byte {
    result := make([]byte, len(f.data))
    copy(result, f.data)
    return result
}

// CAREFUL: View returns
func (f *Frame) GetDataView() []byte {
    return f.data // Caller must not modify
}
```

### No Global State

Avoid mutable globals:

```go
// BAD
var globalBuffer []byte

// GOOD
type Connection struct {
    buffer []byte
}
```

## Documentation

Memory characteristics are documented:

- Buffer sizes
- Pool limits
- Allocation patterns
- GC implications
