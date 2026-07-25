# Laboratory Governance Specification

**Document ID**: LAB-GOV-001
**Version**: 1.0.0
**Date**: 2026-07-25
**Status**: DRAFT
**Authority**: KDE Runtime ECU
**Investigation**: Laboratory Governance Investigation

---

## Executive Summary

This specification defines the complete governance model for the KDE Laboratory, based on evidence from the existing repository structure and previous KDE investigations.

### Key Design Principles

1. **Determinism**: Every artifact has one deterministic home
2. **Immutability**: History is never destroyed
3. **Auditability**: Complete provenance trail
4. **Automation**: Runtime ECU manages classification, naming, and placement
5. **Protection**: Completed artifacts are locked

---

## 1. Operation Classification

### 1.1 Classification Rules

The Runtime ECU SHALL automatically classify operations based on the following rules:

| Operation Type | Trigger Keywords | Directory |
|---------------|------------------|-----------|
| **Investigation** | "investigate", "analyze", "assess", "determine", "examine" | investigations/ |
| **Experiment** | "experiment", "test hypothesis", "validate approach" | experiments/ |
| **Implementation** | "implement", "build", "create", "develop" | implementations/ |
| **Testing** | "test", "verify", "conformance" | testing/ |
| **Review** | "review", "assess quality", "audit" | reviews/ |
| **Decision** | "decide", "evaluate options", "select approach" | decisions/ |
| **Planning** | "plan", "roadmap", "schedule" | planning/ |
| **Evidence** | "evidence", "observation", "data collection" | evidence/ |

### 1.2 Classification Algorithm

```
CLASSIFY(operation_description):
    FOR each classification_rule IN rules:
        IF any keyword matches operation_description:
            RETURN classification_rule.type
    IF no match:
        RETURN Investigation (default)
```

### 1.3 Manual Classification Prohibition

- Manual classification SHALL NOT be allowed
- The Runtime ECU is the sole authority for operation classification
- Agents SHALL provide operation descriptions; ECU determines classification

### 1.4 Classification Sufficiency

**Evidence from Repository**: The existing laboratory structure shows:
- investigations/ - 11 investigations
- experiments/ - Available
- decisions/ - Available
- testing/ - Available
- reviews/ - Available
- planning/ - Available
- evidence/ - Available
- implementations/ - Available

**Recommendation**: The existing categories are sufficient. No additional categories are justified based on current evidence.

---

## 2. Artifact Lifecycle

### 2.1 Complete Lifecycle Stages

```
REQUEST
    │
    ▼
BOOTSTRAP VERIFICATION
    │
    ▼
OPERATION CLASSIFICATION
    │
    ▼
OPERATION ID ASSIGNMENT
    │
    ▼
TIMESTAMP ASSIGNMENT
    │
    ▼
WORKSPACE CREATION
    │
    ▼
EXECUTION
    │
    ▼
EVIDENCE COLLECTION
    │
    ▼
REPORT GENERATION
    │
    ▼
REVIEW (if required)
    │
    ▼
LOCK
    │
    ▼
ARCHIVE (after retention)
    │
    ▼
KNOWLEDGE EXTRACTION
    │
    ▼
REPOSITORY INTEGRATION
```

### 2.2 Stage Definitions

| Stage | Description | Owner | Mandatory? |
|-------|-------------|-------|------------|
| Request | Initial operation request | Agent | Yes |
| Bootstrap Verification | Verify runtime initialized | Runtime | Yes |
| Classification | Determine operation type | Runtime ECU | Yes |
| ID Assignment | Assign unique identifier | Runtime ECU | Yes |
| Timestamp Assignment | Record creation timestamp | Runtime ECU | Yes |
| Workspace Creation | Create operation directory | Runtime ECU | Yes |
| Execution | Perform operation work | Agent | Yes |
| Evidence Collection | Gather supporting evidence | Agent | Yes |
| Report Generation | Create operation report | Agent | Yes |
| Review | Human review (if required) | Human | No |
| Lock | Protect completed artifact | Runtime | Yes |
| Archive | Long-term storage | Runtime | No |
| Knowledge Extraction | Extract reusable knowledge | Runtime | Yes |
| Repository Integration | Update knowledge bases | Runtime | Yes |

