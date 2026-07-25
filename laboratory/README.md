# KDE Engineering Laboratory

**Version**: 3.0.0  
**Status**: Initialized  
**Project**: DNP3 Library

---

## Overview

This is the **official** and **only** engineering laboratory for the DNP3 Library project. All engineering artifacts must be written to this directory.

## Directory Structure

```
laboratory/
├── decisions/           # Technology Decision Records (TDRs)
├── investigations/      # Investigation documents
├── experiments/         # Laboratory experiments
├── testing/            # Testing capability (shared testing assets)
├── evidence/           # Evidence artifacts
├── implementations/    # Implementation specifications (IMPs)
├── planning/           # Planning documents
└── reviews/            # Review documents
```

## Laboratory Capabilities

This laboratory contains **ONLY** engineering artifacts organized into capabilities:

| Capability | Directory | Purpose | Status |
|-----------|-----------|---------|--------|
| **Investigation** | `investigations/` | Systematic inquiry and analysis | ✅ Active |
| **Experiment** | `experiments/` | Controlled hypothesis validation | ✅ Available |
| **Testing** | `testing/` | Reusable testing infrastructure | ✅ Active |
| **Decision** | `decisions/` | Technology Decision Records | ✅ Active |
| **Evidence** | `evidence/` | Evidence artifacts | ✅ Available |
| **Implementation** | `implementations/` | Implementation specifications | ✅ Available |
| **Planning** | `planning/` | Planning documents | ✅ Available |
| **Review** | `reviews/` | Review documents | ✅ Available |

## Testing Capability

The **Testing capability** provides shared testing infrastructure for all laboratory activities:

### Purpose

- Owns reusable testing assets (mocks, simulators, fixtures)
- Provides test infrastructure to Investigations and Experiments
- Maintains conformance test data and validation tools
- Runs regression tests and reports results

### Testing Assets

| Category | Examples |
|----------|----------|
| Mock devices | In-memory Master, Outstation, Sensors |
| Simulators | Protocol simulators, device emulators |
| Test fixtures | Sample datasets, certificates |
| Validation tools | Scripts, harnesses, utilities |
| Conformance data | Protocol test vectors |

### Testing Services

1. **Asset Provision**: Provides mock devices, simulators to Investigations/Experiments
2. **Test Execution**: Runs regression tests and reports results
3. **Conformance Coordination**: Maintains conformance test data
4. **Infrastructure Maintenance**: Maintains testing frameworks and tools
5. **Promotion Review**: Reviews test assets for promotion to shared testing

## Separation of Concerns

### Runtime (.kde/)

The KDE Runtime (`/.kde/`) contains **ONLY**:
- Bootstrap
- Runtime
- Engines
- Seeds
- Governance
- Commands
- Capabilities
- Templates
- Verification

### Laboratory (laboratory/)

The Laboratory contains **ALL** engineering artifacts organized by capability:

| Capability | Scope |
|-----------|-------|
| Investigation | Understanding and analysis |
| Experiment | Hypothesis validation |
| Testing | Reusable infrastructure and quality assurance |
| Decision | Technology decisions |
| Evidence | Evidence artifacts |
| Implementation | Implementation specifications |
| Planning | Planning documents |
| Review | Review documents |

### Testing Capability Ownership

The Testing capability owns the following directories (conceptual ownership):

| Directory | Testing Ownership | Current Status |
|-----------|-------------------|---------------|
| `test/` | Conceptual owner | Active |
| `cmd/` | Conceptual owner | Planned |
| `scripts/` | Conceptual owner | Empty |
| `benchmarks/` | Conceptual owner | Active |

## Engineering Principles

1. **Evidence Over Intuition** - All decisions must be evidence-based
2. **Investigation Before Implementation** - No implementation without investigation
3. **Human Authorization** - Significant changes require human approval
4. **Traceability Always** - Every conclusion must trace to evidence

## Naming Conventions

All laboratory artifacts **MUST** follow the naming conventions:

| Artifact Type | Directory | Prefix | Example |
|--------------|-----------|--------|---------|
| Investigation (KDE) | `investigations/` | `KDE-INV-` | `KDE-INV-001/` |
| Investigation (Project) | `investigations/` | `PROJECT-INV-` | `PROJECT-INV-001/` |
| Experiment | `experiments/` | `PROJECT-EXP-` | `PROJECT-EXP-001/` |
| Decision | `decisions/` | `TDR-` | `TDR-001.md` |
| Review | `reviews/` | `PROJECT-REV-` | `PROJECT-REV-001/` |
| Implementation | `implementations/` | `PROJECT-IMP-` | `PROJECT-IMP-001/` |
| Testing Asset | `testing/` | `TEST-` | `TEST-ASSET-001.md` |

**CRITICAL**: 
- Investigations **MUST** use `INV-` prefix (not `EXP-`)
- Experiments **MUST** use `EXP-` prefix
- Testing assets **MUST** use `TEST-` prefix
- Cross-prefixing (e.g., `investigations/PROJECT-EXP-XXX/`) is a **naming violation**

See `.kde/governance/NAMING-CONVENTIONS.md` for full policy.

## Usage

All engineering artifacts for the project should be created in this laboratory directory. The KDE Runtime will reference these artifacts but does not own them.

---

*Generated by KDE Bootstrap Template - 2026-07-25*
