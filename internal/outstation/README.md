# Outstation Package

**Package**: `internal/outstation`
**Status**: Production

This package implements the DNP3 Outstation role for the go-dnp3 library.

## Overview

The Outstation (also called Remote Terminal Unit or RTU) is a device that:
- Responds to requests from the Master
- Provides data (binary inputs, analog inputs, counters)
- Executes control commands
- Reports events via unsolicited responses

## Key Components

- **State Machine**: Manages outstation operational states
- **Data Handler**: Provides access to data points
- **Request Processor**: Handles incoming Master requests
- **Response Builder**: Constructs protocol responses

## Usage

```go
// Create outstation
config := outstation.DefaultConfig()
ost := outstation.NewOutstation(config)

// Set data provider
ost.SetDataHandler(myDataProvider)

// Set transport
ost.SetTransport(myTransport)

// Initialize and start
ost.Initialize()
ost.Start()
```

## Reference

IEEE 1815-2012 Section 8