### 2.3 Additional Stages Justified

Based on evidence, the following additional stages are required:

| Stage | Justification |
|-------|--------------|
| **Bootstrap Verification** | Ensures runtime integrity before operation |
| **Knowledge Extraction** | Preserves lessons for future operations |
| **Repository Integration** | Updates seeds, engines, governance as needed |

---

## 3. Timestamp Governance

### 3.1 Mandatory Timestamps

| Timestamp | Purpose | Format | Immutable? |
|-----------|---------|--------|------------|
| **Created** | Artifact creation time | ISO 8601 UTC | Yes |
| **Bootstrap** | Runtime bootstrap version | Version string | Yes |
| **Started** | Operation start time | ISO 8601 UTC | Yes |
| **Completed** | Operation completion time | ISO 8601 UTC | Yes |
| **Locked** | Protection timestamp | ISO 8601 UTC | Yes |
| **Modified** | Last modification time | ISO 8601 UTC | Yes (append-only) |

### 3.2 Optional Timestamps

| Timestamp | Purpose | Format | When Applied |
|-----------|---------|--------|--------------|
| **Reviewed** | Human review time | ISO 8601 UTC | After review |
| **Archived** | Archive time | ISO 8601 UTC | After retention |
| **Migrated** | Migration time | ISO 8601 UTC | After migration |
| **Superseded** | Replacement time | ISO 8601 UTC | When replaced |

### 3.3 Timestamp Format

```
ISO 8601: 2026-07-25T10:30:00.000000Z
```

### 3.4 Timezone Policy

- All timestamps SHALL be in UTC (Z suffix)
- Timezone conversion SHALL NOT be performed by Runtime ECU
- Display MAY convert to local time for human readability

### 3.5 Immutability Rules

1. **Created timestamp**: Never modified after creation
2. **Bootstrap timestamp**: Never modified
3. **Started timestamp**: Never modified after set
4. **Completed timestamp**: Never modified after set
5. **Locked timestamp**: Never modified after lock
6. **Modified timestamp**: Updated on every change (append-only log)

### 3.6 Timestamp in Identity

**Decision**: Timestamps SHALL NOT participate in artifact identity.

**Rationale**: 
- Artifact ID is the primary identity
- Timestamps are metadata for audit and provenance
- Including timestamps in identity would cause conflicts on regeneration

### 3.7 Timestamp in Audit Trail

**Decision**: Timestamps SHALL be part of the complete audit trail.

**Evidence**: INV-001 recommends provenance tracking; timestamps are essential for chronological ordering.

---

## 4. Laboratory Organization

### 4.1 Deterministic Directory Structure

```
laboratory/
├── investigations/         ← Investigation artifacts
│   ├── KDE-INV-XXX/       ← KDE framework investigations
│   └── PROJECT-INV-XXX/    ← Project investigations
├── experiments/           ← Experiment artifacts
│   └── PROJECT-EXP-XXX/
├── implementations/        ← Implementation specifications
│   └── PROJECT-IMP-XXX/
├── decisions/              ← Technology Decision Records
│   └── TDR-XXX.md
├── reviews/                ← Human review documents
│   └── PROJECT-REV-XXX.md
├── planning/               ← Planning documents
│   └── PLAN-XXX.md
├── evidence/               ← Evidence artifacts
│   └── EVD-XXX.md
├── testing/                ← Testing infrastructure
│   ├── mocks/             ← Mock implementations
│   ├── simulators/         ← Test simulators
│   └── fixtures/           ← Test fixtures
└── archives/               ← Archived artifacts (managed)
    └── (locked artifacts)
```

### 4.2 Automatic Placement Rules

