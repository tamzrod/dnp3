# Command-Line Tools

This directory contains command-line tools built with go-dnp3.

> ⚠️ **Note**: Tools will be added once implementation begins.

## Planned Tools

- [ ] `dnp3-cli` - Command-line DNP3 client
- [ ] `dnp3-server` - Simple DNP3 server
- [ ] `dnp3-proxy` - DNP3 proxy/relay
- [ ] `dnp3-sim` - DNP3 device simulator

## Building

Tools will be built using standard Go tooling:

```bash
go build ./cmd/...
```

## Installation

Tools can be installed using `go install`:

```bash
go install ./cmd/dnp3-cli
```
