---
title: "007 - Performance Goals"
status: approved
---

# Testing Strategy

## Overview

This document defines the testing strategy for go-dnp3. We believe 
in comprehensive testing to ensure correctness, reliability, and 
interoperability.

## Testing Philosophy

### Test-Driven Development

We follow test-driven development:

1. Write failing test
2. Write minimal implementation
3. Make test pass
4. Refactor
5. Repeat

### Test Quality Over Quantity

Good tests matter more than many tests:

- Test meaningful behavior
- Cover edge cases
- Avoid trivial tests
- Focus on failure detection

### Continuous Testing

Tests run at every stage:

- Local development
- Pull request checks
- Pre-release verification
- Post-deployment validation

## Testing Layers

### Unit Tests

**Purpose**: Verify individual component behavior

**Scope**:
- Single function or method
- Single package
- Isolated dependencies

**Characteristics**:
- Fast execution
- No external dependencies
- Deterministic results
- High coverage

**Example Coverage**:
```go
// Frame encoding/decoding
func TestDataLinkFrame_Encode(t *testing.T) { ... }

// CRC calculation
func TestCRC16_Calculate(t *testing.T) { ... }

// Object encoding
func TestBinaryOutput_Encode(t *testing.T) { ... }
```

### Integration Tests

**Purpose**: Verify component interactions

**Scope**:
- Multiple packages
- Layer interactions
- Network communication

**Characteristics**:
- Slower execution
- May have external dependencies
- More realistic scenarios
- Network I/O simulated

**Example Coverage**:
```go
// Data link + Transport integration
func TestDataLinkTransport_EndToEnd(t *testing.T) { ... }

// Application layer encoding/decoding round-trip
func TestApplicationLayer_RoundTrip(t *testing.T) { ... }
```

### Conformance Tests

**Purpose**: Verify protocol compliance

**Scope**:
- IEEE 1815 specification compliance
- All function codes
- All object groups/variations

**Characteristics**:
- Specification-based
- Known inputs/outputs
- Interoperability focus
- Reference implementation comparison

**Test Data Location**:
```
test/conformance/
├── dll/           # Data link layer tests
├── tl/           # Transport layer tests
└── al/           # Application layer tests
```

### Interoperability Tests

**Purpose**: Verify compatibility with other implementations

**Scope**:
- opendnp3
- Other DNP3 masters/outstations
- Hardware devices

**Characteristics**:
- Real-world scenarios
- Diverse implementations
- Production-like conditions

**Test Data Location**:
```
test/interop/
├── captures/     # Packet captures from real devices
├── scenarios/    # Test scenarios
└── reference/    # Reference implementation outputs
```

### Fuzz Tests

**Purpose**: Discover edge cases and vulnerabilities

**Scope**:
- Input parsing
- Error handling
- Boundary conditions

**Characteristics**:
- Randomized inputs
- Automated execution
- Crash detection
- Security focus

**Tool**: Go's native fuzz testing

```go
func FuzzDataLinkFrame_Decode(f *testing.F) {
    // Fuzz test implementation
}
```

## Test Categories

### Happy Path Tests

Verify correct behavior with valid inputs:

```go
func TestDataLinkFrame_ValidFrame(t *testing.T) {
    // Test with valid frame
    // Expect correct decoding
}
```

### Edge Case Tests

Verify handling of boundary conditions:

```go
func TestDataLinkFrame_MinimumLength(t *testing.T) { ... }
func TestDataLinkFrame_MaximumLength(t *testing.T) { ... }
func TestDataLinkFrame_EmptyPayload(t *testing.T) { ... }
```

### Error Handling Tests

Verify graceful handling of invalid inputs:

```go
func TestDataLinkFrame_InvalidCRC(t *testing.T) { ... }
func TestDataLinkFrame_InvalidLength(t *testing.T) { ... }
func TestDataLinkFrame_InvalidAddress(t *testing.T) { ... }
```

### Timeout Tests

Verify timeout behavior:

```go
func TestConnection_Timeout(t *testing.T) { ... }
func TestResponse_Timeout(t *testing.T) { ... }
```

### Concurrency Tests

Verify concurrent access safety:

```go
func TestConnection_ConcurrentReads(t *testing.T) {
    // Test with goroutines
}
```

## Test Data Management

### Inline Test Data

For simple cases, inline test data:

```go
func TestCRC16_KnownValues(t *testing.T) {
    cases := []struct {
        input    []byte
        expected uint16
    }{
        {[]byte{0x00}, 0x1D0F},
        {[]byte{0xFF, 0xFF}, 0x4049},
    }
    // ...
}
```

### Test Data Files

For complex cases, external files:

```go
func TestConformance_DNP3Frames(t *testing.T) {
    data, err := os.ReadFile("testdata/frames.bin")
    // ...
}
```

### Generated Test Data

For comprehensive coverage, generate test data:

```go
func TestObjectGroup_AllVariations(t *testing.T) {
    for group := 0; group <= 255; group++ {
        for variation := 0; variation <= 255; variation++ {
            // Generate and test
        }
    }
}
```

## Test Coverage

### Coverage Goals

| Layer | Target Coverage |
|-------|----------------|
| Data Link | 90% |
| Transport | 90% |
| Application | 85% |
| Secure Auth | 80% |
| Overall | 85% |

### Coverage Measurement

Run coverage with:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Coverage Enforcement

Coverage requirements in CI:

```yaml
# .github/workflows/test.yml
- name: Test Coverage
  run: |
    go test -coverprofile=coverage.out -covermode=atomic ./...
    go tool cover -func=coverage.out | grep total
    # Must be >= 85%
```

## Mocking Strategy

### When to Mock

Use mocks for:

- External dependencies (network, filesystem)
- Slow operations
- Non-deterministic behavior
- Complex dependencies

### When NOT to Mock

Don't mock:

- Simple value objects
- Internal logic
- Protocol implementations

### Mocking Approach

Use interfaces for testability:

```go
type Clock interface {
    Now() time.Time
}

type realClock struct{}
func (c realClock) Now() time.Time { return time.Now() }

// Mock for testing
type mockClock struct {
    now time.Time
}
func (m *mockClock) Now() time.Time { return m.now }
```

## Testing Tools

### Standard Library

```go
import "testing"
```

### Testify

For assertions and mocking:

```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)
```

### GoConvey

For BDD-style tests:

```go
import "github.com/smartystreets/goconvey/convey"
```

### Quick

For property-based testing:

```go
import "github.com/stretchr/testify/quick"
```

## CI/CD Integration

### On Pull Request

1. Run all unit tests
2. Run integration tests
3. Run conformance tests
4. Check coverage
5. Run linters
6. Build binaries

### On Merge

All of the above, plus:

- Generate coverage report
- Run interoperability tests (if available)
- Security scans

### Before Release

- Full conformance suite
- Interop testing with reference implementations
- Performance benchmarks
- Manual review

## Test Execution

### Local Testing

```bash
# Run all tests
make test

# Run with coverage
go test -coverprofile=coverage.out ./...

# Run specific package
go test ./internal/dll/...

# Run with verbose output
go test -v ./...

# Run only short tests
go test -short ./...
```

### Continuous Integration

See `.github/workflows/test.yml` for CI configuration.

## Performance Testing

### Benchmarks

```go
func BenchmarkDataLinkFrame_Encode(b *testing.B) {
    // Benchmark implementation
}
```

### Profiling

```bash
go test -cpuprofile=cpu.out -memprofile=mem.out ./...
go tool pprof cpu.out
```

### Target Performance

See [007 - Performance Goals](007-performance-goals.md).
