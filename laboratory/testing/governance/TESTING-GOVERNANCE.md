# KDE Testing Governance Rules

**Document ID**: TESTING-GOV-001
**Version**: 1.0.0
**Date**: 2026-07-25
**Status**: APPROVED
**Authority**: KDE Laboratory (per KDE-INV-047)

---

## Purpose

This document establishes governance rules for the Testing capability within the KDE Laboratory. These rules define ownership, lifecycle, promotion, and access for testing assets.

---

## 1. Ownership Rules

### Rule T-OWN-001: Testing Asset Ownership

**Rule**: Testing owns all artifacts in the following directories:

| Directory | Testing Ownership | Current Status |
|-----------|-------------------|----------------|
| `test/` | Full ownership | Active |
| `cmd/` | Full ownership | Planned |
| `scripts/` | Full ownership | Empty |
| `benchmarks/` | Full ownership | Active |

**Rationale**: Centralizing testing assets enables reusability across investigations.

### Rule T-OWN-002: Investigation-Specific Tests

**Rule**: Tests created specifically for an investigation remain owned by that investigation.

| Condition | Ownership |
|-----------|-----------|
| Created during investigation | Investigation |
| Explicitly marked for promotion | Investigation → Testing |
| No clear purpose beyond investigation | Investigation |

**Rationale**: Investigation-specific tests may not be reusable and should not burden Testing.

### Rule T-OWN-003: Experiment-Specific Tests

**Rule**: Tests created specifically for an experiment remain owned by that experiment.

| Condition | Ownership |
|-----------|-----------|
| Created during experiment | Experiment |
| Validates specific hypothesis | Experiment |
| Has broader applicability | Experiment → Testing |

**Rationale**: Experiment tests validate hypotheses; broader applicability requires promotion review.

### Rule T-OWN-004: Promotion Requirement

**Rule**: Reusable test assets must be approved by Testing before promotion.

**Promotion Process**:
1. Asset creator submits promotion request
2. Testing reviews for reusability, stability, documentation
3. Testing approves or rejects with feedback
4. If approved, Testing adopts ownership

**Rationale**: Prevents unqualified assets from entering Testing infrastructure.

---

## 2. Lifecycle Rules

### Rule T-LIF-001: Asset Lifecycle

**Rule**: All Testing assets follow a defined lifecycle.

```
Created → Validated → Active → Maintained → Retired
```

| Phase | Description |
|-------|-------------|
| Created | Asset is created and documented |
| Validated | Asset passes Testing validation |
| Active | Asset is available for use |
| Maintained | Asset receives updates and fixes |
| Retired | Asset is deprecated and removed |

### Rule T-LIF-002: Creation Authority

**Rule**: New Testing assets can be created by:

| Creator | Authority |
|---------|-----------|
| Testing maintainer | Full authority |
| Execution Agent | With Testing review |
| Investigation | With Testing approval |
| Experiment | With Testing approval |

**Rationale**: Prevents proliferation of duplicate or incompatible assets.

### Rule T-LIF-003: Retirement Authority

**Rule**: Testing assets can only be retired with governance approval.

**Retirement Process**:
1. Testing identifies asset for retirement
2. Testing documents retirement rationale
3. Governance review approves/rejects
4. If approved, asset is deprecated with migration path

**Rationale**: Prevents breaking changes without proper notice.

---

## 3. Access Rules

### Rule T-ACC-001: Asset Access

**Rule**: All Testing assets are available to all authorized users.

| User Type | Access Level |
|-----------|--------------|
| Execution Agents | Read, Use |
| Investigation | Read, Use, Request |
| Experiment | Read, Use, Request |
| Testing | Full (Read, Write, Maintain, Retire) |

**Rationale**: Testing assets exist to be reused; access should not be restricted.

### Rule T-ACC-002: Asset Contribution

**Rule**: Contributions to Testing assets require Testing review.

| Contribution Type | Review Required |
|-------------------|-----------------|
| Bug fix | Yes (Testing review) |
| Documentation | Yes (Testing review) |
| New asset | Yes (Testing approval) |
| Breaking change | Yes (Governance approval) |

**Rationale**: Ensures quality and compatibility of Testing infrastructure.

### Rule T-ACC-003: Asset Request

