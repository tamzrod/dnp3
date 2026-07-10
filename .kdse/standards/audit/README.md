# KDSE Audit System

## Purpose

The KDSE Audit System provides standardized methods for evaluating software engineering methodologies and their application. Audits ensure that KDSE itself evolves through evidence and that repositories claiming KDSE compliance meet established standards.

## Documents Overview

| Document | Purpose |
|----------|---------|
| [README.md](README.md) | This overview and decision guide |
| [AUDIT_SCORING.md](AUDIT_SCORING.md) | Official scoring model (0-10 scale, 6 maturity levels) |
| [AUDIT_MATURITY.md](AUDIT_MATURITY.md) | Methodology maturity definition |
| [AUDIT_TEMPLATE.md](AUDIT_TEMPLATE.md) | Canonical audit report template |
| [FOUNDATION_AUDIT.md](FOUNDATION_AUDIT.md) | Foundation Audit standard specification |
| [COMPLIANCE_AUDIT.md](COMPLIANCE_AUDIT.md) | Compliance Audit standard specification |
| [KDSE_FOUNDATION_AUDIT.md](KDSE_FOUNDATION_AUDIT.md) | Historical audit record (pre-standardization) |

## Audit Philosophy

Audits in KDSE serve two purposes:

1. **Validation**: Confirm that a methodology or repository meets established standards
2. **Improvement**: Identify gaps and provide actionable recommendations

KDSE audits are evidence-driven. Findings must be supported by direct observation of artifacts, not assumptions or opinions.

## Audit Types

### Foundation Audit

Evaluates the KDSE methodology itself.

**Use When:**
- KDSE is being updated or evolved
- A new version of KDSE is being released
- Internal quality assurance is required

**Scope:**
- All KDSE foundation documents
- Methodology consistency
- Conceptual completeness
- Terminology alignment

**Frequency:**
- Before each major KDSE release
- After significant methodology changes

**See:** [FOUNDATION_AUDIT.md](FOUNDATION_AUDIT.md)

### Compliance Audit

Evaluates a repository that claims to follow KDSE.

**Use When:**
- An external project wants to validate KDSE adoption
- Case study evidence is being collected
- KDSE compliance is being verified

**Scope:**
- Repository artifacts and their relationships
- Traceability between artifact types
- Authority hierarchy implementation
- Verification practices

**Frequency:**
- As needed for validation purposes
- Before claiming KDSE compliance

**See:** [COMPLIANCE_AUDIT.md](COMPLIANCE_AUDIT.md)

## Decision Guide

```
Is the audit target the KDSE methodology itself?
├── YES → Foundation Audit
└── NO → Is the repository claiming KDSE compliance?
    ├── YES → Compliance Audit
    └── NO → Not a KDSE audit scope
```

## Scoring Model

All KDSE audits use a 0-10 scoring scale with defined maturity levels:

| Score Range | Maturity Level | Meaning |
|-------------|-----------------|---------|
| 0-2 | Concept | Early stage, basic ideas exist |
| 2-4 | Defined | Concepts documented, processes informal |
| 4-6 | Structured | Formal processes, documented practices |
| 6-8 | Usable | Applied consistently, produces value |
| 8-9 | Validated | Outcomes measured, benefits demonstrated |
| 9-10 | Proven | Repeated success across contexts |

**See:** [AUDIT_SCORING.md](AUDIT_SCORING.md)

## Methodology Maturity

KDSE distinguishes between methodology maturity and repository maturity:

- **Methodology Maturity**: How developed is the methodology itself?
- **Repository Maturity**: How well does a repository implement KDSE?

Both use the same scoring scale but evaluate different dimensions.

**See:** [AUDIT_MATURITY.md](AUDIT_MATURITY.md)

## Audit Process

All KDSE audits follow a standard process:

```
1. Planning
   ├── Define scope
   ├── Identify prerequisites
   └── Gather required inputs
   
2. Execution
   ├── Examine artifacts
   ├── Evaluate criteria
   └── Document findings
   
3. Analysis
   ├── Score each dimension
   ├── Identify gaps
   └── Analyze evidence
   
4. Reporting
   ├── Complete template
   ├── Provide recommendations
   └── Deliver verdict
```

## Audit Reports

All KDSE audit reports follow the standard template defined in [AUDIT_TEMPLATE.md](AUDIT_TEMPLATE.md).

Reports must include:
- Standard metadata (version, date, auditor, etc.)
- Dimension scores with justifications
- Evidence supporting findings
- Prioritized recommendations
- Final verdict

## Standards Versioning

KDSE audit standards evolve through evidence:

- Current Audit Standard Version: 1.0
- Previous Audits: Use their documented standard version
- Version Changes: Tracked in evolution documentation

## Relationship to KDSE Methodology

The audit system is itself part of KDSE:

1. **Evidence-Driven**: Audits provide evidence for methodology improvement
2. **Formal Process**: Audits follow defined methodology processes
3. **Traceable**: Audit findings trace to specific artifacts
4. **Verifiable**: Audit results can be independently verified

## Extending the Audit System

Future audit types may be added:

- **Adoption Audit**: Evaluates KDSE adoption progress
- **Traceability Audit**: Deep-dives into traceability practices
- **Verification Audit**: Evaluates verification implementation
- **Scale Audit**: Evaluates KDSE at different scales

New audit types should:
- Follow the existing template structure
- Use the standard scoring model
- Be documented in this directory
- Include clear scope definition

## Quick Reference

| Question | Answer |
|----------|--------|
| How do I audit KDSE itself? | Run a Foundation Audit |
| How do I verify KDSE compliance? | Run a Compliance Audit |
| Where are audit standards? | This directory |
| Where are audit templates? | AUDIT_TEMPLATE.md |
| What scoring scale is used? | 0-10 with 6 maturity levels |

---

*This audit system is part of the KDSE methodology. Audits validate that KDSE is followed and provide evidence for KDSE evolution.*
