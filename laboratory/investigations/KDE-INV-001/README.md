# KDE-INV-001: Laboratory Cleanup and Bootstrap Template Preparation

**Status**: IN_PROGRESS
**Date**: 2026-07-25
**Type**: Evidence-Based Investigation
**Author**: OpenHands Agent

---

## Executive Summary

This investigation reviews the repository structure to classify all artifacts and determine the requirements for creating a clean KDE Bootstrap Template suitable for initializing future repositories.

### Key Findings

| Category | Classification | Count | Action |
|----------|---------------|-------|--------|
| **Bootstrap Core** | Required by every KDE project | 15 | Retain in template |
| **Generic KDE** | Useful across KDE projects | 8 | Retain in template |
| **Project-Specific** | Applicable only to current project | 12 | Remove from template |
| **Historical** | Preserve for reference | ~180 | Archive only |

### Recommendation

Create a `bootstrap-template` branch that:
1. Retains only Bootstrap Core and Generic KDE artifacts
2. Removes all Trexa and DNP3-specific identity
3. Creates a clean, generic KDE Bootstrap structure
4. Preserves full git history for reference

---

## Complete Artifact Inventory

### 1. `.kde/` Directory (Runtime Framework)

| Artifact | Path | Classification | Rationale |
|----------|------|---------------|-----------|
| Bootstrap Config | `.kde/bootstrap/config.yaml` | **Bootstrap Core** | Runtime initialization |
| Bootstrap README | `.kde/bootstrap/README.md` | **Bootstrap Core** | Bootstrap documentation |
| Bootstrap Requirements | `.kde/bootstrap/requirements.json` | **Bootstrap Core** | Bootstrap dependencies |
| Runtime State | `.kde/runtime/state.json` | **Bootstrap Core** | Runtime state tracking |
| Capabilities README | `.kde/capabilities/README.md` | **Bootstrap Core** | Capability module |
| Commands README | `.kde/commands/README.md` | **Bootstrap Core** | Command module |
| Engines README | `.kde/engines/README.md` | **Bootstrap Core** | Engine module |
| Experts README | `.kde/experts/README.md` | **Bootstrap Core** | Expert module |
| **Governance README** | `.kde/governance/README.md` | **Bootstrap Core** | Governance module |
| **Naming Conventions** | `.kde/governance/NAMING-CONVENTIONS.md` | **Project-Specific** | Contains TREXA-INV- references |
| Knowledge README | `.kde/knowledge/README.md` | **Bootstrap Core** | Knowledge module |
| Seeds README | `.kde/seeds/README.md` | **Bootstrap Core** | Seeds module |
| Templates README | `.kde/templates/README.md` | **Bootstrap Core** | Template module |
| **IMP Template** | `.kde/templates/IMP.md` | **Project-Specific** | References TREXA-INV-021 |
| Verification README | `.kde/verification/README.md` | **Bootstrap Core** | Verification module |

### 2. `docs/` Directory (Human-Readable Documentation)

| Artifact | Path | Classification | Rationale |
|----------|------|---------------|-----------|
| **docs/README.md** | `docs/README.md` | **Project-Specific** | References "Trexa Documentation" |
| **application/README.md** | `docs/application/README.md` | **Project-Specific** | Entirely about Trexa visual platform |
| application/api/ | `docs/application/api/` | **Project-Specific** | Trexa API docs |
| application/architecture/ | `docs/application/architecture/` | **Project-Specific** | Trexa architecture |
| application/getting-started/ | `docs/application/getting-started/` | **Project-Specific** | Trexa getting started |
| application/guides/ | `docs/application/guides/` | **Project-Specific** | Trexa user guides |
| application/reference/ | `docs/application/reference/` | **Project-Specific** | Trexa reference |
| application/roadmap/ | `docs/application/roadmap/` | **Project-Specific** | Trexa roadmap |
| kde/README.md | `docs/kde/README.md` | **Generic KDE** | KDE methodology overview |
| kde/governance/ | `docs/kde/governance/` | **Generic KDE** | KDE governance docs |
| kde/history/ | `docs/kde/history/` | **Generic KDE** | KDE history |
| kde/methodology/ | `docs/kde/methodology/` | **Generic KDE** | KDE methodology |
| kde/principles/ | `docs/kde/principles/` | **Generic KDE** | KDE engineering principles |
| kde/reviews/ | `docs/kde/reviews/` | **Generic KDE** | KDE reviews |
| kde/runtime-concepts/ | `docs/kde/runtime-concepts/` | **Generic KDE** | KDE runtime concepts |

