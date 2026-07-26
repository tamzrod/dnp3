# Experiment Specification: DNP3-EXP-007

**Experiment ID**: DNP3-EXP-007
**Title**: NRF Article Screening System — Relevance & Categorization Engine
**Status**: IN_PROGRESS
**Date**: 2026-07-26

---

## Experiment Scope

### In Scope
- Part 1: Relevance Check (Rule 0, Gates 0-3)
- Part 2: Categorization (MTC, HHP, USS, SNDE)
- Rule-based classification engine
- Processing of NRF media monitoring data from EXP-006

### Out of Scope
- Ranking or scoring articles
- Second-pass processing
- Output formatting for NRF delivery
- Integration with external systems

---

## Objectives

| ID | Objective | Status |
|----|-----------|--------|
| O1 | Implement Rule 0 Standing Order Override | IN_PROGRESS |
| O2 | Implement Gate 0 (Article Existence) | IN_PROGRESS |
| O3 | Implement Gate 1 (Framework Fit) | IN_PROGRESS |
| O4 | Implement Gate 2 (Portability) | IN_PROGRESS |
| O5 | Implement Gate 3 (Research Substance) | IN_PROGRESS |
| O6 | Implement Part 2 Categorization | IN_PROGRESS |
| O7 | Process sample articles | PENDING |
| O8 | Validate against expected patterns from EXP-006 | PENDING |

---

## Data Sources

| Source | Type | Relevance |
|--------|------|-----------|
| /tmp/spreadsheet.xlsx | Data | Primary - from DNP3-EXP-006 |
| Consolidated SL sheet | Data Extract | Main screening target |
| Headlines column | Content | Article titles |
| URL/Notes column | Content | Article content reference |

---

## Methodology

### Phase 1: Bootstrap Gate Verification
**Result**: ✅ 8/8 GATES PASSED
**Timestamp**: 2026-07-26T11:27:16

### Phase 2: Data Acquisition
1. Load XLSX file from /tmp/spreadsheet.xlsx
2. Parse Consolidated SL sheet
3. Extract headlines and content

### Phase 3: Part 1 - Relevance Check

#### Rule 0: Standing Order Override
Check for mandated organizations:
- IHL: NUS, NTU, SMU, SUTD, SIT, SUSS
- Hospitals: SingHealth, SGH, NUH, TTSH, CGH, KTPH, SKH, NTFGH, AH, WH, KKH, NCCS, NHCS, SNEC, IMH, NDCS, NSC
- Academic Medical: SingHealth Duke-NUS, NUHS, NHG-LKCMedicine
- A*STAR: A*STAR, ID Labs, A*SRL, BII, BTI, GIS, IMCB, SIgN, SICS, SIFBI
- Research Infrastructure: TCOMS, National Supercomputing Centre of Singapore, St John's Island National Marine Laboratory
- Research Institutes: COI BE-AM, MBI, IDMXS, I-FIM, CQT, CSI, ERI@N, NEWRI, SCELSE
- CREATE: CREATE, BEARS, CARES, E2S2, SHARE, SEC, SMART, TUM CREATE, TSCP, CNRS@CREATE

#### Gate 0: Article Existence
Fail if:
- Paywall stub, consent wall, or abstract with nothing behind it
- Roundup or diary entry (not "In brief" section label)

#### Gate 1: Framework Fit
Pass if article fits one of 21 topics:
- MTC: Manufacturing, Robotics, Trade, Electronics
- HHP: GUSTO, PRECISE, Ageing, PREPARE, General Health
- USS: Climate, Sustainability, Resource Resilience, Built Environment
- SNDE: Smart Nation, Cybersecurity, Communications, Quantum, Robotics

Fail if outside framework (after applying cross-cutting topics).

#### Gate 2: Portability
Pass if:
- Location is incidental (findings travel)
- Global frameworks (COP, IPCC, WHO, UN)

Fail if location is load-bearing AND not Singapore/SEA.

#### Gate 3: Research Substance
Track A (Research): Pass if pathway to application, breakthrough, method, clinical results
Track B (News): Pass if R&D, innovation, facilities, tech transfer, collaboration

### Phase 4: Part 2 - Categorization

Apply in order:
1. Rule 1: All nuclear → USS
2. Rule 2: Nvidia/AI/Quantum → SNDE (default)
3. Rule 3: Domain override (what work does)
4. Rule 4: Energy/sustainability → USS
5. Rule 5: Geography doesn't decide category

---

## Success Criteria

| Criterion | Evidence | Status |
|-----------|----------|--------|
| Bootstrap 8/8 | Gate verification | PASSED |
| Rule 0 implemented | Screening output | PENDING |
| Gates 0-3 implemented | Screening output | PENDING |
| Part 2 implemented | Categorization output | PENDING |
| Articles processed | Sample results | PENDING |
| Patterns match EXP-006 | Validation | PENDING |

---

**Spec Status**: IN_PROGRESS
**Created**: 2026-07-26
