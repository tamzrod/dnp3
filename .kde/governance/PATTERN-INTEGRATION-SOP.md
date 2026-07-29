# SOP: External Pattern Integration

**Document ID**: SOP-EXTERNAL-PATTERNS  
**Version**: 1.0.0  
**Date**: 2026-07-28  
**Source**: INV-081 (Caveman/ENZO Evolution Analysis)  
**Status**: APPROVED

---

## Purpose

This SOP provides a systematic approach for integrating external patterns into KDE. It was developed after analyzing the Caveman/ENZO evolution series (INV-055-073) which demonstrated both the value and risks of external pattern integration.

---

## Scope

This SOP applies to:
- External toolkits (e.g., Caveman token reduction)
- External architectures (e.g., ENZO explicitness)
- External methodologies (e.g., from GitHub, papers, other projects)
- Internal patterns that need formalization

---

## Three-Phase Integration Process

### Phase 1: Discovery Phase

**Objective**: Identify and evaluate external patterns for potential integration.

#### Step 1.1: Pattern Identification

| Source | Examples | Search Method |
|--------|----------|---------------|
| GitHub repositories | caveman, enzo, other toolkits | Code search, stars, forks |
| Research papers | New methodologies | Academic databases |
| Other projects | Industry best practices | Benchmarking |
| User requests | Specific needs | Requirements gathering |

#### Step 1.2: Initial Evaluation

Evaluate each pattern against these criteria:

| Criterion | Weight | Description |
|-----------|--------|-------------|
| Token Reduction | 30% | Does it reduce context/token usage? |
| KDE Compatibility | 25% | Can it integrate with KDE architecture? |
| Governance Alignment | 20% | Does it follow KDE principles? |
| Implementation Effort | 15% | How much work to integrate? |
| Risk Level | 10% | Any potential negative impacts? |

#### Step 1.3: Documentation

Document the pattern in an investigation:

```markdown
# INV-{id}: {Pattern Name} Analysis

## Pattern Overview
- Source: {GitHub/Paper/Other}
- Original Purpose: {What it was designed for}
- Token Reduction: {If applicable}

## KDE Evaluation
- Compatibility: {HIGH/MEDIUM/LOW}
- Integration Effort: {X hours/days}
- Risk Level: {LOW/MEDIUM/HIGH}

## Evidence
[EVIDENCE: Source URL or citation]
```

#### Step 1.4: Decision Gate

| Decision | Criteria | Next Action |
|----------|----------|-------------|
| **Approve** | Score ≥ 70%, Risk LOW/MEDIUM | Proceed to Phase 2 |
| **Defer** | Score 50-70%, Risk MEDIUM | Document for future |
| **Reject** | Score < 50%, Risk HIGH | Archive and close |

---

### Phase 2: KDE Evaluation Phase

**Objective**: Test the pattern in a controlled KDE environment.

#### Step 2.1: Environment Setup

Create a test investigation:

```bash
cd laboratory/investigations
mkdir INV-TEST-{pattern}
cd INV-TEST-{pattern}
```

#### Step 2.2: Pre-Flight Check

Run the pre-flight check:

```bash
cd /workspace/project/kde
python3 .kde/commands/check.py
```

Expected output:
```
============================================================
KDE PRE-FLIGHT CHECK
============================================================

  [PASS] Bootstrap Gates: 6/6 checks passed
  [PASS] Runtime State: initialized
  [PASS] ECU Enforcement: ECU check skipped

[OK] Ready for KDE_RUNTIME investigation
```

#### Step 2.3: Pattern Implementation

Implement the pattern in the test investigation:

1. Create README.md with proper header:
```yaml
---
EXECUTION_MODE: KDE_RUNTIME
AUTHENTICITY_SCORE: 100%
RUNTIME_AUTHORITY: Verified
---
```

2. Apply the pattern to a real KDE task
3. Document the application process
4. Measure token usage before/after