### 3. `laboratory/` Directory (Engineering Evidence)

| Artifact | Path | Classification | Rationale |
|----------|------|---------------|-----------|
| Laboratory README | `laboratory/README.md` | **Bootstrap Core** | Laboratory entry point |
| decisions/ | `laboratory/decisions/` | **Bootstrap Core** | Decision records directory |
| decisions/README.md | `laboratory/decisions/README.md` | **Bootstrap Core** | Decision documentation |
| evidence/ | `laboratory/evidence/` | **Bootstrap Core** | Evidence directory |
| evidence/README.md | `laboratory/evidence/README.md` | **Bootstrap Core** | Evidence documentation |
| experiments/ | `laboratory/experiments/` | **Bootstrap Core** | Experiments directory |
| experiments/README.md | `laboratory/experiments/README.md` | **Bootstrap Core** | Experiment documentation |
| implementations/ | `laboratory/implementations/` | **Bootstrap Core** | Implementations directory |
| implementations/README.md | `laboratory/implementations/README.md` | **Bootstrap Core** | Implementation docs |
| planning/ | `laboratory/planning/` | **Bootstrap Core** | Planning directory |
| planning/README.md | `laboratory/planning/README.md` | **Bootstrap Core** | Planning documentation |
| reviews/ | `laboratory/reviews/` | **Bootstrap Core** | Reviews directory |
| reviews/README.md | `laboratory/reviews/README.md` | **Bootstrap Core** | Reviews documentation |
| investigations/ | `laboratory/investigations/` | **Bootstrap Core** | Investigations directory |
| **KDE-INV-042/43/44/45** | `laboratory/investigations/KDE-INV-*` | **Generic KDE** | Generic KDE framework investigations |
| archive/trexa/ | `laboratory/archive/trexa/` | **Historical** | Archived Trexa artifacts (reference only) |

### 4. Root-Level Files

| Artifact | Path | Classification | Rationale |
|----------|------|---------------|-----------|
| **README.md** | `README.md` | **Project-Specific** | DNP3 Influx Data Logger project README |
| **config.example.yaml** | `config.example.yaml` | **Project-Specific** | DNP3 configuration example |
| LICENSE | `LICENSE` | **Bootstrap Core** | MIT License |
| .gitignore | `.gitignore` | **Bootstrap Core** | Git ignore patterns |
| go.mod | `go.mod` | **Project-Specific** | Go module definition |
| cmd/ | `cmd/` | **Project-Specific** | DNP3 application entry point |
| internal/ | `internal/` | **Project-Specific** | DNP3 internal packages |

---

## Detailed Classification Evidence

### Bootstrap Core Artifacts (15)

These artifacts form the core KDE runtime framework and must be retained:

1. **`.kde/bootstrap/`** - Bootstrap initialization
2. **`.kde/runtime/`** - Runtime state tracking  
3. **`.kde/capabilities/`** - Capability definitions
4. **`.kde/commands/`** - Command definitions
5. **`.kde/engines/`** - Engine definitions
6. **`.kde/experts/`** - Expert definitions
7. **`.kde/governance/` (README only)** - Governance documentation
8. **`.kde/knowledge/`** - Knowledge definitions
9. **`.kde/seeds/`** - Seed definitions
10. **`.kde/templates/` (README only)** - Template documentation
11. **`.kde/verification/`** - Verification definitions
12. **`laboratory/` directories** - All laboratory structure
13. **`LICENSE`** - MIT License
14. **`.gitignore`** - Git patterns

### Generic KDE Artifacts (8)

These artifacts provide generic KDE methodology and should be retained:

1. **`docs/kde/`** - KDE methodology documentation
2. **`laboratory/investigations/KDE-INV-042/`** - Bootstrap Compliance
3. **`laboratory/investigations/KDE-INV-043/`** - Knowledge Promotion
4. **`laboratory/investigations/KDE-INV-044/`** - Decision Classification
5. **`laboratory/investigations/KDE-INV-045/`** - Laboratory Cleanup

### Project-Specific Artifacts (12)

These artifacts contain Trexa or DNP3-specific identity and must be removed:

