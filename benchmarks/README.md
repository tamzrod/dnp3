# Performance Benchmarks

This directory contains performance benchmarks for the go-dnp3 implementation.

## Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. -benchmem ./benchmarks/...

# Run specific benchmarks
go test -bench=BenchmarkFrameEncode -benchmem ./benchmarks/...

# Run with longer duration for more accurate results
go test -bench=. -benchmem -benchtime=5s ./benchmarks/...

# Run with CPU profiling
go test -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof ./benchmarks/...
go tool pprof cpu.prof
```

## Performance Targets

Based on [007-performance-goals.md](../docs/architecture/007-performance-goals.md):

| Operation | Target |
|-----------|--------|
| Frame encode | < 10 µs |
| Frame decode | < 15 µs |
| PDU encode | < 50 µs |
| PDU decode | < 75 µs |
| CRC-16 (292 bytes) | < 5 µs |

## Benchmark Categories

### Data Link Layer (dll_bench_test.go)

- CRC-16 calculation
- Frame encoding/decoding
- Control byte operations
- Various frame types

### Transport Layer (tl_bench_test.go)

- Fragmentation
- Reassembly
- Fragment encoding/decoding
- Header operations

### Application Layer (al_bench_test.go)

- APDU encoding/decoding
- Application control field
- IIN operations
- Function code validation

## Interpreting Results

### ns/op (nanoseconds per operation)

Lower is better. Indicates raw performance.

### B/op (bytes per operation)

Lower is better. Indicates memory allocation efficiency.

### allocs/op (allocations per operation)

Lower is better. Indicates GC pressure.

## Baseline Measurements

Run benchmarks before and after optimizations to measure improvements.
