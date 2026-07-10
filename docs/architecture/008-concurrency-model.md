---
title: "008 - Concurrency Model"
status: approved
---

# Concurrency Model

## Overview

This document defines the concurrency model for go-dnp3. Go's concurrency 
primitives (goroutines and channels) are used idiomatically to build a 
responsive, concurrent DNP3 implementation.

## Design Principles

### 1. Go Idiom First

We use Go's concurrency patterns naturally:

- Goroutines for concurrent operations
- Channels for communication
- `sync` package for synchronization
- Context for cancellation

### 2. Encapsulated Concurrency

Concurrency is internal to components:

- Public API is safe for concurrent use
- Internal state is protected
- Goroutines are managed internally
- No channel leaking

### 3. Resource Bounds

Concurrency is bounded:

- Fixed number of goroutines
- Bounded queues
- Backpressure support
- Graceful shutdown

## Concurrency Architecture

### High-Level View

```
┌─────────────────────────────────────────────────────────────┐
│                         User Code                            │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     Public API Layer                         │
│                 (Thread-safe interfaces)                     │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Application Layer                          │
│              (Application goroutine pool)                     │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Transport Layer                           │
│                 (Segment reassembly)                         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Data Link Layer                            │
│              (Frame processing)                              │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Network I/O                            │
│              (net.Conn read/write goroutines)                │
└─────────────────────────────────────────────────────────────┘
```

### Component Goroutines

#### Connection Goroutines

Each connection manages its own goroutines:

```go
type Connection struct {
    conn   net.Conn
    readCh chan []byte   // Incoming frames
    writeCh chan []byte  // Outgoing frames
    done   chan struct{} // Shutdown signal
}
```

**Goroutine responsibilities:**

| Goroutine | Purpose |
|-----------|---------|
| Reader | Read from net.Conn, send frames to readCh |
| Writer | Receive from writeCh, write to net.Conn |
| Processor | Process frames, coordinate layers |

#### Worker Pools

For parallel processing:

```go
type WorkerPool struct {
    work chan Job
    done chan struct{}
}

func (p *WorkerPool) Run(ctx context.Context) {
    for {
        select {
        case job := <-p.work:
            job.Process()
        case <-ctx.Done():
            return
        }
    }
}
```

## Channel Design

### Channel Types

#### Unbuffered Channels

For synchronous communication:

```go
// Frame ready for processing
frameCh := make(chan *Frame)

// Guaranteed delivery
select {
case frame := <-frameCh:
    process(frame)
case <-ctx.Done():
    return
}
```

#### Buffered Channels

For bounded queues:

```go
// Command queue
commandCh := make(chan Command, 100)

// Backpressure when full
select {
case commandCh <- cmd:
    // Queued
default:
    // Queue full, apply backpressure
    return ErrQueueFull
}
```

### Channel Ownership

Clear ownership model:

```go
type Component struct {
    // Owned by Component, closed on shutdown
    internalCh chan struct{}
    
    // Received from external callers
    externalCh <-chan Request
}
```

## Synchronization

### Mutex Usage

For protecting shared state:

```go
type Session struct {
    mu      sync.Mutex
    state   SessionState
    counter uint32
}

func (s *Session) UpdateState(new State) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.state = new
}
```

### RWMutex Usage

For read-heavy workloads:

```go
type Registry struct {
    mu    sync.RWMutex
    items map[uint16]*Object
}

func (r *Registry) Get(id uint16) *Object {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.items[id]
}
```

### Atomic Operations

For simple counters and flags:

```go
type Counter struct {
    value uint64
}

func (c *Counter) Increment() uint64 {
    return atomic.AddUint64(&c.value, 1)
}

func (c *Counter) Load() uint64 {
    return atomic.LoadUint64(&c.value)
}
```

## Context Propagation

### Context Usage

Context flows through all operations:

```go
func (c *Connection) ProcessFrames(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case frame := <-c.readCh:
            if err := c.processFrame(ctx, frame); err != nil {
                return err
            }
        }
    }
}
```

