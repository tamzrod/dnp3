# KDE-ENGINE-003 (Gamma) Pipeline

**Engine ID**: KDE-ENGINE-003
**Version**: 0.1.0
**Codename**: Gamma

---

## Pipeline Overview

Gamma implements an 8-stage processing pipeline that extends Beta's 6-stage pipeline with causal discovery modules.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         GAMMA CAUSAL PIPELINE                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────┐     ┌─────────────────┐     ┌─────────────────┐          │
│  │   Stage 1   │     │     Stage 2     │     │     Stage 3     │          │
│  │   Evidence   │────▶│   Observation   │────▶│    Pattern      │          │
│  │  Ingestion   │     │   Extraction    │     │    Detector     │          │
│  └─────────────┘     └─────────────────┘     └─────────────────┘          │
│         │                    │                       │                      │
│         ▼                    ▼                       ▼                      │
│  ┌─────────────┐     ┌─────────────────┐     ┌─────────────────┐          │
│  │   Stage 4   │     │     Stage 5     │     │     Stage 6     │          │
│  │  Statistical │────▶│     Context     │────▶│    Boundary     │          │
│  │  Validator   │     │    Detector     │     │    Detector     │          │
│  └─────────────┘     └─────────────────┘     └─────────────────┘          │
│         │                    │                       │                      │
│         └────────────────────┴───────────────────────┘                      │
│                              │                                               │
│                              ▼                                               │
│                    ┌─────────────────┐                                      │
│                    │     Stage 7     │                                      │
│                    │    Causal       │──── NEW                               │
│                    │   Discovery     │                                      │
│                    └─────────────────┘                                      │
│                              │                                               │
│                              ▼                                               │
│                    ┌─────────────────┐                                      │
│                    │     Stage 8     │                                      │
│                    │     Causal      │──── NEW                               │
│                    │    Knowledge    │                                      │
│                    │   Generation    │                                      │
│                    └─────────────────┘                                      │
│                              │                                               │
│                              ▼                                               │
│                       Knowledge                                              │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Stage 1: Evidence Ingestion

### Purpose
Load and prepare evidence artifacts for processing.

### Module
```
EvidenceIngestion
├── Evidence Loader
├── Evidence Validator
└── Evidence Store
```

### Input
- Raw evidence artifacts (logs, measurements, documents)

### Output
- Evidence objects with provenance

### Process

1. **Evidence Loading**
   - Load evidence from sources
   - Preserve original content
   - Attach source metadata

2. **Evidence Validation**
   - Verify evidence integrity
   - Check required fields
   - Flag invalid evidence

3. **Evidence Storage**
   - Store evidence with IDs
   - Index for retrieval
   - Maintain provenance links

### Evidence Object Format

```yaml
evidence:
  id: EV-XXX
  type: [log|measurement|document|event|telemetry]
  content: <original content>
  source:
    type: [file|api|database|manual]
    location: <source identifier>
    timestamp: <collection time>
  collected: <ingestion timestamp>
  metadata:
    collector: <identifier>
    environment: <context>
    quality: [high|medium|low]
```

---

## Stage 2: Observation Extraction

### Purpose
Extract measurable observations from evidence.

### Module
```
ObservationExtraction
├── Evidence Reader
├── Measurement Extractor
└── Observation Formatter
```

### Input
- Evidence objects with provenance

### Output
- Observable facts with measurements

### Process

1. **Evidence Reading**
   - Parse evidence content
   - Identify measurable facts
   - Normalize units

2. **Measurement Extraction**
   - Extract quantitative values
   - Assign units
   - Flag uncertainty

3. **Observation Formatting**
   - Assign observation IDs
   - Attach provenance
   - Include dimensions

### Observation Object Format

```yaml
observation:
  id: OBS-XXX
  type: [measurement|fact|event|behavior]
  value: <measurable value>
  unit: <unit if applicable>
  dimensions:
    temporal: <time dimension>
    environmental: <environment dimension>
    operational: <operation dimension>
  provenance:
    evidence_id: EV-XXX
    extraction_method: <method used>
  uncertainty:
    flagged: [true|false]
    marker: <uncertainty reason if any>
```

---

## Stage 3: Pattern Detection

### Purpose
Identify candidate patterns from observations.

### Module
```
PatternDetection
├── Correlation Analyzer
├── Sequence Detector
└── Conditional Analyzer
```

### Input
- Observation objects

### Output
- Candidate patterns

### Process

1. **Correlation Analysis**
   - Identify variable correlations
   - Calculate correlation coefficients
   - Flag significant correlations

2. **Sequence Detection**
   - Identify temporal sequences
   - Detect preceding events
   - Order by time

3. **Conditional Analysis**
   - Identify conditional relationships
   - Extract IF-THEN patterns
   - Assess conditional strength

### Candidate Pattern Format

