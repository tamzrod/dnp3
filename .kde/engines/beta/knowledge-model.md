# Beta Engine Knowledge Model

**Engine ID**: KDE-ENGINE-002
**Version**: 0.1.0

---

## Overview

The Beta Engine Knowledge Model defines the structure of scientific knowledge objects produced by the Beta pipeline. This model extends traditional pattern representation with context, boundaries, and statistical confidence.

---

## Knowledge Object Hierarchy

```
Knowledge
├── Identity
├── Statement
├── Evidence
├── Statistics
├── Confidence
├── Context
├── Boundaries
├── Assumptions
├── Reproducibility
├── Version
└── Provenance
```

---

## Field Specifications

### 1. Identity

```yaml
identity:
  id: KNOW-XXX          # Unique knowledge identifier
  version: "1.0"        # Knowledge version
  name: [optional name] # Human-readable name
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | String | Yes | Unique identifier (format: KNOW-XXX) |
| `version` | String | Yes | Semantic version of knowledge |
| `name` | String | No | Human-readable name |

### 2. Statement

```yaml
statement: |
  [Precise, testable knowledge claim]
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `statement` | String | Yes | Clear, testable claim |

**Statement Guidelines:**
- State "what" is true
- Include "when" conditions
- Be specific, not general
- Avoid absolute claims unless justified
- Include directionality (increases, decreases, correlates)

### 3. Evidence

```yaml
evidence:
  supporting:
    - evidence_id: EV-XXX
      description: [description]
      contribution: [direct_support|indirect_support]
  contradicting:
    - evidence_id: EV-XXX
      description: [description]
      contribution: [contradiction|exception]
  total_count: n
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `supporting` | Array | Yes | Evidence that supports the claim |
| `contradicting` | Array | No | Evidence that contradicts |
| `total_count` | Integer | Yes | Total evidence items |

**Evidence Item Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `evidence_id` | String | Yes | Reference to evidence |
| `description` | String | Yes | What this evidence shows |
| `contribution` | Enum | Yes | How evidence contributes |

### 4. Statistics

```yaml
statistics:
  sample_size: n
  p_value: p
  correlation: r
  confidence_interval:
    lower: x
    upper: y
    confidence_level: 95
  chi_square: χ²
  effect_size: d
  power: 1-β
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `sample_size` | Integer | Yes | Number of observations |
| `p_value` | Float | Yes | Statistical significance |
| `correlation` | Float | No | Correlation coefficient |
| `confidence_interval` | Object | No | Range of likely values |
| `chi_square` | Float | No | Chi-square statistic |
| `effect_size` | Float | No | Cohen's d or similar |
| `power` | Float | No | Statistical power |

### 5. Confidence

```yaml
confidence:
  value: 97           # 0-100
  level: high         # high, medium, low
  basis: statistical  # statistical, expert, heuristic
  factors:
    - Large sample size (n=70)
    - High correlation (r=0.72)
    - Low p-value (p<0.01)
    - Consistent across contexts
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `value` | Integer | Yes | Confidence percentage (0-100) |
| `level` | Enum | Yes | Qualitative assessment |
| `basis` | Enum | Yes | How confidence was derived |
| `factors` | Array | Yes | Contributing factors |

**Confidence Levels:**

| Level | Value Range | Description |
|-------|-------------|-------------|
| `high` | 80-100 | Strong evidence, likely stable |
| `medium` | 50-79 | Moderate evidence, needs validation |
| `low` | 0-49 | Weak evidence, preliminary |

**Confidence Bases:**

| Basis | Description |
|-------|-------------|
| `statistical` | Derived from statistical analysis |
| `expert` | Expert assessment |
| `heuristic` | Rule-of-thumb or experience |

### 6. Context

```yaml
context:
  dimensions:
    - name: dimension_name
      values: [applicable values]
      evidence_count: n
      confidence: high
  conditions: |
    [Detailed description of applicable conditions]
  applicability: conditional  # universal, conditional, limited
  applicability_confidence: high
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `dimensions` | Array | Yes | Applicable dimensions |
| `conditions` | String | No | Detailed conditions |
| `applicability` | Enum | Yes | Scope of applicability |
| `applicability_confidence` | Enum | Yes | Confidence in applicability |

**Dimension Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | String | Yes | Dimension identifier |
| `values` | Array | Yes | Applicable values |
| `evidence_count` | Integer | Yes | Evidence for this dimension |
| `confidence` | Enum | Yes | Confidence in this dimension |

**Applicability Levels:**

| Level | Description |
|-------|-------------|
| `universal` | Applies in all contexts |
| `conditional` | Applies under specific conditions |
| `limited` | Applies in narrow scope |

### 7. Boundaries

```yaml
boundary:
  type: exception  # contradiction, exception, edge_case, reverse, diminishing
  description: |
    [Description of when pattern fails]
  conditions: |
    [Specific conditions where pattern fails]
  severity: minor  # critical, major, minor
  evidence_count: n
  handling: |
    [How to handle when boundary is encountered]
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | Enum | Yes | Type of boundary |
| `description` | String | Yes | What the boundary is |
| `conditions` | String | No | When boundary applies |
| `severity` | Enum | Yes | Impact of boundary |
| `evidence_count` | Integer | Yes | Evidence for boundary |
| `handling` | String | No | How to handle |

**Boundary Types:**

| Type | Description |
|------|-------------|
| `contradiction` | Pattern reverses |
| `exception` | Specific failure cases |
| `edge_case` | Extreme condition failure |
| `reverse` | Opposite relationship |
| `diminishing` | Pattern weakens over time |

**Severity Levels:**

| Level | Description |
|-------|-------------|
| `critical` | Major impact on applicability |
| `major` | Moderate impact |
| `minor` | Small impact |

### 8. Assumptions

```yaml
assumptions:
  - assumption: [assumption text]
    verified: true
    notes: [optional notes]
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `assumption` | String | Yes | Assumption text |
| `verified` | Boolean | Yes | Whether verified |
| `notes` | String | No | Additional notes |