### Timeout Management

Timeouts prevent hanging:

```go
// Read with timeout
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

select {
case frame := <-c.readCh:
    return frame, nil
case <-ctx.Done():
    return nil, ctx.Err()
}
```

### Cancellation

Graceful shutdown:

```go
func (c *Connection) Close() error {
    c.cancel() // Cancel context
    close(c.done) // Signal goroutines
    
    // Wait for goroutines
    c.wg.Wait()
    
    return c.conn.Close()
}
```

## Concurrency Patterns

### Pipeline Pattern

For data flow:

```
Source → Transform → Filter → Sink
   ↓         ↓          ↓       ↓
 Ch1       Ch2         Ch3     Ch4
```

```go
func Pipeline(ctx context.Context, data []byte) error {
    frames := generateFrames(ctx, data)
    decoded := decodeFrames(ctx, frames)
    processed := processFrames(ctx, decoded)
    
    return sendResults(ctx, processed)
}
```

### Fan-Out Pattern

For parallel processing:

```go
func ProcessConcurrently(ctx context.Context, items []Item) []Result {
    results := make([]Result, len(items))
    var wg sync.WaitGroup
    
    for i, item := range items {
        wg.Add(1)
        go func(i int, item Item) {
            defer wg.Done()
            results[i] = process(item)
        }(i, item)
    }
    
    wg.Wait()
    return results
}
```

### Worker Pool Pattern

For bounded concurrency:

```go
func (p *Pool) Submit(ctx context.Context, job Job) error {
    select {
    case p.jobs <- job:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    default:
        return ErrPoolFull
    }
}
```

## Connection Pooling

### Pool Design

For managing multiple connections:

```go
type Pool struct {
    mu       sync.Mutex
    conns    []*Connection
    avail    []*Connection
    factory  func() (*Connection, error)
    maxOpen  int
    maxIdle  int
}
```

### Pool Operations

```go
func (p *Pool) Get(ctx context.Context) (*Connection, error) {
    p.mu.Lock()
    
    if len(p.avail) > 0 {
        conn := p.avail[len(p.avail)-1]
        p.avail = p.avail[:len(p.avail)-1]
        p.mu.Unlock()
        return conn, nil
    }
    
    if p.len() >= p.maxOpen {
        // Wait or error
    }
    
    p.mu.Unlock()
    return p.factory()
}
```

## Error Handling

### Goroutine-Safe Errors

Errors are captured and propagated:

```go
func (c *Connection) readLoop(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            c.handleError(ctx.Err())
            return
        default:
            // Read and handle errors
        }
    }
}

func (c *Connection) handleError(err error) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    if c.onError != nil {
        c.onError(err)
    }
}
```

### Panic Recovery

Prevent goroutine leaks:

```go
func safeGo(fn func()) {
    go func() {
        defer func() {
            if r := recover(); r != nil {
                log.Printf("Recovered: %v", r)
            }
        }()
        fn()
    }()
}
```

## Testing Concurrency

### Race Detector

Always run with race detector:

```bash
go test -race ./...
```

### Concurrent Tests

```go
func TestConnection_ConcurrentReads(t *testing.T) {
    c, _ := NewTestConnection()
    
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            c.Read(context.Background())
        }()
    }
    
    wg.Wait()
}
```

### Stress Tests

```go
func TestConnection_Stress(t *testing.T) {
    const (
        goroutines = 100
        iterations  = 1000
    )
    
    // Stress test implementation
}
```

## Metrics

### Concurrency Metrics

Track concurrency metrics:

```go
type Metrics struct {
    ActiveConnections  atomic.Int64
    ActiveGoroutines   atomic.Int64
    ChannelCapacity    atomic.Int64
    QueueDepth         atomic.Int64
}
```

### Monitoring

Expose metrics for monitoring:

- Active connections
- Goroutine count
- Queue depths
- Blocking operations
