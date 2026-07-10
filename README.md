# go-dnp3

**A native Go implementation of IEEE 1815 (DNP3)**

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Status](https://img.shields.io/badge/status-research%20%26%20architecture-red.svg)](#current-status)

> ⚠️ **Notice**: This project is in the **research and architecture phase**. No protocol implementation exists yet.

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

### 🔴 Research and Architecture Phase

**No implementation exists yet.**

This repository is currently establishing:

1. Engineering foundation and project identity
2. Architecture documentation
3. Development methodology
4. Research into protocol invariants

### Roadmap

| Phase | Status | Description |
|-------|--------|-------------|
| Phase 1: Research | 🔄 In Progress | Protocol analysis and research |
| Phase 2: Architecture | 📋 Planned | Architecture documentation |
| Phase 3: Architecture Review | 📋 Planned | Review and approval |
| Phase 4: Implementation | 📋 Planned | Protocol implementation |
| Phase 5: Conformance Testing | 📋 Planned | Test suite development |
| Phase 6: Optimization | 📋 Planned | Performance tuning |
| Phase 7: Production Release | 📋 Planned | Initial stable release |

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
