# Technology Decision Record Template

**File**: TDR.md
**Version**: 1.0.0
**Date**: 2026-07-26
**Source**: KDE-INV-052

---

## Purpose

This template provides the standard format for Technology Decision Records (TDRs).

## Usage

Copy this template to create a new Decision:

```bash
cp .kde/templates/TDR.md laboratory/decisions/TDR-XXX.md
```

## Template

```markdown
# Technology Decision Record: TDR-XXX

**Decision ID**: TDR-XXX
**Date**: YYYY-MM-DD
**Status**: PROPOSED | ACCEPTED | REJECTED | SUPERSEDED
**Author**: [Author]
**Human Authority**: [Approver]

---

## Context

[Description of the situation that requires a decision.]

## Decision

[Clear statement of the decision made.]

## Rationale

[Explanation of why this decision was made, including alternatives considered.]

### Alternatives Considered

| Alternative | Pros | Cons | Selected |
|-------------|------|------|----------|
| [Alt 1] | [Pros] | [Cons] | [Yes/No] |
| [Alt 2] | [Pros] | [Cons] | [Yes/No] |

## Consequences

### Positive
- [Benefit 1]
- [Benefit 2]

### Negative
- [Drawback 1]
- [Drawback 2]

## Evidence

| Source | Relevance |
|--------|-----------|
| [Source 1] | [Relevance] |

## Related Artifacts

| Artifact | Relationship |
|----------|--------------|
| [Artifact] | [Relationship] |

## Review History

| Version | Date | Reviewer | Decision |
|---------|------|----------|----------|
| 1.0 | YYYY-MM-DD | [Name] | [Decision] |

---

**Status**: [PROPOSED/ACCEPTED/REJECTED/SUPERSEDED]
**Authority**: Human
```

---

*Per KDE-INV-052*