```yaml
candidate_pattern:
  id: CP-XXX
  type: [correlation|sequence|conditional]
  relationship:
    x: <variable X>
    y: <variable Y>
    operator: [→|+|→|if]
    condition: <condition if applicable>
  supporting_observations:
    - OBS-XXX
    - OBS-YYY
  frequency: <occurrence rate>
  consistency: [high|medium|low]
  status: candidate
```

---

## Stage 4: Statistical Validation

### Purpose
Validate patterns using statistical methods.

### Module
```
StatisticalValidation
├── Sample Analyzer
├── Significance Tester
└── Interval Calculator
```

### Input
- Candidate patterns

### Output
- Validated patterns with statistics

### Process

1. **Sample Analysis**
   - Assess sample size
   - Check representativeness
   - Flag insufficient samples

2. **Significance Testing**
   - Calculate p-values
   - Apply significance thresholds
   - Reject non-significant patterns

3. **Interval Calculation**
   - Calculate confidence intervals
   - Assess precision
   - Document uncertainty

### Validated Pattern Format

```yaml
validated_pattern:
  id: VP-XXX
  candidate_pattern_id: CP-XXX
  relationship: <relationship description>
  statistics:
    sample_size: <n>
    p_value: <p>
    correlation: <r>
    confidence_interval: [lower, upper]
    chi_square: <χ²>
    effect_size: <d>
  validation_result: [validated|rejected|pending]
  confidence: [high|medium|low]
```

---

## Stage 5: Context Detection

### Purpose
Identify contexts where patterns apply.

### Module
```
ContextDetection
├── Dimension Analyzer
├── Applicability Classifier
└── Context Formatter
```

### Input
- Validated patterns

### Output
- Patterns with contexts

### Process

1. **Dimension Analysis**
   - Search temporal dimensions
   - Search environmental dimensions
   - Search operational dimensions

2. **Applicability Classification**
   - Classify applicability level
   - Identify applicable values
   - Assess context strength

3. **Context Formatting**
   - Format context objects
   - Include dimension values
   - Document conditions

### Context Object Format

```yaml
context:
  id: CTX-XXX
  pattern_id: VP-XXX
  dimensions:
    - name: <dimension>
      values: [applicable values]
      evidence_count: <n>
      statistical_support: <value>
  applicability: [universal|conditional|limited]
```

---

## Stage 6: Boundary Detection

### Purpose
Identify boundaries where patterns fail.

### Module
```
BoundaryDetection
├── Contradiction Analyzer
├── Exception Detector
└── Edge Case Identifier
```

### Input
- Validated patterns with contexts

### Output
- Patterns with boundaries

### Process

1. **Contradiction Analysis**
   - Search for reversals
   - Identify negative correlations
   - Document contradictions

2. **Exception Detection**
   - Identify outlier cases
   - Analyze failure conditions
   - Classify exception types

3. **Edge Case Identification**
   - Test at dimension extremes
   - Identify boundary conditions
   - Document edge failures

### Boundary Object Format

```yaml
boundary:
  id: BND-XXX
  pattern_id: VP-XXX
  type: [contradiction|exception|edge_case|reverse|diminishing]
  condition: <failure condition>
  evidence:
    - observation_id: OBS-XXX
      description: <evidence>
  severity: [critical|major|minor]
  occurrence_rate: <rate>
```

---

## Stage 7: Causal Discovery (NEW)

### Purpose
Identify causal mechanisms and assess causal evidence.

### Module
```
CausalDiscovery
├── Temporal Analyzer
├── Confounder Detector
├── Mechanism Identifier
└── Causal Hypothesis Generator
```

### Input
- Validated patterns with contexts and boundaries

### Output
- Causal hypotheses with evidence

### Process

1. **Temporal Analysis**
   - Analyze temporal ordering
   - Apply temporal precedence rule
   - Identify potential causal directions

2. **Confounder Detection**
   - Search for common causes
   - Identify confounder candidates
   - Assess confounder effects

3. **Mechanism Identification**
   - Search for mediating variables
   - Identify causal pathway steps
   - Document mechanism chain

4. **Causal Hypothesis Generation**
   - Generate causal hypotheses
   - Assess causal evidence strength
   - Assign causal confidence levels

### Causal Hypothesis Format

```yaml
causal_hypothesis:
  id: CH-XXX
  hypothesis: <X causes Y>
  
  # Causal Direction
  direction:
    cause: <X>
    effect: <Y>
    temporal_precedence: [strong|moderate|weak|none]
    rationale: <reasoning>
  
  # Confounders
  confounders:
    - name: <confounder>
      effect_on_cause: <description>
      effect_on_effect: <description>
      adjustment_method: <method used>
      adjusted: [yes|partial|no]
  
  # Mechanism
  mechanism:
    identified: [yes|partial|no]
    steps:
      - step: 1
        from: <node>
        to: <node>
        description: <step description>
      - step: 2
        from: <node>
        to: <node>
        description: <step description>
  
  # Evidence Assessment
  evidence:
    temporal_precedence: [strong|moderate|weak|none]
    mechanism_identified: [yes|partial|no]
    confounder_adjusted: [yes|partial|no]
    intervention_studied: [yes|no]
  
  # Causal Confidence
  causal_strength: [strong|moderate|weak]
  causal_confidence: <0-100>
  
  # Status
  status: [hypothesis|supported|validated|rejected]
```