#### Step 2.4: Integration Testing

Test integration with KDE components:

| Component | Test | Success Criteria |
|-----------|------|------------------|
| Runtime | Execute test | No errors |
| ECU | Check evidence markers | All present |
| Verification | Run compliance.py | 0 errors |
| Maturity | Run maturity.py | Score ≥ 3 |

#### Step 2.5: Metrics Collection

Measure the pattern's effectiveness:

| Metric | Measurement | Target |
|--------|------------|--------|
| Token Reduction | Compare before/after | ≥ 20% reduction |
| Time Savings | Task completion time | ≥ 15% faster |
| Quality | Output correctness | ≥ 95% accurate |
| Integration | Component compatibility | 100% compatible |

#### Step 2.6: Decision Gate

| Decision | Criteria | Next Action |
|----------|----------|-------------|
| **Approve** | All metrics meet targets | Proceed to Phase 3 |
| **Modify** | Some metrics need work | Return to Step 2.3 |
| **Reject** | Metrics below threshold | Archive and close |

---

### Phase 3: Adoption Phase

**Objective**: Formally integrate the pattern into KDE.

#### Step 3.1: Documentation

Create formal documentation:

1. **Pattern Specification** (in /knowledge/patterns/)
2. **Usage Guide** (in /docs/guides/)
3. **Examples** (in /playground/)
4. **Update CHANGELOG**

#### Step 3.2: Governance Update

If pattern affects governance:

1. Draft DEP (Design Enhancement Proposal)
2. Submit for review
3. Get human approval
4. Update governance documents

#### Step 3.3: Training

Prepare training materials:

| Material | Audience | Format |
|----------|----------|--------|
| Pattern Overview | All users | README.md |
| Usage Tutorial | New users | /docs/guides/ |
| Examples | Developers | /playground/ |
| Best Practices | Advanced users | /knowledge/ |

#### Step 3.4: Rollout

Deploy the pattern:

1. **Phase 1**: Announce in release notes
2. **Phase 2**: Add to default templates
3. **Phase 3**: Mark old patterns as deprecated
4. **Phase 4**: Full migration support

#### Step 3.5: Monitoring

Track pattern usage:

| Metric | Collection Method | Frequency |
|--------|-------------------|-----------|
| Adoption Rate | Git log analysis | Monthly |
| Token Savings | Benchmarking | Quarterly |
| Issue Reports | GitHub issues | As reported |
| User Feedback | Surveys | Annually |

#### Step 3.6: Decision Gate

| Decision | Criteria | Next Action |
|----------|----------|-------------|
| **Promote** | Adoption ≥ 50%, Issues < 5% | Move to production |
| **Maintain** | Adoption 20-50% | Continue support |
| **Deprecate** | Adoption < 20% | Plan sunset |

---

## Caveman Pattern Case Study

Based on INV-081 analysis, here's how Caveman patterns should be integrated:

### From Caveman: Token Reduction Principles

| Caveman Pattern | KDE Application | Feasibility | Integration Status |
|----------------|-----------------|-------------|-------------------|
| Squash over read | Grep/search before full read | HIGH | Phase 1 |
| Diff over re-read | Use git diff | HIGH | Phase 1 |
| Brief tool outputs | Summarize API responses | MEDIUM | Phase 2 |
| One-pass reads | Cache file reads | MEDIUM | Phase 2 |

### Recommended Integration Path

1. **Immediate** (Week 1-2):
   - Add squash/brief pattern to /knowledge/primitives/
   - Create usage examples in /playground/

2. **Short-term** (Month 1):
   - Add to default investigation template
   - Document in /docs/guides/

3. **Long-term** (Quarter 2):
   - Benchmark token savings
   - Update templates based on feedback

---

## ENZO Pattern Case Study

Based on INV-081 analysis, here's how ENZO patterns should be integrated:

### From ENZO: Architecture Principles

