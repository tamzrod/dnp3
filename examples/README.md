# Examples

This directory contains example code for go-dnp3.

## Available Examples

### Outstation

- [x] [`outstation`](outstation/main.go) — minimal DNP3 outstation server using the v0 MVP public API. Listens on TCP, serves a single master, exposes G1V1/G20V1/G30V1 static points, and accepts Group 12 Variation 1 direct binary control. Build-only (DNP3-095). See the [outstation README](outstation/README.md).

> ⚠️ **Note**: Additional examples will be added as implementation progresses.

## Planned Examples

### Basic Examples

- [ ] Simple master connection
- [x] Simple outstation setup (see [`outstation`](outstation/main.go))
- [ ] Reading binary inputs
- [ ] Writing binary outputs
- [ ] Reading analog values

### Advanced Examples

- [ ] Secure authentication setup
- [ ] Event handling
- [ ] Time synchronization
- [ ] File transfer
- [ ] Custom object handling

### Use Case Examples

- [ ] SCADA integration
- [ ] RTU simulation
- [ ] Master station example
- [ ] Performance testing

## Running Examples

Examples will be runnable as:

```bash
go run ./examples/...
```

## Contributing Examples

When contributing examples:

1. Follow Go idioms
2. Include documentation
3. Handle errors properly
4. Add comments for clarity
5. Test on actual DNP3 devices (when possible)
