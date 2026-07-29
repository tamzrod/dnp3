# KDE Timestamp Standard

**Document ID**: TIMESTAMP-STANDARD
**Version**: 1.0.0
**Date**: 2026-07-24
**Authority**: Human Authority
**Status**: PRODUCTION
**Source**: INV-TEMPORAL-PROVENANCE (Human Approved)

---

## Purpose

This document defines the canonical timestamp standard for all KDE repository artifacts. It ensures consistent temporal provenance, reproducibility, traceability, and engineering auditability.

---

## Canonical Timestamp Format

### Standard Format

| Format | Example |
|--------|---------|
| **ISO-8601 UTC** | `2026-07-24T12:00:00Z` |

### Requirements

1. **Timezone**: UTC (Coordinated Universal Time)
2. **Suffix**: Z (indicates UTC)
3. **Precision**: Seconds (HH:MM:SS)
4. **Separator**: Hyphens (-) for date, colons (:) for time

### Invalid Formats

| Invalid | Reason |
|---------|--------|
| `2026-07-24` | Missing time component |
| `2026-07-24 12:00:00` | Missing Z suffix, space separator |
| `2026/07/24T12:00:00Z` | Slashes in date |
| `2026-07-24T12:00:00+00:00` | +00:00 instead of Z |
| `1753368000` | Unix epoch (not human-readable) |

---

## Timestamp Fields by Artifact Type

### All Documents

| Field | Required | Description |
|-------|---------|-------------|
| `created` | **YES** | When document first created |
| `modified` | **YES** | When document last changed |

### Investigations

| Field | Required | Description |
|-------|---------|-------------|
| `created` | **YES** | Investigation started |
| `modified` | **YES** | Last update |
| `completed` | RECOMMENDED | Investigation finished |
| `approved` | FOR CONCLUSION | Human approval |

### Experiments

| Field | Required | Description |
|-------|---------|-------------|
| `created` | **YES** | Experiment created |
| `modified` | **YES** | Last update |
| `started` | RECOMMENDED | First run executed |
| `completed` | RECOMMENDED | All runs finished |

### Experiment Runs

| Field | Required | Description |
|-------|---------|-------------|
| `executed` | **YES** | Run execution time |
| `duration_seconds` | RECOMMENDED | Execution duration |

### Knowledge (PROMOTED)

| Field | Required | Description |
|-------|---------|-------------|
| `created` | **YES** | Original creation |
| `promoted` | **YES** | When promoted to Knowledge |
| `validated` | RECOMMENDED | Last validation |

### Governance Documents

| Field | Required | Description |
|-------|---------|-------------|
| `created` | **YES** | Document creation |
| `modified` | **YES** | Last modification |
| `approved` | **YES** | Human approval |
| `effective` | RECOMMENDED | When effective |

### Seeds

| Field | Required | Description |
|-------|---------|-------------|
| `created` | **YES** | Seed creation |
| `frozen` | **YES** | When seed was frozen |

### Engine Specifications

| Field | Required | Description |
|-------|---------|-------------|
| `created` | **YES** | Engine specification created |
| `modified` | **YES** | Specification changed |
| `effective` | **YES** | When engine became effective |
| `superseded` | FOR HISTORICAL | When engine was superseded |

---

## Document Header Format

Every KDE document SHALL begin with the following header:

```markdown
# [Document Title]

**Document ID**: [ID]
**Version**: X.Y.Z
**created**: YYYY-MM-DDTHH:MM:SSZ
**modified**: YYYY-MM-DDTHH:MM:SSZ
**Authority**: [Authority]
**Status**: [STATUS]
**Source**: [Source Investigation, if applicable]

---
```

### Required Header Fields

| Field | Description |
|-------|-------------|
| `created` | ISO-8601 UTC timestamp of document creation |
| `modified` | ISO-8601 UTC timestamp of last modification |

### Optional Header Fields

| Field | Description |
|-------|-------------|
| `approved` | Human approval timestamp |
| `effective` | When document became effective |
| `completed` | When work was completed |

---

## Version History Format

All versioned documents SHALL include a version history section:

```markdown
## Version History

| Version | Date | Changes | Authority |
|---------|------|---------|-----------|
| 1.0.0 | 2026-07-24T12:00:00Z | Initial release | Human Authority |
| 1.1.0 | 2026-07-25T12:00:00Z | Added section | INV-XXX |
```

### Version History Requirements

| Field | Format |
|-------|--------|
| Version | X.Y.Z semantic versioning |
| Date | ISO-8601 UTC (Z suffix) |
| Changes | Brief description |
| Authority | Who approved the change |

---

## Rationale

### Why ISO-8601 UTC?

| Reason | Explanation |
|--------|-------------|
| **International Standard** | ISO-8601 is the global standard for date/time representation |
| **UTC Only** | Single timezone eliminates ambiguity across global teams |
| **Z Suffix** | Explicit UTC indicator (Zulu time) |
| **Lexicographic Sortable** | Chronological order matches alphabetical order |
| **Human Readable** | Unlike Unix epoch, easily interpreted by humans |
| **Existing Practice** | Already used in Knowledge and Investigation artifacts |

### Why Seconds Precision?

| Reason | Explanation |
|--------|-------------|
| **Audit Trail** | Enables precise reconstruction of events |
| **Uniqueness** | Prevents timestamp collisions within same day |
| **Compatibility** | ISO-8601 supports this precision |
| **Sufficient** | Millisecond precision unnecessary for KDE artifacts |

---

## Migration Guide

### Updating Existing Documents

1. Add `created` field with original creation date (estimate if unknown)
2. Add `modified` field with current date
3. Ensure date-only fields use full ISO-8601 format
4. Update version history with migration note

### Example Migration

**Before**:
```markdown
**Date**: 2026-07-24
**Version**: 1.0.0
```

**After**:
```markdown
**created**: 2026-07-24T00:00:00Z
**modified**: 2026-07-24T12:00:00Z
**Version**: 1.1.0
```

---

## Enforcement

### Template Compliance

All official templates SHALL enforce this standard:
- [x] Investigation template (v2.0.0)
- [x] Experiment template (v2.0.0)
- [x] Run template (v2.0.0)

### New Documents

All new documents created using official templates SHALL comply with this standard.

### Existing Documents

Existing documents are encouraged to migrate but not required.

---

## Related Documents

| Document | Relationship |
|----------|--------------|
| [laboratory/templates/investigation-template.md](../laboratory/templates/investigation-template.md) | Investigation template |
| [laboratory/templates/experiment-template.md](../laboratory/templates/experiment-template.md) | Experiment template |
| [laboratory/templates/run-template.md](../laboratory/templates/run-template.md) | Run template |
| [INV-TEMPORAL-PROVENANCE](../laboratory/investigations/INV-TEMPORAL-PROVENANCE/) | Source investigation |

---

## Version History

| Version | Date | Changes | Authority |
|---------|------|---------|-----------|
| 1.0.0 | 2026-07-24T12:00:00Z | Initial release | Human (INV-TEMPORAL-PROVENANCE approved) |

---

**Status**: PRODUCTION
**Authority**: Human Authority
**Review Date**: Upon evidence of need