| Operation Type | Directory | Subdirectory |
|---------------|-----------|--------------|
| Investigation | investigations/ | KDE-INV-* or PROJECT-INV-* |
| Experiment | experiments/ | PROJECT-EXP-* |
| Implementation | implementations/ | PROJECT-IMP-* |
| Decision | decisions/ | TDR-*.md |
| Review | reviews/ | PROJECT-REV-*.md |
| Planning | planning/ | PLAN-*.md |
| Evidence | evidence/ | EVD-*.md |
| Testing | testing/ | By testing type |

### 4.3 Runtime ECU Responsibilities

The Runtime ECU SHALL:
1. **Never require manual placement** - Always determine destination automatically
2. **Create directories as needed** - Initialize structure before artifact creation
3. **Validate placement** - Ensure artifacts are in correct directories
4. **Report violations** - Detect and report misplaced artifacts

---

## 5. Naming Convention

### 5.1 Identifier Formats

| Artifact Type | Format | Example |
|--------------|--------|---------|
| Investigation (KDE) | `KDE-INV-NNN` | `KDE-INV-001` |
| Investigation (Project) | `PROJECT-INV-NNN` | `DNP3-INV-001` |
| Experiment | `PROJECT-EXP-NNN` | `DNP3-EXP-001` |
| Implementation | `PROJECT-IMP-NNN` | `DNP3-IMP-001` |
| Decision | `TDR-NNN` | `TDR-001` |
| Review | `PROJECT-REV-NNN` | `DNP3-REV-001` |
| Planning | `PLAN-NNN` | `PLAN-001` |
| Evidence | `EVD-NNN` | `EVD-001` |
| Testing | `TEST-NNN` | `TEST-001` |

### 5.2 Numbering Policy

1. **Sequential numbering**: Each type has independent sequence
2. **No reuse**: Completed IDs are never reassigned
3. **Gap tolerance**: Missing numbers are acceptable
4. **Zero-padding**: 3 digits minimum (`001`, `002`, ... `999`)

### 5.3 Timestamp Integration

**Decision**: Timestamps are NOT part of identifiers.

**Rationale**:
- IDs should be stable across operations
- Timestamps change; IDs must remain constant
- Git history provides temporal ordering

### 5.4 Revision Numbering

For artifacts that require revisions:

| Format | Example | Usage |
|--------|---------|-------|
| Major | `INV-001-v1` | Initial version |
| Revision | `INV-001-v2` | Second version |
| Draft | `INV-001-draft` | Working draft |

**Decision**: Use git for version control; include version in metadata, not filename.

### 5.5 Uniqueness Guarantees

1. **ID registry**: Runtime maintains ID registry for each type
2. **Atomic assignment**: ID assigned at creation, never changed
3. **Collision prevention**: Check registry before ID assignment

### 5.6 Collision Prevention

```
ASSIGN_ID(type):
    next = registry.get_next(type)
    IF next IS UNIQUE:
        RETURN type + "-" + next
    ELSE:
        RAISE CollisionError
```

---

## 6. Metadata Governance

### 6.1 Mandatory Metadata Fields

| Field | Format | Example | Purpose |
|-------|--------|---------|---------|
| **ID** | `TYPE-NNN` | `DNP3-INV-001` | Unique identifier |
| **Type** | String | `investigation` | Operation classification |
| **Title** | String | `Artifact Authority Model` | Human-readable name |
| **Status** | Enum | `COMPLETED` | Lifecycle stage |
| **Bootstrap Version** | Semver | `1.0.0` | Runtime version |
| **Runtime Version** | Semver | `1.1.0` | Runtime version |
| **Created** | ISO 8601 | `2026-07-25T10:30:00Z` | Creation time |
| **Completed** | ISO 8601 | `2026-07-25T12:00:00Z` | Completion time |
| **Authority** | String | `KDE Runtime (DNP3)` | Runtime identity |
| **Agent** | String | `OpenHands` | Execution agent |

### 6.2 Optional Metadata Fields