| ENZO Principle | KDE Application | Feasibility | Integration Status |
|---------------|-----------------|-------------|-------------------|
| Boundary preservation | Investigation boundaries | HIGH | Phase 1 |
| Explicitness | Explicit evidence markers | HIGH | Phase 1 |
| Mode detection | Execution mode enforcement | MEDIUM | Phase 2 |
| Frame-based output | Structured format | MEDIUM | Phase 2 |

### Recommended Integration Path

1. **Immediate** (Week 1-2):
   - Add boundary preservation to /knowledge/foundation/
   - Update evidence marker requirements

2. **Short-term** (Month 1):
   - Add ENZO-style documentation to templates
   - Create migration guide

3. **Long-term** (Quarter 2):
   - Update verification to check ENZO compliance
   - Add maturity level for ENZO patterns

---

## Templates

### Pattern Investigation Template

```markdown
# INV-{id}: {Pattern Name} Integration

**Status**: INVESTIGATION  
**Pattern Source**: {External Source}
**Phase**: {1|2|3}

---

## Pattern Overview

[EVIDENCE: Source citation]

### Description
{What the pattern does}

### Original Context
{Where it came from}

---

## Phase 1: Discovery

### Initial Evaluation

| Criterion | Score | Notes |
|-----------|-------|-------|
| Token Reduction | X/10 | |
| KDE Compatibility | X/10 | |
| Governance Alignment | X/10 | |
| Implementation Effort | X/10 | |
| Risk Level | X/10 | |

**Total Score**: X/50

### Decision
[ ] Approve - Proceed to Phase 2
[ ] Defer - Document for future
[ ] Reject - Archive and close

---

## Phase 2: KDE Evaluation

### Environment Setup
{Test investigation details}

### Pre-Flight Check
```
$ python3 .kde/commands/check.py
{output}
```

### Metrics

| Metric | Before | After | Change |
|--------|--------|-------|--------|
| Tokens | X | Y | Z% |
| Time | X | Y | Z% |
| Quality | X% | Y% | Z% |

### Decision
[ ] Approve - Proceed to Phase 3
[ ] Modify - Return to Step 2.3
[ ] Reject - Archive and close

---

## Phase 3: Adoption

### Documentation Created
- [ ] Pattern specification
- [ ] Usage guide
- [ ] Examples

### Rollout Status
[ ] Phase 1: Announced
[ ] Phase 2: Added to templates
[ ] Phase 3: Old patterns deprecated
[ ] Phase 4: Migration complete

---

## Lessons Learned

### What Worked
-

### What Could Be Improved
-

### Recommendations for Future
-
```

---

## Appendix: Quick Reference

### Decision Matrix

| Phase | Decision | Criteria | Action |
|-------|----------|----------|--------|
| 1 | Approve | Score ≥ 70% | Proceed to Phase 2 |
| 1 | Defer | Score 50-70% | Document |
| 1 | Reject | Score < 50% | Archive |
| 2 | Approve | All metrics met | Proceed to Phase 3 |
| 2 | Modify | Some metrics | Return to Step 2.3 |
| 2 | Reject | Metrics below | Archive |
| 3 | Promote | Adoption ≥ 50% | Production |
| 3 | Maintain | Adoption 20-50% | Continue |
| 3 | Deprecate | Adoption < 20% | Sunset |

### Required Commands

```bash
# Pre-flight check
python3 .kde/commands/check.py

# Verification
python3 .kde/verification/compliance.py

# Maturity assessment
python3 .kde/verification/maturity.py <path>

# Bootstrap gates
python3 .kde/bootstrap/gates.py
```

### Contact

For questions about this SOP:
- Review INV-081 for context
- Consult /governance/MIGRATION-GUIDE.md
- Check /laboratory/investigations/ for examples

---

**Document Status**: APPROVED  
**Human Review Required**: No (SOP is guidance, not mandatory)  
**Review Date**: 2026-10-28 (Quarterly review)