### 9. Reproducibility

```yaml
reproducibility:
  verified: true
  verification_count: 3
  replications:
    - experiment_id: LAB-XXX
      result: confirmed
      date: 2026-07-20
    - experiment_id: LAB-YYY
      result: partial
      date: 2026-07-21
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `verified` | Boolean | Yes | Whether independently verified |
| `verification_count` | Integer | Yes | Number of verifications |
| `replications` | Array | No | Individual replications |

**Replication Result Types:**

| Result | Description |
|--------|-------------|
| `confirmed` | Fully reproduced |
| `partial` | Partially reproduced |
| `failed` | Not reproduced |

### 10. Version

```yaml
version:
  major: 1
  minor: 0
  patch: 0
  changelog:
    - version: 1.0.0
      date: 2026-07-20
      changes:
        - Initial knowledge creation
```

### 11. Provenance

```yaml
provenance:
  engine_id: KDE-ENGINE-002
  engine_version: 0.1.0
  engine_codename: Beta
  created: 2026-07-20T12:00:00Z
  created_by: [actor]
  source_pattern: VP-XXX
  source_context: PC-XXX
  source_boundary: PB-XXX
  lineage:
    - parent_knowledge: null
      relationship: derived_from
```

---

## Complete Knowledge Object Example

```yaml
knowledge:
  # Identity
  id: KNOW-001
  version: "1.0"
  name: "Classical Opening Center Control"

  # Statement
  statement: |
    In classical chess openings with time control >= 60 minutes,
    center control (defined as pawns on e4 and d4 or equivalent)
    correlates with winning outcomes.

  # Evidence
  evidence:
    supporting:
      - evidence_id: EV-001
        description: Analysis of 50 classical games
        contribution: direct_support
      - evidence_id: EV-002
        description: Statistical correlation study
        contribution: direct_support
    contradicting:
      - evidence_id: EV-003
        description: Hypermodern exception case
        contribution: exception
    total_count: 52

  # Statistics
  statistics:
    sample_size: 70
    p_value: 0.003
    correlation: 0.72
    confidence_interval:
      lower: 0.58
      upper: 0.82
      confidence_level: 95
    chi_square: 15.3
    effect_size: 0.89

  # Confidence
  confidence:
    value: 97
    level: high
    basis: statistical
    factors:
      - Large sample size (n=70)
      - High correlation (r=0.72)
      - Very low p-value (p<0.01)
      - Effect size indicates practical significance

  # Context
  context:
    dimensions:
      - name: opening_type
        values: [classical, traditional]
        evidence_count: 45
        confidence: high
      - name: time_control
        values: [classical, rapid]
        evidence_count: 52
        confidence: medium
      - name: player_rating
        values: [intermediate, expert]
        evidence_count: 38
        confidence: medium
    conditions: |
      Applies when:
      - Time control is classical (60+ minutes) or rapid (10-60 min)
      - Player rating is intermediate (1400-1800) or above
      - Opening follows traditional principles
    applicability: conditional
    applicability_confidence: high

  # Boundaries
  boundary:
    type: exception
    description: |
      This pattern does NOT apply to:
      - Hypermodern openings (pieces control center without occupying)
      - Flank openings (a-side focus)
      - Endgame positions (when center control becomes less relevant)
    conditions: |
      Boundary applies when:
      - Opening classification is hypermodern
      - Position is past move 30 (endgame phase)
      - Pawn structure severely disrupted
    severity: minor
    evidence_count: 12
    handling: |
      When encountering hypermodern or flank openings,
      use alternative pattern KNOW-002 (Hypermodern Strategy).

  # Assumptions
  assumptions:
    - assumption: Players are competent and follow opening principles
      verified: true
    - assumption: Games are played under standard tournament conditions
      verified: true
    - assumption: "Center" is defined as e4, d4, e5, d5 squares
      verified: false
      notes: Definition may vary by analysis

  # Reproducibility
  reproducibility:
    verified: true
    verification_count: 3
    replications:
      - experiment_id: LAB-012
        result: confirmed
        date: 2026-07-20
      - experiment_id: LAB-013
        result: confirmed
        date: 2026-07-21
      - experiment_id: LAB-014
        result: partial
        date: 2026-07-22
        notes: Partial due to data quality issues

  # Version
  version:
    major: 1
    minor: 0
    patch: 0
    changelog:
      - version: 1.0.0
        date: 2026-07-20
        changes:
          - Initial knowledge creation

  # Provenance
  provenance:
    engine_id: KDE-ENGINE-002
    engine_version: 0.1.0
    engine_codename: Beta
    created: 2026-07-20T12:00:00Z
    created_by: automated
    source_pattern: VP-001
    source_context: PC-001
    source_boundary: PB-001
    lineage:
      - parent_knowledge: null
        relationship: original

  # Status
  status: approved
  review_date: 2026-07-27
  approved_by: governance
```

---

## Knowledge Validation Checklist

Before finalizing a knowledge object, verify:

- [ ] `id` is unique and follows format
- [ ] `statement` is precise and testable
- [ ] `evidence` has at least one supporting item
- [ ] `statistics` include sample_size and p_value
- [ ] `confidence` has value, level, basis, and factors
- [ ] `context` has at least one dimension or applicability
- [ ] `boundary` is documented if pattern has limits
- [ ] `assumptions` are listed and verified status noted
- [ ] `reproducibility` is tracked
- [ ] `provenance` includes engine information

---

**Document Status**: APPROVED