| Field | Format | Example | When Required |
|-------|--------|---------|--------------|
| **Parent** | ID | `DNP3-INV-001` | Child operations |
| **Related** | ID[] | `[INV-001, INV-002]` | Related operations |
| **Human Reviewer** | Name | `John Doe` | Reviewed artifacts |
| **Locked** | ISO 8601 | `2026-07-25T14:00:00Z` | Completed artifacts |
| **Superseded By** | ID | `DNP3-INV-003` | Replaced artifacts |

### 6.3 Complete Metadata Specification

```yaml
artifact:
  id: TYPE-NNN
  type: investigation|experiment|implementation|testing|review|decision|planning|evidence
  title: "Human-readable title"
  status: DRAFT|IN_PROGRESS|COMPLETED|APPROVED|LOCKED|ARCHIVED|SUPERSEDED
  authority: "KDE Runtime (Project Name)"
  agent: "Execution agent name"

runtime:
  bootstrap_version: "1.0.0"
  runtime_version: "1.1.0"
  ecu_version: "1.0.0"

timestamps:
  created: ISO8601
  started: ISO8601
  completed: ISO8601
  locked: ISO8601
  modified: ISO8601

relationships:
  parent: TYPE-NNN|null
  related: [TYPE-NNN, ...]
  supersedes: TYPE-NNN|null
  superseded_by: TYPE-NNN|null

governance:
  reviewer: "Human name"|null
  review_date: ISO8601|null
  approval_status: PENDING|APPROVED|CONDITIONAL|REJECTED
```

### 6.4 Metadata Storage

**Decision**: Metadata SHALL be stored in document frontmatter (YAML block).

**Rationale**:
- Human readable
- Git tracks changes
- Easy to parse
- Standards-based

---

## 7. Artifact Protection

### 7.1 Locking Policy

| Condition | Action |
|-----------|--------|
| Operation COMPLETED | Lock automatically |
| Review APPROVED | Lock automatically |
| Agent requests | Lock allowed |
| Manual unlock | PROHIBITED |

### 7.2 Lock Implementation

**Locked artifacts**:
- Cannot be overwritten
- Cannot be deleted
- CAN be copied to archives
- CAN be referenced by new artifacts

### 7.3 Unlock Policy

**Decision**: Unlocking is PROHIBITED without governance override.

| Scenario | Allowed? | Process |
|----------|----------|---------|
| Agent requests | No | - |
| Human requests | Yes | Governance approval required |
| Runtime error | Yes | Quarantine then restore |

### 7.4 Amendment Policy

**Decision**: Completed artifacts SHALL NOT be amended; new versions SHALL be created.

| If you need to... | Then... |
|-------------------|---------|
| Fix typo | Create revision (INV-001-v2) |
| Add evidence | Create new evidence artifact |
| Update conclusion | Create new investigation |
| Replace artifact | Mark superseded, create new |

### 7.5 Revision Policy

1. **Never modify locked artifacts**
2. **Create new version for changes**
3. **Reference previous version in metadata**
4. **Include changelog in new version**

### 7.6 Superseded Artifact Handling

| Field | Value |
|-------|-------|
| Status | `SUPERSEDED` |
| Superseded By | New artifact ID |
| Timestamp | When superseded |

**Evidence**: GOV-NAMING-001 shows superseded handling.

---

## 8. Rogue Artifact Detection

### 8.1 Violation Types

| Violation | Description | Detection Method |
|-----------|------------|------------------|
| **incorrect_folder** | Artifact in wrong directory | Path validation |
| **invalid_type** | Type doesn't match directory | Prefix validation |
| **duplicate_id** | ID already exists | Registry check |
| **invalid_naming** | Non-standard name format | Pattern validation |
| **orphan** | No parent operation | Relationship check |
| **incomplete_metadata** | Missing mandatory fields | Schema validation |
| **invalid_lifecycle** | Wrong status transition | State machine |
| **unlocked_completed** | Completed but not locked | Status check |
| **manual_folder** | Folder created manually | Audit trail |
| **timestamp_inconsistency** | Invalid timestamp order | Chronological check |

### 8.2 Response Actions