---

## Stage 8: Causal Knowledge Generation (NEW)

### Purpose
Generate complete causal knowledge objects.

### Module
```
CausalKnowledgeGeneration
├── Effect Size Estimator
├── Intervention Predictor
└── Causal Knowledge Assembler
```

### Input
- Causal hypotheses with evidence
- Pattern statistics
- Context and boundary information

### Output
- Complete causal knowledge objects

### Process

1. **Effect Size Estimation**
   - Calculate average treatment effect
   - Estimate confidence intervals
   - Assess effect magnitude

2. **Intervention Prediction**
   - Predict do(X) outcomes
   - Assess intervention effects
   - Identify potential side effects

3. **Causal Knowledge Assembly**
   - Assemble all causal elements
   - Validate completeness
   - Generate knowledge ID

### Causal Knowledge Format

```yaml
causal_knowledge:
  id: CAUSAL-XXX
  version: "1.0"
  
  # Core Statement
  statement: <X causes Y via mechanism M>
  
  # Causal Mechanism
  mechanism:
    description: <mechanism description>
    steps:
      - step: 1
        description: <step 1>
      - step: 2
        description: <step 2>
    confidence: <confidence value>
  
  # Effect Estimation
  effect:
    type: [ATE|ITE|LATE]
    value: <effect size>
    unit: <unit>
    confidence_interval: [lower, upper]
    confidence: <confidence value>
  
  # Confounders
  confounders:
    - name: <name>
      adjusted: [true|false]
      method: <method>
      residual_effect: <description>
  
  # Intervention
  intervention:
    do_operation: <do(X=x)>
    predicted_outcome:
      value: <predicted>
      confidence_interval: [lower, upper]
    side_effects:
      - effect: <description>
        probability: <value>
  
  # Evidence
  evidence:
    supporting:
      - evidence_id: <ID>
        type: <type>
    contradicting:
      - evidence_id: <ID>
        type: <type>
    total_count: <n>
  
  # Statistics
  statistics:
    sample_size: <n>
    p_value: <p>
    effect_size: <d>
    confidence_interval: [lower, upper]
  
  # Confidence
  confidence:
    value: <0-100>
    level: [high|medium|low]
    basis: [statistical|causal_theory|mixed]
    factors:
      - <factor 1>
      - <factor 2>
  
  # Context
  context:
    dimensions: <applicable dimensions>
    conditions: <specific conditions>
    applicability: [universal|conditional|limited]
  
  # Boundaries
  boundary:
    type: <type>
    conditions: <conditions>
    severity: <severity>
  
  # Assumptions
  assumptions:
    - <assumption 1>
    - <assumption 2>
  
  # Reproducibility
  reproducibility:
    verified: [true|false]
    replications: <n>
  
  # Provenance
  provenance:
    engine_id: KDE-ENGINE-003
    engine_version: 0.1.0
    created: <timestamp>
  
  # Status
  status: [draft|validated|approved|published]
```

---

## Data Flow Summary

```
Input: Raw Evidence
    │
    ▼
Stage 1: Evidence → Evidence Objects with Provenance
    │
    ▼
Stage 2: Evidence Objects → Observations with Dimensions
    │
    ▼
Stage 3: Observations → Candidate Patterns
    │
    ▼
Stage 4: Candidate Patterns → Validated Patterns + Statistics
    │
    ▼
Stage 5: Validated Patterns → Patterns with Context
    │
    ▼
Stage 6: Patterns with Context → Patterns with Boundaries
    │
    ▼
Stage 7: Patterns → Causal Hypotheses + Mechanism + Confounders
    │
    ▼
Stage 8: Causal Hypotheses → Causal Knowledge Objects
    │
    ▼
Output: Causal Knowledge
```

---

## Error Handling

### Stage-Specific Errors

| Stage | Error | Recovery |
|-------|-------|----------|
| 1 | Evidence not found | Skip or retry |
| 2 | No observations extracted | Adjust extraction |
| 3 | No patterns detected | Continue observation |
| 4 | Statistical validation fails | Reject pattern |
| 5 | No context found | Flag as context-free |
| 6 | No boundaries detected | Document as untested |
| 7 | No causal evidence | Mark as correlation |
| 8 | Incomplete knowledge | Return for completion |

---

## Related Documents

- [methodology.md](./methodology.md) — Full methodology
- [knowledge-model.md](./knowledge-model.md) — Knowledge structure

---

**Document Status**: EXPERIMENTAL
