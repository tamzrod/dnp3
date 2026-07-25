# KDE Engine Interface Specification

**Version**: 1.0
**Effective**: 2026-07-20

---

## Overview

This document defines the standard interface that all KDE Engines must implement. The interface enables the Laboratory to execute any engine without modification, making the Laboratory **engine-agnostic**.

---

## Design Principle

```
The Laboratory executes engines.
Engines define methodology.
The interface connects them.
```

---

## Interface Definition

### Required Methods

Every KDE Engine MUST implement the following methods:

```yaml
interface:
  name: KDEEngineInterface
  version: "1.0"
  
  methods:
    Initialize:
      description: Initialize the engine
      parameters: []
      returns: EngineState
      
    Analyze:
      description: Process evidence into knowledge
      parameters:
        - name: evidence
          type: Evidence[]
      returns: AnalysisResult
      
    Validate:
      description: Validate knowledge object
      parameters:
        - name: knowledge
          type: Knowledge
      returns: ValidationResult
      
    GenerateKnowledge:
      description: Create knowledge from pipeline output
      parameters:
        - name: pipeline_output
          type: PipelineOutput
      returns: Knowledge
      
    GenerateReport:
      description: Format findings for consumption
      parameters:
        - name: knowledge
          type: Knowledge
      returns: Report
      
    Capabilities:
      description: Return engine capabilities
      parameters: []
      returns: Capabilities
      
    Version:
      description: Return engine version
      parameters: []
      returns: Version
      
    Metadata:
      description: Return engine metadata
      parameters: []
      returns: EngineMetadata
```

---

## Method Specifications

### 1. Initialize()

**Purpose**: Set up the engine for operation.

```yaml
Initialize:
  description: |
    Initialize the engine and all its modules.
    Verify prerequisites.
    Load configuration.
  
  parameters: []
  
  returns:
    type: EngineState
    description: Current state of the engine
    
  errors:
    - InitializationError
    - ConfigurationError
    - PrerequisiteError
```

**State Transitions**:
```
UNINITIALIZED → INITIALIZING → READY
                     ↓
                 ERROR
```

### 2. Analyze(evidence)

**Purpose**: Process evidence through the engine pipeline.

```yaml
Analyze:
  description: |
    Execute the full engine pipeline on the provided evidence.
    Returns analysis results including patterns, contexts, boundaries,
    and knowledge objects.
  
  parameters:
    - name: evidence
      type: Evidence[]
      required: true
      description: Evidence artifacts to analyze
  
  returns:
    type: AnalysisResult
    description: Complete analysis output
```

**AnalysisResult Structure**:
```yaml
AnalysisResult:
  id: RESULT-XXX
  engine_id: KDE-ENGINE-XXX
  engine_version: "X.X.X"
  timestamp: ISO8601
  
  observations:
    - Observation[]
  
  candidate_patterns:
    - CandidatePattern[]
  
  validated_patterns:
    - ValidatedPattern[]
  
  contexts:
    - Context[]
  
  boundaries:
    - Boundary[]
  
  knowledge:
    - Knowledge[]
  
  statistics:
    processing_time: duration
    evidence_count: integer
    observation_count: integer
    pattern_count: integer
    knowledge_count: integer
```

### 3. Validate(knowledge)

**Purpose**: Validate a knowledge object for completeness and correctness.

```yaml
Validate:
  description: |
    Verify that a knowledge object meets all requirements.
    Check statistics, context, boundaries, and provenance.
  
  parameters:
    - name: knowledge
      type: Knowledge
      required: true
      description: Knowledge object to validate
  
  returns:
    type: ValidationResult
    description: Validation outcome with any issues found
```

**ValidationResult Structure**:
```yaml
ValidationResult:
  valid: boolean
  issues:
    - type: [missing_field|invalid_value|incomplete|warning]
      field: field_name
      message: description
      severity: [error|warning|info]
  passed_checks:
    - check_name
    - check_name
```

