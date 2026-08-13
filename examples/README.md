# Examples

This directory contains example code for go-dnp3.

## Available Examples

### Outstation

- [x] [`outstation`](outstation/main.go) — minimal DNP3 outstation server using the v0 MVP public API. Listens on TCP, serves a single master, exposes G1V1/G20V1/G30V1 static points, and accepts Group 12 Variation 1 direct binary control. Build-only (DNP3-095). See the [outstation README](outstation/README.md).

### Master

- [x] [`master`](master/main.go) — minimal DNP3 master client using the v0 MVP public API (MEXT-031). Connects to one outstation over TCP, performs a Class-0 integrity poll (Binary Input G1V1, Counter G20V1, Analog Input G30V1), prints the points, issues one Direct-Operate CROB control (Group 12 Variation 1), and closes. Pair with the outstation example.

  ```bash
  # terminal 1
  go run ./examples/outstation
  # terminal 2
  go run ./examples/master
  ```

  Defaults match the outstation example (`localhost:20000`, outstation `1024`, master `1`). Override with flags: `-host`, `-port`, `-master`, `-outstation`, `-timeout`.

> ⚠️ **Note**: Additional examples will be added as implementation progresses.

## Planned Examples

### Basic Examples

- [x] Simple master connection (see [`master`](master/main.go))
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
