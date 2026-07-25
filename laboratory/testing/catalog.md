# Testing Asset Catalog

**Document ID**: TEST-CATALOG-001
**Version**: 1.1.0
**Date**: 2026-07-25
**Status**: ACTIVE
**Authority**: KDE Testing Capability
**Updated**: 2026-07-25 (KDE-INV-048)

---

## Overview

This catalog lists all testing assets owned by the Testing capability. Assets are organized by category and include ownership, status, and usage information.

---

## Asset Categories

| Category | Description | Count |
|----------|-------------|-------|
| Mock Devices | Simulated protocol endpoints | 4 |
| Integration Tests | End-to-end communication tests | 2 |
| Conformance Data | Protocol test vectors | 3 |
| Fixtures | Test data and certificates | 0 |
| Simulators | Protocol behavior simulators | 0 |
| Benchmarks | Performance tests | 3 |

---

## 1. Mock Devices

### 1.1 In-Memory Master Mock

| Property | Value |
|----------|-------|
| **Asset ID** | TEST-MOCK-MASTER-001 |
| **Location** | test/integration/master_outstation_test.go |
| **Type** | Mock Device |
| **Status** | Active |
| **Created** | 2026-07-25 |
| **Owner** | Testing Capability |
| **Dependencies** | internal/master, internal/al |

**Purpose**: In-memory DNP3 Master for testing Outstation implementations.

**Usage**:
```go
// From test/integration/master_outstation_test.go
ost := outstation.NewOutstation(nil)
ost.Initialize()
ost.Start()
```

**Reusable**: Yes - Can be used by any investigation testing Outstation behavior.

---

### 1.2 In-Memory Outstation Mock

| Property | Value |
|----------|-------|
| **Asset ID** | TEST-MOCK-OUTSTATION-001 |
| **Location** | test/integration/master_outstation_test.go |
| **Type** | Mock Device |
| **Status** | Active |
| **Created** | 2026-07-25 |
| **Owner** | Testing Capability |
| **Dependencies** | internal/outstation, internal/al |

**Purpose**: In-memory DNP3 Outstation with configurable data.

**Usage**:
```go
customData := &CustomDataHandler{
    binaryInputs: []outstation.BinaryInput{{Value: true}},
}
ost.SetDataHandler(customData)
```

**Reusable**: Yes - Configurable for various test scenarios.

---

### 1.3 Custom Data Handler

| Property | Value |
|----------|-------|
| **Asset ID** | TEST-MOCK-DATAHANDLER-001 |
| **Location** | test/integration/master_outstation_test.go |
| **Type** | Mock Component |
| **Status** | Active |
| **Created** | 2026-07-25 |
| **Owner** | Testing Capability |

**Purpose**: Configurable data provider for Outstation testing.

**Types**:
- `CustomDataHandler` - Basic mock with custom data
- `comprehensiveDataHandler` - Full mock with binary, analog, counters

**Reusable**: Yes - Generic interface for any data scenario.

---

### 1.4 Command Handler Mock

| Property | Value |
|----------|-------|
| **Asset ID** | TEST-MOCK-COMMANDHANDLER-001 |
| **Location** | test/integration/tcp_test.go |
| **Type** | Mock Component |
| **Status** | Active |
| **Created** | 2026-07-25 |
| **Owner** | Testing Capability |

**Purpose**: Tracks command execution for validation.

**Usage**:
```go
type comprehensiveCommandHandler struct {
    executedCommands []string
}
```

**Reusable**: Yes - Can track any command execution pattern.

---

## 2. Integration Tests

### 2.1 Master-Outstation Integration Tests

| Property | Value |
|----------|-------|
| **Asset ID** | TEST-INT-MASTER-001 |
| **Location** | test/integration/master_outstation_test.go |
| **Type** | Integration Test Suite |
| **Status** | Active |
| **Created** | 2026-07-25 |
| **Owner** | Testing Capability |
| **Test Count** | 8 tests |

**Tests Included**:
| Test | Purpose |
|------|---------|
| TestOutstationProcessReadRequest | READ request processing |
| TestOutstationProcessWriteRequest | WRITE request processing |
| TestOutstationProcessEnableUnsolicited | ENABLE UNSOLICITED handling |
| TestOutstationProcessUnsupportedRequest | Error handling |
| TestOutstationStateTransitions | State machine |
| TestOutstationIIN | Internal indication |
| TestOutstationDefaultDataHandler | Default data |
| TestOutstationCustomDataHandler | Custom data |

**Reusable**: Yes - Tests validate core Outstation behavior.

---

### 2.2 TCP Integration Tests

| Property | Value |
|----------|-------|
| **Asset ID** | TEST-INT-TCP-001 |
| **Location** | test/integration/tcp_test.go |
| **Type** | Integration Test Suite |
| **Status** | Active |
| **Created** | 2026-07-25 |
| **Owner** | Testing Capability |
| **Test Count** | 4 tests |

**Tests Included**:
| Test | Purpose |
|------|---------|
| TestTCPMasterOutstationRead | TCP read operation |
| TestTCPDirectCommunication | End-to-end TCP |
| TestTCPTransportAcceptMultipleConnections | Multi-connection |
| TestMasterOutstationEndToEndComprehensive | Full capability test |

**Reusable**: Yes - Tests validate TCP transport integration.

---

## 3. Conformance Data

### 3.1 Application Layer Test Data

