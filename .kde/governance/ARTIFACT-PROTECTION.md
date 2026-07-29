# Artifact Protection Matrix

**Document ID**: ARTIFACT-PROTECTION
**Version**: 1.0.0
**Date**: 2026-07-23
**Status**: PRODUCTION
**Authority**: Governance
**Source**: LAB-039 Revision, LAB-040 Implementation

---

## Purpose

This document defines the protection levels for all artifacts in the KDE repository. It provides:
- A single source of truth for artifact protection
- Clear guidance on what actions are permitted
- Maintenance procedures to keep the matrix current

---

## Protection Levels

### Level Definitions

| Level | Description | Enforcement | Override Allowed |
|-------|-------------|-------------|-----------------|
| **ABSOLUTE** | Never modify | Policy + Process | NO |
| **HIGH** | Human approval required | Policy | YES (with approval) |
| **MEDIUM** | Follow SOP | Procedure | YES |
| **LOW** | Mutable | Design | YES |

### Level Behaviors

| Level | CREATE | MODIFY | DELETE | RENAME | MOVE |
|-------|--------|--------|--------|---------|------|
| ABSOLUTE | BLOCKED | BLOCKED | BLOCKED | BLOCKED | BLOCKED |
| HIGH | WARN + ACK | WARN + ACK | WARN + ACK | WARN + ACK | WARN + ACK |
| MEDIUM | WARN | WARN | WARN | WARN | WARN |
| LOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW |

---

## Artifact Protection Table

### Protected Artifacts

| Artifact Type | Pattern | Protection Level | Rules |
|--------------|---------|-----------------|-------|
| Seeds | `seeds/seed-*` | ABSOLUTE | NEVER-MODIFY.md |
| Evidence | `**/evidence/**` | ABSOLUTE | EVIDENCE.md |
| Historical Experiments | `laboratory/experiments/LAB-[0-9]{3}` (XXX < current) | HIGH | This document |
| Promoted Knowledge | `knowledge/**` (PROMOTED status) | ABSOLUTE | STATE-MACHINE.md |
| Draft Knowledge | `knowledge/**` (DRAFT status) | MEDIUM | SOP |
| Governance Documents | `governance/**` | HIGH | GOVERNANCE/README.md |
| Runtime Configuration | `governance/runtime/**` | HIGH | Human Authority |
| Bootstrap | `laboratory/BOOTSTRAP.md` | HIGH | Human Authority |
| Laboratory Rules | `laboratory/LABORATORY-RULES.md` | HIGH | Human Authority |

### Mutable Artifacts

| Artifact Type | Pattern | Protection Level | Rationale |
|--------------|---------|-----------------|-----------|
| Current Experiments | `laboratory/experiments/LAB-[0-9]{3}` (current) | MEDIUM | Active work |
| Investigations | `laboratory/investigations/**` | MEDIUM | Active research |
| Templates | `laboratory/templates/**` | LOW | Reference only |
| Playground | `playground/**` | LOW | Mutable by design |
| Scratch | `scratch/**` | LOW | Mutable by design |

---

## Default Protection

### Unknown Artifacts

If an artifact does not match any defined protection pattern:

| Setting | Value | Rationale |
|---------|-------|-----------|
| **Default Level** | MEDIUM | Conservative protection |
| **Warning** | YES | Log that no pattern matched |
| **Override Allowed** | YES | Human can approve |

**Rule**: All artifacts receive at least MEDIUM protection unless explicitly set to LOW by governance.

---

## Pattern Matching

### Matching Algorithm

When an artifact path matches multiple protection patterns:

1. **Collect all matching patterns**
2. **Identify highest protection level**
3. **Apply highest protection level**

### Priority Order

Protection levels rank from highest to lowest:

| Rank | Level | Priority Value |
|------|-------|---------------|
| 1 | ABSOLUTE | 100 |
| 2 | HIGH | 75 |
| 3 | MEDIUM | 50 |
| 4 | LOW | 25 |

### Conflict Resolution

When multiple patterns have the same protection level:

| Scenario | Resolution |
|----------|------------|
| Specific vs General | More specific pattern wins |
| Explicit vs Wildcard | Explicit pattern wins |
| Recursive vs Single | Recursive pattern wins |

### Pattern Priority Matrix

| Pattern A Level | Pattern B Level | Result |
|----------------|----------------|--------|
| ABSOLUTE | ANY | ABSOLUTE |
| HIGH | MEDIUM | HIGH |
| HIGH | LOW | HIGH |
| MEDIUM | LOW | MEDIUM |
| LOW | LOW | LOW |

---

## Prohibited Actions by Level

### ABSOLUTE Prohibited

- **CREATE**: Not applicable (already exists)
- **MODIFY**: Any change to content
- **DELETE**: Any removal
- **RENAME**: Any identifier change
- **MOVE**: Any location change

### HIGH Prohibited (Without Approval)

- **MODIFY**: Requires acknowledgment
- **DELETE**: Requires acknowledgment
- **RENAME**: Requires acknowledgment
- **MOVE**: Requires acknowledgment

### MEDIUM Guidelines

- Follow SOP for all operations
- Document changes
- Obtain review as needed

### LOW Permitted

- Standard file operations
- No restrictions

---

## Maintenance

### Review Schedule

| Activity | Frequency | Responsible | Evidence |
|----------|-----------|-------------|----------|
| Full protection review | Quarterly | Governance | Review log |
| New artifact pattern review | As needed | AI triggers | Trigger record |
| Pattern validation | On load | Runtime | Validation log |
| Compliance audit | Semi-annually | Governance | Audit report |

### Review Triggers

Protection matrix review is automatically triggered when:

1. **New artifact discovery**: When AI creates or discovers artifact type not in matrix
2. **Protection violation**: When protection check is bypassed or fails
3. **Architecture change**: When KDE directory structure changes
4. **Governance directive**: When human requests review
5. **Scheduled**: Quarterly scheduled review

### Review Process

```
1. Review Triggered
         │
         ▼
2. Preliminary Assessment (AI or Human initiates)
         │
         ▼
3. Document Proposed Changes
         │
         ▼
4. Governance Review
         │
         ▼
5. Human Approval
         │
         ▼
6. Publish Updates
         │
         ▼
7. Notify AI Agents
```

### Maintenance Log

| Date | Activity | Responsible | Changes | Approved By |
|------|-----------|-------------|---------|------------|
| 2026-07-23 | Initial creation | LAB-040 | Initial matrix | Human Review |

---

## Related Documents

| Document | Purpose |
|----------|---------|
| [LABORATORY/BOOTSTRAP.md](../laboratory/BOOTSTRAP.md) | Entry point with protection reference |
| [LABORATORY/LABORATORY-RULES.md](../laboratory/LABORATORY-RULES.md) | AI behavioral rules |
| [LABORATORY/EVIDENCE.md](../laboratory/EVIDENCE.md) | Evidence protection rules |
| [governance/runtime/protection.yaml](./runtime/protection.yaml) | Runtime protection configuration |
| [governance/runtime/RUNTIME-STARTUP.md](./runtime/RUNTIME-STARTUP.md) | Runtime initialization |

---

## Version History

| Version | Date | Changes | Authority |
|---------|------|---------|-----------|
| 1.0.0 | 2026-07-23 | Initial release | Human Review |

---

**Authority**: Governance
**Status**: PRODUCTION
**Review Date**: 2026-10-23 (Quarterly)