| Action | Trigger | Behavior |
|--------|---------|----------|
| **Reject** | Creation-time violation | Block artifact creation |
| **Warn** | Non-critical violation | Log warning, allow operation |
| **Quarantine** | Critical violation | Isolate artifact, block access |
| **Move** | Folder violation | Relocate to correct directory |
| **Archive** | Retention violation | Move to archive |
| **Recycle** | Deletion request | Move to recycle bin |

### 8.3 Detection Rules

```python
VALIDATION_RULES = {
    'naming': [
        ('pattern', r'^[A-Z]+-[A-Z]+-\d{3}$'),
        ('prefix_matches_folder', True),
        ('case', 'upper')
    ],
    'metadata': [
        ('required_fields', ['id', 'type', 'title', 'status', 'created']),
        ('valid_status_transitions', STATE_MACHINE)
    ],
    'lifecycle': [
        ('completed_locked', True),
        ('no_future_timestamps', True)
    ],
    'placement': [
        ('directory_matches_prefix', True)
    ]
}
```

### 8.4 Runtime ECU Response Matrix

| Violation | ECU Response |
|-----------|-------------|
| incorrect_folder | Move + Warn |
| invalid_type | Reject |
| duplicate_id | Reject |
| invalid_naming | Reject + Warn |
| orphan | Warn |
| incomplete_metadata | Reject (creation) / Quarantine (existing) |
| invalid_lifecycle | Quarantine |
| unlocked_completed | Auto-lock |
| manual_folder | Warn + Log |
| timestamp_inconsistency | Reject + Quarantine |

---

## 9. Recycle Bin Specification

### 9.1 Retention Period

| Artifact Type | Retention | Rationale |
|--------------|-----------|-----------|
| Draft | 30 days | Allow recovery from accidental deletion |
| In Progress | 90 days | Longer for active work |
| Completed | 180 days | Ensure review window |

### 9.2 Recovery Policy

1. **Recovery window**: Within retention period
2. **Recovery request**: Via Runtime ECU only
3. **Recovery validation**: Verify artifact integrity
4. **Restoration**: To original location or specified location

### 9.3 Recycle Bin Structure

```
laboratory/.recycle/
├── deleted/
│   └── DNP3-INV-001/
│       ├── artifact.md
│       └── metadata.yaml
├── metadata.json
└── audit.log
```

### 9.4 Metadata Preservation

Every recycled artifact preserves:

```yaml
original:
  path: "laboratory/investigations/DNP3-INV-001/"
  id: "DNP3-INV-001"
  deleted: ISO8601
  reason: "user_request|auto_cleanup|violation"
  
audit:
  deleted_by: "agent_name"
  approved_by: "runtime|governance"
  ticket: "GOV-XXX"  # If governance approval required
```

### 9.5 Deletion Policy

| Scenario | Deletion Allowed? | Process |
|----------|-------------------|---------|
| Retention expired | Yes | Auto-deletion after audit |
| Governance request | Yes | With audit trail |
| Legal requirement | Yes | With documentation |
| Agent request | No | Must go through governance |

### 9.6 Permanent Removal

**Decision**: Permanent deletion is PROHIBITED without governance override.

| Condition | Action |
|-----------|--------|
| Legal hold | Never delete |
| Evidence artifact | Never delete |
| Investigation | Never delete |
| Governance override | Archive-only |

---

## 10. Migration Strategy

### 10.1 Migration Principles

1. **Preserve everything**: No data loss
2. **Preserve history**: Git history intact
3. **Preserve timestamps**: Where possible
4. **No overwrites**: Existing artifacts protected
5. **Auto-classify**: Runtime ECU determines types

### 10.2 Migration Process

```
MIGRATE(source_path):
    1. SCAN source_path
    2. CLASSIFY each artifact
    3. VALIDATE naming conventions
    4. GENERATE new IDs if needed
    5. CREATE directories
    6. COPY artifacts (not move)
    7. UPDATE metadata
    8. VALIDATE completeness
    9. LOCK migrated artifacts
    10. LOG migration audit
```

### 10.3 Timestamp Preservation

