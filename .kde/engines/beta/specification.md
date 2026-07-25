# KDE-ENGINE-002 (Beta) Specification

**Engine ID**: KDE-ENGINE-002
**Version**: 0.1.0
**Codename**: Beta
**Name**: Contextual Knowledge Discovery Engine (CKDE)
**Status**: Active
**Effective Date**: 2026-07-20

---

## Engine Identity

| Field | Value |
|-------|-------|
| **Engine ID** | KDE-ENGINE-002 |
| **Version** | 0.1.0 |
| **Codename** | Beta |
| **Name** | Contextual Knowledge Discovery Engine |
| **Short Name** | CKDE |
| **Status** | Active |
| **Effective Date** | 2026-07-20 |
| **Parent Engine** | KDE-ENGINE-001 (Alpha) |

---

## Purpose

KDE-ENGINE-002 (Beta) is the **Contextual Knowledge Discovery Engine**. It evolves beyond simple pattern discovery to discover not only patterns, but also the **contexts**, **boundaries**, **confidence**, and **applicability** of those patterns.

### Primary Discovery Principle

> Knowledge is rarely universal. Knowledge is usually conditional.

Beta transforms observations into scientific knowledge by discovering:
- **What** is true
- **When** it is true
- **Where** it is true
- **When** it stops being true
- **How confident** we are

---

## Scope

### What This Engine Covers

- Evidence ingestion and observation extraction
- Pattern detection with statistical validation
- Context discovery across all available dimensions
- Boundary detection for pattern applicability limits
- Scientific knowledge generation with full provenance
- Confidence estimation with statistical support

### What This Engine Does NOT Cover

- Software implementation specifics
- Infrastructure decisions
- Tool configurations
- External service integrations

---

## Design Philosophy

### From Pattern to Knowledge

```
Alpha (KDE-ENGINE-001):
    Pattern → Knowledge

Beta (KDE-ENGINE-002):
    Pattern → Context → Boundary → Confidence → Knowledge
```

### Question Transformation

| Engine | Asks | Discovers |
|--------|------|-----------|
| **Alpha** | "Does X correlate with Y?" | Patterns |
| **Beta** | "Under what conditions does X correlate with Y?" | Contextual Knowledge |

---

## Engine Pipeline

Beta implements a six-module processing pipeline:

```
┌─────────────┐
│   Evidence   │
└──────┬──────┘
       │
       ▼
┌─────────────────┐
│  Observation    │
│    Engine       │
└──────┬─────────┘
       │
       ▼
┌─────────────────┐
│   Pattern       │
│   Detector      │
└──────┬─────────┘
       │
       ▼
┌─────────────────┐
│  Statistical    │
│   Validator     │
└──────┬─────────┘
       │
       ▼
┌─────────────────┐
│   Context       │
│   Detector      │
└──────┬─────────┘
       │
       ▼
┌─────────────────┐
│   Boundary      │
│   Detector      │
└──────┬─────────┘
       │
       ▼
┌─────────────────┐
│   Knowledge     │
│   Generator     │
└─────────────────┘
```

---

## Engine Modules

### Module 1: Observation Engine

**Responsibilities:**
- Read evidence
- Extract measurable observations
- Normalize observations
- Preserve provenance

**Output:** Observation objects

### Module 2: Pattern Detector

**Responsibilities:**
- Discover repeated relationships
- Identify A → B correlations
- Identify A + B → C patterns
- Detect repeated sequences and behaviors
- Generate candidate patterns (no conclusions)

**Output:** Candidate Pattern

### Module 3: Statistical Validator

**Responsibilities:**
- Validate candidate patterns statistically
- Compute sample size, confidence, variance
- Calculate correlation coefficients
- Compute p-values and chi-square
- Reject weak patterns

**Output:** Validated Pattern

### Module 4: Context Detector (NEW)

**Responsibilities:**
- Determine conditions where pattern exists
- Search every available dimension
- Identify applicable contexts

**Dimensions searched:**
- Time, Phase, Environment, Configuration
- Operating Mode, Location, Input Size
- Temperature, Voltage, Pressure
- Domain-specific dimensions

**Output:** Applicable Contexts

### Module 5: Boundary Detector (NEW)

**Responsibilities:**
- Discover where validated patterns fail
- Search for contradictions, exceptions, edge cases
- Identify outliers and reverse correlations
- Define applicability boundaries

**Output:** Applicability Boundaries

### Module 6: Knowledge Generator

**Responsibilities:**
- Generate scientific knowledge objects
- Include all required fields (ID, statement, evidence, etc.)
- Set confidence and statistical support
- Document contexts and boundaries

**Output:** Scientific Knowledge

---

## Relationship to Alpha (KDE-ENGINE-001)

| Aspect | Alpha | Beta |
|--------|-------|------|
| **Discovery Type** | Pattern | Contextual Knowledge |
| **Questions** | "Does X?" | "When does X?" |
| **Output** | Patterns | Knowledge with Context |
| **Confidence** | Implicit | Explicit, Statistical |
| **Boundaries** | Not defined | Explicitly documented |
| **Contexts** | Not discovered | Automatically detected |

### Compatibility

- **Alpha** remains unchanged for historical comparison
- **Beta** operates independently
- Laboratory can execute either engine
- Engine selection through configuration

---

## Engine Lifecycle

### Current Status: Active

This engine is currently active and available for new experiments.

### Lifecycle States

| State | Description |
|-------|-------------|
| **Active** | Available for new experiments |
| **Historical** | Former engine, preserved for reference |
| **Deprecated** | Not recommended |

### Transition from Alpha

| Field | Alpha | Beta |
|-------|-------|------|
| Engine ID | KDE-ENGINE-001 | KDE-ENGINE-002 |
| Status | Historical | Active |
| Purpose | Pattern discovery | Contextual knowledge |

---

## Provenance

This engine is the successor to KDE-ENGINE-001 (Alpha).

See [provenance.md](./provenance.md) for detailed history.

---

## Version Information

| Field | Value |
|-------|-------|
| **Engine ID** | KDE-ENGINE-002 |
| **Version** | 0.1.0 |
| **Codename** | Beta |
| **Created** | 2026-07-20 |
| **Effective** | 2026-07-20 |
| **Status** | Active |

---

**Document Status**: APPROVED
**Review Date**: Upon significant methodology change
