# Test Directory

This directory contains test data, conformance tests, and test utilities.

## Structure

```
test/
├── conformance/      # Conformance test data
├── interop/          # Interoperability test data
├── fuzz/             # Fuzzing tests
└── data/             # Test data files
```

## Test Data

### Conformance Tests

Conformance test data will be stored here for verifying protocol compliance.

### Interoperability Tests

Data from real DNP3 devices and other implementations for interoperability testing.

### Fuzzing Tests

Fuzzing tests for finding edge cases and vulnerabilities.

## Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run conformance tests
go test -run Conformance ./...

# Run fuzzing tests
go test -fuzz=Fuzz ./...
```

> ⚠️ **Note**: Tests will be implemented once implementation begins.