| Timestamp | Preservation | Fallback |
|-----------|--------------|----------|
| Created | Git commit date | Artifact file date |
| Modified | Git commit dates | File system |
| Locked | Current time | Set after migration |
| Bootstrap | Unknown | Current version |

### 10.4 Classification During Migration

**Evidence**: Existing investigations use `KDE-INV-*` and `DNP3-INV-*` prefixes.

**Migration Classification**:
1. Detect prefix pattern
2. Match to operation type
3. Place in corresponding directory
4. Validate no conflicts

### 10.5 Post-Migration Validation

| Check | Pass Condition |
|-------|---------------|
| All artifacts migrated | Count matches |
| No duplicates | ID uniqueness verified |
| Naming correct | Pattern validation |
| Metadata complete | Schema validation |
| Directory correct | Path validation |

---

## 11. Runtime Responsibilities

### 11.1 Component Responsibilities

| Component | Responsibilities |
|-----------|------------------|
| **Bootstrap** | Verify runtime state, load configuration, initialize managers |
| **Runtime ECU** | Classify operations, assign IDs, coordinate execution, enforce lifecycle |
| **Laboratory** | Own artifacts, provide structure, enforce placement |
| **Repository** | Own production code, integration points |
| **Governance** | Define policies, approve overrides, audit reviews |

### 11.2 Manager Responsibilities

| Manager | Responsibilities |
|---------|------------------|
| **Artifact Manager** | Create, store, retrieve, delete artifacts |
| **Lifecycle Manager** | Track status, enforce transitions, manage locks |
| **Lock Manager** | Apply locks, verify lock status, process unlock requests |
| **Timestamp Manager** | Assign timestamps, verify consistency, maintain audit trail |
| **Archive Manager** | Archive old artifacts, manage retention, handle recovery |
| **Recycle Manager** | Move to recycle bin, manage retention, process recovery |
| **ID Manager** | Assign IDs, track usage, prevent collisions |
| **Classification Manager** | Classify operations, validate types, update classifications |
| **Metadata Manager** | Validate metadata, update fields, enforce schema |
| **Validation Manager** | Detect violations, apply responses, maintain integrity |

### 11.3 Recommended Runtime Managers

Based on analysis, the following managers are required:

| Manager | Priority | Justification |
|---------|----------|---------------|
| Artifact Manager | Critical | Core artifact handling |
| Lifecycle Manager | Critical | Status tracking |
| ID Manager | Critical | Unique identification |
| Classification Manager | Critical | Operation classification |
| Timestamp Manager | High | Audit trail |
| Lock Manager | High | Protection |
| Metadata Manager | High | Schema enforcement |
| Validation Manager | High | Violation detection |
| Archive Manager | Medium | Long-term storage |
| Recycle Manager | Medium | Recovery capability |

### 11.4 Manager Dependencies

```
Bootstrap
    │
    ├── ID Manager (initialized)
    ├── Timestamp Manager (initialized)
    ├── Classification Manager (initialized)
    │
    ▼
Runtime ECU
    │
    ├── Artifact Manager
    ├── Lifecycle Manager
    ├── Lock Manager
    ├── Metadata Manager
    ├── Validation Manager
    │
    ├── Archive Manager (depends on Artifact)
    └── Recycle Manager (depends on Artifact)
```

---

## 12. Runtime Responsibility Matrix

| Responsibility | Bootstrap | ECU | Lab | Repo | Gov |
|---------------|-----------|-----|-----|------|-----|
| Artifact creation | Init | Assign ID | Own | N/A | N/A |
| Artifact storage | N/A | Coordinate | Manage | N/A | N/A |
| Classification | Init | Primary | Support | N/A | N/A |
| ID assignment | N/A | Primary | N/A | N/A | N/A |
| Timestamp assignment | N/A | Primary | N/A | N/A | N/A |
| Lifecycle management | N/A | Primary | N/A | N/A | N/A |
| Locking | N/A | Primary | N/A | N/A | N/A |
| Metadata validation | N/A | Primary | N/A | N/A | N/A |
| Violation detection | N/A | Primary | N/A | N/A | N/A |
| Archiving | N/A | Coordinate | N/A | N/A | N/A |
| Recycle bin | N/A | Coordinate | N/A | N/A | N/A |
| Policy definition | N/A | N/A | N/A | N/A | Primary |
| Override approval | N/A | N/A | N/A | N/A | Primary |
| Audit review | N/A | N/A | N/A | N/A | Primary |

