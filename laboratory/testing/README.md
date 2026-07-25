# KDE Testing Capability

**Version**: 1.1.0  
**Status**: Active  
**Authority**: KDE Laboratory (DNP3 Library)  
**Parent**: laboratory/  
**Created**: 2026-07-25  
**Updated**: 2026-07-25 (KDE-INV-048)
**Approval**: Recommendation 1 & 2  

---

## Overview

The Testing capability provides **shared testing infrastructure** for all laboratory activities. It owns reusable testing assets that can be consumed by Investigations, Experiments, and other engineering activities.

## Purpose

The Testing capability exists to:

1. **Own reusable testing assets** - Mock devices, simulators, fixtures
2. **Provide test infrastructure** - Shared frameworks and utilities
3. **Maintain conformance data** - Protocol test vectors and validation data
4. **Run regression tests** - Ensure ongoing quality assurance
5. **Enable reusability** - Prevent recreation of similar test assets

## Scope

### Owned Assets

The Testing capability owns the following asset categories:

| Category | Description | Examples |
|----------|-------------|----------|
| **Mock Devices** | Simulated protocol endpoints | Master mock, Outstation mock, Sensor mocks |
| **Simulators** | Protocol behavior simulators | dnp3-sim, device emulators |
| **Fixtures** | Test data and certificates | Sample datasets, TLS certificates |
| **Infrastructure** | Testing frameworks | Test harnesses, utilities |
| **Validation Tools** | Compliance verification | Scripts, validators |
| **Command-Line Tools** | Reusable CLI applications | dnp3-cli, dnp3-server |

### Simulator Ownership (per KDE-INV-048 Recommendation 1)

The Testing capability **owns** the following planned simulators:

| Simulator | Status | Location | Owner |
|-----------|--------|----------|-------|
| **dnp3-sim** | Planned | `cmd/dnp3-sim/` | Testing Capability |
| dnp3-proxy | Future | `cmd/dnp3-proxy/` | Testing Capability |
| dnp3-cli | Future | `cmd/dnp3-cli/` | Testing Capability |
| dnp3-server | Future | `cmd/dnp3-server/` | Testing Capability |

### Services

| Service | Description |
|---------|-------------|
| **Asset Provision** | Provides mock devices, simulators to Investigations/Experiments |
| **Test Execution** | Runs regression tests and reports results |
| **Conformance Coordination** | Maintains conformance test data |
| **Infrastructure Maintenance** | Maintains testing frameworks and tools |
| **Promotion Review** | Reviews test assets for promotion |

## Directory Structure

```
laboratory/testing/
├── README.md              # This file
├── assets/               # Shared testing assets (catalog)
├── mocks/               # Mock device specifications
├── simulators/          # Simulator specifications
├── fixtures/            # Test fixtures and data
├── governance/          # Testing governance rules
└── catalog.md           # Asset catalog
```

## Ownership Boundaries

### What Testing Owns

| Asset Type | Location | Ownership |
|------------|----------|-----------|
| Integration tests | `test/integration/` | Testing |
| Conformance data | `test/conformance/` | Testing |
| Mock devices | In test files | Testing |
| Simulators | `cmd/` (planned) | Testing |
| Validation scripts | `scripts/` (planned) | Testing |
| Benchmarks | `benchmarks/` | Testing |

### What Testing Does NOT Own

| Asset Type | Owner | Reason |
|------------|-------|--------|
| Investigation-specific tests | Investigation | Created for specific investigation |
| Experiment-specific tests | Experiment | Created for specific experiment |
| Production code | Respective packages | Not testing infrastructure |
| Investigation artifacts | Investigation | Investigation scope |
| Experiment artifacts | Experiment | Experiment scope |

## Lifecycle

### Asset Lifecycle

```
Investigation creates test asset
    ↓
Asset evaluated for reusability
    ↓
Testing reviews for promotion
    ↓
Asset promoted to Testing (if reusable)
    ↓
Testing maintains asset
    ↓
Available to all future activities
```

### Promotion Criteria

An asset is promoted to Testing ownership if:
1. **Reusable**: Can be used by multiple investigations
2. **Stable**: Has been validated and tested
3. **Documented**: Has adequate documentation
4. **Maintained**: Can be maintained by Testing

## Usage

### For Investigations

1. **Request assets** from Testing before creating new test infrastructure
2. **Use existing mocks** instead of creating new ones
3. **Document new assets** that could be promoted
4. **Return reusable assets** to Testing after investigation

### For Experiments

1. **Request simulators** from Testing for validation
2. **Use existing fixtures** instead of creating new ones
3. **Document reusable assets** created during experiment
4. **Coordinate with Testing** for validation scripts

### For Contributors

1. **Follow naming conventions** (TEST-* prefix)
2. **Document assets** before contributing
3. **Include tests** with contributed assets
4. **Submit for review** to Testing maintainer

## Governance

### Rules

| Rule ID | Description |
|---------|-------------|
| T-GOV-001 | Testing owns all assets in test/, cmd/, scripts/, benchmarks/ |
| T-GOV-002 | Investigation-specific tests remain owned by Investigation |
| T-GOV-003 | Experiment-specific tests remain owned by Experiment |
| T-GOV-004 | Reusable assets must be approved by Testing before promotion |
| T-GOV-005 | All Testing assets must follow TEST-* naming convention |

### Authority

| Decision | Authority |
|----------|-----------|
| Create new Testing asset | Testing maintainer |
| Promote asset to Testing | Testing review |
| Retire Testing asset | Testing + Governance approval |
| Modify Testing infrastructure | Testing maintainer |

## Execution Environment (per KDE-INV-048 Recommendation 2)

### Runtime Requirements

The Testing capability requires the following execution environment:

| Component | Version | Source | Owner |
|-----------|---------|--------|-------|
| Go compiler | 1.22.0+ | go.dev | Project (go.mod) |
| Go toolchain | Latest stable | go.dev | Project |

### Environment Policy

1. **Minimum Version**: Go 1.22.0 (per `go.mod`)
2. **Recommended Version**: Latest stable Go release
3. **Installation**: User-managed (see `go.mod` for requirements)
4. **Verification**: `go version` should confirm >= 1.22.0

### Build Verification

All Testing assets must verify the environment before execution:

```bash
go version  # Must be >= 1.22.0
go build ./...  # Must succeed
go test ./...  # Must pass
```

## Related Documents

| Document | Purpose |
|----------|---------|
| `laboratory/investigations/KDE-INV-047/` | Testing capability investigation |
| `laboratory/README.md` | Laboratory overview |
| `.kde/governance/NAMING-CONVENTIONS.md` | Naming policy |

---

*Generated per KDE-INV-047 recommendation - Model B*
*Created: 2026-07-25*