**Rule**: Users can request Testing assets for their activities.

**Request Process**:
1. User identifies needed asset
2. User submits request to Testing
3. Testing provides asset or alternative
4. User acknowledges receipt

**Rationale**: Testing should proactively provide assets to avoid duplication.

---

## 4. Naming Rules

### Rule T-NAM-001: Naming Convention

**Rule**: All Testing assets must follow the TEST-* naming convention.

| Asset Type | Prefix | Example |
|------------|--------|---------|
| Testing document | `TEST-` | `TEST-ASSET-001.md` |
| Mock device | `TEST-MOCK-` | `TEST-MOCK-MASTER-001/` |
| Simulator | `TEST-SIM-` | `TEST-SIM-DNP3-001/` |
| Fixture | `TEST-FIX-` | `TEST-FIX-DNP3-CERTS/` |

**Rationale**: Consistent naming enables tooling and discovery.

### Rule T-NAM-002: Directory Placement

**Rule**: Testing assets must be placed in appropriate directories.

| Asset Type | Location |
|------------|----------|
| Documents | `laboratory/testing/` |
| Mock specifications | `laboratory/testing/mocks/` |
| Simulator specifications | `laboratory/testing/simulators/` |
| Fixtures | `laboratory/testing/fixtures/` |
| Implementation | Respective directory (test/, cmd/, etc.) |

**Rationale**: Logical organization aids discoverability.

---

## 5. Quality Rules

### Rule T-QLT-001: Documentation Requirement

**Rule**: All Testing assets must include documentation.

| Asset Type | Documentation Required |
|------------|----------------------|
| Mock device | Purpose, usage, examples |
| Simulator | Purpose, interface, limitations |
| Fixture | Purpose, format, maintenance |
| Script | Purpose, usage, dependencies |

**Rationale**: Documentation enables reuse and reduces support burden.

### Rule T-QLT-002: Test Coverage

**Rule**: Testing infrastructure assets should have test coverage.

| Asset Type | Test Coverage |
|------------|---------------|
| Mock devices | Unit tests for behavior |
| Simulators | Integration tests for protocol |
| Fixtures | Validation of data format |
| Scripts | Functional tests |

**Rationale**: Tests ensure assets work as documented.

### Rule T-QLT-003: Version Compatibility

**Rule**: Testing assets should document compatibility.

| Information | Required |
|-------------|----------|
| DNP3 version | Yes (if applicable) |
| Go version | Yes |
| External dependencies | Yes |
| Breaking changes | Yes (if any) |

**Rationale**: Compatibility information prevents integration issues.

---

## 6. RACI Matrix

| Activity | Testing | Investigation | Experiment | Governance |
|----------|---------|----------------|-------------|------------|
| Create Testing asset | R, A | C | C | I |
| Maintain Testing asset | R, A | I | I | I |
| Request Testing asset | I | R | R | I |
| Provide Testing asset | R, A | I | I | I |
| Promote asset to Testing | C | R | R | A |
| Retire Testing asset | R | C | C | A |
| Review Testing changes | A | C | C | I |

**Legend**:
- **R** = Responsible (performs the work)
- **A** = Accountable (final decision maker)
- **C** = Consulted (provides input)
- **I** = Informed (kept updated)

---

## 7. Enforcement

### Compliance

All laboratory activities must comply with these governance rules. Non-compliance may result in:

| Violation | Consequence |
|-----------|-------------|
| Missing naming prefix | Warning, then correction required |
| Unauthorized creation | Review, potential removal |
| Unauthorized promotion | Rejection, rework required |
| Retirement without approval | Restoration required |

### Exceptions

Exceptions to these rules require:

1. Written justification
2. Testing review
3. Governance approval
4. Documented rationale

---

## 8. Related Documents

| Document | Purpose |
|---------|---------|
| KDE-INV-047 | Testing capability investigation |
| GOV-HIERARCHY-001 | Authority hierarchy |
| GOV-NAMING-001 | Naming conventions |
| laboratory/testing/README.md | Testing capability overview |

---

**Status**: ENFORCED
**Review Date**: Upon any governance incident or KDE-INV-047 follow-up

---

*Generated per KDE-INV-047 recommendation*
*Approved: 2026-07-25*