| Property | Value |
|----------|-------|
| **Asset ID** | TEST-CONF-AL-001 |
| **Location** | test/conformance/al/ |
| **Type** | Conformance Data |
| **Status** | Planned |
| **Owner** | Testing Capability |

**Purpose**: Test vectors for application layer compliance.

**Planned Content**:
- APDU encoding/decoding test vectors
- Function code validation data
- IIN flag test cases

---

### 3.2 Data Link Layer Test Data

| Property | Value |
|----------|-------|
| **Asset ID** | TEST-CONF-DLL-001 |
| **Location** | test/conformance/dll/ |
| **Type** | Conformance Data |
| **Status** | Planned |
| **Owner** | Testing Capability |

**Purpose**: Test vectors for data link layer compliance.

**Planned Content**:
- Frame encoding/decoding test vectors
- CRC16 validation data
- Control byte test cases

---

### 3.3 Transport Layer Test Data

| Property | Value |
|----------|-------|
| **Asset ID** | TEST-CONF-TL-001 |
| **Location** | test/conformance/tl/ |
| **Type** | Conformance Data |
| **Status** | Planned |
| **Owner** | Testing Capability |

**Purpose**: Test vectors for transport layer compliance.

**Planned Content**:
- Fragmentation test vectors
- Reassembly test data
- Sequence number validation

---

## 4. Fixtures

No fixtures currently cataloged. Fixtures will be added as they are created.

**Planned Fixtures**:
| Fixture | Purpose | Status |
|---------|---------|--------|
| TLS Certificates | Secure transport testing | Planned |
| Sample Datasets | Mock data files | Planned |
| Configuration Files | Test configurations | Planned |

---

## 5. Simulators (per KDE-INV-048 Recommendation 1)

### 5.1 DNP3 Device Simulator

| Property | Value |
|----------|-------|
| **Asset ID** | TEST-SIM-DNP3-001 |
| **Name** | dnp3-sim |
| **Location** | `cmd/dnp3-sim/` (planned) |
| **Type** | Simulator |
| **Status** | **OWNED - Implementation Pending** |
| **Owner** | Testing Capability |
| **Created** | 2026-07-25 |
| **Dependency** | go 1.22.0+, dnp3 library |

**Purpose**: DNP3 device simulator for testing DNP3 Master implementations and protocol compliance.

**Planned Features**:
- Configurable Outstation behavior
- Support for Binary, Analog, Counter data types
- Configurable response patterns
- Error injection capabilities
- Unsolicited response support

**Implementation Status**: Planned - awaiting implementation

**Priority**: HIGH - Required for Testing capability validation

---

**Planned Simulators**:
| Simulator | Purpose | Status | Owner |
|-----------|---------|--------|-------|
| **dnp3-sim** | DNP3 device simulator | **OWNED** | Testing Capability |
| dnp3-cli | Command-line client | Future | Testing Capability |
| dnp3-server | Test server | Future | Testing Capability |
| Traffic Generator | Packet generation | Future | Testing Capability |

---

## 6. Benchmarks

### 6.1 Application Layer Benchmarks

| Property | Value |
|----------|-------|
| **Asset ID** | TEST-BENCH-AL-001 |
| **Location** | benchmarks/al_bench_test.go |
| **Type** | Benchmark Suite |
| **Status** | Active |
| **Owner** | Testing Capability |

**Benchmarks**:
- APDU encoding/decoding performance
- Application control field operations
- IIN flag operations

---

### 6.2 Data Link Layer Benchmarks

| Property | Value |
|----------|-------|
| **Asset ID** | TEST-BENCH-DLL-001 |
| **Location** | benchmarks/dll_bench_test.go |
| **Type** | Benchmark Suite |
| **Status** | Active |
| **Owner** | Testing Capability |

**Benchmarks**:
- Frame encoding/decoding performance
- CRC16 calculation speed
- Control byte operations

---

### 6.3 Transport Layer Benchmarks

| Property | Value |
|----------|-------|
| **Asset ID** | TEST-BENCH-TL-001 |
| **Location** | benchmarks/tl_bench_test.go |
| **Type** | Benchmark Suite |
| **Status** | Active |
| **Owner** | Testing Capability |

**Benchmarks**:
- Fragmentation performance
- Reassembly performance
- Header operations

---

## Asset Promotion Candidates

The following assets were created during investigations and are candidates for Testing ownership:

| Asset | Source | Status | Recommendation |
|-------|--------|--------|----------------|
| Mock data handlers | KDE-INV-046 | ✅ Already Testing | Adopt |
| TCP integration tests | KDE-INV-046 | ✅ Already Testing | Adopt |
| Comprehensive test data | KDE-INV-046 | ✅ Already Testing | Adopt |

---

## Asset Requests

No pending asset requests.

---

## Maintenance Log

| Date | Asset | Action | Maintainer |
|------|-------|--------|------------|
| 2026-07-25 | All assets | Initial cataloging | Testing Capability |
| 2026-07-25 | dnp3-sim | Assigned ownership (KDE-INV-048 Rec 1) | Testing Capability |
| 2026-07-25 | Execution Environment | Documented requirements (KDE-INV-048 Rec 2) | Testing Capability |

---

**Status**: ACTIVE
**Last Updated**: 2026-07-25
**Next Review**: Quarterly or upon significant change

---

*Generated per KDE-INV-047 recommendation*
*Updated per KDE-INV-048 recommendations*
*Maintained by: Testing Capability*
