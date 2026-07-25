# go-dnp3

**A native Go implementation of IEEE 1815 (DNP3)**

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Status](https://img.shields.io/badge/status-architecture%20complete%20%F0%9F%9F%97-green.svg)](#current-status)
[![KDE](https://img.shields.io/badge/KDE-Governed-blue.svg)](.kde/README.md)

> ✅ **Notice**: This project is governed by the KDE (Knowledge Discovery Engine) Runtime. See [KDE Runtime Environment](#kde-runtime-environment) below.

---

## KDE Runtime Environment

This repository is initialized with the **KDE Runtime Environment** (`.kde/`), providing evidence-based engineering governance.

### About .kde/

The `.kde/` directory is the local engineering governance directory, similar to how `.git/` manages version control. It contains:

- **bootstrap/** - Runtime bootstrap and initialization
- **runtime/** - Core runtime system
- **engines/** - Investigation and decision engines
- **experts/** - Domain expert knowledge bases
- **knowledge/** - Engineering knowledge base
- **governance/** - Governance policies
- **seeds/** - Seed knowledge
- **commands/** - System commands
- **capabilities/** - System capabilities
- **templates/** - Artifact templates
- **verification/** - Verification system

### Quick Reference

```bash
# Check KDE status
cat .kde/runtime/state.json

# View bootstrap configuration
cat .kde/bootstrap/config.yaml

# View governance policies
cat .kde/governance/NAMING-CONVENTIONS.md
```

### KDE Resources

- [KDE Runtime Environment](.kde/README.md)
- [KDE Governance Documentation](docs/kde/README.md)
- [Engineering Laboratory](laboratory/README.md)
- [KDE Bootstrap Report](KDE-BOOTSTRAP-REPORT.md)

---

## What is DNP3?

**DNP3** (Distributed Network Protocol 3) is a communication protocol widely used in SCADA systems, particularly in the electric utility, water, and wastewater industries. It was originally developed by GE Harris and is now maintained by IEEE Standards Association as **IEEE 1815**.

DNP3 is designed for:
- **Supervisory Control and Data Acquisition (SCADA)** systems
- **Critical infrastructure** communication
- **Industrial automation** applications
- **Real-time data exchange** between master stations and outstations

### Protocol Characteristics

- **Transport**: TCP/IP, Serial (ANSI/TIA-232-E, EIA-485)
- **Layers**: Data link, transport, and application layers
- **Features**: Event-based reporting, time synchronization, file transfer, security (IEEE 1815-2012)
- **Use Cases**: Power grid management, water treatment, oil and gas pipelines

---

## Why This Project?

The Go ecosystem lacks a **native** DNP3 implementation. Existing solutions are wrappers or ports that:

- Do not follow Go idioms
- Embed dependencies on C libraries
- Lack proper Go concurrency patterns
- Cannot leverage Go's modern tooling

This project exists to create **the canonical native Go implementation** of DNP3, designed from protocol invariants rather than ported from other languages.

---

## Project Philosophy

### Native Go, Not a Wrapper

This is not a wrapper around lib60870, opendnp3, or any other implementation.

- Pure Go codebase
- No C dependencies
- Designed from protocol specifications
- Validated against multiple mature implementations

### Architecture First

We believe good software requires good architecture.

- Document before implementing
- Research before designing
- Test before deploying
- Every decision has rationale

### Long-Term Maintainability

We prioritize sustainable software over short-term development speed.

- Clear, documented code
- Comprehensive test coverage
- Minimal dependencies
- Predictable behavior

---

## Scope

### In Scope

- Complete IEEE 1815-2012 implementation
- Data link layer (unbalanced and balanced modes)
- Transport layer
- Application layer (all function codes)
- Secure authentication (IEC 62351-6)
- Conformance testing infrastructure
- Performance benchmarking

### Out of Scope

- DNP3 over serial (future consideration)
- Wrapper/compatibility layers for other implementations
- Protocol variations not in IEEE 1815

---

## Current Status

### 🟡 Partial Implementation - Integration Incomplete

**Implementation has begun.** The repository contains 41 Go source files implementing core protocol layers. Integration between layers and end-to-end functionality requires completion.

**Implemented**:
- ✅ Data Link Layer (DLL) - Frame encoding/decoding, CRC16, link state machine
- ✅ Transport Layer (TL) - Segmentation, reassembly, flow control
- ✅ Application Layer (AL) - APDU encoding/decoding, function codes, IIN
- ✅ Secure Authentication (SA) - Challenge handling, key management
- ✅ Master Role - Client interface, state machine, read/operate commands
- ✅ Outstation Role - Server interface, state machine, data handling
- ✅ TCP Transport - TCP/IP connectivity
- ✅ TLS Transport - Stub implementation
- ✅ Unit tests (22 test files)
- ✅ Performance benchmarks (3 benchmark suites)

**Incomplete / Missing**:
- ⚠️ Public API wiring to internal implementations
- ⚠️ End-to-end integration tests (CI tests disabled)
- ⚠️ Object group variations (partial)
- ❌ Examples (none exist)
- ❌ CLI tools (none)
- ❌ Serial transport (out of scope)

**Repository Statistics**:
| Metric | Count |
|--------|-------|
| Source files | 41 |
| Test files | 22 |
| Documentation | 265 |
| Architecture docs | 11 |

---

## Architecture Documentation

The architecture documentation is maintained in `docs/architecture/`:

- [000 - Philosophy](docs/architecture/000-philosophy.md)
- [001 - Goals](docs/architecture/001-goals.md)
- [002 - Non-Goals](docs/architecture/002-non-goals.md)
- [003 - Guiding Principles](docs/architecture/003-guiding-principles.md)
- [004 - Package Architecture](docs/architecture/004-package-architecture.md)
- [005 - Development Methodology](docs/architecture/005-development-methodology.md)
- [006 - Testing Strategy](docs/architecture/006-testing-strategy.md)
- [007 - Performance Goals](docs/architecture/007-performance-goals.md)
- [008 - Concurrency Model](docs/architecture/008-concurrency-model.md)
- [009 - Memory Model](docs/architecture/009-memory-model.md)

---

## Contributing

We welcome contributions, but please read our [Contributing Guide](CONTRIBUTING.md) before submitting changes.

**Important**: This project follows an **architecture-first development methodology**. Implementation is prohibited until the architecture has been defined and approved.

## Code of Conduct

Please read our [Code of Conduct](CODE_OF_CONDUCT.md) before participating in this project.

## Security

For security concerns, please read our [Security Policy](SECURITY.md).

---

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

Copyright 2024 the go-dnp3 contributors
