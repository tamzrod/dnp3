# Validation: Experiment Standards

**Seed ID**: SEED-002
**Section**: Validation (NEW)
**Addresses**: LESSON-008

---

## Purpose

Experiment standards ensure consistency, reproducibility, and comparability across all KDE experiments.

**Problem Addressed**: LESSON-008 - "Experiment consistency varied"

---

## Standard Requirements

### Required Experiment Structure

Every experiment MUST have:

```
LAB-XXX/
├── README.md              # Experiment overview
├── experiment.md         # Experiment definition
├── TRACKER.md            # Progress tracking
├── runs/                # Individual runs
│   ├── RUN-001.md
│   └── RUN-N.md
├── evidence/            # Evidence collected
│   └── evidence.md
└── [Optional]
    ├── impact.md        # Impact assessment
    └── synthesis.md     # Final synthesis
```

### Required Fields

#### experiment.md

| Field | Required |
|-------|----------|
| Experiment ID | YES |
| Title | YES |
| Date | YES |
| Engine | YES |
| Status | YES |
| Research Question | YES |
| Methodology | YES |
| Success Criteria | YES |

#### TRACKER.md

| Field | Required |
|-------|----------|
| Experiment ID | YES |
| Status | YES |
| Progress | YES |
| Deliverables | YES |
| Open Issues | YES |

---

## Quality Gates

### Gate 1: Design Review
- Experiment design approved
- Success criteria defined
- Engine selected

### Gate 2: Execution Review
- Runs executed per methodology
- Evidence collected per standards
- Data integrity verified

### Gate 3: Analysis Review
- Results analyzed per methodology
- Findings documented
- Confidence assigned

### Gate 4: Conclusion Review
- Conclusions supported by evidence
- Limitations acknowledged
- Recommendations justified

---

## Confidence Assignment Standards

### Level Definitions

| Level | Criteria | Evidence Required |
|-------|----------|------------------|
| **HIGH** | Strong support, multiple sources | 3+ consistent sources |
| **MEDIUM** | Moderate support | 2+ sources with minor gaps |
| **LOW** | Preliminary support | 1+ source, significant gaps |
| **UNKNOWN** | Insufficient data | No evidence yet |

### Decision Thresholds

| Confidence | Action |
|------------|--------|
| ≥80% | Trust for implementation |
| 60-79% | Consider with monitoring |
| 40-59% | Experimental, use caution |
| <40% | Requires validation |

---

## Document Quality Standards

### Evidence Quality

| Criterion | Requirement |
|-----------|-------------|
| Traceability | Every claim traces to evidence |
| Verification | Evidence can be verified |
| Linkage | Evidence connected to experiments |
| Preservation | Evidence permanently stored |
| Integrity | Evidence unchanged after collection |

### Conclusion Quality

| Criterion | Requirement |
|-----------|-------------|
| Support | Conclusions backed by evidence |
| Acknowledgment | Limitations stated |
| Uncertainty | Confidence properly calibrated |
| Alternatives | Alternative interpretations noted |

---

## Reproducibility Requirements

Each experiment must document:
1. **Environment**: What system/configuration used
2. **Methodology**: How experiment executed
3. **Data**: What data collected
4. **Analysis**: How results analyzed

---

## Changes from Seed-001

| Aspect | Seed-001 | Seed-002 |
|--------|----------|----------|
| Experiment Standards | Evolving | **Defined** |
| Required Structure | Variable | **Mandatory** |
| Quality Gates | Ad-hoc | **Standardized** |
| Confidence Levels | Basic | **Enhanced criteria** |

---

## Lessons Addressed

- **LESSON-008**: "Experiment consistency varied"

---

**Status**: NEW IN SEED-002
**Evidence**: LESSON-008
