# KDE-INV-001: Laboratory Cleanup and Bootstrap Template Preparation - Conclusions

**ID**: KDE-INV-001
**Status**: IN_PROGRESS
**Date**: 2026-07-25

---

## Research Question Answers

### Q1: Which artifacts in `.kde/` are Bootstrap Core vs. Project-Specific?

**Answer**: 13 Bootstrap Core, 2 Project-Specific

| Classification | Artifacts |
|----------------|-----------|
| **Bootstrap Core** | bootstrap/, runtime/, capabilities/, commands/, engines/, experts/, governance/README.md, knowledge/, seeds/, templates/README.md, verification/ |
| **Project-Specific** | governance/NAMING-CONVENTIONS.md (contains TREXA-INV-), templates/IMP.md (references TREXA-INV-021) |

### Q2: Which artifacts in `docs/` are Generic KDE vs. Project-Specific?

**Answer**: 8 Generic KDE, 9 Project-Specific

| Classification | Artifacts |
|----------------|-----------|
| **Generic KDE** | kde/README.md, kde/governance/, kde/history/, kde/methodology/, kde/principles/, kde/reviews/, kde/runtime-concepts/ |
| **Project-Specific** | docs/README.md ("Trexa Documentation"), application/ entire directory |

### Q3: Which artifacts in `laboratory/` are Generic KDE vs. Project-Specific?

**Answer**: 14 Bootstrap Core, 4 Generic KDE, ~180 Historical

| Classification | Artifacts |
|----------------|-----------|
| **Bootstrap Core** | All directory READMEs and structure |
| **Generic KDE** | KDE-INV-042, KDE-INV-043, KDE-INV-044, KDE-INV-045 |
| **Historical** | archive/trexa/ (all ~180 archived artifacts) |

### Q4: Which root-level files are Bootstrap Core vs. Project-Specific?

**Answer**: 2 Bootstrap Core, 5 Project-Specific

| Classification | Artifacts |
|----------------|-----------|
| **Bootstrap Core** | LICENSE, .gitignore |
| **Project-Specific** | README.md, config.example.yaml, go.mod, cmd/, internal/ |

### Q5: What is the proposed Bootstrap Template directory structure?

**Answer**: See README.md Section "Proposed Bootstrap Template Directory Structure"

The template consists of:
- `.kde/` - Runtime framework (13 modules)
- `docs/kde/` - Generic KDE methodology
- `laboratory/` - Laboratory structure with empty directories
- `.gitignore`, `LICENSE`

### Q6: What changes are required to create a clean Bootstrap Template?

**Answer**:

1. **Remove 12 Project-Specific artifacts**:
   - README.md
   - docs/README.md
   - docs/application/ (entire directory)
   - config.example.yaml
   - go.mod
   - cmd/
   - internal/
   - .kde/governance/NAMING-CONVENTIONS.md
   - .kde/templates/IMP.md
   - laboratory/archive/

2. **Update 3 artifacts**:
   - .kde/bootstrap/config.yaml (remove project name)
   - .kde/governance/NAMING-CONVENTIONS.md (genericize)
   - .kde/templates/IMP.md (genericize)
   - docs/kde/README.md (update references)
   - laboratory/README.md (update references)

3. **Create 4 Generic KDE Investigations** in template:
   - KDE-INV-042 (Bootstrap Compliance)
   - KDE-INV-043 (Knowledge Promotion)
   - KDE-INV-044 (Decision Classification)
   - KDE-INV-045 (Laboratory Cleanup)

---

## Summary Statistics

| Category | Count | Percentage | Template Inclusion |
|----------|-------|------------|-------------------|
| **Bootstrap Core** | 15 | 8% | ✅ Yes |
| **Generic KDE** | 8 | 4% | ✅ Yes |
| **Project-Specific** | 12 | 7% | ❌ No |
| **Historical** | ~180 | 81% | ❌ No (archive) |

---

## Final Recommendations

### Retention List (23 artifacts for template)

1. **`.kde/bootstrap/`** - All 3 files
2. **`.kde/runtime/`** - state.json
3. **`.kde/capabilities/`** - README.md
4. **`.kde/commands/`** - README.md
5. **`.kde/engines/`** - README.md
6. **`.kde/experts/`** - README.md
7. **`.kde/governance/`** - README.md, NAMING-CONVENTIONS.md (updated)
8. **`.kde/knowledge/`** - README.md
9. **`.kde/seeds/`** - README.md
10. **`.kde/templates/`** - README.md, IMP.md (updated), + create INV.md, EXP.md, TDR.md
11. **`.kde/verification/`** - README.md
12. **`docs/kde/`** - All 7 subdirectories
13. **`laboratory/`** - All directory READMEs and structure
14. **`.gitignore`** - Standard Go/IDE patterns
15. **`LICENSE`** - MIT License

### Removal List (12 artifacts)

1. README.md
2. docs/README.md
3. docs/application/ (entire directory)
4. config.example.yaml
5. go.mod
6. cmd/ (entire directory)
7. internal/ (entire directory)
8. laboratory/archive/ (entire directory)
9. .kde/governance/NAMING-CONVENTIONS.md (replace with generic version)
10. .kde/templates/IMP.md (replace with generic version)

### Create in Template (7 new files)

1. docs/README.md - Generic KDE Bootstrap Documentation
2. docs/kde/README.md - Updated for generic KDE
3. .kde/templates/INV.md - Investigation template
4. .kde/templates/EXP.md - Experiment template
5. .kde/templates/TDR.md - Decision template
6. .kde/governance/NAMING-CONVENTIONS.md - Generic naming conventions
7. laboratory/README.md - Updated for generic KDE

---

## Implementation Plan

### Phase 1: Create bootstrap-template branch
```bash
git checkout -b bootstrap-template
```

### Phase 2: Remove Project-Specific Artifacts
```bash
rm README.md
rm -rf docs/application/
rm config.example.yaml
rm go.mod
rm -rf cmd/
rm -rf internal/
rm -rf laboratory/archive/
```

### Phase 3: Update Artifacts
```bash
# Update .kde/bootstrap/config.yaml
# Update .kde/governance/NAMING-CONVENTIONS.md
# Update .kde/templates/IMP.md
# Update docs/kde/README.md
# Update laboratory/README.md
# Create docs/README.md
```

### Phase 4: Create Template Files
```bash
# Create docs/README.md
# Create .kde/templates/INV.md
# Create .kde/templates/EXP.md
# Create .kde/templates/TDR.md
```

### Phase 5: Commit and Validate
```bash
git commit -m "Create KDE Bootstrap Template"
git checkout main
```

---

*Conclusions completed: 2026-07-25*
*Awaiting human approval before implementation*
