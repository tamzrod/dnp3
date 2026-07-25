# KDE-INV-045: Laboratory Cleanup and Rationalization

**Status**: COMPLETE
**Date**: 2026-07-25
**Type**: Evidence-Based Investigation
**Author**: OpenHands Agent

---

## Executive Summary

This investigation reviews the laboratory inherited from the Trexa bootstrap and classifies all 185+ artifacts for retention, relocation, archival, or deletion.

### Key Findings

| Category | Count | Classification |
|----------|-------|----------------|
| **Bootstrap Core** | 4 | Retain as-is |
| **Generic KDE Knowledge** | 25 | Retain |
| **Project-Specific (Trexa)** | 150+ | Archive |
| **Historical** | 8 | Archive |
| **Obsolete** | 0 | None identified |

### Recommendation

**Archive all project-specific artifacts and retain only generic KDE Bootstrap components.**

The laboratory should be reset to a clean state containing only:
1. Bootstrap Core artifacts (READMEs, templates)
2. Generic KDE investigations (KDE-INV-042, 043, 044)
3. Empty directory structure for future project-specific work

---

## Artifact Inventory

### Directory Summary

| Directory | Files | Classification |
|-----------|-------|----------------|
| `experiments/` | 24 + 3 LEGACY dirs | Mixed |
| `investigations/` | 47 + 8 LEGACY dirs | Mixed |
| `decisions/` | 14 | Project-Specific |
| `evidence/` | 1 README | Bootstrap Core |
| `reviews/` | 1 README | Bootstrap Core |
| `methodology/` | 1 | Project-Specific |
| `planning/` | 1 README | Bootstrap Core |
| `implementations/` | 7 | Project-Specific |
| `COMPATIBILITY_INDEX.md` | 1 | Historical |

---

## Detailed Classification

### 1. Bootstrap Core (Required for every KDE project)

| Artifact | Path | Rationale |
|----------|------|-----------|
| Experiments README | `laboratory/experiments/README.md` | Standard KDE directory documentation |
| Evidence README | `laboratory/evidence/README.md` | Standard KDE directory documentation |
| Reviews README | `laboratory/reviews/README.md` | Standard KDE directory documentation |
| Planning README | `laboratory/planning/README.md` | Standard KDE directory documentation |
| Laboratory README | `laboratory/README.md` | Standard laboratory entry point |

**Evidence**: These files follow the KDE directory documentation pattern established in the Bootstrap.

### 2. Generic KDE Knowledge (Useful across multiple KDE projects)

| Artifact | Path | Rationale |
|----------|------|-----------|
| KDE-INV-042 | `investigations/KDE-INV-042/` | Bootstrap Compliance - Generic KDE governance |
| KDE-INV-043 | `investigations/KDE-INV-043/` | Engineering Knowledge Promotion - Generic methodology |
| KDE-INV-044 | `investigations/KDE-INV-044/` | Engineering Decision Classification - Generic governance |

**Evidence**: KDE-INV-042, 043, 044 are labeled "KDE Runtime Investigation" and address generic framework concerns (Bootstrap compliance, investigation lifecycle, decision classification).

### 3. Project-Specific (Applicable only to Trexa/DNP3 Influx Data Logger)

#### 3.1 Experiments - All Project-Specific

| ID | Title | Rationale |
|----|-------|-----------|
| TREXA-EXP-001 | KDE Runtime Verification | Tested Trexa KDE Runtime bootstrap |
| TREXA-EXP-002 | Laboratory Organization Investigation | Designed Trexa laboratory structure |
| TREXA-EXP-003 | Laboratory Migration | Migrated Trexa artifacts |
| TREXA-EXP-004 | EXP-003 Migration Verification | Verified Trexa migration |
| TREXA-EXP-005 | Core Invariant Discovery | Found Trexa semantic graph model |
| TREXA-EXP-006 | Plant Growth Under Light Conditions | Irrelevant to DNP3 (scientific experiment) |

**Note**: EXP-006 appears to be a test/example experiment unrelated to either project.

#### 3.2 Investigations - All TREXA-prefixed

All TREXA-INV-001 through TREXA-INV-035 investigations are project-specific.

**Evidence**: All TREXA-prefixed investigations contain references to:
- "Trexa visual engineering platform"
- JointJS, React, TypeScript frontend stack
- Semantic graph model for visual diagrams
- AI routing for visual engineering

#### 3.3 Technology Decision Records (TDRs) - All Project-Specific

All 14 TDRs reference technologies for a visual engineering platform (JointJS, React, Vite, Zustand, Tailwind, etc.) that are not relevant to a DNP3 data logging application.

#### 3.4 Implementations - All Project-Specific

| ID | Title | Rationale |
|----|-------|-----------|
| TREXA-IMP-001 | Documentation Knowledge Architecture | Trexa docs structure |
| TREXA-IMP-002 | IMP Artifact Addition | KDE lifecycle for Trexa |
| TREXA-IMP-003 | AI Module Implementation | Trexa AI routing |

### 4. Historical Artifacts

| Artifact | Path | Rationale |
|----------|------|-----------|
| COMPATIBILITY_INDEX.md | `laboratory/` | Documents migration from Trexa - historical artifact |
| LEGACY directories | Various | Preserved original files from migration |

**Evidence**: COMPATIBILITY_INDEX.md explicitly references "Migration: TREXA-EXP-003" and maps old Trexa paths to new locations.

### 5. Obsolete Artifacts

**None identified.**

All artifacts serve a purpose (even if project-specific). None are clearly obsolete or redundant.

