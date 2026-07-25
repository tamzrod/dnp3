# KDE Runtime Concepts

**Domain**: KDE Methodology
**Version**: 1.0.0
**Date**: 2026-07-24

---

## Overview

This directory contains documentation on KDE runtime concepts. These are technical concepts that explain how the Knowledge Discovery Engine operates.

## Key Concepts

### Investigation

A structured analysis of a problem or question, resulting in:
- **SPEC.md**: Full investigation report
- **CONCLUSION.md**: Key conclusions
- **ANALYSIS.md**: Supporting analysis
- **LEGACY/**: Previous versions (if any)

### Decision Record

A formal record of a technology or architectural decision:
- **TDR**: Technology Decision Record
- Contains rationale and evidence
- References supporting investigations

### Experiment

A hypothesis validation test:
- **DNP3-EXP-XXX**: Experiment identifier
- Documented methodology
- Reproducible results

### Evidence

Supporting data for investigations and decisions:
- Raw data and analysis
- Test results
- Benchmark data

---

## Runtime Structure

```
.kde/
├── bootstrap/          # Bootstrap configuration
├── runtime/            # Core runtime system
├── engines/            # Investigation engines
├── experts/            # Domain expert knowledge
├── knowledge/          # Engineering knowledge
├── governance/         # Governance policies
├── seeds/             # Seed knowledge
├── commands/           # System commands
├── capabilities/       # System capabilities
├── templates/          # Artifact templates
└── verification/       # Verification system
```

---

## Documentation Architecture

The KDE separates runtime assets from human documentation:

| Directory | Type | Purpose |
|-----------|------|---------|
| `.kde/` | Runtime | Framework consumed by KDE |
| `docs/kde/` | Human | Methodology understanding |
| `laboratory/` | Evidence | Engineering records |

---

*For DNP3 Library*
