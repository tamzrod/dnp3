# pkg/dnp3 - Public API for DNP3 Protocol

This package provides the official public API for the DNP3 protocol library.

## Overview

The `pkg/dnp3` package is the single entry point for external applications that want to use the DNP3 protocol. It provides high-level interfaces for both Master clients and Outstation servers.

## Package Structure

```
pkg/dnp3/
├── dnp3.go          # Main package, error types, transport types
├── master/          # Master client API
│   ├── client.go    # Client interface and implementation
│   └── client_test.go
├── outstation/      # Outstation server API
│   ├── server.go    # Server interface and implementation
│   └── server_test.go
└── types/           # Common data types
    ├── types.go     # Data point types (Binary, Analog, Counter)
    ├── commands.go  # Command types
    └── types_test.go
```

## Usage

### Master Client

```go
import (
    "context"
    "log"
    
    dnp3 "dnp3/pkg/dnp3"
    "dnp3/pkg/dnp3/master"
    "dnp3/pkg/dnp3/types"
)

// Create a Master client
client, err := master.NewClient(
    master.WithOutstationAddress(1024),
    master.WithTransport(dnp3.TCP, "192.168.1.100", 20000),
    master.WithTimeout(5 * time.Second),
)
if err != nil {
    log.Fatal(err)
}

// Connect to the outstation
ctx := context.Background()
if err := client.Connect(ctx); err != nil {
    log.Fatal(err)
}
defer client.Disconnect(ctx)

// Read data
resp, err := client.Read(ctx, &types.ReadRequest{
    Groups: types.ReadAllStatic,
})
if err != nil {
    log.Fatal(err)
}

// Process data
for _, bi := range resp.BinaryInputs {
    log.Printf("Binary Input %d: %v", bi.Index, bi.Value)
}
```

### Outstation Server

```go
import (
    "context"
    "log"
    
    "dnp3/pkg/dnp3/outstation"
    "dnp3/pkg/dnp3/types"
)

// Create a data handler
type MyDataHandler struct{}

func (m *MyDataHandler) GetBinaryInputs() []*types.BinaryInput {
    return []*types.BinaryInput{
        {Index: 0, Value: true, Quality: types.QualityOnline},
    }
}

func (m *MyDataHandler) GetAnalogInputs() []*types.AnalogInput {
    return []*types.AnalogInput{
        {Index: 0, Value: 123.45, Quality: types.QualityOnline},
    }
}

func (m *MyDataHandler) GetCounters() []*types.Counter {
    return []*types.Counter{
        {Index: 0, Value: 1000, Quality: types.QualityOnline},
    }
}

// Create an Outstation server
server, err := outstation.NewServer(
    outstation.WithAddress(1024),
    outstation.WithTransport(dnp3.TCP, "", 20000),
)
if err != nil {
    log.Fatal(err)
}

// Set the data handler
server.SetDataHandler(&MyDataHandler{})

// Start the server
ctx := context.Background()
if err := server.Start(ctx); err != nil {
    log.Fatal(err)
}
defer server.Stop(ctx)
```

## Architecture

This package provides a **public facade layer** that wraps the internal protocol implementations. The internal packages (`internal/*`) contain the detailed protocol logic and are not part of the public API.

```
External Code
      ↓
pkg/dnp3 (Public API)
      ↓
internal/* (Implementation)
      ↓
pkg/transport (Network I/O)
```

## Public API Design Principles

1. **Single Entry Point**: All external consumers import from `dnp3/pkg/dnp3`
2. **Facade Pattern**: Public types wrap internal implementations
3. **Interface-Based Consumers**: Public interfaces define contracts
4. **Builder Pattern**: Fluent configuration API

## Stability

Types in this package are part of the public API and are guaranteed to be stable across versions. Internal packages (`internal/*`) may change without notice.

## See Also

- [Master Client API](master/)
- [Outstation Server API](outstation/)
- [Common Types](types/)
- [Internal Implementation](../internal/)