1. **`README.md`** - Project-specific project documentation
2. **`docs/README.md`** - References "Trexa Documentation"
3. **`docs/application/`** - Entire Trexa application documentation
4. **`config.example.yaml`** - DNP3-specific configuration
5. **`go.mod`** - Go module with DNP3 references
6. **`cmd/`** - DNP3 application entry point
7. **`internal/`** - DNP3 internal packages
8. **`.kde/governance/NAMING-CONVENTIONS.md`** - Contains TREXA-INV- references
9. **`.kde/templates/IMP.md`** - References TREXA-INV-021

### Historical Artifacts (~180)

Archived in `laboratory/archive/trexa/` - preserved for reference but not part of template.

---

## Proposed Bootstrap Template Directory Structure

```
kde-bootstrap/
├── .kde/                      # KDE Runtime Framework
│   ├── bootstrap/              # Bootstrap initialization
│   │   ├── README.md
│   │   ├── config.yaml        # Generic config (no project name)
│   │   └── requirements.json
│   ├── runtime/               # Runtime state
│   │   └── state.json
│   ├── capabilities/          # Capability module
│   │   └── README.md
│   ├── commands/              # Command module
│   │   └── README.md
│   ├── engines/               # Engine module
│   │   └── README.md
│   ├── experts/                # Expert module
│   │   └── README.md
│   ├── governance/            # Governance module
│   │   ├── README.md
│   │   └── NAMING-CONVENTIONS.md  # Generic naming conventions
│   ├── knowledge/             # Knowledge module
│   │   └── README.md
│   ├── seeds/                 # Seeds module
│   │   └── README.md
│   ├── templates/              # Templates module
│   │   ├── README.md
│   │   ├── INV.md             # Investigation template
│   │   ├── EXP.md             # Experiment template
│   │   ├── TDR.md             # Decision template
│   │   └── IMP.md             # Implementation template
│   └── verification/           # Verification module
│       └── README.md
│
├── docs/                      # Human-Readable Documentation
│   └── kde/                   # KDE Methodology (generic)
│       ├── README.md
│       ├── methodology/
│       ├── principles/
│       ├── governance/
│       ├── runtime-concepts/
│       ├── reviews/
│       └── history/
│
├── laboratory/                # Engineering Laboratory
│   ├── README.md
│   ├── decisions/             # Technology Decision Records
│   │   └── README.md
│   ├── evidence/              # Evidence artifacts
│   │   └── README.md
│   ├── experiments/           # Laboratory experiments
│   │   └── README.md
│   ├── implementations/        # Implementation specifications
│   │   └── README.md
│   ├── investigations/        # Investigations
│   │   └── README.md
│   ├── planning/              # Planning documents
│   │   └── README.md
│   └── reviews/               # Review documents
│       └── README.md
│
├── .gitignore                 # Git ignore patterns
└── LICENSE                    # MIT License
```

---

## Required Changes for Bootstrap Template

### 1. Remove Project-Specific Artifacts

```bash
# Remove DNP3-specific artifacts
rm README.md
rm -rf docs/application/
rm config.example.yaml
rm go.mod
rm -rf cmd/
rm -rf internal/
```

### 2. Update .kde/ Artifacts

```bash
# Update .kde/bootstrap/config.yaml - Remove "DNP3 Influx Data Logger"
# Update .kde/governance/NAMING-CONVENTIONS.md - Remove TREXA-INV- references
# Update .kde/templates/IMP.md - Remove TREXA-INV- references
```

### 3. Update docs/ Artifacts

```bash
# Create generic docs/README.md - KDE Bootstrap Documentation
# Update docs/kde/ - Already generic, may need minor updates
```

### 4. Update laboratory/ Artifacts

```bash
# Update laboratory/README.md - Remove project references
# Remove KDE-INV-042/43/44/45 - Move to Generic KDE category in template
```

### 5. Remove Historical Archive

```bash
# Archive is reference only, not part of Bootstrap Template
rm -rf laboratory/archive/
```

---

## Risk Assessment

| Action | Risk | Mitigation |
|--------|------|------------|
| Remove project-specific docs | Loss of documentation structure | Preserve KDE docs structure |
| Update naming conventions | Naming conflicts | Use generic PROJECT- prefix |
| Remove archive directory | Loss of historical reference | Archive remains in main branch |
| Update templates | Template breakage | Test templates after update |

---

## Verification Checklist

Before creating bootstrap-template branch:

- [ ] All project-specific artifacts identified
- [ ] All artifacts classified with evidence
- [ ] Generic KDE artifacts verified
- [ ] Proposed structure reviewed
- [ ] Risk assessment completed
- [ ] Human approval obtained

---

*Investigation initiated: 2026-07-25*