### 4. GenerateKnowledge(pipeline_output)

**Purpose**: Create a knowledge object from pipeline output.

```yaml
GenerateKnowledge:
  description: |
    Transform validated patterns, contexts, and boundaries
    into a complete knowledge object.
  
  parameters:
    - name: pipeline_output
      type: PipelineOutput
      required: true
      description: Output from Analyze() method
  
  returns:
    type: Knowledge
    description: Complete knowledge object
```

### 5. GenerateReport(knowledge)

**Purpose**: Format knowledge for human consumption.

```yaml
GenerateReport:
  description: |
    Generate a formatted report from knowledge objects.
    Include all relevant details in readable format.
  
  parameters:
    - name: knowledge
      type: Knowledge
      required: true
      description: Knowledge object to format
  
  returns:
    type: Report
    description: Formatted report
```

**Report Structure**:
```yaml
Report:
  id: REPORT-XXX
  knowledge_id: KNOW-XXX
  format: [markdown|html|pdf|json]
  
  sections:
    - title: "Knowledge Statement"
      content: formatted_statement
    - title: "Evidence"
      content: formatted_evidence
    - title: "Statistics"
      content: formatted_statistics
    - title: "Confidence"
      content: formatted_confidence
    - title: "Context"
      content: formatted_context
    - title: "Boundaries"
      content: formatted_boundaries
  
  metadata:
    generated: timestamp
    engine_id: engine_id
    version: version
```

### 6. Capabilities()

**Purpose**: Return what the engine can do.

```yaml
Capabilities:
  description: |
    Return the engine's capabilities including supported
    dimensions, statistics, and features.
  
  parameters: []
  
  returns:
    type: Capabilities
    description: Engine capabilities
```

**Capabilities Structure**:
```yaml
Capabilities:
  engine_id: KDE-ENGINE-XXX
  version: "X.X.X"
  
  supported_modules:
    - ObservationEngine
    - PatternDetector
    - StatisticalValidator
    - ContextDetector
    - BoundaryDetector
    - KnowledgeGenerator
  
  supported_dimensions:
    temporal: [list]
    environmental: [list]
    operational: [list]
    domain_specific: [list]
  
  supported_statistics:
    - p_value
    - correlation
    - chi_square
    - confidence_interval
    - effect_size
    - sample_size
  
  supported_pattern_types:
    - correlation
    - sequence
    - conditional
    - behavioral
    - comparative
  
  maximum_sample_size: integer
  minimum_sample_size: integer
  
  features:
    context_detection: boolean
    boundary_detection: boolean
    causal_inference: boolean
    temporal_tracking: boolean
```

### 7. Version()

**Purpose**: Return engine version information.

```yaml
Version:
  description: |
    Return semantic version of the engine.
  
  parameters: []
  
  returns:
    type: Version
    description: Version information
```

**Version Structure**:
```yaml
Version:
  major: integer
  minor: integer
  patch: integer
  codename: string
  full: "X.Y.Z"
```

### 8. Metadata()

**Purpose**: Return engine identity and status.

```yaml
Metadata:
  description: |
    Return complete engine metadata including
    identity, status, and documentation links.
  
  parameters: []
  
  returns:
    type: EngineMetadata
    description: Engine metadata
```

**EngineMetadata Structure**:
```yaml
EngineMetadata:
  id: KDE-ENGINE-XXX
  name: "Engine Name"
  codename: "codename"
  version: "X.Y.Z"
  status: [active|historical|deprecated]
  effective_date: ISO8601
  
  documentation:
    specification: path
    methodology: path
    pipeline: path
    knowledge_model: path
  
  parent_engine: engine_id_or_null
  child_engine: engine_id_or_null
  
  laboratory_version_required: version
```

---

## Data Types

### Evidence

```yaml
Evidence:
  id: EV-XXX
  type: [log|screenshot|measurement|document|event|telemetry]
  content: any
  source: string
  timestamp: ISO8601
  metadata: object
```

### Observation