---

## 13. Engineering Recommendations

### 13.1 Immediate Actions

| Recommendation | Priority | Action |
|----------------|----------|--------|
| Implement ID Registry | High | Runtime ECU tracks next ID per type |
| Add auto-classification | High | ECU classifies based on keywords |
| Implement locking | High | Auto-lock on completion |
| Add validation | High | Detect violations at creation |
| Create recycle bin | Medium | Enable recovery |

### 13.2 Short-term Actions

| Recommendation | Priority | Action |
|----------------|----------|--------|
| Timestamp standardization | Medium | Enforce ISO 8601 UTC |
| Metadata schema | Medium | Validate all artifacts |
| Archive strategy | Medium | Define retention periods |
| Migration tool | Medium | Auto-classify existing |

### 13.3 Long-term Actions

| Recommendation | Priority | Action |
|----------------|----------|--------|
| Manager implementation | Low | Implement each manager |
| Dashboard | Low | Visualize laboratory health |
| Analytics | Low | Track patterns over time |
| AI classification | Future | ML-based classification |

### 13.4 Governance Recommendations

| Recommendation | Priority | Action |
|----------------|----------|--------|
| Formalize this spec | High | Adopt as GOV-GOV-001 |
| Create templates | High | Standardize artifact formats |
| Document overrides | Medium | Track governance decisions |
| Train agents | Medium | Ensure consistent behavior |

---

## 14. Specification Summary

### 14.1 Classification

- 8 operation types: Investigation, Experiment, Implementation, Testing, Review, Decision, Planning, Evidence
- Auto-classification by Runtime ECU
- Manual classification prohibited

### 14.2 Lifecycle

- 13 stages from Request to Repository Integration
- Bootstrap Verification added
- Knowledge Extraction added
- Repository Integration added

### 14.3 Timestamps

- 6 mandatory: Created, Bootstrap, Started, Completed, Locked, Modified
- 4 optional: Reviewed, Archived, Migrated, Superseded
- ISO 8601 UTC format
- Immutability enforced

### 14.4 Organization

- Deterministic directory structure
- Runtime ECU handles all placement
- Manual placement prohibited

### 14.5 Naming

- Standard formats per type
- Sequential numbering
- Timestamps not in identifiers
- Collision prevention via registry

### 14.6 Metadata

- 10 mandatory fields
- Complete schema defined
- YAML frontmatter storage

### 14.7 Protection

- Auto-lock on completion
- Unlock prohibited without governance
- Amendment creates new version
- Superseded artifacts tracked

### 14.8 Governance

- 10 violation types
- 6 response actions
- Complete response matrix

### 14.9 Recycle Bin

- Retention periods by type
- Recovery within window
- Permanent deletion prohibited

### 14.10 Migration

- Preserve everything
- Auto-classification
- Post-migration validation

### 14.11 Runtime

- 9 managers recommended
- Clear dependencies
- Complete responsibility matrix

---

## 15. Adoption Recommendation

This specification SHOULD be adopted as the official **KDE Laboratory Governance Standard (GOV-LAB-001)**.

**Implementation Phases**:

| Phase | Content | Priority |
|-------|---------|----------|
| Phase 1 | Classification, Lifecycle, Timestamps | High |
| Phase 2 | Naming, Metadata, Organization | High |
| Phase 3 | Protection, Governance, Validation | Medium |
| Phase 4 | Recycle Bin, Archive, Migration | Medium |
| Phase 5 | Manager Implementation | Low |

---

*Specification created by Laboratory Governance Investigation*
*Runtime ECU: Bootstrap SUCCESS, 8 engines, 2 seeds, fully operational*
