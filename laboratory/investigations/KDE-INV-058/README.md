---
id: KDE-INV-058
type: investigation
title: "Safe KDE Runtime Upgrade: tamzrod/kde Merge Strategy"
authority: "KDE Runtime (DNP3 Library)"
status: IN_PROGRESS
created: "2026-07-27"
execution_agent: "OpenHands Agent"
engine: KDE-ENGINE-004 (Delta)
---

# Safe KDE Runtime Upgrade: tamzrod/kde Merge Strategy

**Investigation ID**: KDE-INV-058
**Engine**: KDE-ENGINE-004 (Delta)
**Title**: Safe KDE Runtime Upgrade: tamzrod/kde Merge Strategy
**Status**: IN_PROGRESS
**Date**: 2026-07-27
**Authority**: KDE Runtime (DNP3 Library)

---

## Executive Summary

This investigation analyzes the feasibility and strategy for safely upgrading the KDE Runtime Environment (`.kde/`) from the upstream repository [tamzrod/kde](https://github.com/tamzrod/kde). The study focuses on identifying merge-safe components (bootstrap, runtime, skills) while preserving project-specific customizations (experts, laboratory data) to ensure minimal disruption to ongoing engineering work.

**Key Finding**: The tamzrod/kde repository offers an evolved KDE architecture with enhanced governance documentation, additional experts (GIS, SLD), and refined SOPs. A safe three-phase merge strategy is recommended, prioritizing `.kde/bootstrap/` and `.kde/runtime/skills/` as primary upgrade targets.

---

## Research Questions

| ID | Question | Finding |
|----|----------|---------|
| RQ1 | What components in tamzrod/kde differ from current KDE? | **Major differences**: Experts (GIS/SLD), Knowledge taxonomy, Governance SOPs |
| RQ2 | Which components are safe to merge without conflicts? | **Safe**: bootstrap/, runtime/skills/, templates/, commands/, capabilities/, verification/ |
| RQ3 | Which components require preservation of local customizations? | **Preserve**: experts/ (dnp3-*), laboratory/, knowledge/ (project-specific) |
| RQ4 | What is the recommended merge strategy? | **Three-phase approach**: Phase 1 (bootstrap+runtime), Phase 2 (engines+seeds+governance), Phase 3 (expert merging) |

---

## Evidence

### Evidence E1: Repository Structure Comparison

**Type**: Direct
**Source**: GitHub API analysis of tamzrod/kde vs current `.kde/`
**Relevance**: Identifies structural differences for merge planning

```
Current KDE (.kde/):
├── bootstrap/         [COMPATIBLE]
├── capabilities/       [COMPATIBLE]
├── commands/           [COMPATIBLE]
├── engines/            [NEEDS COMPARISON]
├── experts/            [PRESERVE: dnp3-protocol, dnp3-security, dnp3-testing]
├── governance/         [MERGE CANDIDATE]
├── knowledge/           [PRESERVE: project-specific]
├── laboratory/          [PRESERVE: all data]
├── requirements.txt     [COMPATIBLE]
├── runtime/            [MERGE CANDIDATE]
├── seeds/              [NEEDS COMPARISON]
├── templates/           [COMPATIBLE]
└── verification/        [COMPATIBLE]

tamzrod/kde (.kde/):
├── README.md           [NEW: Contains better documentation]
├── bootstrap/          [COMPATIBLE - similar structure]
├── capabilities/       [COMPATIBLE]
├── commands/           [COMPATIBLE]
├── runtime/            [CONTAINS: aliases/, preflight.py]
└── verification/       [COMPATIBLE]

tamzrod/kde (root level):
├── engines/            [NEEDS COMPARISON - same as .kde/engines/]
├── experts/            [CONTAINS: gis/, sld/, kde-governance/]
├── governance/         [EXTENSIVE: 26+ SOPs vs current 6]
├── knowledge/          [EXTENSIVE: bus-voltage/, foundation/, etc.]
├── laboratory/         [RESEARCH DATA - do not merge]
├── runtime/            [SKILLS: loader.py, registry.json]
└── seeds/              [NEEDS COMPARISON]
```

---

### Evidence E2: Skills Registry Comparison

**Type**: Direct
**Source**: `registry.json` files from both repositories
**Relevance**: Confirms skill compatibility

Both repositories contain **8 identical skills**:
1. skill-investigation-planning
2. skill-experiment-design
3. skill-knowledge-retrieval
4. skill-evidence-collection
5. skill-decision-attribution
6. skill-artifact-traceability
7. skill-governance-review
8. skill-frontend-design

**Finding**: Skills are version-locked (v1.0.0) and synchronized. No merge conflicts expected.

---

### Evidence E3: Experts Directory Comparison

**Type**: Direct
**Source**: Directory listing of `experts/` in both repositories
**Relevance**: Identifies domain-specific customizations to preserve

| Current KDE Experts | tamzrod/kde Experts | Merge Action |
|---------------------|---------------------|--------------|
| dnp3-protocol/ | - | **PRESERVE** (project-specific) |
| dnp3-security/ | - | **PRESERVE** (project-specific) |
| dnp3-testing/ | - | **PRESERVE** (project-specific) |
| kde-governance/ | kde-governance/ | **MERGE** (add missing files) |
| registry/ | registry/ | **MERGE** (combine registries) |
| - | gis/ | **ADD** (new domain) |
| - | sld/ | **ADD** (new domain) |

---

### Evidence E4: Governance Documentation Comparison

**Type**: Document
**Source**: File listings of `governance/` directories
**Relevance**: tamzrod/kde has significantly more governance SOPs

**Current KDE Governance (6 files)**:
- AUTHORITY-DEFINITIONS.md
- DEP-001.md
- ENV-001.md
- GOVERNANCE-HIERARCHY.md
- NAMING-CONVENTIONS.md
- README.md

**tamzrod/kde Governance (26+ files)** includes:
- ARCHIVE-SOP.md
- ARTIFACT-PROTECTION.md
- ENGINE-VERSIONING.md
- FUTURE-EXPERIMENT-GUIDELINES.md
- INVESTIGATION-CLOSURE-SOP.md
- KDE-GOVERNANCE-DEPENDENCY-TRACKING.md
- KDE-GOVERNANCE-META-VALIDATION.md
- KDE-GOVERNANCE-STATE-VOCABULARY.md
- LABORATORY-SOP.md
- LESSONS-LEARNED-SOP.md
- NUMBERING-INVESTIGATION.md
- REC-001-APPROVAL.md
- SOP-COMPLEXITY-BUDGET.md
- STATE-MACHINE.md
- TIMESTAMP-STANDARD.md
- VERSION.md
- +promotions/, runtime/ subdirectories

**Finding**: tamzrod/kde governance is more mature. Consider adopting new SOPs selectively.

---

### Evidence E5: Runtime System Comparison

**Type**: Direct
**Source**: Runtime Python modules in both repositories
**Relevance**: Identifies new capabilities in tamzrod/kde

| Component | Current KDE | tamzrod/kde | Action |
|-----------|-------------|--------------|--------|
| attribution.py | ✓ | ✓ | Compare for updates |
| catalog.json | ✓ | ✓ | Compare for updates |
| ecu/ | ✓ | ✓ | Sync |
| instrumentation.py | ✓ | ✓ | Compare for updates |
| preflight.py | - | ✓ | **ADD** (new capability) |
| retrieval.py | ✓ | ✓ | Compare for updates |
| runtime.py | ✓ | ✓ | Compare for updates |
| skills/ | ✓ | ✓ | Sync registries |
| sop005.py | ✓ | ✓ | Compare for updates |
| validators/ | ✓ | ✓ | Sync |
| **aliases/** | - | ✓ | **ADD** (new capability) |
| **orchestrator/** | ✓ | ✓ | Compare for updates |

**New in tamzrod/kde**:
- `runtime/aliases/` - Command aliasing system
- `preflight.py` - Enhanced preflight checks

---

## Findings

### Finding F1: Structural Compatibility

**Classification**: Architecture
**Evidence**: E1
**Confidence**: HIGH

The KDE Runtime uses a consistent directory structure across repositories. The `.kde/` directory in the current project contains all expected components (bootstrap, runtime, engines, etc.), making it structurally compatible with tamzrod/kde for merge operations.

### Finding F2: Skills System Synchronized

**Classification**: Configuration
**Evidence**: E2
**Confidence**: HIGH

Both repositories have identical skills registries (8 skills, v1.0.0). The skills loader and registry files can be synchronized without conflict. The tamzrod/kde `runtime/skills/` at root level contains the same loader.py and registry.json as `.kde/runtime/skills/`.

### Finding F3: Experts Require Selective Merge

**Classification**: Strategy
**Evidence**: E3
**Confidence**: HIGH

Project-specific experts (`dnp3-protocol`, `dnp3-security`, `dnp3-testing`) must be preserved. The tamzrod/kde experts (`gis`, `sld`) are new domain experts that can be added. The `kde-governance` expert should be merged to incorporate any new governance knowledge.

### Finding F4: Governance SOPs Need Selective Adoption

**Classification**: Governance
**Evidence**: E4
**Confidence**: MEDIUM

The tamzrod/kde has significantly more mature governance documentation. New SOPs like INVESTIGATION-CLOSURE-SOP.md, LABORATORY-SOP.md, and SOP-COMPLEXITY-BUDGET.md could enhance the current KDE governance. However, some SOPs may conflict with project-specific policies.

### Finding F5: Runtime Enhancements Available

**Classification**: Enhancement
**Evidence**: E5
**Confidence**: MEDIUM

The tamzrod/kde runtime includes:
1. `preflight.py` - Standalone preflight verification script
2. `aliases/` - Command aliasing for improved UX

These can be safely added to enhance the current runtime.

---

## Recommendations

| Recommendation | Priority | Owner | Rationale |
|----------------|----------|-------|------------|
| REC-1: Adopt new `preflight.py` from tamzrod/kde | HIGH | Agent | Enhances runtime verification |
| REC-2: Add `runtime/aliases/` from tamzrod/kde | MEDIUM | Agent | Improves command UX |
| REC-3: Merge `experts/kde-governance/` from tamzrod/kde | HIGH | Agent | Updates governance knowledge |
| REC-4: Add new experts (`gis`, `sld`) from tamzrod/kde | LOW | Agent | New domains, optional |
| REC-5: Selectively adopt new governance SOPs | MEDIUM | Agent | Enhance governance maturity |
| REC-6: Do NOT merge `laboratory/` from tamzrod/kde | HIGH | Agent | Contains project-specific research |

---

## Merge Strategy: Three-Phase Approach

### Phase 1: Bootstrap & Runtime (Low Risk)

```
Components:
├── .kde/bootstrap/    → Sync with tamzrod/kde
├── .kde/runtime/      → Add preflight.py, aliases/
└── .kde/skills/       → Sync registry.json
```

**Actions**:
1. Copy `runtime/preflight.py` from tamzrod/kde
2. Copy `runtime/aliases/` directory
3. Compare and sync `runtime/*.py` files
4. Verify skills registry synchronization

**Validation**: Run `python3 .kde/bootstrap/gates.py --quick`

---

### Phase 2: Engines, Seeds, Governance (Medium Risk)

```
Components:
├── .kde/engines/       → Compare engine implementations
├── .kde/seeds/         → Compare seed content
├── .kde/governance/    → Selective SOP adoption
└── .kde/templates/    → Sync if updated
```

**Actions**:
1. Compare engines directory with `git diff`
2. Merge any new engine implementations
3. Compare seeds for new content
4. Selectively add new governance SOPs:
   - ARCHIVE-SOP.md
   - INVESTIGATION-CLOSURE-SOP.md
   - LABORATORY-SOP.md
   - SOP-COMPLEXITY-BUDGET.md

**Validation**: Run KDE preflight checks

---

### Phase 3: Experts & Knowledge (High Risk)

```
Components:
├── .kde/experts/       → Selective merge
├── .kde/knowledge/     → Selective merge
└── .kde/laboratory/    → DO NOT MERGE
```

**Actions**:
1. Preserve existing `dnp3-*` experts
2. Merge `kde-governance` expert updates
3. Optionally add `gis` and `sld` experts
4. Compare `knowledge/` directories
5. **DO NOT touch laboratory/ directory**

**Validation**: Expert registry integrity check

---

## Related Artifacts

| Artifact | Type | Relationship |
|----------|------|--------------|
| KDE-INV-056 | Investigation | Prior KDE structure analysis |
| tamzrod/kde | External | Source repository |
| .kde/README.md | Documentation | Current KDE documentation |

---

## Conclusion

This investigation confirms that a safe KDE upgrade from tamzrod/kde is feasible using a three-phase merge strategy. The recommended approach prioritizes low-risk components (bootstrap, runtime skills) while preserving project-specific customizations (DNP3 experts, laboratory data). The tamzrod/kde repository offers valuable enhancements in governance documentation and runtime capabilities that can incrementally improve the current KDE implementation.

**Status**: Investigation ready for review
**Human Review Required**: Yes

---

**Investigation Status**: IN_PROGRESS
**Human Review Required**: Yes
