# go-dnp3

**A native Go implementation of IEEE 1815 (DNP3)**

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Status](https://img.shields.io/badge/status-core%20working%20%F0%9F%9F%97-green.svg)](#current-status)
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

## Quick Start

```bash
# Clone and build
git clone https://github.com/tamzrod/dnp3.git
cd dnp3
go build ./...

# Run tests
go test ./pkg/... ./internal/... ./test/integration/... -count=1

# Build workbench
go build -o workbench ./cmd/workbench

# Test Master↔Outstation (two terminals)
# Terminal 1:
./workbench -mode outstation -address 0.0.0.0 -port 20000
# Terminal 2:
./workbench -mode master -address 127.0.0.1 -port 20000
```

See [cmd/workbench/README.md](cmd/workbench/README.md) for workbench usage.

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

### 🟢 Verified Master MVP (v0 interoperability profile)

The repository implements and internally verifies the v0 Master interoperability
profile: one TCP master connects to one outstation, performs a Class-0 read of
Binary Input (G1V1), Analog Input (G30V1), and Counter (G20V1), and issues one
Direct-Operate control (G12V1 CROB). The full MVP path is exercised end-to-end
against the deterministic in-memory simulator in
[`test/integration/mvp_loopback_test.go`](test/integration/mvp_loopback_test.go)
(DNP3-045), with no network I/O.

**Verified capabilities (each has an in-repo test reference — see
[`active_work/supported-profile.md`](active_work/supported-profile.md) for the
authoritative matrix):**
- ✅ Data Link Layer (DLL) — frame encode/decode, CRC16, link handshake
  (`internal/dll/link/link_test.go`)
- ✅ Transport Layer (TL) — segmentation/reassembly, fragment sequence
  (`internal/tl`)
- ✅ Application Layer (AL) — APDU encode/decode, object-header model
  (`internal/al/object_header_test.go`)
- ✅ LSB-first wire encoding for all MVP multi-octet fields (indices, analog/
  counter values, CROB times) — corrected and golden-verified (DNP3-001)
- ✅ Master `Connect` / `Disconnect` / `Close` lifecycle with context
  cancellation; client is reusable after `Close` (DNP3-050,
  `test/integration/close_reuse_test.go`)
- ✅ Master `Read` / `IntegrityPoll` (Class-0 G1/G30/G20) and `Operate`
  (Direct-Operate G12V1) with command-status reporting
- ✅ Retry, timeout, and outstanding-request tracking; optional idle-timeout
  keep-alive close
- ✅ Public error taxonomy + `ClassifyError` (DNP3-043)
- ✅ Optional diagnostic logger hook, no-op/silent by default (DNP3-044)
- ✅ Master/outstation link-layer address validation (DNP3-049)
- ✅ Workbench TUI — Master and Outstation modes

**Explicitly unsupported in v0 (rejected by the public API until a separately
tested profile adds them):**
- ⛔ TLS and Serial transports (TCP only)
- ⛔ Unsolicited responses / event delivery
- ⛔ Secure authentication, time sync, file transfer, device attributes
- ⛔ Restart, delay measurement, freeze operations
- ⛔ Select-before-operate and direct-operate-no-response
- ⛔ Object groups/variations outside G1V1, G30V1, G20V1, G12V1

**External interop status (MEXT series lock):**
- 🔒 **External interoperability is NOT claimed.** The Master is verified for
  **internal use only** (deterministic fixtures + in-repo simulator loopback).
- 🔒 Do not assume production-ready or third-party-interop status. An external
  claim is blocked until **MEXT-035** passes (`./scripts/verify-external-mvp.sh`
  exit 0, plus CROB/Operate/multi-header gates). See
  [`active_work/MEXT_MASTER_ROADMAP.md`](active_work/MEXT_MASTER_ROADMAP.md) and
  [`active_work/external-acceptance.md`](active_work/external-acceptance.md).
- Known external blockers (residuals R1–R5) are listed in
  [`active_work/supported-profile.md`](active_work/supported-profile.md).

**Not yet verified:**
- ⚠️ External interoperability (independent raw capture / VEC-01) is still
  pending. All current verification is internal (deterministic fixtures and
  the in-repo simulator loopback).
- ⚠️ Not IEEE 1815 complete (only the v0 object subset is implemented).
- ⚠️ Workbench is an engineering tool, not production SCADA.

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