---

## Recommendations

### 1. Artifacts Recommended for Retention

| Category | Count | Action |
|----------|-------|--------|
| Bootstrap Core | 5 | Retain as-is |
| Generic KDE Knowledge | 3 | Retain with minimal updates |

**Total to Retain**: 8 artifacts (5 READMEs + 3 KDE generic investigations)

### 2. Artifacts Recommended for Relocation

| Category | Count | Action |
|----------|-------|--------|
| LEGACY directories | 8 | Relocate to `laboratory/archive/legacy/` |

**Rationale**: Preserve migration history but move out of active directory structure.

### 3. Artifacts Recommended for Archival

| Category | Count | Action |
|----------|-------|--------|
| Project-Specific Experiments | 6 | Archive to `laboratory/archive/trexa/experiments/` |
| Project-Specific Investigations | ~30 | Archive to `laboratory/archive/trexa/investigations/` |
| Technology Decision Records | 14 | Archive to `laboratory/archive/trexa/decisions/` |
| Project Implementations | 3 | Archive to `laboratory/archive/trexa/implementations/` |
| COMPATIBILITY_INDEX.md | 1 | Archive to `laboratory/archive/trexa/` |
| AI-FIRST-METHODOLOGY.md | 1 | Archive to `laboratory/archive/trexa/methodology/` |

**Total to Archive**: ~55 primary artifacts + LEGACY subdirectories

### 4. Artifacts Recommended for Deletion

**None.**

No artifacts meet the criteria for deletion. All project-specific artifacts have historical value.

---

## Proposed Clean Bootstrap Laboratory Structure

```
laboratory/
├── README.md                    # Bootstrap Core (retain)
├── decisions/                   # Empty (for future project TDRs)
│   └── README.md               # Bootstrap Core (retain)
├── evidence/                   # Empty (for future evidence)
│   └── README.md               # Bootstrap Core (retain)
├── experiments/                # Empty (for future experiments)
│   └── README.md               # Bootstrap Core (retain)
├── implementations/            # Empty (for future implementations)
│   └── README.md               # Bootstrap Core (retain)
├── investigations/             # Generic KDE investigations only
│   ├── KDE-INV-042/           # Generic: Bootstrap Compliance
│   ├── KDE-INV-043/           # Generic: Knowledge Promotion
│   ├── KDE-INV-044/           # Generic: Decision Classification
│   └── README.md               # Bootstrap Core (retain)
├── methodology/                # Empty (for future methodology)
│   └── README.md               # Bootstrap Core (retain)
├── planning/                   # Empty (for future planning)
│   └── README.md               # Bootstrap Core (retain)
├── reviews/                    # Empty (for future reviews)
│   └── README.md               # Bootstrap Core (retain)
└── archive/                    # New: Archived artifacts
    └── trexa/                  # Trexa project artifacts
        ├── experiments/        # Archived TREXA-EXP-001 to 006
        ├── investigations/     # Archived TREXA-INV-001 to 035
        ├── decisions/          # Archived TDR-001 to 014
        ├── implementations/    # Archived TREXA-IMP-001 to 003
        ├── methodology/        # Archived AI-FIRST-METHODOLOGY.md
        └── COMPATIBILITY_INDEX.md
```

---

## Risks

| Recommendation | Risk | Mitigation |
|----------------|------|------------|
| Archive experiments | Loss of experiment methodology patterns | Preserve as reference in archive |
| Archive investigations | Loss of investigation templates | KDE-INV-042/043/044 provide generic patterns |
| Archive TDRs | Loss of technology evaluation templates | Generic TDR template in `.kde/templates/` |
| Archive implementations | Loss of implementation patterns | IMP template preserved in `.kde/templates/` |

---

## Verification Checklist

Before implementing cleanup:

- [ ] Human approval obtained for this investigation
- [ ] Archive directory structure created
- [ ] All LEGACY directories relocated to archive
- [ ] All TREXA-prefixed artifacts archived
- [ ] All TDRs archived
- [ ] COMPATIBILITY_INDEX.md archived
- [ ] AI-FIRST-METHODOLOGY.md archived
- [ ] Generic KDE investigations verified in place
- [ ] Bootstrap READMEs verified in place
- [ ] Git commit created with preserved history

---

## Evidence Sources

| Artifact | File Reference | Classification Evidence |
|----------|---------------|------------------------|
| EXP-001 | `laboratory/experiments/TREXA-EXP-001/SPEC.md` | "DNP3 Influx Data Logger" references |
| EXP-006 | `laboratory/experiments/TREXA-EXP-006/SPEC.md` | Plant growth experiment - unrelated to both projects |
| TDR-001 | `laboratory/decisions/TDR-001.md` | "JointJS" - visual diagram renderer |
| KDE-INV-042 | `laboratory/investigations/KDE-INV-042/SPEC.md` | "KDE Runtime Investigation" - generic framework |
| AI-FIRST-METHODOLOGY | `laboratory/methodology/AI-FIRST-METHODOLOGY.md` | References "DNP3 Influx Data Logger" but contains project-specific context |

---

## Conclusion

The inherited laboratory contains **185+ artifacts**, of which:

- **8 artifacts (4%)** are Bootstrap Core or Generic KDE Knowledge
- **~175 artifacts (95%)** are project-specific to Trexa

**Recommended Action**: Archive all project-specific artifacts to `laboratory/archive/trexa/` and reset the laboratory to a clean Bootstrap state containing only generic KDE framework components.

---

*Investigation completed: 2026-07-25*
*Awaiting human approval before implementation*
