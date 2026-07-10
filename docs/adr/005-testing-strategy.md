# ADR-005: Testing Strategy

## Status

Accepted

## Context

We need a testing strategy that:
1. Ensures protocol conformance
2. Validates interoperability with other DNP3 implementations
3. Achieves high code coverage
4. Tests both unit and integration scenarios
5. Enables fuzz testing for robustness

## Decision

We will use a multi-layered testing approach:

### Test Pyramid

```
         ┌─────────────────┐
         │  Fuzz Testing   │  Rare, high value
         ├─────────────────┤
         │ Integration     │  Layer interactions
         ├─────────────────┤
         │    Unit Tests   │  Common, fast
         └─────────────────┘
```

### Unit Tests

Each package has corresponding `_test.go` files:

```
internal/dll/
├── frame/
│   ├── frame_test.go
│   └── encode_test.go
├── link/
│   └── link_test.go
└── crc/
    └── crc_test.go
```

#### Test Categories

| Category | Purpose | Example |
|----------|---------|---------|
| Happy path | Normal operation | Encode/decode valid frame |
| Edge cases | Boundary conditions | Empty data, max size |
| Error cases | Invalid input | Invalid CRC, wrong length |
| Property tests | Invariant verification | Round-trip encoding |

#### Example Test Structure

```go
func TestFrameEncode(t *testing.T) {
    tests := []struct {
        name    string
        frame   Frame
        want    []byte
        wantErr bool
   }{
        {
            name:  "reset link stations",
            frame: Frame{Function: ResetLinkStations, Primary: true},
            want:  []byte{0x05, 0x64, 0x49, 0xFF, 0xFF},
            wantErr: false,
        },
        // ... more cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Encode(tt.frame)
            if (err != nil) != tt.wantErr {
                t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !tt.wantErr && !bytes.Equal(got, tt.want) {
                t.Errorf("Encode() = %x, want %x", got, tt.want)
            }
        })
    }
}
```

### Table-Driven Tests

Use table-driven tests for comprehensive coverage:

```go
func TestCRCCalculation(t *testing.T) {
    cases := []struct {
        data    []byte
        expect  uint16
    }{
        {[]byte{0x01, 0x02}, 0x1234},
        {[]byte{}, 0x0000},
        // ... generate from known values
    }
    
    for _, c := range cases {
        got := CRC16(c.data)
        if got != c.expect {
            t.Errorf("CRC16(%x) = %x, want %x", c.data, got, c.expect)
        }
    }
}
```

### Integration Tests

Cross-layer tests in `test/integration/`:

```
test/
├── integration/
│   ├── master_test.go
│   └── outstation_test.go
└── conformance/
    ├── dll_test.go
    ├── tl_test.go
    └── al_test.go
```

### Test Fixtures

Use golden files for complex test data:

```
test/
└── fixtures/
    ├── frames/
    │   ├── reset_link.bin
    │   └── confirmed_data.bin
    └── pdus/
        ├── read_request.bin
        └── response.bin
```

### Conformance Test Data

Protocol conformance test cases from specification:

```go
func TestConformance_LinkLayerReset(t *testing.T) {
    // From IEEE 1815-2012 Annex B
    // Verify exact behavior for reset link stations
}
```

### Fuzz Testing

Fuzz tests for robustness:

```go
func FuzzFrameDecode(f *testing.F) {
    // Corpus of valid frames
    f.Add([]byte{0x05, 0x64, 0x49, 0xFF, 0xFF, 0x12, 0x34})
    
    f.Fuzz(func(t *testing.T, data []byte) {
        _, err := Decode(data)
        // Just verify it doesn't panic
    })
}
```

### Race Detection

Run all tests with race detector:

```bash
go test -race ./...
```

### Coverage Requirements

| Layer | Target Coverage |
|-------|---------------|
| internal/dll | 90% |
| internal/tl | 90% |
| internal/al | 85% |
| pkg/dnp3 | 80% |

### Mock Strategy

Minimize mocks. Prefer:
1. Real implementations for internal tests
2. Interfaces for external dependencies
3. Test implementations for interfaces

```go
// Define interface for testability
type Logger interface {
    Debug(msg string, args ...interface{})
    Info(msg string, args ...interface{})
    Error(msg string, args ...interface{})
}

// Real implementation
type ProductionLogger struct{}

func (l *ProductionLogger) Debug(msg string, args ...interface{}) { /* ... */ }

// Test implementation
type TestLogger struct {
    Logs []string
}

func (l *TestLogger) Debug(msg string, args ...interface{}) {
    l.Logs = append(l.Logs, msg)
}
```

## Consequences

### Positive

- High confidence in correctness
- Protocol conformance verified
- Regression protection
- Fuzzing finds edge cases
- Clear test organization

### Negative

- More test code to maintain
- Test fixtures need management
- Fuzzing can be slow
- Coverage targets may be aspirational

### Trade-offs

We prioritize correctness and confidence over development speed.

## Traceability

- Architecture: [docs/architecture/006-testing-strategy.md](docs/architecture/006-testing-strategy.md)
- Protocol: Conformance testing per IEEE 1815-2012

## Related Decisions

- ADR-001: Package Structure (tests follow package layout)
- ADR-002: Error Handling Strategy (errors are testable)