```yaml
Observation:
  id: OBS-XXX
  type: [measurement|fact|event|behavior]
  value: any
  unit: string_or_null
  dimensions: object
  provenance:
    evidence_id: EV-XXX
    extraction_method: string
  uncertainty:
    flagged: boolean
    marker: string_or_null
```

### CandidatePattern

```yaml
CandidatePattern:
  id: CP-XXX
  type: [correlation|sequence|conditional|behavioral|comparative]
  pattern:
    antecedent: any
    consequent: any
    condition: any
  supporting_observations: [OBS-XXX]
  frequency: number
  consistency: [high|medium|low]
  status: candidate
```

### ValidatedPattern

```yaml
ValidatedPattern:
  id: VP-XXX
  candidate_pattern_id: CP-XXX
  statistics:
    sample_size: integer
    p_value: float
    correlation: float
    confidence_interval: [float, float]
    chi_square: float
    effect_size: float
  result: [validated|rejected|pending_review]
  confidence: [high|medium|low]
```

### Context

```yaml
Context:
  id: CTX-XXX
  pattern_id: VP-XXX
  dimension:
    name: string
    values: [any]
    evidence_count: integer
    statistical_support: float
  applicability: [universal|conditional|limited]
```

### Boundary

```yaml
Boundary:
  id: BND-XXX
  pattern_id: VP-XXX
  type: [contradiction|exception|edge_case|reverse|diminishing]
  condition: string
  evidence:
    - observation_id: OBS-XXX
      description: string
  severity: [critical|major|minor]
  occurrence_rate: float
```

### Knowledge

```yaml
Knowledge:
  id: KNOW-XXX
  version: "X.Y"
  statement: string
  evidence:
    supporting: [EvidenceRef]
    contradicting: [EvidenceRef]
    total_count: integer
  statistics: Statistics
  confidence:
    value: integer
    level: [high|medium|low]
    basis: [statistical|expert|heuristic]
    factors: [string]
  context: Context
  boundary: Boundary_or_null
  assumptions: [Assumption]
  reproducibility:
    verified: boolean
    verification_count: integer
  provenance:
    engine_id: string
    engine_version: string
    created: ISO8601
  status: [draft|validated|approved|published]
```

---

## Error Handling

### Standard Errors

| Error | Description | Handling |
|-------|-------------|----------|
| `InitializationError` | Engine failed to initialize | Retry or abort |
| `ConfigurationError` | Invalid configuration | Fix config |
| `PrerequisiteError` | Prerequisites not met | Install deps |
| `ValidationError` | Evidence invalid | Skip invalid |
| `ProcessingError` | Pipeline failed | Log and retry |
| `ResourceError` | Insufficient resources | Reduce scope |

### Error Response Format

```yaml
Error:
  code: ERROR_CODE
  message: human_readable_message
  details: object
  recoverable: boolean
  retry_after: duration_or_null
```

---

## Implementation Requirements

### Language Independence

The interface is language-agnostic. Implementations may be in any language that supports:
- Struct/mapping data types
- Arrays/lists
- Type definitions
- Error handling

### Serialization

Engines MUST support JSON serialization for:
- Input parameters
- Return values
- Error responses

### Thread Safety

Engines SHOULD be thread-safe for concurrent execution.

### Resource Management

Engines MUST:
- Clean up resources on shutdown
- Report resource usage
- Support resource limits

---

## Versioning

### Interface Versioning

The interface follows semantic versioning:
- MAJOR: Breaking changes
- MINOR: Additive changes
- PATCH: Bug fixes

### Engine Versioning

Each engine maintains its own version independent of the interface.

---

## Related Documents

- [engine/README.md](./README.md) — Engine framework overview
- [alpha/methodology.md](./alpha/methodology.md) — Alpha methodology
- [beta/methodology.md](./beta/methodology.md) — Beta methodology
- [beta/pipeline.md](./beta/pipeline.md) — Pipeline documentation

---

**Document Status**: APPROVED
